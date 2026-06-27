package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bookshelf/monolith/internal/domain"
	"github.com/bookshelf/monolith/internal/service"
	"github.com/go-chi/chi/v5"
)

// ListBooks — список книг с пагинацией/поиском/сортировкой (GET /books, публичный).
func (h *Handler) ListBooks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// page/limit парсим мягко: некорректное значение -> 0, дефолты проставит сервис.
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := domain.BookFilter{
		Search: q.Get("search"),
		Sort:   q.Get("sort"),
		Order:  q.Get("order"),
		Page:   page,
		Limit:  limit,
	}

	resp, err := h.services.Book.List(r.Context(), filter)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetBook — одна книга по ID (GET /books/{bookId}, публичный).
func (h *Handler) GetBook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "bookId")

	resp, err := h.services.Book.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrBookNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Book not found")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// CreateBook — создание книги (POST /books, требует auth).
func (h *Handler) CreateBook(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())

	var req domain.CreateBookRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	// Валидация обязательных полей.
	var details []domain.ErrorDetail
	if req.Title == "" {
		details = append(details, domain.ErrorDetail{Field: "title", Message: "title is required"})
	}
	if req.Author == "" {
		details = append(details, domain.ErrorDetail{Field: "author", Message: "author is required"})
	}
	if len(details) > 0 {
		writeValidationError(w, r, details)
		return
	}

	resp, err := h.services.Book.Create(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookTitleEmpty), errors.Is(err, service.ErrBookAuthorEmpty):
			writeError(w, r, http.StatusBadRequest, "validation_error", err.Error())
		default:
			writeError(w, r, http.StatusInternalServerError, "internal_error", "Internal server error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// UpdateBook — обновление книги владельцем (PUT /books/{bookId}, требует auth).
func (h *Handler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	bookID := chi.URLParam(r, "bookId")

	var req domain.UpdateBookRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	resp, err := h.services.Book.Update(r.Context(), userID, bookID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookNotFound):
			writeError(w, r, http.StatusNotFound, "not_found", "Book not found")
		case errors.Is(err, service.ErrNotBookOwner):
			writeError(w, r, http.StatusForbidden, "forbidden", err.Error())
		case errors.Is(err, service.ErrBookTitleEmpty), errors.Is(err, service.ErrBookAuthorEmpty):
			writeError(w, r, http.StatusBadRequest, "validation_error", err.Error())
		default:
			writeError(w, r, http.StatusInternalServerError, "internal_error", "Internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// DeleteBook — удаление книги владельцем (DELETE /books/{bookId}, требует auth).
func (h *Handler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	bookID := chi.URLParam(r, "bookId")

	if err := h.services.Book.Delete(r.Context(), userID, bookID); err != nil {
		switch {
		case errors.Is(err, service.ErrBookNotFound):
			writeError(w, r, http.StatusNotFound, "not_found", "Book not found")
		case errors.Is(err, service.ErrNotBookOwner):
			writeError(w, r, http.StatusForbidden, "forbidden", err.Error())
		default:
			writeError(w, r, http.StatusInternalServerError, "internal_error", "Internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
