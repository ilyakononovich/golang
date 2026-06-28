package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bookshelf/monolith/internal/config"
	"github.com/bookshelf/monolith/internal/handler"
	"github.com/bookshelf/monolith/internal/repository"
	"github.com/bookshelf/monolith/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // регистрирует драйвер PostgreSQL через init()
)

func main() {
	// 1) Конфигурация.
	cfg := config.Load()

	// 2) Подключение к БД + настройки пула соединений.
	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("не удалось подключиться к базе данных: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	log.Println("Connected to database")

	// 3) Dependency Injection: DB -> Repositories -> Services -> Handlers.
	repos := repository.New(db)
	services := service.New(repos, cfg.JWTSecret)
	handlers := handler.New(services, cfg.JWTSecret)

	// 4) Роутер.
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health-эндпоинты (вне /api/v1).
	r.Get("/health", handlers.Health)
	r.Get("/ready", handlers.Ready)

	r.Route("/api/v1", func(r chi.Router) {
		// --- Публичные роуты ---
		r.Post("/auth/register", handlers.Register)
		r.Post("/auth/login", handlers.Login)

		r.Get("/books", handlers.ListBooks)
		r.Get("/books/{bookId}", handlers.GetBook)
		r.Get("/books/{bookId}/reviews", handlers.ListBookReviews)
		r.Get("/reviews/{reviewId}", handlers.GetReview)

		// --- Защищённые роуты (AuthMiddleware) ---
		r.Group(func(r chi.Router) {
			r.Use(handlers.AuthMiddleware)

			r.Get("/users/me", handlers.GetCurrentUser)
			r.Put("/users/me", handlers.UpdateCurrentUser)

			r.Post("/books", handlers.CreateBook)
			r.Put("/books/{bookId}", handlers.UpdateBook)
			r.Delete("/books/{bookId}", handlers.DeleteBook)

			r.Post("/books/{bookId}/reviews", handlers.CreateReview)
			r.Put("/reviews/{reviewId}", handlers.UpdateReview)
			r.Delete("/reviews/{reviewId}", handlers.DeleteReview)
		})
	})

	// 5) HTTP-сервер с таймаутами.
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запускаем сервер в горутине, чтобы main мог ждать сигнал завершения.
	go func() {
		log.Printf("HTTP-сервер запускается на %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("сервер остановлен с ошибкой: %v", err)
		}
	}()

	// Graceful shutdown: ждём SIGINT/SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Получен сигнал завершения, останавливаем сервер...")

	// Даём активным запросам до 10 секунд завершиться.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown не удался: %v", err)
	}
	log.Println("Сервер остановлен корректно")
}
