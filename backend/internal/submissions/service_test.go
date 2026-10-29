package submissions

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"heritage-backend/internal/auth"
	"heritage-backend/internal/catalog"
	"heritage-backend/internal/httpx"
	"heritage-backend/internal/media"
	"heritage-backend/internal/notifications"
	"heritage-backend/internal/testdb"
)

// fakeMedia хранит сведения о файлах в памяти: MinIO в тестах недоступен,
// а модулю заявок важно лишь, что файлы привязываются к заявке и переходят к объекту.
type fakeMedia struct {
	mu       sync.Mutex
	files    []media.File
	attached map[int64]int64 // заявка → объект
}

func newFakeMedia() *fakeMedia { return &fakeMedia{attached: map[int64]int64{}} }

func (f *fakeMedia) Upload(_ context.Context, in media.UploadInput) (media.File, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	sid := in.SubmissionID
	file := media.File{ID: int64(len(f.files) + 1), SubmissionID: &sid, Kind: in.Kind, OriginalName: in.OriginalName,
		Caption: in.Caption, YearTaken: in.YearTaken, Source: in.Source, SizeBytes: in.Size}
	f.files = append(f.files, file)
	return file, nil
}

func (f *fakeMedia) ListBySubmission(_ context.Context, id int64) ([]media.File, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var list []media.File
	for _, file := range f.files {
		if *file.SubmissionID == id {
			list = append(list, file)
		}
	}
	return list, nil
}

func (f *fakeMedia) AttachSubmissionTx(_ context.Context, tx pgx.Tx, submissionID, objectID int64) error {
	if tx == nil {
		return errors.New("перенос файлов должен идти в транзакции")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.attached[submissionID] = objectID
	for i := range f.files {
		if *f.files[i].SubmissionID == submissionID {
			oid := objectID
			f.files[i].ObjectID = &oid
		}
	}
	return nil
}

func (f *fakeMedia) MaxUploadBytes() int64 { return 100 << 20 }

func wantStatus(t *testing.T, err error, status int) {
	t.Helper()
	var he *httpx.Error
	if !errors.As(err, &he) || he.Status != status {
		t.Fatalf("ожидалась ошибка %d, получено: %v", status, err)
	}
}

type env struct {
	db    *pgxpool.Pool
	svc   *Service
	media *fakeMedia
	cat   *catalog.Service
	users *auth.Service
	mod   Moderator
}

func setup(t *testing.T) env {
	t.Helper()
	db := testdb.Connect(t)
	cat := catalog.NewService(db)
	users := auth.NewService(db, auth.NewTokenIssuer("0123456789abcdef0123456789abcdef", time.Hour))
	fm := newFakeMedia()
	svc := NewService(db, cat, fm, users, notifications.NewService(db, notifications.LogSender{}), "https://heritage.test/")

	e := env{db: db, svc: svc, media: fm, cat: cat, users: users}
	m := e.register(t, "moderator")
	if _, err := users.SetRole(context.Background(), 0, m.ID, auth.RoleModerator); err != nil {
		t.Fatal(err)
	}
	e.mod = Moderator{UserID: m.ID, Role: auth.RoleModerator}
	return e
}

func (e env) register(t *testing.T, prefix string) auth.User {
	t.Helper()
	u, err := e.users.Register(context.Background(), auth.RegisterInput{
		Email: testdb.Unique(prefix) + "@example.ru", Password: "password1", Name: "Тест " + prefix,
	})
	if err != nil {
		t.Fatal(err)
	}
	return u
}

// emails возвращает письма из очереди для адреса. Очередь общая с другими тестами,
// поэтому ProcessDue не вызываем, а смотрим только задачи своего получателя.
func (e env) emails(t *testing.T, to string) []notifications.Message {
	t.Helper()
	rows, err := e.db.Query(context.Background(), `SELECT subject, body FROM email_jobs WHERE to_email = $1 ORDER BY id`, to)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var list []notifications.Message
	for rows.Next() {
		m := notifications.Message{To: to}
		if err := rows.Scan(&m.Subject, &m.Body); err != nil {
			t.Fatal(err)
		}
		list = append(list, m)
	}
	return list
}

func newObjectInput() CreateInput {
	return CreateInput{
		Kind:        KindNewObject,
		Title:       "Дом купца Тестова",
		Description: "Двухэтажный деревянный дом с резными наличниками.",
		Address:     "ул. Тестовая, 1",
		Coordinates: &catalog.Coordinates{Lat: 46.35, Lng: 48.04},
		Period:      "1890-е",
		Source:      "Личный архив",
	}
}

func TestCreateValidation(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	u := e.register(t, "author")

	cases := map[string]func(*CreateInput){
		"неизвестный вид":   func(in *CreateInput) { in.Kind = "other" },
		"без источника":     func(in *CreateInput) { in.Source = "  " },
		"без названия":      func(in *CreateInput) { in.Title = "" },
		"без места":         func(in *CreateInput) { in.Address, in.Coordinates = "", nil },
		"без периода":       func(in *CreateInput) { in.Period = "" },
		"плохие координаты": func(in *CreateInput) { in.Coordinates = &catalog.Coordinates{Lat: 100, Lng: 48} },
		"дополнение без объекта": func(in *CreateInput) {
			in.Kind = KindAddition
		},
	}
	for name, mutate := range cases {
		in := newObjectInput()
		mutate(&in)
		_, err := e.svc.Create(ctx, u.ID, in)
		if err == nil {
			t.Fatalf("%s: заявка должна отклоняться", name)
		}
		wantStatus(t, err, 400)
	}

	// Достаточно адреса без координат.
	in := newObjectInput()
	in.Coordinates = nil
	sub, err := e.svc.Create(ctx, u.ID, in)
	if err != nil {
		t.Fatal(err)
	}
	if sub.Status != StatusPending || sub.Coordinates != nil || sub.Kind != KindNewObject || sub.AuthorID != u.ID {
		t.Fatalf("созданная заявка: %+v", sub)
	}

	// Дополнение — только к существующему объекту.
	_, err = e.svc.Create(ctx, u.ID, CreateInput{Kind: KindAddition, ObjectSlug: "no-such-object", Source: "Книга"})
	wantStatus(t, err, 404)
	add, err := e.svc.Create(ctx, u.ID, CreateInput{Kind: KindAddition, ObjectSlug: "birzha", Source: "Книга", Description: "Фото 1950-х"})
	if err != nil {
		t.Fatal(err)
	}
	if add.ObjectID == nil || add.ObjectSlug != "birzha" || !strings.Contains(add.Title, "Дополнение") {
		t.Fatalf("дополнение: %+v", add)
	}
}

func TestBlockedUserCannotSubmit(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	u := e.register(t, "blocked")
	if _, err := e.users.SetBlocked(ctx, e.mod.UserID, e.mod.Role, u.ID, true); err != nil {
		t.Fatal(err)
	}
	_, err := e.svc.Create(ctx, u.ID, newObjectInput())
	wantStatus(t, err, 403)
}

func TestViewAndUploadRights(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	author, other := e.register(t, "author"), e.register(t, "other")
	sub, err := e.svc.Create(ctx, author.ID, newObjectInput())
	if err != nil {
		t.Fatal(err)
	}

	// Чужую заявку пользователь не видит и не может пополнять — 404, как будто её нет.
	_, err = e.svc.Get(ctx, sub.ID, Viewer{UserID: other.ID})
	wantStatus(t, err, 404)
	_, err = e.svc.UploadMedia(ctx, sub.ID, other.ID, UploadInput{Kind: "photo", OriginalName: "x.jpg", Body: strings.NewReader("x")})
	wantStatus(t, err, 404)

	for i := 0; i < maxFiles; i++ {
		if _, err := e.svc.UploadMedia(ctx, sub.ID, author.ID, UploadInput{Kind: "photo", OriginalName: "x.jpg", Body: strings.NewReader("x")}); err != nil {
			t.Fatal(err)
		}
	}
	_, err = e.svc.UploadMedia(ctx, sub.ID, author.ID, UploadInput{Kind: "photo", OriginalName: "x.jpg", Body: strings.NewReader("x")})
	wantStatus(t, err, 409)

	mine, err := e.svc.Get(ctx, sub.ID, Viewer{UserID: author.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(mine.Media) != maxFiles || mine.Author != nil {
		t.Fatalf("автор видит: файлов %d, author=%+v", len(mine.Media), mine.Author)
	}
	asMod, err := e.svc.Get(ctx, sub.ID, Viewer{UserID: e.mod.UserID, IsModerator: true})
	if err != nil {
		t.Fatal(err)
	}
	if asMod.Author == nil || asMod.Author.Email != author.Email {
		t.Fatalf("модератор должен видеть автора: %+v", asMod.Author)
	}

	// После решения файлы добавлять нельзя.
	if _, err := e.svc.Reject(ctx, sub.ID, e.mod, RejectInput{Reason: "Дубликат"}); err != nil {
		t.Fatal(err)
	}
	_, err = e.svc.UploadMedia(ctx, sub.ID, author.ID, UploadInput{Kind: "photo", OriginalName: "x.jpg", Body: strings.NewReader("x")})
	wantStatus(t, err, 409)
}

// pendingPositions проходит всю очередь по страницам и возвращает позиции нужных заявок.
func pendingPositions(t *testing.T, svc *Service, ids ...int64) []int {
	t.Helper()
	pos := make([]int, len(ids))
	for i := range pos {
		pos[i] = -1
	}
	n := 0
	for offset := int32(0); ; offset += maxListLimit {
		page, err := svc.ListByStatus(context.Background(), StatusPending, maxListLimit, offset)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range page {
			for i, id := range ids {
				if s.ID == id {
					pos[i] = n
				}
			}
			n++
		}
		if len(page) < maxListLimit {
			return pos
		}
	}
}

func TestQueueOrderAndMyList(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	u := e.register(t, "queue")
	first, err := e.svc.Create(ctx, u.ID, newObjectInput())
	if err != nil {
		t.Fatal(err)
	}
	second, err := e.svc.Create(ctx, u.ID, CreateInput{Kind: KindAddition, ObjectSlug: "birzha", Source: "Архив"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { // не оставляем заявки в общей очереди
		_, _ = e.svc.Reject(ctx, first.ID, e.mod, RejectInput{Reason: "тест"})
		_, _ = e.svc.Reject(ctx, second.ID, e.mod, RejectInput{Reason: "тест"})
	})

	pos := pendingPositions(t, e.svc, first.ID, second.ID)
	if pos[0] < 0 || pos[1] < 0 || pos[0] > pos[1] {
		t.Fatalf("в очереди старые заявки должны идти первыми: позиции %v", pos)
	}
	_, err = e.svc.ListByStatus(ctx, "unknown", 10, 0)
	wantStatus(t, err, 400)

	mine, err := e.svc.ListByAuthor(ctx, u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(mine) != 2 || mine[0].ID != second.ID || mine[1].ID != first.ID || mine[0].ObjectSlug != "birzha" {
		t.Fatalf("мои заявки (новые первыми): %+v", mine)
	}
}

func TestApproveNewObject(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	author := e.register(t, "newobj")
	sub, err := e.svc.Create(ctx, author.ID, newObjectInput())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.svc.UploadMedia(ctx, sub.ID, author.ID, UploadInput{Kind: "photo", OriginalName: "f.jpg", Body: strings.NewReader("x")}); err != nil {
		t.Fatal(err)
	}

	// Автор не может решать по своей заявке, даже будучи модератором.
	_, err = e.svc.Approve(ctx, sub.ID, Moderator{UserID: author.ID, Role: auth.RoleModerator}, catalog.ObjectInput{})
	wantStatus(t, err, 403)
	// Без slug каталог отклонит объект, и заявка останется ожидающей.
	_, err = e.svc.Approve(ctx, sub.ID, e.mod, catalog.ObjectInput{Category: "manor", ProtectionStatus: "identified"})
	wantStatus(t, err, 400)

	slug := testdb.Unique("test-submission-obj")
	approved, err := e.svc.Approve(ctx, sub.ID, e.mod, catalog.ObjectInput{
		Slug: slug, Category: "manor", ProtectionStatus: "identified",
	})
	if err != nil {
		t.Fatal(err)
	}
	obj, err := e.cat.GetBySlug(ctx, slug)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = e.cat.Delete(ctx, obj.ID) })

	if !obj.IsPublished || obj.Title != sub.Title || obj.Address != sub.Address ||
		obj.Coordinates != *sub.Coordinates || obj.FullHistory != sub.Description ||
		obj.YearBuilt == nil || *obj.YearBuilt != sub.Period {
		t.Fatalf("объект из заявки: %+v", obj)
	}
	if approved.Status != StatusApproved || approved.ObjectID == nil || *approved.ObjectID != obj.ID ||
		approved.ObjectSlug != slug || approved.Decision == nil || approved.Decision.Decision != StatusApproved {
		t.Fatalf("одобренная заявка: %+v", approved)
	}
	if e.media.attached[sub.ID] != obj.ID || len(approved.Media) != 1 || approved.Media[0].ObjectID == nil {
		t.Fatalf("файлы заявки не перешли объекту: %v, %+v", e.media.attached, approved.Media)
	}
	mails := e.emails(t, author.Email)
	if len(mails) != 1 || mails[0].Subject != "Ваша заявка одобрена" ||
		!strings.Contains(mails[0].Body, "https://heritage.test/objects/"+slug) {
		t.Fatalf("письмо автору: %+v", mails)
	}

	// Повторное решение — 409, и второго письма нет.
	_, err = e.svc.Reject(ctx, sub.ID, e.mod, RejectInput{Reason: "передумал"})
	wantStatus(t, err, 409)
	if n := len(e.emails(t, author.Email)); n != 1 {
		t.Fatalf("писем после повторного решения: %d", n)
	}
}

func TestApproveNewObjectNeedsCoordinates(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	author := e.register(t, "nocoords")
	in := newObjectInput()
	in.Coordinates = nil
	sub, err := e.svc.Create(ctx, author.ID, in)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = e.svc.Reject(ctx, sub.ID, e.mod, RejectInput{Reason: "тест"}) })
	_, err = e.svc.Approve(ctx, sub.ID, e.mod, catalog.ObjectInput{
		Slug: testdb.Unique("test-nocoords"), Category: "manor", ProtectionStatus: "identified",
	})
	wantStatus(t, err, 400)
}

func TestApproveAdditionConcurrently(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	author := e.register(t, "addition")
	sub, err := e.svc.Create(ctx, author.ID, CreateInput{Kind: KindAddition, ObjectSlug: "birzha", Source: "Архив", Description: "Фото 1950-х"})
	if err != nil {
		t.Fatal(err)
	}

	// Два модератора одобряют одновременно: проходит ровно одно решение, второе получает 409.
	errs := make([]error, 2)
	var wg sync.WaitGroup
	for i := range errs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = e.svc.Approve(ctx, sub.ID, e.mod, catalog.ObjectInput{})
		}()
	}
	wg.Wait()
	ok := 0
	for _, err := range errs {
		if err == nil {
			ok++
		} else {
			wantStatus(t, err, 409)
		}
	}
	if ok != 1 {
		t.Fatalf("одобрений прошло %d, ожидалось 1: %v", ok, errs)
	}

	got, err := e.svc.Get(ctx, sub.ID, Viewer{UserID: author.ID})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusApproved || got.ObjectSlug != "birzha" || e.media.attached[sub.ID] != *got.ObjectID {
		t.Fatalf("одобренное дополнение: %+v, attached=%v", got, e.media.attached)
	}
	mails := e.emails(t, author.Email)
	if len(mails) != 1 || !strings.Contains(mails[0].Body, "/objects/birzha") {
		t.Fatalf("письмо автору: %+v", mails)
	}
}

func TestRejectWithReasonAndBlock(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	author := e.register(t, "rejected")
	sub, err := e.svc.Create(ctx, author.ID, newObjectInput())
	if err != nil {
		t.Fatal(err)
	}

	_, err = e.svc.Reject(ctx, sub.ID, e.mod, RejectInput{Reason: "   "})
	wantStatus(t, err, 400)
	_, err = e.svc.Reject(ctx, sub.ID+1_000_000, e.mod, RejectInput{Reason: "x"})
	wantStatus(t, err, 404)

	rejected, err := e.svc.Reject(ctx, sub.ID, e.mod, RejectInput{Reason: "Сведения не подтверждены", BlockAuthor: true})
	if err != nil {
		t.Fatal(err)
	}
	if rejected.Status != StatusRejected || rejected.Decision == nil || rejected.Decision.Reason != "Сведения не подтверждены" {
		t.Fatalf("отклонённая заявка: %+v", rejected)
	}
	if rejected.Author == nil || !rejected.Author.IsBlocked {
		t.Fatalf("автор должен быть заблокирован: %+v", rejected.Author)
	}
	mails := e.emails(t, author.Email)
	if len(mails) != 1 || mails[0].Subject != "Ваша заявка отклонена" ||
		!strings.Contains(mails[0].Body, "Сведения не подтверждены") || !strings.Contains(mails[0].Body, "заблокирована") {
		t.Fatalf("письмо автору: %+v", mails)
	}

	// Автор видит причину отказа в своём списке, но подать новую заявку уже не может.
	mine, err := e.svc.ListByAuthor(ctx, author.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(mine) != 1 || mine[0].Decision == nil || mine[0].Decision.Reason != "Сведения не подтверждены" {
		t.Fatalf("мои заявки: %+v", mine)
	}
	_, err = e.svc.Create(ctx, author.ID, newObjectInput())
	wantStatus(t, err, 403)

	_, err = e.svc.Approve(ctx, sub.ID, e.mod, catalog.ObjectInput{})
	wantStatus(t, err, 409)
}

func TestRejectCannotBlockModerator(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	author := e.register(t, "modauthor")
	if _, err := e.users.SetRole(ctx, 0, author.ID, auth.RoleModerator); err != nil {
		t.Fatal(err)
	}
	sub, err := e.svc.Create(ctx, author.ID, newObjectInput())
	if err != nil {
		t.Fatal(err)
	}
	// Модератор не может заблокировать другого модератора — и отказ тогда тоже не сохраняется.
	_, err = e.svc.Reject(ctx, sub.ID, e.mod, RejectInput{Reason: "x", BlockAuthor: true})
	wantStatus(t, err, 403)
	got, err := e.svc.Get(ctx, sub.ID, Viewer{UserID: author.ID})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusPending {
		t.Fatalf("заявка должна остаться ожидающей: %s", got.Status)
	}
	if _, err := e.svc.Reject(ctx, sub.ID, e.mod, RejectInput{Reason: "тест"}); err != nil {
		t.Fatal(err)
	}
}
