CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR(255) PRIMARY KEY,
    track_number VARCHAR(255),
    entry VARCHAR(50),
    locale VARCHAR(50),
    internal_signature TEXT,
    customer_id VARCHAR(255),
    delivery_service VARCHAR(255),
    shardkey VARCHAR(50),
    sm_id INTEGER,
    date_created TIMESTAMP WITH TIME ZONE,
    oof_shard VARCHAR(50)
);

-- Таблица для информации о доставке
-- order_uid первичный ключ
CREATE TABLE IF NOT EXISTS delivery (
    uid VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE, -- Ссылка на заказ
    name VARCHAR(255),
    phone VARCHAR(50),
    zip VARCHAR(50),
    city VARCHAR(255),
    address TEXT,
    region VARCHAR(255),
    email VARCHAR(255)
);

-- таблица для информации о платеже
-- transaction первичный ключ
CREATE TABLE IF NOT EXISTS payment (
    transaction VARCHAR(255) PRIMARY KEY, -- ID транзакции
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

-- таблица для основной информации о заказе
-- order_uid первичный ключ
-- Связь с delivery через delivery_uid
-- Связь с payment через payment_transaction

    
    delivery_uid VARCHAR(255) UNIQUE REFERENCES delivery(uid) ON DELETE SET NULL, -- Уникальная ссылка
    payment_transaction VARCHAR(255) UNIQUE REFERENCES payment(transaction) ON DELETE SET NULL -- Уникальная ссылка


-- Таблица для товаров в заказе
CREATE TABLE IF NOT EXISTS items (
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

