package backend

import (
	"database/sql"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
)

func JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("id")
	if playerID == "" {
		http.Error(w, "ID игрока не указан", http.StatusBadRequest)
		return
	}

	game.Mutex.Lock()
	player, exists := game.Players[playerID]
	if !exists {
		player = &Player{
			ID:      playerID,
			IsAlive: true,
		}
		game.Players[playerID] = player
		log.Printf("Создан новый игрок %s через joinRoomHandler", playerID)
	}
	game.Mutex.Unlock()

	room := joinRoom(player)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"roomId":  room.ID,
		"players": len(room.Players),
	})
}

func LeaveRoomHandler(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("id")
	if playerID == "" {
		http.Error(w, "ID игрока не указан", http.StatusBadRequest)
		return
	}

	roomLock.Lock()
	for _, room := range rooms {
		if _, exists := room.Players[playerID]; exists {
			delete(room.Players, playerID)
			log.Printf("Игрок %s покинул комнату %s", playerID, room.ID)
			break
		}
	}
	roomLock.Unlock()

	game.Mutex.Lock()
	delete(game.Players, playerID)
	game.Mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
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
	err := Db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)", req.Username).Scan(&exists)
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Пользователь с таким именем уже существует", http.StatusBadRequest)
		return
	}

	// Хэширование пароля с использованием bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Ошибка шифрования пароля", http.StatusInternalServerError)
		return
	}

	_, err = Db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", req.Username, string(hashedPassword))
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":     "success",
		"message":    "Пользователь успешно зарегистрирован. Добро пожаловать в ",
		"addMessage": " Mafia Game",
	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
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

	var storedPassword string
	err := Db.QueryRow("SELECT password FROM users WHERE username=$1", req.Username).Scan(&storedPassword)
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

	// Сравниваем хэшированный пароль из базы с введённым паролем
	if err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(req.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный пароль"})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":     "success",
		"message":    "Вход выполнен успешно ",
		"addMessage": " Mafia game",
	})
}

func ServeProfile(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID не указан", http.StatusBadRequest)
		return
	}
	http.ServeFile(w, r, "./frontend/profile.html")
}

func ServeWelcome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./frontend/welcome.html")

}

func ServeGame(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./frontend/index.html")
}

func GameStatus(w http.ResponseWriter, r *http.Request) {
	game.Mutex.Lock()
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

// Получение списка друзей по username (передаётся через параметр id)
func GetFriendsHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("id")
	if username == "" {
		http.Error(w, "ID игрока не указан", http.StatusBadRequest)
		return
	}
	rows, err := Db.Query("SELECT friend_username FROM friends WHERE user_username = $1", username)
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	friends := []string{}
	for rows.Next() {
		var friend string
		rows.Scan(&friend)
		friends = append(friends, friend)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"friends": friends,
	})
}

// Добавление друга (ожидается POST с JSON: { "user": "username", "friend": "friend_username" })
func AddFriendHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("ADDFRIEND")
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	type AddFriendRequest struct {
		User   string `json:"user"`
		Friend string `json:"friend"`
	}
	var req AddFriendRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.User == "" || req.Friend == "" {
		http.Error(w, "Пользователь или друг не указан", http.StatusBadRequest)
		return
	}

	// Проверяем, что friend существует в таблице users
	var exists bool
	err := Db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)", req.Friend).Scan(&exists)
	if err != nil {
		http.Error(w, "Ошибка при проверке пользователя в БД", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Пользователь с логином "+req.Friend+" не существует", http.StatusBadRequest)
		return
	}

	var alreadyAdded bool
	err = Db.QueryRow("SELECT EXISTS(SELECT 1 FROM friends WHERE user_username=$1 AND friend_username=$2)", req.User, req.Friend).Scan(&alreadyAdded)
	if err != nil {
		http.Error(w, "Ошибка при проверке списка друзей", http.StatusInternalServerError)
		return
	}
	if alreadyAdded {
		http.Error(w, "Пользователь уже добавлен в друзья", http.StatusBadRequest)
		return
	}

	_, err = Db.Exec("INSERT INTO friends (user_username, friend_username) VALUES ($1, $2), ($2, $1) ON CONFLICT DO NOTHING", req.User, req.Friend)

	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Друг успешно добавлен",
	})
}

// Удаление друга (ожидается POST с JSON: { "user": "username", "friend": "friend_username" })
func RemoveFriendHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	type RemoveFriendRequest struct {
		User   string `json:"user"`
		Friend string `json:"friend"`
	}
	var req RemoveFriendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	if req.User == "" || req.Friend == "" {
		http.Error(w, "Пользователь или друг не указан", http.StatusBadRequest)
		return
	}

	var alreadyAdded bool
	err := Db.QueryRow("SELECT EXISTS(SELECT 1 FROM friends WHERE user_username=$1 AND friend_username=$2)", req.User, req.Friend).Scan(&alreadyAdded)
	if err != nil {
		http.Error(w, "Ошибка при проверке списка друзей", http.StatusInternalServerError)
		return
	}
	if !alreadyAdded {
		http.Error(w, "Пользователя нет в друзьях", http.StatusBadRequest)
		return
	}

	_, err = Db.Exec("DELETE FROM friends WHERE (user_username = $1 AND friend_username = $2) OR (user_username = $2 AND friend_username = $1)", req.User, req.Friend)
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Друг успешно удалён",
	})
}
