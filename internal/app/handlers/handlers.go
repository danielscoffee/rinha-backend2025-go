// Package handlers
package handlers

import (
	"fmt"
	"math"
	"time"

	"github.com/danielscoffee/rinha-backend2025-go/internal/pkg/models"
	"github.com/danielscoffee/rinha-backend2025-go/internal/pkg/processor"
	"github.com/danielscoffee/rinha-backend2025-go/internal/pkg/storage"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

type Handlers struct {
	storage   *storage.MemoryStorage
	processor *processor.PaymentProcessor
	parser    *fastjson.Parser
}

func NewHandlers(storage *storage.MemoryStorage, processor *processor.PaymentProcessor) *Handlers {
	return &Handlers{
		storage:   storage,
		processor: processor,
		parser:    &fastjson.Parser{},
	}
}

func (h *Handlers) PostPayments(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")

	body := ctx.PostBody()
	if len(body) == 0 {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(`{"error":"empty body"}`)
		return
	}

	v, err := h.parser.ParseBytes(body)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(`{"error":"invalid json"}`)
		return
	}

	correlationID := string(v.GetStringBytes("correlationId"))
	if correlationID == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(`{"error":"correlationId is required"}`)
		return
	}

	amount := v.GetFloat64("amount")
	amountCents := int64(math.Round(amount * 100))
	if amountCents <= 0 {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(`{"error":"amount must be positive"}`)
		return
	}

	payment := &models.Payment{
		CorrelationID: correlationID,
		AmountCents:   amountCents,
		RequestedAt:   time.Now().UTC(),
	}

	go func() {
		h.processor.ProcessPayment(payment, h.storage)
	}()

	ctx.SetStatusCode(fasthttp.StatusAccepted)
	ctx.SetBodyString(`{"status":"accepted"}`)
}

func (h *Handlers) GetPaymentsSummary(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")

	args := ctx.QueryArgs()
	fromBytes := args.Peek("from")
	toBytes := args.Peek("to")

	var from, to *time.Time
	if len(fromBytes) > 0 {
		if t, err := time.Parse(time.RFC3339, string(fromBytes)); err == nil {
			from = &t
		}
	}
	if len(toBytes) > 0 {
		if t, err := time.Parse(time.RFC3339, string(toBytes)); err == nil {
			to = &t
		}
	}

	summary := h.storage.GetSummary(from, to)
	response := buildSummaryResponse(summary)
	ctx.SetBodyString(response)
}

func (h *Handlers) GetHealth(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString(`{"status":"healthy","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`)
}

func buildSummaryResponse(summary *models.PaymentSummary) string {
	return `{"default":{"totalRequests":` +
		fmt.Sprintf("%d", summary.Default.TotalRequests) +
		`,"totalAmount":` +
		processor.CentsToDecimalString(summary.Default.TotalAmountCents) +
		`},"fallback":{"totalRequests":` +
		fmt.Sprintf("%d", summary.Fallback.TotalRequests) +
		`,"totalAmount":` +
		processor.CentsToDecimalString(summary.Fallback.TotalAmountCents) +
		`}}`
}
