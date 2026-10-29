package media

import (
	"context"
	"io"
)

// Порты сервиса: техника и другие модули видны ему только через эти интерфейсы.

// FileStorage — хранилище содержимого файлов (S3/MinIO). Реализация — infrastructure/s3.
type FileStorage interface {
	Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	// Get возвращает поток, который умеет Seek (нужно для Range-запросов).
	Get(ctx context.Context, key string) (io.ReadSeekCloser, error)
	Remove(ctx context.Context, key string) error
}

// Objects — то, что медиафайлам нужно знать об объектах каталога.
// Адаптер к catalog.Service пишется при сборке приложения.
type Objects interface {
	Exists(ctx context.Context, id int64) (bool, error)
}
