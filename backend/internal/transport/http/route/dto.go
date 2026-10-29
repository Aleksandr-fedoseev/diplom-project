package route

import (
	"encoding/json"
	"time"

	"heritage-backend/internal/domain/route"
)

// DTO — формат JSON в API, как в прежнем модуле routes. У построенного, но не сохранённого
// маршрута (POST /routes/build) id, slug, title и дата пустые и в ответ не попадают.

type buildRequest struct {
	Objects []string `json:"objects"`
}

type saveRequest struct {
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Objects     []string `json:"objects"` // slug'и объектов в порядке посещения
}

func (in saveRequest) toInput() route.SaveInput {
	return route.SaveInput{Slug: in.Slug, Title: in.Title, Description: in.Description, Objects: in.Objects}
}

type stopResponse struct {
	Position int     `json:"position"`
	ObjectID int64   `json:"objectId"`
	Slug     string  `json:"slug"`
	Title    string  `json:"title"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
}

type routeResponse struct {
	ID          int64           `json:"id,omitempty"`
	Slug        string          `json:"slug,omitempty"`
	Title       string          `json:"title,omitempty"`
	Description string          `json:"description,omitempty"`
	OwnerID     *int64          `json:"ownerId,omitempty"`
	DistanceM   float64         `json:"distanceM"`
	DurationS   float64         `json:"durationS"`
	Geometry    json.RawMessage `json:"geometry,omitempty"`
	Stops       []stopResponse  `json:"stops"`
	CreatedAt   *time.Time      `json:"createdAt,omitempty"`
}

func toRouteResponse(r route.Route) routeResponse {
	out := routeResponse{
		ID: r.ID, Slug: r.Slug, Title: r.Title, Description: r.Description, OwnerID: r.OwnerID,
		DistanceM: r.DistanceM, DurationS: r.DurationS, Geometry: r.Geometry,
		Stops: make([]stopResponse, len(r.Stops)), // пустой список уходит как [], а не null
	}
	if !r.CreatedAt.IsZero() {
		out.CreatedAt = &r.CreatedAt
	}
	for i, st := range r.Stops {
		out.Stops[i] = stopResponse{Position: st.Position, ObjectID: st.ObjectID, Slug: st.Slug, Title: st.Title, Lat: st.Lat, Lng: st.Lng}
	}
	return out
}

func toRouteList(list []route.Route) []routeResponse {
	out := make([]routeResponse, len(list)) // пустой список — [], а не null
	for i, r := range list {
		out[i] = toRouteResponse(r)
	}
	return out
}
