package submission

import (
	"context"
	"errors"
	"strings"
	"testing"

	"heritage-backend/internal/domain/errs"
	"heritage-backend/internal/domain/submission"
)

// Тесты сервиса на фейках: репозиторий, каталог, медиа, пользователи, события и транзакции — в памяти.

const (
	authorID = 1
	otherID  = 2
	modID    = 10
	mod2ID   = 11 // модератор, который сам подаёт заявки
)

var moderator = submission.Moderator{UserID: modID, Role: "moderator"}

type env struct {
	svc     *Service
	repo    *fakeRepo
	events  *fakePublisher
	catalog *fakeCatalog
	media   *fakeMedia
	users   *fakeUsers
}

func setup() env {
	repo, events := newFakeRepo(), &fakePublisher{}
	e := env{
		repo: repo, events: events, catalog: newFakeCatalog(), media: newFakeMedia(),
		users: &fakeUsers{users: map[int64]Author{
			authorID: {ID: authorID, Name: "Автор", Email: "author@example.ru", Role: "user"},
			otherID:  {ID: otherID, Name: "Другой", Email: "other@example.ru", Role: "user"},
			modID:    {ID: modID, Name: "Модератор", Email: "mod@example.ru", Role: "moderator"},
			mod2ID:   {ID: mod2ID, Name: "Модератор-автор", Email: "mod2@example.ru", Role: "moderator"},
		}},
	}
	e.svc = NewService(repo, &fakeTx{repo: repo, events: events}, events, e.catalog, e.media, e.users)
	return e
}

func wantKind(t *testing.T, err error, kind errs.Kind) {
	t.Helper()
	if !errs.Is(err, kind) {
		t.Fatalf("ожидалась ошибка рода %d, получено: %v", kind, err)
	}
}

func newObjectInput() submission.NewSubmission {
	return submission.NewSubmission{
		Kind:        submission.KindNewObject,
		Title:       "Дом купца Тестова",
		Description: "Двухэтажный деревянный дом с резными наличниками.",
		Address:     "ул. Тестовая, 1",
		Coordinates: &submission.Coordinates{Lat: 46.35, Lng: 48.04},
		Period:      "1890-е",
		Source:      "Личный архив",
	}
}

func photo() submission.MediaInput {
	return submission.MediaInput{Kind: "photo", OriginalName: "x.jpg", Body: strings.NewReader("x"), SourceNote: " архив "}
}

func (e env) submit(t *testing.T, author int64, in submission.NewSubmission) submission.Submission {
	t.Helper()
	sub, err := e.svc.Submit(context.Background(), author, in)
	if err != nil {
		t.Fatal(err)
	}
	return sub
}

func TestSubmitValidation(t *testing.T) {
	e := setup()
	ctx := context.Background()

	bad := newObjectInput()
	bad.Source = ""
	_, err := e.svc.Submit(ctx, authorID, bad)
	wantKind(t, err, errs.Invalid)

	// Достаточно адреса без координат.
	in := newObjectInput()
	in.Coordinates = nil
	sub := e.submit(t, authorID, in)
	if sub.Status != submission.StatusPending || sub.Coordinates != nil || sub.Kind != submission.KindNewObject ||
		sub.AuthorID != authorID || sub.Media == nil || sub.Author != nil {
		t.Fatalf("созданная заявка: %+v", sub)
	}

	// Дополнение — только к существующему опубликованному объекту.
	_, err = e.svc.Submit(ctx, authorID, submission.NewSubmission{Kind: submission.KindAddition, ObjectSlug: "no-such", Source: "Книга"})
	wantKind(t, err, errs.NotFound)
	_, err = e.svc.Submit(ctx, authorID, submission.NewSubmission{Kind: submission.KindAddition, ObjectSlug: "hidden", Source: "Книга"})
	wantKind(t, err, errs.NotFound)

	add := e.submit(t, authorID, submission.NewSubmission{Kind: submission.KindAddition, ObjectSlug: "Birzha", Source: "Книга", Description: "Фото 1950-х"})
	if add.ObjectID == nil || *add.ObjectID != 1 || add.ObjectSlug != "birzha" || add.ObjectTitle != "Биржа" ||
		add.Title != "Дополнение к объекту «Биржа»" {
		t.Fatalf("дополнение: %+v", add)
	}
}

func TestBlockedUserCannotSubmit(t *testing.T) {
	e := setup()
	e.users.users[authorID] = Author{ID: authorID, Role: "user", IsBlocked: true}
	_, err := e.svc.Submit(context.Background(), authorID, newObjectInput())
	if !errors.Is(err, errUserBlocked) {
		t.Fatalf("заблокированный автор: %v", err)
	}
}

func TestViewAndUploadRights(t *testing.T) {
	e := setup()
	ctx := context.Background()
	sub := e.submit(t, authorID, newObjectInput())

	// Чужую заявку пользователь не видит и не может пополнять — как будто её нет.
	_, err := e.svc.Get(ctx, sub.ID, submission.Viewer{UserID: otherID})
	wantKind(t, err, errs.NotFound)
	_, err = e.svc.AddMedia(ctx, sub.ID, otherID, photo())
	wantKind(t, err, errs.NotFound)
	_, err = e.svc.AddMedia(ctx, sub.ID+100, authorID, photo())
	wantKind(t, err, errs.NotFound)

	for i := 0; i < submission.MaxFiles; i++ {
		f, err := e.svc.AddMedia(ctx, sub.ID, authorID, photo())
		if err != nil {
			t.Fatal(err)
		}
		if f.SourceNote != "архив" {
			t.Fatalf("подписи файла не обрезаны: %+v", f)
		}
	}
	_, err = e.svc.AddMedia(ctx, sub.ID, authorID, photo())
	if !errors.Is(err, submission.ErrTooManyFiles) {
		t.Fatalf("11-й файл: %v", err)
	}

	mine, err := e.svc.Get(ctx, sub.ID, submission.Viewer{UserID: authorID})
	if err != nil {
		t.Fatal(err)
	}
	if len(mine.Media) != submission.MaxFiles || mine.Author != nil {
		t.Fatalf("автор видит: файлов %d, author=%+v", len(mine.Media), mine.Author)
	}
	asMod, err := e.svc.Get(ctx, sub.ID, submission.Viewer{UserID: modID, IsModerator: true})
	if err != nil {
		t.Fatal(err)
	}
	if asMod.Author == nil || asMod.Author.Email != "author@example.ru" {
		t.Fatalf("модератор должен видеть автора: %+v", asMod.Author)
	}

	// Заблокированный автор файлы не добавляет.
	sub2 := e.submit(t, authorID, newObjectInput())
	e.users.users[authorID] = Author{ID: authorID, Role: "user", IsBlocked: true}
	_, err = e.svc.AddMedia(ctx, sub2.ID, authorID, photo())
	wantKind(t, err, errs.Forbidden)
	e.users.users[authorID] = Author{ID: authorID, Role: "user"}

	// После решения файлы добавлять нельзя.
	if _, err := e.svc.Reject(ctx, sub2.ID, moderator, submission.RejectInput{Reason: "Дубликат"}); err != nil {
		t.Fatal(err)
	}
	_, err = e.svc.AddMedia(ctx, sub2.ID, authorID, photo())
	if !errors.Is(err, submission.ErrDecidedNoMedia) {
		t.Fatalf("файл после решения: %v", err)
	}
}

func TestQueueAndMyList(t *testing.T) {
	e := setup()
	ctx := context.Background()
	first := e.submit(t, authorID, newObjectInput())
	second := e.submit(t, authorID, submission.NewSubmission{Kind: submission.KindAddition, ObjectSlug: "birzha", Source: "Архив"})
	e.submit(t, otherID, newObjectInput())

	queue, err := e.svc.Queue(ctx, submission.StatusPending, 0, -5)
	if err != nil {
		t.Fatal(err)
	}
	if len(queue) != 3 || queue[0].ID != first.ID || queue[1].ObjectSlug != "birzha" || e.repo.lastLimit != maxListLimit {
		t.Fatalf("очередь: %+v (limit %d)", queue, e.repo.lastLimit)
	}
	if _, err := e.svc.Queue(ctx, submission.StatusPending, 1000, 0); err != nil || e.repo.lastLimit != maxListLimit {
		t.Fatalf("слишком большой limit: %d, %v", e.repo.lastLimit, err)
	}
	_, err = e.svc.Queue(ctx, "unknown", 10, 0)
	wantKind(t, err, errs.Invalid)

	mine, err := e.svc.ListMine(ctx, authorID)
	if err != nil {
		t.Fatal(err)
	}
	if len(mine) != 2 || mine[0].ID != second.ID || mine[1].ID != first.ID || mine[0].ObjectSlug != "birzha" {
		t.Fatalf("мои заявки (новые первыми): %+v", mine)
	}
}

func TestApproveNewObject(t *testing.T) {
	e := setup()
	ctx := context.Background()
	sub := e.submit(t, authorID, newObjectInput())
	if _, err := e.svc.AddMedia(ctx, sub.ID, authorID, photo()); err != nil {
		t.Fatal(err)
	}

	// Автор не может решать по своей заявке, даже будучи модератором.
	own := e.submit(t, mod2ID, newObjectInput())
	_, err := e.svc.Approve(ctx, own.ID, submission.Moderator{UserID: mod2ID, Role: "moderator"}, submission.ApproveInput{Slug: "own"})
	wantKind(t, err, errs.Forbidden)
	// Без slug каталог отклонит объект, и заявка останется ожидающей, а событие не запишется.
	_, err = e.svc.Approve(ctx, sub.ID, moderator, submission.ApproveInput{Category: "manor"})
	wantKind(t, err, errs.Invalid)
	if got, _ := e.repo.ByID(ctx, sub.ID); got.Status != submission.StatusPending || len(e.events.events) != 0 {
		t.Fatalf("после ошибки каталога: %s, событий %d", got.Status, len(e.events.events))
	}
	_, err = e.svc.Approve(ctx, sub.ID+100, moderator, submission.ApproveInput{Slug: "x"})
	wantKind(t, err, errs.NotFound)

	approved, err := e.svc.Approve(ctx, sub.ID, moderator, submission.ApproveInput{
		Slug: "dom-testova", Category: "manor", ProtectionStatus: "identified",
	})
	if err != nil {
		t.Fatal(err)
	}
	// Объект заполнен из заявки.
	if len(e.catalog.created) != 1 {
		t.Fatalf("создано объектов: %d", len(e.catalog.created))
	}
	obj := e.catalog.created[0]
	if obj.Slug != "dom-testova" || obj.Category != "manor" || obj.Title != sub.Title || obj.Address != sub.Address ||
		obj.Lat != 46.35 || obj.Lng != 48.04 || obj.FullHistory != sub.Description || obj.ShortDescription != sub.Description ||
		obj.YearBuilt == nil || *obj.YearBuilt != sub.Period || !obj.IsPublished {
		t.Fatalf("объект из заявки: %+v", obj)
	}
	created := e.catalog.objects["dom-testova"]
	if approved.Status != submission.StatusApproved || approved.ObjectID == nil || *approved.ObjectID != created.ID ||
		approved.ObjectSlug != "dom-testova" || approved.Decision == nil ||
		approved.Decision.Decision != submission.StatusApproved || *approved.Decision.ModeratorID != modID ||
		approved.Author == nil {
		t.Fatalf("одобренная заявка: %+v", approved)
	}
	if e.media.attached[sub.ID] != created.ID || len(approved.Media) != 1 || approved.Media[0].ObjectID == nil {
		t.Fatalf("файлы заявки не перешли объекту: %v, %+v", e.media.attached, approved.Media)
	}
	want := publishedEvent{submission.EventApproved, submission.ApprovedEvent{
		SubmissionID: sub.ID, AuthorID: authorID, Title: sub.Title, ObjectSlug: "dom-testova",
	}}
	if len(e.events.events) != 1 || e.events.events[0] != want {
		t.Fatalf("события: %+v", e.events.events)
	}

	// Повторное решение — конфликт, и второго события нет.
	_, err = e.svc.Reject(ctx, sub.ID, moderator, submission.RejectInput{Reason: "передумал"})
	if !errors.Is(err, submission.ErrAlreadyDecided) {
		t.Fatalf("повторное решение: %v", err)
	}
	_, err = e.svc.Approve(ctx, sub.ID, moderator, submission.ApproveInput{Slug: "again"})
	wantKind(t, err, errs.Conflict)
	if len(e.events.events) != 1 {
		t.Fatalf("событий после повторного решения: %d", len(e.events.events))
	}
}

func TestApproveNewObjectNeedsCoordinates(t *testing.T) {
	e := setup()
	in := newObjectInput()
	in.Coordinates = nil
	sub := e.submit(t, authorID, in)
	_, err := e.svc.Approve(context.Background(), sub.ID, moderator, submission.ApproveInput{Slug: "x", Category: "manor"})
	if !errors.Is(err, submission.ErrNoCoordinates) {
		t.Fatalf("без координат: %v", err)
	}
	// Координаты от модератора подходят.
	if _, err := e.svc.Approve(context.Background(), sub.ID, moderator, submission.ApproveInput{
		Slug: "x", Coordinates: &submission.Coordinates{Lat: 46, Lng: 48},
	}); err != nil {
		t.Fatal(err)
	}
}

func TestApproveAddition(t *testing.T) {
	e := setup()
	ctx := context.Background()
	sub := e.submit(t, authorID, submission.NewSubmission{Kind: submission.KindAddition, ObjectSlug: "birzha", Source: "Архив"})
	if _, err := e.svc.AddMedia(ctx, sub.ID, authorID, photo()); err != nil {
		t.Fatal(err)
	}

	got, err := e.svc.Approve(ctx, sub.ID, moderator, submission.ApproveInput{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != submission.StatusApproved || got.ObjectSlug != "birzha" || e.media.attached[sub.ID] != 1 ||
		len(e.catalog.created) != 0 {
		t.Fatalf("одобренное дополнение: %+v, attached=%v", got, e.media.attached)
	}
	want := publishedEvent{submission.EventApproved, submission.ApprovedEvent{
		SubmissionID: sub.ID, AuthorID: authorID, Title: sub.Title, ObjectSlug: "birzha",
	}}
	if len(e.events.events) != 1 || e.events.events[0] != want {
		t.Fatalf("события: %+v", e.events.events)
	}

	// Объект дополнения удалён (ON DELETE SET NULL) — одобрить нельзя.
	gone := e.submit(t, authorID, submission.NewSubmission{Kind: submission.KindAddition, ObjectSlug: "birzha", Source: "Архив"})
	s := e.repo.subs[gone.ID]
	s.ObjectID = nil
	e.repo.subs[gone.ID] = s
	_, err = e.svc.Approve(ctx, gone.ID, moderator, submission.ApproveInput{})
	if !errors.Is(err, submission.ErrObjectDeleted) {
		t.Fatalf("дополнение без объекта: %v", err)
	}
}

// При сбое внутри транзакции ничего не сохраняется и событие не публикуется.
func TestApproveRollsBackOnFailure(t *testing.T) {
	e := setup()
	ctx := context.Background()
	sub := e.submit(t, authorID, submission.NewSubmission{Kind: submission.KindAddition, ObjectSlug: "birzha", Source: "Архив"})

	boom := errors.New("хранилище недоступно")
	e.media.failAttach = boom
	if _, err := e.svc.Approve(ctx, sub.ID, moderator, submission.ApproveInput{}); !errors.Is(err, boom) {
		t.Fatalf("ожидалась ошибка переноса файлов: %v", err)
	}
	got, _ := e.repo.ByID(ctx, sub.ID)
	if got.Status != submission.StatusPending || got.Decision != nil || len(e.events.events) != 0 {
		t.Fatalf("после отката: %+v, событий %d", got, len(e.events.events))
	}

	// Сбой записи события тоже откатывает решение.
	e.media.failAttach = nil
	e.events.fail = boom
	if _, err := e.svc.Reject(ctx, sub.ID, moderator, submission.RejectInput{Reason: "x"}); !errors.Is(err, boom) {
		t.Fatalf("ожидалась ошибка события: %v", err)
	}
	if got, _ := e.repo.ByID(ctx, sub.ID); got.Status != submission.StatusPending || got.Decision != nil {
		t.Fatalf("после сбоя события: %+v", got)
	}
}

func TestRejectWithReasonAndBlock(t *testing.T) {
	e := setup()
	ctx := context.Background()
	sub := e.submit(t, authorID, newObjectInput())

	_, err := e.svc.Reject(ctx, sub.ID, moderator, submission.RejectInput{Reason: "   "})
	wantKind(t, err, errs.Invalid)
	_, err = e.svc.Reject(ctx, sub.ID+1_000_000, moderator, submission.RejectInput{Reason: "x"})
	wantKind(t, err, errs.NotFound)

	rejected, err := e.svc.Reject(ctx, sub.ID, moderator, submission.RejectInput{Reason: " Сведения не подтверждены ", BlockAuthor: true})
	if err != nil {
		t.Fatal(err)
	}
	if rejected.Status != submission.StatusRejected || rejected.Decision == nil || rejected.Decision.Reason != "Сведения не подтверждены" {
		t.Fatalf("отклонённая заявка: %+v", rejected)
	}
	if rejected.Author == nil || !rejected.Author.IsBlocked {
		t.Fatalf("автор должен быть заблокирован: %+v", rejected.Author)
	}
	want := publishedEvent{submission.EventRejected, submission.RejectedEvent{
		SubmissionID: sub.ID, AuthorID: authorID, Title: sub.Title, Reason: "Сведения не подтверждены", AuthorBlocked: true,
	}}
	if len(e.events.events) != 1 || e.events.events[0] != want {
		t.Fatalf("события: %+v", e.events.events)
	}

	// Автор видит причину отказа в своём списке, но подать новую заявку уже не может.
	mine, err := e.svc.ListMine(ctx, authorID)
	if err != nil {
		t.Fatal(err)
	}
	if len(mine) != 1 || mine[0].Decision == nil || mine[0].Decision.Reason != "Сведения не подтверждены" {
		t.Fatalf("мои заявки: %+v", mine)
	}
	_, err = e.svc.Submit(ctx, authorID, newObjectInput())
	wantKind(t, err, errs.Forbidden)

	_, err = e.svc.Approve(ctx, sub.ID, moderator, submission.ApproveInput{})
	wantKind(t, err, errs.Conflict)
	if len(e.events.events) != 1 {
		t.Fatalf("событий после повторного решения: %d", len(e.events.events))
	}
}

func TestRejectWithoutBlock(t *testing.T) {
	e := setup()
	sub := e.submit(t, authorID, newObjectInput())
	if _, err := e.svc.Reject(context.Background(), sub.ID, moderator, submission.RejectInput{Reason: "Дубликат"}); err != nil {
		t.Fatal(err)
	}
	ev := e.events.events[0].Payload.(submission.RejectedEvent)
	if ev.AuthorBlocked || e.users.users[authorID].IsBlocked {
		t.Fatalf("автор не должен блокироваться: %+v", ev)
	}
}

func TestRejectCannotBlockModerator(t *testing.T) {
	e := setup()
	ctx := context.Background()
	sub := e.submit(t, mod2ID, newObjectInput())

	// Модератор не может заблокировать другого модератора — и отказ тогда тоже не сохраняется.
	_, err := e.svc.Reject(ctx, sub.ID, moderator, submission.RejectInput{Reason: "x", BlockAuthor: true})
	wantKind(t, err, errs.Forbidden)
	got, err := e.svc.Get(ctx, sub.ID, submission.Viewer{UserID: mod2ID})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != submission.StatusPending || len(e.events.events) != 0 {
		t.Fatalf("заявка должна остаться ожидающей: %s, событий %d", got.Status, len(e.events.events))
	}
	if _, err := e.svc.Reject(ctx, sub.ID, moderator, submission.RejectInput{Reason: "тест"}); err != nil {
		t.Fatal(err)
	}
}
