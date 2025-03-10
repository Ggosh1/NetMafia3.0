package main

import ( 
	"log"
	"net/http"
	_ "net/http/pprof"

	"NetMafia3/backend"
	_ "github.com/lib/pq"
)

func main() {
	// Инициализация базы данных
	backend.InitDB()
	defer backend.Db.Close()

	// Настройка HTTP-обработчиков
	http.HandleFunc("/login", backend.LoginHandler)
	http.HandleFunc("/register", backend.RegisterHandler)
	http.HandleFunc("/profile", backend.ServeProfile) // Страница профиля
	http.HandleFunc("/", backend.ServeWelcome)        // Приветственная страница
	http.HandleFunc("/game", backend.ServeGame)       // Игровая страница
	http.HandleFunc("/ws", backend.HandleConnections) // WebSocket соединения
	http.HandleFunc("/status", backend.GameStatus)
	http.HandleFunc("/joinroom", backend.JoinRoomHandler)
	http.HandleFunc("/leaveroom", backend.LeaveRoomHandler)
	http.HandleFunc("/friends", backend.GetFriendsHandler)
	http.HandleFunc("/addfriend", backend.AddFriendHandler)
	http.HandleFunc("/removefriend", backend.RemoveFriendHandler)
	// Раздача статических файлов (css, js, картинки)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./frontend/static"))))

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
