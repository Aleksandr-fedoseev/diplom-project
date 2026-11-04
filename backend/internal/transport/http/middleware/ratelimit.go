package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"heritage-backend/internal/transport/http/httpx"
)

// RateLimiter ограничивает частоту запросов: не больше limit за окно window на один ключ
// (IP-адрес, email). Счётчики живут в памяти процесса — этого достаточно для одного
// экземпляра приложения; при нескольких экземплярах их нужно хранить в общей БД.
type RateLimiter struct {
	limit  int
	window time.Duration
	now    func() time.Time // подменяется в тестах

	mu        sync.Mutex
	hits      map[string]*rateWindow
	lastSweep time.Time
}

type rateWindow struct {
	start time.Time
	count int
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{limit: limit, window: window, now: time.Now, hits: map[string]*rateWindow{}}
}

// Allow засчитывает попытку по ключу. Если лимит исчерпан — false и время до нового окна.
func (l *RateLimiter) Allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	// Раз в окно выбрасываем истёкшие счётчики, чтобы карта не росла бесконечно.
	if now.Sub(l.lastSweep) >= l.window {
		for k, w := range l.hits {
			if now.Sub(w.start) >= l.window {
				delete(l.hits, k)
			}
		}
		l.lastSweep = now
	}

	w, ok := l.hits[key]
	if !ok || now.Sub(w.start) >= l.window {
		l.hits[key] = &rateWindow{start: now, count: 1}
		return true, 0
	}
	if w.count >= l.limit {
		return false, w.start.Add(l.window).Sub(now)
	}
	w.count++
	return true, 0
}

// LimitByIP отвечает 429, если с одного IP-адреса пришло больше запросов, чем разрешает l.
// Адрес берётся из RemoteAddr, который chi RealIP уже заменил адресом из X-Forwarded-For —
// поэтому снаружи бэкенд должен быть доступен только через nginx.
func LimitByIP(l *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ok, retry := l.Allow("ip:" + clientIP(r)); !ok {
				httpx.WriteError(w, r, httpx.TooManyRequests(w, "слишком много запросов, повторите позже", retry))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
