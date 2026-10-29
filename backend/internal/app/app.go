// Package app собирает приложение и управляет его жизненным циклом.
// Это единственное место, где создаются конкретные типы: пул БД, клиенты внешних
// сервисов, репозитории, сервисы и обработчики. Остальные пакеты знают друг о друге
// только через интерфейсы.
package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"heritage-backend/internal/infrastructure/config"
)

// worker — фоновая задача, которая работает, пока не отменён ctx (например, отправка писем).
type worker func(ctx context.Context)

type App struct {
	cfg     config.Config
	pool    *pgxpool.Pool
	server  *http.Server
	workers []worker
}

// New подключается к инфраструктуре и собирает модули. Ошибка — приложение запускать нельзя.
func New(ctx context.Context, cfg config.Config) (*App, error) {
	a := &App{cfg: cfg}
	if err := a.initInfra(ctx); err != nil {
		return nil, err
	}
	handler, err := a.wire(ctx)
	if err != nil {
		a.pool.Close()
		return nil, err
	}
	a.server = newHTTPServer(cfg.HTTPAddr, handler)
	return a, nil
}

// Run запускает HTTP-сервер и фоновые задачи и ждёт отмены ctx (сигнал ОС)
// или ошибки сервера. Затем корректно останавливает всё и закрывает пул БД.
func (a *App) Run(ctx context.Context) error {
	defer a.pool.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	for _, w := range a.workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w(ctx)
		}()
	}

	err := serve(ctx, a.server)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("HTTP-сервер остановлен с ошибкой", "err", err)
	}

	// Сервер остановлен: отменяем фоновые задачи и дожидаемся, пока они допишут текущую работу.
	cancel()
	wg.Wait()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
