package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bookshelf/monolith/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Ошибки репозитория пользователей.
var (
	// ErrUserNotFound — пользователь с таким идентификатором/email/username не найден.
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists — email или username уже заняты.
	ErrUserAlreadyExists = errors.New("user already exists")
)

// userColumns — список колонок таблицы users в фиксированном порядке.
const userColumns = "id, username, email, password_hash, created_at, updated_at"

// UserRepository отвечает за доступ к данным пользователей.
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository создаёт репозиторий пользователей.
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create сохраняет нового пользователя в БД (POST /auth/register).
// ID генерируется в коде через google/uuid; created_at/updated_at проставляет БД.
// Если email или username уже заняты, возвращает ErrUserAlreadyExists.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	user.ID = uuid.NewString()

	const query = `
		INSERT INTO users (id, username, email, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// GetByID получает пользователя по ID (GET /users/me, формирование UserSummary).
func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	const query = "SELECT " + userColumns + " FROM users WHERE id = $1"
	return r.getOne(ctx, query, id)
}

// GetByEmail ищет пользователя по email (логин: проверка существования + хеш пароля).
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = "SELECT " + userColumns + " FROM users WHERE email = $1"
	return r.getOne(ctx, query, email)
}

// GetByUsername ищет пользователя по имени (регистрация: проверка уникальности).
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	const query = "SELECT " + userColumns + " FROM users WHERE username = $1"
	return r.getOne(ctx, query, username)
}

// getOne — общий помощник: выполняет запрос на одну запись и маппит её в *domain.User.
// При отсутствии строки возвращает ErrUserNotFound.
func (r *UserRepository) getOne(ctx context.Context, query string, arg any) (*domain.User, error) {
	var user domain.User
	if err := r.db.GetContext(ctx, &user, query, arg); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &user, nil
}

// Update обновляет данные пользователя (PUT /users/me).
// updated_at проставляется триггером БД; забираем его через RETURNING.
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	const query = `
		UPDATE users
		SET username = $1, email = $2, password_hash = $3
		WHERE id = $4
		RETURNING updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		user.Username, user.Email, user.PasswordHash, user.ID,
	).Scan(&user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		if isUniqueViolation(err) {
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

// EmailExists проверяет, занят ли email (быстрая проверка без загрузки записи).
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	return r.exists(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email)
}

// UsernameExists проверяет, занят ли username.
func (r *UserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	return r.exists(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", username)
}

// exists — общий помощник для проверок ...Exists.
func (r *UserRepository) exists(ctx context.Context, query, arg string) (bool, error) {
	var exists bool
	if err := r.db.GetContext(ctx, &exists, query, arg); err != nil {
		return false, fmt.Errorf("check exists: %w", err)
	}
	return exists, nil
}

// GetByIDs пакетно загружает пользователей по списку ID одним запросом.
// Возвращает map id -> *User для быстрого поиска (решает N+1 при списке рецензий).
func (r *UserRepository) GetByIDs(ctx context.Context, ids []string) (map[string]*domain.User, error) {
	result := make(map[string]*domain.User, len(ids))
	if len(ids) == 0 {
		return result, nil
	}

	// sqlx.In разворачивает срез в (?, ?, ...), затем Rebind меняет ? на $1,$2 для postgres.
	query, args, err := sqlx.In("SELECT "+userColumns+" FROM users WHERE id IN (?)", ids)
	if err != nil {
		return nil, fmt.Errorf("build IN query: %w", err)
	}
	query = r.db.Rebind(query)

	var users []domain.User
	if err := r.db.SelectContext(ctx, &users, query, args...); err != nil {
		return nil, fmt.Errorf("get users by ids: %w", err)
	}

	for i := range users {
		u := users[i]
		result[u.ID] = &u
	}
	return result, nil
}

// isUniqueViolation сообщает, является ли ошибка нарушением UNIQUE-ограничения
// PostgreSQL (код 23505).
func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
