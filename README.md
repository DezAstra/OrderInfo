# OrderInfo
OrderInfo Service - это микросервис на Go, предназначенный для получения заказов через NATS Streaming, их хранения в PostgreSQL и предоставления доступа к ним через HTTP REST API. Также реализовано кэширование в памяти для быстрого доступа к последним заказам.

## Функциональность

*   **Подписка на NATS Streaming:** Сервис подписывается на указанный канал и получает сообщения с данными заказов в формате JSON.
*   **Хранение в PostgreSQL:** Полученные заказы сохраняются в реляционной БД PostgreSQL.
*   **In-Memory кэширование:** Заказы кэшируются в памяти для быстрого доступа.
*   **Восстановление кэша:** При запуске сервис восстанавливает кэш из PostgreSQL.
*   **HTTP API:** Предоставляет эндпоинт `GET /order/{uid}` для получения информации о заказе по его уникальному идентификатору из кэша.
*   **Простой веб-интерфейс:** Веб-страница по корневому пути `/` для удобного просмотра заказов.
*   **Автотесты:** Проект покрыт модульными и интеграционными тестами.

## Зависимости

*   **Golang** 
*   **PostgreSQL 18** (через Docker)
*   **NATS Streaming Server** (через Docker)
*   **Docker & Docker Compose** (для запуска зависимостей)

## Установка и запуск

### 1. Клонирование репозитория

```bash
git clone https://github.com/DezAstra/OrderInfo.git
cd OrderInfo
```

### 2. Запуск зависимостей (PostgreSQL, NATS Streaming)

Сервис использует Docker Compose для управления зависимостями.

```bash
docker-compose up -d
```

### 3. (Опционально) Создание таблиц в БД

Если таблицы ещё не созданы, выполните SQL-скрипт `scripts/db.sql` в БД `order_db` внутри контейнера `order-postgres` или через `psql`.

### 4. Запуск сервиса

Убедитесь, что `go.mod` и зависимости актуальны:

```bash
go mod tidy
```

Запустите сервис:

```bash
go run main.go
```

Сервис будет доступен по адресу `http://localhost:8080`.

## Использование

### 1. Веб-интерфейс

Откройте в браузере `http://localhost:8080/`. Вы увидите форму для ввода `order_uid`.

### 2. HTTP API

Получение информации о заказе по его `order_uid`:

```bash
curl http://localhost:8080/order/b563feb7b2b84b6test
```

### 3. Отправка тестовых данных в NATS Streaming

Для тестирования подписки на NATS Streaming можно использовать скрипт `scripts/publish_testdata.go`:

```bash
cd scripts
go run publish_testdata.go
```

Это отправит содержимое `model.json` в канал `orders` NATS Streaming. Сервис должен получить сообщение, сохранить заказ в БД и обновить кэш.

## Тестирование

Проект покрыт автотестами. Для запуска тестов убедитесь, что Docker Compose с зависимостями запущен (`docker-compose up -d`).

### Запуск всех тестов

```bash
go test -v ./...
```


### Запуск стресс-тестов (wrk, vegeta)

Для запуска стресс-тестов запустите соответствующий bat-файл.
```bash
scripts\wrk_run.bat
```
```bash
scripts\vegeta_run.bat
```

## Структура проекта

```
OrderInfo/
├── config/                  # Конфигурация приложения
├── internal/                # Внутренние пакеты приложения
│   ├── cache/               # In-memory кэш
│   │   ├── cache.go
│   │   └── cache_test.go
│   ├── db/                  # Подключение к БД
│   │   ├── connection.go
│   │   └── connection_test.go
│   ├── http/                # HTTP-сервер и обработчики
│   │   ├── handlers.go
│   │   ├── handlers_test.go
│   │   └── server.go
│   ├── model/               # Модель данных
│   │   ├── order.go
│   │   └── model.json
│   ├── nats/                # Клиент и обработчик NATS Streaming
│   │   ├── client.go
│   │   ├── client_test.go
│   │   ├── handler.go
│   │   └── handler_test.go
│   ├── repository/          # Репозиторий для работы с БД
│   │   ├── repository.go
│   │   └── repository_test.go
│   └── service/             # Основная бизнес-логика сервиса
│       ├── order_service.go
│       └── order_service_test.go
├── scripts/                 # Вспомогательные скрипты
│   ├── publish_testdata.go  # Скрипт для публикации тестовых данных в NATS
│   ├── sdb.sql              # SQL-скрипт для создания таблиц
│   ├── wrk_run.bat          # Скрипт для запуска wrk стресс-теста
│   ├── vegeta_run.bat       # Скрипт для запуска vegeta стресс-теста
│   └── model.json           # Тестовые данные в формате JSON
├── static/                  # Статические файлы для веб-интерфейса
│   ├── index.html
│   ├── style.css
│   └── script.js
├── go.mod
├── go.sum
├── DockerFile
├── main.go                  # Точка входа в приложение
├── docker-compose.yml       # Файл для запуска зависимостей через Docker Compose
└── README.md
```

## Демонстрация работы (видео)


https://github.com/user-attachments/assets/dc45cd2a-17aa-4537-a900-17e0f6666fb4


## Лицензия

Этот проект лицензирован под лицензией MIT 
