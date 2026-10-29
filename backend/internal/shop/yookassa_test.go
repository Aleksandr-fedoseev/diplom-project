package shop

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFormatAmount(t *testing.T) {
	cases := map[int64]string{450000: "4500.00", 5: "0.05", 100: "1.00", 199: "1.99", 12345678: "123456.78", 0: "0.00"}
	for kopecks, want := range cases {
		if got := formatAmount(kopecks); got != want {
			t.Errorf("formatAmount(%d) = %q, ожидалось %q", kopecks, got, want)
		}
	}
}

func TestYooKassaClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "123456" || pass != "test_secret" {
			t.Errorf("Basic auth: %q %q %v", user, pass, ok)
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/payments":
			if got := r.Header.Get("Idempotence-Key"); got != "order-42" {
				t.Errorf("Idempotence-Key = %q", got)
			}
			var body struct {
				Amount struct {
					Value    string `json:"value"`
					Currency string `json:"currency"`
				} `json:"amount"`
				Capture      bool `json:"capture"`
				Confirmation struct {
					Type      string `json:"type"`
					ReturnURL string `json:"return_url"`
				} `json:"confirmation"`
				Description string            `json:"description"`
				Metadata    map[string]string `json:"metadata"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("тело запроса: %v", err)
			}
			if body.Amount.Value != "4500.00" || body.Amount.Currency != "RUB" || !body.Capture ||
				body.Confirmation.Type != "redirect" || body.Confirmation.ReturnURL != "https://site.test/orders?order=42" ||
				body.Description != "Заказ №42" || body.Metadata["order_id"] != "42" {
				t.Errorf("тело запроса: %+v", body)
			}
			_, _ = w.Write([]byte(`{"id":"2d0f6b3e-000f-5000-8000-1a2b3c4d5e6f","status":"pending","paid":false,
				"amount":{"value":"4500.00","currency":"RUB"},
				"confirmation":{"type":"redirect","confirmation_url":"https://yoomoney.ru/checkout/payments/v2/contract?orderId=2d0f"},
				"metadata":{"order_id":"42"}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/payments/2d0f6b3e-000f-5000-8000-1a2b3c4d5e6f":
			_, _ = w.Write([]byte(`{"id":"2d0f6b3e-000f-5000-8000-1a2b3c4d5e6f","status":"succeeded","paid":true,
				"amount":{"value":"4500.00","currency":"RUB"},"metadata":{"order_id":"42"}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/payments/unknown":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"type":"error","code":"not_found"}`))
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	c := NewYooKassaClient("123456", "test_secret")
	c.baseURL = srv.URL
	ctx := context.Background()

	p, err := c.CreatePayment(ctx, PaymentRequest{
		OrderID: 42, AmountKopecks: 450000, Description: "Заказ №42", ReturnURL: "https://site.test/orders?order=42",
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.ID != "2d0f6b3e-000f-5000-8000-1a2b3c4d5e6f" || p.Status != PaymentPending ||
		p.ConfirmationURL != "https://yoomoney.ru/checkout/payments/v2/contract?orderId=2d0f" || p.OrderID != "42" {
		t.Fatalf("созданный платёж: %+v", p)
	}

	p, err = c.GetPayment(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != PaymentSucceeded || p.Amount != "4500.00" || p.OrderID != "42" {
		t.Fatalf("запрошенный платёж: %+v", p)
	}

	if _, err := c.GetPayment(ctx, "unknown"); !errors.Is(err, ErrPaymentNotFound) {
		t.Fatalf("неизвестный платёж: %v", err)
	}
	if _, err := c.GetPayment(ctx, "broken"); err == nil || errors.Is(err, ErrPaymentNotFound) {
		t.Fatalf("ошибка сервера должна быть временной ошибкой: %v", err)
	}

	// Без shopId клиент к ЮKassa не обращается.
	if _, err := NewYooKassaClient("", "").CreatePayment(ctx, PaymentRequest{OrderID: 1, AmountKopecks: 100}); !errors.Is(err, ErrPaymentsDisabled) {
		t.Fatalf("ненастроенный клиент: %v", err)
	}
}
