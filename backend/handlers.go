package backend

import (
	"NetMafia3/backend/GameFiles"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

var roomManager = GameFiles.NewRoomManager()

func writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Ошибка кодирования JSON: %v", err)
	}
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
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}
	sessionID := cookie.Value
	fmt.Println(sessionID)
	login, err := getLoginBySession(sessionID)
	if err != nil {
		fmt.Println(err)
	}

	if login == "" {
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
	if req.RoomID == "" && len(roomManager.Rooms) == 0 {
		writeJSONResponse(w, map[string]string{"error": "Не указан roomId"}, http.StatusBadRequest)
		return
	}

	if req.RoomID == "" {
		req.RoomID = roomManager.GetBestRoom().ID
	}

	log.Printf("Принят запрос на добавление игрока %s в комнату %s", login, req.RoomID)

	// Если у вас есть какой-то общий мьютекс для синхронизации, используйте его:
	// Добавляем игрока в комнату
	log.Printf("Пытаемся добавить игрока %s в комнату %s", login, req.RoomID)

	if err := roomManager.AddPlayerToRoom(req.RoomID, login); err != nil {
		log.Printf("Ошибка при добавлении игрока %s в комнату %s: %v", login, req.RoomID, err)
		writeJSONResponse(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	log.Printf("Вроде добавили игрока %s в комнату %s", login, req.RoomID)

	// Успешное добавление
	writeJSONResponse(w, map[string]interface{}{
		"status":  "ok",
		"message": fmt.Sprintf("Игрок %s добавлен в комнату %s", login, req.RoomID),
		"roomId":  req.RoomID,
		"login":   login,
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
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}
	sessionID := cookie.Value
	login, err := getLoginBySession(sessionID)
	if err != nil {
		fmt.Println(err)
	}
	roomManager.AddPlayerToRoom(room.ID, login)
	writeJSONResponse(w, map[string]interface{}{
		"roomId": room.ID,
		"login":  login,
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

	_, err := roomManager.GetRoom(roomID)
	if err != nil {
		log.Printf("Ошибка при получении комнаты: %v", err)
		http.Error(w, "Комната не найдена", http.StatusNotFound)
		return
	}

	log.Printf("Удаление игрока %s из комнаты %s", playerID, roomID)
	roomManager.RemovePlayerFromRoom(roomID, playerID)

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

	//создаем sessionID
	sessionID, err1 := GenerateSessionID(16)
	if err1 != nil {
		http.Error(w, "Ошибка генерации Session ID", http.StatusInternalServerError)
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
		MaxAge:   3600,
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	// Установка куки
	http.SetCookie(w, cookie)
	roomManager.CreatePlayer(req.Username)
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

	//создаем sessionID
	sessionID, err := GenerateSessionID(16)
	if err != nil {
		http.Error(w, "Ошибка генерации Session ID", http.StatusInternalServerError)
		return
	}
	//обова sessionID в бд
	_, err = Db.Exec("UPDATE users SET session_id=$1 WHERE username=$2", sessionID, req.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка обновления сессии в базе данных"})
		return
	}
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   3600,
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
	roomManager.CreatePlayer(req.Username)
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
	id, err := getLoginBySession(cookie.Value)
	roomManager.CreatePlayer(id)
	http.Redirect(w, r, "/profile", http.StatusFound)
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
func RoomsGame(w http.ResponseWriter, r *http.Request) {
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
	http.ServeFile(w, r, "./frontend/rooms.html")
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

func GetLogin(w http.ResponseWriter, r *http.Request) {
	type ResponseData struct {
		Login string `json:"login"`
		Error string `json:"error,omitempty"`
	}
	// Получаем cookie с session_id
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	sessionID := cookie.Value

	// Получаем login по sessionID
	login, err := getLoginBySession(sessionID)
	resp := ResponseData{}

	if err != nil {
		resp.Error = "Сессия не найдена"
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		resp.Login = login
	}

	// Устанавливаем заголовок и отправляем JSON-ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getLoginBySession(sessionID string) (string, error) {
	var login string
	query := "SELECT username FROM users WHERE session_id = $1;"
	err := Db.QueryRow(query, sessionID).Scan(&login)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("сессия не найдена")
		}
		return "", err
	}
	fmt.Println(login)
	return login, nil
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
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	sessionID := cookie.Value

	// Получаем login по sessionID
	login, _ := getLoginBySession(sessionID)
	response := struct {
		Login string      `json:"login"`
		Rooms interface{} `json:"rooms"`
	}{
		Login: login,
		Rooms: roomsList,
	}

	log.Printf("Доступные комнаты: %v, Login: %s", roomsList, login)
	json.NewEncoder(w).Encode(response)
}
func GetFriendsHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	sessionID := cookie.Value

	// Получаем login по sessionID
	username, _ := getLoginBySession(sessionID)
	fmt.Println("пролучил куки")
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
	fmt.Println(friends)
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

	// Получаем session_id из cookies
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	sessionID := cookie.Value

	// Получаем логин пользователя по sessionID
	username, err := getLoginBySession(sessionID)
	if err != nil {
		http.Error(w, "Ошибка аутентификации", http.StatusUnauthorized)
		return
	}

	// Структура для парсинга JSON-запроса
	type AddFriendRequest struct {
		Friend string `json:"friend"`
	}
	var req AddFriendRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Friend == "" {
		http.Error(w, "Друг не указан", http.StatusBadRequest)
		return
	}

	// Проверяем, что friend существует в таблице users
	var exists bool
	err = Db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)", req.Friend).Scan(&exists)
	if err != nil {
		http.Error(w, "Ошибка при проверке пользователя в БД", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Пользователь с логином "+req.Friend+" не существует", http.StatusBadRequest)
		return
	}

	// Проверяем, не является ли friend уже другом
	var alreadyAdded bool
	err = Db.QueryRow("SELECT EXISTS(SELECT 1 FROM friends WHERE user_username=$1 AND friend_username=$2)", username, req.Friend).Scan(&alreadyAdded)
	if err != nil {
		http.Error(w, "Ошибка при проверке списка друзей", http.StatusInternalServerError)
		return
	}
	if alreadyAdded {
		http.Error(w, "Пользователь уже добавлен в друзья", http.StatusBadRequest)
		return
	}

	// Добавляем в друзья (в обе стороны)
	_, err = Db.Exec("INSERT INTO friends (user_username, friend_username) VALUES ($1, $2), ($2, $1) ON CONFLICT DO NOTHING", username, req.Friend)

	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
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

	// Извлекаем session_id из куки
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	sessionID := cookie.Value

	// Получаем логин пользователя по sessionID
	username, err := getLoginBySession(sessionID)
	if err != nil {
		http.Error(w, "Ошибка аутентификации", http.StatusUnauthorized)
		return
	}

	// Структура для запроса на удаление друга
	type RemoveFriendRequest struct {
		Friend string `json:"friend"`
	}
	var req RemoveFriendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	if req.Friend == "" {
		http.Error(w, "Друг не указан", http.StatusBadRequest)
		return
	}

	// Проверка, есть ли пользователь в списке друзей
	var alreadyAdded bool
	err = Db.QueryRow("SELECT EXISTS(SELECT 1 FROM friends WHERE user_username=$1 AND friend_username=$2)", username, req.Friend).Scan(&alreadyAdded)
	if err != nil {
		http.Error(w, "Ошибка при проверке списка друзей", http.StatusInternalServerError)
		return
	}
	if !alreadyAdded {
		http.Error(w, "Пользователя нет в друзьях", http.StatusBadRequest)
		return
	}

	// Удаляем друга из базы данных
	_, err = Db.Exec("DELETE FROM friends WHERE (user_username = $1 AND friend_username = $2) OR (user_username = $2 AND friend_username = $1)", username, req.Friend)
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Друг успешно удалён",
	})
}

func FriendsList(w http.ResponseWriter, r *http.Request) {
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
	http.ServeFile(w, r, "./frontend/addFriend.html")
}
