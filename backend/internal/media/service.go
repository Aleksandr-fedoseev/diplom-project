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
)

type File struct {
	ID           int64     `json:"id"`
	ObjectID     *int64    `json:"objectId,omitempty"`
	Kind         string    `json:"kind"`
	OriginalName string    `json:"originalName"`
	ContentType  string    `json:"contentType"`
	SizeBytes    int64     `json:"sizeBytes"`
	Caption      string    `json:"caption"`
	URL          string    `json:"url"`
	CreatedAt    time.Time `json:"createdAt"`
	storageKey   string
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

type Service struct {
	db     *pgxpool.Pool
	store  *minio.Client
	bucket string
	maxAll int64
}

func NewService(ctx context.Context, db *pgxpool.Pool, endpoint, accessKey, secretKey, bucket string, useSSL bool, maxUploadMB int64) (*Service, error) {
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
	return &Service{db: db, store: client, bucket: bucket, maxAll: maxUploadMB * mb}, nil
}

// MaxUploadBytes — общий предел размера тела запроса при загрузке.
func (s *Service) MaxUploadBytes() int64 { return s.maxAll }

type UploadInput struct {
	ObjectID     int64
	Kind         string
	Caption      string
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

	var exists bool
	if err := s.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM heritage_objects WHERE id = $1)`, in.ObjectID).Scan(&exists); err != nil {
		return File{}, err
	}
	if !exists {
		return File{}, httpx.NotFound("объект не найден")
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
	key := fmt.Sprintf("%s/%d/%s%s", in.Kind, in.ObjectID, uuid.NewString(), ext)

	if _, err := s.store.PutObject(ctx, s.bucket, key, br, in.Size, minio.PutObjectOptions{ContentType: contentType}); err != nil {
		return File{}, fmt.Errorf("загрузка в MinIO: %w", err)
	}

	f := File{storageKey: key}
	err := s.db.QueryRow(ctx, `
		INSERT INTO media_files (object_id, kind, storage_key, original_name, content_type, size_bytes, caption, uploaded_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, object_id, kind, original_name, content_type, size_bytes, caption, created_at`,
		in.ObjectID, in.Kind, key, in.OriginalName, contentType, in.Size, in.Caption, in.UploadedBy,
	).Scan(&f.ID, &f.ObjectID, &f.Kind, &f.OriginalName, &f.ContentType, &f.SizeBytes, &f.Caption, &f.CreatedAt)
	if err != nil {
		// Запись в БД не удалась — убираем уже загруженный файл, чтобы не оставлять «сирот».
		_ = s.store.RemoveObject(context.WithoutCancel(ctx), s.bucket, key, minio.RemoveObjectOptions{})
		return File{}, fmt.Errorf("сохранение сведений о файле: %w", err)
	}
	f.URL = fileURL(f.ID)
	return f, nil
}

func (s *Service) ListByObject(ctx context.Context, objectID int64) ([]File, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, object_id, kind, original_name, content_type, size_bytes, caption, created_at, storage_key
		FROM media_files WHERE object_id = $1 ORDER BY kind, id`, objectID)
	if err != nil {
		return nil, err
	}
	files, err := pgx.CollectRows(rows, scanFile)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (s *Service) get(ctx context.Context, id int64) (File, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, object_id, kind, original_name, content_type, size_bytes, caption, created_at, storage_key
		FROM media_files WHERE id = $1`, id)
	if err != nil {
		return File{}, err
	}
	f, err := pgx.CollectExactlyOneRow(rows, scanFile)
	if errors.Is(err, pgx.ErrNoRows) {
		return File{}, httpx.NotFound("файл не найден")
	}
	return f, err
}

// Open возвращает поток файла из MinIO. minio.Object умеет Seek,
// поэтому http.ServeContent отдаёт файл частями (Range) — это нужно для аудио и 3D.
func (s *Service) Open(ctx context.Context, id int64) (File, *minio.Object, error) {
	f, err := s.get(ctx, id)
	if err != nil {
		return File{}, nil, err
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
	if _, err := s.db.Exec(ctx, `DELETE FROM media_files WHERE id = $1`, id); err != nil {
		return err
	}
	return s.store.RemoveObject(ctx, s.bucket, f.storageKey, minio.RemoveObjectOptions{})
}

func scanFile(row pgx.CollectableRow) (File, error) {
	var f File
	err := row.Scan(&f.ID, &f.ObjectID, &f.Kind, &f.OriginalName, &f.ContentType, &f.SizeBytes, &f.Caption, &f.CreatedAt, &f.storageKey)
	f.URL = fileURL(f.ID)
	return f, err
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
