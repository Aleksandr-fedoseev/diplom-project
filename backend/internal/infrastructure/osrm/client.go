// Package osrm — клиент маршрутизатора OSRM: строит пеший маршрут по улицам через заданные точки.
package osrm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Point struct {
	Lat float64
	Lng float64
}

type Route struct {
	DistanceM float64         // длина в метрах
	DurationS float64         // время в пути в секундах
	Geometry  json.RawMessage // GeoJSON LineString — линия, которую рисует карта
}

type Client struct {
	baseURL string
	http    *http.Client
}

// baseURL — адрес без хвоста /route/v1/..., например:
// http://osrm:5000 (свой контейнер) или https://routing.openstreetmap.de/routed-foot (публичный).
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// Route строит маршрут, проходящий через точки в заданном порядке.
func (c *Client) Route(ctx context.Context, points []Point) (Route, error) {
	if len(points) < 2 {
		return Route{}, fmt.Errorf("нужно минимум 2 точки")
	}

	// OSRM ждёт координаты в порядке «долгота,широта», точки разделены «;».
	coords := make([]string, len(points))
	for i, p := range points {
		coords[i] = fmt.Sprintf("%.6f,%.6f", p.Lng, p.Lat)
	}
	// Имя профиля в адресе osrm-routed не проверяет: профиль (пеший) задаётся при подготовке данных.
	url := fmt.Sprintf("%s/route/v1/foot/%s?overview=full&geometries=geojson", c.baseURL, strings.Join(coords, ";"))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Route{}, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return Route{}, fmt.Errorf("запрос к OSRM: %w", err)
	}
	defer resp.Body.Close()

	var body struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Routes  []struct {
			Distance float64         `json:"distance"`
			Duration float64         `json:"duration"`
			Geometry json.RawMessage `json:"geometry"`
		} `json:"routes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return Route{}, fmt.Errorf("ответ OSRM (HTTP %d): %w", resp.StatusCode, err)
	}
	if body.Code != "Ok" || len(body.Routes) == 0 {
		return Route{}, fmt.Errorf("OSRM: %s %s", body.Code, body.Message)
	}

	r := body.Routes[0]
	return Route{DistanceM: r.Distance, DurationS: r.Duration, Geometry: r.Geometry}, nil
}
