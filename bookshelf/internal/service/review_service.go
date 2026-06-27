package service

import (
	"context"
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/bookshelf/monolith/internal/domain"
	"github.com/bookshelf/monolith/internal/repository"
)

// minReviewContentLength — минимальная длина текста рецензии.
const minReviewContentLength = 10

// Ошибки сервиса рецензий.
var (
	// ErrReviewNotFound — рецензия не найдена.
	ErrReviewNotFound = errors.New("review not found")
	// ErrNotReviewOwner — пользователь не является автором рецензии.
	ErrNotReviewOwner = errors.New("you are not the owner of this review")
	// ErrAlreadyReviewed — пользователь уже оставлял рецензию на эту книгу.
	ErrAlreadyReviewed = errors.New("you have already reviewed this book")
	// ErrInvalidRating — рейтинг вне диапазона 1..5.
	ErrInvalidRating = errors.New("rating must be between 1 and 5")
	// ErrReviewContentTooShort — слишком короткий текст рецензии.
	ErrReviewContentTooShort = errors.New("review content must be at least 10 characters")
)

// ReviewService содержит бизнес-логику работы с рецензиями.
type ReviewService struct {
	reviewRepo *repository.ReviewRepository
	bookRepo   *repository.BookRepository
	userRepo   *repository.UserRepository
}

// NewReviewService создаёт сервис рецензий.
func NewReviewService(reviewRepo *repository.ReviewRepository, bookRepo *repository.BookRepository, userRepo *repository.UserRepository) *ReviewService {
	return &ReviewService{reviewRepo: reviewRepo, bookRepo: bookRepo, userRepo: userRepo}
}

// Create создаёт рецензию пользователя userID на книгу bookID (POST /books/{id}/reviews).
func (s *ReviewService) Create(ctx context.Context, userID, bookID string, req domain.CreateReviewRequest) (*domain.ReviewResponse, error) {
	// 1) Книга должна существовать.
	if _, err := s.bookRepo.GetByID(ctx, bookID); err != nil {
		if errors.Is(err, repository.ErrBookNotFound) {
			return nil, ErrBookNotFound
		}
		return nil, fmt.Errorf("get book: %w", err)
	}

	// 2) Пользователь не должен был ещё оставлять рецензию на эту книгу.
	reviewed, err := s.reviewRepo.UserHasReviewedBook(ctx, userID, bookID)
	if err != nil {
		return nil, fmt.Errorf("check existing review: %w", err)
	}
	if reviewed {
		return nil, ErrAlreadyReviewed
	}

	// 3) Валидация рейтинга и текста.
	if req.Rating < 1 || req.Rating > 5 {
		return nil, ErrInvalidRating
	}
	if utf8.RuneCountInString(req.Content) < minReviewContentLength {
		return nil, ErrReviewContentTooShort
	}

	// 4) Создаём рецензию (repo заполнит ID, CreatedAt, UpdatedAt).
	review := &domain.Review{
		BookID:  bookID,
		UserID:  userID,
		Rating:  req.Rating,
		Title:   ptrToNullString(req.Title),
		Content: req.Content,
	}
	if err := s.reviewRepo.Create(ctx, review); err != nil {
		return nil, fmt.Errorf("create review: %w", err)
	}

	// 5) Подгружаем автора и возвращаем ответ с встроенным UserSummary.
	author, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get review author: %w", err)
	}
	resp := review.ToResponse(author)
	return &resp, nil
}
