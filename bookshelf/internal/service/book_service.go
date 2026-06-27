package service

import (
	"github.com/bookshelf/monolith/internal/repository"
)

// BookService содержит бизнес-логику работы с книгами.
type BookService struct {
	repo *repository.BookRepository
}

// NewBookService создаёт сервис книг.
func NewBookService(repo *repository.BookRepository) *BookService {
	return &BookService{repo: repo}
}
