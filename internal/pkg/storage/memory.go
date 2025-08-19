// Package storage
package storage

import (
	"sync"
	"time"

	"github.com/danielscoffee/rinha-backend2025-go/internal/pkg/models"
)

type MemoryStorage struct {
	mutex           sync.RWMutex
	payments        []models.Payment
	defaultSummary  models.ProcessorSummary
	fallbackSummary models.ProcessorSummary
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		payments: make([]models.Payment, 0, 10000),
	}
}

func (s *MemoryStorage) StorePayment(payment *models.Payment) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.payments = append(s.payments, *payment)

	if payment.ProcessorType == "default" {
		s.defaultSummary.TotalRequests++
		s.defaultSummary.TotalAmountCents += payment.AmountCents
	} else {
		s.fallbackSummary.TotalRequests++
		s.fallbackSummary.TotalAmountCents += payment.AmountCents
	}
}

func (s *MemoryStorage) GetSummary(from, to *time.Time) *models.PaymentSummary {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if from == nil && to == nil {
		return &models.PaymentSummary{
			Default:  s.defaultSummary,
			Fallback: s.fallbackSummary,
		}
	}

	var defaultSummary, fallbackSummary models.ProcessorSummary
	for _, payment := range s.payments {
		if from != nil && payment.RequestedAt.Before(*from) {
			continue
		}
		if to != nil && payment.RequestedAt.After(*to) {
			continue
		}
		if payment.ProcessorType == "default" {
			defaultSummary.TotalRequests++
			defaultSummary.TotalAmountCents += payment.AmountCents
		} else {
			fallbackSummary.TotalRequests++
			fallbackSummary.TotalAmountCents += payment.AmountCents
		}
	}

	return &models.PaymentSummary{Default: defaultSummary, Fallback: fallbackSummary}
}
