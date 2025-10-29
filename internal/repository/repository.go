package repository

import (
	"OrderInfo/internal/model"
	"database/sql"
	"fmt"
	"log"
)

// OrderRepository определяет интерфейс для работы с заказами в хранилище.
type OrderRepository interface {
	SaveOrder(order model.Order) error
	GetOrder(uid string) (model.Order, error)
	GetAllOrders() ([]model.Order, error)
}

// PostgresOrderRepository реализует OrderRepository с использованием PostgreSQL.
type PostgresOrderRepository struct {
	db *sql.DB
}

// NewPostgresOrderRepository создает новый экземпляр PostgresOrderRepository.
func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{
		db: db,
	}
}

// SaveOrder сохраняет заказ в базе данных.
// Предполагается, что order.OrderUID уникален. Если заказ с таким UID уже существует, может произойти ошибка вставки (в зависимости от настроек БД/SQL).
func (r *PostgresOrderRepository) SaveOrder(order model.Order) error {
	tx, err := r.db.Begin() // для согласованности
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}

	// Откатываем транзакцию в случае ошибки, если она не была зафиксирована
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("Ошибка при откате транзакции: %v", rbErr)
			}
		}
	}()

	// 1. Вставка данных доставки
	deliveryUID := order.OrderUID
	_, err = tx.Exec(`
        INSERT INTO delivery (uid, name, phone, zip, city, address, region, email)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (uid) DO UPDATE SET -- Обновляем, если уже существует (опционально)
        name = EXCLUDED.name,
        phone = EXCLUDED.phone,
        zip = EXCLUDED.zip,
        city = EXCLUDED.city,
        address = EXCLUDED.address,
        region = EXCLUDED.region,
        email = EXCLUDED.email
    `,
		deliveryUID,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email,
	)
	if err != nil {
		return fmt.Errorf("ошибка при вставке данных доставки: %w", err)
	}

	// 2. Вставка данных платежа
	_, err = tx.Exec(`
        INSERT INTO payment (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        ON CONFLICT (transaction) DO UPDATE SET -- Обновляем, если уже существует (опционально)
        request_id = EXCLUDED.request_id,
        currency = EXCLUDED.currency,
        provider = EXCLUDED.provider,
        amount = EXCLUDED.amount,
        payment_dt = EXCLUDED.payment_dt,
        bank = EXCLUDED.bank,
        delivery_cost = EXCLUDED.delivery_cost,
        goods_total = EXCLUDED.goods_total,
        custom_fee = EXCLUDED.custom_fee
    `,
		order.Payment.Transaction,
		order.Payment.RequestID,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDt,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("ошибка при вставке данных платежа: %w", err)
	}

	// 3. Вставка основных данных заказа
	_, err = tx.Exec(`
        INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard, delivery_uid, payment_transaction)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
        ON CONFLICT (order_uid) DO UPDATE SET -- Обновляем, если уже существует (опционально)
        track_number = EXCLUDED.track_number,
        entry = EXCLUDED.entry,
        locale = EXCLUDED.locale,
        internal_signature = EXCLUDED.internal_signature,
        customer_id = EXCLUDED.customer_id,
        delivery_service = EXCLUDED.delivery_service,
        shardkey = EXCLUDED.shardkey,
        sm_id = EXCLUDED.sm_id,
        date_created = EXCLUDED.date_created,
        oof_shard = EXCLUDED.oof_shard,
        delivery_uid = EXCLUDED.delivery_uid,
        payment_transaction = EXCLUDED.payment_transaction
    `,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.Shardkey,
		order.SmID,
		order.DateCreated,
		order.OofShard,
		deliveryUID,
		order.Payment.Transaction,
	)
	if err != nil {
		return fmt.Errorf("ошибка при вставке основных данных заказа: %w", err)
	}

	// 4. Вставка данных товаров

	for _, item := range order.Items {
		_, err = tx.Exec(`
            INSERT INTO items (chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status, order_uid)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
            ON CONFLICT (order_uid, chrt_id) DO NOTHING -- Изменено: используем столбцы, а не имя ограничения
            -- Или ON CONFLICT (order_uid, chrt_id) DO UPDATE SET ...
        `,
			item.ChrtID,
			item.TrackNumber,
			item.Price,
			item.Rid,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NmID,
			item.Brand,
			item.Status,
			order.OrderUID,
		)
		if err != nil {
			return fmt.Errorf("ошибка при вставке товара: %w", err)
		}
	}

	// 5. Фиксируем транзакцию
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("ошибка при фиксации транзакции: %w", err)
	}

	log.Printf("Заказ %s успешно сохранен в БД", order.OrderUID)
	return nil
}

// GetOrder извлекает заказ из базы данных по его UID.
func (r *PostgresOrderRepository) GetOrder(uid string) (model.Order, error) {
	var order model.Order

	var (
		deliveryUIDFromOrder        string
		paymentTransactionFromOrder string
	)
	// 1. Получаем основные данные заказа
	err := r.db.QueryRow(`
        SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard,
               delivery_uid, payment_transaction
        FROM orders WHERE order_uid = $1
    `, uid).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
		&order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
		&deliveryUIDFromOrder, &paymentTransactionFromOrder,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return order, fmt.Errorf("заказ с UID %s не найден", uid)
		}
		return order, fmt.Errorf("ошибка при получении основных данных заказа: %w", err)
	}

	// 2. Получаем данные доставки по UID доставки
	err = r.db.QueryRow(`
        SELECT name, phone, zip, city, address, region, email
        FROM delivery WHERE uid = $1
    `, deliveryUIDFromOrder).Scan(
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
		&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
	)
	if err != nil {
		return order, fmt.Errorf("ошибка при получении данных доставки для заказа %s (delivery UID: %s): %w", uid, deliveryUIDFromOrder, err)
	}

	// 3. Получаем данные платежа по transaction
	err = r.db.QueryRow(`
        SELECT request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee, transaction
        FROM payment WHERE transaction = $1
    `, paymentTransactionFromOrder).Scan(
		&order.Payment.RequestID, &order.Payment.Currency, &order.Payment.Provider,
		&order.Payment.Amount, &order.Payment.PaymentDt, &order.Payment.Bank,
		&order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee,
		&order.Payment.Transaction,
	)
	if err != nil {
		return order, fmt.Errorf("ошибка при получении данных платежа для заказа %s (payment Transaction: %s): %w", uid, paymentTransactionFromOrder, err)
	}

	// 4. Получаем список товаров
	rows, err := r.db.Query(`
        SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
        FROM items WHERE order_uid = $1
    `, uid)
	if err != nil {
		return order, fmt.Errorf("ошибка при получении списка товаров для заказа %s: %w", uid, err)
	}
	defer rows.Close()

	var items []model.Item
	for rows.Next() {
		var item model.Item
		err = rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid,
			&item.Name, &item.Sale, &item.Size, &item.TotalPrice,
			&item.NmID, &item.Brand, &item.Status,
		)
		if err != nil {
			return order, fmt.Errorf("ошибка при сканировании строки товара для заказа %s: %w", uid, err)
		}
		items = append(items, item)
	}

	// Проверяем, были ли ошибки во время итерации rows.Next()
	if err = rows.Err(); err != nil {
		return order, fmt.Errorf("ошибка при итерации по строкам товаров для заказа %s: %w", uid, err)
	}

	order.Items = items

	log.Printf("Заказ %s успешно загружен из БД", order.OrderUID)
	return order, nil
}

// GetAllOrders извлекает все заказы из базы данных.
func (r *PostgresOrderRepository) GetAllOrders() ([]model.Order, error) {
	// 1. Получаем все основные данные заказов, включая связи на delivery и payment
	rows, err := r.db.Query(`
        SELECT o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
               d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
               p.request_id, p.currency, p.provider, p.amount, p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee, p.transaction -- <-- Добавили p.transaction
        FROM orders o
        JOIN delivery d ON o.delivery_uid = d.uid
        JOIN payment p ON o.payment_transaction = p.transaction
    `)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении списка всех заказов из БД: %w", err)
	}
	defer rows.Close()

	// Слайс для хранения всех заказов
	var allOrders []model.Order

	for rows.Next() {
		var order model.Order
		err = rows.Scan(
			&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
			&order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
			&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
			&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
			&order.Payment.RequestID, &order.Payment.Currency, &order.Payment.Provider,
			&order.Payment.Amount, &order.Payment.PaymentDt, &order.Payment.Bank,
			&order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee,
			&order.Payment.Transaction,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании строки заказа: %w", err)
		}

		// 2. Для каждого заказа список товаров
		itemsRows, err := r.db.Query(`
            SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
            FROM items WHERE order_uid = $1
        `, order.OrderUID)
		if err != nil {
			return nil, fmt.Errorf("ошибка при получении списка товаров для заказа %s: %w", order.OrderUID, err)
		}

		var items []model.Item
		for itemsRows.Next() {
			var item model.Item
			err = itemsRows.Scan(
				&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid,
				&item.Name, &item.Sale, &item.Size, &item.TotalPrice,
				&item.NmID, &item.Brand, &item.Status,
			)
			if err != nil {
				itemsRows.Close() // Закрыть текущий результат перед ошибкой
				return nil, fmt.Errorf("ошибка при сканировании строки товара для заказа %s: %w", order.OrderUID, err)
			}
			items = append(items, item)
		}
		itemsRows.Close() // Закрыть результат товаров для текущего заказа

		order.Items = items
		allOrders = append(allOrders, order)
	}

	// Проверяем, были ли ошибки во время итерации rows.Next()
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам заказов: %w", err)
	}

	log.Printf("Загружено %d заказов из БД", len(allOrders))
	return allOrders, nil
}
