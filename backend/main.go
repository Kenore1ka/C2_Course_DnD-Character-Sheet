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

type HitPoints struct {
	Current int `json:"current"`
	Max     int `json:"max"`
}

type AbilityScores struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}

// Character представляет "сырые" данные персонажа.
type Character struct {
	ID                       string
	Name                     string
	Class                    string
	Race                     string
	Alignment                string
	Level                    int
	CurrentHitPoints         int
	AbilityScores            AbilityScores
	SkillProficiencies       []string
	SavingThrowProficiencies []string
}

// CharacterSheet - это полная структура для отправки на фронтенд.
type CharacterSheet struct {
	Name                       string            `json:"name"`
	Class                      string            `json:"class"`
	Race                       string            `json:"race"`
	Alignment                  string            `json:"alignment"`
	Level                      int               `json:"level"`
	ProficiencyBonus           int               `json:"proficiencyBonus"`
	HitPoints                  HitPoints         `json:"hitPoints"`
	ArmorClass                 int               `json:"armorClass"`
	Initiative                 int               `json:"initiative"`
	AbilityScores              AbilityScores     `json:"abilityScores"`
	AbilityModifiers           AbilityScores     `json:"abilityModifiers"`
	SavingThrows               map[string]int    `json:"savingThrows"`
	Skills                     map[string]int    `json:"skills"`
	SkillMap                   map[string]string `json:"skillMap"`
	SkillProficiencies         []string          `json:"skillProficiencies"`
	SavingThrowProficiencies   []string          `json:"savingThrowProficiencies"`
}

// --- КОНФИГУРАЦИЯ И КОНСТАНТЫ ---

var skillAbilityMap = map[string]string{
	"Акробатика":      "Dexterity", "Анализ": "Intelligence", "Атлетика": "Strength",
	"Внимательность":  "Wisdom", "Выживание": "Wisdom", "Выступление": "Charisma",
	"Запугивание":     "Charisma", "История": "Intelligence", "Ловкость рук": "Dexterity",
	"Магия":           "Intelligence", "Медицина": "Wisdom", "Обман": "Charisma",
	"Природа":         "Intelligence", "Проницательность": "Wisdom", "Скрытность": "Dexterity",
	"Убеждение":       "Charisma", "Уход за животными": "Wisdom",
}

type AppConfig struct {
	WelcomeMessage string `json:"welcomeMessage"`
}

var config AppConfig

// --- ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ (БИЗНЕС-ЛОГИКА) ---

func calculateProficiencyBonus(level int) int {
	return 2 + (level-1)/4
}

func calculateModifier(score int) int {
	return int(math.Floor(float64(score-10) / 2))
}

func getModifierByName(modifiers AbilityScores, name string) int {
	switch name {
	case "Strength": return modifiers.Strength
	case "Dexterity": return modifiers.Dexterity
	case "Constitution": return modifiers.Constitution
	case "Intelligence": return modifiers.Intelligence
	case "Wisdom": return modifiers.Wisdom
	case "Charisma": return modifiers.Charisma
	default: return 0
	}
}

func proficiencyBonusIf(condition bool, bonus int) int {
	if condition {
		return bonus
	}
	return 0
}

func createSheetFromCharacter(character Character) CharacterSheet {
	proficiencyBonus := calculateProficiencyBonus(character.Level)
	modifiers := AbilityScores{
		Strength:     calculateModifier(character.AbilityScores.Strength),
		Dexterity:    calculateModifier(character.AbilityScores.Dexterity),
		Constitution: calculateModifier(character.AbilityScores.Constitution),
		Intelligence: calculateModifier(character.AbilityScores.Intelligence),
		Wisdom:       calculateModifier(character.AbilityScores.Wisdom),
		Charisma:     calculateModifier(character.AbilityScores.Charisma),
	}

	savingThrows := make(map[string]int)
	stProfSet := make(map[string]bool)
	for _, prof := range character.SavingThrowProficiencies {
		stProfSet[prof] = true
	}
	savingThrows["strength"] = modifiers.Strength + proficiencyBonusIf(stProfSet["strength"], proficiencyBonus)
	savingThrows["dexterity"] = modifiers.Dexterity + proficiencyBonusIf(stProfSet["dexterity"], proficiencyBonus)
	savingThrows["constitution"] = modifiers.Constitution + proficiencyBonusIf(stProfSet["constitution"], proficiencyBonus)
	savingThrows["intelligence"] = modifiers.Intelligence + proficiencyBonusIf(stProfSet["intelligence"], proficiencyBonus)
	savingThrows["wisdom"] = modifiers.Wisdom + proficiencyBonusIf(stProfSet["wisdom"], proficiencyBonus)
	savingThrows["charisma"] = modifiers.Charisma + proficiencyBonusIf(stProfSet["charisma"], proficiencyBonus)

	skills := make(map[string]int)
	skillProfSet := make(map[string]bool)
	for _, prof := range character.SkillProficiencies {
		skillProfSet[prof] = true
	}
	for skillName, abilityName := range skillAbilityMap {
		skills[skillName] = getModifierByName(modifiers, abilityName) + proficiencyBonusIf(skillProfSet[skillName], proficiencyBonus)
	}
	
	initiative := modifiers.Dexterity
	armorClass := 10 + modifiers.Dexterity
	maxHitPoints := 8 + (modifiers.Constitution * character.Level)
	currentHitPoints := character.CurrentHitPoints
	if currentHitPoints > maxHitPoints {
		currentHitPoints = maxHitPoints
	}

	return CharacterSheet{
		Name:                       character.Name,
		Class:                      character.Class,
		Race:                       character.Race,
		Alignment:                  character.Alignment,
		Level:                      character.Level,
		ProficiencyBonus:           proficiencyBonus,
		HitPoints:                  HitPoints{Current: currentHitPoints, Max: maxHitPoints},
		ArmorClass:                 armorClass,
		Initiative:                 initiative,
		AbilityScores:              character.AbilityScores,
		AbilityModifiers:           modifiers,
		SavingThrows:               savingThrows,
		Skills:                     skills,
		SkillMap:                   skillAbilityMap,
		SkillProficiencies:         character.SkillProficiencies,
		SavingThrowProficiencies:   character.SavingThrowProficiencies,
	}
}

// --- HTTP ОБРАБОТЧИКИ ---

func characterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

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

func getCharacter(w http.ResponseWriter, r *http.Request) {
	mockCharacter := Character{
		ID:                       "1",
		Name:                     "Дриззт До'Урден",
		Class:                    "Следопыт (Ranger)",
		Race:                     "Темный эльф (Дроу)",
		Alignment:                "Хаотично-добрый",
		Level:                    8,
		CurrentHitPoints:         65,
		AbilityScores: AbilityScores{
			Strength: 13, Dexterity: 20, Constitution: 15,
			Intelligence: 17, Wisdom: 17, Charisma: 14,
		},
		SkillProficiencies:       []string{"Акробатика", "Внимательность", "Скрытность", "Выживание"},
		SavingThrowProficiencies: []string{"dexterity", "wisdom"},
	}
	sheet := createSheetFromCharacter(mockCharacter)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sheet)
}

func updateCharacter(w http.ResponseWriter, r *http.Request) {
	var receivedData struct {
		Name                       string        `json:"name"`
		Class                      string        `json:"class"`
		Race                       string        `json:"race"`
		Alignment                  string        `json:"alignment"`
		Level                      int           `json:"level"`
		HitPoints                  HitPoints     `json:"hitPoints"`
		AbilityScores              AbilityScores `json:"abilityScores"`
		SkillProficiencies         []string      `json:"skillProficiencies"`
		SavingThrowProficiencies   []string      `json:"savingThrowProficiencies"`
	}

	if err := json.NewDecoder(r.Body).Decode(&receivedData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tempChar := Character{
		Name:                       receivedData.Name,
		Class:                      receivedData.Class,
		Race:                       receivedData.Race,
		Alignment:                  receivedData.Alignment,
		Level:                      receivedData.Level,
		CurrentHitPoints:           receivedData.HitPoints.Current,
		AbilityScores:              receivedData.AbilityScores,
		SkillProficiencies:         receivedData.SkillProficiencies,
		SavingThrowProficiencies:   receivedData.SavingThrowProficiencies,
	}
	
	updatedSheet := createSheetFromCharacter(tempChar)
	log.Println("Персонаж обновлен (включая инфо):", updatedSheet.Name)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedSheet)
}

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
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/app/config/config.json"
	}
	if err := loadConfig(configPath); err != nil {
		log.Fatalf("Ошибка при загрузке конфигурации: %v", err)
	}

	http.HandleFunc("/api/health", healthCheckHandler)
	http.HandleFunc("/api/character", characterHandler)

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