package repository

import (
	"github.com/jmoiron/sqlx"
)

// BookRepository отвечает за доступ к данным книг.
type BookRepository struct {
	db *sqlx.DB
}

// NewBookRepository создаёт репозиторий книг.
func NewBookRepository(db *sqlx.DB) *BookRepository {
	return &BookRepository{db: db}
}
