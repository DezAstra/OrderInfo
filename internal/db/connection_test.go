package db

import (
	"OrderInfo/config"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// getEnvOrDefault возвращает значение переменной окружения или значение по умолчанию
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// TestNewConnection_ValidConfig тестирует подключение с правильной конфигурацией.
func TestNewConnection_ValidConfig(t *testing.T) {
	// 1. Подготовка: Используем переменные окружения или значения по умолчанию
	cfg := config.DBConfig{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvOrDefault("DB_PORT", "5432"),
		User:     getEnvOrDefault("DB_USER", "order_user"), // Правильный пользователь
		Password: getEnvOrDefault("DB_PASSWORD", "admin"),  // Правильный пароль
		Name:     getEnvOrDefault("DB_NAME", "order_db"),   // Правильная БД
		SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
	}

	// 2. Тестирование: Вызываем тестируемую функцию
	dbConn, err := NewConnection(&cfg)

	// 3. Проверка: Убедимся, что подключение успешно
	require.NoError(t, err, "NewConnection should not return an error for valid config")
	assert.NotNil(t, dbConn, "DB connection should not be nil for valid config")

	// 4. Очистка: Закрываем подключение
	if dbConn != nil {
		err = dbConn.Close()
		assert.NoError(t, err, "Closing DB connection should not return an error")
	}
}

// TestNewConnection_InvalidConfig тестирует подключение с неправильной конфигурацией.
func TestNewConnection_InvalidConfig(t *testing.T) {
	// 1. Подготовка: Используем правильные параметры, но с НЕПРАВИЛЬНЫМ паролем
	cfg := config.DBConfig{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvOrDefault("DB_PORT", "5432"),
		User:     getEnvOrDefault("DB_USER", "order_user"), // Правильный пользователь
		Password: "wrong_password",                         // неправильный пароль
		Name:     getEnvOrDefault("DB_NAME", "order_db"),   // Правильная БД
		SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
	}

	// 2. Тестирование: Вызываем тестируемую функцию
	dbConn, err := NewConnection(&cfg)

	// 3. Проверка: Убедимся, что подключение НЕ успешно и возвращена ошибка
	assert.Error(t, err, "NewConnection should return an error for invalid password")
	assert.Nil(t, dbConn, "DB connection should be nil for invalid config")
}

// TestNewConnection_InvalidDBName тестирует подключение с неправильным именем БД.
func TestNewConnection_InvalidDBName(t *testing.T) {
	// 1. Подготовка: Используем правильные параметры, но с НЕПРАВИЛЬНЫМ именем БД
	cfg := config.DBConfig{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvOrDefault("DB_PORT", "5432"),
		User:     getEnvOrDefault("DB_USER", "order_user"), // Правильный пользователь
		Password: getEnvOrDefault("DB_PASSWORD", "admin"),  // Правильный пароль
		Name:     "non_existent_db",                        // НЕПРАВИЛЬНАЯ БД
		SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
	}

	// 2. Тестирование: Вызываем тестируемую функцию
	dbConn, err := NewConnection(&cfg)

	// 3. Проверка: Убедимся, что подключение НЕ успешно и возвращена ошибка
	assert.Error(t, err, "NewConnection should return an error for invalid DB name")
	assert.Nil(t, dbConn, "DB connection should be nil for invalid DB name")
}
