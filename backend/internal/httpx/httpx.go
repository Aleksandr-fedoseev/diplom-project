// Package httpx — общие помощники для HTTP-обработчиков: ответы в JSON и ошибки.
package httpx

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// Error — ошибка, которую можно показать клиенту. Всё остальное
// клиент видит как «внутренняя ошибка», а подробности уходят в лог.
type Error struct {
	Status  int    `json:"-"`
	Message string `json:"error"`
}

func (e *Error) Error() string { return e.Message }

func BadRequest(msg string) *Error   { return &Error{http.StatusBadRequest, msg} }
func Unauthorized(msg string) *Error { return &Error{http.StatusUnauthorized, msg} }
func Forbidden(msg string) *Error    { return &Error{http.StatusForbidden, msg} }
func NotFound(msg string) *Error     { return &Error{http.StatusNotFound, msg} }
func Conflict(msg string) *Error     { return &Error{http.StatusConflict, msg} }

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var he *Error
	if errors.As(err, &he) {
		JSON(w, he.Status, he)
		return
	}
	slog.Error("внутренняя ошибка", "method", r.Method, "path", r.URL.Path, "err", err)
	JSON(w, http.StatusInternalServerError, &Error{Message: "внутренняя ошибка сервера"})
}

// DecodeJSON читает тело запроса в v и отклоняет неизвестные поля,
// чтобы опечатка в названии поля не проходила молча.
func DecodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return BadRequest("некорректный JSON: " + err.Error())
	}
	return nil
}

// Handler — обработчик, который возвращает ошибку вместо того, чтобы писать её сам.
// Так в каждом обработчике остаётся только «полезная» логика.
type Handler func(w http.ResponseWriter, r *http.Request) error

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		WriteError(w, r, err)
	}
}
