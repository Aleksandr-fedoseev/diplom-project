package notifications

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// SMTPSender отправляет письма через SMTP-сервер (порт 587, шифрование STARTTLS).
type SMTPSender struct {
	addr string // host:port
	auth smtp.Auth
	from string
}

func NewSMTPSender(host string, port int, user, password, from string) *SMTPSender {
	var auth smtp.Auth
	if user != "" {
		auth = smtp.PlainAuth("", user, password, host)
	}
	return &SMTPSender{addr: net.JoinHostPort(host, fmt.Sprint(port)), auth: auth, from: from}
}

func (s *SMTPSender) Send(_ context.Context, msg Message) error {
	// smtp.SendMail сам включает STARTTLS, если сервер его поддерживает.
	return smtp.SendMail(s.addr, s.auth, s.from, []string{msg.To}, buildMessage(s.from, msg))
}

// buildMessage собирает письмо: тема в кодировке для заголовков, текст — UTF-8 в base64,
// чтобы кириллица дошла без искажений через любой почтовый сервер.
func buildMessage(from string, msg Message) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", msg.To)
	fmt.Fprintf(&b, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", msg.Subject))
	fmt.Fprintf(&b, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	b.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
	body := base64.StdEncoding.EncodeToString([]byte(msg.Body))
	for len(body) > 76 { // по стандарту строки письма не длиннее 76 символов
		b.WriteString(body[:76] + "\r\n")
		body = body[76:]
	}
	b.WriteString(body + "\r\n")
	return []byte(b.String())
}

// LogSender не отправляет письма, а пишет их в лог. Используется, когда SMTP не настроен (разработка).
type LogSender struct{}

func (LogSender) Send(_ context.Context, msg Message) error {
	slog.Info("письмо (SMTP не настроен, только в лог)", "to", msg.To, "subject", msg.Subject, "body", msg.Body)
	return nil
}
