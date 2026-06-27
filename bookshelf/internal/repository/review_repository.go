package repository

import (
	"github.com/jmoiron/sqlx"
)

// ReviewRepository отвечает за доступ к данным рецензий.
type ReviewRepository struct {
	db *sqlx.DB
}

// NewReviewRepository создаёт репозиторий рецензий.
func NewReviewRepository(db *sqlx.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}
