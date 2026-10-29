// Package httptransport собирает HTTP-обработчики всех модулей в один маршрутизатор.
// (Имя пакета не «http», чтобы не путать со стандартным net/http.)
package httptransport

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"heritage-backend/internal/catalog"
	"heritage-backend/internal/media"
	"heritage-backend/internal/routes"
	"heritage-backend/internal/shop"
	"heritage-backend/internal/submissions"
	authhttp "heritage-backend/internal/transport/http/auth"
	"heritage-backend/internal/transport/http/httpx"
	"heritage-backend/internal/transport/http/middleware"
)

type Deps struct {
	CORSOrigins   []string
	Authenticator middleware.Authenticator
	// Health проверяет, что приложение может работать (БД доступна).
	Health func(ctx context.Context) error

	Auth *authhttp.Handler

	// Модули, которые ещё не переехали на новую раскладку.
	Catalog     *catalog.Handler
	Media       *media.Handler
	Routes      *routes.Handler
	Submissions *submissions.Handler
	Shop        *shop.Handler
}

func New(d Deps) http.Handler {
	r := chi.NewRouter()

	// Middleware выполняются по порядку для каждого запроса.
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer) // паника в обработчике не роняет весь сервер
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: d.CORSOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	}))
	r.Use(middleware.Authenticate(d.Authenticator))

	r.Route("/api", func(r chi.Router) {
		r.Method(http.MethodGet, "/health", httpx.Handler(func(w http.ResponseWriter, r *http.Request) error {
			if err := d.Health(r.Context()); err != nil {
				httpx.JSON(w, http.StatusServiceUnavailable, map[string]string{"error": "БД недоступна"})
				return nil
			}
			httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
			return nil
		}))

		r.Route("/auth", d.Auth.Routes)
		r.Route("/admin/users", d.Auth.AdminRoutes)

		r.Route("/media", d.Media.Routes)
		d.Catalog.Routes(r)
		d.Routes.Routes(r)
		d.Submissions.Routes(r)
		d.Shop.Routes(r)
	})
	return r
}
