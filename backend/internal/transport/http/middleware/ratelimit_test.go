package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterWindow(t *testing.T) {
	l := NewRateLimiter(2, time.Minute)
	now := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	l.now = func() time.Time { return now }

	for i := range 2 {
		if ok, _ := l.Allow("a"); !ok {
			t.Fatalf("попытка %d в пределах лимита отклонена", i+1)
		}
	}
	ok, retry := l.Allow("a")
	if ok || retry != time.Minute {
		t.Fatalf("третья попытка: ok=%v retry=%v", ok, retry)
	}
	if ok, _ := l.Allow("b"); !ok {
		t.Fatal("у другого ключа свой счётчик")
	}

	now = now.Add(time.Minute)
	if ok, _ := l.Allow("a"); !ok {
		t.Fatal("в новом окне попытки снова разрешены")
	}
}

func TestLimitByIP(t *testing.T) {
	h := LimitByIP(NewRateLimiter(1, time.Minute))(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	call := func(addr string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
		req.RemoteAddr = addr
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}

	if rec := call("203.0.113.5:1000"); rec.Code != http.StatusNoContent {
		t.Fatalf("первый запрос: %d", rec.Code)
	}
	rec := call("203.0.113.5:2000") // тот же IP с другого порта
	if rec.Code != http.StatusTooManyRequests || rec.Header().Get("Retry-After") != "60" {
		t.Fatalf("второй запрос: %d, Retry-After=%q", rec.Code, rec.Header().Get("Retry-After"))
	}
	if rec := call("203.0.113.6:1000"); rec.Code != http.StatusNoContent {
		t.Fatalf("другой IP: %d", rec.Code)
	}
}
