package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
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

// AuthMiddleware проверяет JWT из заголовка Authorization и кладёт userID в контекст.
// При любой проблеме с токеном отвечает 401 и прерывает цепочку (next не вызывается).
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1) Заголовок Authorization обязателен.
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, r, http.StatusUnauthorized, "unauthorized", "Authorization header required")
			return
		}

		// 2-4) Разбираем формат "Bearer <token>".
		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			writeError(w, r, http.StatusUnauthorized, "unauthorized", "Invalid authorization header format")
			return
		}
		token := parts[1]

		// 5-6) Валидируем токен и извлекаем userID.
		userID, err := h.services.User.ValidateToken(token)
		if err != nil {
			writeError(w, r, http.StatusUnauthorized, "unauthorized", "Invalid or expired token")
			return
		}

		// 7-8) Кладём userID в контекст и передаём управление дальше.
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
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
