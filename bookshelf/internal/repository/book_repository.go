package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/bookshelf/monolith/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ErrBookNotFound — книга с таким ID не найдена.
var ErrBookNotFound = errors.New("book not found")

// bookColumns — реальные колонки таблицы books (без вычисляемых average_rating/reviews_count).
const bookColumns = "id, title, author, description, isbn, published_year, created_by, created_at, updated_at"

// allowedBookSort — белый список колонок для сортировки (защита от SQL-инъекции:
// значение Sort приходит от клиента и не может подставляться в запрос напрямую).
var allowedBookSort = map[string]string{
	"title":          "title",
	"author":         "author",
	"created_at":     "created_at",
	"published_year": "published_year",
}

// BookRepository отвечает за доступ к данным книг.
type BookRepository struct {
	db *sqlx.DB
}

// NewBookRepository создаёт репозиторий книг.
func NewBookRepository(db *sqlx.DB) *BookRepository {
	return &BookRepository{db: db}
}

// Create сохраняет новую книгу в БД (POST /books).
// ID генерируется в коде через google/uuid; created_at/updated_at проставляет БД.
// Nullable-поля (Description, ISBN, PublishedYear) передаются как sql.Null* —
// при Valid == false в колонку запишется NULL.
func (r *BookRepository) Create(ctx context.Context, book *domain.Book) error {
	book.ID = uuid.NewString()

	const query = `
		INSERT INTO books (id, title, author, description, isbn, published_year, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		book.ID, book.Title, book.Author,
		book.Description, book.ISBN, book.PublishedYear,
		book.CreatedBy,
	).Scan(&book.CreatedAt, &book.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create book: %w", err)
	}
	return nil
}

// GetByID получает книгу по ID (GET /books/{id}).
// Дополнительно вычисляет average_rating и reviews_count через LEFT JOIN с reviews.
// Если рецензий нет: average_rating = NULL, reviews_count = 0.
func (r *BookRepository) GetByID(ctx context.Context, id string) (*domain.Book, error) {
	const query = `
		SELECT
			b.id, b.title, b.author, b.description, b.isbn, b.published_year,
			b.created_by, b.created_at, b.updated_at,
			AVG(r.rating)::float8 AS average_rating,
			COUNT(r.id)::int      AS reviews_count
		FROM books b
		LEFT JOIN reviews r ON r.book_id = b.id
		WHERE b.id = $1
		GROUP BY b.id`

	var book domain.Book
	if err := r.db.GetContext(ctx, &book, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBookNotFound
		}
		return nil, fmt.Errorf("get book by id: %w", err)
	}
	return &book, nil
}

// List возвращает страницу книг с пагинацией, поиском и сортировкой (GET /books).
// Возвращает книги, общее количество (для пагинатора) и ошибку.
func (r *BookRepository) List(ctx context.Context, filter domain.BookFilter) ([]domain.Book, int, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Условие поиска (общее для COUNT и для выборки данных).
	where := ""
	args := []any{}
	if filter.Search != "" {
		where = " WHERE title ILIKE $1 OR author ILIKE $1"
		args = append(args, "%"+filter.Search+"%")
	}

	// 1) Общее количество — отдельным запросом.
	var total int
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM books"+where, args...); err != nil {
		return nil, 0, fmt.Errorf("count books: %w", err)
	}

	// 2) Сама страница данных.
	orderBy := buildBookOrderBy(filter.Sort, filter.Order)
	query := fmt.Sprintf(
		"SELECT %s FROM books%s%s LIMIT $%d OFFSET $%d",
		bookColumns, where, orderBy, len(args)+1, len(args)+2,
	)
	args = append(args, limit, offset)

	var books []domain.Book
	if err := r.db.SelectContext(ctx, &books, query, args...); err != nil {
		return nil, 0, fmt.Errorf("list books: %w", err)
	}
	return books, total, nil
}

// Update обновляет данные книги (PUT /books/{id}).
// Проверка прав владельца — в сервисном слое. updated_at ставит триггер БД.
func (r *BookRepository) Update(ctx context.Context, book *domain.Book) error {
	const query = `
		UPDATE books
		SET title = $1, author = $2, description = $3, isbn = $4, published_year = $5
		WHERE id = $6
		RETURNING updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		book.Title, book.Author, book.Description, book.ISBN, book.PublishedYear, book.ID,
	).Scan(&book.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrBookNotFound
		}
		return fmt.Errorf("update book: %w", err)
	}
	return nil
}

// Delete удаляет книгу (DELETE /books/{id}). Связанные рецензии удаляются каскадно.
func (r *BookRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM books WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete book: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete book rows affected: %w", err)
	}
	if n == 0 {
		return ErrBookNotFound
	}
	return nil
}

// buildBookOrderBy строит безопасный ORDER BY из белого списка колонок.
// Неизвестная колонка -> сортировка по created_at; направление -> только ASC/DESC.
func buildBookOrderBy(sort, order string) string {
	col, ok := allowedBookSort[sort]
	if !ok {
		col = "created_at"
	}
	dir := "DESC"
	if strings.EqualFold(order, "asc") {
		dir = "ASC"
	}
	return " ORDER BY " + col + " " + dir
}
