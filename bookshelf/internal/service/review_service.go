package service

import (
	"github.com/bookshelf/monolith/internal/repository"
)

// ReviewService содержит бизнес-логику работы с рецензиями.
type ReviewService struct {
	repo *repository.ReviewRepository
}

// NewReviewService создаёт сервис рецензий.
func NewReviewService(repo *repository.ReviewRepository) *ReviewService {
	return &ReviewService{repo: repo}
}
