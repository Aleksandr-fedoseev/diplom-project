package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"heritage-backend/internal/domain/media"
	"heritage-backend/internal/repository/postgres/sqlc"
)

// Media — сведения о медиафайлах. Сами файлы лежат в S3-хранилище.
type Media struct {
	pool *pgxpool.Pool
}

var _ media.Repository = (*Media)(nil)

func NewMedia(pool *pgxpool.Pool) *Media { return &Media{pool: pool} }

func toFile(r sqlc.MediaFile) media.File {
	f := media.File{
		ID:           r.ID,
		ObjectID:     r.ObjectID,
		SubmissionID: r.SubmissionID,
		Kind:         media.Kind(r.Kind),
		StorageKey:   r.StorageKey,
		OriginalName: r.OriginalName,
		ContentType:  r.ContentType,
		SizeBytes:    r.SizeBytes,
		Caption:      r.Caption,
		YearTaken:    r.YearTaken,
		SourceNote:   r.SourceNote,
		Source:       media.Source(r.Source),
		Author:       r.Author,
		License:      r.License,
		UploadedBy:   r.UploadedBy,
		CreatedAt:    r.CreatedAt,
	}
	if r.Projection != nil {
		p := media.Projection(*r.Projection)
		f.Projection = &p
	}
	if r.HAngle != nil {
		a := int(*r.HAngle)
		f.HAngle = &a
	}
	return f
}

func toFiles(rows []sqlc.MediaFile) []media.File {
	files := make([]media.File, len(rows))
	for i, r := range rows {
		files[i] = toFile(r)
	}
	return files
}

func (r *Media) Create(ctx context.Context, f media.File) (media.File, error) {
	var projection *string
	if f.Projection != nil {
		p := string(*f.Projection)
		projection = &p
	}
	var hAngle *int32
	if f.HAngle != nil {
		a := int32(*f.HAngle)
		hAngle = &a
	}
	row, err := queries(ctx, r.pool).CreateMediaFile(ctx, sqlc.CreateMediaFileParams{
		ObjectID:     f.ObjectID,
		SubmissionID: f.SubmissionID,
		Kind:         string(f.Kind),
		StorageKey:   f.StorageKey,
		OriginalName: f.OriginalName,
		ContentType:  f.ContentType,
		SizeBytes:    f.SizeBytes,
		Caption:      f.Caption,
		YearTaken:    f.YearTaken,
		SourceNote:   f.SourceNote,
		Source:       string(f.Source),
		Projection:   projection,
		HAngle:       hAngle,
		Author:       f.Author,
		License:      f.License,
		UploadedBy:   f.UploadedBy,
	})
	if err != nil {
		return media.File{}, err
	}
	return toFile(row), nil
}

func (r *Media) ByID(ctx context.Context, id int64) (media.File, error) {
	row, err := queries(ctx, r.pool).GetMediaFile(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return media.File{}, media.ErrFileNotFound
	}
	if err != nil {
		return media.File{}, err
	}
	return toFile(row), nil
}

func (r *Media) ListByObject(ctx context.Context, objectID int64) ([]media.File, error) {
	rows, err := queries(ctx, r.pool).ListMediaFilesByObject(ctx, &objectID)
	if err != nil {
		return nil, err
	}
	return toFiles(rows), nil
}

func (r *Media) ListBySubmission(ctx context.Context, submissionID int64) ([]media.File, error) {
	rows, err := queries(ctx, r.pool).ListMediaFilesBySubmission(ctx, &submissionID)
	if err != nil {
		return nil, err
	}
	return toFiles(rows), nil
}

func (r *Media) AttachToObject(ctx context.Context, submissionID, objectID int64) error {
	return queries(ctx, r.pool).AttachMediaFilesToObject(ctx, sqlc.AttachMediaFilesToObjectParams{
		ObjectID: &objectID, SubmissionID: &submissionID,
	})
}

func (r *Media) Delete(ctx context.Context, id int64) error {
	n, err := queries(ctx, r.pool).DeleteMediaFile(ctx, id)
	if err != nil {
		return err
	}
	if n == 0 {
		return media.ErrFileNotFound
	}
	return nil
}
