package http

import (
	"OrderInfo/config"
	"OrderInfo/internal/cache"
	"context"
	"log"
	"net/http"
	"time"
)

// Server инкапсулирует HTTP-сервер и его зависимости.
type Server struct {
	httpServer *http.Server
	cache      cache.Cache
}

// NewServer создает новый экземпляр HTTP-сервера.
func NewServer(cfg *config.ServerConfig, cache cache.Cache) *Server {
	// Создание маршрутизатора
	mux := http.NewServeMux()

	// обработчик для получения заказа по UID
	getOrderHandler := GetOrderHandler(cache)
	mux.HandleFunc("/order/", getOrderHandler) // Обработчик будет вызван для /order/{uid}

	// обработчик для статических файлов
	fs := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs)) // <-- Маппим на /static/

	//  обработчик для корневого URL (/)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Проверка, что запрашивается корень /
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "./static/index.html") // index.html в static
	})

	// Создаем HTTP-сервер
	httpServer := &http.Server{
		Addr:         cfg.HTTPPort,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		cache:      cache,
	}
}

// Start запускает HTTP-сервер в отдельной горутине.
func (s *Server) Start() error {
	log.Printf("Запуск HTTP-сервера на %s", s.httpServer.Addr)
	// http.ListenAndServe блокирует выполнение, поэтому запускаем в горутине
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Ошибка при запуске HTTP-сервера: %v", err)
		}
	}()
	return nil
}

// Stop останавливает HTTP-сервер.
func (s *Server) Stop(ctx context.Context) error {
	log.Println("Остановка HTTP-сервера...")
	return s.httpServer.Shutdown(ctx)
}
