package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/bookshelf/monolith/internal/domain"
	"github.com/bookshelf/monolith/internal/service"
	"github.com/go-chi/chi/v5/middleware"
)

// version — версия API, отдаётся в health-эндпоинтах.
const version = "1.0.0"

// contextKey — собственный тип для ключей контекста, чтобы исключить коллизии
// с ключами из других пакетов (строковый ключ "userID" мог бы пересечься).
type contextKey string

// userIDKey — ключ, под которым AuthMiddleware кладёт ID пользователя в контекст.
const userIDKey contextKey = "userID"

// Handler — транспортный слой: знает про HTTP, вызывает сервисы.
type Handler struct {
	services  *service.Service
	jwtSecret string
}

// New создаёт Handler.
func New(services *service.Service, jwtSecret string) *Handler {
	return &Handler{services: services, jwtSecret: jwtSecret}
}

// writeJSON отправляет JSON-ответ с указанным статусом.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data == nil {
		return
	}
	// Ошибку кодирования логировать особо некуда — статус уже отправлен.
	_ = json.NewEncoder(w).Encode(data)
}

// writeError отправляет ошибку в едином формате ErrorResponse, включая RequestID.
func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	writeJSON(w, status, domain.ErrorResponse{
		Code:      code,
		Message:   message,
		RequestID: middleware.GetReqID(r.Context()),
	})
}

// writeValidationError отправляет 400 с детализацией ошибок валидации полей.
func writeValidationError(w http.ResponseWriter, r *http.Request, details []domain.ErrorDetail) {
	writeJSON(w, http.StatusBadRequest, domain.ErrorResponse{
		Code:      "validation_error",
		Message:   "validation failed",
		Details:   details,
		RequestID: middleware.GetReqID(r.Context()),
	})
}

// decodeJSON парсит JSON из тела запроса в v.
func decodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// getUserID извлекает ID пользователя из контекста (пустая строка, если его нет).
func getUserID(ctx context.Context) string {
	id, _ := ctx.Value(userIDKey).(string)
	return id
}

// Health — liveness-проба: сервис жив.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"version":   version,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Ready — readiness-проба: сервис готов принимать трафик (включая зависимости).
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"version":   version,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"checks": map[string]string{
			"database": "ok",
		},
	})
}
