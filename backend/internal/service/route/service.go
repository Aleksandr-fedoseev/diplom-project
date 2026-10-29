// Package route — сценарии модуля маршрутов: построение, сохранение, списки, просмотр, удаление.
// Линия маршрута строится маршрутизатором один раз и хранится в БД. Реализует route.Service из домена.
package route

import (
	"context"
	"log/slog"

	"heritage-backend/internal/domain/route"
)

type Service struct {
	routes route.Repository
	places Places
	router Router
	tx     TxManager
}

// Проверка при сборке: Service реализует интерфейс домена.
var _ route.Service = (*Service)(nil)

func NewService(routes route.Repository, places Places, router Router, tx TxManager) *Service {
	return &Service{routes: routes, places: places, router: router, tx: tx}
}

// Build находит объекты по slug в заданном порядке и строит по ним маршрут, ничего не сохраняя.
// Это и есть «построить свой маршрут» на карте.
func (s *Service) Build(ctx context.Context, slugs []string) (route.Route, error) {
	if err := route.CheckStopCount(len(slugs)); err != nil {
		return route.Route{}, err
	}

	found, err := s.places.PublishedPlaces(ctx, slugs)
	if err != nil {
		return route.Route{}, err
	}
	bySlug := make(map[string]Place, len(found))
	for _, p := range found {
		bySlug[p.Slug] = p
	}

	// Восстанавливаем порядок, который задал пользователь: каталог возвращает объекты в произвольном порядке.
	stops := make([]route.Stop, 0, len(slugs))
	points := make([]Point, 0, len(slugs))
	for i, slug := range slugs {
		p, ok := bySlug[slug]
		if !ok {
			return route.Route{}, route.ObjectNotFound(slug)
		}
		st := stopFromPlace(p)
		st.Position = i + 1
		stops = append(stops, st)
		points = append(points, Point{Lat: p.Lat, Lng: p.Lng})
	}

	line, err := s.router.Route(ctx, points)
	if err != nil {
		// Подробности — в лог, пользователю — понятное сообщение.
		slog.Error("маршрутизатор не построил маршрут", "err", err)
		return route.Route{}, route.ErrRouterUnavailable
	}
	return route.Route{DistanceM: line.DistanceM, DurationS: line.DurationS, Geometry: line.Geometry, Stops: stops}, nil
}

// Save строит маршрут и сохраняет его. ownerID == nil — тематический маршрут администратора.
func (s *Service) Save(ctx context.Context, in route.SaveInput, ownerID *int64) (route.Route, error) {
	if err := in.Normalize(ownerID); err != nil {
		return route.Route{}, err
	}
	built, err := s.Build(ctx, in.Objects)
	if err != nil {
		return route.Route{}, err
	}
	built.Slug, built.Title, built.Description, built.OwnerID = in.Slug, in.Title, in.Description, ownerID

	// Маршрут и его точки пишем в одной транзакции: либо сохраняется всё, либо ничего.
	err = s.tx.Do(ctx, func(ctx context.Context) error {
		id, err := s.routes.Create(ctx, built)
		if err != nil {
			return err
		}
		return s.routes.AddStops(ctx, id, built.Stops)
	})
	if err != nil {
		return route.Route{}, err
	}
	// Читаем сохранённое заново: так ответ совпадает с тем, что потом отдаст GET (id, дата, линия из БД).
	return s.load(ctx, in.Slug)
}

// ListThematic — тематические маршруты без линий (для списка линии не нужны, они тяжёлые).
func (s *Service) ListThematic(ctx context.Context) ([]route.Route, error) {
	list, err := s.routes.ListThematic(ctx)
	if err != nil {
		return nil, err
	}
	return list, s.fillStops(ctx, list)
}

func (s *Service) ListByOwner(ctx context.Context, ownerID int64) ([]route.Route, error) {
	list, err := s.routes.ListByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	return list, s.fillStops(ctx, list)
}

func (s *Service) Get(ctx context.Context, slug string, viewer *route.Actor) (route.Route, error) {
	r, err := s.load(ctx, slug)
	if err != nil {
		return route.Route{}, err
	}
	// Чужой сохранённый маршрут не показываем и не выдаём, что он существует.
	if !r.CanView(viewer) {
		return route.Route{}, route.ErrNotFound
	}
	return r, nil
}

// load — маршрут целиком (с линией и остановками), без проверки прав.
func (s *Service) load(ctx context.Context, slug string) (route.Route, error) {
	r, err := s.routes.BySlug(ctx, slug)
	if err != nil {
		return route.Route{}, err
	}
	one := []route.Route{r}
	if err := s.fillStops(ctx, one); err != nil {
		return route.Route{}, err
	}
	return one[0], nil
}

// Delete удаляет маршрут. Свой маршрут может удалить владелец, любой — администратор.
func (s *Service) Delete(ctx context.Context, id int64, actor route.Actor) error {
	ownerID, err := s.routes.Owner(ctx, id)
	if err != nil {
		return err
	}
	if !route.CanDelete(ownerID, actor) {
		return route.ErrDeleteDenied
	}
	return s.routes.Delete(ctx, id)
}

// fillStops заполняет остановки сразу для всех маршрутов списка: один запрос за точками
// и один к каталогу за объектами, а не по запросу на каждый маршрут.
func (s *Service) fillStops(ctx context.Context, list []route.Route) error {
	if len(list) == 0 {
		return nil
	}
	routeIDs := make([]int64, len(list))
	for i, r := range list {
		routeIDs[i] = r.ID
	}
	refs, err := s.routes.StopRefs(ctx, routeIDs)
	if err != nil {
		return err
	}

	objectIDs := make([]int64, 0, len(refs))
	for _, ref := range refs {
		objectIDs = append(objectIDs, ref.ObjectID)
	}
	places, err := s.places.PlacesByIDs(ctx, objectIDs)
	if err != nil {
		return err
	}
	byID := make(map[int64]Place, len(places))
	for _, p := range places {
		byID[p.ID] = p
	}

	// Точки приходят отсортированными по маршруту и номеру остановки.
	stops := make(map[int64][]route.Stop, len(list))
	for _, ref := range refs {
		st := stopFromPlace(byID[ref.ObjectID])
		st.ObjectID = ref.ObjectID
		st.Position = ref.Position
		stops[ref.RouteID] = append(stops[ref.RouteID], st)
	}
	for i := range list {
		list[i].Stops = stops[list[i].ID]
		if list[i].Stops == nil {
			list[i].Stops = []route.Stop{}
		}
	}
	return nil
}

func stopFromPlace(p Place) route.Stop {
	return route.Stop{ObjectID: p.ID, Slug: p.Slug, Title: p.Title, Lat: p.Lat, Lng: p.Lng}
}
