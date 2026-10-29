// Package shop — HTTP-обработчики магазина: товары, корзина, заказы, оплата и уведомления ЮKassa.
package shop

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"heritage-backend/internal/domain/auth"
	"heritage-backend/internal/domain/shop"
	"heritage-backend/internal/transport/http/httpx"
	"heritage-backend/internal/transport/http/middleware"
)

type Handler struct {
	svc shop.Service
}

func NewHandler(svc shop.Service) *Handler { return &Handler{svc: svc} }

// Routes подключается по адресу /api.
func (h *Handler) Routes(r chi.Router) {
	// Каталог товаров — для всех.
	r.Method(http.MethodGet, "/products", httpx.Handler(h.listProducts))
	r.Method(http.MethodGet, "/products/{slug}", httpx.Handler(h.getProduct))

	// Уведомления ЮKassa — без входа: ЮKassa токена не присылает.
	r.Method(http.MethodPost, "/payments/yookassa/webhook", httpx.Handler(h.webhook))

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(auth.RoleUser))
		r.Method(http.MethodGet, "/cart", httpx.Handler(h.getCart))
		r.Method(http.MethodPut, "/cart/items/{productId}", httpx.Handler(h.setCartItem))
		r.Method(http.MethodDelete, "/cart/items/{productId}", httpx.Handler(h.removeCartItem))

		r.Method(http.MethodPost, "/orders", httpx.Handler(h.checkout))
		r.Method(http.MethodGet, "/me/orders", httpx.Handler(h.myOrders))
		r.Method(http.MethodGet, "/orders/{id}", httpx.Handler(h.getOrder))
		r.Method(http.MethodPost, "/orders/{id}/pay", httpx.Handler(h.pay))
		r.Method(http.MethodPost, "/orders/{id}/cancel", httpx.Handler(h.cancel))
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(auth.RoleAdmin))
		r.Method(http.MethodPost, "/products", httpx.Handler(h.createProduct))
		r.Method(http.MethodPut, "/products/{id}", httpx.Handler(h.updateProduct))
		r.Method(http.MethodDelete, "/products/{id}", httpx.Handler(h.deleteProduct))

		r.Method(http.MethodGet, "/admin/orders", httpx.Handler(h.listOrders))
		r.Method(http.MethodPut, "/admin/orders/{id}/status", httpx.Handler(h.setStatus))
	})
}

func isAdmin(r *http.Request) bool {
	p, ok := middleware.PrincipalFrom(r.Context())
	return ok && p.Role.AtLeast(auth.RoleAdmin)
}

// userID — id вошедшего пользователя (маршрут уже защищён RequireRole).
func userID(r *http.Request) int64 {
	p, _ := middleware.PrincipalFrom(r.Context())
	return p.UserID
}

// Товары.

// listProducts: GET /api/products?category=clothes
// Администратор с ?all=true видит и снятые с продажи товары.
func (h *Handler) listProducts(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	products, err := h.svc.ListProducts(r.Context(), q.Get("category"), isAdmin(r) && q.Get("all") == "true")
	if err != nil {
		return err
	}
	out := make([]productResponse, len(products))
	for i, p := range products {
		out[i] = toProductResponse(p)
	}
	httpx.JSON(w, http.StatusOK, out)
	return nil
}

func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request) error {
	p, err := h.svc.GetProduct(r.Context(), chi.URLParam(r, "slug"), isAdmin(r))
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toProductResponse(p))
	return nil
}

func (h *Handler) createProduct(w http.ResponseWriter, r *http.Request) error {
	var in productRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	p, err := h.svc.CreateProduct(r.Context(), in.toInput())
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, toProductResponse(p))
	return nil
}

func (h *Handler) updateProduct(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.IDParam(r, "id")
	if err != nil {
		return err
	}
	var in productRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	p, err := h.svc.UpdateProduct(r.Context(), id, in.toInput())
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toProductResponse(p))
	return nil
}

// deleteProduct снимает товар с продажи (физически не удаляет).
func (h *Handler) deleteProduct(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.IDParam(r, "id")
	if err != nil {
		return err
	}
	if err := h.svc.DeactivateProduct(r.Context(), id); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// Корзина.

func (h *Handler) getCart(w http.ResponseWriter, r *http.Request) error {
	cart, err := h.svc.GetCart(r.Context(), userID(r))
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toCartResponse(cart))
	return nil
}

// setCartItem: PUT /api/cart/items/{productId} {"quantity": 2}; 0 убирает товар из корзины.
func (h *Handler) setCartItem(w http.ResponseWriter, r *http.Request) error {
	productID, err := httpx.IDParam(r, "productId")
	if err != nil {
		return err
	}
	var in quantityRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	cart, err := h.svc.SetCartItem(r.Context(), userID(r), productID, in.Quantity)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toCartResponse(cart))
	return nil
}

func (h *Handler) removeCartItem(w http.ResponseWriter, r *http.Request) error {
	productID, err := httpx.IDParam(r, "productId")
	if err != nil {
		return err
	}
	cart, err := h.svc.RemoveCartItem(r.Context(), userID(r), productID)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toCartResponse(cart))
	return nil
}

// Заказы.

// checkout: POST /api/orders {"deliveryAddress": "...", "phone": "+7 900 000-00-00"}
// В ответе заказ с paymentUrl — туда фронтенд отправляет покупателя на оплату.
func (h *Handler) checkout(w http.ResponseWriter, r *http.Request) error {
	var in checkoutRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	o, err := h.svc.Checkout(r.Context(), userID(r), shop.CheckoutInput{DeliveryAddress: in.DeliveryAddress, Phone: in.Phone})
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, toOrderResponse(o))
	return nil
}

// pay: POST /api/orders/{id}/pay — ссылка на оплату неоплаченного заказа.
func (h *Handler) pay(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.IDParam(r, "id")
	if err != nil {
		return err
	}
	o, err := h.svc.Pay(r.Context(), id, userID(r))
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toOrderResponse(o))
	return nil
}

// cancel: POST /api/orders/{id}/cancel — отмена своего неоплаченного заказа.
func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.IDParam(r, "id")
	if err != nil {
		return err
	}
	o, err := h.svc.CancelMine(r.Context(), id, userID(r))
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toOrderResponse(o))
	return nil
}

func (h *Handler) myOrders(w http.ResponseWriter, r *http.Request) error {
	orders, err := h.svc.ListMyOrders(r.Context(), userID(r))
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toOrderList(orders))
	return nil
}

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.IDParam(r, "id")
	if err != nil {
		return err
	}
	o, err := h.svc.GetOrder(r.Context(), id, shop.Viewer{UserID: userID(r), IsAdmin: isAdmin(r)})
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toOrderResponse(o))
	return nil
}

// listOrders: GET /api/admin/orders?status=paid&limit=50&offset=0
func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	limit, offset := 50, 0
	if v, err := strconv.Atoi(q.Get("limit")); err == nil && v > 0 && v <= 200 {
		limit = v
	}
	if v, err := strconv.Atoi(q.Get("offset")); err == nil && v >= 0 {
		offset = v
	}
	orders, err := h.svc.ListOrders(r.Context(), q.Get("status"), limit, offset)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toOrderList(orders))
	return nil
}

// setStatus: PUT /api/admin/orders/{id}/status {"status": "shipped"}
func (h *Handler) setStatus(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.IDParam(r, "id")
	if err != nil {
		return err
	}
	var in statusRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	o, err := h.svc.SetStatus(r.Context(), id, in.Status)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toOrderResponse(o))
	return nil
}

// webhook: POST /api/payments/yookassa/webhook — уведомление ЮKassa о платеже.
// Тело читается без DecodeJSON: ЮKassa присылает много полей, а нужен только object.id.
// Отвечаем 200, если уведомление обработано или платёж нам неизвестен (иначе ЮKassa
// будет повторять его сутки), и 500 при временном сбое — тогда ЮKassa повторит позже.
func (h *Handler) webhook(w http.ResponseWriter, r *http.Request) error {
	var n webhookNotification
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&n); err != nil || n.Object.ID == "" {
		return httpx.BadRequest("ожидается уведомление ЮKassa с object.id")
	}
	slog.Info("уведомление ЮKassa", "event", n.Event, "payment", n.Object.ID)
	if err := h.svc.HandlePaymentNotification(r.Context(), n.Object.ID); err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	return nil
}
