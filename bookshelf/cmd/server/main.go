package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	// mux — маршрутизатор: сопоставляет путь запроса с обработчиком.
	mux := http.NewServeMux()

	// Health-check: GET /health -> {"status": "ok"}.
	// Нужен frontend-команде и мониторингу для проверки, что сервис жив.
	mux.HandleFunc("GET /health", healthHandler)

	addr := ":8080"
	log.Printf("HTTP-сервер запускается на %s", addr)

	// ListenAndServe блокирует выполнение и слушает входящие запросы.
	// Возвращает ошибку, только если сервер не смог запуститься или упал.
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("сервер остановлен с ошибкой: %v", err)
	}
}

// healthHandler отвечает JSON-ом {"status": "ok"}.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Заголовок выставляем до записи тела ответа.
	w.Header().Set("Content-Type", "application/json")

	// Кодируем map напрямую в тело ответа.
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		log.Printf("не удалось закодировать ответ /health: %v", err)
	}
}
