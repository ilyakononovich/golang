package repository

import (
	"github.com/jmoiron/sqlx"
)

// Repository агрегирует все репозитории приложения.
// Через него сервисный слой получает доступ к данным пользователей, книг и рецензий.
type Repository struct {
	User   *UserRepository
	Book   *BookRepository
	Review *ReviewRepository
}

// New создаёт Repository и инициализирует все вложенные репозитории
// одним общим подключением к БД.
func New(db *sqlx.DB) *Repository {
	return &Repository{
		User:   NewUserRepository(db),
		Book:   NewBookRepository(db),
		Review: NewReviewRepository(db),
	}
}
