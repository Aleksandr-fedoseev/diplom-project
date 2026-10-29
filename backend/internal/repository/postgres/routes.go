package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"heritage-backend/internal/domain/route"
	"heritage-backend/internal/repository/postgres/sqlc"
)

// Routes — хранилище маршрутов и их точек (таблицы routes и route_points).
type Routes struct {
	pool *pgxpool.Pool
}

var _ route.Repository = (*Routes)(nil)

func NewRoutes(pool *pgxpool.Pool) *Routes { return &Routes{pool: pool} }

func (r *Routes) Create(ctx context.Context, rt route.Route) (int64, error) {
	id, err := queries(ctx, r.pool).CreateRoute(ctx, sqlc.CreateRouteParams{
		Slug:        rt.Slug,
		Title:       rt.Title,
		Description: rt.Description,
		OwnerID:     rt.OwnerID,
		DistanceM:   rt.DistanceM,
		DurationS:   rt.DurationS,
		Geometry:    string(rt.Geometry),
	})
	if pgCode(err) == codeUniqueViolation {
		return 0, route.ErrSlugTaken
	}
	return id, err
}

func (r *Routes) AddStops(ctx context.Context, routeID int64, stops []route.Stop) error {
	q := queries(ctx, r.pool)
	for _, st := range stops {
		err := q.CreateRoutePoint(ctx, sqlc.CreateRoutePointParams{
			RouteID: routeID, Position: int32(st.Position), ObjectID: st.ObjectID,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Routes) BySlug(ctx context.Context, slug string) (route.Route, error) {
	row, err := queries(ctx, r.pool).GetRouteBySlug(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return route.Route{}, route.ErrNotFound
	}
	if err != nil {
		return route.Route{}, err
	}
	rt := toRoute(row.ID, row.Slug, row.Title, row.Description, row.OwnerID, row.DistanceM, row.DurationS, row.CreatedAt)
	rt.Geometry = json.RawMessage(row.Geometry)
	return rt, nil
}

func (r *Routes) ListThematic(ctx context.Context) ([]route.Route, error) {
	rows, err := queries(ctx, r.pool).ListThematicRoutes(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]route.Route, len(rows))
	for i, x := range rows {
		list[i] = toRoute(x.ID, x.Slug, x.Title, x.Description, x.OwnerID, x.DistanceM, x.DurationS, x.CreatedAt)
	}
	return list, nil
}

func (r *Routes) ListByOwner(ctx context.Context, ownerID int64) ([]route.Route, error) {
	rows, err := queries(ctx, r.pool).ListRoutesByOwner(ctx, &ownerID)
	if err != nil {
		return nil, err
	}
	list := make([]route.Route, len(rows))
	for i, x := range rows {
		list[i] = toRoute(x.ID, x.Slug, x.Title, x.Description, x.OwnerID, x.DistanceM, x.DurationS, x.CreatedAt)
	}
	return list, nil
}

// toRoute собирает сущность из полей: у каждого запроса sqlc свой тип строки, а поля одни и те же.
func toRoute(id int64, slug, title, description string, ownerID *int64, distanceM, durationS float64, createdAt time.Time) route.Route {
	return route.Route{
		ID: id, Slug: slug, Title: title, Description: description, OwnerID: ownerID,
		DistanceM: distanceM, DurationS: durationS, CreatedAt: createdAt,
	}
}

func (r *Routes) StopRefs(ctx context.Context, routeIDs []int64) ([]route.StopRef, error) {
	rows, err := queries(ctx, r.pool).ListRoutePoints(ctx, routeIDs)
	if err != nil {
		return nil, err
	}
	refs := make([]route.StopRef, len(rows))
	for i, x := range rows {
		refs[i] = route.StopRef{RouteID: x.RouteID, Position: int(x.Position), ObjectID: x.ObjectID}
	}
	return refs, nil
}

func (r *Routes) Owner(ctx context.Context, id int64) (*int64, error) {
	ownerID, err := queries(ctx, r.pool).GetRouteOwner(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, route.ErrNotFound
	}
	return ownerID, err
}

// Delete удаляет маршрут; его точки удаляет сама БД (ON DELETE CASCADE).
func (r *Routes) Delete(ctx context.Context, id int64) error {
	return queries(ctx, r.pool).DeleteRoute(ctx, id)
}
