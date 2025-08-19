// Package processor
package processor

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/danielscoffee/rinha-backend2025-go/internal/pkg/cache"
	"github.com/danielscoffee/rinha-backend2025-go/internal/pkg/models"
	"github.com/danielscoffee/rinha-backend2025-go/internal/pkg/storage"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

const (
	StateClosed = iota
	StateOpen
	StateHalfOpen
)

var ErrCircuitOpen = fmt.Errorf("circuit breaker is open")

type CircuitBreaker struct {
	state        int32
	failures     int32
	lastFailTime int64
	maxFailures  int32
	timeout      time.Duration
}

type PaymentProcessor struct {
	httpClient       *fasthttp.Client
	defaultURL       string
	fallbackURL      string
	lastHealthCheck  map[string]time.Time
	healthCheckMutex sync.RWMutex
	parser           *fastjson.Parser
	defaultCB        *CircuitBreaker
	fallbackCB       *CircuitBreaker
	cache            *cache.RedisCache
}

func NewCircuitBreaker(maxFailures int32, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:       StateClosed,
		maxFailures: maxFailures,
		timeout:     timeout,
	}
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	state := atomic.LoadInt32(&cb.state)

	if state == StateOpen {
		if time.Since(time.Unix(0, atomic.LoadInt64(&cb.lastFailTime))) > cb.timeout {
			atomic.StoreInt32(&cb.state, StateHalfOpen)
		} else {
			return ErrCircuitOpen
		}
	}

	err := fn()
	if err != nil {
		atomic.AddInt32(&cb.failures, 1)
		atomic.StoreInt64(&cb.lastFailTime, time.Now().UnixNano())

		if atomic.LoadInt32(&cb.failures) >= cb.maxFailures {
			atomic.StoreInt32(&cb.state, StateOpen)
		}
		return err
	}

	atomic.StoreInt32(&cb.failures, 0)
	atomic.StoreInt32(&cb.state, StateClosed)
	return nil
}

func NewPaymentProcessor(c *cache.RedisCache) *PaymentProcessor {
	client := &fasthttp.Client{
		MaxConnsPerHost:     50,
		MaxIdleConnDuration: 30 * time.Second,
		ReadTimeout:         3 * time.Second,
		WriteTimeout:        3 * time.Second,
		MaxConnWaitTimeout:  1 * time.Second,
	}

	return &PaymentProcessor{
		httpClient:      client,
		defaultURL:      "http://payment-processor-default:8080",
		fallbackURL:     "http://payment-processor-fallback:8080",
		lastHealthCheck: make(map[string]time.Time),
		parser:          &fastjson.Parser{},
		defaultCB:       NewCircuitBreaker(5, 30*time.Second),
		fallbackCB:      NewCircuitBreaker(5, 30*time.Second),
		cache:           c,
	}
}

func (p *PaymentProcessor) ProcessPayment(payment *models.Payment, storage *storage.MemoryStorage) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	err := p.defaultCB.Call(func() error {
		return p.tryProcessor(ctx, payment, "default")
	})

	if err == nil {
		payment.ProcessorType = "default"
		storage.StorePayment(payment)
		return
	}

	err = p.fallbackCB.Call(func() error {
		return p.tryProcessor(ctx, payment, "fallback")
	})

	if err == nil {
		payment.ProcessorType = "fallback"
		storage.StorePayment(payment)
		return
	}

	log.Printf("Failed to process payment %s: %v", payment.CorrelationID, err)
}

func (p *PaymentProcessor) tryProcessor(ctx context.Context, payment *models.Payment, processorType string) error {
	url := p.defaultURL
	if processorType == "fallback" {
		url = p.fallbackURL
	}

	if !p.isHealthy(processorType) {
		return fmt.Errorf("processor %s unhealthy", processorType)
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url + "/payments")
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")

	body := buildPaymentRequest(payment)
	req.SetBodyString(body)

	errCh := make(chan error, 1)
	go func() {
		errCh <- p.httpClient.Do(req, resp)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
		if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
			return fmt.Errorf("processor returned status %d", resp.StatusCode())
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *PaymentProcessor) isHealthy(processorType string) bool {
	if p.cache != nil {
		key := "healthcheck:" + processorType
		if ok, _ := p.cache.SetNX(key, "1", 5*time.Second); !ok {
			return true
		}
	}

	p.healthCheckMutex.RLock()
	lastCheck, exists := p.lastHealthCheck[processorType]
	p.healthCheckMutex.RUnlock()

	if exists && time.Since(lastCheck) < 5*time.Second {
		return true
	}

	p.healthCheckMutex.Lock()
	p.lastHealthCheck[processorType] = time.Now()
	p.healthCheckMutex.Unlock()

	url := p.defaultURL
	if processorType == "fallback" {
		url = p.fallbackURL
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url + "/payments/service-health")
	req.Header.SetMethod("GET")

	err := p.httpClient.DoTimeout(req, resp, 1*time.Second)
	if err != nil || resp.StatusCode() != 200 {
		return false
	}

	v, err := p.parser.ParseBytes(resp.Body())
	if err != nil {
		return false
	}

	return !v.GetBool("failing")
}

func buildPaymentRequest(payment *models.Payment) string {
	var buf strings.Builder
	buf.Grow(256)
	buf.WriteString(`{"correlationId":"`)
	buf.WriteString(payment.CorrelationID)
	buf.WriteString(`","amount":`)
	buf.WriteString(CentsToDecimalString(payment.AmountCents))
	buf.WriteString(`,"requestedAt":"`)
	buf.WriteString(payment.RequestedAt.Format(time.RFC3339))
	buf.WriteString(`"}`)
	return buf.String()
}

func CentsToDecimalString(cents int64) string {
	whole := cents / 100
	frac := cents % 100

	res := ""
	res += fmt.Sprintf("%d", whole)
	if cents < 0 {
		res = "-"
	}
	if frac == 0 {
		return res
	}

	res += "."
	if frac < 10 {
		res += "0"
	}
	res += fmt.Sprintf("%d", frac)
	return res
}
