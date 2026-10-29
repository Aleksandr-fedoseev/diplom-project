// Package submission — домен заявок пользователей и их модерации.
// Пользователь предлагает новый объект или дополнение (фото, сведения) к существующему,
// модератор одобряет или отклоняет заявку. Модуль владеет таблицами submissions и
// moderation_decisions; объекты, файлы и пользователи принадлежат другим модулям.
// Только стандартная библиотека: ни БД, ни HTTP здесь нет.
package submission

import (
	"io"
	"strings"
	"time"
	"unicode/utf8"
)

// Kind — вид заявки.
type Kind string

const (
	KindNewObject Kind = "new_object" // предложение нового объекта
	KindAddition  Kind = "addition"   // фото и сведения к существующему объекту
)

// Status — состояние заявки. Решение принимается один раз: pending → approved или rejected.
type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)

func (s Status) Valid() bool {
	return s == StatusPending || s == StatusApproved || s == StatusRejected
}

const (
	MaxFiles     = 10  // файлов в одной заявке
	ShortDescLen = 300 // длина краткого описания объекта, если модератор его не задал
)

type Coordinates struct {
	Lat float64
	Lng float64
}

func (c Coordinates) Valid() bool {
	return c.Lat >= -90 && c.Lat <= 90 && c.Lng >= -180 && c.Lng <= 180
}

// Decision — решение модератора; по заявке оно одно.
type Decision struct {
	Decision    Status // approved или rejected
	Reason      string // причина отказа; при одобрении пустая
	ModeratorID *int64 // nil, если модератор удалён
	CreatedAt   time.Time
}

// Author — автор заявки, как его видит модератор. Данные приходят из модуля пользователей.
type Author struct {
	ID        int64
	Name      string
	Email     string
	Role      string
	IsBlocked bool
}

// File — файл заявки. Хранит его модуль медиа; здесь — только то, что показывается в карточке.
type File struct {
	ID           int64
	ObjectID     *int64 // заполняется после одобрения: файл перешёл к объекту
	SubmissionID *int64
	Kind         string
	OriginalName string
	ContentType  string
	SizeBytes    int64
	Caption      string
	YearTaken    string // год съёмки
	SourceNote   string // источник: архив, книга, личный архив
	URL          string
	CreatedAt    time.Time
}

type Submission struct {
	ID          int64
	AuthorID    int64
	Kind        Kind
	ObjectID    *int64 // для дополнения — объект; после одобрения новой заявки — созданный объект
	Title       string
	Description string
	Address     string
	Coordinates *Coordinates // nil, если автор указал только адрес
	Period      string
	Source      string
	Status      Status
	Decision    *Decision // nil, пока решения нет
	CreatedAt   time.Time

	// Поля ниже не хранятся в таблице заявок: сервис дописывает их из других модулей.
	ObjectSlug  string
	ObjectTitle string
	Author      *Author // только в карточке для модератора
	Media       []File  // только в карточке заявки
}

// Viewer — кто смотрит заявку.
type Viewer struct {
	UserID      int64
	IsModerator bool
}

// Moderator — модератор, принимающий решение. Роль нужна модулю пользователей
// для проверки прав при блокировке автора.
type Moderator struct {
	UserID int64
	Role   string
}

// VisibleTo — автор видит только свои заявки, модератор — любые.
func (s Submission) VisibleTo(v Viewer) bool {
	return v.IsModerator || s.AuthorID == v.UserID
}

// CanDecide — правило решения: только по ожидающей заявке и не по своей
// (даже если автор сам модератор).
func (s Submission) CanDecide(moderatorID int64) error {
	if s.Status != StatusPending {
		return ErrAlreadyDecided
	}
	if s.AuthorID == moderatorID {
		return ErrOwnSubmission
	}
	return nil
}

// CanAddMedia — файлы добавляет только автор и только пока заявка ожидает решения.
// Чужая заявка — «не найдена», как и при просмотре.
func (s Submission) CanAddMedia(userID int64) error {
	if s.AuthorID != userID {
		return ErrNotFound
	}
	if s.Status != StatusPending {
		return ErrDecidedNoMedia
	}
	return nil
}

// CheckFilesLimit — к заявке уже прикреплено filesCount файлов; можно ли добавить ещё один.
func CheckFilesLimit(filesCount int) error {
	if filesCount >= MaxFiles {
		return ErrTooManyFiles
	}
	return nil
}

// NewSubmission — заявка пользователя. Для дополнения объект указывается по slug
// (как в адресе карточки /objects/{slug}).
type NewSubmission struct {
	Kind        Kind
	ObjectSlug  string
	Title       string
	Description string
	Address     string
	Coordinates *Coordinates
	Period      string
	Source      string
}

// Normalize обрезает пробелы и проверяет правила: источник обязателен всегда;
// новому объекту нужны название, период и место (адрес или координаты), дополнению — objectSlug.
func (in *NewSubmission) Normalize() error {
	in.Kind = Kind(strings.TrimSpace(string(in.Kind)))
	in.ObjectSlug = strings.ToLower(strings.TrimSpace(in.ObjectSlug))
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	in.Address = strings.TrimSpace(in.Address)
	in.Period = strings.TrimSpace(in.Period)
	in.Source = strings.TrimSpace(in.Source)

	if in.Coordinates != nil && !in.Coordinates.Valid() {
		return ErrInvalidCoordinates
	}
	if in.Source == "" {
		return ErrSourceRequired
	}
	switch in.Kind {
	case KindNewObject:
		switch {
		case in.Title == "":
			return ErrTitleRequired
		case in.Address == "" && in.Coordinates == nil:
			return ErrPlaceRequired
		case in.Period == "":
			return ErrPeriodRequired
		}
	case KindAddition:
		if in.ObjectSlug == "" {
			return ErrObjectSlugRequired
		}
	default:
		return ErrInvalidKind
	}
	return nil
}

// AdditionTitle — название дополнения, если автор его не задал.
func AdditionTitle(objectTitle string) string {
	return "Дополнение к объекту «" + objectTitle + "»"
}

// MediaInput — файл к заявке; заполняется обработчиком из multipart-формы.
type MediaInput struct {
	Kind         string
	Caption      string
	YearTaken    string
	SourceNote   string
	OriginalName string
	Size         int64
	Body         io.Reader
}

// Normalize обрезает пробелы в подписях.
func (in *MediaInput) Normalize() {
	in.Caption = strings.TrimSpace(in.Caption)
	in.YearTaken = strings.TrimSpace(in.YearTaken)
	in.SourceNote = strings.TrimSpace(in.SourceNote)
}

// RejectInput — отказ: причина обязательна; BlockAuthor — заодно заблокировать автора
// (например, за спам или заведомо ложные сведения).
type RejectInput struct {
	Reason      string
	BlockAuthor bool
}

func (in *RejectInput) Normalize() error {
	in.Reason = strings.TrimSpace(in.Reason)
	if in.Reason == "" {
		return ErrReasonRequired
	}
	return nil
}

// ApproveInput — итоговые поля объекта каталога, которые модератор присылает при одобрении
// новой заявки. Для дополнения не нужны. Правильность slug, категории и охранного статуса
// проверяет каталог при создании объекта.
type ApproveInput struct {
	Slug               string
	Title              string
	Category           string
	ProtectionStatus   string
	RegistryNumber     *string
	YearBuilt          *string
	Architect          *string
	ArchitecturalStyle *string
	Address            string
	Coordinates        *Coordinates
	ShortDescription   string
	FullHistory        string
	Tags               []string
	IsPublished        *bool // nil — опубликовать сразу
}

// Prefill заполняет пустые поля объекта из заявки: название, адрес, координаты,
// период (→ год постройки), описание (→ полная история, а его начало — краткое описание).
// Slug, категорию и охранный статус задаёт модератор. Без координат объект на карту не поставить.
func (in ApproveInput) Prefill(s Submission) (ApproveInput, error) {
	if strings.TrimSpace(in.Title) == "" {
		in.Title = s.Title
	}
	if strings.TrimSpace(in.Address) == "" {
		in.Address = s.Address
	}
	// Нулевые координаты старый API тоже считал «не заданы».
	if in.Coordinates == nil || *in.Coordinates == (Coordinates{}) {
		in.Coordinates = s.Coordinates
	}
	if in.Coordinates == nil {
		return in, ErrNoCoordinates
	}
	if in.YearBuilt == nil && s.Period != "" {
		period := s.Period
		in.YearBuilt = &period
	}
	if strings.TrimSpace(in.FullHistory) == "" {
		in.FullHistory = s.Description
	}
	if strings.TrimSpace(in.ShortDescription) == "" {
		in.ShortDescription = truncate(s.Description, ShortDescLen)
	}
	if in.IsPublished == nil {
		published := true
		in.IsPublished = &published
	}
	return in, nil
}

// truncate обрезает текст до n символов (не байтов) и ставит многоточие.
func truncate(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n]) + "…"
}
