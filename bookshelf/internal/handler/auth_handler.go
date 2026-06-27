package handler

import (
	"errors"
	"net/http"

	"github.com/bookshelf/monolith/internal/domain"
	"github.com/bookshelf/monolith/internal/repository"
	"github.com/bookshelf/monolith/internal/service"
)

// Register — регистрация нового пользователя (POST /auth/register).
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	// Валидация обязательных полей.
	var details []domain.ErrorDetail
	if req.Username == "" {
		details = append(details, domain.ErrorDetail{Field: "username", Message: "username is required"})
	}
	if req.Email == "" {
		details = append(details, domain.ErrorDetail{Field: "email", Message: "email is required"})
	}
	if req.Password == "" {
		details = append(details, domain.ErrorDetail{Field: "password", Message: "password is required"})
	}
	if len(details) > 0 {
		writeValidationError(w, r, details)
		return
	}

	resp, err := h.services.User.Register(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserExists):
			writeError(w, r, http.StatusConflict, "user_exists", err.Error())
		case errors.Is(err, service.ErrUsernameExists):
			writeError(w, r, http.StatusConflict, "username_exists", err.Error())
		case errors.Is(err, service.ErrInvalidUsername),
			errors.Is(err, service.ErrInvalidPassword),
			errors.Is(err, service.ErrInvalidEmail):
			writeError(w, r, http.StatusBadRequest, "validation_error", err.Error())
		default:
			writeError(w, r, http.StatusInternalServerError, "internal_error", "Internal server error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// Login — вход в систему (POST /auth/login).
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	var details []domain.ErrorDetail
	if req.Email == "" {
		details = append(details, domain.ErrorDetail{Field: "email", Message: "email is required"})
	}
	if req.Password == "" {
		details = append(details, domain.ErrorDetail{Field: "password", Message: "password is required"})
	}
	if len(details) > 0 {
		writeValidationError(w, r, details)
		return
	}

	resp, err := h.services.User.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, r, http.StatusUnauthorized, "invalid_credentials", err.Error())
			return
		}
		writeError(w, r, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetCurrentUser — профиль текущего пользователя (GET /users/me).
func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())

	user, err := h.services.User.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "User not found")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, user.ToPublic())
}

// UpdateCurrentUser — обновление профиля (PUT /users/me).
func (h *Handler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())

	var req domain.UpdateUserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	user, err := h.services.User.Update(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUsernameExists):
			writeError(w, r, http.StatusConflict, "username_exists", err.Error())
		case errors.Is(err, service.ErrInvalidUsername):
			writeError(w, r, http.StatusBadRequest, "validation_error", err.Error())
		case errors.Is(err, repository.ErrUserNotFound):
			writeError(w, r, http.StatusNotFound, "not_found", "User not found")
		default:
			writeError(w, r, http.StatusInternalServerError, "internal_error", "Internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, user.ToPublic())
}
