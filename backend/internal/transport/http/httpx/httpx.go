// Package httpx — общие помощники HTTP-обработчиков: ответы в JSON,
// разбор тела запроса и перевод ошибок домена в HTTP-коды.
package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"heritage-backend/internal/domain/errs"
)

// statusByKind — единственное место, где род ошибки домена превращается в HTTP-код.
var statusByKind = map[errs.Kind]int{
	errs.Invalid:      http.StatusBadRequest,
	errs.Unauthorized: http.StatusUnauthorized,
	errs.Forbidden:    http.StatusForbidden,
	errs.NotFound:     http.StatusNotFound,
	errs.Conflict:     http.StatusConflict,
	errs.Unavailable:  http.StatusServiceUnavailable,
}

type errorBody struct {
	Error string `json:"error"`
}

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// WriteError отвечает клиенту по ошибке. Ошибку домена показываем как есть,
// всё остальное — «внутренняя ошибка», а подробности уходят только в лог.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	if e, ok := errs.As(err); ok {
		if status, known := statusByKind[e.Kind]; known {
			JSON(w, status, errorBody{e.Message})
			return
		}
	}
	slog.Error("внутренняя ошибка", "method", r.Method, "path", r.URL.Path, "err", err)
	JSON(w, http.StatusInternalServerError, errorBody{"внутренняя ошибка сервера"})
}

// BadRequest — ошибка формата запроса (не JSON, не число в адресе). Это забота транспорта, не домена.
func BadRequest(msg string) error { return errs.New(errs.Invalid, msg) }

// Unauthorized и Forbidden — ответы middleware на отсутствие входа и прав.
func Unauthorized(msg string) error { return errs.New(errs.Unauthorized, msg) }
func Forbidden(msg string) error    { return errs.New(errs.Forbidden, msg) }

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

// IDParam читает числовой параметр адреса, например {id} в /objects/{id}.
func IDParam(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, name), 10, 64)
	if err != nil || id <= 0 {
		return 0, BadRequest("некорректный " + name)
	}
	return id, nil
}

// Handler — обработчик, который возвращает ошибку вместо того, чтобы писать её сам.
// Так в каждом обработчике остаётся только «полезная» логика.
type Handler func(w http.ResponseWriter, r *http.Request) error

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		WriteError(w, r, err)
	}
}
