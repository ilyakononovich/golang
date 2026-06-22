package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/bookshelf/monolith/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.Load()
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Get("/health", healthHandler)

	addr := ":" + cfg.Port
	log.Printf("HTTP-сервер запускается на %s", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("сервер остановлен с ошибкой: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		log.Printf("не удалось закодировать ответ /health: %v", err)
	}
}
