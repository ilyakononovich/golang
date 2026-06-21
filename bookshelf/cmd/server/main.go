package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/bookshelf/monolith/internal/config"
)

func main() {
	cfg := config.Load()

	http.HandleFunc("/health", healthHandler)

	addr := ":" + cfg.Port
	log.Printf("HTTP-сервер запускается на %s", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("сервер остановлен с ошибкой: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		log.Printf("не удалось закодировать ответ /health: %v", err)
	}
}
