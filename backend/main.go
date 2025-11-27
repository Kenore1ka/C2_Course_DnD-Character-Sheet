// backend/main.go
package main

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"os"
)

// --- СТРУКТУРЫ ДАННЫХ ---

// AbilityScores хранит значения шести основных характеристик.
// json:"..." - это теги, которые говорят Go, как называть эти поля при преобразовании в JSON.
type AbilityScores struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}

// Character представляет "сырые" данные персонажа, как если бы они хранились в базе данных.
type Character struct {
	ID            string
	Name          string
	AbilityScores AbilityScores
}

// CharacterSheet - это полная структура, которую мы отправляем на фронтенд.
// Она содержит как базовые данные, так и все вычисляемые значения (модификаторы, навыки).
type CharacterSheet struct {
	Name             string            `json:"name"`
	AbilityScores    AbilityScores     `json:"abilityScores"`
	AbilityModifiers AbilityScores     `json:"abilityModifiers"`
	Skills           map[string]int    `json:"skills"`
	SkillMap         map[string]string `json:"skillMap"` // Карта "Навык -> Характеристика" для удобства фронтенда
}

// --- КОНФИГУРАЦИЯ И КОНСТАНТЫ ---

// skillAbilityMap определяет, какая характеристика используется для какого навыка.
var skillAbilityMap = map[string]string{
	"Акробатика":      "Dexterity",
	"Анализ":          "Intelligence",
	"Атлетика":        "Strength",
	"Внимательность":  "Wisdom",
	"Выживание":       "Wisdom",
	"Выступление":     "Charisma",
	"Запугивание":     "Charisma",
	"История":         "Intelligence",
	"Ловкость рук":    "Dexterity",
	"Магия":           "Intelligence",
	"Медицина":        "Wisdom",
	"Обман":           "Charisma",
	"Природа":         "Intelligence",
	"Проницательность": "Wisdom",
	"Скрытность":      "Dexterity",
	"Убеждение":       "Charisma",
	"Уход за животными": "Wisdom",
}

// Структура и переменная для хранения конфигурации из config.json
type AppConfig struct {
	WelcomeMessage string `json:"welcomeMessage"`
}

var config AppConfig

// --- ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ (БИЗНЕС-ЛОГИКА) ---

// calculateModifier вычисляет модификатор характеристики по правилам D&D 5e.
func calculateModifier(score int) int {
	return int(math.Floor(float64(score-10) / 2))
}

// getModifierByName - удобная функция для получения модификатора по его строковому названию.
func getModifierByName(modifiers AbilityScores, name string) int {
	switch name {
	case "Strength":
		return modifiers.Strength
	case "Dexterity":
		return modifiers.Dexterity
	case "Constitution":
		return modifiers.Constitution
	case "Intelligence":
		return modifiers.Intelligence
	case "Wisdom":
		return modifiers.Wisdom
	case "Charisma":
		return modifiers.Charisma
	default:
		return 0
	}
}

// createSheetFromCharacter - центральная функция, которая принимает "сырые" данные персонажа
// и возвращает полностью рассчитанный лист для отправки на фронтенд.
func createSheetFromCharacter(character Character) CharacterSheet {
	// 1. Рассчитываем модификаторы
	modifiers := AbilityScores{
		Strength:     calculateModifier(character.AbilityScores.Strength),
		Dexterity:    calculateModifier(character.AbilityScores.Dexterity),
		Constitution: calculateModifier(character.AbilityScores.Constitution),
		Intelligence: calculateModifier(character.AbilityScores.Intelligence),
		Wisdom:       calculateModifier(character.AbilityScores.Wisdom),
		Charisma:     calculateModifier(character.AbilityScores.Charisma),
	}

	// 2. Рассчитываем навыки на основе модификаторов
	skills := make(map[string]int)
	for skillName, abilityName := range skillAbilityMap {
		skills[skillName] = getModifierByName(modifiers, abilityName)
	}

	// 3. Собираем и возвращаем финальный объект
	return CharacterSheet{
		Name:             character.Name,
		AbilityScores:    character.AbilityScores,
		AbilityModifiers: modifiers,
		Skills:           skills,
		SkillMap:         skillAbilityMap,
	}
}

// --- HTTP ОБРАБОТЧИКИ ---

// characterHandler - это универсальный обработчик для эндпоинта /api/character.
// Он определяет, какой метод был использован (GET или POST), и вызывает соответствующую функцию.
func characterHandler(w http.ResponseWriter, r *http.Request) {
	// Устанавливаем заголовки CORS, чтобы разрешить запросы от нашего фронтенда
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Браузер может отправить "preflight" запрос методом OPTIONS перед POST.
	// Нам нужно просто ответить на него, что всё в порядке.
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getCharacter(w, r)
	case http.MethodPost:
		updateCharacter(w, r)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// getCharacter обрабатывает GET-запросы: создает персонажа-заглушку и отправляет его.
func getCharacter(w http.ResponseWriter, r *http.Request) {
	mockCharacter := Character{
		ID:   "1",
		Name: "Дриззт До'Урден",
		AbilityScores: AbilityScores{
			Strength: 13, Dexterity: 20, Constitution: 15,
			Intelligence: 17, Wisdom: 17, Charisma: 14,
		},
	}
	sheet := createSheetFromCharacter(mockCharacter)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sheet)
}

// updateCharacter обрабатывает POST-запросы: принимает обновленные данные,
// пересчитывает все значения и отправляет полный лист обратно.
func updateCharacter(w http.ResponseWriter, r *http.Request) {
	// Структура для приема данных от фронтенда. Принимаем только то, что может изменить пользователь.
	var receivedData struct {
		Name          string        `json:"name"`
		AbilityScores AbilityScores `json:"abilityScores"`
	}

	if err := json.NewDecoder(r.Body).Decode(&receivedData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Создаем временный объект, чтобы передать его в нашу центральную логику
	tempChar := Character{Name: receivedData.Name, AbilityScores: receivedData.AbilityScores}
	
	// Используем нашу универсальную функцию, которая пересчитает всё: и модификаторы, и навыки
	updatedSheet := createSheetFromCharacter(tempChar)

	log.Println("Персонаж обновлен (включая навыки):", updatedSheet.Name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedSheet)
}


// healthCheckHandler и loadConfig - служебные функции из предыдущих шагов.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Status: "ok", Message: config.WelcomeMessage,
	}
	json.NewEncoder(w).Encode(response)
}

func loadConfig(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, &config)
}


// --- ТОЧКА ВХОДА В ПРИЛОЖЕНИЕ ---

func main() {
	// Загружаем конфигурацию
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/app/config/config.json"
	}
	if err := loadConfig(configPath); err != nil {
		log.Fatalf("Ошибка при загрузке конфигурации: %v", err)
	}

	// Регистрируем наши HTTP обработчики
	http.HandleFunc("/api/health", healthCheckHandler)
	http.HandleFunc("/api/character", characterHandler)

	// Получаем порт из переменных окружения и запускаем сервер
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Сервер запускается на порту :", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("Ошибка при запуске сервера: ", err)
	}
}