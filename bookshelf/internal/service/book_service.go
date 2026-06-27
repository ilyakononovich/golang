package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bookshelf/monolith/internal/domain"
	"github.com/bookshelf/monolith/internal/repository"
)

// Ошибки сервиса книг.
var (
	ErrBookNotFound    = errors.New("book not found")
	ErrNotBookOwner    = errors.New("you are not the owner of this book")
	ErrBookTitleEmpty  = errors.New("title is required")
	ErrBookAuthorEmpty = errors.New("author is required")
)

// BookService содержит бизнес-логику работы с книгами.
type BookService struct {
	bookRepo *repository.BookRepository
	userRepo *repository.UserRepository
}

// NewBookService создаёт сервис книг.
func NewBookService(bookRepo *repository.BookRepository, userRepo *repository.UserRepository) *BookService {
	return &BookService{bookRepo: bookRepo, userRepo: userRepo}
}

// Create создаёт новую книгу от имени пользователя userID (POST /books).
func (s *BookService) Create(ctx context.Context, userID string, req domain.CreateBookRequest) (*domain.BookResponse, error) {
	// 1) Валидация: обязательные поля.
	if req.Title == "" {
		return nil, ErrBookTitleEmpty
	}
	if req.Author == "" {
		return nil, ErrBookAuthorEmpty
	}

	// 2) Формируем доменную модель; опциональные поля -> sql.Null*.
	book := &domain.Book{
		Title:         req.Title,
		Author:        req.Author,
		Description:   ptrToNullString(req.Description),
		ISBN:          ptrToNullString(req.ISBN),
		PublishedYear: ptrToNullInt32(req.PublishedYear),
		CreatedBy:     userID,
	}

	// 3) Сохраняем (repo заполнит ID, CreatedAt, UpdatedAt).
	if err := s.bookRepo.Create(ctx, book); err != nil {
		return nil, fmt.Errorf("create book: %w", err)
	}

	// 4) Возвращаем представление для API.
	resp := book.ToResponse()
	return &resp, nil
}

// GetByID возвращает книгу по ID вместе с данными её создателя (GET /books/{id}).
func (s *BookService) GetByID(ctx context.Context, id string) (*domain.BookResponse, error) {
	// 1) Получаем книгу (с вычисленными average_rating и reviews_count).
	book, err := s.bookRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrBookNotFound) {
			return nil, ErrBookNotFound
		}
		return nil, fmt.Errorf("get book: %w", err)
	}

	resp := book.ToResponse()

	// 2) Подгружаем создателя и встраиваем его краткое представление.
	creator, err := s.userRepo.GetByID(ctx, book.CreatedBy)
	if err != nil {
		// Создателя нет (например, удалён) — не критично, оставляем Creator пустым.
		if !errors.Is(err, repository.ErrUserNotFound) {
			return nil, fmt.Errorf("get book creator: %w", err)
		}
	} else {
		summary := creator.ToSummary()
		resp.Creator = &summary
	}

	return &resp, nil
}

// Значения по умолчанию для пагинации/сортировки списка книг.
const (
	defaultPage  = 1
	defaultLimit = 20
	maxLimit     = 100
	defaultSort  = "created_at"
	defaultOrder = "desc"
)

// List возвращает страницу книг с пагинацией, поиском и сортировкой (GET /books).
// Creator намеренно не заполняется — клиент получает только created_by (UUID),
// а поле Creator не попадает в JSON благодаря omitempty.
func (s *BookService) List(ctx context.Context, filter domain.BookFilter) (*domain.BookListResponse, error) {
	// 1) Значения по умолчанию и защита от некорректных значений.
	if filter.Page < 1 {
		filter.Page = defaultPage
	}
	if filter.Limit < 1 {
		filter.Limit = defaultLimit
	}
	if filter.Limit > maxLimit {
		filter.Limit = maxLimit
	}
	if filter.Sort == "" {
		filter.Sort = defaultSort
	}
	if filter.Order == "" {
		filter.Order = defaultOrder
	}

	// 2) Запрос данных + общего количества.
	books, total, err := s.bookRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list books: %w", err)
	}

	// 3) Конвертация в представление для API (без Creator).
	data := make([]domain.BookResponse, len(books))
	for i := range books {
		data[i] = books[i].ToResponse()
	}

	return &domain.BookListResponse{
		Data:       data,
		Pagination: domain.NewPagination(filter.Page, filter.Limit, total),
	}, nil
}

// Update обновляет книгу. Менять может только её владелец (PUT /books/{id}).
func (s *BookService) Update(ctx context.Context, userID, bookID string, req domain.UpdateBookRequest) (*domain.BookResponse, error) {
	// 1) Получаем книгу.
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		if errors.Is(err, repository.ErrBookNotFound) {
			return nil, ErrBookNotFound
		}
		return nil, fmt.Errorf("get book: %w", err)
	}

	// 2) Проверка владельца.
	if book.CreatedBy != userID {
		return nil, ErrNotBookOwner
	}

	// 3) Применяем только переданные поля (nil = не менять).
	if req.Title != nil {
		if *req.Title == "" {
			return nil, ErrBookTitleEmpty
		}
		book.Title = *req.Title
	}
	if req.Author != nil {
		if *req.Author == "" {
			return nil, ErrBookAuthorEmpty
		}
		book.Author = *req.Author
	}
	if req.Description != nil {
		book.Description = ptrToNullString(req.Description)
	}
	if req.ISBN != nil {
		book.ISBN = ptrToNullString(req.ISBN)
	}
	if req.PublishedYear != nil {
		book.PublishedYear = ptrToNullInt32(req.PublishedYear)
	}

	// 4) Сохраняем.
	if err := s.bookRepo.Update(ctx, book); err != nil {
		if errors.Is(err, repository.ErrBookNotFound) {
			return nil, ErrBookNotFound
		}
		return nil, fmt.Errorf("update book: %w", err)
	}

	resp := book.ToResponse()
	return &resp, nil
}

// Delete удаляет книгу. Удалить может только её владелец (DELETE /books/{id}).
func (s *BookService) Delete(ctx context.Context, userID, bookID string) error {
	// 1) Получаем книгу.
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		if errors.Is(err, repository.ErrBookNotFound) {
			return ErrBookNotFound
		}
		return fmt.Errorf("get book: %w", err)
	}

	// 2) Проверка владельца.
	if book.CreatedBy != userID {
		return ErrNotBookOwner
	}

	// 3) Удаляем.
	if err := s.bookRepo.Delete(ctx, bookID); err != nil {
		if errors.Is(err, repository.ErrBookNotFound) {
			return ErrBookNotFound
		}
		return fmt.Errorf("delete book: %w", err)
	}
	return nil
}

// ptrToNullString превращает *string в sql.NullString (nil -> NULL).
func ptrToNullString(p *string) sql.NullString {
	if p == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *p, Valid: true}
}

// ptrToNullInt32 превращает *int в sql.NullInt32 (nil -> NULL).
func ptrToNullInt32(p *int) sql.NullInt32 {
	if p == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: int32(*p), Valid: true}
}
