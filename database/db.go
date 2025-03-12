package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite" // SQLite driver
)

type DB struct {
	Conn *sql.DB
}

type User struct {
	ID             int64
	CalculateCount int64
}

// ConnectDB подключается к базе данных и создает таблицу, если она не существует
func ConnectDB() (*DB, error) {
	conn, err := sql.Open("sqlite", "wannacar-bot.db")
	if err != nil {
		return nil, err
	}

	if err = conn.Ping(); err != nil {
		return nil, err
	}

	// Создаем таблицу, если она не существует
	err = createUsersTable(conn)
	if err != nil {
		return nil, err
	}

	return &DB{Conn: conn}, nil
}

// createUsersTable создает таблицу пользователей, если она не существует
func createUsersTable(conn *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		calculate_count INTEGER DEFAULT NULL
	);
	`
	_, err := conn.Exec(query)
	if err != nil {
		log.Printf("Ошибка при создании таблицы: %v", err)
		return err
	}
	log.Println("Таблица пользователей успешно создана/обновлена")
	return nil
}

func (db *DB) Close() {
	db.Conn.Close()
}

func (db *DB) CreateUser(userID int64) error {
	query := "INSERT INTO users (id, calculate_count) VALUES (?, 1)"
	_, err := db.Conn.Exec(query, userID)
	return err
}

func (db *DB) GetUserByID(userID int64) *User {
	var user User
	query := "SELECT id, calculate_count FROM users WHERE id = ?"
	err := db.Conn.QueryRow(query, userID).Scan(&user.ID, &user.CalculateCount)
	if err != nil {
		return nil
	}
	return &user
}

func (db *DB) UpdateCalculateCount(userID int64) error {
	query := "UPDATE users SET calculate_count = calculate_count + 1 WHERE id = ?"
	_, err := db.Conn.Exec(query, userID)
	return err
}
