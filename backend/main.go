package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// healthCheckHandler - это наш обработчик запросов.
// Он будет вызываться каждый раз, когда кто-то заходит на нужный нам URL.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Разрешаем нашему фронтенду делать запросы
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	// Устанавливаем заголовок, чтобы клиент знал, что мы отвечаем в формате JSON.
	w.Header().Set("Content-Type", "application/json")

	// Создаем структуру для ответа.
	// `json:"status"` - это тег, который говорит, как назвать это поле в JSON.
	response := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}

	// Кодируем нашу структуру в JSON и отправляем ее в качестве ответа.
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Регистрируем наш обработчик.
	// Теперь все запросы на адрес "/api/health" будет обрабатывать функция healthCheckHandler.
	http.HandleFunc("/api/health", healthCheckHandler)

	// Запускаем сервер на порту 8080.
	log.Println("Сервер запускается на порту :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Ошибка при запуске сервера: ", err)
	}
}