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

// Login аутентифицирует пользователя по email и паролю и возвращает токен.
// В обоих случаях (нет такого email / неверный пароль) возвращает ErrInvalidCredentials,
// чтобы не раскрывать, зарегистрирован ли email.
func (s *UserService) Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error) {
	// 1) Ищем пользователя по email.
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	// 2) Сравниваем пароль с хешем. Несовпадение -> те же ErrInvalidCredentials.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// 3) Генерируем токен.
	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	// 4) Ответ: публичные данные + токен.
	return &domain.AuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(tokenTTL.Seconds()),
		User:        user.ToPublic(),
	}, nil
}

// GetByID возвращает пользователя по ID. Делегирует в репозиторий.
func (s *UserService) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	return s.repo.GetByID(ctx, userID)
}

// Update обновляет профиль пользователя (PUT /users/me).
// В MVP меняется только username; пустой username трактуется как «не менять».
func (s *UserService) Update(ctx context.Context, userID string, req domain.UpdateUserRequest) (*domain.User, error) {
	// 1) Получаем текущего пользователя.
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 2) Если username передан — валидируем длину и уникальность.
	if req.Username != "" {
		if len(req.Username) < 3 {
			return nil, ErrInvalidUsername
		}

		// Уникальность через GetByUsername (а не UsernameExists), чтобы отличить
		// «имя занято другим» от «это его же имя».
		existing, err := s.repo.GetByUsername(ctx, req.Username)
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			// имя свободно — продолжаем
		case err != nil:
			return nil, fmt.Errorf("check username: %w", err)
		case existing.ID != userID:
			return nil, ErrUsernameExists // имя занято другим пользователем
		}
		// если existing.ID == userID — имя его собственное, ничего не проверяем

		user.Username = req.Username
	}

	// 3) Сохраняем изменения.
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return user, nil
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
