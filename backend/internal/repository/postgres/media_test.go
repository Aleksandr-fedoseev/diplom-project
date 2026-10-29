package postgres_test

import (
	"context"
	"errors"
	"testing"

	"heritage-backend/internal/domain/media"
	"heritage-backend/internal/infrastructure/postgres"
	repo "heritage-backend/internal/repository/postgres"
	"heritage-backend/internal/testdb"
)

func TestMediaRepository(t *testing.T) {
	pool := testdb.Connect(t)
	files := repo.NewMedia(pool)
	ctx := context.Background()

	// Автор, заявка и объект, к которым крепятся файлы. Заявки — чужой модуль, поэтому прямой SQL.
	var userID, submissionID, objectID int64
	if err := pool.QueryRow(ctx, `INSERT INTO auth.users (email, password_hash, name, role) VALUES ($1, 'h', 'Автор', 'user') RETURNING id`,
		testdb.Unique("media")+"@example.ru").Scan(&userID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO submission.submissions (author_id, kind, title, source) VALUES ($1, 'new_object', 'Дом', 'Архив') RETURNING id`,
		userID).Scan(&submissionID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT id FROM catalog.heritage_objects WHERE slug = 'birzha'`).Scan(&objectID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM auth.users WHERE id = $1`, userID) })

	partial, angle := media.ProjectionPartial, 120
	pano, err := files.Create(ctx, media.File{
		SubmissionID: &submissionID, Kind: media.KindPanorama, StorageKey: testdb.Unique("panorama/key"),
		OriginalName: "pano.jpg", ContentType: "image/jpeg", SizeBytes: 100, Caption: "Вид с крыши",
		YearTaken: "2024", SourceNote: "Личный архив", Source: media.SourceAIReconstruction,
		Projection: &partial, HAngle: &angle, Author: "Иванов", License: "CC BY 4.0", UploadedBy: &userID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if pano.ID == 0 || pano.CreatedAt.IsZero() || pano.ObjectID != nil || pano.Projection == nil || *pano.Projection != partial ||
		pano.HAngle == nil || *pano.HAngle != 120 || pano.Source != media.SourceAIReconstruction || pano.SourceNote != "Личный архив" ||
		pano.Author != "Иванов" || pano.License != "CC BY 4.0" || *pano.UploadedBy != userID {
		t.Fatalf("созданная панорама: %+v", pano)
	}

	photo, err := files.Create(ctx, media.File{
		SubmissionID: &submissionID, Kind: media.KindPhoto, StorageKey: testdb.Unique("photo/key"),
		OriginalName: "a.jpg", ContentType: "image/jpeg", SizeBytes: 10, Source: media.SourcePhoto, UploadedBy: &userID,
	})
	if err != nil || photo.Projection != nil || photo.HAngle != nil {
		t.Fatalf("фото: %+v, %v", photo, err)
	}

	// Ограничения миграции 00004: у панорамы обязательна проекция, у фото её нет.
	if _, err := files.Create(ctx, media.File{SubmissionID: &submissionID, Kind: media.KindPanorama,
		StorageKey: testdb.Unique("bad/key"), Source: media.SourcePhoto}); err == nil {
		t.Fatal("панорама без проекции сохранилась")
	}
	if _, err := files.Create(ctx, media.File{SubmissionID: &submissionID, Kind: media.KindPhoto, Projection: &partial,
		StorageKey: testdb.Unique("bad/key"), Source: media.SourcePhoto}); err == nil {
		t.Fatal("фото с проекцией сохранилось")
	}

	list, err := files.ListBySubmission(ctx, submissionID)
	if err != nil || len(list) != 2 || list[0].ID != pano.ID {
		t.Fatalf("файлы заявки: %+v, %v", list, err)
	}

	// Передача объекту в откаченной транзакции не сохраняется.
	boom := errors.New("откат")
	_ = postgres.NewTxManager(pool).Do(ctx, func(ctx context.Context) error {
		if err := files.AttachToObject(ctx, submissionID, objectID); err != nil {
			t.Fatal(err)
		}
		return boom
	})
	if got, _ := files.ByID(ctx, pano.ID); got.ObjectID != nil {
		t.Fatal("передача из откаченной транзакции сохранилась")
	}

	if err := files.AttachToObject(ctx, submissionID, objectID); err != nil {
		t.Fatal(err)
	}
	got, err := files.ByID(ctx, pano.ID)
	if err != nil || got.ObjectID == nil || *got.ObjectID != objectID || got.SubmissionID == nil {
		t.Fatalf("после передачи объекту: %+v, %v", got, err)
	}
	byObject, err := files.ListByObject(ctx, objectID)
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, f := range byObject {
		if f.ID == pano.ID || f.ID == photo.ID {
			found++
		}
	}
	if found != 2 {
		t.Fatalf("файлы объекта: %+v", byObject)
	}

	for _, f := range []media.File{pano, photo} {
		if err := files.Delete(ctx, f.ID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := files.ByID(ctx, pano.ID); !errors.Is(err, media.ErrFileNotFound) {
		t.Fatalf("удалённый файл: %v", err)
	}
	if err := files.Delete(ctx, pano.ID); !errors.Is(err, media.ErrFileNotFound) {
		t.Fatalf("повторное удаление: %v", err)
	}
}
