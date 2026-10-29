package app

import (
	"context"
	"net/http"

	"heritage-backend/internal/infrastructure/osrm"
	"heritage-backend/internal/infrastructure/outbox"
	"heritage-backend/internal/infrastructure/postgres"
	"heritage-backend/internal/infrastructure/s3"
	"heritage-backend/internal/infrastructure/security"
	"heritage-backend/internal/infrastructure/smtp"
	"heritage-backend/internal/infrastructure/yookassa"
	repo "heritage-backend/internal/repository/postgres"
	authsvc "heritage-backend/internal/service/auth"
	catalogsvc "heritage-backend/internal/service/catalog"
	mediasvc "heritage-backend/internal/service/media"
	notificationsvc "heritage-backend/internal/service/notification"
	routesvc "heritage-backend/internal/service/route"
	shopsvc "heritage-backend/internal/service/shop"
	submissionsvc "heritage-backend/internal/service/submission"
	httptransport "heritage-backend/internal/transport/http"
	authhttp "heritage-backend/internal/transport/http/auth"
	cataloghttp "heritage-backend/internal/transport/http/catalog"
	mediahttp "heritage-backend/internal/transport/http/media"
	routehttp "heritage-backend/internal/transport/http/route"
	shophttp "heritage-backend/internal/transport/http/shop"
	submissionhttp "heritage-backend/internal/transport/http/submission"
)

// wire создаёт модули и связывает их между собой: репозитории → сервисы → обработчики.
// Возвращает корневой HTTP-обработчик; фоновые задачи добавляет в a.workers.
func (a *App) wire(ctx context.Context) (http.Handler, error) {
	cfg, db := a.cfg, a.pool
	tx := postgres.NewTxManager(db)
	events := outbox.NewPublisher(db) // события пишутся в транзакции сервиса

	// Пользователи: bcrypt и JWT подставляются через порты сервиса.
	authService := authsvc.NewService(repo.NewUsers(db), security.BcryptHasher{},
		security.NewJWTIssuer(cfg.JWTSecret, cfg.JWTTTL))
	if err := authService.EnsureAdmin(ctx, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		return nil, err
	}

	// Каталог — владелец объектов; остальные модули получают сведения об объектах через него.
	catalogService := catalogsvc.NewService(repo.NewObjects(db), repo.NewCategories(db))

	storage, err := s3.New(ctx, cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinIOBucket, cfg.MinIOUseSSL)
	if err != nil {
		return nil, err
	}
	mediaService := mediasvc.NewService(repo.NewMedia(db), storage, catalogService, cfg.MaxUploadMB)
	mediaHandler := mediahttp.NewHandler(mediaService)

	routeService := routesvc.NewService(repo.NewRoutes(db), routePlaces{catalogService},
		routeRouter{osrm.NewClient(cfg.OSRMURL)}, tx)

	submissionService := submissionsvc.NewService(repo.NewSubmissions(db), tx, events,
		submissionCatalog{catalogService}, submissionMedia{mediaService}, submissionUsers{authService})

	// Без YOOKASSA_SHOP_ID клиент не обращается к ЮKassa: заказ создаётся без ссылки на оплату.
	shopService := shopsvc.NewService(shopsvc.Deps{
		Tx:        tx,
		Products:  repo.NewProducts(db),
		Carts:     repo.NewCarts(db),
		Orders:    repo.NewOrders(db),
		Users:     authService,
		Events:    events,
		Payments:  shopPayments{yookassa.NewClient(cfg.YooKassaShopID, cfg.YooKassaSecretKey)},
		ReturnURL: cfg.YooKassaReturnURL,
	})

	// Уведомления: обработчик outbox доставляет события модулю notification,
	// тот превращает их в письма. Без SMTP_HOST письма только пишутся в лог.
	var mailer notificationsvc.Mailer = smtp.LogMailer{}
	if cfg.SMTPHost != "" {
		mailer = smtp.New(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPFrom)
	}
	notificationService := notificationsvc.NewService(notificationUsers{authService}, mailer, cfg.PublicURL)
	processor := outbox.NewProcessor(db)
	for eventType, handle := range notificationService.Handlers() {
		processor.Subscribe(eventType, handle)
	}
	a.workers = append(a.workers, processor.Run)

	return httptransport.New(httptransport.Deps{
		CORSOrigins:   cfg.CORSOrigins,
		Authenticator: authService,
		Health:        db.Ping,
		Auth:          authhttp.NewHandler(authService),
		Catalog:       cataloghttp.NewHandler(catalogService, mediaService, mediaHandler),
		Media:         mediaHandler,
		Route:         routehttp.NewHandler(routeService),
		Submission:    submissionhttp.NewHandler(submissionService, cfg.MaxUploadMB<<20),
		Shop:          shophttp.NewHandler(shopService),
	}), nil
}
