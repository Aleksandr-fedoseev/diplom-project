// Package server собирает все модули монолита в один HTTP-маршрутизатор.
package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"heritage-backend/internal/auth"
	"heritage-backend/internal/catalog"
	"heritage-backend/internal/httpx"
	"heritage-backend/internal/media"
	"heritage-backend/internal/routes"
	"heritage-backend/internal/shop"
	"heritage-backend/internal/submissions"
)

type Deps struct {
	DB          *pgxpool.Pool
	Tokens      *auth.TokenIssuer
	CORSOrigins []string

	Auth        *auth.Handler
	Catalog     *catalog.Handler
	Media       *media.Handler
	Routes      *routes.Handler
	Submissions *submissions.Handler
	Shop        *shop.Handler
}

func New(d Deps) http.Handler {
	r := chi.NewRouter()

	// Middleware выполняются по порядку для каждого запроса.
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer) // паника в обработчике не роняет весь сервер
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: d.CORSOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	}))
	r.Use(auth.Authenticate(d.Tokens))

	r.Route("/api", func(r chi.Router) {
		r.Method(http.MethodGet, "/health", httpx.Handler(func(w http.ResponseWriter, r *http.Request) error {
			if err := d.DB.Ping(r.Context()); err != nil {
				return &httpx.Error{Status: http.StatusServiceUnavailable, Message: "БД недоступна"}
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
