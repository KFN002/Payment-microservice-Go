package service

import (
	"context"
	"fmt"
	"gitlab.crja72.ru/gospec/go8/payment/internal/clients"
	"gitlab.crja72.ru/gospec/go8/payment/internal/db"
	"gitlab.crja72.ru/gospec/go8/payment/internal/models"
	"gitlab.crja72.ru/gospec/go8/payment/internal/repository"
	"go.uber.org/zap"
)

// PaymentService структура для сервиса
type PaymentService struct {
	repo          repository.PaymentRepository
	logger        *zap.Logger
	converter     *clients.ForexClient
	paymentClient *clients.YooMoneyClient
	paymentsQueue *db.LockFreeQueue
}

// NewPaymentService создание экземпляра сервиса
func NewPaymentService(repo repository.PaymentRepository, logger *zap.Logger, converter *clients.ForexClient, paymentClient *clients.YooMoneyClient, paymentsQueue *db.LockFreeQueue) *PaymentService {
	return &PaymentService{
		repo:          repo,
		logger:        logger,
		converter:     converter,
		paymentClient: paymentClient,
		paymentsQueue: paymentsQueue,
	}
}

// GetPaymentLink создание ссылки для оплаты
func (s *PaymentService) GetPaymentLink(ctx context.Context, paymentID string) (string, error) {
	s.logger.Info("Getting payment link", zap.String("payment_id", paymentID))

	amount, currency, err := s.repo.GetPaymentDetails(ctx, paymentID) // получаем данные по оплате
	if err != nil {
		return "", fmt.Errorf("failed to get payment details: %w", err)
	}

	convertedAmount, err := s.converter.ConvertToRub(amount, currency) // конвертируем сумму в рубли
	if err != nil {
		return "", fmt.Errorf("failed to convert amount: %w", err)
	}

	err = s.repo.UpdatePaymentStatus(ctx, paymentID, models.StatusPending) // если создали ссылку, то скоро оплата, меняем статус
	if err != nil {
		return "", fmt.Errorf("error changing payment status to pending: %w", err)
	}

	payment, err := s.repo.GetPaymentByID(ctx, paymentID) // получаем данные платежа
	if err != nil {
		s.logger.Error("Failed to fetch payment by ID", zap.String("payment_id", paymentID), zap.Error(err))
		return "", fmt.Errorf("error fetching payment: %w", err)
	}
	// создаем ссылку оплаты на основной счет банка
	link, err := s.paymentClient.QuickPayment(models.CoreAccount, paymentID, "AC", convertedAmount, paymentID, paymentID, paymentID, "")
	if err != nil {
		s.logger.Error("Failed to create payment link", zap.String("payment_id", paymentID), zap.Error(err))
		return "", fmt.Errorf("error creating payment link: %w", err)
	}

	s.paymentsQueue.Enqueue(*payment) // добавляем платеж в очередь для проверки статуса и избежания проблем с оплатой

	return link, nil
}

// GetPayment получение оплаты (статуса)
func (s *PaymentService) GetPayment(ctx context.Context, paymentID string) (string, error) {
	s.logger.Info("Getting payment", zap.String("payment_id", paymentID))

	status, err := s.paymentClient.CheckPaymentStatus(paymentID) // проверка статуса оплаты
	if err != nil {
		s.logger.Error("Failed to check payment status", zap.String("payment_id", paymentID), zap.Error(err))
		return "error", fmt.Errorf("error getting payment status: %w", err)
	}

	switch status {
	case "success": // все хорошо и деньги получены
		payment, err := s.repo.GetPaymentByID(ctx, paymentID)
		if err != nil {
			return "error", fmt.Errorf("error fetching payment: %w", err)
		}
		if payment.Status != "COMPLETE" { // деньги получены и счет не закрыт
			err := s.repo.UpdatePaymentStatus(ctx, paymentID, models.StatusSuccess)
			if err != nil {
				return "error", fmt.Errorf("error changing payment status to success: %w", err)
			}
		} else {
			return "complete", nil // деньги получены и счет закрыт
		}
	case "pending": // деньги не получены, но ждет оплаты
		err := s.repo.UpdatePaymentStatus(ctx, paymentID, models.StatusPending)
		if err != nil {
			return "pending", nil
		}
	case "failed": // ошибка
		err := s.repo.UpdatePaymentStatus(ctx, paymentID, models.StatusFailed)
		if err != nil {
			return "error", fmt.Errorf("error changing payment status to failed: %w", err)
		}
	}

	s.logger.Info("GetPayment: ", zap.String("payment_status", status))
	return status, nil
}

// CreatePayment создание счета оплаты
func (s *PaymentService) CreatePayment(ctx context.Context, fromUserID, toUserID string, amount float64, currency string) (string, error) {
	s.logger.Info("Creating payment", zap.String("user_id", fromUserID), zap.Float64("amount", amount), zap.String("currency", currency))

	paymentID, err := s.repo.CreatePayment(ctx, fromUserID, toUserID, amount, currency)
	if err != nil {
		s.logger.Error("Failed to create payment", zap.Error(err))
		return "", err
	}

	s.logger.Info("Payment created successfully", zap.String("payment_id", paymentID))
	return paymentID, nil
}

// RefundPayment возврат средств
func (s *PaymentService) RefundPayment(ctx context.Context, paymentID string) error {
	s.logger.Info("Refunding payment", zap.String("payment_id", paymentID))

	payment, err := s.repo.GetPaymentByID(ctx, paymentID) // получение данных о счете
	if err != nil {
		s.logger.Error("Failed to get payment by ID", zap.String("payment_id", paymentID), zap.Error(err))
		return fmt.Errorf("error fetching payment by ID: %w", err)
	}

	if payment.Status != "COMPLETE" {
		err = s.repo.UpdatePaymentStatus(ctx, paymentID, "REFUNDED") // обновление статуса счета
		if err != nil {
			return fmt.Errorf("error updating payment status: %w", err)
		}
		// создание счета с инвертированными получателями, чтобы счет банка не терял денег и все было легально
		newPaymentID, err := s.repo.CreatePayment(ctx, payment.ToUserID, payment.FromUserID, payment.Amount, payment.Currency)
		if err != nil {
			s.logger.Error("Failed to create new payment", zap.String("payment_id", newPaymentID), zap.Error(err))
			return fmt.Errorf("error creating payment: %w", err)
		}

		s.logger.Info("Payment refund process initiated successfully", zap.String("new_payment_id", newPaymentID))
		return nil
	}
	return fmt.Errorf("payment has not been paid")
}

// GetPaymentByID получение данных о счете по номеру
func (s *PaymentService) GetPaymentByID(ctx context.Context, paymentID string) (*models.Payment, error) {
	s.logger.Info("Getting payment by ID", zap.String("payment_id", paymentID))

	payment, err := s.repo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		s.logger.Error("Failed to get payment", zap.String("payment_id", paymentID), zap.Error(err))
		return nil, err
	}

	s.logger.Info("Payment retrieved", zap.String("payment_id", paymentID))
	return payment, nil
}

// GetPaymentHistory получение истории счетов пользователя
func (s *PaymentService) GetPaymentHistory(ctx context.Context, userID string, page, limit int) ([]*models.Payment, error) {
	s.logger.Info("Getting payment history", zap.String("user_id", userID), zap.Int("page", page), zap.Int("limit", limit))

	payments, err := s.repo.GetPaymentHistory(ctx, userID, page, limit)
	if err != nil {
		s.logger.Error("Failed to get payment history", zap.String("user_id", userID), zap.Error(err))
		return nil, err
	}

	s.logger.Info("Payment history retrieved", zap.String("user_id", userID))
	return payments, nil
}

// UpdatePaymentStatus обновление статуса счета
func (s *PaymentService) UpdatePaymentStatus(ctx context.Context, paymentID string, status models.PaymentStatus) error {
	s.logger.Info("Updating payment status", zap.String("payment_id", paymentID), zap.String("status", string(status)))

	err := s.repo.UpdatePaymentStatus(ctx, paymentID, status)
	if err != nil {
		s.logger.Error("Failed to update payment status", zap.String("payment_id", paymentID), zap.String("status", string(status)), zap.Error(err))
		return err
	}

	s.logger.Info("Payment status updated successfully", zap.String("payment_id", paymentID), zap.String("status", string(status)))
	return nil
}

// GetActivePayments Получение активных счетов пользователя
func (s *PaymentService) GetActivePayments(ctx context.Context, userID string) ([]*models.Payment, error) {
	s.logger.Info("Getting active payments", zap.String("user_id", userID))

	activePayments, err := s.repo.GetActivePayments(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active payments: %w", err)
	}

	return activePayments, nil
}
