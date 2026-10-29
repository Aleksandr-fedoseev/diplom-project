package media

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"heritage-backend/internal/domain/errs"
	"heritage-backend/internal/domain/media"
)

// Фейки: хранилище файлов и БД в памяти, каталог из одного объекта.

type fakeRepo struct {
	files     map[int64]media.File
	failNext  bool // следующий Create вернёт ошибку
	attached  map[int64]int64
	deletedID int64
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{files: map[int64]media.File{}, attached: map[int64]int64{}}
}

func (r *fakeRepo) Create(_ context.Context, f media.File) (media.File, error) {
	if r.failNext {
		return media.File{}, errors.New("БД недоступна")
	}
	f.ID = int64(len(r.files) + 1)
	r.files[f.ID] = f
	return f, nil
}

func (r *fakeRepo) ByID(_ context.Context, id int64) (media.File, error) {
	f, ok := r.files[id]
	if !ok {
		return media.File{}, media.ErrFileNotFound
	}
	return f, nil
}

func (r *fakeRepo) ListByObject(context.Context, int64) ([]media.File, error)     { return nil, nil }
func (r *fakeRepo) ListBySubmission(context.Context, int64) ([]media.File, error) { return nil, nil }

func (r *fakeRepo) AttachToObject(_ context.Context, submissionID, objectID int64) error {
	r.attached[submissionID] = objectID
	return nil
}

func (r *fakeRepo) Delete(_ context.Context, id int64) error {
	delete(r.files, id)
	r.deletedID = id
	return nil
}

type fakeStorage struct {
	data map[string][]byte
	ct   map[string]string
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{data: map[string][]byte{}, ct: map[string]string{}}
}

func (s *fakeStorage) Put(_ context.Context, key string, body io.Reader, _ int64, contentType string) error {
	b, err := io.ReadAll(body)
	s.data[key], s.ct[key] = b, contentType
	return err
}

func (s *fakeStorage) Get(_ context.Context, key string) (io.ReadSeekCloser, error) {
	b, ok := s.data[key]
	if !ok {
		return nil, errors.New("нет такого ключа")
	}
	return nopCloser{bytes.NewReader(b)}, nil
}

func (s *fakeStorage) Remove(_ context.Context, key string) error {
	delete(s.data, key)
	return nil
}

type nopCloser struct{ io.ReadSeeker }

func (nopCloser) Close() error { return nil }

type fakeObjects map[int64]bool

func (o fakeObjects) Exists(_ context.Context, id int64) (bool, error) { return o[id], nil }

// Первые байты файлов разных форматов: по ним сервис определяет тип.
var (
	jpeg = "\xff\xd8\xff\xe0" + strings.Repeat("x", 100)
	png  = "\x89PNG\r\n\x1a\n" + strings.Repeat("x", 100)
	webp = "RIFF\x00\x00\x00\x00WEBPVP8 " + strings.Repeat("x", 100)
	glb  = "glTF" + strings.Repeat("x", 100)
)

type env struct {
	svc     *Service
	repo    *fakeRepo
	storage *fakeStorage
}

func newEnv() env {
	repo, storage := newFakeRepo(), newFakeStorage()
	return env{svc: NewService(repo, storage, fakeObjects{1: true}, 50), repo: repo, storage: storage}
}

func upload(kind media.Kind, body string) media.UploadInput {
	return media.UploadInput{ObjectID: 1, Kind: kind, OriginalName: "file", Size: int64(len(body)),
		Body: strings.NewReader(body), UploadedBy: 5}
}

func TestUploadDetectsTypeAndStores(t *testing.T) {
	e := newEnv()
	ctx := context.Background()

	in := upload(media.KindPanorama, webp)
	in.Projection, in.HAngle, in.Source = media.ProjectionPartial, new(int), media.SourceAIReconstruction
	*in.HAngle = 180
	in.OriginalName, in.Author, in.License = "Pano.WEBP", " Иванов ", "CC BY 4.0"
	f, err := e.svc.Upload(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	if f.ContentType != "image/webp" || f.Projection == nil || *f.Projection != media.ProjectionPartial ||
		*f.HAngle != 180 || f.Source != media.SourceAIReconstruction || f.Author != "Иванов" || f.License != "CC BY 4.0" {
		t.Fatalf("панорама: %+v", f)
	}
	if !strings.HasPrefix(f.StorageKey, "panorama/object/1/") || !strings.HasSuffix(f.StorageKey, ".webp") {
		t.Fatalf("ключ в хранилище: %s", f.StorageKey)
	}
	if string(e.storage.data[f.StorageKey]) != webp {
		t.Fatal("в хранилище не всё содержимое файла (первые байты потерялись при определении типа?)")
	}

	// Фото без источника считается настоящей фотографией; у проекции нет значения.
	f, err = e.svc.Upload(ctx, upload(media.KindPhoto, png))
	if err != nil || f.Source != media.SourcePhoto || f.Projection != nil || !strings.HasSuffix(f.StorageKey, ".jpg") {
		t.Fatalf("фото: %+v, %v", f, err)
	}
	if f, err = e.svc.Upload(ctx, upload(media.KindModel3D, glb)); err != nil || f.ContentType != "model/gltf-binary" {
		t.Fatalf("GLB: %+v, %v", f, err)
	}
}

func TestUploadRejects(t *testing.T) {
	e := newEnv()
	ctx := context.Background()

	// PNG как панорама не принимается, текст под видом фото — тоже.
	pano := upload(media.KindPanorama, png)
	pano.Projection = media.ProjectionFull360
	if _, err := e.svc.Upload(ctx, pano); !errs.Is(err, errs.Invalid) {
		t.Fatalf("PNG-панорама: %v", err)
	}
	if _, err := e.svc.Upload(ctx, upload(media.KindPhoto, "MZ это exe")); !errs.Is(err, errs.Invalid) {
		t.Fatalf("не изображение: %v", err)
	}
	if _, err := e.svc.Upload(ctx, upload(media.KindPanorama, jpeg)); !errors.Is(err, media.ErrProjectionRequired) {
		t.Fatalf("панорама без проекции: %v", err)
	}

	big := upload(media.KindPhoto, jpeg)
	big.Size = 16 * media.MB
	if _, err := e.svc.Upload(ctx, big); !errs.Is(err, errs.Invalid) {
		t.Fatalf("фото больше 15 МБ: %v", err)
	}

	missing := upload(media.KindPhoto, jpeg)
	missing.ObjectID = 404
	if _, err := e.svc.Upload(ctx, missing); !errors.Is(err, media.ErrObjectNotFound) {
		t.Fatalf("несуществующий объект: %v", err)
	}
	if len(e.storage.data) != 0 {
		t.Fatal("отклонённый файл попал в хранилище")
	}
}

// Если сведения о файле не сохранились, файл убирается из хранилища.
func TestUploadRemovesFileWhenDBFails(t *testing.T) {
	e := newEnv()
	e.repo.failNext = true
	if _, err := e.svc.Upload(context.Background(), upload(media.KindPhoto, jpeg)); err == nil {
		t.Fatal("ожидалась ошибка")
	}
	if len(e.storage.data) != 0 {
		t.Fatalf("в хранилище остался файл-сирота: %v", e.storage.data)
	}
}

func TestOpenChecksVisibility(t *testing.T) {
	e := newEnv()
	ctx := context.Background()

	in := upload(media.KindPhoto, jpeg)
	in.ObjectID, in.SubmissionID = 0, 3
	f, err := e.svc.Upload(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(f.StorageKey, "photo/submission/3/") {
		t.Fatalf("ключ файла заявки: %s", f.StorageKey)
	}

	if _, _, err := e.svc.Open(ctx, f.ID, media.Viewer{}); !errors.Is(err, media.ErrFileNotFound) {
		t.Fatalf("гость открыл файл заявки: %v", err)
	}
	if _, _, err := e.svc.Open(ctx, f.ID, media.Viewer{UserID: 6}); !errors.Is(err, media.ErrFileNotFound) {
		t.Fatalf("чужой пользователь открыл файл заявки: %v", err)
	}
	for _, v := range []media.Viewer{{UserID: 5}, {UserID: 9, Moderator: true}} {
		_, body, err := e.svc.Open(ctx, f.ID, v)
		if err != nil {
			t.Fatalf("%+v: %v", v, err)
		}
		b, _ := io.ReadAll(body)
		_ = body.Close()
		if string(b) != jpeg {
			t.Fatal("содержимое файла не совпало")
		}
	}

	// После передачи объекту (одобрение заявки) — файл виден всем; это решает репозиторий,
	// здесь проверяем только, что сервис передаёт вызов.
	if err := e.svc.AttachToObject(ctx, 3, 1); err != nil || e.repo.attached[3] != 1 {
		t.Fatalf("передача объекту: %v", err)
	}
	if _, _, err := e.svc.Open(ctx, 999, media.Viewer{Moderator: true}); !errors.Is(err, media.ErrFileNotFound) {
		t.Fatalf("несуществующий файл: %v", err)
	}
}

func TestDeleteRemovesFromStorage(t *testing.T) {
	e := newEnv()
	ctx := context.Background()
	f, err := e.svc.Upload(ctx, upload(media.KindPhoto, jpeg))
	if err != nil {
		t.Fatal(err)
	}
	if err := e.svc.Delete(ctx, f.ID); err != nil {
		t.Fatal(err)
	}
	if e.repo.deletedID != f.ID || len(e.storage.data) != 0 {
		t.Fatal("файл не удалён из БД или хранилища")
	}
	if err := e.svc.Delete(ctx, f.ID); !errors.Is(err, media.ErrFileNotFound) {
		t.Fatalf("повторное удаление: %v", err)
	}
}
