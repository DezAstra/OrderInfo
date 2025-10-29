package nats

import (
	"OrderInfo/internal/cache"
	"OrderInfo/internal/model"
	"OrderInfo/internal/repository"
	"encoding/json"
	"log"

	"github.com/nats-io/stan.go"
)

// NewMessageHandler возвращает функцию-обработчик сообщений NATS Streaming.
// Эта функция замыкает (closes over) зависимости repo и cache.
func NewMessageHandler(repo repository.OrderRepository, cache cache.Cache) stan.MsgHandler {
	return func(msg *stan.Msg) {
		// Вызываем внутреннюю функцию для лучшей читаемости
		handleMessage(msg, repo, cache)
	}
}

// handleMessage содержит основную логику обработки полученного сообщения.
func handleMessage(msg *stan.Msg, repo repository.OrderRepository, cache cache.Cache) {
	log.Printf("Получено сообщение из NATS, Sequence: %d, размер данных: %d байт", msg.Sequence, len(msg.Data))

	var order model.Order

	// 1. Десериализация JSON
	if err := json.Unmarshal(msg.Data, &order); err != nil {
		log.Printf("Ошибка десериализации JSON из сообщения NATS (Sequence: %d): %v. Данные: %s", msg.Sequence, err, string(msg.Data))
		return
	}

	// 2. Проверка валидности структуры
	if order.OrderUID == "" {
		log.Printf("Получен заказ без OrderUID (Sequence: %d). Данные: %s", msg.Sequence, string(msg.Data))
		return
	}

	log.Printf("Десериализован заказ с UID: %s (из сообщения NATS Sequence: %d)", order.OrderUID, msg.Sequence)

	// 3. Сохранение в БД
	if err := repo.SaveOrder(order); err != nil {
		log.Printf("Ошибка сохранения заказа %s (из сообщения NATS Sequence: %d) в БД: %v", order.OrderUID, msg.Sequence, err)
		return
	}
	log.Printf("Заказ %s (из сообщения NATS Sequence: %d) успешно сохранён в БД", order.OrderUID, msg.Sequence)

	// 4. Обновление кэша
	cache.Set(order.OrderUID, order)
	log.Printf("Заказ %s (из сообщения NATS Sequence: %d) успешно обновлён в кэше", order.OrderUID, msg.Sequence)

	// 5. Логирование успеха
	log.Printf("Обработка сообщения Sequence %d для заказа %s завершена успешно.", msg.Sequence, order.OrderUID)
}
