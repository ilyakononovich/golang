package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bookshelf/monolith/internal/domain"
	"github.com/bookshelf/monolith/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// tokenTTL — время жизни access-токена (для MVP: 1 час).
const tokenTTL = time.Hour

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

	// 5) Генерация JWT-токена для нового пользователя.
	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	// 6) Ответ: публичные данные (без пароля) + токен.
	return &domain.AuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(tokenTTL.Seconds()),
		User:        user.ToPublic(),
	}, nil
}

// generateToken создаёт подписанный JWT (HS256) с claims sub/exp/iat.
func (s *UserService) generateToken(userID string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userID,                   // кому принадлежит токен
		"exp": now.Add(tokenTTL).Unix(), // когда истекает
		"iat": now.Unix(),               // когда выдан
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// ValidateToken проверяет подпись и срок действия токена и возвращает ID пользователя.
// Вызывается из AuthMiddleware при каждом запросе к защищённым эндпоинтам.
func (s *UserService) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		// Защита от подмены алгоритма: принимаем только HMAC (HS256).
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		// сюда попадают и истёкший токен, и неверная подпись, и битый формат
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token")
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", errors.New("invalid token subject")
	}
	return sub, nil
}
