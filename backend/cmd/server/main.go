// Точка входа бэкенда: читает настройки, подключается к БД и MinIO, запускает HTTP-сервер.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"heritage-backend/internal/auth"
	"heritage-backend/internal/catalog"
	"heritage-backend/internal/config"
	"heritage-backend/internal/database"
	"heritage-backend/internal/media"
	"heritage-backend/internal/osrm"
	"heritage-backend/internal/routes"
	"heritage-backend/internal/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("сервер остановлен с ошибкой", "err", err)
		os.Exit(1)
	}
}

func run() error {
	// ctx отменяется по Ctrl+C или по сигналу остановки от Docker.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		return err
	}

	// Собираем модули: каждый получает только то, что ему нужно.
	tokens := auth.NewTokenIssuer(cfg.JWTSecret, cfg.JWTTTL)
	authSvc := auth.NewService(db, tokens)
	if err := authSvc.EnsureAdmin(ctx, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		return err
	}

	mediaSvc, err := media.NewService(ctx, db, cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey,
		cfg.MinIOBucket, cfg.MinIOUseSSL, cfg.MaxUploadMB)
	if err != nil {
		return err
	}
	mediaHandler := media.NewHandler(mediaSvc)

	handler := server.New(server.Deps{
		DB:          db,
		Tokens:      tokens,
		CORSOrigins: cfg.CORSOrigins,
		Auth:        auth.NewHandler(authSvc),
		Catalog:     catalog.NewHandler(catalog.NewRepository(db), mediaSvc, mediaHandler),
		Media:       mediaHandler,
		Routes:      routes.NewHandler(routes.NewService(db, osrm.NewClient(cfg.OSRMURL))),
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Сервер работает в отдельной горутине, а main ждёт сигнала остановки.
	errCh := make(chan error, 1)
	go func() {
		slog.Info("сервер запущен", "addr", cfg.HTTPAddr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		slog.Info("получен сигнал остановки, завершаем текущие запросы")
	}

	// Даём текущим запросам до 15 секунд на завершение.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
