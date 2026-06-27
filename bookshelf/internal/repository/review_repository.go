package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bookshelf/monolith/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ErrReviewNotFound — рецензия с таким ID не найдена.
var ErrReviewNotFound = errors.New("review not found")

// reviewColumns — колонки таблицы reviews.
const reviewColumns = "id, book_id, user_id, rating, title, content, created_at, updated_at"

// ReviewRepository отвечает за доступ к данным рецензий.
type ReviewRepository struct {
	db *sqlx.DB
}

// NewReviewRepository создаёт репозиторий рецензий.
func NewReviewRepository(db *sqlx.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

// Create сохраняет новую рецензию (POST /books/{id}/reviews).
// ID генерируется через google/uuid; created_at/updated_at проставляет БД.
func (r *ReviewRepository) Create(ctx context.Context, review *domain.Review) error {
	review.ID = uuid.NewString()

	const query = `
		INSERT INTO reviews (id, book_id, user_id, rating, title, content)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		review.ID, review.BookID, review.UserID, review.Rating, review.Title, review.Content,
	).Scan(&review.CreatedAt, &review.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create review: %w", err)
	}
	return nil
}

// GetByID получает рецензию по ID (для проверок перед update/delete).
func (r *ReviewRepository) GetByID(ctx context.Context, id string) (*domain.Review, error) {
	const query = "SELECT " + reviewColumns + " FROM reviews WHERE id = $1"

	var review domain.Review
	if err := r.db.GetContext(ctx, &review, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReviewNotFound
		}
		return nil, fmt.Errorf("get review by id: %w", err)
	}
	return &review, nil
}

// ListByBookID возвращает страницу рецензий книги с пагинацией (GET /books/{id}/reviews).
// Возвращает рецензии, общее количество и ошибку.
func (r *ReviewRepository) ListByBookID(ctx context.Context, bookID string, page, limit int) ([]domain.Review, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Общее количество рецензий книги.
	var total int
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM reviews WHERE book_id = $1", bookID); err != nil {
		return nil, 0, fmt.Errorf("count reviews: %w", err)
	}

	// Страница данных, новые сверху.
	const query = "SELECT " + reviewColumns + `
		FROM reviews
		WHERE book_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	var reviews []domain.Review
	if err := r.db.SelectContext(ctx, &reviews, query, bookID, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("list reviews: %w", err)
	}
	return reviews, total, nil
}

// Update обновляет рецензию (PUT /reviews/{id}).
// Проверка авторства — в сервисе. updated_at ставит триггер БД.
func (r *ReviewRepository) Update(ctx context.Context, review *domain.Review) error {
	const query = `
		UPDATE reviews
		SET rating = $1, title = $2, content = $3
		WHERE id = $4
		RETURNING updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		review.Rating, review.Title, review.Content, review.ID,
	).Scan(&review.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrReviewNotFound
		}
		return fmt.Errorf("update review: %w", err)
	}
	return nil
}

// Delete удаляет рецензию (DELETE /reviews/{id}).
func (r *ReviewRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM reviews WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete review: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete review rows affected: %w", err)
	}
	if n == 0 {
		return ErrReviewNotFound
	}
	return nil
}

// UserHasReviewedBook проверяет, оставлял ли пользователь рецензию на книгу
// (ограничение «одна рецензия на книгу от пользователя»).
func (r *ReviewRepository) UserHasReviewedBook(ctx context.Context, userID, bookID string) (bool, error) {
	const query = "SELECT EXISTS(SELECT 1 FROM reviews WHERE user_id = $1 AND book_id = $2)"

	var exists bool
	if err := r.db.GetContext(ctx, &exists, query, userID, bookID); err != nil {
		return false, fmt.Errorf("check user reviewed book: %w", err)
	}
	return exists, nil
}
