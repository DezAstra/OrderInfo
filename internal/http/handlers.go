package http

import (
	"OrderInfo/internal/cache"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// GetOrderHandler возвращает http.HandlerFunc для получения заказа по UID.
// Это фабричная функция, которая замыкает (closes over) зависимость cache.
func GetOrderHandler(cache cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем
		if r.Method != http.MethodGet {
			http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
			return
		}

		// ServeMux не предоставляет встроенной маршрутизации с переменными, поэтому извлекаем вручную.
		path := r.URL.Path
		prefix := "/order/"
		if !strings.HasPrefix(path, prefix) {
			http.Error(w, "Неправильный путь", http.StatusNotFound)
			return
		}
		uid := strings.TrimPrefix(path, prefix)

		// Проверка, не пустой ли UID
		if uid == "" {
			http.Error(w, "UID заказа отсутствует в URL", http.StatusBadRequest)
			return
		}

		log.Printf("Получен запрос на получение заказа с UID: %s", uid)

		// Получение заказа из кэша
		order, found := cache.Get(uid)

		if !found {
			log.Printf("Заказ с UID %s не найден в кэше", uid)
			http.Error(w, "Заказ не найден", http.StatusNotFound)
			return
		}

		// заголовок Content-Type
		w.Header().Set("Content-Type", "application/json")

		// Сериализация заказа в JSON и отправляем клиенту
		if err := json.NewEncoder(w).Encode(order); err != nil {
			log.Printf("Ошибка при сериализации заказа %s в JSON: %v", uid, err)
			return
		}

		log.Printf("Заказ %s успешно отправлен клиенту", uid)
	}
}
