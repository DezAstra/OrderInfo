// main.go

package main

import (
	"OrderInfo/config"
	"OrderInfo/internal/service"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("Запуск Order Display Service...")

	// 1. Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// 2. Создание OrderService
	orderService, err := service.NewOrderService(cfg)
	if err != nil {
		log.Fatalf("Ошибка инициализации OrderService: %v", err)
	}

	// 3. Запуск OrderService
	if err := orderService.Start(); err != nil {
		log.Fatalf("Ошибка запуска OrderService: %v", err)
	}

	// 4. Настройка обработки сигналов для корректного завершения
	// Создаём канал для получения сигналов
	sigChan := make(chan os.Signal, 1)
	// Notify заполняет канал sigChan при получении указанных сигналов
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Ждём сигнал
	sig := <-sigChan
	log.Printf("Получен сигнал: %s. Начинаю завершение работы...", sig)

	// 5. Корректная остановка OrderService
	// Создаём контекст для Stop (можно добавить таймаут)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := orderService.Stop(ctx); err != nil {
		log.Printf("Ошибки при остановке OrderService: %v", err)
		// Здесь можно решить, стоит ли os.Exit(1) в случае ошибки остановки
		// В данном случае просто логируем.
	} else {
		log.Println("OrderService корректно остановлен.")
	}

	log.Println("Order Display Service завершён.")
}
