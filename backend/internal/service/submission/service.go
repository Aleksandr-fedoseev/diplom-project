// Package submission — сценарии модуля заявок: подача, файлы, просмотр, очередь модерации,
// одобрение и отказ. Реализует submission.Service из домена.
package submission

import (
	"context"
	"fmt"

	"heritage-backend/internal/domain/submission"
)

const maxListLimit = 200 // заявок в одном ответе очереди

type Service struct {
	repo    submission.Repository
	tx      TxManager
	events  EventPublisher
	catalog Catalog
	media   Media
	users   Users
}

// Проверка при сборке: Service реализует интерфейс домена.
var _ submission.Service = (*Service)(nil)

func NewService(repo submission.Repository, tx TxManager, events EventPublisher,
	catalog Catalog, media Media, users Users) *Service {
	return &Service{repo: repo, tx: tx, events: events, catalog: catalog, media: media, users: users}
}

func (s *Service) Submit(ctx context.Context, authorID int64, in submission.NewSubmission) (submission.Submission, error) {
	if err := in.Normalize(); err != nil {
		return submission.Submission{}, err
	}
	if err := s.users.EnsureActive(ctx, authorID); err != nil {
		return submission.Submission{}, err
	}

	sub := submission.Submission{
		AuthorID: authorID, Kind: in.Kind, Title: in.Title, Description: in.Description,
		Address: in.Address, Coordinates: in.Coordinates, Period: in.Period, Source: in.Source,
		Status: submission.StatusPending,
	}
	if in.Kind == submission.KindAddition {
		obj, err := s.catalog.ObjectBySlug(ctx, in.ObjectSlug)
		if err != nil {
			return submission.Submission{}, err
		}
		if !obj.Published {
			return submission.Submission{}, submission.ErrObjectNotFound // о скрытых объектах не сообщаем
		}
		sub.ObjectID = &obj.ID
		if sub.Title == "" {
			sub.Title = submission.AdditionTitle(obj.Title)
		}
	}

	id, err := s.repo.Create(ctx, sub)
	if err != nil {
		return submission.Submission{}, fmt.Errorf("создание заявки: %w", err)
	}
	return s.Get(ctx, id, submission.Viewer{UserID: authorID})
}

func (s *Service) AddMedia(ctx context.Context, submissionID, userID int64, in submission.MediaInput) (submission.File, error) {
	sub, err := s.repo.ByID(ctx, submissionID)
	if err != nil {
		return submission.File{}, err
	}
	if err := sub.CanAddMedia(userID); err != nil {
		return submission.File{}, err
	}
	if err := s.users.EnsureActive(ctx, userID); err != nil {
		return submission.File{}, err
	}
	files, err := s.media.ListBySubmission(ctx, submissionID)
	if err != nil {
		return submission.File{}, err
	}
	if err := submission.CheckFilesLimit(len(files)); err != nil {
		return submission.File{}, err
	}
	in.Normalize()
	return s.media.Upload(ctx, MediaUpload{
		SubmissionID: submissionID,
		Kind:         in.Kind,
		Caption:      in.Caption,
		YearTaken:    in.YearTaken,
		SourceNote:   in.SourceNote,
		OriginalName: in.OriginalName,
		Size:         in.Size,
		Body:         in.Body,
		UploadedBy:   userID,
	})
}

// Get возвращает карточку заявки. Чужую заявку пользователь не видит — ErrNotFound.
func (s *Service) Get(ctx context.Context, id int64, v submission.Viewer) (submission.Submission, error) {
	sub, err := s.repo.ByID(ctx, id)
	if err != nil {
		return submission.Submission{}, err
	}
	if !sub.VisibleTo(v) {
		return submission.Submission{}, submission.ErrNotFound
	}
	list := []submission.Submission{sub}
	if err := s.fillObjects(ctx, list); err != nil {
		return submission.Submission{}, err
	}
	sub = list[0]

	if sub.Media, err = s.media.ListBySubmission(ctx, id); err != nil {
		return submission.Submission{}, err
	}
	if sub.Media == nil {
		sub.Media = []submission.File{} // в карточке пустой список, а не отсутствие поля
	}
	if v.IsModerator {
		author, err := s.users.User(ctx, sub.AuthorID)
		if err != nil {
			return submission.Submission{}, err
		}
		sub.Author = &author
	}
	return sub, nil
}

func (s *Service) ListMine(ctx context.Context, authorID int64) ([]submission.Submission, error) {
	list, err := s.repo.ListByAuthor(ctx, authorID)
	if err != nil {
		return nil, err
	}
	return list, s.fillObjects(ctx, list)
}

func (s *Service) Queue(ctx context.Context, status submission.Status, limit, offset int) ([]submission.Submission, error) {
	if !status.Valid() {
		return nil, submission.ErrInvalidStatus
	}
	if limit <= 0 || limit > maxListLimit {
		limit = maxListLimit
	}
	list, err := s.repo.ListByStatus(ctx, status, limit, max(offset, 0))
	if err != nil {
		return nil, err
	}
	return list, s.fillObjects(ctx, list)
}

// fillObjects дописывает slug и название объекта сразу для всего списка — одним запросом к каталогу.
func (s *Service) fillObjects(ctx context.Context, list []submission.Submission) error {
	ids := make([]int64, 0, len(list))
	for _, sub := range list {
		if sub.ObjectID != nil {
			ids = append(ids, *sub.ObjectID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	objects, err := s.catalog.ObjectsByIDs(ctx, ids)
	if err != nil {
		return err
	}
	byID := make(map[int64]ObjectRef, len(objects))
	for _, o := range objects {
		byID[o.ID] = o
	}
	for i := range list {
		if list[i].ObjectID == nil {
			continue
		}
		if o, ok := byID[*list[i].ObjectID]; ok {
			list[i].ObjectSlug, list[i].ObjectTitle = o.Slug, o.Title
		}
	}
	return nil
}
