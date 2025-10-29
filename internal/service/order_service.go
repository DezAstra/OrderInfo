package service

import (
	"OrderInfo/config"
	"OrderInfo/internal/cache"
	"OrderInfo/internal/db"
	"OrderInfo/internal/http"
	"OrderInfo/internal/nats"
	"OrderInfo/internal/repository"
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/nats-io/stan.go"
)

// OrderService
type OrderService struct {
	config     *config.Config
	db         *sql.DB
	repo       repository.OrderRepository
	cache      cache.Cache
	natsClient *nats.Client
	natsSub    stan.Subscription
	httpServer *http.Server
}

// NewOrderService creates a new instance of OrderService.
func NewOrderService(cfg *config.Config) (*OrderService, error) {
	log.Println("Инициализация OrderService...")

	// 1. Подключение к БД
	dbConn, err := db.NewConnection(&cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации подключения к БД: %w", err)
	}

	// 2. Создание репозитория
	repo := repository.NewPostgresOrderRepository(dbConn)

	// 3. Создание кэша
	cacheInstance := cache.NewInMemoryCache()

	// 4. Восстановление кэша из БД (требование задания)
	log.Println("Восстановление кэша из БД...")
	if err := initializeCacheFromDB(cacheInstance, repo); err != nil {
		log.Printf("Предупреждение: не удалось восстановить кэш из БД: %v", err)
	} else {
		log.Println("Кэш успешно восстановлен из БД.")
	}

	// 5. Инициализация клиента NATS
	natsClient, err := nats.NewClient(&cfg.NATS)
	if err != nil {
		// Закрываем подключение к БД, если NATS не удалось инициализировать
		dbConn.Close()
		return nil, fmt.Errorf("ошибка инициализации клиента NATS: %w", err)
	}

	// 6. Инициализация HTTP-сервера
	httpServer := http.NewServer(&cfg.Server, cacheInstance)

	service := &OrderService{
		config:     cfg,
		db:         dbConn,
		repo:       repo,
		cache:      cacheInstance,
		natsClient: natsClient,
		httpServer: httpServer,
	}

	return service, nil
}

// Start запускает компоненты сервиса: подписку на NATS и HTTP-сервер.
func (s *OrderService) Start() error {
	log.Println("Запуск OrderService...")

	// 1. Подписка на NATS
	msgHandler := nats.NewMessageHandler(s.repo, s.cache)
	sub, err := s.natsClient.Subscribe(s.config.NATS.Channel, msgHandler)
	if err != nil {
		return fmt.Errorf("ошибка подписки на канал NATS: %w", err)
	}
	s.natsSub = sub

	// 2. Запуск HTTP-сервера
	if err := s.httpServer.Start(); err != nil {
		// отменить подписку, если HTTP-сервер не запустился
		if s.natsSub != nil {
			s.natsSub.Close() // Close() доступен у stan.Subscription
		}
		return fmt.Errorf("ошибка запуска HTTP-сервера: %w", err)
	}

	log.Println("OrderService запущен.")
	return nil
}

// Stop останавливает компоненты сервиса.
func (s *OrderService) Stop(ctx context.Context) error {
	log.Println("Остановка OrderService...")
	var errs []error

	// Остановка HTTP-сервера
	if err := s.httpServer.Stop(ctx); err != nil {
		errs = append(errs, fmt.Errorf("ошибка остановки HTTP-сервера: %w", err))
	}

	// Закрытие подписки NATS
	if s.natsSub != nil { // s.natsSub теперь типа stan.Subscription
		if err := s.natsSub.Close(); err != nil { // Close() доступен у stan.Subscription
			errs = append(errs, fmt.Errorf("ошибка закрытия подписки NATS: %w", err))
		}
	}

	// Закрытие клиента NATS
	if err := s.natsClient.Close(); err != nil {
		errs = append(errs, fmt.Errorf("ошибка закрытия клиента NATS: %w", err))
	}

	// Закрытие подключения к БД
	if err := s.db.Close(); err != nil {
		errs = append(errs, fmt.Errorf("ошибка закрытия подключения к БД: %w", err))
	}

	if len(errs) > 0 {
		for _, e := range errs {
			log.Printf("Ошибка при остановке: %v", e)
		}
		// Возвращаем первую ошибку или nil, если нужно просто логировать
		return errs[0]
	}

	log.Println("OrderService остановлен.")
	return nil
}

// initializeCacheFromDB загружает все заказы из репозитория в кэш.
func initializeCacheFromDB(cacheInstance cache.Cache, repo repository.OrderRepository) error {
	allOrders, err := repo.GetAllOrders()
	if err != nil {
		return fmt.Errorf("ошибка получения всех заказов из БД: %w", err)
	}

	for _, order := range allOrders {
		cacheInstance.Set(order.OrderUID, order)
	}

	log.Printf("Загружено %d заказов в кэш из БД.", len(allOrders))
	return nil
}
