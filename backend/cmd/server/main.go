// Точка входа бэкенда: сигналы ОС, загрузка настроек и запуск приложения.
// Сборка модулей и жизненный цикл — в internal/app.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"heritage-backend/internal/app"
	"heritage-backend/internal/infrastructure/config"
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
	a, err := app.New(ctx, cfg)
	if err != nil {
		return err
	}
	return a.Run(ctx)
}
