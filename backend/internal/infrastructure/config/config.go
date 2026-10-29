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

	// SMTP для уведомлений. Без SMTP_HOST письма не отправляются, а пишутся в лог.
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string

	// ЮKassa. Без YOOKASSA_SHOP_ID заказы создаются, но оплату принять нельзя.
	YooKassaShopID    string
	YooKassaSecretKey string
	YooKassaReturnURL string // куда ЮKassa вернёт покупателя после оплаты

	PublicURL string // адрес сайта для ссылок в письмах
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

		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     getInt("SMTP_PORT", 587),
		SMTPUser:     os.Getenv("SMTP_USER"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:     getEnv("SMTP_FROM", "noreply@localhost"),

		YooKassaShopID:    os.Getenv("YOOKASSA_SHOP_ID"),
		YooKassaSecretKey: os.Getenv("YOOKASSA_SECRET_KEY"),
		YooKassaReturnURL: getEnv("YOOKASSA_RETURN_URL", "http://localhost:3000/shop/orders"),

		PublicURL: getEnv("PUBLIC_URL", "http://localhost:3000"),
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
