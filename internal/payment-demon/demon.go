package payments_demon

import (
	"context"
	"gitlab.crja72.ru/gospec/go8/payment/internal/clients"
	"gitlab.crja72.ru/gospec/go8/payment/internal/db"
	"gitlab.crja72.ru/gospec/go8/payment/internal/models"
	"gitlab.crja72.ru/gospec/go8/payment/internal/repository"
	"gitlab.crja72.ru/gospec/go8/payment/internal/service"
	"go.uber.org/zap"
	"log"
	"time"
)

// PaymentDemon Структура демона для проверки счетов
type PaymentDemon struct {
	service       service.PaymentService
	repo          repository.PaymentRepository
	paymentClient *clients.YooMoneyClient
	paymentsQueue *db.LockFreeQueue
	authClient    *clients.AuthClient
	logger        *zap.Logger
}

// NewPaymentDemon Создание экземпляра демона
func NewPaymentDemon(service service.PaymentService, repo repository.PaymentRepository, paymentClient *clients.YooMoneyClient, paymentQueue *db.LockFreeQueue, logger *zap.Logger, authClient *clients.AuthClient) *PaymentDemon {
	return &PaymentDemon{
		service:       service,
		paymentClient: paymentClient,
		paymentsQueue: paymentQueue,
		logger:        logger,
		authClient:    authClient,
	}
}

// Start Бесконечный цикл проверки счетов
func (d PaymentDemon) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Payment demon stopped")
			return
		default:
			payment, ok := d.paymentsQueue.Dequeue()
			if !ok {
				time.Sleep(1 * time.Second)
				continue
			}

			status, err := d.service.GetPayment(ctx, payment.ID) // получаем данные по статусу оплаты
			if err != nil {
				d.paymentsQueue.Enqueue(payment) // если ошибка, то добавляем в очередь снова
				d.logger.Error("Failed to check payment status", zap.String("payment_id", payment.ID), zap.Error(err))
				continue
			}

			switch status {
			case "success":
				receiverData, err := d.authClient.GetUserById(ctx, payment.ToUserID) // если успешно, запрашиваем счет для перевода средств
				if err != nil {
					d.paymentsQueue.Enqueue(payment) // если ошибка, то добавляем в очередь снова
					d.logger.Error("Failed to get receiver", zap.String("user_id", payment.ToUserID), zap.Error(err))
				}

				receiver := receiverData.YoomoneyId // получаем идентификатор получателя средств

				err = d.repo.UpdatePaymentStatus(ctx, payment.ID, models.StatusComplete)
				if err != nil {
					d.paymentsQueue.Enqueue(payment) // если ошибка, то добавляем в очередь снова
					d.logger.Error("Failed to update payment status", zap.String("payment_id", payment.ID), zap.Error(err))
				}

				paymentStatus, err := d.paymentClient.CreateTransfer(&payment, receiver)
				if err != nil { // если ошибка перевода, то возвращаем статус платежа на success
					err = d.repo.UpdatePaymentStatus(ctx, payment.ID, models.StatusSuccess)
					if err != nil {
						d.logger.Error("Failed to update payment status", zap.String("payment_id", payment.ID), zap.Error(err))
					}
					d.paymentsQueue.Enqueue(payment) // если ошибка, то добавляем в очередь снова
					d.logger.Error("Failed to create new transfer", zap.String("original_payment_id", payment.ID), zap.Error(err))
				} else {
					d.logger.Info("New transfer created", zap.String("status", paymentStatus))
				}

			case "pending", "failed":
				d.paymentsQueue.Enqueue(payment) // если ошибка или статус pending, то добавляем в очередь снова
				d.logger.Info("Re-enqueued payment for further processing", zap.String("payment_id", payment.ID))
			case "complete":
				d.logger.Info("Payment complete", zap.String("payment_id", payment.ID))
			default:
				d.paymentsQueue.Enqueue(payment) // если ошибка, то добавляем в очередь снова
				d.logger.Warn("Unexpected payment status", zap.String("payment_id", payment.ID), zap.String("status", status))
			}
		}
	}
}
