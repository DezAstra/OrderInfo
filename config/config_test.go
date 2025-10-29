package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadConfig_Defaults тестирует загрузку конфигурации с значениями по умолчанию.
func TestLoadConfig_Defaults(t *testing.T) {
	// 1. Подготовка: проверка, что переменные окружения НЕ установлены
	// Сохранение текущих значений
	originalEnvVars := map[string]*string{
		"DB_HOST":         getEnvPtr("DB_HOST"),
		"DB_PORT":         getEnvPtr("DB_PORT"),
		"DB_USER":         getEnvPtr("DB_USER"),
		"DB_PASSWORD":     getEnvPtr("DB_PASSWORD"),
		"DB_NAME":         getEnvPtr("DB_NAME"),
		"DB_SSLMODE":      getEnvPtr("DB_SSLMODE"),
		"NATS_URL":        getEnvPtr("NATS_URL"),
		"NATS_CLUSTER_ID": getEnvPtr("NATS_CLUSTER_ID"),
		"NATS_CLIENT_ID":  getEnvPtr("NATS_CLIENT_ID"),
		"NATS_CHANNEL":    getEnvPtr("NATS_CHANNEL"),
		"HTTP_PORT":       getEnvPtr("HTTP_PORT"),
	}

	// удаление переменных окружения перед тестом
	for key := range originalEnvVars {
		os.Unsetenv(key)
	}

	// восстановление переменных окружения после теста
	defer func() {
		for key, valuePtr := range originalEnvVars {
			if valuePtr != nil {
				os.Setenv(key, *valuePtr)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// 2. Тестирование: вызов тестируемой функцию
	cfg, err := LoadConfig()

	// 3. Проверка что конфигурация загружена без ошибок
	require.NoError(t, err, "LoadConfig should not return an error with default values")
	assert.NotNil(t, cfg, "Config should not be nil")

	// 4. Проверка, что используются значения по умолчанию
	// DB
	assert.Equal(t, "localhost", cfg.DB.Host, "DB Host should be default")
	assert.Equal(t, "5432", cfg.DB.Port, "DB Port should be default")
	assert.Equal(t, "order_user", cfg.DB.User, "DB User should be default")
	assert.Equal(t, "admin", cfg.DB.Password, "DB Password should be default")
	assert.Equal(t, "order_db", cfg.DB.Name, "DB Name should be default")
	assert.Equal(t, "disable", cfg.DB.SSLMode, "DB SSLMode should be default")

	// NATS
	assert.Equal(t, "nats://localhost:4223", cfg.NATS.URL, "NATS URL should be default")
	assert.Equal(t, "test-cluster", cfg.NATS.ClusterID, "NATS ClusterID should be default")
	assert.Equal(t, "order_service_client_1", cfg.NATS.ClientID, "NATS ClientID should be default")
	assert.Equal(t, "orders", cfg.NATS.Channel, "NATS Channel should be default")

	// Server
	assert.Equal(t, ":8080", cfg.Server.HTTPPort, "HTTP Port should be default")
}

// TestLoadConfig_FromEnv тестирует загрузку конфигурации из переменных окружения.
func TestLoadConfig_FromEnv(t *testing.T) {
	// 1. Подготовка: переменные окружения
	testDBHost := "test-db-host"
	testDBPort := "12345"
	testDBUser := "test-db-user"
	testDBPassword := "test-db-password"
	testDBName := "test-db-name"
	testDBSSLMode := "require"

	testNATSURL := "nats://test-nats:4223"
	testNATSClusterID := "test-nats-cluster"
	testNATSClientID := "test-nats-client"
	testNATSChannel := "test-orders"

	testHTTPPort := ":9090"

	// Сохранение текущих значений, если они есть, чтобы восстановить их позже
	originalEnvVars := map[string]*string{
		"DB_HOST":         getEnvPtr("DB_HOST"),
		"DB_PORT":         getEnvPtr("DB_PORT"),
		"DB_USER":         getEnvPtr("DB_USER"),
		"DB_PASSWORD":     getEnvPtr("DB_PASSWORD"),
		"DB_NAME":         getEnvPtr("DB_NAME"),
		"DB_SSLMODE":      getEnvPtr("DB_SSLMODE"),
		"NATS_URL":        getEnvPtr("NATS_URL"),
		"NATS_CLUSTER_ID": getEnvPtr("NATS_CLUSTER_ID"),
		"NATS_CLIENT_ID":  getEnvPtr("NATS_CLIENT_ID"),
		"NATS_CHANNEL":    getEnvPtr("NATS_CHANNEL"),
		"HTTP_PORT":       getEnvPtr("HTTP_PORT"),
	}

	// Устанавливаем тестовые переменные окружения
	os.Setenv("DB_HOST", testDBHost)
	os.Setenv("DB_PORT", testDBPort)
	os.Setenv("DB_USER", testDBUser)
	os.Setenv("DB_PASSWORD", testDBPassword)
	os.Setenv("DB_NAME", testDBName)
	os.Setenv("DB_SSLMODE", testDBSSLMode)

	os.Setenv("NATS_URL", testNATSURL)
	os.Setenv("NATS_CLUSTER_ID", testNATSClusterID)
	os.Setenv("NATS_CLIENT_ID", testNATSClientID)
	os.Setenv("NATS_CHANNEL", testNATSChannel)

	os.Setenv("HTTP_PORT", testHTTPPort)

	// Восстанавливаем переменные окружения после теста
	defer func() {
		for key, valuePtr := range originalEnvVars {
			if valuePtr != nil {
				os.Setenv(key, *valuePtr)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// 2. Тестирование: вызов тестируемой функции
	cfg, err := LoadConfig()

	// 3. Проверка что конфигурация загружена без ошибок
	require.NoError(t, err, "LoadConfig should not return an error with env values")
	assert.NotNil(t, cfg, "Config should not be nil")

	// 4. Проверка что используются значения из переменных окружения
	// DB
	assert.Equal(t, testDBHost, cfg.DB.Host, "DB Host should be from env")
	assert.Equal(t, testDBPort, cfg.DB.Port, "DB Port should be from env")
	assert.Equal(t, testDBUser, cfg.DB.User, "DB User should be from env")
	assert.Equal(t, testDBPassword, cfg.DB.Password, "DB Password should be from env")
	assert.Equal(t, testDBName, cfg.DB.Name, "DB Name should be from env")
	assert.Equal(t, testDBSSLMode, cfg.DB.SSLMode, "DB SSLMode should be from env")

	// NATS
	assert.Equal(t, testNATSURL, cfg.NATS.URL, "NATS URL should be from env")
	assert.Equal(t, testNATSClusterID, cfg.NATS.ClusterID, "NATS ClusterID should be from env")
	assert.Equal(t, testNATSClientID, cfg.NATS.ClientID, "NATS ClientID should be from env")
	assert.Equal(t, testNATSChannel, cfg.NATS.Channel, "NATS Channel should be from env")

	// Server
	assert.Equal(t, testHTTPPort, cfg.Server.HTTPPort, "HTTP Port should be from env")
}

// TestLoadConfig_PartialEnv тестирует загрузку конфигурации с частично заданными переменными окружения.
func TestLoadConfig_PartialEnv(t *testing.T) {
	// 1. Подготовка: Устанавливаем НЕКОТОРЫЕ переменные окружения, другие оставляем неустановленными
	testDBHost := "partial-test-db-host"
	testNATSURL := "nats://partial-test-nats:4223"
	testHTTPPort := ":7070"

	// Сохраняем текущие значения, если они есть, чтобы восстановить их позже
	originalEnvVars := map[string]*string{
		"DB_HOST":         getEnvPtr("DB_HOST"),
		"DB_PORT":         getEnvPtr("DB_PORT"),
		"DB_USER":         getEnvPtr("DB_USER"),
		"DB_PASSWORD":     getEnvPtr("DB_PASSWORD"),
		"DB_NAME":         getEnvPtr("DB_NAME"),
		"DB_SSLMODE":      getEnvPtr("DB_SSLMODE"),
		"NATS_URL":        getEnvPtr("NATS_URL"),
		"NATS_CLUSTER_ID": getEnvPtr("NATS_CLUSTER_ID"),
		"NATS_CLIENT_ID":  getEnvPtr("NATS_CLIENT_ID"),
		"NATS_CHANNEL":    getEnvPtr("NATS_CHANNEL"),
		"HTTP_PORT":       getEnvPtr("HTTP_PORT"),
	}

	// Убираем все переменные окружения
	for key := range originalEnvVars {
		os.Unsetenv(key)
	}

	// Устанавливаем только некоторые тестовые переменные окружения
	os.Setenv("DB_HOST", testDBHost)
	os.Setenv("NATS_URL", testNATSURL)
	os.Setenv("HTTP_PORT", testHTTPPort)

	// Восстанавливаем переменные окружения после теста
	defer func() {
		for key, valuePtr := range originalEnvVars {
			if valuePtr != nil {
				os.Setenv(key, *valuePtr)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// 2. Тестирование: Вызываем тестируемую функцию
	cfg, err := LoadConfig()

	// 3. Проверка, что конфигурация загружена без ошибок
	require.NoError(t, err, "LoadConfig should not return an error with partial env values")
	assert.NotNil(t, cfg, "Config should not be nil")

	// 4. Проверка, что используются значения из переменных окружения ТАМ, ГДЕ ОНИ ЗАДАНЫ
	// DB
	assert.Equal(t, testDBHost, cfg.DB.Host, "DB Host should be from env")
	// Остальные DB параметры должны быть по умолчанию
	assert.Equal(t, "5432", cfg.DB.Port, "DB Port should be default")
	assert.Equal(t, "order_user", cfg.DB.User, "DB User should be default")
	assert.Equal(t, "admin", cfg.DB.Password, "DB Password should be default")
	assert.Equal(t, "order_db", cfg.DB.Name, "DB Name should be default")
	assert.Equal(t, "disable", cfg.DB.SSLMode, "DB SSLMode should be default")

	// NATS
	assert.Equal(t, testNATSURL, cfg.NATS.URL, "NATS URL should be from env")
	// Остальные NATS параметры должны быть по умолчанию
	assert.Equal(t, "test-cluster", cfg.NATS.ClusterID, "NATS ClusterID should be default")
	assert.Equal(t, "order_service_client_1", cfg.NATS.ClientID, "NATS ClientID should be default")
	assert.Equal(t, "orders", cfg.NATS.Channel, "NATS Channel should be default")

	// Server
	assert.Equal(t, testHTTPPort, cfg.Server.HTTPPort, "HTTP Port should be from env")
}

// getEnvPtr возвращает указатель на значение переменной окружения или nil, если она не установлена.
func getEnvPtr(key string) *string {
	if value, exists := os.LookupEnv(key); exists {
		return &value // Возвращаем указатель на значение
	}
	return nil // Возвращаем nil, если переменная не установлена
}
