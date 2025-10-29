package cache

import (
	"OrderInfo/internal/model"
	"sync"
)

// Cache определяет интерфейс для кэширования заказов.
type Cache interface {
	Set(uid string, order model.Order)
	Get(uid string) (model.Order, bool)
	Delete(uid string)
}

// InMemoryCache реализует Cache с использованием map и мьютекса
type InMemoryCache struct {
	orders map[string]model.Order
	mutex  sync.RWMutex // RWMutex позволяет множественные чтения, но эксклюзивную запись
}

// NewInMemoryCache создает новый экземпляр InMemoryCache
func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		orders: make(map[string]model.Order),
	}
}

// Set добавляет или обновляет заказ в кэше
func (c *InMemoryCache) Set(uid string, order model.Order) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.orders[uid] = order
}

// Get возвращает заказ по его UID и флаг, указывающий, был ли он найден.
func (c *InMemoryCache) Get(uid string) (model.Order, bool) {
	c.mutex.RLock() // для безопасного параллельного чтения
	defer c.mutex.RUnlock()
	order, found := c.orders[uid]
	return order, found
}

// Delete удаляет заказ из кэша
func (c *InMemoryCache) Delete(uid string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.orders, uid)
}
