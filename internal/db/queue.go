package db

import (
	"gitlab.crja72.ru/gospec/go8/payment/internal/models"
	"sync/atomic"
	"unsafe"
)

// Queue асинхронная очередь
type Queue interface {
	Enqueue(element models.Payment)
	EnqueueList(data []models.Payment)
	Dequeue() (models.Payment, bool)
}

type QueueNode struct {
	expression models.Payment
	next       unsafe.Pointer
}

type LockFreeQueue struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}

func NewPaymentsQueue() *LockFreeQueue {
	dummy := &QueueNode{}
	return &LockFreeQueue{
		head: unsafe.Pointer(dummy),
		tail: unsafe.Pointer(dummy),
	}
}

func (q *LockFreeQueue) Enqueue(element models.Payment) {
	newNode := &QueueNode{expression: element}

	for {
		tail := atomic.LoadPointer(&q.tail)
		next := atomic.LoadPointer(&((*QueueNode)(tail)).next)

		if tail == atomic.LoadPointer(&q.tail) {
			if next == nil {
				if atomic.CompareAndSwapPointer(&((*QueueNode)(tail)).next, nil, unsafe.Pointer(newNode)) {
					atomic.CompareAndSwapPointer(&q.tail, tail, unsafe.Pointer(newNode))
					return
				}
			} else {
				atomic.CompareAndSwapPointer(&q.tail, tail, next)
			}
		}
	}
}

func (q *LockFreeQueue) EnqueueList(data []models.Payment) {
	for _, expr := range data {
		q.Enqueue(expr)
	}
}

func (q *LockFreeQueue) Dequeue() (models.Payment, bool) {
	for {
		head := atomic.LoadPointer(&q.head)
		next := atomic.LoadPointer(&((*QueueNode)(head)).next)

		if head == atomic.LoadPointer(&q.head) {
			if next == nil {
				return models.Payment{}, false
			}
			if atomic.CompareAndSwapPointer(&q.head, head, next) {
				return (*QueueNode)(next).expression, true
			}
		}
	}
}
