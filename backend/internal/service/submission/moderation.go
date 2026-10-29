package submission

import (
	"context"

	"heritage-backend/internal/domain/submission"
)

// Approve одобряет заявку. Объект (или перенос файлов к существующему), статус, решение
// и событие submission.approved сохраняются в одной транзакции: либо всё, либо ничего.
func (s *Service) Approve(ctx context.Context, id int64, mod submission.Moderator, in submission.ApproveInput) (submission.Submission, error) {
	err := s.decide(ctx, id, mod, func(ctx context.Context, sub submission.Submission) error {
		var obj ObjectRef
		var err error
		if sub.Kind == submission.KindNewObject {
			obj, err = s.createObject(ctx, sub, in)
		} else {
			obj, err = s.additionObject(ctx, sub)
		}
		if err != nil {
			return err
		}

		if err := s.repo.SetStatus(ctx, id, submission.StatusApproved, &obj.ID); err != nil {
			return err
		}
		if err := s.repo.CreateDecision(ctx, id, submission.Decision{
			Decision: submission.StatusApproved, ModeratorID: &mod.UserID,
		}); err != nil {
			return err
		}
		if err := s.media.AttachToObject(ctx, id, obj.ID); err != nil {
			return err
		}
		return s.events.Publish(ctx, submission.EventApproved, submission.ApprovedEvent{
			SubmissionID: id, AuthorID: sub.AuthorID, Title: sub.Title, ObjectSlug: obj.Slug,
		})
	})
	if err != nil {
		return submission.Submission{}, err
	}
	return s.Get(ctx, id, submission.Viewer{UserID: mod.UserID, IsModerator: true})
}

// createObject создаёт объект каталога из полей модератора; пустые берутся из заявки.
func (s *Service) createObject(ctx context.Context, sub submission.Submission, in submission.ApproveInput) (ObjectRef, error) {
	in, err := in.Prefill(sub)
	if err != nil {
		return ObjectRef{}, err
	}
	return s.catalog.CreateObject(ctx, NewObject{
		Slug:               in.Slug,
		Title:              in.Title,
		Category:           in.Category,
		ProtectionStatus:   in.ProtectionStatus,
		RegistryNumber:     in.RegistryNumber,
		YearBuilt:          in.YearBuilt,
		Architect:          in.Architect,
		ArchitecturalStyle: in.ArchitecturalStyle,
		Address:            in.Address,
		Lat:                in.Coordinates.Lat,
		Lng:                in.Coordinates.Lng,
		ShortDescription:   in.ShortDescription,
		FullHistory:        in.FullHistory,
		Tags:               in.Tags,
		IsPublished:        *in.IsPublished,
	})
}

// additionObject — объект, к которому относится дополнение (slug нужен для события).
func (s *Service) additionObject(ctx context.Context, sub submission.Submission) (ObjectRef, error) {
	if sub.ObjectID == nil {
		return ObjectRef{}, submission.ErrObjectDeleted
	}
	objects, err := s.catalog.ObjectsByIDs(ctx, []int64{*sub.ObjectID})
	if err != nil {
		return ObjectRef{}, err
	}
	if len(objects) == 0 {
		return ObjectRef{}, submission.ErrObjectDeleted
	}
	return objects[0], nil
}

// Reject отклоняет заявку с причиной. Блокировка автора, статус, решение и событие
// submission.rejected сохраняются в одной транзакции: если автора заблокировать нельзя
// (например, он модератор), отказ тоже не сохраняется.
func (s *Service) Reject(ctx context.Context, id int64, mod submission.Moderator, in submission.RejectInput) (submission.Submission, error) {
	if err := in.Normalize(); err != nil {
		return submission.Submission{}, err
	}
	err := s.decide(ctx, id, mod, func(ctx context.Context, sub submission.Submission) error {
		if in.BlockAuthor {
			if err := s.users.Block(ctx, mod.UserID, mod.Role, sub.AuthorID); err != nil {
				return err
			}
		}
		if err := s.repo.SetStatus(ctx, id, submission.StatusRejected, nil); err != nil {
			return err
		}
		if err := s.repo.CreateDecision(ctx, id, submission.Decision{
			Decision: submission.StatusRejected, Reason: in.Reason, ModeratorID: &mod.UserID,
		}); err != nil {
			return err
		}
		return s.events.Publish(ctx, submission.EventRejected, submission.RejectedEvent{
			SubmissionID: id, AuthorID: sub.AuthorID, Title: sub.Title,
			Reason: in.Reason, AuthorBlocked: in.BlockAuthor,
		})
	})
	if err != nil {
		return submission.Submission{}, err
	}
	return s.Get(ctx, id, submission.Viewer{UserID: mod.UserID, IsModerator: true})
}

// decide — общая часть одобрения и отказа: транзакция, блокировка строки заявки
// (SELECT … FOR UPDATE) и проверка правил решения. apply делает саму запись в той же транзакции.
// Если два модератора решают одну заявку, второй ждёт окончания транзакции первого
// и получает ErrAlreadyDecided.
func (s *Service) decide(ctx context.Context, id int64, mod submission.Moderator,
	apply func(ctx context.Context, sub submission.Submission) error) error {
	return s.tx.Do(ctx, func(ctx context.Context) error {
		sub, err := s.repo.LockForDecision(ctx, id)
		if err != nil {
			return err
		}
		if err := sub.CanDecide(mod.UserID); err != nil {
			return err
		}
		return apply(ctx, sub)
	})
}
