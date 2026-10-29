// Package s3 — небольшой клиент S3-совместимого хранилища (MinIO) поверх minio-go.
// Умеет положить, отдать и удалить файл по ключу; о медиафайлах и правилах домена не знает.
package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	mc     *minio.Client
	bucket string
}

// New подключается к хранилищу и создаёт бакет, если его ещё нет (первый запуск).
func New(ctx context.Context, endpoint, accessKey, secretKey, bucket string, useSSL bool) (*Client, error) {
	mc, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("клиент S3: %w", err)
	}
	c := &Client{mc: mc, bucket: bucket}
	if err := c.ensureBucket(ctx); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) ensureBucket(ctx context.Context) error {
	exists, err := c.mc.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("проверка бакета %q: %w", c.bucket, err)
	}
	if exists {
		return nil
	}
	if err := c.mc.MakeBucket(ctx, c.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("создание бакета %q: %w", c.bucket, err)
	}
	return nil
}

// Put загружает файл потоком: целиком в память он не читается.
func (c *Client) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	_, err := c.mc.PutObject(ctx, c.bucket, key, body, size, minio.PutObjectOptions{ContentType: contentType})
	return err
}

// Get возвращает поток файла. minio.Object умеет Seek, поэтому HTTP может отдавать
// файл частями (Range) — это нужно для аудио, 3D-моделей и больших панорам.
func (c *Client) Get(ctx context.Context, key string) (io.ReadSeekCloser, error) {
	obj, err := c.mc.GetObject(ctx, c.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	// GetObject не обращается к серверу до первого чтения; Stat сразу покажет, что файла нет.
	if _, err := obj.Stat(); err != nil {
		_ = obj.Close()
		return nil, err
	}
	return obj, nil
}

func (c *Client) Remove(ctx context.Context, key string) error {
	return c.mc.RemoveObject(ctx, c.bucket, key, minio.RemoveObjectOptions{})
}
