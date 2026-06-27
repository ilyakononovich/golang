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

// GetByID возвращает рецензию по ID вместе с данными автора.
func (s *ReviewService) GetByID(ctx context.Context, id string) (*domain.ReviewResponse, error) {
	review, err := s.reviewRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrReviewNotFound) {
			return nil, ErrReviewNotFound
		}
		return nil, fmt.Errorf("get review: %w", err)
	}

	// Автор (если его нет — ToResponse оставит User пустым).
	author, err := s.userRepo.GetByID(ctx, review.UserID)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, fmt.Errorf("get review author: %w", err)
	}

	resp := review.ToResponse(author)
	return &resp, nil
}

// ListByBookID возвращает страницу рецензий книги с авторами и пагинацией
// (GET /books/{id}/reviews).
func (s *ReviewService) ListByBookID(ctx context.Context, bookID string, page, limit int) (*domain.ReviewListResponse, error) {
	// 1) Книга должна существовать.
	if _, err := s.bookRepo.GetByID(ctx, bookID); err != nil {
		if errors.Is(err, repository.ErrBookNotFound) {
			return nil, ErrBookNotFound
		}
		return nil, fmt.Errorf("get book: %w", err)
	}

	// 2) Значения по умолчанию для пагинации.
	if page < 1 {
		page = defaultPage
	}
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	// 3) Сами рецензии + общее количество.
	reviews, total, err := s.reviewRepo.ListByBookID(ctx, bookID, page, limit)
	if err != nil {
		return nil, fmt.Errorf("list reviews: %w", err)
	}

	// 4) Авторы одним запросом (без N+1): собираем ID -> GetByIDs -> map.
	userIDs := make([]string, 0, len(reviews))
	for i := range reviews {
		userIDs = append(userIDs, reviews[i].UserID)
	}
	usersByID, err := s.userRepo.GetByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("get review authors: %w", err)
	}

	// 5) Собираем ответы, подставляя автора из map.
	data := make([]domain.ReviewResponse, len(reviews))
	for i := range reviews {
		author := usersByID[reviews[i].UserID] // nil, если автор не найден
		data[i] = reviews[i].ToResponse(author)
	}

	return &domain.ReviewListResponse{
		Data:       data,
		Pagination: domain.NewPagination(page, limit, total),
	}, nil
}

// Update обновляет рецензию. Менять может только её автор (PUT /reviews/{id}).
func (s *ReviewService) Update(ctx context.Context, userID, reviewID string, req domain.UpdateReviewRequest) (*domain.ReviewResponse, error) {
	// 1) Получаем рецензию.
	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		if errors.Is(err, repository.ErrReviewNotFound) {
			return nil, ErrReviewNotFound
		}
		return nil, fmt.Errorf("get review: %w", err)
	}

	// 2) Проверка авторства.
	if review.UserID != userID {
		return nil, ErrNotReviewOwner
	}

	// 3) Применяем и валидируем только переданные поля (nil = не менять).
	if req.Rating != nil {
		if *req.Rating < 1 || *req.Rating > 5 {
			return nil, ErrInvalidRating
		}
		review.Rating = *req.Rating
	}
	if req.Content != nil {
		if utf8.RuneCountInString(*req.Content) < minReviewContentLength {
			return nil, ErrReviewContentTooShort
		}
		review.Content = *req.Content
	}
	if req.Title != nil {
		review.Title = ptrToNullString(req.Title)
	}

	// 4) Сохраняем.
	if err := s.reviewRepo.Update(ctx, review); err != nil {
		if errors.Is(err, repository.ErrReviewNotFound) {
			return nil, ErrReviewNotFound
		}
		return nil, fmt.Errorf("update review: %w", err)
	}

	// 5) Подгружаем автора и возвращаем ответ.
	author, err := s.userRepo.GetByID(ctx, review.UserID)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, fmt.Errorf("get review author: %w", err)
	}
	resp := review.ToResponse(author)
	return &resp, nil
}

// Delete удаляет рецензию. Удалить может только её автор (DELETE /reviews/{id}).
func (s *ReviewService) Delete(ctx context.Context, userID, reviewID string) error {
	// 1) Получаем рецензию.
	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		if errors.Is(err, repository.ErrReviewNotFound) {
			return ErrReviewNotFound
		}
		return fmt.Errorf("get review: %w", err)
	}

	// 2) Проверка авторства.
	if review.UserID != userID {
		return ErrNotReviewOwner
	}

	// 3) Удаляем.
	if err := s.reviewRepo.Delete(ctx, reviewID); err != nil {
		if errors.Is(err, repository.ErrReviewNotFound) {
			return ErrReviewNotFound
		}
		return fmt.Errorf("delete review: %w", err)
	}
	return nil
}
