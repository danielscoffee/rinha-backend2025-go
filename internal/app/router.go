// Package app
package app

import (
	"bytes"

	"github.com/danielscoffee/rinha-backend2025-go/internal/app/handlers"
	"github.com/valyala/fasthttp"
)

type Router struct {
	handlers *handlers.Handlers
}

func NewRouter(h *handlers.Handlers) *Router {
	return &Router{handlers: h}
}

func (r *Router) Handler(ctx *fasthttp.RequestCtx) {
	path := ctx.Path()
	method := ctx.Method()

	switch {
	case bytes.Equal(method, []byte("POST")) && bytes.Equal(path, []byte("/payments")):
		r.handlers.PostPayments(ctx)
	case bytes.Equal(method, []byte("GET")) && bytes.HasPrefix(path, []byte("/payments-summary")):
		r.handlers.GetPaymentsSummary(ctx)
	case bytes.Equal(method, []byte("GET")) && bytes.Equal(path, []byte("/health")):
		r.handlers.GetHealth(ctx)
	default:
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetBodyString(`{"error":"not found"}`)
	}
}
