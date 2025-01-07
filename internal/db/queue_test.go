package db

import (
	"gitlab.crja72.ru/gospec/go8/payment/internal/models"
	"testing"
)

func TestLockFreeQueue_EnqueueDequeue(t *testing.T) {
	queue := NewPaymentsQueue()
	payment := models.Payment{
		ID:     "1234",
		Amount: 100.0,
	}
	queue.Enqueue(payment)
	dequeuedPayment, ok := queue.Dequeue()
	if !ok {
		t.Errorf("Dequeue returned false, expected true")
	}
	if dequeuedPayment.ID != payment.ID {
		t.Errorf("Expected payment ID %s, but got %s", payment.ID, dequeuedPayment.ID)
	}
	if dequeuedPayment.Amount != payment.Amount {
		t.Errorf("Expected payment amount %f, but got %f", payment.Amount, dequeuedPayment.Amount)
	}
}

func TestLockFreeQueue_EnqueueListDequeue(t *testing.T) {
	queue := NewPaymentsQueue()
	payments := []models.Payment{
		{ID: "1234", Amount: 100.0},
		{ID: "5678", Amount: 200.0},
		{ID: "9101", Amount: 300.0},
	}
	queue.EnqueueList(payments)
	for _, payment := range payments {
		dequeuedPayment, ok := queue.Dequeue()
		if !ok {
			t.Errorf("Dequeue returned false, expected true")
		}
		if dequeuedPayment.ID != payment.ID {
			t.Errorf("Expected payment ID %s, but got %s", payment.ID, dequeuedPayment.ID)
		}
		if dequeuedPayment.Amount != payment.Amount {
			t.Errorf("Expected payment amount %f, but got %f", payment.Amount, dequeuedPayment.Amount)
		}
	}
}

func TestLockFreeQueue_DequeueFromEmptyQueue(t *testing.T) {
	queue := NewPaymentsQueue()
	dequeuedPayment, ok := queue.Dequeue()
	if ok {
		t.Errorf("Dequeue returned true, expected false when queue is empty")
	}
	if dequeuedPayment != (models.Payment{}) {
		t.Errorf("Expected empty payment, but got %+v", dequeuedPayment)
	}
}
