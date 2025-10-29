package model

import "time"

// Order представляет основную сущность заказа.
type Order struct {
	OrderUID          string    `json:"order_uid"`          // ID заказа
	TrackNumber       string    `json:"track_number"`       // Номер отслеживания
	Entry             string    `json:"entry"`              // Код точки входа
	Delivery          Delivery  `json:"delivery"`           // Информация о доставке
	Payment           Payment   `json:"payment"`            // Информация о платеже
	Items             []Item    `json:"items"`              // Список товаров
	Locale            string    `json:"locale"`             // Язык
	InternalSignature string    `json:"internal_signature"` // Внутренняя подпись (часто пустая)
	CustomerID        string    `json:"customer_id"`        // ID клиента
	DeliveryService   string    `json:"delivery_service"`   // Сервис доставки
	Shardkey          string    `json:"shardkey"`           // Ключ шардирования
	SmID              int       `json:"sm_id"`              // ID службы доставки (SM - Service Manager?)
	DateCreated       time.Time `json:"date_created"`       // Дата создания (как time.Time)
	// DateCreated    string    `json:"date_created"`       // Дата создания (как строка)
	OofShard string `json:"oof_shard"` // OOF шард (Order of Fulfillment?)
}

// Delivery содержит информацию о получателе и доставке.
type Delivery struct {
	Name    string `json:"name"`    // Имя получателя
	Phone   string `json:"phone"`   // Телефон получателя
	Zip     string `json:"zip"`     // Почтовый индекс
	City    string `json:"city"`    // Город
	Address string `json:"address"` // Полный адрес
	Region  string `json:"region"`  // Регион
	Email   string `json:"email"`   // Email получателя
}

// Payment содержит информацию о платеже.
type Payment struct {
	Transaction  string `json:"transaction"`   // ID транзакции
	RequestID    string `json:"request_id"`    // ID запроса (часто пустой)
	Currency     string `json:"currency"`      // Валюта (USD)
	Provider     string `json:"provider"`      // Провайдер оплаты (wbpay)
	Amount       int    `json:"amount"`        // Общая сумма
	PaymentDt    int64  `json:"payment_dt"`    // Дата платежа (Unix timestamp)
	Bank         string `json:"bank"`          // Банк
	DeliveryCost int    `json:"delivery_cost"` // Стоимость доставки
	GoodsTotal   int    `json:"goods_total"`   // Стоимость товаров
	CustomFee    int    `json:"custom_fee"`    // Таможенная пошлина
}

// Item представляет товар в заказе.
type Item struct {
	ChrtID      int    `json:"chrt_id"`      // ID характеристики товара
	TrackNumber string `json:"track_number"` // Номер отслеживания (дублируется?)
	Price       int    `json:"price"`        // Цена за единицу
	Rid         string `json:"rid"`          // Уникальный ID записи товара
	Name        string `json:"name"`         // Название товара
	Sale        int    `json:"sale"`         // Скидка в процентах
	Size        string `json:"size"`         // Размер
	TotalPrice  int    `json:"total_price"`  // Итоговая цена (с учётом скидки?)
	NmID        int    `json:"nm_id"`        // Артикул WB
	Brand       string `json:"brand"`        // Бренд
	Status      int    `json:"status"`       // Статус товара
}
