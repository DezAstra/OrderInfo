// OrderInfo/internal/http/handlers_test.go
package http

import (
	"OrderInfo/internal/cache"
	"OrderInfo/internal/model"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOrderHandler(t *testing.T) {
	// 1. создание мок-кэш и добавляем туда тестовый заказ
	testCache := cache.NewInMemoryCache()
	testOrderUID := "b563feb7b2b84b6test-handler-final-check"
	testTime, err := time.Parse(time.RFC3339, "2021-11-26T06:22:19Z")
	require.NoError(t, err, "Failed to parse test time for order")

	testOrder := model.Order{
		OrderUID:        testOrderUID,
		TrackNumber:     "WBILMTESTTRACK-FINAL",
		Entry:           "WBIL-FINAL",
		Locale:          "en-final",
		CustomerID:      "test_customer_final",
		DeliveryService: "meest_final",
		Shardkey:        "9-final",
		SmID:            99,
		OofShard:        "1-final",
		DateCreated:     testTime,
		Delivery: model.Delivery{
			Name:    "Test Testov Final",
			Phone:   "+9720000000-final",
			Zip:     "2639809-final",
			City:    "Kiryat Mozkin Final",
			Address: "Ploshad Mira 15 Final",
			Region:  "Kraiot Final",
			Email:   "test_final@gmail.com",
		},
		Payment: model.Payment{
			Transaction:  "b563feb7b2b84b6test-final", // <-- Уникальный Transaction
			RequestID:    "req-final-123",
			Currency:     "USD-final",
			Provider:     "wbpay-final",
			Amount:       1817,
			PaymentDt:    1637907727,
			Bank:         "alpha-final",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []model.Item{
			{
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK-FINAL",
				Price:       453,
				Rid:         "ab4219087a764ae0btest-final",
				Name:        "Mascaras Final",
				Sale:        30,
				Size:        "0-final",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo Final",
				Status:      202,
			},
		},
	}

	testCache.Set(testOrderUID, testOrder)

	// 2. создание HTTP-обработчика (http.HandlerFunc)
	handlerFunc := GetOrderHandler(testCache)

	// 3. создание тестового HTTP-запроса (GET /order/{uid})
	req, err := http.NewRequest("GET", "/order/"+testOrderUID, nil)
	require.NoError(t, err, "Failed to create test HTTP request")

	// 4. Создание Recorder для записи ответа
	rr := httptest.NewRecorder()

	// 5. Вызов обработчика (http.HandlerFunc)
	handlerFunc.ServeHTTP(rr, req)

	// 6. Проверка кода состояния
	assert.Equal(t, http.StatusOK, rr.Code, "Handler returned wrong status code")

	// 7. Проверка, что тело ответа содержит ожидаемый JSON
	var responseOrder model.Order
	err = json.Unmarshal(rr.Body.Bytes(), &responseOrder)
	require.NoError(t, err, "Response body should be valid JSON")

	// 8. Проверка поля возвращенного заказа
	assert.Equal(t, testOrderUID, responseOrder.OrderUID)
	assert.Equal(t, "WBILMTESTTRACK-FINAL", responseOrder.TrackNumber)
	assert.True(t, testTime.Equal(responseOrder.DateCreated))
	assert.Equal(t, "Test Testov Final", responseOrder.Delivery.Name)
	assert.Equal(t, "b563feb7b2b84b6test-final", responseOrder.Payment.Transaction) // <-- Проверяем уникальный Transaction
	assert.Equal(t, 1, len(responseOrder.Items))
	if len(responseOrder.Items) > 0 {
		assert.Equal(t, "Mascaras Final", responseOrder.Items[0].Name)
		assert.Equal(t, 9934930, responseOrder.Items[0].ChrtID)
	}
}

func TestGetOrderHandlerNotFound(t *testing.T) {
	// 1. Создание пустого мок-кэша
	testCache := cache.NewInMemoryCache()

	// 2. Создание HTTP-обработчика (http.HandlerFunc)
	handlerFunc := GetOrderHandler(testCache)

	// 3. Создание тестового HTTP-запрос для несуществующего UID
	req, err := http.NewRequest("GET", "/order/non-existent-uid-final-check", nil)
	require.NoError(t, err, "Failed to create test HTTP request for non-existent order")

	// 4. Создание Recorder
	rr := httptest.NewRecorder()

	// 5. Вызов обработчика (http.HandlerFunc)
	handlerFunc.ServeHTTP(rr, req)

	// 6. Проверка кода состояния 404
	assert.Equal(t, http.StatusNotFound, rr.Code, "Handler should return 404 for non-existent order")
}
