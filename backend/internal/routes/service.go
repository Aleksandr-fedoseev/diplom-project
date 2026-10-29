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

	"heritage-backend/internal/catalog"
	"heritage-backend/internal/httpx"
	"heritage-backend/internal/osrm"
	"heritage-backend/internal/routes/store"
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

// Places — то, что маршрутам нужно от каталога: координаты и названия объектов.
// Реализует catalog.Service; таблицу объектов модуль маршрутов сам не читает.
type Places interface {
	PublishedPlaces(ctx context.Context, slugs []string) ([]catalog.Place, error)
	PlacesByIDs(ctx context.Context, ids []int64) ([]catalog.Place, error)
}

type Service struct {
	db     *pgxpool.Pool  // нужен только для транзакций
	q      *store.Queries // запросы из queries.sql, код сгенерирован sqlc
	places Places
	osrm   *osrm.Client
}

func NewService(db *pgxpool.Pool, places Places, client *osrm.Client) *Service {
	return &Service{db: db, q: store.New(db), places: places, osrm: client}
}

// Build находит объекты по slug в заданном порядке и строит по ним маршрут, ничего не сохраняя.
// Это и есть «построить свой маршрут» на карте.
func (s *Service) Build(ctx context.Context, slugs []string) (Route, error) {
	if len(slugs) < 2 || len(slugs) > maxStops {
		return Route{}, httpx.BadRequest(fmt.Sprintf("в маршруте должно быть от 2 до %d объектов", maxStops))
	}

	found, err := s.places.PublishedPlaces(ctx, slugs)
	if err != nil {
		return Route{}, err
	}
	bySlug := make(map[string]Stop, len(found))
	for _, p := range found {
		bySlug[p.Slug] = stopFromPlace(p)
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

	qtx := s.q.WithTx(tx)
	id, err := qtx.CreateRoute(ctx, store.CreateRouteParams{
		Slug:        in.Slug,
		Title:       in.Title,
		Description: in.Description,
		OwnerID:     ownerID,
		DistanceM:   built.DistanceM,
		DurationS:   built.DurationS,
		Geometry:    string(built.Geometry),
	})
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return Route{}, httpx.Conflict("маршрут с таким slug уже есть")
	}
	if err != nil {
		return Route{}, err
	}

	for _, st := range built.Stops {
		err := qtx.CreateRoutePoint(ctx, store.CreateRoutePointParams{
			RouteID: id, Position: int32(st.Position), ObjectID: st.ObjectID,
		})
		if err != nil {
			return Route{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Route{}, err
	}
	return s.Get(ctx, in.Slug)
}

// ListThematic — тематические маршруты без линий (для списка линии не нужны, они тяжёлые).
func (s *Service) ListThematic(ctx context.Context) ([]Route, error) {
	rows, err := s.q.ListThematicRoutes(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]Route, len(rows))
	for i, r := range rows {
		list[i] = Route{ID: r.ID, Slug: r.Slug, Title: r.Title, Description: r.Description, OwnerID: r.OwnerID,
			DistanceM: r.DistanceM, DurationS: r.DurationS, CreatedAt: &r.CreatedAt}
	}
	return list, s.fillStops(ctx, list)
}

func (s *Service) ListByOwner(ctx context.Context, ownerID int64) ([]Route, error) {
	rows, err := s.q.ListRoutesByOwner(ctx, &ownerID)
	if err != nil {
		return nil, err
	}
	list := make([]Route, len(rows))
	for i, r := range rows {
		list[i] = Route{ID: r.ID, Slug: r.Slug, Title: r.Title, Description: r.Description, OwnerID: r.OwnerID,
			DistanceM: r.DistanceM, DurationS: r.DurationS, CreatedAt: &r.CreatedAt}
	}
	return list, s.fillStops(ctx, list)
}

// Get возвращает маршрут целиком: с линией и остановками.
func (s *Service) Get(ctx context.Context, slug string) (Route, error) {
	r, err := s.q.GetRouteBySlug(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return Route{}, httpx.NotFound("маршрут не найден")
	}
	if err != nil {
		return Route{}, err
	}
	route := []Route{{ID: r.ID, Slug: r.Slug, Title: r.Title, Description: r.Description, OwnerID: r.OwnerID,
		DistanceM: r.DistanceM, DurationS: r.DurationS, CreatedAt: &r.CreatedAt, Geometry: json.RawMessage(r.Geometry)}}
	return route[0], s.fillStops(ctx, route)
}

// fillStops заполняет остановки сразу для всех маршрутов списка: один запрос за точками
// и один к каталогу за объектами, а не по запросу на каждый маршрут.
func (s *Service) fillStops(ctx context.Context, list []Route) error {
	if len(list) == 0 {
		return nil
	}
	routeIDs := make([]int64, len(list))
	for i, r := range list {
		routeIDs[i] = r.ID
	}
	points, err := s.q.ListRoutePoints(ctx, routeIDs)
	if err != nil {
		return err
	}

	objectIDs := make([]int64, 0, len(points))
	for _, p := range points {
		objectIDs = append(objectIDs, p.ObjectID)
	}
	places, err := s.places.PlacesByIDs(ctx, objectIDs)
	if err != nil {
		return err
	}
	byID := make(map[int64]catalog.Place, len(places))
	for _, p := range places {
		byID[p.ID] = p
	}

	// Точки приходят отсортированными по маршруту и номеру остановки.
	stops := make(map[int64][]Stop, len(list))
	for _, p := range points {
		st := stopFromPlace(byID[p.ObjectID])
		st.Position = int(p.Position)
		stops[p.RouteID] = append(stops[p.RouteID], st)
	}
	for i := range list {
		list[i].Stops = stops[list[i].ID]
		if list[i].Stops == nil {
			list[i].Stops = []Stop{}
		}
	}
	return nil
}

func stopFromPlace(p catalog.Place) Stop {
	return Stop{ObjectID: p.ID, Slug: p.Slug, Title: p.Title, Lat: p.Lat, Lng: p.Lng}
}

// Delete удаляет маршрут. Свой маршрут может удалить владелец, тематический — только администратор.
func (s *Service) Delete(ctx context.Context, id int64, userID int64, isAdmin bool) error {
	ownerID, err := s.q.GetRouteOwner(ctx, id)
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
	return s.q.DeleteRoute(ctx, id)
}
