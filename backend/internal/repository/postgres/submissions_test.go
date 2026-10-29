package postgres_test

import (
	"context"
	"errors"
	"testing"

	"heritage-backend/internal/domain/auth"
	"heritage-backend/internal/domain/submission"
	"heritage-backend/internal/infrastructure/postgres"
	repo "heritage-backend/internal/repository/postgres"
	"heritage-backend/internal/testdb"
)

// Тест репозитория заявок на настоящей БД: SQL, порядок списков, блокировка и единственность решения.
func TestSubmissionsRepository(t *testing.T) {
	pool := testdb.Connect(t)
	subs := repo.NewSubmissions(pool)
	ctx := context.Background()

	author, err := repo.NewUsers(pool).Create(ctx, auth.User{
		Email: testdb.Unique("sub-repo") + "@example.ru", Name: "Автор", Role: auth.RoleUser, PasswordHash: "h",
	})
	if err != nil {
		t.Fatal(err)
	}
	var birzhaID int64
	if err := pool.QueryRow(ctx, `SELECT id FROM catalog.heritage_objects WHERE slug = 'birzha'`).Scan(&birzhaID); err != nil {
		t.Fatal(err)
	}

	// Новый объект с координатами и дополнение без координат.
	firstID, err := subs.Create(ctx, submission.Submission{
		AuthorID: author.ID, Kind: submission.KindNewObject, Title: "Дом", Description: "Описание",
		Address: "ул. 1", Coordinates: &submission.Coordinates{Lat: 46.35, Lng: 48.04}, Period: "1890-е", Source: "Архив",
	})
	if err != nil {
		t.Fatal(err)
	}
	secondID, err := subs.Create(ctx, submission.Submission{
		AuthorID: author.ID, Kind: submission.KindAddition, ObjectID: &birzhaID, Title: "Дополнение", Source: "Книга",
	})
	if err != nil {
		t.Fatal(err)
	}

	first, err := subs.ByID(ctx, firstID)
	if err != nil {
		t.Fatal(err)
	}
	if first.Status != submission.StatusPending || first.Kind != submission.KindNewObject || first.Coordinates == nil ||
		first.Coordinates.Lat != 46.35 || first.Coordinates.Lng != 48.04 || first.Decision != nil || first.CreatedAt.IsZero() {
		t.Fatalf("заявка из БД: %+v", first)
	}
	second, _ := subs.ByID(ctx, secondID)
	if second.Coordinates != nil || second.ObjectID == nil || *second.ObjectID != birzhaID {
		t.Fatalf("дополнение из БД: %+v", second)
	}
	if _, err := subs.ByID(ctx, -1); !errors.Is(err, submission.ErrNotFound) {
		t.Fatalf("несуществующая заявка: %v", err)
	}

	// Свои заявки — новые первыми.
	mine, err := subs.ListByAuthor(ctx, author.ID)
	if err != nil || len(mine) != 2 || mine[0].ID != secondID || mine[1].ID != firstID {
		t.Fatalf("заявки автора: %+v, %v", mine, err)
	}
	// Очередь — старые первыми (очередь общая с другими тестами, поэтому смотрим позиции).
	pos := queuePositions(t, subs, firstID, secondID)
	if pos[0] < 0 || pos[1] < 0 || pos[0] > pos[1] {
		t.Fatalf("в очереди старые заявки должны идти первыми: позиции %v", pos)
	}

	// Блокировка и решение в одной транзакции.
	tm := postgres.NewTxManager(pool)
	err = tm.Do(ctx, func(ctx context.Context) error {
		locked, err := subs.LockForDecision(ctx, firstID)
		if err != nil {
			return err
		}
		if locked.ID != firstID || locked.Status != submission.StatusPending || locked.Coordinates == nil {
			t.Fatalf("заблокированная заявка: %+v", locked)
		}
		if err := subs.SetStatus(ctx, firstID, submission.StatusRejected, nil); err != nil {
			return err
		}
		return subs.CreateDecision(ctx, firstID, submission.Decision{
			Decision: submission.StatusRejected, Reason: "Дубликат", ModeratorID: &author.ID,
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := tm.Do(ctx, func(ctx context.Context) error {
		_, err := subs.LockForDecision(ctx, -1)
		return err
	}); !errors.Is(err, submission.ErrNotFound) {
		t.Fatalf("блокировка несуществующей: %v", err)
	}

	first, _ = subs.ByID(ctx, firstID)
	if first.Status != submission.StatusRejected || first.Decision == nil || first.Decision.Reason != "Дубликат" ||
		first.Decision.Decision != submission.StatusRejected || first.Decision.ModeratorID == nil || first.Decision.CreatedAt.IsZero() {
		t.Fatalf("решение: %+v", first.Decision)
	}
	// Решение по заявке одно.
	if err := subs.CreateDecision(ctx, firstID, submission.Decision{Decision: submission.StatusApproved}); !errors.Is(err, submission.ErrAlreadyDecided) {
		t.Fatalf("второе решение: %v", err)
	}

	// SetStatus без objectID не трогает объект, с ним — запоминает.
	if err := subs.SetStatus(ctx, secondID, submission.StatusApproved, nil); err != nil {
		t.Fatal(err)
	}
	if got, _ := subs.ByID(ctx, secondID); got.ObjectID == nil || *got.ObjectID != birzhaID || got.Status != submission.StatusApproved {
		t.Fatalf("статус дополнения: %+v", got)
	}
	if err := subs.SetStatus(ctx, -1, submission.StatusApproved, nil); !errors.Is(err, submission.ErrNotFound) {
		t.Fatalf("статус несуществующей: %v", err)
	}

	// Список по статусу отдаёт решение вместе с заявкой.
	rejected, err := subs.ListByStatus(ctx, submission.StatusRejected, 1000, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, s := range rejected {
		if s.ID == firstID {
			found = s.Decision != nil && s.Decision.Reason == "Дубликат"
		}
	}
	if !found {
		t.Fatal("отклонённая заявка с причиной не найдена в списке по статусу")
	}
}

// queuePositions проходит очередь ожидающих заявок по страницам и возвращает позиции нужных.
func queuePositions(t *testing.T, subs *repo.Submissions, ids ...int64) []int {
	t.Helper()
	const page = 200
	pos := make([]int, len(ids))
	for i := range pos {
		pos[i] = -1
	}
	n := 0
	for offset := 0; ; offset += page {
		list, err := subs.ListByStatus(context.Background(), submission.StatusPending, page, offset)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range list {
			for i, id := range ids {
				if s.ID == id {
					pos[i] = n
				}
			}
			n++
		}
		if len(list) < page {
			return pos
		}
	}
}
