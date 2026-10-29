// Package media — загрузка фото, аудио и 3D-моделей в MinIO и выдача их клиенту.
// В PostgreSQL хранятся только сведения о файле, сами байты — в объектном хранилище.
package media

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"heritage-backend/internal/httpx"
	"heritage-backend/internal/media/store"
)

type File struct {
	ID           int64     `json:"id"`
	ObjectID     *int64    `json:"objectId,omitempty"`
	SubmissionID *int64    `json:"submissionId,omitempty"`
	Kind         string    `json:"kind"`
	OriginalName string    `json:"originalName"`
	ContentType  string    `json:"contentType"`
	SizeBytes    int64     `json:"sizeBytes"`
	Caption      string    `json:"caption"`
	YearTaken    string    `json:"yearTaken"` // год съёмки
	Source       string    `json:"source"`    // источник: архив, книга, личный архив
	URL          string    `json:"url"`
	CreatedAt    time.Time `json:"createdAt"`
	storageKey   string
	uploadedBy   *int64
}

// IsPublic — файл прикреплён к объекту и виден всем. Файлы заявки на модерации
// видят только автор и модераторы.
func (f File) IsPublic() bool { return f.ObjectID != nil }

// CanView проверяет, может ли пользователь скачать файл, который ещё не опубликован.
func (f File) CanView(userID int64, isModerator bool) bool {
	return f.IsPublic() || isModerator || (f.uploadedBy != nil && *f.uploadedBy == userID)
}

// kindRule описывает, какие файлы принимаются для каждого вида медиа.
type kindRule struct {
	maxBytes int64
	allowed  []string // допустимые MIME-типы, определённые по содержимому файла
	ext      string   // расширение по умолчанию для ключа в хранилище
}

const mb = 1 << 20

var rules = map[string]kindRule{
	"photo":         {maxBytes: 15 * mb, allowed: []string{"image/jpeg", "image/png", "image/webp"}, ext: ".jpg"},
	"archive_photo": {maxBytes: 15 * mb, allowed: []string{"image/jpeg", "image/png", "image/webp"}, ext: ".jpg"},
	"audio":         {maxBytes: 40 * mb, allowed: []string{"audio/mpeg", "audio/ogg", "application/ogg", "audio/wave", "audio/mp4", "video/mp4"}, ext: ".mp3"},
	"model3d":       {maxBytes: 50 * mb, allowed: []string{"model/gltf-binary"}, ext: ".glb"},
}

// Objects — то, что медиафайлам нужно знать об объектах каталога.
// Интерфейс объявлен здесь, а реализует его catalog.Service: так media
// не читает таблицы каталога и не импортирует пакет catalog.
type Objects interface {
	Exists(ctx context.Context, id int64) (bool, error)
}

type Service struct {
	q       *store.Queries // запросы из queries.sql, код сгенерирован sqlc
	objects Objects
	store   *minio.Client
	bucket  string
	maxAll  int64
}

func NewService(ctx context.Context, db *pgxpool.Pool, objects Objects, endpoint, accessKey, secretKey, bucket string, useSSL bool, maxUploadMB int64) (*Service, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("клиент MinIO: %w", err)
	}

	// Бакет создаётся при первом запуске.
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("проверка бакета %q: %w", bucket, err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("создание бакета %q: %w", bucket, err)
		}
	}
	return &Service{q: store.New(db), objects: objects, store: client, bucket: bucket, maxAll: maxUploadMB * mb}, nil
}

// MaxUploadBytes — общий предел размера тела запроса при загрузке.
func (s *Service) MaxUploadBytes() int64 { return s.maxAll }

// UploadInput — файл к объекту (ObjectID) или к заявке (SubmissionID), заполняется ровно одно.
// Права на заявку проверяет модуль заявок до вызова Upload.
type UploadInput struct {
	ObjectID     int64
	SubmissionID int64
	Kind         string
	Caption      string
	YearTaken    string
	Source       string
	OriginalName string
	Size         int64
	Body         io.Reader
	UploadedBy   int64
}

func (s *Service) Upload(ctx context.Context, in UploadInput) (File, error) {
	rule, ok := rules[in.Kind]
	if !ok {
		return File{}, httpx.BadRequest("kind должен быть одним из: photo, archive_photo, audio, model3d")
	}
	if in.Size > rule.maxBytes || in.Size > s.maxAll {
		return File{}, httpx.BadRequest(fmt.Sprintf("файл больше допустимых %d МБ", min(rule.maxBytes, s.maxAll)/mb))
	}

	var objectID, submissionID *int64
	owner := fmt.Sprintf("object/%d", in.ObjectID)
	switch {
	case in.ObjectID != 0 && in.SubmissionID == 0:
		exists, err := s.objects.Exists(ctx, in.ObjectID)
		if err != nil {
			return File{}, err
		}
		if !exists {
			return File{}, httpx.NotFound("объект не найден")
		}
		objectID = &in.ObjectID
	case in.SubmissionID != 0 && in.ObjectID == 0:
		submissionID = &in.SubmissionID
		owner = fmt.Sprintf("submission/%d", in.SubmissionID)
	default:
		return File{}, fmt.Errorf("файл должен относиться либо к объекту, либо к заявке")
	}

	// Тип определяем по первым байтам файла, а не по расширению или заголовку от клиента:
	// переименованный .exe в .jpg так не пройдёт.
	br := bufio.NewReaderSize(in.Body, 512)
	head, _ := br.Peek(512)
	contentType := detectContentType(head)
	if !contains(rule.allowed, contentType) {
		return File{}, httpx.BadRequest(fmt.Sprintf("неподдерживаемый тип файла %q для %s", contentType, in.Kind))
	}

	ext := strings.ToLower(path.Ext(in.OriginalName))
	if ext == "" {
		ext = rule.ext
	}
	key := fmt.Sprintf("%s/%s/%s%s", in.Kind, owner, uuid.NewString(), ext)

	if _, err := s.store.PutObject(ctx, s.bucket, key, br, in.Size, minio.PutObjectOptions{ContentType: contentType}); err != nil {
		return File{}, fmt.Errorf("загрузка в MinIO: %w", err)
	}

	row, err := s.q.CreateMediaFile(ctx, store.CreateMediaFileParams{
		ObjectID:     objectID,
		SubmissionID: submissionID,
		Kind:         in.Kind,
		StorageKey:   key,
		OriginalName: in.OriginalName,
		ContentType:  contentType,
		SizeBytes:    in.Size,
		Caption:      in.Caption,
		YearTaken:    in.YearTaken,
		SourceNote:   in.Source, // после миграции 00004 свободный текст хранится в source_note
		UploadedBy:   &in.UploadedBy,
	})
	if err != nil {
		// Запись в БД не удалась — убираем уже загруженный файл, чтобы не оставлять «сирот».
		_ = s.store.RemoveObject(context.WithoutCancel(ctx), s.bucket, key, minio.RemoveObjectOptions{})
		return File{}, fmt.Errorf("сохранение сведений о файле: %w", err)
	}
	return toFile(row), nil
}

func (s *Service) ListByObject(ctx context.Context, objectID int64) ([]File, error) {
	rows, err := s.q.ListMediaFilesByObject(ctx, &objectID)
	if err != nil {
		return nil, err
	}
	files := make([]File, len(rows))
	for i, r := range rows {
		files[i] = toFile(r)
	}
	return files, nil
}

func (s *Service) ListBySubmission(ctx context.Context, submissionID int64) ([]File, error) {
	rows, err := s.q.ListMediaFilesBySubmission(ctx, &submissionID)
	if err != nil {
		return nil, err
	}
	files := make([]File, len(rows))
	for i, r := range rows {
		files[i] = toFile(r)
	}
	return files, nil
}

// AttachSubmissionTx передаёт файлы одобренной заявки объекту в транзакции модуля заявок.
func (s *Service) AttachSubmissionTx(ctx context.Context, tx pgx.Tx, submissionID, objectID int64) error {
	return s.q.WithTx(tx).AttachSubmissionMedia(ctx, store.AttachSubmissionMediaParams{
		ObjectID: &objectID, SubmissionID: &submissionID,
	})
}

func (s *Service) get(ctx context.Context, id int64) (File, error) {
	row, err := s.q.GetMediaFile(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return File{}, httpx.NotFound("файл не найден")
	}
	if err != nil {
		return File{}, err
	}
	return toFile(row), nil
}

// Open возвращает поток файла из MinIO. minio.Object умеет Seek,
// поэтому http.ServeContent отдаёт файл частями (Range) — это нужно для аудио и 3D.
// canView вызывается для неопубликованных файлов (файлов заявки); на «нет» отвечаем 404,
// чтобы по ответу нельзя было узнать, что такой файл существует.
func (s *Service) Open(ctx context.Context, id int64, canView func(File) bool) (File, *minio.Object, error) {
	f, err := s.get(ctx, id)
	if err != nil {
		return File{}, nil, err
	}
	if !f.IsPublic() && !canView(f) {
		return File{}, nil, httpx.NotFound("файл не найден")
	}
	obj, err := s.store.GetObject(ctx, s.bucket, f.storageKey, minio.GetObjectOptions{})
	if err != nil {
		return File{}, nil, fmt.Errorf("чтение из MinIO: %w", err)
	}
	return f, obj, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	f, err := s.get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.q.DeleteMediaFile(ctx, id); err != nil {
		return err
	}
	return s.store.RemoveObject(ctx, s.bucket, f.storageKey, minio.RemoveObjectOptions{})
}

// toFile переводит строку таблицы в ответ API. Ключ в хранилище наружу не отдаётся.
func toFile(r store.MediaFile) File {
	return File{
		ID:           r.ID,
		ObjectID:     r.ObjectID,
		SubmissionID: r.SubmissionID,
		Kind:         r.Kind,
		OriginalName: r.OriginalName,
		ContentType:  r.ContentType,
		SizeBytes:    r.SizeBytes,
		Caption:      r.Caption,
		YearTaken:    r.YearTaken,
		Source:       r.SourceNote,
		URL:          fileURL(r.ID),
		CreatedAt:    r.CreatedAt,
		storageKey:   r.StorageKey,
		uploadedBy:   r.UploadedBy,
	}
}

func fileURL(id int64) string { return fmt.Sprintf("/api/media/%d/file", id) }

// detectContentType дополняет стандартный http.DetectContentType:
// он не знает GLB, а у GLB файл всегда начинается с байтов «glTF».
func detectContentType(head []byte) string {
	if bytes.HasPrefix(head, []byte("glTF")) {
		return "model/gltf-binary"
	}
	return http.DetectContentType(head)
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}
