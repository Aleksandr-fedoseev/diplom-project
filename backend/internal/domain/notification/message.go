package notification

import (
	"fmt"
	"strings"
)

// Message — письмо одному получателю.
type Message struct {
	To      string
	Subject string
	Body    string
}

// Validate проверяет письмо перед отправкой: адрес обязателен, а перевод строки
// в адресе или теме позволил бы дописать в письмо чужие заголовки.
func (m Message) Validate() error {
	if strings.TrimSpace(m.To) == "" {
		return ErrNoRecipientEmail
	}
	if strings.ContainsAny(m.To+m.Subject, "\r\n") {
		return ErrHeaderInjection
	}
	return nil
}

// FormatRubles переводит копейки в рубли для письма: 450000 → "4500.00".
func FormatRubles(kopecks int64) string {
	sign := ""
	if kopecks < 0 {
		sign, kopecks = "-", -kopecks
	}
	return fmt.Sprintf("%s%d.%02d", sign, kopecks/100, kopecks%100)
}
