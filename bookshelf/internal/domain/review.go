package domain

import (
	"database/sql"
	"time"
)

// Review — модель рецензии для хранения в БД.
type Review struct {
	ID        string         `json:"id" db:"id"`
	BookID    string         `json:"book_id" db:"book_id"`
	UserID    string         `json:"user_id" db:"user_id"`
	Rating    int            `json:"rating" db:"rating"`
	Title     sql.NullString `json:"title" db:"title"`
	Content   string         `json:"content" db:"content"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
}

// ReviewResponse — представление рецензии для API со встроенными данными автора.
type ReviewResponse struct {
	ID        string      `json:"id"`
	BookID    string      `json:"book_id"`
	UserID    string      `json:"user_id"`
	Rating    int         `json:"rating"`
	Title     *string     `json:"title"`
	Content   string      `json:"content"`
	User      UserSummary `json:"user"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// CreateReviewRequest — данные для создания рецензии (POST /books/{bookId}/reviews).
type CreateReviewRequest struct {
	Rating  int     `json:"rating"`
	Title   *string `json:"title"`
	Content string  `json:"content"`
}

// UpdateReviewRequest — данные для обновления рецензии.
// Все поля опциональные: nil означает «не менять».
type UpdateReviewRequest struct {
	Rating  *int    `json:"rating"`
	Title   *string `json:"title"`
	Content *string `json:"content"`
}

// ReviewListResponse — ответ на список рецензий книги.
type ReviewListResponse struct {
	Data       []ReviewResponse `json:"data"`
	Pagination Pagination       `json:"pagination"`
}

// ToResponse разворачивает sql.Null*-поля в указатели и встраивает данные автора.
// Принимает полную модель *User и сам вызывает ToSummary().
// Если user == nil, поле User остаётся нулевым значением.
func (r *Review) ToResponse(user *User) ReviewResponse {
	resp := ReviewResponse{
		ID:        r.ID,
		BookID:    r.BookID,
		UserID:    r.UserID,
		Rating:    r.Rating,
		Title:     nullStringToPtr(r.Title),
		Content:   r.Content,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
	if user != nil {
		resp.User = user.ToSummary()
	}
	return resp
}
