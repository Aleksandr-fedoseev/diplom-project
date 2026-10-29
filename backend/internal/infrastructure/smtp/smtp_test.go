package smtp

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
)

func TestBuildMessageEncodesCyrillic(t *testing.T) {
	raw, err := buildMessage("noreply@heritage.ru", "a@b.ru", "Решение по заявке", "Текст")
	if err != nil {
		t.Fatal(err)
	}
	msg := string(raw)
	if !strings.Contains(msg, "Subject: =?utf-8?q?") || !strings.Contains(msg, "0KLQtdC60YHRgg==") {
		t.Fatalf("письмо закодировано неверно:\n%s", msg)
	}
	if !strings.Contains(msg, "To: a@b.ru\r\n") || !strings.Contains(msg, "Content-Transfer-Encoding: base64\r\n\r\n") {
		t.Fatalf("заголовки письма неверны:\n%s", msg)
	}
}

func TestBuildMessageWrapsLongBody(t *testing.T) {
	body := strings.Repeat("Длинный текст письма. ", 20)
	raw, err := buildMessage("noreply@heritage.ru", "a@b.ru", "Тема", body)
	if err != nil {
		t.Fatal(err)
	}
	_, encoded, _ := strings.Cut(string(raw), "\r\n\r\n")
	lines := strings.Split(strings.TrimSuffix(encoded, "\r\n"), "\r\n")
	for _, l := range lines {
		if len(l) > 76 {
			t.Fatalf("строка длиннее 76 символов: %d", len(l))
		}
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.Join(lines, ""))
	if err != nil || string(decoded) != body {
		t.Fatalf("текст не восстанавливается: err=%v", err)
	}
}

func TestHeaderInjectionRejected(t *testing.T) {
	if _, err := buildMessage("noreply@heritage.ru", "a@b.ru", "плохая\r\nBcc: x@y.z", ""); !errors.Is(err, ErrHeaderInjection) {
		t.Fatalf("тема с переводом строки: %v", err)
	}
	if _, err := buildMessage("noreply@heritage.ru", "a@b.ru\nBcc: x@y.z", "Тема", ""); !errors.Is(err, ErrHeaderInjection) {
		t.Fatalf("адрес с переводом строки: %v", err)
	}
	if err := (LogMailer{}).Send(context.Background(), "a@b.ru", "x\ny", ""); !errors.Is(err, ErrHeaderInjection) {
		t.Fatalf("LogMailer: %v", err)
	}
	if err := (LogMailer{}).Send(context.Background(), "a@b.ru", "Тема", "Текст\nс переводами строк"); err != nil {
		t.Fatalf("LogMailer: %v", err)
	}
}
