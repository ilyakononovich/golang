package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	// http.HandleFunc регистрирует обработчик в стандартном роутере
	// DefaultServeMux — он встроен в пакет net/http и используется по умолчанию.
	// Health-check нужен frontend-команде и мониторингу для проверки, что сервис жив.
	http.HandleFunc("/health", healthHandler)

	addr := ":8080"
	log.Printf("HTTP-сервер запускается на %s", addr)

	// Второй аргумент nil означает "использовать DefaultServeMux",
	// в котором мы только что зарегистрировали /health.
	// ListenAndServe блокирует выполнение и слушает входящие запросы.
	if err := http.ListenAndServe(addr, nil); err != nil {
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
