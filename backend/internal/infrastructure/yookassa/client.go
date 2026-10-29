// Package yookassa — клиент API ЮKassa (https://yookassa.ru/developers/api).
// Ничего не знает о магазине: принимает и возвращает простые данные платежа.
// Сервис магазина видит его через свой порт PaymentGateway (адаптер — при сборке приложения).
package yookassa

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Статусы платежа в ЮKassa.
const (
	PaymentPending           = "pending"
	PaymentWaitingForCapture = "waiting_for_capture"
	PaymentSucceeded         = "succeeded"
	PaymentCanceled          = "canceled"
)

var (
	// ErrPaymentsDisabled — магазин не подключён к ЮKassa (не задан YOOKASSA_SHOP_ID).
	ErrPaymentsDisabled = errors.New("приём оплаты не настроен")
	// ErrPaymentNotFound — ЮKassa не знает платежа с таким id.
	ErrPaymentNotFound = errors.New("платёж не найден")
)

// PaymentRequest — данные для создания платежа по заказу.
type PaymentRequest struct {
	OrderID       int64
	AmountKopecks int64
	Description   string
	ReturnURL     string // куда ЮKassa вернёт покупателя после оплаты
	// IdempotenceKey — ключ идемпотентности; пустой — "order-<OrderID>".
	IdempotenceKey string
}

// Payment — платёж в ЮKassa (нужные поля).
type Payment struct {
	ID              string
	Status          string
	Amount          string // сумма строкой, как её пишет ЮKassa: "4500.00"
	ConfirmationURL string // страница оплаты; есть только у нового платежа
	OrderID         string // metadata.order_id
}

type Client struct {
	baseURL   string
	shopID    string
	secretKey string
	http      *http.Client
}

// NewClient создаёт клиента. С пустым shopID клиент не обращается к ЮKassa,
// а на все вызовы возвращает ErrPaymentsDisabled: заказы оформляются без оплаты.
func NewClient(shopID, secretKey string) *Client {
	return &Client{
		baseURL:   "https://api.yookassa.ru/v3",
		shopID:    shopID,
		secretKey: secretKey,
		http:      &http.Client{Timeout: 10 * time.Second},
	}
}

// Enabled — клиент настроен (задан shopID).
func (c *Client) Enabled() bool { return c.shopID != "" }

// formatAmount переводит копейки в сумму для ЮKassa: 450000 → "4500.00".
// Считаем целыми числами: у float64 бывают ошибки округления.
func formatAmount(kopecks int64) string {
	return fmt.Sprintf("%d.%02d", kopecks/100, kopecks%100)
}

// paymentResponse — ответ ЮKassa на создание и запрос платежа (нужные нам поля).
type paymentResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Amount struct {
		Value string `json:"value"`
	} `json:"amount"`
	Confirmation struct {
		ConfirmationURL string `json:"confirmation_url"`
	} `json:"confirmation"`
	Metadata struct {
		OrderID string `json:"order_id"`
	} `json:"metadata"`
}

// CreatePayment создаёт платёж с одностадийной оплатой (capture: true — деньги списываются сразу).
// Если ответ потерялся и запрос повторили с тем же ключом идемпотентности,
// ЮKassa вернёт тот же платёж, а не создаст второй.
func (c *Client) CreatePayment(ctx context.Context, req PaymentRequest) (Payment, error) {
	if !c.Enabled() {
		return Payment{}, ErrPaymentsDisabled
	}
	orderID := strconv.FormatInt(req.OrderID, 10)
	key := req.IdempotenceKey
	if key == "" {
		key = "order-" + orderID
	}
	body := map[string]any{
		"amount":       map[string]string{"value": formatAmount(req.AmountKopecks), "currency": "RUB"},
		"capture":      true,
		"confirmation": map[string]string{"type": "redirect", "return_url": req.ReturnURL},
		"description":  req.Description,
		"metadata":     map[string]string{"order_id": orderID},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return Payment{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/payments", bytes.NewReader(payload))
	if err != nil {
		return Payment{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotence-Key", key)
	return c.do(httpReq)
}

// GetPayment запрашивает платёж по id — так магазин узнаёт его настоящий статус.
func (c *Client) GetPayment(ctx context.Context, id string) (Payment, error) {
	if !c.Enabled() {
		return Payment{}, ErrPaymentsDisabled
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/payments/"+url.PathEscape(id), nil)
	if err != nil {
		return Payment{}, err
	}
	return c.do(httpReq)
}

func (c *Client) do(req *http.Request) (Payment, error) {
	req.SetBasicAuth(c.shopID, c.secretKey)
	resp, err := c.http.Do(req)
	if err != nil {
		return Payment{}, fmt.Errorf("запрос к ЮKassa: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return Payment{}, ErrPaymentNotFound
	}
	if resp.StatusCode != http.StatusOK {
		// В теле ошибки ЮKassa пишет code и description — пригодится в логе.
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return Payment{}, fmt.Errorf("ЮKassa ответила HTTP %d: %s", resp.StatusCode, msg)
	}
	var body paymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return Payment{}, fmt.Errorf("ответ ЮKassa: %w", err)
	}
	return Payment{
		ID:              body.ID,
		Status:          body.Status,
		Amount:          body.Amount.Value,
		ConfirmationURL: body.Confirmation.ConfirmationURL,
		OrderID:         body.Metadata.OrderID,
	}, nil
}
