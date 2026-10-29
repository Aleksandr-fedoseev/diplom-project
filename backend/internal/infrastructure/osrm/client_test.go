package osrm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Клиент проверяется на поддельном сервере OSRM: формат запроса и разбор ответа.
func TestClientRoute(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if strings.Contains(r.URL.Path, "0.000000,0.000000") {
			_, _ = w.Write([]byte(`{"code":"NoRoute","message":"нет пути"}`))
			return
		}
		_, _ = w.Write([]byte(`{"code":"Ok","routes":[{"distance":420.5,"duration":300,
			"geometry":{"type":"LineString","coordinates":[[48.03,46.35],[48.02,46.34]]}}]}`))
	}))
	defer srv.Close()
	c := NewClient(srv.URL + "/")

	r, err := c.Route(context.Background(), []Point{{Lat: 46.35, Lng: 48.03}, {Lat: 46.34, Lng: 48.02}})
	if err != nil || r.DistanceM != 420.5 || r.DurationS != 300 || len(r.Geometry) == 0 {
		t.Fatalf("маршрут: %+v, %v", r, err)
	}
	// OSRM ждёт «долгота,широта».
	if gotPath != "/route/v1/foot/48.030000,46.350000;48.020000,46.340000" {
		t.Fatalf("адрес запроса: %s", gotPath)
	}

	if _, err := c.Route(context.Background(), []Point{{}, {}}); err == nil {
		t.Fatal("ожидалась ошибка NoRoute")
	}
	if _, err := c.Route(context.Background(), []Point{{}}); err == nil {
		t.Fatal("ожидалась ошибка для одной точки")
	}
}
