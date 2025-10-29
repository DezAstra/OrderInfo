package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"OrderInfo/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	testDB *sql.DB
	repo   OrderRepository
)

const (
	testSchemaName = "test_repo_schema_final"
)

// TestMain - точка входа для тестов пакета. Используется для настройки и очистки тестовой среды.
func TestMain(m *testing.M) {
	log.Println("TestMain: Начало инициализации тестовой среды...")

	// Используем переменные окружения или значения по умолчанию для тестовой БД
	// Подключаемся к основной БД, но будем использовать изолированную схему
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
	log.Println("TestMain: Удаление тестовой схемы (если существует)...")
	_, err = testDB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE;", testSchemaName))
	if err != nil {
		log.Fatalf("TestMain: Ошибка удаления существующей тестовой схемы %s: %v", testSchemaName, err)
	}

	// 2. Создаём новую тестовую схему
	log.Println("TestMain: Создание новой тестовой схемы...")
	_, err = testDB.Exec(fmt.Sprintf("CREATE SCHEMA %s;", testSchemaName))
	if err != nil {
		log.Fatalf("TestMain: Ошибка создания тестовой схемы %s: %v", testSchemaName, err)
	}

	// 3. Переключаемся на тестовую схему для этой сессии подключения testDB
	log.Println("TestMain: Установка search_path на тестовую схему...")
	_, err = testDB.Exec(fmt.Sprintf("SET search_path TO %s;", testSchemaName))
	if err != nil {
		log.Fatalf("TestMain: Ошибка установки search_path на %s: %v", testSchemaName, err)
	}

	// 4. Создаём таблицы в тестовой схеме
	// Порядок создания важен из-за внешних ключей.
	log.Println("TestMain: Создание таблиц в тестовой схеме...")
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
		log.Fatalf("TestMain: Ошибка создания таблиц в тестовой схеме: %v", err)
	}
	log.Println("TestMain: Таблицы успешно созданы.")

	// Создаём репозиторий, используя соединение с установленным search_path
	// repo будет использовать testDB, у которого search_path установлен на test_repo_schema.
	repo = NewPostgresOrderRepository(testDB)
	log.Println("TestMain: Репозиторий для тестов создан.")

	// Запуск тестов
	log.Println("TestMain: Запуск тестов...")
	exitCode := m.Run()
	log.Printf("TestMain: Тесты завершены с кодом выхода: %d", exitCode)

	// 5. удаление тестовой схемы
	log.Println("TestMain: Очистка - удаление тестовой схемы...")
	_, err = testDB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE;", testSchemaName))
	if err != nil {
		log.Printf("TestMain: Warning: Failed to drop test schema %s: %v", testSchemaName, err)
	}

	// testDB.Close() будет вызван defer
	log.Println("TestMain: Завершение.")
	os.Exit(exitCode)
}

// getEnvOrDefault возвращает значение переменной окружения или значение по умолчанию.
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// clearTables очищает все таблицы в тестовой схеме.
func clearTables(t *testing.T) {
	t.Helper() // Помечает функцию как помощник для тестов, чтобы трейсы указывали на вызывающий тест
	_, err := testDB.Exec(`
        DELETE FROM items;
        DELETE FROM orders;
        DELETE FROM payment;
        DELETE FROM delivery;
    `)
	require.NoError(t, err, "Failed to clear test tables")
}

// Тесты

func TestPostgresOrderRepository_SaveAndGet(t *testing.T) {
	// очистка таблиц перед тестом для изоляции
	clearTables(t)

	// Подготовка тестовых данных
	testTime, err := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
	require.NoError(t, err, "Failed to parse test time")

	order := model.Order{
		OrderUID:        "test-save-get-uid-final-corrected",
		TrackNumber:     "TESTTRACKFINALCORRECTED",
		Entry:           "TEST",
		Locale:          "en",
		CustomerID:      "test_customer_final_corrected",
		DeliveryService: "test_service_final_corrected",
		Shardkey:        "1",
		SmID:            1,
		OofShard:        "1",
		DateCreated:     testTime,
		Delivery: model.Delivery{
			Name:    "Test Name Final Corrected",
			Phone:   "+1111111111",
			Zip:     "11111",
			City:    "Test City Final Corrected",
			Address: "Test Address Final Corrected",
			Region:  "Test Region Final Corrected",
			Email:   "testfinalcorrected@example.com",
		},
		Payment: model.Payment{
			Transaction:  "test-transaction-id-save-get-final-corrected",
			RequestID:    "reqfinalcorrected123",
			Currency:     "USD",
			Provider:     "test_provider_final_corrected",
			Amount:       1500,
			PaymentDt:    1672531201, // Unix timestamp
			Bank:         "Test Bank Final Corrected",
			DeliveryCost: 300,
			GoodsTotal:   1200,
			CustomFee:    0,
		},
		Items: []model.Item{
			{
				ChrtID:      99999,
				TrackNumber: "TESTTRACKFINALCORRECTED",
				Price:       600,
				Rid:         "ridfinalcorrected1",
				Name:        "Test Item Final Corrected 1",
				Sale:        0,
				Size:        "L",
				TotalPrice:  600,
				NmID:        99999,
				Brand:       "Test Brand Final Corrected",
				Status:      202,
			},
		},
	}

	// Тестируем SaveOrder
	err = repo.SaveOrder(order)
	require.NoError(t, err, "SaveOrder should not return an error")

	// Тестируем GetOrder
	retrievedOrder, err := repo.GetOrder(order.OrderUID)
	require.NoError(t, err, "GetOrder should not return an error for existing order")
	assert.Equal(t, order.OrderUID, retrievedOrder.OrderUID)
	assert.Equal(t, order.TrackNumber, retrievedOrder.TrackNumber)
	assert.True(t, order.DateCreated.Equal(retrievedOrder.DateCreated))
	assert.Equal(t, order.Delivery.Name, retrievedOrder.Delivery.Name)
	assert.Equal(t, order.Payment.Transaction, retrievedOrder.Payment.Transaction)
	assert.Equal(t, len(order.Items), len(retrievedOrder.Items))
	if len(order.Items) > 0 && len(retrievedOrder.Items) > 0 {
		assert.Equal(t, order.Items[0].Name, retrievedOrder.Items[0].Name)
		assert.Equal(t, order.Items[0].ChrtID, retrievedOrder.Items[0].ChrtID)
	}
}

func TestPostgresOrderRepository_GetNonExistent(t *testing.T) {
	// очистка таблиц перед тестом для изоляции
	clearTables(t)

	_, err := repo.GetOrder("non-existent-uid-final-corrected")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не найден") // Проверяем сообщение из кода ошибки в GetOrder
}

func TestPostgresOrderRepository_GetAllOrders(t *testing.T) {
	// Очищаем таблицы перед тестом для изоляции
	clearTables(t)

	// проверка, что изначально БД пуста для теста
	allOrdersBefore, err := repo.GetAllOrders()
	require.NoError(t, err)
	assert.Equal(t, 0, len(allOrdersBefore))

	// создание и сохранение 2 заказов
	testTime1, err := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
	require.NoError(t, err)
	testTime2, err := time.Parse(time.RFC3339, "2023-01-02T00:00:00Z")
	require.NoError(t, err)

	order1 := model.Order{
		OrderUID:    "get-all-test-1-final-corrected",
		TrackNumber: "TRACK1FINALCORRECTED",
		DateCreated: testTime1,
		Delivery: model.Delivery{
			Name:    "Test Name 1 Final Corrected",
			Phone:   "+1234567890",
			Zip:     "12345",
			City:    "Test City 1 Final Corrected",
			Address: "Test Address 1 Final Corrected",
			Region:  "Test Region 1 Final Corrected",
			Email:   "test1finalcorrected@example.com",
		},
		Payment: model.Payment{
			Transaction:  "test-transaction-id-get-all-1-final-corrected",
			RequestID:    "req123finalcorrected",
			Currency:     "USD",
			Provider:     "test_provider_final_corrected",
			Amount:       1000,
			PaymentDt:    1672531200,
			Bank:         "Test Bank Final Corrected",
			DeliveryCost: 200,
			GoodsTotal:   800,
			CustomFee:    0,
		},
		Items: []model.Item{
			{
				ChrtID:      11111,
				TrackNumber: "TRACK1FINALCORRECTED",
				Price:       100,
				Rid:         "rid11finalcorrected",
				Name:        "Test Item 1.1 Final Corrected",
				Sale:        0,
				Size:        "S",
				TotalPrice:  100,
				NmID:        11111,
				Brand:       "Test Brand 1 Final Corrected",
				Status:      202,
			},
		},
	}
	order2 := model.Order{
		OrderUID:    "get-all-test-2-final-corrected",
		TrackNumber: "TRACK2FINALCORRECTED",
		DateCreated: testTime2,
		Delivery: model.Delivery{
			Name:    "Test Name 2 Final Corrected",
			Phone:   "+0987654321",
			Zip:     "54321",
			City:    "Test City 2 Final Corrected",
			Address: "Test Address 2 Final Corrected",
			Region:  "Test Region 2 Final Corrected",
			Email:   "test2finalcorrected@example.com",
		},
		Payment: model.Payment{
			Transaction:  "test-transaction-id-get-all-2-final-corrected",
			RequestID:    "req456finalcorrected",
			Currency:     "EUR",
			Provider:     "test_provider_2_final_corrected",
			Amount:       2000,
			PaymentDt:    1672617600,
			Bank:         "Test Bank 2 Final Corrected",
			DeliveryCost: 300,
			GoodsTotal:   1700,
			CustomFee:    10,
		},
		Items: []model.Item{
			{
				ChrtID:      22222,
				TrackNumber: "TRACK2FINALCORRECTED",
				Price:       200,
				Rid:         "rid22finalcorrected",
				Name:        "Test Item 2.1 Final Corrected",
				Sale:        5,
				Size:        "M",
				TotalPrice:  190,
				NmID:        22222,
				Brand:       "Test Brand 2 Final Corrected",
				Status:      203,
			},
		},
	}

	err1 := repo.SaveOrder(order1)
	err2 := repo.SaveOrder(order2)
	require.NoError(t, err1)
	require.NoError(t, err2)

	// Получение всех заказов
	allOrders, err := repo.GetAllOrders()
	require.NoError(t, err)

	// Проверка количества
	assert.Equal(t, 2, len(allOrders))

	// Проверка, что оба заказа есть (проверим UID и Transaction)
	foundUID1 := false
	foundUID2 := false
	for _, o := range allOrders {
		if o.OrderUID == order1.OrderUID {
			foundUID1 = true
			assert.Equal(t, order1.TrackNumber, o.TrackNumber)
			assert.True(t, order1.DateCreated.Equal(o.DateCreated))
			assert.Equal(t, order1.Payment.Transaction, o.Payment.Transaction)
		}
		if o.OrderUID == order2.OrderUID {
			foundUID2 = true
			assert.Equal(t, order2.TrackNumber, o.TrackNumber)
			assert.True(t, order2.DateCreated.Equal(o.DateCreated))
			assert.Equal(t, order2.Payment.Transaction, o.Payment.Transaction)
		}
	}

	assert.True(t, foundUID1, "Order 1 should be in GetAllOrders result")
	assert.True(t, foundUID2, "Order 2 should be in GetAllOrders result")

	// Убедимся, что не нашли лишнего
	assert.Equal(t, 2, len(allOrders))
}
