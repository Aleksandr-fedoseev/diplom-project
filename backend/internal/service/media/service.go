// Package media — сценарии медиафайлов: загрузка в хранилище, выдача с проверкой прав,
// удаление и передача файлов заявки объекту. Реализует media.Service из домена.
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

	"github.com/google/uuid"

	"heritage-backend/internal/domain/errs"
	"heritage-backend/internal/domain/media"
)

type Service struct {
	repo    media.Repository
	storage FileStorage
	objects Objects
	maxAll  int64
}

// Проверка при сборке: Service реализует интерфейс домена.
var _ media.Service = (*Service)(nil)

// NewService: maxUploadMB — общий предел размера файла в мегабайтах (MAX_UPLOAD_MB).
func NewService(repo media.Repository, storage FileStorage, objects Objects, maxUploadMB int64) *Service {
	return &Service{repo: repo, storage: storage, objects: objects, maxAll: maxUploadMB * media.MB}
}

func (s *Service) MaxUploadBytes() int64 { return s.maxAll }

func (s *Service) Upload(ctx context.Context, in media.UploadInput) (media.File, error) {
	rule, err := in.Normalize()
	if err != nil {
		return media.File{}, err
	}
	if err := rule.CheckSize(in.Size, s.maxAll); err != nil {
		return media.File{}, err
	}

	var objectID, submissionID *int64
	var owner string
	switch {
	case in.ObjectID != 0 && in.SubmissionID == 0:
		exists, err := s.objects.Exists(ctx, in.ObjectID)
		if err != nil {
			return media.File{}, err
		}
		if !exists {
			return media.File{}, media.ErrObjectNotFound
		}
		objectID, owner = &in.ObjectID, fmt.Sprintf("object/%d", in.ObjectID)
	case in.SubmissionID != 0 && in.ObjectID == 0:
		submissionID, owner = &in.SubmissionID, fmt.Sprintf("submission/%d", in.SubmissionID)
	default:
		// Ошибка программиста, а не пользователя: вызывающий модуль заполнил оба поля или ни одного.
		return media.File{}, errors.New("файл должен относиться либо к объекту, либо к заявке")
	}

	// Тип определяем по первым байтам файла, а не по расширению или заголовку от клиента:
	// переименованный .exe в .jpg так не пройдёт.
	br := bufio.NewReaderSize(in.Body, 512)
	head, _ := br.Peek(512)
	contentType := detectContentType(head)
	if !rule.Allows(contentType) {
		return media.File{}, errs.New(errs.Invalid, fmt.Sprintf("неподдерживаемый тип файла %q для %s", contentType, in.Kind))
	}

	ext := strings.ToLower(path.Ext(in.OriginalName))
	if ext == "" {
		ext = rule.DefaultExt
	}
	key := fmt.Sprintf("%s/%s/%s%s", in.Kind, owner, uuid.NewString(), ext)

	if err := s.storage.Put(ctx, key, br, in.Size, contentType); err != nil {
		return media.File{}, fmt.Errorf("загрузка в хранилище: %w", err)
	}

	f := media.File{
		ObjectID:     objectID,
		SubmissionID: submissionID,
		Kind:         in.Kind,
		StorageKey:   key,
		OriginalName: in.OriginalName,
		ContentType:  contentType,
		SizeBytes:    in.Size,
		Caption:      in.Caption,
		YearTaken:    in.YearTaken,
		SourceNote:   in.SourceNote,
		Source:       in.Source,
		HAngle:       in.HAngle,
		Author:       in.Author,
		License:      in.License,
		UploadedBy:   &in.UploadedBy,
	}
	if in.Projection != "" {
		f.Projection = &in.Projection
	}
	saved, err := s.repo.Create(ctx, f)
	if err != nil {
		// Запись в БД не удалась — убираем уже загруженный файл, чтобы не оставлять «сирот».
		_ = s.storage.Remove(context.WithoutCancel(ctx), key)
		return media.File{}, fmt.Errorf("сохранение сведений о файле: %w", err)
	}
	return saved, nil
}

func (s *Service) ListByObject(ctx context.Context, objectID int64) ([]media.File, error) {
	return s.repo.ListByObject(ctx, objectID)
}

func (s *Service) ListBySubmission(ctx context.Context, submissionID int64) ([]media.File, error) {
	return s.repo.ListBySubmission(ctx, submissionID)
}

func (s *Service) Open(ctx context.Context, id int64, v media.Viewer) (media.File, io.ReadSeekCloser, error) {
	f, err := s.repo.ByID(ctx, id)
	if err != nil {
		return media.File{}, nil, err
	}
	if !f.CanView(v) {
		return media.File{}, nil, media.ErrFileNotFound
	}
	body, err := s.storage.Get(ctx, f.StorageKey)
	if err != nil {
		return media.File{}, nil, fmt.Errorf("чтение из хранилища: %w", err)
	}
	return f, body, nil
}

// Delete удаляет сведения о файле, затем сам файл. Если хранилище не ответит,
// в нём останется «сирота», но в API файла уже не будет.
func (s *Service) Delete(ctx context.Context, id int64) error {
	f, err := s.repo.ByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	return s.storage.Remove(ctx, f.StorageKey)
}

// AttachToObject вызывается модулем заявок внутри его транзакции: репозиторий берёт её из ctx.
func (s *Service) AttachToObject(ctx context.Context, submissionID, objectID int64) error {
	return s.repo.AttachToObject(ctx, submissionID, objectID)
}

// detectContentType дополняет стандартный http.DetectContentType:
// он не знает GLB, а файл GLB всегда начинается с байтов «glTF».
func detectContentType(head []byte) string {
	if bytes.HasPrefix(head, []byte("glTF")) {
		return "model/gltf-binary"
	}
	return http.DetectContentType(head)
}
