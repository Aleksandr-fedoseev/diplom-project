package app

import (
	"context"
	"net/http"

	"heritage-backend/internal/auth"
	"heritage-backend/internal/catalog"
	"heritage-backend/internal/media"
	"heritage-backend/internal/notifications"
	"heritage-backend/internal/osrm"
	"heritage-backend/internal/routes"
	"heritage-backend/internal/server"
	"heritage-backend/internal/shop"
	"heritage-backend/internal/submissions"
)

// wire создаёт модули и связывает их между собой. Возвращает корневой HTTP-обработчик;
// фоновые задачи добавляет в a.workers.
func (a *App) wire(ctx context.Context) (http.Handler, error) {
	cfg, db := a.cfg, a.pool

	tokens := auth.NewTokenIssuer(cfg.JWTSecret, cfg.JWTTTL)
	authSvc := auth.NewService(db, tokens)
	if err := authSvc.EnsureAdmin(ctx, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		return nil, err
	}

	// Письма уходят через SMTP, а без него (при разработке) — только в лог.
	var sender notifications.Sender = notifications.LogSender{}
	if cfg.SMTPHost != "" {
		sender = notifications.NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPFrom)
	}
	notifySvc := notifications.NewService(db, sender)
	a.workers = append(a.workers, notifySvc.Run)

	// Каталог — владелец объектов; медиафайлы и маршруты получают сведения об объектах через него.
	catalogSvc := catalog.NewService(db)

	mediaSvc, err := media.NewService(ctx, db, catalogSvc, cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey,
		cfg.MinIOBucket, cfg.MinIOUseSSL, cfg.MaxUploadMB)
	if err != nil {
		return nil, err
	}
	mediaHandler := media.NewHandler(mediaSvc)

	return server.New(server.Deps{
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
	}), nil
}
