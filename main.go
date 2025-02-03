package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)
import _ "net/http/pprof"
import _ "github.com/lib/pq"

type Player struct {
	ID                      string          `json:"id"`
	Conn                    *websocket.Conn `json:"-"`
	Role                    string          `json:"role"`
	IsAlive                 bool            `json:"is_alive"`
	VotedFor                string          `json:"voted_for"`
	Action                  string          `json:"action"` // Used for night actions
	Aura                    string          `json:"aura"`
	TargetedScreamerPlayer  string          `json:"targeted_screamer_player"`
	TargetedSunFlowerPlayer string          `json:"targeted_sun_flower_player"`
	Hacked                  bool            `json:"hacked"`
}

type Game struct {
	Players       map[string]*Player
	Roles         []string
	Phase         string
	Votes         map[string]int
	Mutex         sync.Mutex
	GameStarted   bool
	CurrentPhase  string
	DayNumber     int
	TimeRemaining int
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var game = Game{
	Players: make(map[string]*Player),
	Votes:   make(map[string]int),
	Roles:   []string{"mafia", "detective", "villager", "villager"}, // Example roles
}

type Room struct {
	ID      string             `json:"id"`
	Players map[string]*Player `json:"players"` // Ключ – ID игрока
	// Здесь можно добавить дополнительные поля (например, состояние игры, таймер, и т.д.)
}

var (
	rooms    = make(map[string]*Room) // все созданные комнаты
	roomLock sync.Mutex               // для синхронизации доступа к rooms
)

func generateRoomID() string {
	return fmt.Sprintf("room-%d", time.Now().UnixNano()+int64(rand.Intn(1000)))
}

// joinRoom ищет свободную комнату (где количество игроков < 16) и выбирает ту, где уже больше всего игроков.
// Если ни одной такой комнаты нет, создается новая.
func joinRoom(p *Player) *Room {
	roomLock.Lock()
	defer roomLock.Unlock()

	var bestRoom *Room
	for _, room := range rooms {
		if len(room.Players) < 16 {
			if bestRoom == nil || len(room.Players) > len(bestRoom.Players) {
				bestRoom = room
			}
		}
	}
	if bestRoom == nil {
		// Нет свободной комнаты — создаём новую
		bestRoom = &Room{
			ID:      generateRoomID(),
			Players: make(map[string]*Player),
		}
		rooms[bestRoom.ID] = bestRoom
		log.Printf("Создана новая комната: %s", bestRoom.ID)
	}
	// Добавляем игрока в выбранную комнату
	bestRoom.Players[p.ID] = p
	log.Printf("Игрок %s добавлен в комнату %s (игроков: %d)", p.ID, bestRoom.ID, len(bestRoom.Players))
	return bestRoom
}

// joinRoomHandler — HTTP-обработчик для присоединения к игровой комнате.
// Если запись о игроке не найдена в game.Players, создаём её.
func joinRoomHandler(w http.ResponseWriter, r *http.Request) {
	// Ожидается, что игрок передаст свой идентификатор в параметре "id"
	playerID := r.URL.Query().Get("id")
	if playerID == "" {
		http.Error(w, "ID игрока не указан", http.StatusBadRequest)
		return
	}

	// Проверяем, существует ли запись об игроке в game.Players
	game.Mutex.Lock()
	player, exists := game.Players[playerID]
	if !exists {
		// Если записи нет, создаём нового игрока (без активного WebSocket, пока что)
		player = &Player{
			ID:      playerID,
			IsAlive: true,
		}
		game.Players[playerID] = player
		log.Printf("Создан новый игрок %s через joinRoomHandler", playerID)
	}
	game.Mutex.Unlock()

	// Теперь пытаемся добавить игрока в комнату
	room := joinRoom(player)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"roomId":  room.ID,
		"players": len(room.Players),
	})
}

// leaveRoomHandler удаляет игрока из комнаты и из глобальной карты game.Players.
func leaveRoomHandler(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("id")
	if playerID == "" {
		http.Error(w, "ID игрока не указан", http.StatusBadRequest)
		return
	}

	// Удаляем игрока из глобальной карты комнат.
	roomLock.Lock()
	for _, room := range rooms {
		if _, exists := room.Players[playerID]; exists {
			delete(room.Players, playerID)
			log.Printf("Игрок %s покинул комнату %s", playerID, room.ID)
			break
		}
	}
	roomLock.Unlock()

	// Также удаляем игрока из глобальной карты game.Players.
	game.Mutex.Lock()
	delete(game.Players, playerID)
	game.Mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

var db *sql.DB

// registerHandler – обработчик регистрации пользователя (пример)
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	type RegistrationRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var req RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Не указано имя или пароль", http.StatusBadRequest)
		return
	}

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)", req.Username).Scan(&exists)
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Пользователь с таким именем уже существует", http.StatusBadRequest)
		return
	}

	// Здесь пароль сохраняется в открытом виде — для демонстрации.
	_, err = db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", req.Username, req.Password)
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Пользователь успешно зарегистрирован",
	})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Устанавливаем заголовок, чтобы клиент ожидал JSON
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Метод не поддерживается"})
		return
	}

	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат данных"})
		return
	}

	if req.Username == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Не указано имя или пароль"})
		return
	}

	// Извлекаем пароль пользователя из БД
	var storedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username=$1", req.Username).Scan(&storedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Пользователь не найден"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
		return
	}

	// Для демонстрационных целей сравниваем пароли как есть (в реальном приложении используйте хэширование)
	if req.Password != storedPassword {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный пароль"})
		return
	}

	// Если всё успешно, возвращаем JSON с успехом
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Вход выполнен успешно",
	})
}

// serveProfile отдает страницу профиля (profile.html)
func serveProfile(w http.ResponseWriter, r *http.Request) {
	// Можно, например, проверить наличие параметра id в URL.
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID не указан", http.StatusBadRequest)
		return
	}
	http.ServeFile(w, r, "./static/profile.html")
}

// serveWelcome отдает страницу приветствия (welcome.html)
func serveWelcome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/welcome.html")
}

// serveGame отдает игровую страницу (index.html)
func serveGame(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/index.html")
}

func main() {
	sysConnStr := "host=localhost port=5432 user=postgres password=123 dbname=postgres sslmode=disable"
	sysDB, err := sql.Open("postgres", sysConnStr)
	if err != nil {
		log.Fatal("Ошибка подключения к системной БД:", err)
	}
	defer sysDB.Close()

	if err = sysDB.Ping(); err != nil {
		log.Fatal("Ошибка подключения к системной БД:", err)
	}

	// Имя целевой базы данных
	targetDBName := "mafia_game"

	// Проверяем, существует ли база данных targetDBName.
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)"
	err = sysDB.QueryRow(checkQuery, targetDBName).Scan(&exists)
	if err != nil {
		log.Fatal("Ошибка проверки существования базы данных:", err)
	}

	if !exists {
		log.Printf("База данных %s не существует. Создаем...", targetDBName)
		// Создаем базу данных.
		_, err = sysDB.Exec(fmt.Sprintf("CREATE DATABASE %s", targetDBName))
		if err != nil {
			log.Fatal("Ошибка создания базы данных:", err)
		}
		log.Printf("База данных %s успешно создана.", targetDBName)
	} else {
		log.Printf("База данных %s уже существует.", targetDBName)
	}

	// Закрываем системное соединение и подключаемся к целевой базе данных.
	targetConnStr := fmt.Sprintf("host=localhost port=5432 user=postgres password=123 dbname=%s sslmode=disable", targetDBName)
	db, err = sql.Open("postgres", targetConnStr)
	if err != nil {
		log.Fatal("Ошибка подключения к целевой базе данных:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Ошибка подключения к целевой базе данных:", err)
	}

	// Создаем таблицу users, если её нет.
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
	);
	`
	if _, err := db.Exec(createTableQuery); err != nil {
		log.Fatal("Ошибка создания таблицы:", err)
	}

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/profile", serveProfile) // Страница профиля
	// Обработчик для приветственной страницы
	http.HandleFunc("/", serveWelcome)
	// Обработчик для игровой страницы – пользователь должен переходить по /game?id=<имя_пользователя>
	http.HandleFunc("/game", serveGame)
	// WebSocket и остальные обработчики остаются без изменений
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/status", gameStatus)
	http.HandleFunc("/joinroom", joinRoomHandler)
	http.HandleFunc("/leaveroom", leaveRoomHandler)

	// Если вам нужны статические файлы (например, картинки, css, js) – можно раздать их по префиксу /static/
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Глобальный мап для хранения ID игроков, которые покинули игру навсегда.
var disconnectedPlayers = make(map[string]bool)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Обновляем соединение через upgrader
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	// Получаем ID игрока из параметра URL. Если не указан, генерируем.
	playerID := r.URL.Query().Get("id")
	if playerID == "" {
		playerID = fmt.Sprintf("player-%d", len(game.Players)+1)
	}

	// Блокируем глобальный мьютекс для проверки и добавления игрока.
	game.Mutex.Lock()
	// Если игрок уже покинул игру навсегда, не разрешаем новое подключение.
	if disconnectedPlayers[playerID] {
		game.Mutex.Unlock()
		log.Printf("Reject connection: player %s is marked as disconnected permanently", playerID)
		conn.WriteMessage(websocket.CloseMessage, []byte("You have left the game permanently"))
		return
	}
	// Если уже существует активное соединение с этим ID, отклоняем второе.
	if existing, ok := game.Players[playerID]; ok && existing.Conn != nil {
		game.Mutex.Unlock()
		log.Printf("Reject connection: player %s is already connected", playerID)
		conn.WriteMessage(websocket.CloseMessage, []byte("Player already connected"))
		return
	}

	// Создаем нового игрока и добавляем его в глобальную карту.
	player := &Player{
		ID:      playerID,
		Conn:    conn,
		IsAlive: true,
	}
	game.Players[playerID] = player

	// Формируем снимок активных игроков (только с открытыми соединениями)
	playersSnapshot := make(map[string]bool)
	for id, p := range game.Players {
		if p.Conn != nil { // только активные соединения
			playersSnapshot[id] = p.IsAlive
		}
	}
	game.Mutex.Unlock()

	// Отправляем новому клиенту начальное сообщение со списком игроков
	initialStatus := struct {
		Type    string          `json:"type"`
		Players map[string]bool `json:"players"`
	}{
		Type:    "playerList",
		Players: playersSnapshot,
	}
	if err := conn.WriteJSON(initialStatus); err != nil {
		log.Printf("Ошибка отправки начального состояния игроку %s: %v", playerID, err)
	}

	// Немедленно обновляем список для всех подключенных клиентов
	broadcastPlayerList()

	log.Printf("Player %s connected. Total active players: %d", playerID, len(game.Players))
	// Основной цикл чтения сообщений
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Player %s disconnected: %v", playerID, err)
			// При разрыве соединения удаляем игрока и помечаем его как отключённого
			game.Mutex.Lock()
			delete(game.Players, playerID)
			disconnectedPlayers[playerID] = true
			game.Mutex.Unlock()

			// Если используется логика комнат, удаляем игрока и из них (если есть)
			roomLock.Lock()
			for _, room := range rooms {
				if _, exists := room.Players[playerID]; exists {
					delete(room.Players, playerID)
					log.Printf("Player %s removed from room %s", playerID, room.ID)
				}
			}
			roomLock.Unlock()

			// Обновляем список активных игроков для оставшихся клиентов
			broadcastPlayerList()
			break
		}
		log.Printf("Message from %s: %s", playerID, string(message))
		processMessage(playerID, message)
	}
}

// Функция для рассылки обновлённого списка активных игроков
func broadcastPlayerList() {
	game.Mutex.Lock()
	playersSnapshot := make(map[string]bool)
	for id, p := range game.Players {
		if p.Conn != nil { // учитываем только активные соединения
			playersSnapshot[id] = p.IsAlive
		}
	}
	game.Mutex.Unlock()

	update := struct {
		Type    string          `json:"type"`
		Players map[string]bool `json:"players"`
	}{
		Type:    "playerList",
		Players: playersSnapshot,
	}

	game.Mutex.Lock()
	for _, p := range game.Players {
		if p.Conn != nil {
			if err := p.Conn.WriteJSON(update); err != nil {
				log.Printf("Ошибка отправки списка игрока %s: %v", p.ID, err)
			}
		}
	}
	game.Mutex.Unlock()
}

func startGame(w http.ResponseWriter, r *http.Request) {
	game.Mutex.Lock()
	//log.Println("Mutex Locked")
	game.Mutex.Unlock()
	//log.Println("Mutex UNLocked")

	if game.GameStarted {
		http.Error(w, "Game already started", http.StatusBadRequest)
		return
	}

	if len(game.Players) < 4 {
		http.Error(w, "Not enough players to start the game", http.StatusBadRequest)
		return
	}
	game.Roles = generateRoles(len(game.Players))
	log.Println("Starting game...")
	assignRoles()
	game.GameStarted = true
	game.DayNumber = 1
	startDayPhase()
}

func assignRoles() {
	roles := shuffleRoles(game.Roles)
	index := 0
	for _, player := range game.Players {
		player.Role = roles[index]
		if player.Role == "Альфа оборотень" || player.Role == "Волчий провидец" || player.Role == "Малыш оборотень" || player.Role == "Волчий страж" {
			player.Aura = "bad"
		} else if player.Role == "Шут" || player.Role == "Хакер" || player.Role == "Тюремщик" || player.Role == "Линчеватель" {
			player.Aura = "unknown"
		} else {
			player.Aura = "good"
		}
		index++
		log.Printf("Assigned role %s to player %s", player.Role, player.ID)
	}
	broadcastRoles()
}

func generateRoles(playerCount int) []string {
	roles := []string{}

	//// Добавляем мафию (1 мафия на каждые 4 игрока)
	//mafiaCount := playerCount / 4
	//for i := 0; i < mafiaCount; i++ {
	//	roles = append(roles, "mafia")
	//}
	//
	//// Добавляем детектива (1 детектив на каждые 6 игроков)
	//if playerCount >= 6 {
	//	roles = append(roles, "detective")
	//}
	//
	//// Добавляем доктора, если игроков больше 5
	//if playerCount >= 5 {
	//	roles = append(roles, "doctor")
	//}
	//
	//// Остальные роли - мирные жители
	//villagerCount := playerCount - len(roles)
	//for i := 0; i < villagerCount; i++ {
	//	roles = append(roles, "villager")
	//}

	switch playerCount {
	case 4:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор"}
	case 5:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун"}
	case 6:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун", "Дитя цветов"}
	case 7:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун", "Дитя цветов", "Хакер"}
	case 8:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун", "Дитя цветов", "Хакер", "Волчий провидец"}
	case 9:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун", "Дитя цветов", "Хакер", "Волчий провидец", "Медиум"}
	case 10:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун", "Дитя цветов", "Хакер", "Волчий провидец", "Медиум", "Тюремщик"}
	case 11:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун", "Дитя цветов", "Хакер", "Волчий провидец", "Медиум", "Тюремщик", "Линчеватель"}
	case 12:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун", "Дитя цветов", "Хакер", "Волчий провидец", "Медиум", "Тюремщик", "Линчеватель", "Малыш оборотень"}
	case 13:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун", "Дитя цветов", "Хакер", "Волчий провидец", "Медиум", "Тюремщик", "Линчеватель", "Малыш оборотень", "Провидец ауры"}
	case 14:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун", "Дитя цветов", "Хакер", "Волчий провидец", "Медиум", "Тюремщик", "Линчеватель", "Малыш оборотень", "Провидец ауры", "Охотник на зверей"}
	case 15:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун", "Дитя цветов", "Хакер", "Волчий провидец", "Медиум", "Тюремщик", "Линчеватель", "Малыш оборотень", "Провидец ауры", "Охотник на зверей", "Купидон"}
	case 16:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун", "Дитя цветов", "Хакер", "Волчий провидец", "Медиум", "Тюремщик", "Линчеватель", "Малыш оборотень", "Провидец ауры", "Охотник на зверей", "Купидон", "Волчий страж"}

	}

	return roles
}

func shuffleRoles(roles []string) []string {
	shuffled := make([]string, len(roles))
	copy(shuffled, roles)
	for i := range shuffled {
		j := i + rand.Intn(len(shuffled)-i)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	return shuffled
}

func broadcastRoles() {
	for _, player := range game.Players {
		roleMessage, _ := json.Marshal(struct {
			Role string `json:"role"`
		}{
			Role: player.Role,
		})
		player.Conn.WriteMessage(websocket.TextMessage, roleMessage)
	}
}

func startDayPhase() {
	//log.Println("1")
	game.Mutex.Lock()
	//log.Println("Mutex Locked")
	//log.Println("2")
	game.CurrentPhase = "day"
	game.Votes = make(map[string]int)
	log.Println("Day phase started.")
	broadcastGameStatus() // Отправить клиентам обновление о фазе
	//log.Println("3")
	game.Mutex.Unlock()
	//log.Println("Mutex UNLocked")
	//log.Println("4")
	startPhaseTimer(30, endDayPhase)
}

func startNightPhase() {
	game.Mutex.Lock()
	//log.Println("Mutex Locked")
	game.CurrentPhase = "night"
	log.Println("Night phase started.")
	broadcastGameStatus() // Отправить клиентам обновление о фазе
	game.Mutex.Unlock()
	//log.Println("Mutex UNLocked")
	startPhaseTimer(30, func() {
		log.Println("Night phase timer ended.")
		processNightActions()
		endNightPhase()
	})
}

func processNightActions() {
	log.Println("Processing night actions...")

	// Собираем голоса (действия) только от оборотней
	werewolfVotes := make(map[string]int)
	nightActions := make(map[string]string)
	game.Mutex.Lock()
	log.Println("#5")
	for _, player := range game.Players {
		if player.Action != "" && player.IsAlive {
			nightActions[player.ID] = player.Action
			log.Println("####!!!", player.ID, player.Action)
			log.Println("#6")
		}
		player.Action = "" // Reset actions after processing
	}

	aliveWerewolves := 0
	for _, player := range game.Players {
		if player.IsAlive && player.Aura == "bad" {
			aliveWerewolves++
		}
	}
	doctorTarget := ""
	hackerTarget := ""
	game.Mutex.Unlock()
	log.Println("#7")
	// Mafia's action: eliminate a player
	for id, targetID := range nightActions {
		p := game.Players[id]
		// Если aura=bad и игрок жив, учитываем его голос
		log.Println("####id-targetid", id, targetID)
		if p != nil && p.IsAlive && p.Aura == "bad" {
			werewolfVotes[targetID]++
			log.Println("####", targetID, werewolfVotes[targetID])
		}
		if p != nil && p.IsAlive && p.Role == "Доктор" {
			doctorTarget = targetID
			log.Println("####doctorTarget", doctorTarget)

		}
		if p != nil && p.IsAlive && p.Role == "Хакер" {
			hackerTarget = targetID
			log.Printf("Hacker targeted %s", hackerTarget)

		}
	}

	if hackerTarget != "" {
		if target, exists := game.Players[hackerTarget]; exists {
			target.Hacked = true
			log.Printf("Player %s has been hacked and will lose voting/chat rights", target.ID)
			message, _ := json.Marshal(struct {
				PlayerID string `json:"playerID"`
				Chat     string `json:"chat"`
			}{
				PlayerID: "[SERVER]",
				Chat:     "Вы были взломаны! Вы не можете голосовать и писать в чат. Вы погибните в конце дня.",
			})

			target.Conn.WriteMessage(websocket.TextMessage, message)
		}
	}

	// С threshold определяем, сколько нужно голосов
	// Для упрощения логики используем округление вверх: (aliveWerewolves/2 + 1), если нечётно
	voteThreshold := aliveWerewolves / 2
	if aliveWerewolves%2 != 0 {
		voteThreshold = aliveWerewolves/2 + 1
	}

	// Определяем лидера голосования среди оборотней
	maxVotes := 0
	var candidates []string
	for targetID, count := range werewolfVotes {
		if count > maxVotes {
			maxVotes = count
			candidates = []string{targetID}
		} else if count == maxVotes {
			candidates = append(candidates, targetID)
		}
	}

	log.Printf("[Night] Werewolf votes: %v, threshold=%d, maxVotes=%d, candidates=%v",
		werewolfVotes, voteThreshold, maxVotes, candidates,
	)

	// Убийство совершается, только если:
	// 1) Есть ровно один лидер (candidates имеет длину 1)
	// 2) Лидер набрал >= порога
	if len(candidates) == 1 && maxVotes >= voteThreshold {
		targetID := candidates[0]
		targetPlayer, ok := game.Players[targetID]
		if ok && targetPlayer.IsAlive && targetID != doctorTarget {
			if targetPlayer.Role == "Крикун" {
				log.Println("##Крикун1")
				if targetPlayer.TargetedScreamerPlayer != "" {
					targetPlayer := game.Players[targetPlayer.TargetedScreamerPlayer]
					if targetPlayer != nil {
						log.Println("##Крикун2")
						broadcastChatMessage("[SERVER]", fmt.Sprintf("Крикун раскрыл роль игрока %s - %s", targetPlayer.ID, targetPlayer.Role))
					}
				}
			}
			targetPlayer.IsAlive = false
			log.Printf("[Night] Werewolves killed player %s", targetID)
		}
	} else {
		log.Println("[Night] No one was killed by werewolves this night.")
	}

	// Detective's action: check a player's role
	for id, action := range nightActions {
		if game.Players[id].Role == "Провидец" {
			if target, exists := game.Players[action]; exists {
				log.Printf("Detective checked player %s, role: %s", target.ID, target.Role)
				message := fmt.Sprintf("Player %s is %s", target.ID, target.Role)
				log.Printf("Sending message to detective %s: %s", id, message)
				teamCheckMessage, _ := json.Marshal(struct {
					Team string `json:"team"`
				}{
					Team: target.Role,
				})
				game.Players[id].Conn.WriteMessage(websocket.TextMessage, teamCheckMessage)
			}
		}
		if game.Players[id].Role == "Провидец ауры" {
			if target, exists := game.Players[action]; exists {
				log.Printf("Aura seer checked player %s, aura: %s", target.ID, target.Aura)
				message := fmt.Sprintf("Player %s is %s", target.ID, target.Role)
				log.Printf("Sending message to detective %s: %s", id, message)
				teamCheckMessage, _ := json.Marshal(struct {
					Team string `json:"team"`
				}{
					Team: target.Aura,
				})
				game.Players[id].Conn.WriteMessage(websocket.TextMessage, teamCheckMessage)
			}
		}
	}
}

func endDayPhase() {
	log.Println("Ending day phase. Processing votes...")
	processVotes()

	if gameOver, winner := checkGameOver(); gameOver {
		log.Println(winner)
		broadcastWinner(winner)
		game.GameStarted = false // Останавливаем игру
		return
	}

	startNightPhase()
}

func endNightPhase() {
	log.Println("Ending night phase. Starting new day.")

	if gameOver, winner := checkGameOver(); gameOver {
		log.Println(winner)
		broadcastWinner(winner)
		game.GameStarted = false // Останавливаем игру
		return
	}

	game.DayNumber++
	startDayPhase()
}

func checkGameOver() (bool, string) {
	aliveMafia := 0
	aliveVillagers := 0
	hackerAlive := false
	for _, player := range game.Players {
		if player.IsAlive {
			if player.Role == "Хакер" {
				hackerAlive = true
			}
			if player.Aura == "bad" {
				aliveMafia++
			} else {
				aliveVillagers++
			}
		}
	}

	if aliveMafia == 0 && !hackerAlive {
		return true, "Villagers win!"
	}

	if aliveMafia >= aliveVillagers && !hackerAlive {
		return true, "Mafia wins!"
	}
	if hackerAlive && aliveVillagers == 0 && aliveMafia == 0 {
		return true, "Hacker win"
	}

	return false, ""
}

func broadcastWinner(winner string) {
	message, _ := json.Marshal(struct {
		Winner string `json:"winner"`
	}{
		Winner: winner,
	})

	for _, player := range game.Players {
		if err := player.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Failed to send winner message to player %s: %v", player.ID, err)
		}
	}
}

func broadcastGameStatus() {
	for _, player := range game.Players {
		// Базовый статус, который отправляется всем
		status := struct {
			Phase                   string          `json:"phase"`
			Players                 map[string]bool `json:"players"`
			Day                     int             `json:"day"`
			TargetedScreamPlayer    string          `json:"targeted_scream_player,omitempty"`
			TargetedSunFlowerPlayer string          `json:"targeted_sun_flower_player,omitempty"`
			TimeRemaining           int             `json:"time_remaining"`
			Votes                   map[string]int  `json:"votes"`
		}{
			Phase: game.CurrentPhase,
			Players: func() map[string]bool {
				players := make(map[string]bool)
				for id, p := range game.Players {
					players[id] = p.IsAlive
				}
				return players
			}(),
			Day:           game.DayNumber,
			TimeRemaining: game.TimeRemaining,
			Votes:         game.Votes,
		}

		// Добавляем информацию о цели только для "Крикуна"
		if player.Role == "Крикун" && player.TargetedScreamerPlayer != "" {
			status.TargetedScreamPlayer = player.TargetedScreamerPlayer
		}
		if player.Role == "Дитя цветов" && player.TargetedSunFlowerPlayer != "" {
			status.TargetedSunFlowerPlayer = player.TargetedSunFlowerPlayer
		}

		// Сериализация в JSON
		data, err := json.Marshal(status)
		if err != nil {
			log.Printf("Failed to marshal game status for player %s: %v", player.ID, err)
			continue
		}

		// Отправка данных игроку
		err = player.Conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Printf("Failed to send game status to player %s: %v", player.ID, err)
		}
	}
}

func processMessage(playerID string, message []byte) {
	game.Mutex.Lock()
	//log.Println("Mutex Locked")
	game.Mutex.Unlock()
	//log.Println("Mutex UNLocked")

	var msg struct {
		Action  string `json:"action"`
		Target  string `json:"vote"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Failed to parse message: %s", err)
		return
	}

	player, exists := game.Players[playerID]
	if !exists || !player.IsAlive {
		return
	}

	if game.CurrentPhase == "day" && msg.Action == "vote" && !player.Hacked {
		if msg.Target == player.VotedFor {
			// Удаляем предыдущий голос
			game.Votes[player.VotedFor]--
			if game.Votes[player.VotedFor] < 0 {
				game.Votes[player.VotedFor] = 0
			}
			player.VotedFor = ""

		} else {
			if player.VotedFor != "" {
				game.Votes[player.VotedFor]--
				if game.Votes[player.VotedFor] < 0 {
					game.Votes[player.VotedFor] = 0
				}
			}
			// Ставим новый
			if _, ok := game.Players[msg.Target]; ok {
				player.VotedFor = msg.Target
				game.Votes[msg.Target]++
				log.Printf("Player %s voted for %s", playerID, msg.Target)
			}
		}
	} else if game.CurrentPhase == "night" && (player.Aura == "bad" || player.Role == "Провидец" || player.Role == "Провидец ауры" || player.Role == "Доктор" || player.Role == "Хакер") && msg.Action != "cancel_vote" {
		player.Action = msg.Target
		log.Printf("Player %s (%s) targets %s", playerID, player.Role, msg.Target)
	} else if msg.Action == "start_game" {
		log.Printf("Player %s requested to start the game", playerID)
		startGame(nil, nil) // Запуск игры
	} else if msg.Action == "chat" && !player.Hacked {
		broadcastChatMessage(playerID, msg.Message)
	} else if game.CurrentPhase == "day" && msg.Action == "cancel_vote" && !player.Hacked {
		game.Votes[msg.Target]--
	} else if game.CurrentPhase == "night" && msg.Action == "cancel_vote" {
		player.Action = ""
	} else if player.Role == "Крикун" && msg.Action == "scream_target" {
		game.Mutex.Lock()
		player.TargetedScreamerPlayer = msg.Target
		game.Mutex.Unlock()
		log.Printf("Screamer selected target: %s", msg.Target)
		broadcastGameStatus()
	} else if player.Role == "Дитя цветов" && msg.Action == "scream_target" {
		game.Mutex.Lock()
		player.TargetedSunFlowerPlayer = msg.Target
		game.Mutex.Unlock()
		log.Printf("FlowerChild selected target: %s", msg.Target)
		broadcastGameStatus()
	}

}

func broadcastChatMessage(playerID, chatMessage string) {
	message, _ := json.Marshal(struct {
		PlayerID string `json:"playerID"`
		Chat     string `json:"chat"`
	}{
		PlayerID: playerID,
		Chat:     chatMessage,
	})

	for _, player := range game.Players {
		if err := player.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Failed to send chat message to player %s: %v", player.ID, err)
		}
	}
}

func processVotes() {
	flowerTarget := ""
	// Подсчет количества живых игроков
	alivePlayers := 0
	for _, player := range game.Players {
		if player.IsAlive {
			alivePlayers++
		}
		if player.Role == "Дитя цветов" {
			if player.TargetedSunFlowerPlayer != "" {
				flowerTarget = game.Players[player.TargetedSunFlowerPlayer].ID
			}
		}
		if player.Hacked {
			player.IsAlive = false
			log.Printf("Player %s was killed by hacker", player.ID)
		}
	}

	// Порог голосов для исключения
	voteThreshold := calculateVoteThreshold(alivePlayers)

	// Подсчет голосов
	maxVotes := 0
	candidates := []string{}
	for playerID, votes := range game.Votes {
		if votes > maxVotes {
			maxVotes = votes
			candidates = []string{playerID}
		} else if votes == maxVotes {
			candidates = append(candidates, playerID)
		}
	}

	log.Printf("Vote threshold: %d, Max votes: %d, Candidates: %v", voteThreshold, maxVotes, candidates)

	// Проверка, есть ли кандидат с достаточным количеством голосов
	if maxVotes >= voteThreshold && len(candidates) == 1 {
		excludedPlayerID := candidates[0]
		if player, exists := game.Players[excludedPlayerID]; exists {
			flag := true
			if player.ID == flowerTarget {
				broadcastChatMessage("[SERVER]", fmt.Sprintf("Этого игрока нельзя казнить сегодня."))
				flag = false
			} else if player.Role == "Шут" {
				broadcastWinner("Шут победил!")
				game.GameStarted = false // Останавливаем игру
				return
			} else if player.Role == "Крикун" {
				log.Println("##Крикун1")
				if player.TargetedScreamerPlayer != "" {
					targetPlayer := game.Players[player.TargetedScreamerPlayer]
					if targetPlayer != nil {
						log.Println("##Крикун2")
						broadcastChatMessage("[SERVER]", fmt.Sprintf("Крикун раскрыл роль игрока %s - %s", targetPlayer.ID, targetPlayer.Role))
					}
				}
			}
			if flag {
				player.IsAlive = false
				log.Printf("Player %s was voted out.", excludedPlayerID)
			}
		}
	} else {
		log.Println("No player was excluded.")
	}

	// Очистка голосов
	game.Votes = make(map[string]int)

	// Обновление статуса игры для всех игроков
	broadcastGameStatus()
}

func calculateVoteThreshold(alivePlayers int) int {
	if alivePlayers%2 == 0 {
		return alivePlayers / 2
	}
	return alivePlayers/2 + 1
}

func gameStatus(w http.ResponseWriter, r *http.Request) {
	game.Mutex.Lock()
	//log.Println("Mutex Locked")
	//game.Mutex.Unlock()
	log.Println("Mutex UNLocked")

	status, _ := json.Marshal(struct {
		Phase   string          `json:"phase"`
		Players map[string]bool `json:"players"`
		Day     int             `json:"day"`
	}{
		Phase: game.CurrentPhase,
		Players: func() map[string]bool {
			players := make(map[string]bool)
			for id, player := range game.Players {
				players[id] = player.IsAlive
			}
			return players
		}(),
		Day: game.DayNumber,
	})

	w.Header().Set("Content-Type", "application/json")
	w.Write(status)
}

func startPhaseTimer(duration int, endPhaseFunc func()) {
	game.TimeRemaining = duration
	broadcastGameStatus() // Отправляем начальное значение таймера

	go func() {
		for game.TimeRemaining > 0 {
			time.Sleep(1 * time.Second)

			game.Mutex.Lock()
			game.TimeRemaining--
			game.Mutex.Unlock()

			broadcastGameStatus() // Обновляем таймер у всех клиентов
		}

		// Когда таймер истекает, вызываем переданную функцию (завершение фазы)
		endPhaseFunc()
	}()
}
