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
	"heritage-backend/internal/notifications"
	"heritage-backend/internal/osrm"
	"heritage-backend/internal/routes"
	"heritage-backend/internal/server"
	"heritage-backend/internal/shop"
	"heritage-backend/internal/submissions"
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

	// Письма уходят через SMTP, а без него (при разработке) — только в лог.
	var sender notifications.Sender = notifications.LogSender{}
	if cfg.SMTPHost != "" {
		sender = notifications.NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPFrom)
	}
	notifySvc := notifications.NewService(db, sender)
	// Фоновый обработчик очереди писем; останавливается вместе с сервером по ctx.
	workerDone := make(chan struct{})
	go func() {
		defer close(workerDone)
		notifySvc.Run(ctx)
	}()
	defer func() { stop(); <-workerDone }() // дожидаемся, пока обработчик допишет текущее письмо

	// Каталог — владелец объектов; медиафайлы и маршруты получают сведения об объектах через него.
	catalogSvc := catalog.NewService(db)

	mediaSvc, err := media.NewService(ctx, db, catalogSvc, cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey,
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
		Catalog:     catalog.NewHandler(catalogSvc, mediaSvc, mediaHandler),
		Media:       mediaHandler,
		Routes:      routes.NewHandler(routes.NewService(db, catalogSvc, osrm.NewClient(cfg.OSRMURL))),
		Submissions: submissions.NewHandler(submissions.NewService(db, catalogSvc, mediaSvc, authSvc, notifySvc, cfg.PublicURL)),
		// Без YOOKASSA_SHOP_ID клиент не обращается к ЮKassa: заказ создаётся без ссылки на оплату.
		Shop: shop.NewHandler(shop.NewService(db, authSvc, notifySvc,
			shop.NewYooKassaClient(cfg.YooKassaShopID, cfg.YooKassaSecretKey), cfg.YooKassaReturnURL)),
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
