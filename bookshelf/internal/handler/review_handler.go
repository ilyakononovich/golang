package handler

import (
	"errors"
	"net/http"
	"strconv"
	"unicode/utf8"

	"github.com/bookshelf/monolith/internal/domain"
	"github.com/bookshelf/monolith/internal/service"
	"github.com/go-chi/chi/v5"
)

// ListBookReviews — рецензии книги с пагинацией (GET /books/{bookId}/reviews, публичный).
func (h *Handler) ListBookReviews(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "bookId")

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	resp, err := h.services.Review.ListByBookID(r.Context(), bookID, page, limit)
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

// GetReview — одна рецензия по ID (GET /reviews/{reviewId}, публичный).
func (h *Handler) GetReview(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "reviewId")

	resp, err := h.services.Review.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrReviewNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Review not found")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// CreateReview — создание рецензии (POST /books/{bookId}/reviews, требует auth).
func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	bookID := chi.URLParam(r, "bookId")

	var req domain.CreateReviewRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	// Валидация рейтинга и текста.
	var details []domain.ErrorDetail
	if req.Rating < 1 || req.Rating > 5 {
		details = append(details, domain.ErrorDetail{Field: "rating", Message: "rating must be between 1 and 5"})
	}
	if utf8.RuneCountInString(req.Content) < 10 {
		details = append(details, domain.ErrorDetail{Field: "content", Message: "content must be at least 10 characters"})
	}
	if len(details) > 0 {
		writeValidationError(w, r, details)
		return
	}

	resp, err := h.services.Review.Create(r.Context(), userID, bookID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookNotFound):
			writeError(w, r, http.StatusNotFound, "not_found", "Book not found")
		case errors.Is(err, service.ErrAlreadyReviewed):
			writeError(w, r, http.StatusConflict, "already_reviewed", err.Error())
		case errors.Is(err, service.ErrInvalidRating), errors.Is(err, service.ErrReviewContentTooShort):
			writeError(w, r, http.StatusBadRequest, "validation_error", err.Error())
		default:
			writeError(w, r, http.StatusInternalServerError, "internal_error", "Internal server error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// UpdateReview — обновление рецензии автором (PUT /reviews/{reviewId}, требует auth).
func (h *Handler) UpdateReview(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	reviewID := chi.URLParam(r, "reviewId")

	var req domain.UpdateReviewRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	resp, err := h.services.Review.Update(r.Context(), userID, reviewID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrReviewNotFound):
			writeError(w, r, http.StatusNotFound, "not_found", "Review not found")
		case errors.Is(err, service.ErrNotReviewOwner):
			writeError(w, r, http.StatusForbidden, "forbidden", err.Error())
		case errors.Is(err, service.ErrInvalidRating), errors.Is(err, service.ErrReviewContentTooShort):
			writeError(w, r, http.StatusBadRequest, "validation_error", err.Error())
		default:
			writeError(w, r, http.StatusInternalServerError, "internal_error", "Internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// DeleteReview — удаление рецензии автором (DELETE /reviews/{reviewId}, требует auth).
func (h *Handler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	reviewID := chi.URLParam(r, "reviewId")

	if err := h.services.Review.Delete(r.Context(), userID, reviewID); err != nil {
		switch {
		case errors.Is(err, service.ErrReviewNotFound):
			writeError(w, r, http.StatusNotFound, "not_found", "Review not found")
		case errors.Is(err, service.ErrNotReviewOwner):
			writeError(w, r, http.StatusForbidden, "forbidden", err.Error())
		default:
			writeError(w, r, http.StatusInternalServerError, "internal_error", "Internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
