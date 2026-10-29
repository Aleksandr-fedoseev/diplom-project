package shop

import (
	"errors"
	"testing"

	"heritage-backend/internal/domain/errs"
)

func TestFormatRubles(t *testing.T) {
	cases := map[int64]string{450000: "4500.00", 5: "0.05", 100: "1.00", 199: "1.99", 12345678: "123456.78", 0: "0.00", -150: "-1.50"}
	for kopecks, want := range cases {
		if got := FormatRubles(kopecks); got != want {
			t.Errorf("FormatRubles(%d) = %q, ожидалось %q", kopecks, got, want)
		}
	}
}

func TestCheckTransition(t *testing.T) {
	allowed := map[[2]Status]bool{
		{StatusPending, StatusCancelled}: true,
		{StatusPaid, StatusShipped}:      true,
		{StatusPaid, StatusCancelled}:    true,
		{StatusShipped, StatusCompleted}: true,
	}
	all := []Status{StatusPending, StatusPaid, StatusShipped, StatusCompleted, StatusCancelled}
	for _, from := range all {
		for _, to := range all {
			err := CheckTransition(from, to)
			if allowed[[2]Status{from, to}] != (err == nil) {
				t.Errorf("%s → %s: %v", from, to, err)
			}
			if err != nil && !errs.Is(err, errs.Conflict) {
				t.Errorf("запрещённый переход должен быть конфликтом: %v", err)
			}
		}
	}
}

func TestParseStatus(t *testing.T) {
	if st, err := ParseStatus("shipped"); err != nil || st != StatusShipped {
		t.Fatalf("shipped: %v %v", st, err)
	}
	if _, err := ParseStatus("lost"); !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("lost: %v", err)
	}
}

func TestCheckoutInputNormalize(t *testing.T) {
	in := CheckoutInput{DeliveryAddress: "  г. Астрахань, ул. Никольская, 1 ", Phone: " +7 (900) 123-45-67 "}
	if err := in.Normalize(); err != nil || in.DeliveryAddress != "г. Астрахань, ул. Никольская, 1" || in.Phone != "+7 (900) 123-45-67" {
		t.Fatalf("нормализация: %+v, %v", in, err)
	}
	long := make([]rune, 501)
	for i := range long {
		long[i] = 'а'
	}
	bad := []CheckoutInput{
		{DeliveryAddress: "  ", Phone: "+79001234567"},
		{DeliveryAddress: string(long), Phone: "+79001234567"},
		{DeliveryAddress: "адрес", Phone: "12-34"},            // мало цифр
		{DeliveryAddress: "адрес", Phone: "1234567890123456"}, // 16 цифр
		{DeliveryAddress: "адрес", Phone: "+7 900 abc 45 67"}, // буквы
		{DeliveryAddress: "адрес", Phone: "8 900 123+45 67"},  // «+» не в начале
	}
	for _, b := range bad {
		if err := b.Normalize(); !errs.Is(err, errs.Invalid) {
			t.Errorf("ожидалась ошибка для %+v: %v", b, err)
		}
	}
	ok := CheckoutInput{DeliveryAddress: "адрес", Phone: "123456789012345"} // ровно 15 цифр
	if err := ok.Normalize(); err != nil {
		t.Fatal(err)
	}
}

func TestProductInputNormalize(t *testing.T) {
	in := ProductInput{Slug: " Merch-Cap ", Title: " Кепка ", PriceKopecks: 100}
	if err := in.Normalize(); err != nil || in.Slug != "merch-cap" || in.Title != "Кепка" || in.IsActive == nil || !*in.IsActive {
		t.Fatalf("нормализация: %+v, %v", in, err)
	}
	cases := map[error]ProductInput{
		ErrInvalidSlug:   {Slug: "a b", Title: "x", PriceKopecks: 1},
		ErrTitleRequired: {Slug: "a", Title: " ", PriceKopecks: 1},
		ErrInvalidPrice:  {Slug: "a", Title: "x", PriceKopecks: 0},
		ErrNegativeStock: {Slug: "a", Title: "x", PriceKopecks: 1, Stock: -1},
	}
	for want, in := range cases {
		if err := in.Normalize(); !errors.Is(err, want) {
			t.Errorf("%+v: %v, ожидалось %v", in, err, want)
		}
	}
	inactive := false
	p := ProductInput{Slug: "a", Title: "x", PriceKopecks: 1, IsActive: &inactive}.Product(7)
	if p.ID != 7 || p.IsActive || p.InStock() {
		t.Fatalf("товар: %+v", p)
	}
}

func TestCartRules(t *testing.T) {
	cart := Cart{Items: []CartItem{
		{PriceKopecks: 450000, Quantity: 2, Stock: 5, IsActive: true},
		{PriceKopecks: 120000, Quantity: 3, Stock: 2, IsActive: true},
	}}
	if cart.Total() != 2*450000+3*120000 || cart.Items[0].LineTotal() != 900000 {
		t.Fatalf("сумма корзины: %d", cart.Total())
	}
	if !cart.Items[0].Available() || cart.Items[1].Available() {
		t.Fatal("доступность позиций")
	}

	p := Product{Title: "Кепка", Stock: 5, IsActive: true}
	if err := CheckCartQuantity(p, 5); err != nil {
		t.Fatal(err)
	}
	if err := CheckCartQuantity(p, 6); !errs.Is(err, errs.Conflict) {
		t.Fatalf("больше остатка: %v", err)
	}
	if err := CheckCartQuantity(p, -1); !errors.Is(err, ErrInvalidQuantity) {
		t.Fatalf("отрицательное: %v", err)
	}
	p.IsActive = false
	if err := CheckCartQuantity(p, 1); !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("снятый товар: %v", err)
	}
}

func TestNewOrder(t *testing.T) {
	in := CheckoutInput{DeliveryAddress: "адрес", Phone: "+79001234567"}
	if _, err := NewOrder(1, in, nil); !errors.Is(err, ErrCartEmpty) {
		t.Fatalf("пустая корзина: %v", err)
	}
	lines := []CheckoutLine{
		{ProductID: 1, Title: "Худи", PriceKopecks: 450000, Quantity: 2, Stock: 5, IsActive: true},
		{ProductID: 2, Title: "Блокнот", PriceKopecks: 90000, Quantity: 3, Stock: 3, IsActive: true},
	}
	o, err := NewOrder(7, in, lines)
	if err != nil || o.TotalKopecks != 2*450000+3*90000 || len(o.Items) != 2 || o.Status != StatusPending ||
		!o.OwnedBy(7) || o.Items[0].Title != "Худи" || o.Items[0].PriceKopecks != 450000 {
		t.Fatalf("заказ: %+v, %v", o, err)
	}

	lines[1].Stock = 2
	if _, err := NewOrder(7, in, lines); !errs.Is(err, errs.Conflict) {
		t.Fatalf("не хватает товара: %v", err)
	}
	lines[1].Stock, lines[1].IsActive = 3, false
	if _, err := NewOrder(7, in, lines); !errs.Is(err, errs.Conflict) {
		t.Fatalf("снятый товар: %v", err)
	}
}

func TestOrderVisibleTo(t *testing.T) {
	owner := int64(5)
	o := Order{UserID: &owner}
	if !o.VisibleTo(Viewer{UserID: 5}) || o.VisibleTo(Viewer{UserID: 6}) || !o.VisibleTo(Viewer{UserID: 6, IsAdmin: true}) {
		t.Fatal("видимость заказа")
	}
	if (Order{}).OwnedBy(0) {
		t.Fatal("заказ удалённого покупателя ничей")
	}
}
