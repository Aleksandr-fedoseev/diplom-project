// Package submissions — заявки пользователей и их модерация.
// Пользователь предлагает новый объект или дополнение (фото, сведения) к существующему,
// модератор одобряет или отклоняет заявку, автор получает письмо о решении.
// Модуль — владелец таблиц submissions и moderation_decisions: объекты, файлы,
// пользователей и письма он получает через сервисы других модулей.
package submissions

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"heritage-backend/internal/auth"
	"heritage-backend/internal/catalog"
	"heritage-backend/internal/httpx"
	"heritage-backend/internal/media"
	"heritage-backend/internal/notifications"
	"heritage-backend/internal/submissions/store"
)

const (
	KindNewObject = "new_object" // предложение нового объекта
	KindAddition  = "addition"   // фото и сведения к существующему объекту

	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"

	maxFiles     = 10  // файлов в одной заявке
	maxListLimit = 200 // заявок в одном ответе очереди
	shortDescLen = 300 // длина краткого описания, если модератор его не задал
)

type Decision struct {
	Decision    string    `json:"decision"` // approved или rejected
	Reason      string    `json:"reason,omitempty"`
	ModeratorID *int64    `json:"moderatorId,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Author — автор заявки; показывается модератору в карточке заявки.
type Author struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	IsBlocked bool   `json:"isBlocked"`
}

type Submission struct {
	ID          int64                `json:"id"`
	AuthorID    int64                `json:"authorId"`
	Author      *Author              `json:"author,omitempty"` // только в карточке для модератора
	Kind        string               `json:"kind"`
	ObjectID    *int64               `json:"objectId,omitempty"`
	ObjectSlug  string               `json:"objectSlug,omitempty"`
	ObjectTitle string               `json:"objectTitle,omitempty"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Address     string               `json:"address"`
	Coordinates *catalog.Coordinates `json:"coordinates,omitempty"`
	Period      string               `json:"period"`
	Source      string               `json:"source"`
	Status      string               `json:"status"`
	Decision    *Decision            `json:"decision,omitempty"`
	Media       []media.File         `json:"media,omitempty"` // только в карточке заявки
	CreatedAt   time.Time            `json:"createdAt"`
}

// Интерфейсы ниже объявлены на стороне модуля заявок, а реализуют их сервисы других модулей.
// Так заявки не читают чужие таблицы, а в тестах вместо MinIO подставляется заглушка.

// Catalog реализует catalog.Service.
type Catalog interface {
	CreateTx(ctx context.Context, tx pgx.Tx, in catalog.ObjectInput) (catalog.Object, error)
	GetBySlug(ctx context.Context, slug string) (catalog.Object, error)
	PlacesByIDs(ctx context.Context, ids []int64) ([]catalog.Place, error)
}

// Media реализует media.Service.
type Media interface {
	Upload(ctx context.Context, in media.UploadInput) (media.File, error)
	ListBySubmission(ctx context.Context, submissionID int64) ([]media.File, error)
	AttachSubmissionTx(ctx context.Context, tx pgx.Tx, submissionID, objectID int64) error
	MaxUploadBytes() int64
}

// Users реализует auth.Service.
type Users interface {
	GetByID(ctx context.Context, id int64) (auth.User, error)
	EnsureActive(ctx context.Context, userID int64) error
	SetBlockedTx(ctx context.Context, tx pgx.Tx, actorID int64, actorRole string, userID int64, blocked bool) (auth.User, error)
}

// Notifier реализует notifications.Service.
type Notifier interface {
	EnqueueTx(ctx context.Context, tx pgx.Tx, msg notifications.Message) error
}

type Service struct {
	db        *pgxpool.Pool  // нужен только для транзакций
	q         *store.Queries // запросы из queries.sql, код сгенерирован sqlc
	catalog   Catalog
	media     Media
	users     Users
	notify    Notifier
	publicURL string // адрес сайта для ссылок в письмах
}

func NewService(db *pgxpool.Pool, catalog Catalog, media Media, users Users, notify Notifier, publicURL string) *Service {
	return &Service{
		db: db, q: store.New(db), catalog: catalog, media: media, users: users, notify: notify,
		publicURL: strings.TrimRight(publicURL, "/"),
	}
}

// Viewer — кто смотрит заявку: автор видит только свои, модератор — любые.
type Viewer struct {
	UserID      int64
	IsModerator bool
}

// CreateInput — заявка пользователя.
// Для дополнения объект указывается по slug (как в адресе карточки /objects/{slug}).
type CreateInput struct {
	Kind        string               `json:"kind"` // new_object или addition
	ObjectSlug  string               `json:"objectSlug"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Address     string               `json:"address"`
	Coordinates *catalog.Coordinates `json:"coordinates"`
	Period      string               `json:"period"`
	Source      string               `json:"source"`
}

func (in *CreateInput) normalize() {
	in.Kind = strings.TrimSpace(in.Kind)
	in.ObjectSlug = strings.TrimSpace(strings.ToLower(in.ObjectSlug))
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	in.Address = strings.TrimSpace(in.Address)
	in.Period = strings.TrimSpace(in.Period)
	in.Source = strings.TrimSpace(in.Source)
}

// validate проверяет поля, общие для обоих видов заявок, и требования к новому объекту.
func (in *CreateInput) validate() error {
	if c := in.Coordinates; c != nil && (c.Lat < -90 || c.Lat > 90 || c.Lng < -180 || c.Lng > 180) {
		return httpx.BadRequest("некорректные координаты")
	}
	if in.Source == "" {
		return httpx.BadRequest("укажите источник сведений (архив, книга, личный архив и т. п.)")
	}
	switch in.Kind {
	case KindNewObject:
		switch {
		case in.Title == "":
			return httpx.BadRequest("укажите название объекта")
		case in.Address == "" && in.Coordinates == nil:
			return httpx.BadRequest("укажите место: адрес или координаты")
		case in.Period == "":
			return httpx.BadRequest("укажите период или дату")
		}
	case KindAddition:
		if in.ObjectSlug == "" {
			return httpx.BadRequest("укажите объект, к которому относится дополнение (objectSlug)")
		}
	default:
		return httpx.BadRequest("kind: new_object или addition")
	}
	return nil
}

// Create подаёт заявку. Файлы прикрепляются отдельными запросами (UploadMedia) после создания.
func (s *Service) Create(ctx context.Context, authorID int64, in CreateInput) (Submission, error) {
	in.normalize()
	if err := in.validate(); err != nil {
		return Submission{}, err
	}
	if err := s.users.EnsureActive(ctx, authorID); err != nil {
		return Submission{}, err
	}

	var objectID *int64
	if in.Kind == KindAddition {
		obj, err := s.catalog.GetBySlug(ctx, in.ObjectSlug)
		if err != nil {
			return Submission{}, err // 404, если объекта нет
		}
		if !obj.IsPublished {
			return Submission{}, httpx.NotFound("объект не найден") // о скрытых объектах не сообщаем
		}
		objectID = &obj.ID
		if in.Title == "" {
			in.Title = "Дополнение к объекту «" + obj.Title + "»"
		}
	}

	var lat, lng *float64
	if c := in.Coordinates; c != nil {
		lat, lng = &c.Lat, &c.Lng
	}
	id, err := s.q.CreateSubmission(ctx, store.CreateSubmissionParams{
		AuthorID:    authorID,
		Kind:        in.Kind,
		ObjectID:    objectID,
		Title:       in.Title,
		Description: in.Description,
		Address:     in.Address,
		Lat:         lat,
		Lng:         lng,
		Period:      in.Period,
		Source:      in.Source,
	})
	if err != nil {
		return Submission{}, fmt.Errorf("создание заявки: %w", err)
	}
	return s.Get(ctx, id, Viewer{UserID: authorID})
}

// UploadInput — файл к заявке; заполняется обработчиком из multipart-формы.
type UploadInput struct {
	Kind         string
	Caption      string
	YearTaken    string
	Source       string
	OriginalName string
	Size         int64
	Body         io.Reader
}

// MaxUploadBytes — предел размера тела запроса с файлом (берётся у модуля медиа).
func (s *Service) MaxUploadBytes() int64 { return s.media.MaxUploadBytes() }

// UploadMedia прикрепляет файл к заявке. Можно только автору и только пока заявка ожидает решения.
func (s *Service) UploadMedia(ctx context.Context, submissionID, userID int64, in UploadInput) (media.File, error) {
	sub, err := s.q.GetSubmission(ctx, submissionID)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && sub.AuthorID != userID) {
		return media.File{}, httpx.NotFound("заявка не найдена")
	}
	if err != nil {
		return media.File{}, err
	}
	if sub.Status != StatusPending {
		return media.File{}, httpx.Conflict("по заявке уже принято решение, добавлять файлы нельзя")
	}
	if err := s.users.EnsureActive(ctx, userID); err != nil {
		return media.File{}, err
	}
	files, err := s.media.ListBySubmission(ctx, submissionID)
	if err != nil {
		return media.File{}, err
	}
	if len(files) >= maxFiles {
		return media.File{}, httpx.Conflict(fmt.Sprintf("к заявке можно прикрепить не больше %d файлов", maxFiles))
	}
	return s.media.Upload(ctx, media.UploadInput{
		SubmissionID: submissionID,
		Kind:         in.Kind,
		Caption:      strings.TrimSpace(in.Caption),
		YearTaken:    strings.TrimSpace(in.YearTaken),
		Source:       strings.TrimSpace(in.Source),
		OriginalName: in.OriginalName,
		Size:         in.Size,
		Body:         in.Body,
		UploadedBy:   userID,
	})
}

// Get возвращает карточку заявки с файлами и решением. Чужую заявку пользователь не видит:
// отвечаем 404, чтобы по ответу нельзя было узнать, что такая заявка есть.
// Модератору дополнительно показываются сведения об авторе.
func (s *Service) Get(ctx context.Context, id int64, v Viewer) (Submission, error) {
	row, err := s.q.GetSubmission(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && row.AuthorID != v.UserID && !v.IsModerator) {
		return Submission{}, httpx.NotFound("заявка не найдена")
	}
	if err != nil {
		return Submission{}, err
	}
	list := []Submission{fromRow(row)}
	if err := s.fillObjects(ctx, list); err != nil {
		return Submission{}, err
	}
	sub := list[0]

	if sub.Media, err = s.media.ListBySubmission(ctx, id); err != nil {
		return Submission{}, err
	}
	if sub.Media == nil {
		sub.Media = []media.File{}
	}
	if v.IsModerator {
		u, err := s.users.GetByID(ctx, sub.AuthorID)
		if err != nil {
			return Submission{}, err
		}
		sub.Author = &Author{ID: u.ID, Name: u.Name, Email: u.Email, IsBlocked: u.IsBlocked}
	}
	return sub, nil
}

// ListByAuthor — свои заявки пользователя со статусом и причиной отказа, новые первыми.
func (s *Service) ListByAuthor(ctx context.Context, authorID int64) ([]Submission, error) {
	rows, err := s.q.ListSubmissionsByAuthor(ctx, authorID)
	if err != nil {
		return nil, err
	}
	list := make([]Submission, len(rows))
	for i, r := range rows {
		list[i] = fromRow(store.GetSubmissionRow(r))
	}
	return list, s.fillObjects(ctx, list)
}

// ListByStatus — очередь модерации: заявки с данным статусом, старые первыми.
func (s *Service) ListByStatus(ctx context.Context, status string, limit, offset int32) ([]Submission, error) {
	if status != StatusPending && status != StatusApproved && status != StatusRejected {
		return nil, httpx.BadRequest("status: pending, approved или rejected")
	}
	if limit <= 0 || limit > maxListLimit {
		limit = maxListLimit
	}
	rows, err := s.q.ListSubmissionsByStatus(ctx, store.ListSubmissionsByStatusParams{
		Status: status, Lim: limit, Off: max(offset, 0),
	})
	if err != nil {
		return nil, err
	}
	list := make([]Submission, len(rows))
	for i, r := range rows {
		list[i] = fromRow(store.GetSubmissionRow(r))
	}
	return list, s.fillObjects(ctx, list)
}

func fromRow(r store.GetSubmissionRow) Submission {
	sub := Submission{
		ID:          r.ID,
		AuthorID:    r.AuthorID,
		Kind:        r.Kind,
		ObjectID:    r.ObjectID,
		Title:       r.Title,
		Description: r.Description,
		Address:     r.Address,
		Period:      r.Period,
		Source:      r.Source,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt,
	}
	if r.HasLocation {
		sub.Coordinates = &catalog.Coordinates{Lat: r.Lat, Lng: r.Lng}
	}
	if r.Decision != nil {
		sub.Decision = &Decision{Decision: *r.Decision, ModeratorID: r.ModeratorID, CreatedAt: r.DecidedAt.Time}
		if r.Reason != nil {
			sub.Decision.Reason = *r.Reason
		}
	}
	return sub
}

// fillObjects дописывает slug и название объекта сразу для всего списка — одним запросом к каталогу.
func (s *Service) fillObjects(ctx context.Context, list []Submission) error {
	ids := make([]int64, 0, len(list))
	for _, sub := range list {
		if sub.ObjectID != nil {
			ids = append(ids, *sub.ObjectID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	places, err := s.catalog.PlacesByIDs(ctx, ids)
	if err != nil {
		return err
	}
	byID := make(map[int64]catalog.Place, len(places))
	for _, p := range places {
		byID[p.ID] = p
	}
	for i := range list {
		if list[i].ObjectID == nil {
			continue
		}
		if p, ok := byID[*list[i].ObjectID]; ok {
			list[i].ObjectSlug, list[i].ObjectTitle = p.Slug, p.Title
		}
	}
	return nil
}

// Moderator — модератор, принимающий решение.
type Moderator struct {
	UserID int64
	Role   string
}

// Approve одобряет заявку. Для новой заявки модератор присылает итоговые поля объекта
// (slug, категория, охранный статус и т. д.); незаполненные поля берутся из заявки (см. objectDraft).
// Для дополнения obj не нужен: файлы заявки переходят к объекту.
// Объект, статус заявки, решение, перенос файлов и письмо автору сохраняются в одной транзакции.
func (s *Service) Approve(ctx context.Context, id int64, mod Moderator, obj catalog.ObjectInput) (Submission, error) {
	err := s.decide(ctx, id, mod, func(tx pgx.Tx, sub store.GetSubmissionRow, author auth.User) error {
		var objectID int64
		var link string
		switch sub.Kind {
		case KindNewObject:
			draft := objectDraft(sub, obj)
			if draft.Coordinates == (catalog.Coordinates{}) {
				return httpx.BadRequest("в заявке нет координат — укажите их для объекта (coordinates)")
			}
			created, err := s.catalog.CreateTx(ctx, tx, draft)
			if err != nil {
				return err
			}
			objectID, link = created.ID, s.publicURL+"/objects/"+created.Slug
		default:
			if sub.ObjectID == nil {
				return httpx.Conflict("объект, к которому относится дополнение, удалён")
			}
			objectID = *sub.ObjectID
			link = s.publicURL + "/objects/"
			if places, err := s.catalog.PlacesByIDs(ctx, []int64{objectID}); err == nil && len(places) == 1 {
				link += places[0].Slug
			}
		}

		qtx := s.q.WithTx(tx)
		if err := qtx.SetSubmissionStatus(ctx, store.SetSubmissionStatusParams{
			ID: id, Status: StatusApproved, ObjectID: &objectID,
		}); err != nil {
			return err
		}
		if err := qtx.CreateDecision(ctx, store.CreateDecisionParams{
			SubmissionID: id, ModeratorID: &mod.UserID, Decision: StatusApproved,
		}); err != nil {
			return err
		}
		if err := s.media.AttachSubmissionTx(ctx, tx, id, objectID); err != nil {
			return err
		}
		return s.notify.EnqueueTx(ctx, tx, notifications.Message{
			To:      author.Email,
			Subject: "Ваша заявка одобрена",
			Body: fmt.Sprintf("Здравствуйте, %s!\n\nВаша заявка «%s» одобрена модератором. "+
				"Спасибо за вклад в сохранение наследия Астрахани!\n\nПосмотреть объект: %s\n",
				author.Name, sub.Title, link),
		})
	})
	if err != nil {
		return Submission{}, err
	}
	return s.Get(ctx, id, Viewer{UserID: mod.UserID, IsModerator: true})
}

// RejectInput — отказ: причина обязательна; blockAuthor — заодно заблокировать автора
// (например, за спам или заведомо ложные сведения).
type RejectInput struct {
	Reason      string `json:"reason"`
	BlockAuthor bool   `json:"blockAuthor"`
}

// Reject отклоняет заявку с причиной и сообщает автору по почте.
// Отказ, письмо и блокировка автора сохраняются в одной транзакции: либо всё, либо ничего.
func (s *Service) Reject(ctx context.Context, id int64, mod Moderator, in RejectInput) (Submission, error) {
	in.Reason = strings.TrimSpace(in.Reason)
	if in.Reason == "" {
		return Submission{}, httpx.BadRequest("укажите причину отказа")
	}
	err := s.decide(ctx, id, mod, func(tx pgx.Tx, sub store.GetSubmissionRow, author auth.User) error {
		if in.BlockAuthor {
			if _, err := s.users.SetBlockedTx(ctx, tx, mod.UserID, mod.Role, author.ID, true); err != nil {
				return err
			}
		}
		qtx := s.q.WithTx(tx)
		if err := qtx.SetSubmissionStatus(ctx, store.SetSubmissionStatusParams{ID: id, Status: StatusRejected}); err != nil {
			return err
		}
		if err := qtx.CreateDecision(ctx, store.CreateDecisionParams{
			SubmissionID: id, ModeratorID: &mod.UserID, Decision: StatusRejected, Reason: in.Reason,
		}); err != nil {
			return err
		}
		body := fmt.Sprintf("Здравствуйте, %s!\n\nВаша заявка «%s» отклонена модератором.\n\nПричина: %s\n",
			author.Name, sub.Title, in.Reason)
		if in.BlockAuthor {
			body += "\nКроме того, ваша учётная запись заблокирована: подавать новые заявки нельзя.\n"
		} else {
			body += fmt.Sprintf("\nВы можете исправить сведения и подать заявку снова: %s/contribute\n", s.publicURL)
		}
		return s.notify.EnqueueTx(ctx, tx, notifications.Message{To: author.Email, Subject: "Ваша заявка отклонена", Body: body})
	})
	if err != nil {
		return Submission{}, err
	}
	return s.Get(ctx, id, Viewer{UserID: mod.UserID, IsModerator: true})
}

// decide — общая часть одобрения и отказа: транзакция, блокировка строки заявки
// и проверка, что решение по ней ещё не принято. apply делает саму запись в той же транзакции.
func (s *Service) decide(ctx context.Context, id int64, mod Moderator,
	apply func(tx pgx.Tx, sub store.GetSubmissionRow, author auth.User) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }() // после Commit ничего не делает

	qtx := s.q.WithTx(tx)
	// SELECT ... FOR UPDATE: если два модератора решают одну заявку, второй ждёт
	// окончания транзакции первого и затем получает 409.
	locked, err := qtx.LockSubmission(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return httpx.NotFound("заявка не найдена")
	}
	if err != nil {
		return err
	}
	if locked.Status != StatusPending {
		return httpx.Conflict("по заявке уже принято решение")
	}
	if locked.AuthorID == mod.UserID {
		return httpx.Forbidden("нельзя принимать решение по своей заявке")
	}
	sub, err := qtx.GetSubmission(ctx, id)
	if err != nil {
		return err
	}
	author, err := s.users.GetByID(ctx, sub.AuthorID)
	if err != nil {
		return err
	}
	if err := apply(tx, sub, author); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// objectDraft собирает объект каталога из полей модератора; пустые поля заполняются из заявки:
// название, адрес, координаты, период (→ год постройки), описание (→ полная история,
// а его начало — краткое описание). Slug, категорию и охранный статус задаёт модератор.
func objectDraft(sub store.GetSubmissionRow, in catalog.ObjectInput) catalog.ObjectInput {
	if strings.TrimSpace(in.Title) == "" {
		in.Title = sub.Title
	}
	if strings.TrimSpace(in.Address) == "" {
		in.Address = sub.Address
	}
	if in.Coordinates == (catalog.Coordinates{}) && sub.HasLocation {
		in.Coordinates = catalog.Coordinates{Lat: sub.Lat, Lng: sub.Lng}
	}
	if in.YearBuilt == nil && sub.Period != "" {
		in.YearBuilt = &sub.Period
	}
	if strings.TrimSpace(in.FullHistory) == "" {
		in.FullHistory = sub.Description
	}
	if strings.TrimSpace(in.ShortDescription) == "" {
		in.ShortDescription = truncate(sub.Description, shortDescLen)
	}
	return in
}

// truncate обрезает текст до n символов (не байтов) и ставит многоточие.
func truncate(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n]) + "…"
}
