// Package models
package models

import "time"

type Payment struct {
	CorrelationID string    `json:"correlationId"`
	AmountCents   int64     `json:"-"`
	RequestedAt   time.Time `json:"requestedAt"`
	ProcessorType string    `json:"processorType"`
}

type PaymentSummary struct {
	Default  ProcessorSummary `json:"default"`
	Fallback ProcessorSummary `json:"fallback"`
}

type ProcessorSummary struct {
	TotalRequests    int   `json:"totalRequests"`
	TotalAmountCents int64 `json:"-"`
}

type ProcessorRequest struct {
	CorrelationID string  `json:"correlationId"`
	Amount        float64 `json:"amount"`
	RequestedAt   string  `json:"requestedAt"`
}

type ProcessorResponse struct {
	Message string `json:"message"`
}

type HealthCheckResponse struct {
	Failing         bool `json:"failing"`
	MinResponseTime int  `json:"minResponseTime"`
}
