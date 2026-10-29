// Package shop — домен магазина сувениров: товары, корзина, заказы и их статусы.
// Только стандартная библиотека: ни БД, ни HTTP, ни ЮKassa здесь нет.
//
// Деньги везде — целые копейки (int64), как и в БД: 4500 ₽ — это 450000.
// Дробные числа для денег не используются: у float64 бывают ошибки округления.
package shop

import (
	"fmt"
	"strings"
	"time"
)

// FormatRubles переводит копейки в сумму с двумя знаками: 450000 → "4500.00".
// В таком виде сумму принимает и возвращает ЮKassa. Считаем целыми числами.
func FormatRubles(kopecks int64) string {
	sign := ""
	if kopecks < 0 {
		sign, kopecks = "-", -kopecks
	}
	return fmt.Sprintf("%s%d.%02d", sign, kopecks/100, kopecks%100)
}

// Товары.

type Product struct {
	ID           int64
	Slug         string
	Title        string
	Description  string
	Category     string
	PriceKopecks int64
	Stock        int
	ImageURL     string
	IsActive     bool // снятый с продажи товар остаётся в БД: на него ссылаются старые заказы
	UpdatedAt    time.Time
}

// InStock — товар можно купить: он в продаже и есть на складе.
func (p Product) InStock() bool { return p.IsActive && p.Stock > 0 }

// ProductInput — данные для создания и изменения товара администратором.
type ProductInput struct {
	Slug         string
	Title        string
	Description  string
	Category     string
	PriceKopecks int64
	Stock        int
	ImageURL     string
	IsActive     *bool // nil — «в продаже»
}

// Normalize приводит поля к одному виду и проверяет правила товара.
func (in *ProductInput) Normalize() error {
	in.Slug = strings.TrimSpace(strings.ToLower(in.Slug))
	in.Title = strings.TrimSpace(in.Title)
	in.Category = strings.TrimSpace(in.Category)
	switch {
	case in.Slug == "" || strings.ContainsAny(in.Slug, " /?#"):
		return ErrInvalidSlug
	case in.Title == "":
		return ErrTitleRequired
	case in.PriceKopecks <= 0:
		return ErrInvalidPrice
	case in.Stock < 0:
		return ErrNegativeStock
	}
	if in.IsActive == nil {
		active := true
		in.IsActive = &active
	}
	return nil
}

// Product собирает товар из проверенных данных (сначала вызовите Normalize).
func (in ProductInput) Product(id int64) Product {
	active := in.IsActive == nil || *in.IsActive
	return Product{
		ID: id, Slug: in.Slug, Title: in.Title, Description: in.Description, Category: in.Category,
		PriceKopecks: in.PriceKopecks, Stock: in.Stock, ImageURL: in.ImageURL, IsActive: active,
	}
}

// Корзина.

// CartItem — позиция корзины вместе с текущими данными товара.
type CartItem struct {
	ProductID    int64
	Slug         string
	Title        string
	ImageURL     string
	PriceKopecks int64 // текущая цена товара
	Quantity     int
	Stock        int
	IsActive     bool
}

func (i CartItem) LineTotal() int64 { return i.PriceKopecks * int64(i.Quantity) }

// Available — товар в продаже и его хватает; иначе оформить заказ не получится.
func (i CartItem) Available() bool { return i.IsActive && i.Quantity <= i.Stock }

type Cart struct {
	Items []CartItem
}

// Total — сумма корзины по текущим ценам.
func (c Cart) Total() int64 {
	var total int64
	for _, it := range c.Items {
		total += it.LineTotal()
	}
	return total
}

// CheckCartQuantity — правило «положить в корзину»: больше, чем есть на складе, нельзя.
// Окончательно наличие проверяется при оформлении: пока товар лежит в корзине, его могут купить другие.
func CheckCartQuantity(p Product, quantity int) error {
	if quantity < 0 {
		return ErrInvalidQuantity
	}
	if !p.IsActive {
		return ErrProductNotFound
	}
	if quantity > p.Stock {
		return errNotEnoughForCart(p.Title, p.Stock)
	}
	return nil
}

// Заказы.

// Status — статус заказа.
type Status string

const (
	StatusPending   Status = "pending"   // ждёт оплаты
	StatusPaid      Status = "paid"      // оплачен
	StatusShipped   Status = "shipped"   // отправлен
	StatusCompleted Status = "completed" // выполнен
	StatusCancelled Status = "cancelled" // отменён
)

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusPaid, StatusShipped, StatusCompleted, StatusCancelled:
		return true
	}
	return false
}

// ParseStatus проверяет статус из запроса.
func ParseStatus(s string) (Status, error) {
	if st := Status(s); st.Valid() {
		return st, nil
	}
	return "", ErrInvalidStatus
}

// transitions — какие смены статуса может сделать администратор.
// Оплата (pending → paid) сюда не входит: её подтверждает только ЮKassa.
// Отмена оплаченного заказа не возвращает деньги автоматически: возврат
// делается вручную в личном кабинете ЮKassa (автоматические возвраты — вне рамок проекта).
var transitions = map[Status][]Status{
	StatusPending: {StatusCancelled},
	StatusPaid:    {StatusShipped, StatusCancelled},
	StatusShipped: {StatusCompleted},
}

// CheckTransition — правило смены статуса администратором.
func CheckTransition(from, to Status) error {
	for _, next := range transitions[from] {
		if next == to {
			return nil
		}
	}
	return errTransitionNotAllowed(from, to)
}

// OrderItem — позиция заказа. Название и цена — снимок на момент покупки:
// смена цены товара оформленные заказы не меняет.
type OrderItem struct {
	ProductID    int64
	Title        string
	Quantity     int
	PriceKopecks int64
}

func (i OrderItem) LineTotal() int64 { return i.PriceKopecks * int64(i.Quantity) }

type Order struct {
	ID              int64
	UserID          *int64 // nil, если покупатель удалён
	Status          Status
	TotalKopecks    int64
	PaymentID       string // id платежа в ЮKassa; пусто, если платёж не создан
	PaymentURL      string // страница оплаты ЮKassa
	DeliveryAddress string
	Phone           string
	Items           []OrderItem
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// OwnedBy — заказ принадлежит пользователю userID.
func (o Order) OwnedBy(userID int64) bool { return o.UserID != nil && *o.UserID == userID }

// Viewer — кто смотрит заказ. Домен магазина не знает про роли модуля auth,
// поэтому transport передаёт только «кто» и «администратор ли».
type Viewer struct {
	UserID  int64
	IsAdmin bool
}

// VisibleTo — чужой заказ видит только администратор.
func (o Order) VisibleTo(v Viewer) bool { return v.IsAdmin || o.OwnedBy(v.UserID) }

// CheckoutInput — данные покупателя при оформлении заказа.
type CheckoutInput struct {
	DeliveryAddress string
	Phone           string
}

// Normalize обрезает пробелы и проверяет адрес и телефон.
func (in *CheckoutInput) Normalize() error {
	in.DeliveryAddress = strings.TrimSpace(in.DeliveryAddress)
	in.Phone = strings.TrimSpace(in.Phone)
	if in.DeliveryAddress == "" || len([]rune(in.DeliveryAddress)) > 500 {
		return ErrInvalidAddress
	}
	// Телефон: цифры, пробелы, скобки, дефисы и «+» в начале; цифр — от 10 до 15.
	digits := 0
	for i, r := range in.Phone {
		switch {
		case r >= '0' && r <= '9':
			digits++
		case r == ' ' || r == '(' || r == ')' || r == '-' || (r == '+' && i == 0):
		default:
			return ErrInvalidPhone
		}
	}
	if digits < 10 || digits > 15 {
		return ErrInvalidPhone
	}
	return nil
}

// CheckoutLine — позиция корзины с заблокированной строкой товара (см. CartRepository.LockForCheckout).
type CheckoutLine struct {
	ProductID    int64
	Title        string
	PriceKopecks int64
	Quantity     int
	Stock        int
	IsActive     bool
}

// NewOrder — правило оформления: корзина не пуста, все товары в продаже и их хватает.
// Сумма считается на сервере по текущим ценам, в позиции попадает снимок названия и цены.
func NewOrder(userID int64, in CheckoutInput, lines []CheckoutLine) (Order, error) {
	if len(lines) == 0 {
		return Order{}, ErrCartEmpty
	}
	o := Order{
		UserID:          &userID,
		Status:          StatusPending,
		DeliveryAddress: in.DeliveryAddress,
		Phone:           in.Phone,
		Items:           make([]OrderItem, 0, len(lines)),
	}
	for _, l := range lines {
		if !l.IsActive {
			return Order{}, errProductInactive(l.Title)
		}
		if l.Quantity > l.Stock {
			return Order{}, errNotEnoughForOrder(l.Title, l.Stock)
		}
		item := OrderItem{ProductID: l.ProductID, Title: l.Title, Quantity: l.Quantity, PriceKopecks: l.PriceKopecks}
		o.Items = append(o.Items, item)
		o.TotalKopecks += item.LineTotal()
	}
	return o, nil
}
