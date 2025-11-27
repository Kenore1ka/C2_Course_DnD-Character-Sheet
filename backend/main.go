package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type AppConfig struct {
	WelcomeMessage string `json:"welcomeMessage"`
}

// Глобальная переменная для хранения конфига
var config AppConfig

// Функция для загрузки конфигурации из файла
func loadConfig(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, &config)
}

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
		Message string `json:"message"`
	}{
		Status: "ok",
		Message: config.WelcomeMessage, // <-- Используем данные из конфига
	}

	// Кодируем нашу структуру в JSON и отправляем ее в качестве ответа.
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Загружаем конфиг. Путь к файлу будем передавать через переменную окружения.
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/app/config/config.json" // Путь по умолчанию внутри контейнера
	}

	if err := loadConfig(configPath); err != nil {
		log.Fatalf("Ошибка при загрузке конфигурации: %v", err)
	}

	http.HandleFunc("/api/health", healthCheckHandler)

	// Получаем порт из переменной окружения
	port := os.Getenv("APP_PORT")
	// Если переменная не установлена, используем порт 8080 по умолчанию
	if port == "" {
		port = "8080"
	}

	log.Println("Сервер запускается на порту :", port) // <-- Изменяем лог
	// Запускаем сервер на нужном порту
	log.Println("Сообщение из конфига:", config.WelcomeMessage) // <-- Выведем в лог для проверки
	err := http.ListenAndServe(":"+port, nil) // <-- Используем переменную port
	if err != nil {
		log.Fatal("Ошибка при запуске сервера: ", err)
	}
}