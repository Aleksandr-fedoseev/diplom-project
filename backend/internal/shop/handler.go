package shop

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"heritage-backend/internal/auth"
	"heritage-backend/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Routes подключается по адресу /api.
func (h *Handler) Routes(r chi.Router) {
	// Каталог товаров — для всех.
	r.Method(http.MethodGet, "/products", httpx.Handler(h.listProducts))
	r.Method(http.MethodGet, "/products/{slug}", httpx.Handler(h.getProduct))

	// Уведомления ЮKassa — без входа: ЮKassa токена не присылает.
	r.Method(http.MethodPost, "/payments/yookassa/webhook", httpx.Handler(h.webhook))

	r.Group(func(r chi.Router) {
		r.Use(auth.RequireRole(auth.RoleUser))
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
		r.Use(auth.RequireRole(auth.RoleAdmin))
		r.Method(http.MethodPost, "/products", httpx.Handler(h.createProduct))
		r.Method(http.MethodPut, "/products/{id}", httpx.Handler(h.updateProduct))
		r.Method(http.MethodDelete, "/products/{id}", httpx.Handler(h.deleteProduct))

		r.Method(http.MethodGet, "/admin/orders", httpx.Handler(h.listOrders))
		r.Method(http.MethodPut, "/admin/orders/{id}/status", httpx.Handler(h.setStatus))
	})
}

func isAdmin(r *http.Request) bool {
	p, ok := auth.FromContext(r.Context())
	return ok && p.Role == auth.RoleAdmin
}

func idParam(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, name), 10, 64)
	if err != nil {
		return 0, httpx.BadRequest("некорректный " + name)
	}
	return id, nil
}

// listProducts: GET /api/products?category=clothes
// Администратор с ?all=true видит и снятые с продажи товары.
func (h *Handler) listProducts(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	products, err := h.svc.ListProducts(r.Context(), q.Get("category"), isAdmin(r) && q.Get("all") == "true")
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, products)
	return nil
}

func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request) error {
	p, err := h.svc.GetProduct(r.Context(), chi.URLParam(r, "slug"), isAdmin(r))
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, p)
	return nil
}

func (h *Handler) createProduct(w http.ResponseWriter, r *http.Request) error {
	var in ProductInput
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	p, err := h.svc.CreateProduct(r.Context(), in)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, p)
	return nil
}

func (h *Handler) updateProduct(w http.ResponseWriter, r *http.Request) error {
	id, err := idParam(r, "id")
	if err != nil {
		return err
	}
	var in ProductInput
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	p, err := h.svc.UpdateProduct(r.Context(), id, in)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, p)
	return nil
}

// deleteProduct снимает товар с продажи (см. Service.DeactivateProduct).
func (h *Handler) deleteProduct(w http.ResponseWriter, r *http.Request) error {
	id, err := idParam(r, "id")
	if err != nil {
		return err
	}
	if err := h.svc.DeactivateProduct(r.Context(), id); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) getCart(w http.ResponseWriter, r *http.Request) error {
	p, _ := auth.FromContext(r.Context())
	cart, err := h.svc.GetCart(r.Context(), p.UserID)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, cart)
	return nil
}

// setCartItem: PUT /api/cart/items/{productId} {"quantity": 2}; 0 убирает товар из корзины.
func (h *Handler) setCartItem(w http.ResponseWriter, r *http.Request) error {
	productID, err := idParam(r, "productId")
	if err != nil {
		return err
	}
	var in struct {
		Quantity int32 `json:"quantity"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	p, _ := auth.FromContext(r.Context())
	cart, err := h.svc.SetCartItem(r.Context(), p.UserID, productID, in.Quantity)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, cart)
	return nil
}

func (h *Handler) removeCartItem(w http.ResponseWriter, r *http.Request) error {
	productID, err := idParam(r, "productId")
	if err != nil {
		return err
	}
	p, _ := auth.FromContext(r.Context())
	cart, err := h.svc.RemoveCartItem(r.Context(), p.UserID, productID)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, cart)
	return nil
}

// cancel: POST /api/orders/{id}/cancel — отмена своего неоплаченного заказа.
func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) error {
	id, err := idParam(r, "id")
	if err != nil {
		return err
	}
	p, _ := auth.FromContext(r.Context())
	o, err := h.svc.CancelMine(r.Context(), id, p.UserID)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, o)
	return nil
}

// checkout: POST /api/orders {"deliveryAddress": "...", "phone": "+7 900 000-00-00"}
// В ответе заказ с paymentUrl — туда фронтенд отправляет покупателя на оплату.
func (h *Handler) checkout(w http.ResponseWriter, r *http.Request) error {
	var in CheckoutInput
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	p, _ := auth.FromContext(r.Context())
	o, err := h.svc.Checkout(r.Context(), p.UserID, in)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, o)
	return nil
}

// pay: POST /api/orders/{id}/pay — ссылка на оплату неоплаченного заказа.
func (h *Handler) pay(w http.ResponseWriter, r *http.Request) error {
	id, err := idParam(r, "id")
	if err != nil {
		return err
	}
	p, _ := auth.FromContext(r.Context())
	o, err := h.svc.Pay(r.Context(), id, p.UserID)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, o)
	return nil
}

func (h *Handler) myOrders(w http.ResponseWriter, r *http.Request) error {
	p, _ := auth.FromContext(r.Context())
	orders, err := h.svc.ListMyOrders(r.Context(), p.UserID)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, orders)
	return nil
}

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) error {
	id, err := idParam(r, "id")
	if err != nil {
		return err
	}
	p, _ := auth.FromContext(r.Context())
	o, err := h.svc.GetOrder(r.Context(), id, p)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, o)
	return nil
}

// listOrders: GET /api/admin/orders?status=paid&limit=50&offset=0
func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	limit, offset := int32(50), int32(0)
	if v, err := strconv.ParseInt(q.Get("limit"), 10, 32); err == nil && v > 0 && v <= 200 {
		limit = int32(v)
	}
	if v, err := strconv.ParseInt(q.Get("offset"), 10, 32); err == nil && v >= 0 {
		offset = int32(v)
	}
	orders, err := h.svc.ListOrders(r.Context(), q.Get("status"), limit, offset)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, orders)
	return nil
}

// setStatus: PUT /api/admin/orders/{id}/status {"status": "shipped"}
func (h *Handler) setStatus(w http.ResponseWriter, r *http.Request) error {
	id, err := idParam(r, "id")
	if err != nil {
		return err
	}
	var in struct {
		Status string `json:"status"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	o, err := h.svc.SetStatus(r.Context(), id, in.Status)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, o)
	return nil
}

// webhook: POST /api/payments/yookassa/webhook — уведомление ЮKassa о платеже.
// Тело читается без DecodeJSON: ЮKassa присылает много полей, а нужен только object.id.
// Отвечаем 200, если уведомление обработано или платёж нам неизвестен (иначе ЮKassa
// будет повторять его сутки), и 500 при временном сбое — тогда ЮKassa повторит позже.
func (h *Handler) webhook(w http.ResponseWriter, r *http.Request) error {
	var n struct {
		Event  string `json:"event"`
		Object struct {
			ID string `json:"id"`
		} `json:"object"`
	}
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
