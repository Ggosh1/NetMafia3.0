package backend

import (
	"NetMafia3/backend/GameFiles"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

var roomManager = GameFiles.NewRoomManager()

func writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Ошибка кодирования JSON: %v", err)
	}
}

func JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("id")
	if playerID == "" {
		http.Error(w, "ID игрока не указан", http.StatusBadRequest)
		return
	}

	/*player, exists := game.Players[playerID]
	if !exists {
		player = &Player{
			ID:      playerID,
			IsAlive: true,
		}
		game.Players[playerID] = player
		log.Printf("Создан новый игрок %s через joinRoomHandler", playerID)
	}*/

	/*room := joinRoom(player)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"roomId":  room.ID,
		"players": len(room.Players),
	})*/
}

type CreateRoomRequest struct {
	RoomID string `json:"roomId"`
}

func JoinRoomByIDHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Попытка подключиться к комнате")

	if r.Method != http.MethodPost {
		writeJSONResponse(w, map[string]string{"error": "Метод не разрешён"}, http.StatusMethodNotAllowed)
		return
	}

	// Получаем playerID из query-параметров
	playerID := r.URL.Query().Get("id")
	if playerID == "" {
		writeJSONResponse(w, map[string]string{"error": "ID игрока не указан"}, http.StatusBadRequest)
		return
	}

	// Структура для парсинга тела запроса (roomId будем брать отсюда)
	type JoinRoomRequest struct {
		RoomID string `json:"roomId"`
	}

	var req JoinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, map[string]string{"error": "Некорректное тело запроса"}, http.StatusBadRequest)
		return
	}
	if req.RoomID == "" {
		writeJSONResponse(w, map[string]string{"error": "Не указан roomId"}, http.StatusBadRequest)
		return
	}

	log.Printf("Принят запрос на добавление игрока %s в комнату %s", playerID, req.RoomID)

	// Если у вас есть какой-то общий мьютекс для синхронизации, используйте его:
	// Добавляем игрока в комнату
	log.Printf("Пытаемся добавить игрока %s в комнату %s", playerID, req.RoomID)

	if err := roomManager.AddPlayerToRoom(req.RoomID, playerID); err != nil {
		log.Printf("Ошибка при добавлении игрока %s в комнату %s: %v", playerID, req.RoomID, err)
		writeJSONResponse(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	log.Printf("Вроде добавили игрока %s в комнату %s", playerID, req.RoomID)

	// Успешное добавление
	writeJSONResponse(w, map[string]interface{}{
		"status":  "ok",
		"message": fmt.Sprintf("Игрок %s добавлен в комнату %s", playerID, req.RoomID),
		"roomId":  req.RoomID,
	}, http.StatusOK)
}

func CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Попытка создать комнату")

	if r.Method != http.MethodPost {
		writeJSONResponse(w, map[string]string{"error": "Метод не разрешён"}, http.StatusMethodNotAllowed)
		return
	}

	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, map[string]string{"error": "Некорректное тело запроса"}, http.StatusBadRequest)
		return
	}

	if req.RoomID == "" {
		writeJSONResponse(w, map[string]string{"error": "ID комнаты не указан"}, http.StatusBadRequest)
		return
	}

	log.Printf("Принят запрос на создание комнаты %s", req.RoomID)

	room, err := roomManager.CreateRoom(req.RoomID)

	if err != nil {
		log.Printf("Ошибка при создании комнаты %s - %s", req.RoomID, err)
		writeJSONResponse(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	log.Printf("ID созданной комнаты %s", room.ID)

	writeJSONResponse(w, map[string]interface{}{
		"roomId": room.ID,
		// "players": len(room.Players), // Можно добавить дополнительные данные при необходимости
	}, http.StatusOK)
}

func LeaveRoomHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Обработка выхода из комнаты")

	playerID := r.URL.Query().Get("id")
	if playerID == "" {
		http.Error(w, "ID игрока не указан", http.StatusBadRequest)
		return
	}

	roomID := r.URL.Query().Get("roomId")
	if roomID == "" {
		http.Error(w, "ID комнаты не указан", http.StatusBadRequest)
		return
	}

	log.Printf("Попытка получения комнаты %s", roomID)

	room, err := roomManager.GetRoom(roomID)
	if err != nil {
		log.Printf("Ошибка при получении комнаты: %v", err)
		http.Error(w, "Комната не найдена", http.StatusNotFound)
		return
	}

	log.Printf("Удаление игрока %s из комнаты %s", playerID, roomID)
	room.RemovePlayer(playerID)

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

	_, err = Db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", req.Username, req.Password)
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	roomManager.CreatePlayer(req.Username)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Пользователь успешно зарегистрирован",
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

	roomManager.CreatePlayer(req.Username)

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Вход выполнен успешно",
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

func AvailableRoomsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var roomsList []struct {
		RoomID     string `json:"roomID"`
		PlayersCnt int    `json:"playersCnt"`
	}

	for _, room := range roomManager.GetRoomsList() {
		roomsList = append(roomsList, struct {
			RoomID     string `json:"roomID"`
			PlayersCnt int    `json:"playersCnt"`
		}{
			RoomID:     room.ID,
			PlayersCnt: len(room.Game.Players),
		})
	}

	log.Printf("Available Rooms: %v", roomsList)
	json.NewEncoder(w).Encode(roomsList)
}
func ServeGame(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./frontend/index.html")
}

func GameStatus(w http.ResponseWriter, r *http.Request) {
	/*game.Mutex.Lock()
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
	w.Write(status)*/
}
