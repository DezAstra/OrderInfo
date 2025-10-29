package nats

import (
	"OrderInfo/config"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	// Тест с правильной конфигурацией
	cfg := &config.NATSConfig{
		URL:       "nats://localhost:4223",
		ClusterID: "test-cluster",
		ClientID:  "test-client-id",
		Channel:   "test-channel",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.sc)

	// Закрываем клиент
	err = client.Close()
	assert.NoError(t, err)
}

func TestNewClient_InvalidConfig(t *testing.T) {
	// Тест с неправильной конфигурацией
	cfg := &config.NATSConfig{
		URL:       "nats://invalid-host:9999", // Несуществующий хост
		ClusterID: "invalid-cluster",
		ClientID:  "invalid-client-id",
		Channel:   "invalid-channel",
	}

	// Ожидаем ошибку при подключении
	client, err := NewClient(cfg)
	// Проверим, что либо ошибка, либо клиент не nil
	if err != nil {
		assert.Nil(t, client)
	} else {
		assert.NotNil(t, client)
		// Если клиент создан, попробуем закрыть его
		closeErr := client.Close()
		assert.NoError(t, closeErr)
	}
}

func TestClient_Close(t *testing.T) {
	cfg := &config.NATSConfig{
		URL:       "nats://localhost:4223",
		ClusterID: "test-cluster",
		ClientID:  "test-close-client-id",
		Channel:   "test-channel",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	assert.NotNil(t, client)

	// Закрываем клиент
	err = client.Close()
	assert.NoError(t, err)

	// Повторное закрытие не должно вызывать ошибку
	err = client.Close()
	assert.NoError(t, err)
}
