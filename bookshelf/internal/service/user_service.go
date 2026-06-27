package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/bookshelf/monolith/internal/domain"
	"github.com/bookshelf/monolith/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// Ошибки сервиса пользователей.
var (
	// ErrInvalidCredentials — неверный email или пароль (логин).
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUserExists — email уже занят.
	ErrUserExists = errors.New("user with this email already exists")
	// ErrUsernameExists — username уже занят.
	ErrUsernameExists = errors.New("username already taken")
	// ErrInvalidPassword — пароль не прошёл валидацию.
	ErrInvalidPassword = errors.New("password must be at least 8 characters")
	// ErrInvalidUsername — username не прошёл валидацию.
	ErrInvalidUsername = errors.New("username must be at least 3 characters")
	// ErrInvalidEmail — email не прошёл валидацию.
	ErrInvalidEmail = errors.New("email is required")
)

// UserService содержит бизнес-логику работы с пользователями.
type UserService struct {
	repo      *repository.UserRepository
	jwtSecret string
}

// NewUserService создаёт сервис пользователей.
func NewUserService(repo *repository.UserRepository, jwtSecret string) *UserService {
	return &UserService{repo: repo, jwtSecret: jwtSecret}
}

// Register регистрирует нового пользователя: валидация -> проверка уникальности ->
// хеширование пароля -> сохранение -> формирование ответа с токеном.
func (s *UserService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.AuthResponse, error) {
	// 1) Валидация входных данных (до похода в БД).
	if len(req.Username) < 3 {
		return nil, ErrInvalidUsername
	}
	if len(req.Password) < 8 {
		return nil, ErrInvalidPassword
	}
	if req.Email == "" {
		return nil, ErrInvalidEmail
	}

	// 2) Проверка уникальности email и username.
	emailTaken, err := s.repo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("check email: %w", err)
	}
	if emailTaken {
		return nil, ErrUserExists
	}
	usernameTaken, err := s.repo.UsernameExists(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("check username: %w", err)
	}
	if usernameTaken {
		return nil, ErrUsernameExists
	}

	// 3) Хеширование пароля (никогда не храним пароль в открытом виде).
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// 4) Создание пользователя (repo заполнит ID, CreatedAt, UpdatedAt).
	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
	}
	if err := s.repo.Create(ctx, user); err != nil {
		// На случай гонки: между проверкой Exists и Create кто-то занял email/username.
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, ErrUserExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	// 5) Генерация токена — будет реализована в следующем этапе (JWT).
	token := ""

	// 6) Ответ: публичные данные (без пароля) + токен.
	return &domain.AuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		User:        user.ToPublic(),
	}, nil
}
