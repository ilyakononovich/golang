package domain

import (
	"database/sql"
	"time"
)

// Book — модель книги для хранения в БД.
// Nullable-колонки используют sql.Null*-типы, чтобы корректно читать NULL.
type Book struct {
	ID            string          `json:"id" db:"id"`
	Title         string          `json:"title" db:"title"`
	Author        string          `json:"author" db:"author"`
	Description   sql.NullString  `json:"description" db:"description"`
	ISBN          sql.NullString  `json:"isbn" db:"isbn"`
	PublishedYear sql.NullInt32   `json:"published_year" db:"published_year"`
	AverageRating sql.NullFloat64 `json:"average_rating" db:"average_rating"`
	// ReviewsCount — вычисляемое поле: колонки в таблице books нет,
	// значение приходит из SELECT COUNT(...) AS reviews_count (JOIN с reviews).
	ReviewsCount int       `json:"reviews_count" db:"reviews_count"`
	CreatedBy    string    `json:"created_by" db:"created_by"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// BookResponse — представление книги для API.
// Nullable-поля — указатели (nil = отсутствует), чтобы JSON был чистым.
type BookResponse struct {
	ID            string       `json:"id"`
	Title         string       `json:"title"`
	Author        string       `json:"author"`
	Description   *string      `json:"description"`
	ISBN          *string      `json:"isbn"`
	PublishedYear *int         `json:"published_year"`
	AverageRating *float64     `json:"average_rating"`
	ReviewsCount  int          `json:"reviews_count"`
	CreatedBy     string       `json:"created_by"`
	Creator       *UserSummary `json:"creator,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

// CreateBookRequest — данные для создания книги (POST /books).
type CreateBookRequest struct {
	Title         string  `json:"title"`
	Author        string  `json:"author"`
	Description   *string `json:"description"`
	ISBN          *string `json:"isbn"`
	PublishedYear *int    `json:"published_year"`
}

// UpdateBookRequest — данные для обновления книги (PUT /books/{id}).
// Все поля опциональные: nil означает «не менять».
type UpdateBookRequest struct {
	Title         *string `json:"title"`
	Author        *string `json:"author"`
	Description   *string `json:"description"`
	ISBN          *string `json:"isbn"`
	PublishedYear *int    `json:"published_year"`
}

// BookFilter — параметры выборки списка книг (GET /books).
type BookFilter struct {
	Search string `json:"search"`
	Sort   string `json:"sort"`
	Order  string `json:"order"`
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
}

// BookListResponse — ответ на список книг.
type BookListResponse struct {
	Data       []BookResponse `json:"data"`
	Pagination Pagination     `json:"pagination"`
}

// ToResponse разворачивает sql.Null*-поля в указатели.
// Данные создателя (Creator) заполняются отдельно в сервисе.
func (b *Book) ToResponse() BookResponse {
	return BookResponse{
		ID:            b.ID,
		Title:         b.Title,
		Author:        b.Author,
		Description:   nullStringToPtr(b.Description),
		ISBN:          nullStringToPtr(b.ISBN),
		PublishedYear: nullInt32ToPtr(b.PublishedYear),
		AverageRating: nullFloat64ToPtr(b.AverageRating),
		ReviewsCount:  b.ReviewsCount,
		CreatedBy:     b.CreatedBy,
		CreatedAt:     b.CreatedAt,
		UpdatedAt:     b.UpdatedAt,
	}
}
