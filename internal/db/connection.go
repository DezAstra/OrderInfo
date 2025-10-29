package db

import (
	"OrderInfo/config"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// NewConnection создает новое подключение к базе данных PostgreSQL.
func NewConnection(cfg *config.DBConfig) (*sql.DB, error) {
	// строка подключения (DSN)
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.Host,
		cfg.Port,
		cfg.SSLMode)

	log.Printf("Подключение к БД: %s", dsn) // Логирование строки подключения

	// Открываем подключение
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("Неудача при открытии соединения БД: %w", err)
	}

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		// Если ping не прошёл, закрываем соединение
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Ошибка при закрытии соединения с БД (fail ping): %v", closeErr)
		}
		return nil, fmt.Errorf("Ошибка при проверке соединения с БД: %w", err)
	}

	log.Println("Успешное подключение к БД")
	return db, nil
}
