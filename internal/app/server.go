package app

import (
	"log"

	"github.com/danielscoffee/rinha-backend2025-go/internal/app/handlers"
	"github.com/danielscoffee/rinha-backend2025-go/internal/pkg/cache"
	"github.com/danielscoffee/rinha-backend2025-go/internal/pkg/processor"
	"github.com/danielscoffee/rinha-backend2025-go/internal/pkg/storage"
	"github.com/valyala/fasthttp"
)

type Server struct {
	router    *Router
	storage   *storage.MemoryStorage
	processor *processor.PaymentProcessor
}

func NewServer() *Server {
	storage := storage.NewMemoryStorage()
	redisCache := cache.NewRedisCache()
	processor := processor.NewPaymentProcessor(redisCache)
	
	h := handlers.NewHandlers(storage, processor)
	router := NewRouter(h)

	return &Server{
		router:    router,
		storage:   storage,
		processor: processor,
	}
}

func (s *Server) Listen(addr string) error {
	log.Printf("Server listening on %s", addr)
	return fasthttp.ListenAndServe(addr, s.router.Handler)
}
