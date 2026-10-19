// Package routes — тематические маршруты администратора и сохранённые маршруты пользователей.
// Линия маршрута строится через OSRM один раз и хранится в БД.
package routes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"heritage-backend/internal/httpx"
	"heritage-backend/internal/osrm"
)

const maxStops = 25

type Stop struct {
	Position int     `json:"position"`
	ObjectID int64   `json:"objectId"`
	Slug     string  `json:"slug"`
	Title    string  `json:"title"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
}

type Route struct {
	ID          int64           `json:"id,omitempty"`
	Slug        string          `json:"slug,omitempty"`
	Title       string          `json:"title,omitempty"`
	Description string          `json:"description,omitempty"`
	OwnerID     *int64          `json:"ownerId,omitempty"`
	DistanceM   float64         `json:"distanceM"`
	DurationS   float64         `json:"durationS"`
	Geometry    json.RawMessage `json:"geometry,omitempty"`
	Stops       []Stop          `json:"stops"`
	CreatedAt   *time.Time      `json:"createdAt,omitempty"`
}

type Service struct {
	db   *pgxpool.Pool
	osrm *osrm.Client
}

func NewService(db *pgxpool.Pool, client *osrm.Client) *Service {
	return &Service{db: db, osrm: client}
}

// Build находит объекты по slug в заданном порядке и строит по ним маршрут, ничего не сохраняя.
// Это и есть «построить свой маршрут» на карте.
func (s *Service) Build(ctx context.Context, slugs []string) (Route, error) {
	if len(slugs) < 2 || len(slugs) > maxStops {
		return Route{}, httpx.BadRequest(fmt.Sprintf("в маршруте должно быть от 2 до %d объектов", maxStops))
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, slug, title, ST_Y(location::geometry), ST_X(location::geometry)
		FROM heritage_objects WHERE slug = ANY($1) AND is_published`, slugs)
	if err != nil {
		return Route{}, err
	}
	found, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Stop, error) {
		var st Stop
		err := row.Scan(&st.ObjectID, &st.Slug, &st.Title, &st.Lat, &st.Lng)
		return st, err
	})
	if err != nil {
		return Route{}, err
	}
	bySlug := make(map[string]Stop, len(found))
	for _, st := range found {
		bySlug[st.Slug] = st
	}

	// Восстанавливаем порядок, который задал пользователь: БД возвращает строки в произвольном порядке.
	stops := make([]Stop, 0, len(slugs))
	points := make([]osrm.Point, 0, len(slugs))
	for i, slug := range slugs {
		st, ok := bySlug[slug]
		if !ok {
			return Route{}, httpx.BadRequest("объект не найден: " + slug)
		}
		st.Position = i + 1
		stops = append(stops, st)
		points = append(points, osrm.Point{Lat: st.Lat, Lng: st.Lng})
	}

	built, err := s.osrm.Route(ctx, points)
	if err != nil {
		slog.Error("OSRM не построил маршрут", "err", err)
		return Route{}, &httpx.Error{Status: 502, Message: "сервис построения маршрутов недоступен"}
	}
	return Route{DistanceM: built.DistanceM, DurationS: built.DurationS, Geometry: built.Geometry, Stops: stops}, nil
}

type SaveInput struct {
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Objects     []string `json:"objects"` // slug'и объектов в порядке посещения
}

// Save строит маршрут и сохраняет его. ownerID == nil — тематический маршрут администратора.
func (s *Service) Save(ctx context.Context, in SaveInput, ownerID *int64) (Route, error) {
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		return Route{}, httpx.BadRequest("укажите название маршрута")
	}
	if ownerID != nil {
		// У пользовательских маршрутов slug служебный: пользователю его придумывать незачем.
		in.Slug = fmt.Sprintf("u%d-%s", *ownerID, uuid.NewString()[:8])
	} else if in.Slug = strings.TrimSpace(strings.ToLower(in.Slug)); in.Slug == "" {
		return Route{}, httpx.BadRequest("укажите slug маршрута")
	}

	built, err := s.Build(ctx, in.Objects)
	if err != nil {
		return Route{}, err
	}

	// Маршрут и его точки пишем в одной транзакции: либо сохраняется всё, либо ничего.
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Route{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }() // после Commit ничего не делает

	var id int64
	err = tx.QueryRow(ctx, `
		INSERT INTO routes (slug, title, description, owner_id, distance_m, duration_s, geometry)
		VALUES ($1, $2, $3, $4, $5, $6, ST_GeomFromGeoJSON($7)::geography)
		RETURNING id`,
		in.Slug, in.Title, in.Description, ownerID, built.DistanceM, built.DurationS, string(built.Geometry),
	).Scan(&id)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return Route{}, httpx.Conflict("маршрут с таким slug уже есть")
	}
	if err != nil {
		return Route{}, err
	}

	for _, st := range built.Stops {
		if _, err := tx.Exec(ctx, `INSERT INTO route_points (route_id, position, object_id) VALUES ($1, $2, $3)`,
			id, st.Position, st.ObjectID); err != nil {
			return Route{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Route{}, err
	}
	return s.Get(ctx, in.Slug)
}

const routeColumns = `id, slug, title, description, owner_id, distance_m, duration_s, created_at`

// ListThematic — тематические маршруты без линий (для списка линии не нужны, они тяжёлые).
func (s *Service) ListThematic(ctx context.Context) ([]Route, error) {
	return s.list(ctx, `WHERE owner_id IS NULL ORDER BY title`)
}

func (s *Service) ListByOwner(ctx context.Context, ownerID int64) ([]Route, error) {
	return s.list(ctx, `WHERE owner_id = $1 ORDER BY created_at DESC`, ownerID)
}

func (s *Service) list(ctx context.Context, tail string, args ...any) ([]Route, error) {
	rows, err := s.db.Query(ctx, `SELECT `+routeColumns+` FROM routes `+tail, args...)
	if err != nil {
		return nil, err
	}
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Route, error) {
		var r Route
		r.CreatedAt = new(time.Time)
		err := row.Scan(&r.ID, &r.Slug, &r.Title, &r.Description, &r.OwnerID, &r.DistanceM, &r.DurationS, r.CreatedAt)
		return r, err
	})
	if err != nil {
		return nil, err
	}
	for i := range list {
		if list[i].Stops, err = s.stops(ctx, list[i].ID); err != nil {
			return nil, err
		}
	}
	return list, nil
}

// Get возвращает маршрут целиком: с линией и остановками.
func (s *Service) Get(ctx context.Context, slug string) (Route, error) {
	var (
		r        Route
		geometry string
	)
	r.CreatedAt = new(time.Time)
	err := s.db.QueryRow(ctx, `SELECT `+routeColumns+`, ST_AsGeoJSON(geometry) FROM routes WHERE slug = $1`, slug).
		Scan(&r.ID, &r.Slug, &r.Title, &r.Description, &r.OwnerID, &r.DistanceM, &r.DurationS, r.CreatedAt, &geometry)
	if errors.Is(err, pgx.ErrNoRows) {
		return Route{}, httpx.NotFound("маршрут не найден")
	}
	if err != nil {
		return Route{}, err
	}
	r.Geometry = json.RawMessage(geometry)
	r.Stops, err = s.stops(ctx, r.ID)
	return r, err
}

func (s *Service) stops(ctx context.Context, routeID int64) ([]Stop, error) {
	rows, err := s.db.Query(ctx, `
		SELECT p.position, o.id, o.slug, o.title, ST_Y(o.location::geometry), ST_X(o.location::geometry)
		FROM route_points p JOIN heritage_objects o ON o.id = p.object_id
		WHERE p.route_id = $1 ORDER BY p.position`, routeID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Stop, error) {
		var st Stop
		err := row.Scan(&st.Position, &st.ObjectID, &st.Slug, &st.Title, &st.Lat, &st.Lng)
		return st, err
	})
}

// Delete удаляет маршрут. Свой маршрут может удалить владелец, тематический — только администратор.
func (s *Service) Delete(ctx context.Context, id int64, userID int64, isAdmin bool) error {
	var ownerID *int64
	err := s.db.QueryRow(ctx, `SELECT owner_id FROM routes WHERE id = $1`, id).Scan(&ownerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return httpx.NotFound("маршрут не найден")
	}
	if err != nil {
		return err
	}
	isOwner := ownerID != nil && *ownerID == userID
	if !isOwner && !isAdmin {
		return httpx.Forbidden("удалить можно только свой маршрут")
	}
	_, err = s.db.Exec(ctx, `DELETE FROM routes WHERE id = $1`, id)
	return err
}
