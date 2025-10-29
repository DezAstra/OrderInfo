package config

import (
	"log"
	"os"
)

// Config хранит все настройки приложения.
type Config struct {
	DB     DBConfig
	NATS   NATSConfig
	Server ServerConfig
}

// DBConfig хранит настройки подключения к базе данных.
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string // часто "disable", "require", "verify-full" и т.д.
}

// NATSConfig хранит настройки подключения к NATS Streaming.
type NATSConfig struct {
	URL       string
	ClusterID string // ID кластера NATS Streaming
	ClientID  string // ID клиента для подписки
	Channel   string // Имя подписываемого канала
}

// ServerConfig хранит настройки HTTP-сервера.
type ServerConfig struct {
	HTTPPort string
}

// LoadConfig загружает конфигурацию
func LoadConfig() (*Config, error) {
	dbHost := getEnvOrDefault("DB_HOST", "localhost")
	dbPort := getEnvOrDefault("DB_PORT", "5432")
	dbUser := getEnvOrDefault("DB_USER", "order_user")    // локальный пользователь БД
	dbPassword := getEnvOrDefault("DB_PASSWORD", "admin") // локальный пароль
	dbName := getEnvOrDefault("DB_NAME", "order_db")      // локальная БД
	dbSSLMode := getEnvOrDefault("DB_SSLMODE", "disable")

	natsURL := getEnvOrDefault("NATS_URL", "nats://localhost:4223")
	natsClusterID := getEnvOrDefault("NATS_CLUSTER_ID", "test-cluster")
	natsClientID := getEnvOrDefault("NATS_CLIENT_ID", "order_service_client_1") // ID клиента
	natsChannel := getEnvOrDefault("NATS_CHANNEL", "orders")                    // Имя канала

	httpPort := getEnvOrDefault("HTTP_PORT", ":8080")

	config := &Config{
		DB: DBConfig{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPassword,
			Name:     dbName,
			SSLMode:  dbSSLMode,
		},
		NATS: NATSConfig{
			URL:       natsURL,
			ClusterID: natsClusterID,
			ClientID:  natsClientID,
			Channel:   natsChannel,
		},
		Server: ServerConfig{
			HTTPPort: httpPort,
		},
	}

	// конфигурация для отладки
	log.Printf("Configuration loaded: %+v", config)

	return config, nil
}

// getEnvOrDefault возвращает значение переменной окружения или значение по умолчанию
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		log.Printf("Использование переменной окружения %s=%s", key, value)
		return value
	}
	log.Printf("Переменная окружения %s не установлена. используется по умолчанию: %s", key, defaultValue)
	return defaultValue
}
