package app

import (
	"context"
	"net/http"

	legacyauth "heritage-backend/internal/auth"
	"heritage-backend/internal/catalog"
	"heritage-backend/internal/infrastructure/security"
	"heritage-backend/internal/media"
	"heritage-backend/internal/notifications"
	"heritage-backend/internal/osrm"
	repo "heritage-backend/internal/repository/postgres"
	"heritage-backend/internal/routes"
	authsvc "heritage-backend/internal/service/auth"
	"heritage-backend/internal/shop"
	"heritage-backend/internal/submissions"
	httptransport "heritage-backend/internal/transport/http"
	authhttp "heritage-backend/internal/transport/http/auth"
)

// wire создаёт модули и связывает их между собой. Возвращает корневой HTTP-обработчик;
// фоновые задачи добавляет в a.workers.
func (a *App) wire(ctx context.Context) (http.Handler, error) {
	cfg, db := a.cfg, a.pool

	// Пользователи: репозиторий → сервис (bcrypt и JWT подставляются через порты) → обработчик.
	authService := authsvc.NewService(repo.NewUsers(db), security.BcryptHasher{},
		security.NewJWTIssuer(cfg.JWTSecret, cfg.JWTTTL))
	if err := authService.EnsureAdmin(ctx, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		return nil, err
	}
	// Старый сервис пользователей нужен submissions и shop, пока они не переехали.
	legacyAuth := legacyauth.NewService(db)

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

	return httptransport.New(httptransport.Deps{
		CORSOrigins:   cfg.CORSOrigins,
		Authenticator: authService,
		Health:        db.Ping,
		Auth:          authhttp.NewHandler(authService),
		Catalog:       catalog.NewHandler(catalogSvc, mediaSvc, mediaHandler),
		Media:         mediaHandler,
		Routes:        routes.NewHandler(routes.NewService(db, catalogSvc, osrm.NewClient(cfg.OSRMURL))),
		Submissions:   submissions.NewHandler(submissions.NewService(db, catalogSvc, mediaSvc, legacyAuth, notifySvc, cfg.PublicURL)),
		// Без YOOKASSA_SHOP_ID клиент не обращается к ЮKassa: заказ создаётся без ссылки на оплату.
		Shop: shop.NewHandler(shop.NewService(db, legacyAuth, notifySvc,
			shop.NewYooKassaClient(cfg.YooKassaShopID, cfg.YooKassaSecretKey), cfg.YooKassaReturnURL)),
	}), nil
}
