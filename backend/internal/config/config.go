// Package config читает настройки приложения из переменных окружения.
// Все настройки собраны в одной структуре, чтобы было видно, что нужно для запуска.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr    string   // адрес, на котором слушает сервер, например ":8080"
	DatabaseURL string   // строка подключения к PostgreSQL
	CORSOrigins []string // откуда фронтенду разрешено обращаться к API

	JWTSecret string
	JWTTTL    time.Duration

	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIOUseSSL    bool

	MaxUploadMB int64 // общий предел размера загружаемого файла

	OSRMURL string // базовый адрес OSRM, например http://osrm:5000

	// Первый администратор создаётся при старте, если такого email ещё нет.
	AdminEmail    string
	AdminPassword string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:       getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		CORSOrigins:    strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000"), ","),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		JWTTTL:         getDuration("JWT_TTL", 24*time.Hour),
		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey: os.Getenv("MINIO_ACCESS_KEY"),
		MinIOSecretKey: os.Getenv("MINIO_SECRET_KEY"),
		MinIOBucket:    getEnv("MINIO_BUCKET", "media"),
		MinIOUseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",
		MaxUploadMB:    int64(getInt("MAX_UPLOAD_MB", 50)),
		OSRMURL:        getEnv("OSRM_URL", "https://routing.openstreetmap.de/routed-foot"),
		AdminEmail:     os.Getenv("ADMIN_EMAIL"),
		AdminPassword:  os.Getenv("ADMIN_PASSWORD"),
	}

	// Без этих значений приложение работать не может — лучше упасть сразу при старте.
	if cfg.DatabaseURL == "" {
		return cfg, fmt.Errorf("DATABASE_URL не задан")
	}
	if len(cfg.JWTSecret) < 32 {
		return cfg, fmt.Errorf("JWT_SECRET должен быть не короче 32 символов")
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return v
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	if v, err := time.ParseDuration(os.Getenv(key)); err == nil {
		return v
	}
	return def
}
