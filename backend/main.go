// backend/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

// --- ГЛОБАЛЬНЫЕ ПЕРЕМЕННЫЕ ---
var dbpool *pgxpool.Pool
var config AppConfig

// --- СТРУКТURY ДАННЫХ ---
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
type AppConfig struct {
	WelcomeMessage string `json:"welcomeMessage"`
}
type Item struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}
type CharacterItem struct {
	CharacterID int `json:"-"`
	ItemID      int `json:"itemId"`
	Quantity    int `json:"quantity"`
}
type InventoryItem struct {
	Item     Item `json:"item"`
	Quantity int  `json:"quantity"`
}

// --- КОНСТАНТЫ ---
var skillAbilityMap = map[string]string{
	"Акробатика": "Dexterity", "Анализ": "Intelligence", "Атлетика": "Strength",
	"Внимательность": "Wisdom", "Выживание": "Wisdom", "Выступление": "Charisma",
	"Запугивание": "Charisma", "История": "Intelligence", "Ловкость рук": "Dexterity",
	"Магия": "Intelligence", "Медицина": "Wisdom", "Обман": "Charisma",
	"Природа": "Intelligence", "Проницательность": "Wisdom", "Скрытность": "Dexterity",
	"Убеждение": "Charisma", "Уход за животными": "Wisdom",
}

// --- ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ---
func calculateProficiencyBonus(level int) int {
	return 2 + (level-1)/4
}
func calculateModifier(score int) int {
	return int(math.Floor(float64(score-10) / 2))
}
func proficiencyBonusIf(condition bool, bonus int) int {
	if condition {
		return bonus
	}
	return 0
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
func createSheetFromCharacter(character Character) CharacterSheet {
	proficiencyBonus := calculateProficiencyBonus(character.Level)
	modifiers := AbilityScores{
		Strength: calculateModifier(character.AbilityScores.Strength), Dexterity: calculateModifier(character.AbilityScores.Dexterity),
		Constitution: calculateModifier(character.AbilityScores.Constitution), Intelligence: calculateModifier(character.AbilityScores.Intelligence),
		Wisdom: calculateModifier(character.AbilityScores.Wisdom), Charisma: calculateModifier(character.AbilityScores.Charisma),
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
		Name: character.Name, Class: character.Class, Race: character.Race, Alignment: character.Alignment, Level: character.Level,
		ProficiencyBonus: proficiencyBonus, HitPoints: HitPoints{Current: currentHitPoints, Max: maxHitPoints},
		ArmorClass: armorClass, Initiative: initiative, AbilityScores: character.AbilityScores, AbilityModifiers: modifiers,
		SavingThrows: savingThrows, Skills: skills, SkillMap: skillAbilityMap, SkillProficiencies: character.SkillProficiencies,
		SavingThrowProficiencies: character.SavingThrowProficiencies,
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
	var character Character
	query := `SELECT id, name, class, race, alignment, level, current_hit_points, ability_scores, skill_proficiencies, saving_throw_proficiencies FROM characters WHERE id = 1`
	err := dbpool.QueryRow(context.Background(), query).Scan(&character.ID, &character.Name, &character.Class, &character.Race, &character.Alignment, &character.Level, &character.CurrentHitPoints, &character.AbilityScores, &character.SkillProficiencies, &character.SavingThrowProficiencies)
	if err != nil {
		log.Printf("Ошибка при загрузке персонажа из БД: %v", err)
		http.Error(w, "Персонаж не найден", http.StatusInternalServerError)
		return
	}
	sheet := createSheetFromCharacter(character)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sheet)
}

func updateCharacter(w http.ResponseWriter, r *http.Request) {
	var receivedData struct {
		Name                     string        `json:"name"`
		Class                    string        `json:"class"`
		Race                     string        `json:"race"`
		Alignment                string        `json:"alignment"`
		Level                    int           `json:"level"`
		HitPoints                HitPoints     `json:"hitPoints"`
		AbilityScores            AbilityScores `json:"abilityScores"`
		SkillProficiencies       []string      `json:"skillProficiencies"`
		SavingThrowProficiencies []string      `json:"savingThrowProficiencies"`
	}
	if err := json.NewDecoder(r.Body).Decode(&receivedData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `UPDATE characters SET name = $1, class = $2, race = $3, alignment = $4, level = $5, current_hit_points = $6, ability_scores = $7, skill_proficiencies = $8, saving_throw_proficiencies = $9 WHERE id = 1`
	_, err := dbpool.Exec(context.Background(), query, receivedData.Name, receivedData.Class, receivedData.Race, receivedData.Alignment, receivedData.Level, receivedData.HitPoints.Current, receivedData.AbilityScores, receivedData.SkillProficiencies, receivedData.SavingThrowProficiencies)
	if err != nil {
		log.Printf("Ошибка при обновлении персонажа в БД: %v", err)
		http.Error(w, "Не удалось обновить персонажа", http.StatusInternalServerError)
		return
	}
	log.Println("Персонаж с ID=1 обновлен в базе данных.")
	getCharacter(w, r)
}

func inventoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	switch r.Method {
	case http.MethodGet:
		getInventory(w, r)
	case http.MethodPost:
		addItemToInventory(w, r)
	default:
		http.Error(w, "Метод не поддерживается для этого пути", http.StatusMethodNotAllowed)
	}
}

func inventoryItemHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	switch r.Method {
	case http.MethodDelete:
		deleteItemFromInventory(w, r)
	default:
		http.Error(w, "Метод не поддерживается для этого пути", http.StatusMethodNotAllowed)
	}
}

func getInventory(w http.ResponseWriter, r *http.Request) {
	characterID := 1
	query := `SELECT i.id, i.name, i.type, i.description, ci.quantity FROM items i JOIN character_items ci ON i.id = ci.item_id WHERE ci.character_id = $1`
	rows, err := dbpool.Query(context.Background(), query, characterID)
	if err != nil {
		log.Printf("Ошибка при загрузке инвентаря из БД: %v", err)
		http.Error(w, "Не удалось загрузить инвентарь", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var inventory []InventoryItem
	for rows.Next() {
		var invItem InventoryItem
		if err := rows.Scan(&invItem.Item.ID, &invItem.Item.Name, &invItem.Item.Type, &invItem.Item.Description, &invItem.Quantity); err != nil {
			log.Printf("Ошибка при сканировании строки инвентаря: %v", err)
			continue
		}
		inventory = append(inventory, invItem)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Ошибка при итерации строк инвентаря: %v", err)
		http.Error(w, "Ошибка при обработке инвентаря", http.StatusInternalServerError)
		return
	}
	if inventory == nil {
		inventory = make([]InventoryItem, 0)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inventory)
}

func addItemToInventory(w http.ResponseWriter, r *http.Request) {
	characterID := 1
	var newItem CharacterItem
	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `INSERT INTO character_items (character_id, item_id, quantity) VALUES ($1, $2, $3) ON CONFLICT (character_id, item_id) DO UPDATE SET quantity = character_items.quantity + EXCLUDED.quantity;`
	_, err := dbpool.Exec(context.Background(), query, characterID, newItem.ItemID, newItem.Quantity)
	if err != nil {
		log.Printf("Ошибка при добавлении предмета в инвентарь: %v", err)
		http.Error(w, "Не удалось добавить предмет", http.StatusInternalServerError)
		return
	}
	log.Printf("Предмет %d (кол-во: %d) добавлен/обновлен в инвентаре персонажа %d", newItem.ItemID, newItem.Quantity, characterID)
	getInventory(w, r)
}

func deleteItemFromInventory(w http.ResponseWriter, r *http.Request) {
	characterID := 1
	itemIDStr := r.URL.Path[len("/api/character/inventory/item/"):]
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Некорректный ID предмета", http.StatusBadRequest)
		return
	}
	query := `DELETE FROM character_items WHERE character_id = $1 AND item_id = $2`
	commandTag, err := dbpool.Exec(context.Background(), query, characterID, itemID)
	if err != nil {
		log.Printf("Ошибка при удалении предмета из инвентаря: %v", err)
		http.Error(w, "Не удалось удалить предмет", http.StatusInternalServerError)
		return
	}
	if commandTag.RowsAffected() > 0 {
		log.Printf("Предмет %d удален из инвентаря персонажа %d", itemID, characterID)
	}
	getInventory(w, r)
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{Status: "ok", Message: config.WelcomeMessage})
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
	if err := loadConfig(os.Getenv("CONFIG_PATH")); err != nil {
		if err := loadConfig("/app/config/config.json"); err != nil {
			log.Fatalf("Ошибка при загрузке конфигурации: %v", err)
		}
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", os.Getenv("DATABASE_USER"), os.Getenv("DATABASE_PASSWORD"), os.Getenv("DATABASE_HOST"), os.Getenv("DATABASE_PORT"), os.Getenv("DATABASE_NAME"))
	var err error
	dbpool, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer dbpool.Close()
	log.Println("Успешно подключено к базе данных!")

	// Регистрация обработчиков
	http.HandleFunc("/api/health", healthCheckHandler)
	http.HandleFunc("/api/character", characterHandler)
	http.HandleFunc("/api/character/inventory", inventoryHandler)
	http.HandleFunc("/api/character/inventory/item/", inventoryItemHandler)

	// Запуск сервера
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Сервер запускается на порту :", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Ошибка при запуске сервера: ", err)
	}
}