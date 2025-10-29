package main

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/nats-io/stan.go"
)

func main() {
	clusterID := "test-cluster"         // config.go
	clientID := "publisher_test_script" // ClientID для паблишера
	natsURL := "nats://localhost:4223"  // config.go

	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))
	if err != nil {
		log.Fatalf("Ошибка подключения к NATS Streaming: %v", err)
	}
	defer sc.Close()

	channel := "orders" // config.go

	// Чтение JSON из файла
	jsonFilePath := "scripts/model.json"

	file, err := os.Open(jsonFilePath)
	if err != nil {
		log.Fatalf("Ошибка открытия файла %s: %v", jsonFilePath, err)
	}
	defer file.Close() // Закрываем файл после завершения main

	orderJSONBytes, err := io.ReadAll(file) // чтение содержимое файла в []byte
	if err != nil {
		log.Fatalf("Ошибка чтения файла %s: %v", jsonFilePath, err)
	}

	orderUID := "b563feb7b2b84b6test"

	err = sc.Publish(channel, orderJSONBytes) // Публикуем []byte
	if err != nil {
		log.Fatalf("Ошибка публикации сообщения: %v", err)
	}

	log.Printf("Сообщение с заказом (UID: %s) отправлено в канал %s", orderUID, channel)

	// Небольшая задержка, чтобы убедиться, что сообщение обработано
	time.Sleep(1 * time.Second)
}
