package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"heritage-backend/internal/domain/submission"
	"heritage-backend/internal/repository/postgres/sqlc"
)

// Submissions — заявки и решения модераторов (таблицы submissions и moderation_decisions).
type Submissions struct {
	pool *pgxpool.Pool
}

var _ submission.Repository = (*Submissions)(nil)

func NewSubmissions(pool *pgxpool.Pool) *Submissions { return &Submissions{pool: pool} }

func (r *Submissions) Create(ctx context.Context, s submission.Submission) (int64, error) {
	var lat, lng *float64
	if c := s.Coordinates; c != nil {
		lat, lng = &c.Lat, &c.Lng
	}
	return queries(ctx, r.pool).CreateSubmission(ctx, sqlc.CreateSubmissionParams{
		AuthorID:    s.AuthorID,
		Kind:        string(s.Kind),
		ObjectID:    s.ObjectID,
		Title:       s.Title,
		Description: s.Description,
		Address:     s.Address,
		Lat:         lat,
		Lng:         lng,
		Period:      s.Period,
		Source:      s.Source,
	})
}

func (r *Submissions) ByID(ctx context.Context, id int64) (submission.Submission, error) {
	row, err := queries(ctx, r.pool).GetSubmission(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return submission.Submission{}, submission.ErrNotFound
	}
	if err != nil {
		return submission.Submission{}, err
	}
	return toSubmission(row), nil
}

func (r *Submissions) LockForDecision(ctx context.Context, id int64) (submission.Submission, error) {
	row, err := queries(ctx, r.pool).LockSubmission(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return submission.Submission{}, submission.ErrNotFound
	}
	if err != nil {
		return submission.Submission{}, err
	}
	// У строки без решения те же поля, что у GetSubmissionRow до колонок решения.
	return toSubmission(sqlc.GetSubmissionRow{
		ID: row.ID, AuthorID: row.AuthorID, Kind: row.Kind, ObjectID: row.ObjectID, Title: row.Title,
		Description: row.Description, Address: row.Address, HasLocation: row.HasLocation, Lat: row.Lat,
		Lng: row.Lng, Period: row.Period, Source: row.Source, Status: row.Status, CreatedAt: row.CreatedAt,
	}), nil
}

func (r *Submissions) ListByAuthor(ctx context.Context, authorID int64) ([]submission.Submission, error) {
	rows, err := queries(ctx, r.pool).ListSubmissionsByAuthor(ctx, authorID)
	if err != nil {
		return nil, err
	}
	list := make([]submission.Submission, len(rows))
	for i, row := range rows {
		list[i] = toSubmission(sqlc.GetSubmissionRow(row))
	}
	return list, nil
}

func (r *Submissions) ListByStatus(ctx context.Context, status submission.Status, limit, offset int) ([]submission.Submission, error) {
	rows, err := queries(ctx, r.pool).ListSubmissionsByStatus(ctx, sqlc.ListSubmissionsByStatusParams{
		Status: string(status), Lim: int32(limit), Off: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	list := make([]submission.Submission, len(rows))
	for i, row := range rows {
		list[i] = toSubmission(sqlc.GetSubmissionRow(row))
	}
	return list, nil
}

func (r *Submissions) SetStatus(ctx context.Context, id int64, status submission.Status, objectID *int64) error {
	n, err := queries(ctx, r.pool).SetSubmissionStatus(ctx, sqlc.SetSubmissionStatusParams{
		ID: id, Status: string(status), ObjectID: objectID,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return submission.ErrNotFound
	}
	return nil
}

func (r *Submissions) CreateDecision(ctx context.Context, submissionID int64, d submission.Decision) error {
	err := queries(ctx, r.pool).CreateModerationDecision(ctx, sqlc.CreateModerationDecisionParams{
		SubmissionID: submissionID, ModeratorID: d.ModeratorID, Decision: string(d.Decision), Reason: d.Reason,
	})
	switch pgCode(err) {
	case codeUniqueViolation: // по заявке уже есть решение
		return submission.ErrAlreadyDecided
	case codeForeignKeyViolation:
		return submission.ErrNotFound
	}
	return err
}

func toSubmission(r sqlc.GetSubmissionRow) submission.Submission {
	s := submission.Submission{
		ID:          r.ID,
		AuthorID:    r.AuthorID,
		Kind:        submission.Kind(r.Kind),
		ObjectID:    r.ObjectID,
		Title:       r.Title,
		Description: r.Description,
		Address:     r.Address,
		Period:      r.Period,
		Source:      r.Source,
		Status:      submission.Status(r.Status),
		CreatedAt:   r.CreatedAt,
	}
	if r.HasLocation {
		s.Coordinates = &submission.Coordinates{Lat: r.Lat, Lng: r.Lng}
	}
	if r.Decision != nil {
		s.Decision = &submission.Decision{
			Decision:    submission.Status(*r.Decision),
			ModeratorID: r.ModeratorID,
			CreatedAt:   r.DecidedAt.Time,
		}
		if r.Reason != nil {
			s.Decision.Reason = *r.Reason
		}
	}
	return s
}
