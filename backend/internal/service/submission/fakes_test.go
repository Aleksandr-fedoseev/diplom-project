package submission

import (
	"context"
	"errors"
	"maps"
	"sort"
	"sync"

	"heritage-backend/internal/domain/errs"
	"heritage-backend/internal/domain/submission"
	"heritage-backend/internal/infrastructure/postgres"
)

// Фейки портов для тестов сервиса: всё в памяти, без БД и HTTP.

var errNotInTx = errors.New("вызов должен идти внутри транзакции")

type fakeTxKey struct{}

// inTx — идёт ли транзакция: фейковая (fakeTx) или настоящая (postgres.TxManager в тесте гонки).
func inTx(ctx context.Context) bool {
	return ctx.Value(fakeTxKey{}) != nil || postgres.InTx(ctx)
}

// fakeTx вызывает fn и при ошибке «откатывает» состояние фейкового репозитория и
// записанные события — так проверяется, что при сбое ничего не сохраняется.
type fakeTx struct {
	repo   *fakeRepo
	events *fakePublisher
}

func (t *fakeTx) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	repoSnap := t.repo.snapshot()
	eventsLen := len(t.events.events)
	if err := fn(context.WithValue(ctx, fakeTxKey{}, true)); err != nil {
		t.repo.restore(repoSnap)
		t.events.events = t.events.events[:eventsLen]
		return err
	}
	return nil
}

type fakeRepo struct {
	mu        sync.Mutex
	subs      map[int64]submission.Submission
	nextID    int64
	lastLimit int
}

func newFakeRepo() *fakeRepo { return &fakeRepo{subs: map[int64]submission.Submission{}} }

func (r *fakeRepo) snapshot() map[int64]submission.Submission {
	r.mu.Lock()
	defer r.mu.Unlock()
	return maps.Clone(r.subs)
}

func (r *fakeRepo) restore(m map[int64]submission.Submission) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.subs = m
}

func (r *fakeRepo) Create(_ context.Context, s submission.Submission) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	s.ID = r.nextID
	s.Status = submission.StatusPending
	r.subs[s.ID] = s
	return s.ID, nil
}

func (r *fakeRepo) ByID(_ context.Context, id int64) (submission.Submission, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.subs[id]
	if !ok {
		return submission.Submission{}, submission.ErrNotFound
	}
	return s, nil
}

func (r *fakeRepo) LockForDecision(ctx context.Context, id int64) (submission.Submission, error) {
	if !inTx(ctx) {
		return submission.Submission{}, errNotInTx
	}
	return r.ByID(ctx, id)
}

func (r *fakeRepo) list(keep func(submission.Submission) bool, less func(a, b submission.Submission) bool) []submission.Submission {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []submission.Submission
	for _, s := range r.subs {
		if keep(s) {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return less(out[i], out[j]) })
	return out
}

func (r *fakeRepo) ListByAuthor(_ context.Context, authorID int64) ([]submission.Submission, error) {
	return r.list(func(s submission.Submission) bool { return s.AuthorID == authorID },
		func(a, b submission.Submission) bool { return a.ID > b.ID }), nil
}

func (r *fakeRepo) ListByStatus(_ context.Context, status submission.Status, limit, offset int) ([]submission.Submission, error) {
	r.lastLimit = limit
	out := r.list(func(s submission.Submission) bool { return s.Status == status },
		func(a, b submission.Submission) bool { return a.ID < b.ID })
	out = out[min(offset, len(out)):]
	return out[:min(limit, len(out))], nil
}

func (r *fakeRepo) SetStatus(ctx context.Context, id int64, status submission.Status, objectID *int64) error {
	if !inTx(ctx) {
		return errNotInTx
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.subs[id]
	if !ok {
		return submission.ErrNotFound
	}
	s.Status = status
	if objectID != nil {
		s.ObjectID = objectID
	}
	r.subs[id] = s
	return nil
}

func (r *fakeRepo) CreateDecision(ctx context.Context, id int64, d submission.Decision) error {
	if !inTx(ctx) {
		return errNotInTx
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.subs[id]
	if s.Decision != nil {
		return submission.ErrAlreadyDecided
	}
	s.Decision = &d
	r.subs[id] = s
	return nil
}

type publishedEvent struct {
	Type    string
	Payload any
}

type fakePublisher struct {
	mu     sync.Mutex
	events []publishedEvent
	fail   error
}

func (p *fakePublisher) Publish(ctx context.Context, eventType string, payload any) error {
	if !inTx(ctx) {
		return errNotInTx
	}
	if p.fail != nil {
		return p.fail
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.events = append(p.events, publishedEvent{eventType, payload})
	return nil
}

type fakeCatalog struct {
	mu      sync.Mutex
	objects map[string]ObjectRef
	created []NewObject
	nextID  int64
}

func newFakeCatalog() *fakeCatalog {
	return &fakeCatalog{nextID: 100, objects: map[string]ObjectRef{
		"birzha": {ID: 1, Slug: "birzha", Title: "Биржа", Published: true},
		"hidden": {ID: 2, Slug: "hidden", Title: "Скрытый", Published: false},
	}}
}

func (c *fakeCatalog) ObjectBySlug(_ context.Context, slug string) (ObjectRef, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	o, ok := c.objects[slug]
	if !ok {
		return ObjectRef{}, submission.ErrObjectNotFound
	}
	return o, nil
}

// CreateObject повторяет главные проверки каталога: slug обязателен и уникален.
func (c *fakeCatalog) CreateObject(ctx context.Context, in NewObject) (ObjectRef, error) {
	if !inTx(ctx) {
		return ObjectRef{}, errNotInTx
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if in.Slug == "" {
		return ObjectRef{}, errs.New(errs.Invalid, "укажите slug")
	}
	if _, ok := c.objects[in.Slug]; ok {
		return ObjectRef{}, errs.New(errs.Conflict, "slug занят")
	}
	c.nextID++
	o := ObjectRef{ID: c.nextID, Slug: in.Slug, Title: in.Title, Published: in.IsPublished}
	c.objects[in.Slug] = o
	c.created = append(c.created, in)
	return o, nil
}

func (c *fakeCatalog) ObjectsByIDs(_ context.Context, ids []int64) ([]ObjectRef, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []ObjectRef
	for _, o := range c.objects {
		for _, id := range ids {
			if o.ID == id {
				out = append(out, o)
			}
		}
	}
	return out, nil
}

type fakeMedia struct {
	mu         sync.Mutex
	files      []MediaFile
	attached   map[int64]int64 // заявка → объект
	failAttach error
}

func newFakeMedia() *fakeMedia { return &fakeMedia{attached: map[int64]int64{}} }

func (m *fakeMedia) Upload(_ context.Context, in MediaUpload) (MediaFile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	sid := in.SubmissionID
	f := MediaFile{ID: int64(len(m.files) + 1), SubmissionID: &sid, Kind: in.Kind, OriginalName: in.OriginalName,
		Caption: in.Caption, YearTaken: in.YearTaken, SourceNote: in.SourceNote, SizeBytes: in.Size}
	m.files = append(m.files, f)
	return f, nil
}

func (m *fakeMedia) ListBySubmission(_ context.Context, id int64) ([]MediaFile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []MediaFile
	for _, f := range m.files {
		if *f.SubmissionID == id {
			out = append(out, f)
		}
	}
	return out, nil
}

func (m *fakeMedia) AttachToObject(ctx context.Context, submissionID, objectID int64) error {
	if !inTx(ctx) {
		return errNotInTx
	}
	if m.failAttach != nil {
		return m.failAttach
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.attached[submissionID] = objectID
	for i := range m.files {
		if *m.files[i].SubmissionID == submissionID {
			oid := objectID
			m.files[i].ObjectID = &oid
		}
	}
	return nil
}

type fakeUsers struct {
	mu    sync.Mutex
	users map[int64]Author
}

var errUserBlocked = errs.New(errs.Forbidden, "пользователь заблокирован")

func (u *fakeUsers) User(_ context.Context, id int64) (Author, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	a, ok := u.users[id]
	if !ok {
		return Author{}, errs.New(errs.NotFound, "пользователь не найден")
	}
	return a, nil
}

func (u *fakeUsers) EnsureActive(ctx context.Context, id int64) error {
	a, err := u.User(ctx, id)
	if err != nil {
		return err
	}
	if a.IsBlocked {
		return errUserBlocked
	}
	return nil
}

// Block повторяет правило модуля пользователей: модератор блокирует только обычных пользователей.
func (u *fakeUsers) Block(ctx context.Context, _ int64, actorRole string, userID int64) error {
	if !inTx(ctx) {
		return errNotInTx
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	a, ok := u.users[userID]
	if !ok {
		return errs.New(errs.NotFound, "пользователь не найден")
	}
	if actorRole != "admin" && a.Role != "user" {
		return errs.New(errs.Forbidden, "модератор может блокировать только пользователей")
	}
	a.IsBlocked = true
	u.users[userID] = a
	return nil
}
