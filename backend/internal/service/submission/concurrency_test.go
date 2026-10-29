package submission

import (
	"context"
	"sync"
	"testing"

	"heritage-backend/internal/domain/auth"
	"heritage-backend/internal/domain/errs"
	"heritage-backend/internal/domain/submission"
	"heritage-backend/internal/infrastructure/postgres"
	repo "heritage-backend/internal/repository/postgres"
	"heritage-backend/internal/testdb"
)

// Гонку двух модераторов можно проверить только на настоящей БД: SELECT … FOR UPDATE
// в транзакции заставляет второго ждать первого. Остальные порты — фейки.
func TestApproveConcurrently(t *testing.T) {
	pool := testdb.Connect(t)
	ctx := context.Background()

	users := repo.NewUsers(pool)
	newUser := func(prefix string, role auth.Role) auth.User {
		u, err := users.Create(ctx, auth.User{Email: testdb.Unique(prefix) + "@example.ru", Name: prefix, Role: role, PasswordHash: "h"})
		if err != nil {
			t.Fatal(err)
		}
		return u
	}
	author, mod := newUser("race-author", auth.RoleUser), newUser("race-mod", auth.RoleModerator)

	var birzhaID int64
	if err := pool.QueryRow(ctx, `SELECT id FROM catalog.heritage_objects WHERE slug = 'birzha'`).Scan(&birzhaID); err != nil {
		t.Fatal(err)
	}
	catalog := &fakeCatalog{objects: map[string]ObjectRef{"birzha": {ID: birzhaID, Slug: "birzha", Title: "Биржа", Published: true}}}
	events, media := &fakePublisher{}, newFakeMedia()
	fakeUsers := &fakeUsers{users: map[int64]Author{
		author.ID: {ID: author.ID, Role: "user"}, mod.ID: {ID: mod.ID, Role: "moderator"},
	}}
	svc := NewService(repo.NewSubmissions(pool), postgres.NewTxManager(pool), events, catalog, media, fakeUsers)

	sub, err := svc.Submit(ctx, author.ID, submission.NewSubmission{Kind: submission.KindAddition, ObjectSlug: "birzha", Source: "Архив"})
	if err != nil {
		t.Fatal(err)
	}

	// Два модератора одобряют одновременно: проходит ровно одно решение, второе получает конфликт.
	results := make([]error, 2)
	var wg sync.WaitGroup
	for i := range results {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, results[i] = svc.Approve(ctx, sub.ID, submission.Moderator{UserID: mod.ID, Role: "moderator"}, submission.ApproveInput{})
		}()
	}
	wg.Wait()

	ok := 0
	for _, err := range results {
		switch {
		case err == nil:
			ok++
		case !errs.Is(err, errs.Conflict):
			t.Fatalf("ожидался конфликт, получено: %v", err)
		}
	}
	if ok != 1 {
		t.Fatalf("одобрений прошло %d, ожидалось 1: %v", ok, results)
	}
	if len(events.events) != 1 {
		t.Fatalf("событий: %d, ожидалось 1", len(events.events))
	}
	got, err := svc.Get(ctx, sub.ID, submission.Viewer{UserID: author.ID})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != submission.StatusApproved || got.ObjectSlug != "birzha" || media.attached[sub.ID] != birzhaID {
		t.Fatalf("одобренное дополнение: %+v, attached=%v", got, media.attached)
	}
}
