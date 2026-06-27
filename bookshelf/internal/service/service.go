package service

import (
	"github.com/bookshelf/monolith/internal/repository"
)

// Service агрегирует все сервисы приложения (бизнес-логика).
type Service struct {
	User   *UserService
	Book   *BookService
	Review *ReviewService
}

// New создаёт Service и инициализирует все вложенные сервисы.
func New(repos *repository.Repository, jwtSecret string) *Service {
	return &Service{
		User:   NewUserService(repos.User, jwtSecret),
		Book:   NewBookService(repos.Book),
		Review: NewReviewService(repos.Review),
	}
}
