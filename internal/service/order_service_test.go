// internal/service/order_service_test.go
package service

import (
	"OrderInfo/config"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	testDB *sql.DB
)

const (
	testSchemaName = "test_service_schema_final_v2"
)

// TestMain - точка входа для тестов пакета. Используется для настройки и очистки тестовой среды.
func TestMain(m *testing.M) {
	log.Println("TestMain for OrderService tests: Setting up test environment...")
	// Подключаемся к основной БД, используя изолированную схему
	dsn := "user=" + getEnvOrDefault("DB_USER", "order_user") +
		" password=" + getEnvOrDefault("DB_PASSWORD", "admin") +
		" dbname=" + getEnvOrDefault("DB_NAME", "order_db") +
		" host=" + getEnvOrDefault("DB_HOST", "localhost") +
		" port=" + getEnvOrDefault("DB_PORT", "5432") +
		" sslmode=" + getEnvOrDefault("DB_SSLMODE", "disable")

	var err error
	testDB, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal("TestMain: Failed to connect to test DB: ", err)
	}
	defer func() {
		if testDB != nil {
			testDB.Close()
		}
	}()

	// Настройка изолированной тестовой среды
	// 1. Удаляем тестовую схему, если она существует (на случай сбоя предыдущего запуска)
	log.Println("TestMain: Dropping test schema if exists...")
	_, err = testDB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE;", testSchemaName))
	if err != nil {
		log.Fatalf("TestMain: Failed to drop existing test schema %s: %v", testSchemaName, err)
	}

	// 2. Создаём новую тестовую схему
	log.Println("TestMain: Creating new test schema...")
	_, err = testDB.Exec(fmt.Sprintf("CREATE SCHEMA %s;", testSchemaName)) // Убираем IF NOT EXISTS для чистоты эксперимента
	if err != nil {
		log.Fatalf("TestMain: Failed to create test schema %s: %v", testSchemaName, err)
	}

	// 3. Переключаемся на тестовую схему (для этой сессии)
	log.Println("TestMain: Setting search_path to test schema...")
	_, err = testDB.Exec(fmt.Sprintf("SET search_path TO %s;", testSchemaName))
	if err != nil {
		log.Fatalf("TestMain: Failed to set search_path to test schema: %v", err)
	}

	// 4. Создаём таблицы в тестовой схеме (копия из setup_db.sql)
	// ПРИМЕЧАНИЕ: Таблицы будут созданы внутри test_repo_schema
	// ВАЖНО: Порядок создания важен из-за внешних ключей.
	log.Println("TestMain: Creating tables...")
	createTablesSQL := `
        -- Таблица для информации о доставке
        CREATE TABLE delivery (
            uid VARCHAR(255) PRIMARY KEY,
            name VARCHAR(255),
            phone VARCHAR(50),
            zip VARCHAR(50),
            city VARCHAR(255),
            address TEXT,
            region VARCHAR(255),
            email VARCHAR(255)
        );

        -- Таблица для информации о платеже
        CREATE TABLE payment (
            transaction VARCHAR(255) PRIMARY KEY,
            request_id VARCHAR(255),
            currency VARCHAR(10),
            provider VARCHAR(100),
            amount INTEGER,
            payment_dt BIGINT, -- Unix timestamp
            bank VARCHAR(100),
            delivery_cost INTEGER,
            goods_total INTEGER,
            custom_fee INTEGER
        );

        -- Таблица для основной информации о заказе
        CREATE TABLE orders (
            order_uid VARCHAR(255) PRIMARY KEY,
            track_number VARCHAR(255),
            entry VARCHAR(50),
            locale VARCHAR(50),
            internal_signature TEXT,
            customer_id VARCHAR(255),
            delivery_service VARCHAR(255),
            shardkey VARCHAR(50),
            sm_id INTEGER,
            date_created TIMESTAMP WITH TIME ZONE, -- или TEXT
            oof_shard VARCHAR(50),

            delivery_uid VARCHAR(255) UNIQUE REFERENCES delivery(uid) ON DELETE SET NULL, -- Уникальная ссылка
            payment_transaction VARCHAR(255) UNIQUE REFERENCES payment(transaction) ON DELETE SET NULL -- Уникальная ссылка
        );

        -- Таблица для товаров в заказе
        CREATE TABLE items (
            id SERIAL PRIMARY KEY, -- Автоинкрементный ID для строки в items
            chrt_id INTEGER,
            track_number VARCHAR(255),
            price INTEGER,
            rid VARCHAR(255),
            name VARCHAR(255),
            sale INTEGER,
            size VARCHAR(50),
            total_price INTEGER,
            nm_id INTEGER,
            brand VARCHAR(255),
            status INTEGER,

            order_uid VARCHAR(255) REFERENCES orders(order_uid) ON DELETE CASCADE -- Внешний ключ на заказ
        );

        -- Индексы (по необходимости)
        CREATE INDEX idx_items_order_uid ON items(order_uid);
        CREATE UNIQUE INDEX idx_items_order_uid_chrt_id ON items (order_uid, chrt_id);
    `
	_, err = testDB.Exec(createTablesSQL)
	if err != nil {
		log.Fatalf("TestMain: Failed to setup test DB schema: %v", err)
	}
	log.Println("TestMain: Tables created successfully.")

	// Создаём репозиторий, используя соединение с установленным search_path
	// repo будет использовать testDB, у которого search_path установлен на test_schema.
	// repo = repository.NewPostgresOrderRepository(testDB)
	log.Println("TestMain: Repository for tests created.")

	// Запускаем тесты
	log.Println("TestMain: Running tests...")
	exitCode := m.Run()
	log.Printf("TestMain: Tests finished with exit code: %d", exitCode)

	// 5. После выполнения тестов - удаляем тестовую схему (и всё её содержимое)
	// Это гарантирует чистоту для следующего запуска тестов.
	log.Println("TestMain: Cleaning up - dropping test schema...")
	_, err = testDB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE;", testSchemaName))
	if err != nil {
		log.Printf("TestMain: Warning: Failed to drop test schema %s: %v", testSchemaName, err)
		// Не вызываем os.Exit(1) здесь, чтобы не перезаписать exitCode от тестов
	}

	// Закрываем соединение
	// testDB.Close() вызывается defer
	log.Println("TestMain: Exiting...")
	os.Exit(exitCode)
}

// getEnvOrDefault возвращает значение переменной окружения или значение по умолчанию.
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// --- Тесты ---

// TestNewOrderService тестирует создание OrderService.
func TestNewOrderService(t *testing.T) {
	// 1. Подготовка: Создаём конфигурацию для теста
	// Используем уникальные порты и идентификаторы для теста, чтобы не конфликтовать с основным сервисом или другими тестами
	testHTTPPort := ":0" // ":0" означает, что ОС выберет свободный порт
	testNATSClientID := "order_service_client_test_new"

	cfg := &config.Config{
		DB: config.DBConfig{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvOrDefault("DB_PORT", "5432"),
			User:     getEnvOrDefault("DB_USER", "order_user"),
			Password: getEnvOrDefault("DB_PASSWORD", "admin"),
			Name:     getEnvOrDefault("DB_NAME", "order_db"), // <-- ВАЖНО: используем основную БД, но search_path установлен на test_schema
			SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
		},
		NATS: config.NATSConfig{
			URL:       getEnvOrDefault("NATS_URL", "nats://localhost:4223"),
			ClusterID: getEnvOrDefault("NATS_CLUSTER_ID", "test-cluster"),
			ClientID:  testNATSClientID, // <-- Уникальный ClientID для теста
			Channel:   getEnvOrDefault("NATS_CHANNEL", "orders"),
		},
		Server: config.ServerConfig{
			HTTPPort: testHTTPPort, // <-- Используем ":0"
		},
	}

	// 2. Тестирование: Вызываем тестируемую функцию
	service, err := NewOrderService(cfg)
	require.NoError(t, err, "NewOrderService should not return an error")

	// 3. Проверка: Убедимся, что сервис создан и его поля инициализированы
	assert.NotNil(t, service)
	assert.NotNil(t, service.config)
	assert.NotNil(t, service.db)
	assert.NotNil(t, service.repo)
	assert.NotNil(t, service.cache)
	assert.NotNil(t, service.natsClient)
	// assert.NotNil(t, service.natsSub) // natsSub инициализируется в Start
	assert.NotNil(t, service.httpServer)

	// 4. Очистка: Останавливаем сервис
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = service.Stop(ctx)
	assert.NoError(t, err, "Stop should not return an error")
}

// TestOrderService_StartStop тестирует запуск и остановку OrderService.
func TestOrderService_StartStop(t *testing.T) {
	// 1. Подготовка: Создаём конфигурацию для теста
	// Используем уникальные порты и идентификаторы для теста
	testHTTPPort := ":0" // ":0" означает, что ОС выберет свободный порт
	testNATSClientID := "order_service_client_test_start_stop"

	cfg := &config.Config{
		DB: config.DBConfig{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvOrDefault("DB_PORT", "5432"),
			User:     getEnvOrDefault("DB_USER", "order_user"),
			Password: getEnvOrDefault("DB_PASSWORD", "admin"),
			Name:     getEnvOrDefault("DB_NAME", "order_db"), // <-- ВАЖНО: используем основную БД, но search_path установлен на test_schema
			SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
		},
		NATS: config.NATSConfig{
			URL:       getEnvOrDefault("NATS_URL", "nats://localhost:4223"),
			ClusterID: getEnvOrDefault("NATS_CLUSTER_ID", "test-cluster"),
			ClientID:  testNATSClientID, // <-- Уникальный ClientID для теста
			Channel:   getEnvOrDefault("NATS_CHANNEL", "orders"),
		},
		Server: config.ServerConfig{
			HTTPPort: testHTTPPort, // <-- Используем ":0"
		},
	}

	// 2. Подготовка: Создаём сервис
	service, err := NewOrderService(cfg)
	require.NoError(t, err, "NewOrderService should not return an error")

	// 3. Тестирование: Запускаем сервис
	err = service.Start()
	require.NoError(t, err, "Start should not return an error")

	// 4. Проверка: Убедимся, что HTTP-сервер запущен (косвенно)
	// Мы не можем получить httpServer.Addr напрямую, так как оно приватное.
	// Но Start() уже проверил, что httpServer.ListenAndServe() не вернул ошибку.
	// Можно попробовать подключиться, но это сложно без знания порта.

	// 5. Проверка: Убедимся, что natsSub инициализирован
	assert.NotNil(t, service.natsSub, "natsSub should be initialized after Start")

	// 6. Очистка: Останавливаем сервис
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = service.Stop(ctx)
	assert.NoError(t, err, "Stop should not return an error")
}
