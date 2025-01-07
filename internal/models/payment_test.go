package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPaymentInitialization(t *testing.T) {
	now := time.Now()
	payment := Payment{
		ID:         "payment-id",
		FromUserID: "user1",
		ToUserID:   "user2",
		Amount:     100.50,
		Currency:   "RUB",
		Status:     StatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	assert.Equal(t, "payment-id", payment.ID)
	assert.Equal(t, "user1", payment.FromUserID)
	assert.Equal(t, "user2", payment.ToUserID)
	assert.Equal(t, 100.50, payment.Amount)
	assert.Equal(t, "RUB", payment.Currency)
	assert.Equal(t, StatusPending, payment.Status)
	assert.Equal(t, now, payment.CreatedAt)
	assert.Equal(t, now, payment.UpdatedAt)
}

func TestPaymentStatusConstants(t *testing.T) {
	assert.Equal(t, "PENDING", string(StatusPending))
	assert.Equal(t, "SUCCESS", string(StatusSuccess))
	assert.Equal(t, "FAILED", string(StatusFailed))
	assert.Equal(t, "REFUNDED", string(StatusRefunded))
	assert.Equal(t, "COMPLETE", string(StatusComplete))
}

func TestCoreAccountConstant(t *testing.T) {
	assert.Equal(t, "4100118177295897", CoreAccount)
}

func TestPaymentStatusTransitions(t *testing.T) {
	payment := Payment{
		ID:         "payment-id",
		FromUserID: "user1",
		ToUserID:   "user2",
		Amount:     150.0,
		Currency:   "USD",
		Status:     StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	payment.Status = StatusSuccess
	assert.Equal(t, StatusSuccess, payment.Status)

	payment.Status = StatusFailed
	assert.Equal(t, StatusFailed, payment.Status)

	payment.Status = StatusRefunded
	assert.Equal(t, StatusRefunded, payment.Status)

	payment.Status = StatusComplete
	assert.Equal(t, StatusComplete, payment.Status)
}

func TestPaymentEdgeCases(t *testing.T) {
	payment := Payment{
		ID:         "zero-amount-id",
		FromUserID: "user1",
		ToUserID:   "user2",
		Amount:     0.0,
		Currency:   "USD",
		Status:     StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	assert.Equal(t, 0.0, payment.Amount)

	payment.Amount = -50.0
	assert.Equal(t, -50.0, payment.Amount)

	payment.FromUserID = ""
	payment.ToUserID = ""
	assert.Equal(t, "", payment.FromUserID)
	assert.Equal(t, "", payment.ToUserID)
}
