package app

import (
	"context"

	"heritage-backend/internal/infrastructure/postgres"
)

// initInfra подключается к БД и применяет миграции. Клиенты S3, SMTP, OSRM и ЮKassa
// создаются при сборке модулей (wiring.go), потому что нужны только им.
func (a *App) initInfra(ctx context.Context) error {
	pool, err := postgres.Connect(ctx, a.cfg.DatabaseURL)
	if err != nil {
		return err
	}
	if err := postgres.Migrate(pool); err != nil {
		pool.Close()
		return err
	}
	a.pool = pool
	return nil
}
