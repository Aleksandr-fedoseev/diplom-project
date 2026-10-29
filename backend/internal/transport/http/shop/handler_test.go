package shop

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"heritage-backend/internal/domain/auth"
	"heritage-backend/internal/domain/shop"
	"heritage-backend/internal/transport/http/middleware"
)

// fakeService — встроенный интерфейс даёт все методы; тест переопределяет нужные.
// Вызов непереопределённого метода паникует — значит, обработчик позвал лишнее.
type fakeService struct {
	shop.Service
	viewer     shop.Viewer
	webhookErr error
	payErr     error
}

func (f *fakeService) GetOrder(_ context.Context, id int64, v shop.Viewer) (shop.Order, error) {
	f.viewer = v
	uid := int64(7)
	return shop.Order{ID: id, UserID: &uid, Status: shop.StatusPending, TotalKopecks: 450000,
		Items: []shop.OrderItem{{ProductID: 1, Title: "Худи", Quantity: 2, PriceKopecks: 225000}}}, nil
}

func (f *fakeService) Pay(context.Context, int64, int64) (shop.Order, error) {
	return shop.Order{}, f.payErr
}

func (f *fakeService) ListMyOrders(context.Context, int64) ([]shop.Order, error) { return nil, nil }

func (f *fakeService) HandlePaymentNotification(context.Context, string) error { return f.webhookErr }

func serve(t *testing.T, svc shop.Service, p *auth.Principal, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	r := chi.NewRouter()
	NewHandler(svc).Routes(r)
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if p != nil {
		req = req.WithContext(middleware.WithPrincipal(req.Context(), *p))
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestHandlerAccess(t *testing.T) {
	svc := &fakeService{}
	user := &auth.Principal{UserID: 7, Role: auth.RoleUser}
	if rec := serve(t, svc, nil, http.MethodGet, "/cart", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("корзина без входа: %d", rec.Code)
	}
	if rec := serve(t, svc, user, http.MethodPost, "/products", `{}`); rec.Code != http.StatusForbidden {
		t.Fatalf("создание товара пользователем: %d", rec.Code)
	}
	if rec := serve(t, svc, user, http.MethodGet, "/admin/orders", ""); rec.Code != http.StatusForbidden {
		t.Fatalf("заказы администратора: %d", rec.Code)
	}
}

func TestHandlerOrderJSON(t *testing.T) {
	svc := &fakeService{}
	admin := &auth.Principal{UserID: 1, Role: auth.RoleAdmin}
	rec := serve(t, svc, admin, http.MethodGet, "/orders/5", "")
	if rec.Code != http.StatusOK || svc.viewer != (shop.Viewer{UserID: 1, IsAdmin: true}) {
		t.Fatalf("заказ: %d, viewer %+v", rec.Code, svc.viewer)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"id", "userId", "status", "totalKopecks", "paymentUrl", "deliveryAddress", "phone", "items", "createdAt", "updatedAt"} {
		if _, ok := body[key]; !ok {
			t.Errorf("в ответе нет поля %s", key)
		}
	}
	item := body["items"].([]any)[0].(map[string]any)
	if item["lineTotalKopecks"] != float64(450000) {
		t.Fatalf("позиция: %+v", item)
	}

	// Пустой список — [], а не null.
	rec = serve(t, svc, admin, http.MethodGet, "/me/orders", "")
	if strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Fatalf("пустой список: %s", rec.Body.String())
	}
	if rec := serve(t, svc, admin, http.MethodGet, "/orders/abc", ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("некорректный id: %d", rec.Code)
	}
}

func TestHandlerPayUnavailable(t *testing.T) {
	user := &auth.Principal{UserID: 7, Role: auth.RoleUser}
	for _, err := range []error{shop.ErrPaymentsDisabled, shop.ErrPaymentFailed} {
		rec := serve(t, &fakeService{payErr: err}, user, http.MethodPost, "/orders/5/pay", "")
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("%v: HTTP %d", err, rec.Code)
		}
	}
}

func TestHandlerWebhook(t *testing.T) {
	svc := &fakeService{}
	if rec := serve(t, svc, nil, http.MethodPost, "/payments/yookassa/webhook", `{"event":"payment.succeeded","object":{"id":"p1"}}`); rec.Code != http.StatusOK {
		t.Fatalf("уведомление: %d", rec.Code)
	}
	if rec := serve(t, svc, nil, http.MethodPost, "/payments/yookassa/webhook", `{}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("без id: %d", rec.Code)
	}
	svc.webhookErr = errors.New("ЮKassa недоступна")
	if rec := serve(t, svc, nil, http.MethodPost, "/payments/yookassa/webhook", `{"object":{"id":"p1"}}`); rec.Code != http.StatusInternalServerError {
		t.Fatalf("временный сбой должен давать 5xx: %d", rec.Code)
	}
}
