package shop

import (
	"time"

	"heritage-backend/internal/domain/shop"
)

// Формат JSON — как в старом модуле shop: фронтенд на него опирается.
// Деньги — целые копейки (priceKopecks, totalKopecks).

type productRequest struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	PriceKopecks int64  `json:"priceKopecks"`
	Stock        int    `json:"stock"`
	ImageURL     string `json:"imageUrl"`
	IsActive     *bool  `json:"isActive"` // по умолчанию true
}

func (in productRequest) toInput() shop.ProductInput {
	return shop.ProductInput{
		Slug: in.Slug, Title: in.Title, Description: in.Description, Category: in.Category,
		PriceKopecks: in.PriceKopecks, Stock: in.Stock, ImageURL: in.ImageURL, IsActive: in.IsActive,
	}
}

type productResponse struct {
	ID           int64     `json:"id"`
	Slug         string    `json:"slug"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Category     string    `json:"category"`
	PriceKopecks int64     `json:"priceKopecks"`
	Stock        int       `json:"stock"`
	InStock      bool      `json:"inStock"`
	ImageURL     string    `json:"imageUrl"`
	IsActive     bool      `json:"isActive"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func toProductResponse(p shop.Product) productResponse {
	return productResponse{
		ID: p.ID, Slug: p.Slug, Title: p.Title, Description: p.Description, Category: p.Category,
		PriceKopecks: p.PriceKopecks, Stock: p.Stock, InStock: p.InStock(), ImageURL: p.ImageURL,
		IsActive: p.IsActive, UpdatedAt: p.UpdatedAt,
	}
}

type quantityRequest struct {
	Quantity int `json:"quantity"`
}

type cartItemResponse struct {
	ProductID        int64  `json:"productId"`
	Slug             string `json:"slug"`
	Title            string `json:"title"`
	ImageURL         string `json:"imageUrl"`
	PriceKopecks     int64  `json:"priceKopecks"`
	Quantity         int    `json:"quantity"`
	Stock            int    `json:"stock"`
	LineTotalKopecks int64  `json:"lineTotalKopecks"`
	Available        bool   `json:"available"` // товар в продаже и его хватает
}

type cartResponse struct {
	Items        []cartItemResponse `json:"items"`
	TotalKopecks int64              `json:"totalKopecks"`
}

func toCartResponse(c shop.Cart) cartResponse {
	out := cartResponse{Items: make([]cartItemResponse, len(c.Items)), TotalKopecks: c.Total()}
	for i, it := range c.Items {
		out.Items[i] = cartItemResponse{
			ProductID: it.ProductID, Slug: it.Slug, Title: it.Title, ImageURL: it.ImageURL,
			PriceKopecks: it.PriceKopecks, Quantity: it.Quantity, Stock: it.Stock,
			LineTotalKopecks: it.LineTotal(), Available: it.Available(),
		}
	}
	return out
}

type checkoutRequest struct {
	DeliveryAddress string `json:"deliveryAddress"`
	Phone           string `json:"phone"`
}

type statusRequest struct {
	Status string `json:"status"`
}

type orderItemResponse struct {
	ProductID        int64  `json:"productId"`
	Title            string `json:"title"`
	Quantity         int    `json:"quantity"`
	PriceKopecks     int64  `json:"priceKopecks"`
	LineTotalKopecks int64  `json:"lineTotalKopecks"`
}

type orderResponse struct {
	ID              int64               `json:"id"`
	UserID          *int64              `json:"userId,omitempty"`
	Status          string              `json:"status"`
	TotalKopecks    int64               `json:"totalKopecks"`
	PaymentURL      string              `json:"paymentUrl"` // страница оплаты ЮKassa; пусто, если платёж не создан
	DeliveryAddress string              `json:"deliveryAddress"`
	Phone           string              `json:"phone"`
	Items           []orderItemResponse `json:"items"`
	CreatedAt       time.Time           `json:"createdAt"`
	UpdatedAt       time.Time           `json:"updatedAt"`
}

func toOrderResponse(o shop.Order) orderResponse {
	out := orderResponse{
		ID: o.ID, UserID: o.UserID, Status: string(o.Status), TotalKopecks: o.TotalKopecks,
		PaymentURL: o.PaymentURL, DeliveryAddress: o.DeliveryAddress, Phone: o.Phone,
		Items: make([]orderItemResponse, len(o.Items)), CreatedAt: o.CreatedAt, UpdatedAt: o.UpdatedAt,
	}
	for i, it := range o.Items {
		out.Items[i] = orderItemResponse{
			ProductID: it.ProductID, Title: it.Title, Quantity: it.Quantity,
			PriceKopecks: it.PriceKopecks, LineTotalKopecks: it.LineTotal(),
		}
	}
	return out
}

func toOrderList(list []shop.Order) []orderResponse {
	out := make([]orderResponse, len(list))
	for i, o := range list {
		out[i] = toOrderResponse(o)
	}
	return out
}

// webhookNotification — уведомление ЮKassa. Нужен только object.id: статус платежа
// сервис перепроверяет у ЮKassa, телу уведомления не доверяет.
type webhookNotification struct {
	Event  string `json:"event"`
	Object struct {
		ID string `json:"id"`
	} `json:"object"`
}
