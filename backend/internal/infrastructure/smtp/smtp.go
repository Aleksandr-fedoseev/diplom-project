// Package smtp отправляет письма. О домене ничего не знает: адрес, тема и текст — строки.
// Mailer — настоящий SMTP-сервер, LogMailer — запись письма в лог (разработка без почты).
package smtp

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"net"
	netsmtp "net/smtp"
	"strings"
	"time"
)

// ErrHeaderInjection — перевод строки в адресе или теме: так можно было бы
// дописать в письмо чужие заголовки (например, скрытую копию).
var ErrHeaderInjection = errors.New("перевод строки в адресе или теме письма")

// Mailer отправляет письма через SMTP-сервер (порт 587, шифрование STARTTLS).
type Mailer struct {
	addr string // host:port
	auth netsmtp.Auth
	from string
}

// New создаёт клиента SMTP. Пустой user — сервер без авторизации (например, локальный).
func New(host string, port int, user, password, from string) *Mailer {
	var auth netsmtp.Auth
	if user != "" {
		auth = netsmtp.PlainAuth("", user, password, host)
	}
	return &Mailer{addr: net.JoinHostPort(host, fmt.Sprint(port)), auth: auth, from: from}
}

func (m *Mailer) Send(_ context.Context, to, subject, body string) error {
	msg, err := buildMessage(m.from, to, subject, body)
	if err != nil {
		return err
	}
	// smtp.SendMail сам включает STARTTLS, если сервер его поддерживает.
	return netsmtp.SendMail(m.addr, m.auth, m.from, []string{to}, msg)
}

// buildMessage собирает письмо: тема в кодировке для заголовков, текст — UTF-8 в base64,
// чтобы кириллица дошла без искажений через любой почтовый сервер.
func buildMessage(from, to, subject, body string) ([]byte, error) {
	if strings.ContainsAny(from+to+subject, "\r\n") {
		return nil, ErrHeaderInjection
	}
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", subject))
	fmt.Fprintf(&b, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	b.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
	encoded := base64.StdEncoding.EncodeToString([]byte(body))
	for len(encoded) > 76 { // по стандарту строки письма не длиннее 76 символов
		b.WriteString(encoded[:76] + "\r\n")
		encoded = encoded[76:]
	}
	b.WriteString(encoded + "\r\n")
	return []byte(b.String()), nil
}

// LogMailer не отправляет письма, а пишет их в лог. Используется, когда SMTP не настроен (разработка).
type LogMailer struct{}

func (LogMailer) Send(_ context.Context, to, subject, body string) error {
	if strings.ContainsAny(to+subject, "\r\n") {
		return ErrHeaderInjection
	}
	slog.Info("письмо (SMTP не настроен, только в лог)", "to", to, "subject", subject, "body", body)
	return nil
}
