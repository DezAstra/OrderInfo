package nats

import (
	"OrderInfo/internal/model"
	"errors"
	"reflect"
	"testing"

	"github.com/nats-io/stan.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOrderRepository имплементирует repository.OrderRepository для тестов.
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) SaveOrder(order model.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetOrder(uid string) (model.Order, error) {
	args := m.Called(uid)
	order, _ := args.Get(0).(model.Order)
	return order, args.Error(1)
}

func (m *MockOrderRepository) GetAllOrders() ([]model.Order, error) {
	args := m.Called()
	orders, _ := args.Get(0).([]model.Order)
	return orders, args.Error(1)
}

// MockCache имплементирует cache.Cache для тестов.
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Set(uid string, order model.Order) {
	m.Called(uid, order)
}

func (m *MockCache) Get(uid string) (model.Order, bool) {
	args := m.Called(uid)
	order, _ := args.Get(0).(model.Order)
	return order, args.Bool(1)
}

func (m *MockCache) Delete(uid string) {
	m.Called(uid)
}

func (m *MockCache) GetAll() []model.Order {
	args := m.Called()
	orders, _ := args.Get(0).([]model.Order)
	return orders
}

func TestNewMessageHandler(t *testing.T) {
	// моки
	mockRepo := new(MockOrderRepository)
	mockCache := new(MockCache)

	// обработчик сообщений
	handler := NewMessageHandler(mockRepo, mockCache)

	// Проверка, что handler не nil
	assert.NotNil(t, handler)
}

func TestMessageHandler_HandleValidJSON(t *testing.T) {
	// 1. Создание моков для репозитория и кэша
	mockRepo := new(MockOrderRepository)
	mockCache := new(MockCache)

	// 2. обработчик сообщений, передача моков
	handler := NewMessageHandler(mockRepo, mockCache)

	// 3. model.JSON
	validJSON := []byte(`{
  "order_uid": "b563feb7b2b84b6test-handler",
  "track_number": "WBILMTESTTRACK-HANDLER",
  "entry": "WBIL-HANDLER",
  "delivery": {
    "name": "Test Testov Handler",
    "phone": "+9720000000-handler",
    "zip": "2639809-handler",
    "city": "Kiryat Mozkin Handler",
    "address": "Ploshad Mira 15 Handler",
    "region": "Kraiot Handler",
    "email": "test_handler@gmail.com"
  },
  "payment": {
    "transaction": "b563feb7b2b84b6test-handler-transaction",
    "request_id": "req-handler-123",
    "currency": "USD-handler",
    "provider": "wbpay-handler",
    "amount": 1817,
    "payment_dt": 1637907727,
    "bank": "alpha-handler",
    "delivery_cost": 1500,
    "goods_total": 317,
    "custom_fee": 0
  },
  "items": [
    {
      "chrt_id": 9934930,
      "track_number": "WBILMTESTTRACK-HANDLER",
      "price": 453,
      "rid": "ab4219087a764ae0btest-handler",
      "name": "Mascaras Handler",
      "sale": 30,
      "size": "0-handler",
      "total_price": 317,
      "nm_id": 2389212,
      "brand": "Vivienne Sabo Handler",
      "status": 202
    }
  ],
  "locale": "en-handler",
  "internal_signature": "",
  "customer_id": "test_customer_handler",
  "delivery_service": "meest_handler",
  "shardkey": "9-handler",
  "sm_id": 99,
  "date_created": "2021-11-26T06:22:19Z",
  "oof_shard": "1-handler"
}`)

	// 4. Создание фейкового сообщения NATS Streaming
	fakeMsg := &stan.Msg{}
	v := reflect.ValueOf(fakeMsg).Elem()
	if dataField := v.FieldByName("Data"); dataField.IsValid() && dataField.CanSet() {
		dataField.Set(reflect.ValueOf(validJSON))
	} else {
		t.Fatalf("Cannot set Data field on stan.Msg. Field may be unexported or not exist.")
	}

	// 5. Настраиваем ожидания для моков
	// Ожидание, что SaveOrder будет вызван с любым model.Order и вернет nil
	mockRepo.On("SaveOrder", mock.AnythingOfType("model.Order")).Return(nil)
	// Ожидание, что cache.Set будет вызван с любыми аргументами
	mockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("model.Order")).Return()

	// 6. Вызов обработчика
	handler(fakeMsg)

	// 7. Проверяем, что моки были вызваны как ожидается
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)

	// Дополнительно: можно проверить, что вызов был ровно 1 раз
	mockRepo.AssertNumberOfCalls(t, "SaveOrder", 1)
	mockCache.AssertNumberOfCalls(t, "Set", 1)
}

func TestMessageHandler_HandleInvalidJSON(t *testing.T) {
	// 1. моки
	mockRepo := new(MockOrderRepository)
	mockCache := new(MockCache)

	// 2. обработчик сообщений
	handler := NewMessageHandler(mockRepo, mockCache)

	// 3. подготавливаем НЕКОРРЕКТНЫЙ JSON
	invalidJSON := []byte(`{"order_uid": "test", "invalid_field": }`) // Некорректный JSON

	// 4. фейковое сообщение NATS Streaming
	fakeMsg := &stan.Msg{}
	v := reflect.ValueOf(fakeMsg).Elem()
	if dataField := v.FieldByName("Data"); dataField.IsValid() && dataField.CanSet() {
		dataField.Set(reflect.ValueOf(invalidJSON))
	} else {
		t.Fatalf("Cannot set Data field on stan.Msg. Field may be unexported or not exist.")
	}

	// 5. Вызов обработчика
	handler(fakeMsg)

	// 6. проверка, что методы моков НЕ БЫЛИ вызваны
	mockRepo.AssertNotCalled(t, "SaveOrder")
	mockCache.AssertNotCalled(t, "Set")
}

func TestMessageHandler_RepoError(t *testing.T) {
	// 1. моки
	mockRepo := new(MockOrderRepository)
	mockCache := new(MockCache)

	// 2. обработчик сообщений
	handler := NewMessageHandler(mockRepo, mockCache)

	// 3. Подготавливаем корректный JSON
	validJSON := []byte(`{
  "order_uid": "b563feb7b2b84b6test-repo-error",
  "track_number": "WBILMTESTTRACK-REPO-ERROR",
  "entry": "WBIL-REPO-ERROR",
  "delivery": { "name": "Test Testov Repo Error", "phone": "+1234567890", "zip": "12345", "city": "Test City", "address": "Test Address", "region": "Test Region", "email": "test@example.com" },
  "payment": { "transaction": "b563feb7b2b84b6test-repo-error-transaction", "request_id": "req123", "currency": "USD", "provider": "test_provider", "amount": 1000, "payment_dt": 1637907727, "bank": "Test Bank", "delivery_cost": 200, "goods_total": 800, "custom_fee": 0 },
  "items": [ { "chrt_id": 12345, "track_number": "WBILMTESTTRACK-REPO-ERROR", "price": 500, "rid": "rid1", "name": "Test Item", "sale": 0, "size": "M", "total_price": 500, "nm_id": 54321, "brand": "Test Brand", "status": 202 } ],
  "locale": "en", "internal_signature": "", "customer_id": "test_customer", "delivery_service": "meest", "shardkey": "1", "sm_id": 99, "date_created": "2021-11-26T06:22:19Z", "oof_shard": "1"
}`)

	// 4. Создаем фейковое сообщение NATS Streaming
	fakeMsg := &stan.Msg{}
	v := reflect.ValueOf(fakeMsg).Elem()
	if dataField := v.FieldByName("Data"); dataField.IsValid() && dataField.CanSet() {
		dataField.Set(reflect.ValueOf(validJSON))
	} else {
		t.Fatalf("Cannot set Data field on stan.Msg. Field may be unexported or not exist.")
	}

	// 5. Настраиваем ожидания для моков
	testError := errors.New("simulated repo error in nats test")
	mockRepo.On("SaveOrder", mock.AnythingOfType("model.Order")).Return(testError)
	// Ожидание, что cache.Set НЕ БУДЕТ вызван, потому что SaveOrder вернул ошибку
	mockCache.AssertNotCalled(t, "Set")

	// 6. Вызываем обработчик
	handler(fakeMsg)

	// 7. Проверяем ожидания
	mockRepo.AssertExpectations(t)
	mockRepo.AssertCalled(t, "SaveOrder", mock.AnythingOfType("model.Order"))
	mockCache.AssertNotCalled(t, "Set")
}
