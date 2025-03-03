package backend

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var Db *sql.DB

func InitDB() {
	sysConnStr := "host=localhost port=5432 user=postgres password=123 dbname=postgres sslmode=disable"
	sysDB, err := sql.Open("postgres", sysConnStr)
	if err != nil {
		log.Fatal("Ошибка подключения к системной БД:", err)
	}
	defer sysDB.Close()

	if err = sysDB.Ping(); err != nil {
		log.Fatal("Ошибка подключения к системной БД:", err)
	}

	targetDBName := "mafia_game"

	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)"
	err = sysDB.QueryRow(checkQuery, targetDBName).Scan(&exists)
	if err != nil {
		log.Fatal("Ошибка проверки существования базы данных:", err)
	}

	if !exists {
		log.Printf("База данных %s не существует. Создаем...", targetDBName)
		_, err = sysDB.Exec(fmt.Sprintf("CREATE DATABASE %s", targetDBName))
		if err != nil {
			log.Fatal("Ошибка создания базы данных:", err)
		}
		log.Printf("База данных %s успешно создана.", targetDBName)
	} else {
		log.Printf("База данных %s уже существует.", targetDBName)
	}

	targetConnStr := fmt.Sprintf("host=localhost port=5432 user=postgres password=123 dbname=%s sslmode=disable", targetDBName)
	Db, err = sql.Open("postgres", targetConnStr)
	if err != nil {
		log.Fatal("Ошибка подключения к целевой базе данных:", err)
	}
	if err = Db.Ping(); err != nil {
		log.Fatal("Ошибка подключения к целевой базе данных:", err)
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
	);
	`
	if _, err := Db.Exec(createTableQuery); err != nil {
		log.Fatal("Ошибка создания таблицы:", err)
	}
}
