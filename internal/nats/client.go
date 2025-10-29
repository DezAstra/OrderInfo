package nats

import (
	"OrderInfo/config"
	"fmt"
	"log"

	"github.com/nats-io/stan.go"
)

// Client инкапсулирует подключение к NATS Streaming.
type Client struct {
	sc stan.Conn
}

// NewClient создает новый клиент NATS Streaming.
func NewClient(cfg *config.NATSConfig) (*Client, error) {
	log.Printf("Подключение к NATS Streaming: URL=%s, ClusterID=%s, ClientID=%s", cfg.URL, cfg.ClusterID, cfg.ClientID)

	sc, err := stan.Connect(cfg.ClusterID, cfg.ClientID, stan.NatsURL(cfg.URL))
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к NATS Streaming: %w", err)
	}

	log.Println("Успешное подключение к NATS Streaming")
	return &Client{sc: sc}, nil
}

// Subscribe подписывается на канал с указанным обработчиком.
func (c *Client) Subscribe(channel string, handler stan.MsgHandler) (stan.Subscription, error) {
	log.Printf("Подписка на канал: %s", channel)
	sub, err := c.sc.Subscribe(channel, handler)
	if err != nil {
		return nil, fmt.Errorf("ошибка подписки на канал %s: %w", channel, err)
	}
	log.Printf("Успешно подписан на канал: %s", channel)
	return sub, nil
}

// Close закрывает соединение с NATS Streaming.
func (c *Client) Close() error {
	if c.sc != nil {
		return c.sc.Close()
	}
	return nil
}
