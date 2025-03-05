package backend

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
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

	sessionID, err1 := GenerateSessionID(16)
	if err1 != nil {
		http.Error(w, "Ошибка генерации Session ID", http.StatusInternalServerError)
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

	_, err = Db.Exec("INSERT INTO users (username, password, session_id) VALUES ($1, $2, $3)", req.Username, req.Password, sessionID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Ошибка базы записи sessionID", http.StatusInternalServerError)
		return
	}

	// Создание куки
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Domain:   "localhost", // Укажите свой домен (или удалите, чтобы куки работала только для текущего домена)
		MaxAge:   3600,
		Secure:   false, // В production используйте HTTPS
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	// Установка куки
	http.SetCookie(w, cookie)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":     "success",
		"message":    "Пользователь успешно зарегистрирован. Добро пожаловать в ",
		"addMessage": " Mafia Game",
		"token":      sessionID,
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

	if req.Password != storedPassword {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный пароль"})
		return
	}
	sessionID, err := GenerateSessionID(16)
	if err != nil {
		http.Error(w, "Ошибка генерации Session ID", http.StatusInternalServerError)
		return
	}

	// Обновляем session_id в базе данных
	_, err = Db.Exec("UPDATE users SET session_id=$1 WHERE username=$2", sessionID, req.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка обновления сессии в базе данных"})
		return
	}

	// Создание куки
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Domain:   "localhost", // Укажите свой домен (или удалите, чтобы куки работала только для текущего домена)
		MaxAge:   3600,
		Secure:   false, // В production используйте HTTPS
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	// Установка куки
	http.SetCookie(w, cookie)

	json.NewEncoder(w).Encode(map[string]string{
		"status":     "success",
		"message":    "Вход выполнен успешно ",
		"addMessage": " Mafia game",
		"token":      sessionID,
	})
}

func GenerateSessionID(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func ServeProfile(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("session_id")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		} else {
			fmt.Println("Ошибка при получении куки:", err)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID не указан", http.StatusBadRequest)
		return
	}
	http.ServeFile(w, r, "./frontend/profile.html")
}

func ServeWelcome(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")

	if err != nil {
		if err == http.ErrNoCookie {
			fmt.Println("куки не найден")
			http.ServeFile(w, r, "./frontend/welcome.html")
			return
		} else {
			fmt.Println("Ошибка при получении куки:", err)
			http.ServeFile(w, r, "./frontend/welcome.html")
			return
		}
	}

	fmt.Println("Куки session_id:", cookie)

	redirectURL := "/profile"

	u, err := url.Parse(redirectURL)
	if err != nil {
		http.Error(w, "Ошибка формирования URL", http.StatusInternalServerError)
		fmt.Println("Ошибка парсинга URL:", err)
		return
	}
	query := u.Query()
	query.Add("id", cookie.Value)
	u.RawQuery = query.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)
	fmt.Println("редирект")
}

func ServeGame(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("session_id")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		} else {
			fmt.Println("Ошибка при получении куки:", err)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	}
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
