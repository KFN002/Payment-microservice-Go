package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.crja72.ru/gospec/go8/payment/internal/models"
	"go.uber.org/zap"
)

type PaymentRepository interface {
	CreatePayment(ctx context.Context, fromUserID, toUserID string, amount float64, currency string) (string, error)
	GetPaymentByID(ctx context.Context, paymentID string) (*models.Payment, error)
	GetPaymentHistory(ctx context.Context, userID string, page, limit int) ([]*models.Payment, error)
	UpdatePaymentStatus(ctx context.Context, paymentID string, status models.PaymentStatus) error
	GetPaymentDetails(ctx context.Context, paymentID string) (float64, string, error)
	GetActivePayments(ctx context.Context, userID string) ([]*models.Payment, error)
}

type paymentRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
	redis  *redis.Client
}

func NewPaymentRepository(db *pgxpool.Pool, logger *zap.Logger, redis *redis.Client) PaymentRepository {
	return &paymentRepository{
		db:     db,
		logger: logger,
		redis:  redis,
	}
}

func (r *paymentRepository) GetPaymentByID(ctx context.Context, paymentID string) (*models.Payment, error) {
	cacheKey := fmt.Sprintf("payment:%s", paymentID)

	cachedPayment, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var payment models.Payment
		if err := json.Unmarshal([]byte(cachedPayment), &payment); err == nil {
			return &payment, nil
		}
	}

	query := `SELECT id, from_user_id, to_user_id, amount, currency, status, created_at, updated_at 
			  FROM payments WHERE id = $1`
	var payment models.Payment
	err = r.db.QueryRow(ctx, query, paymentID).Scan(
		&payment.ID,
		&payment.FromUserID,
		&payment.ToUserID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		r.logger.Error("Failed to fetch payment by ID", zap.String("payment_id", paymentID), zap.Error(err))
		return nil, fmt.Errorf("error fetching payment by ID: %w", err)
	}

	data, err := json.Marshal(payment)
	if err == nil {
		r.redis.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return &payment, nil
}

func (r *paymentRepository) GetPaymentHistory(ctx context.Context, userID string, page, limit int) ([]*models.Payment, error) {
	cacheKey := fmt.Sprintf("payment_history:%s:%d:%d", userID, page, limit)

	cachedHistory, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var payments []*models.Payment
		if err := json.Unmarshal([]byte(cachedHistory), &payments); err == nil {
			return payments, nil
		}
	}

	offset := (page - 1) * limit
	query := `SELECT id, from_user_id, to_user_id, amount, currency, status, created_at, updated_at 
			  FROM payments WHERE from_user_id = $1 LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		r.logger.Error("Failed to fetch payment history", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("error fetching payment history: %w", err)
	}
	defer rows.Close()

	var payments []*models.Payment
	for rows.Next() {
		var payment models.Payment
		if err := rows.Scan(
			&payment.ID,
			&payment.FromUserID,
			&payment.ToUserID,
			&payment.Amount,
			&payment.Currency,
			&payment.Status,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		); err != nil {
			r.logger.Error("Failed to scan payment history row", zap.Error(err))
			return nil, fmt.Errorf("error scanning payment history: %w", err)
		}
		payments = append(payments, &payment)
	}

	data, err := json.Marshal(payments)
	if err == nil {
		r.redis.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return payments, nil
}

func (r *paymentRepository) GetPaymentDetails(ctx context.Context, paymentID string) (float64, string, error) {
	cacheKey := fmt.Sprintf("payment_details:%s", paymentID)

	cachedDetails, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var details struct {
			Amount   float64 `json:"amount"`
			Currency string  `json:"currency"`
		}
		if err := json.Unmarshal([]byte(cachedDetails), &details); err == nil {
			return details.Amount, details.Currency, nil
		}
	}

	query := `SELECT amount, currency FROM payments WHERE id = $1`
	var amount float64
	var currency string
	err = r.db.QueryRow(ctx, query, paymentID).Scan(&amount, &currency)
	if err != nil {
		r.logger.Error("Failed to fetch payment details", zap.String("payment_id", paymentID), zap.Error(err))
		return 0, "", fmt.Errorf("error fetching payment details: %w", err)
	}

	details := struct {
		Amount   float64 `json:"amount"`
		Currency string  `json:"currency"`
	}{Amount: amount, Currency: currency}
	data, err := json.Marshal(details)
	if err == nil {
		r.redis.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return amount, currency, nil
}

func (r *paymentRepository) CreatePayment(ctx context.Context, fromUserID, toUserID string, amount float64, currency string) (string, error) {
	id := uuid.New().String()
	query := `INSERT INTO payments (id, from_user_id, to_user_id, amount, currency, status) 
			  VALUES ($1, $2, $3, $4, $5, 'PENDING') RETURNING id`

	var paymentID string
	err := r.db.QueryRow(ctx, query, id, fromUserID, toUserID, amount, currency).Scan(&paymentID)
	if err != nil {
		r.logger.Error("Failed to create payment", zap.Error(err))
		return "", fmt.Errorf("error creating payment: %w", err)
	}

	r.logger.Info("Payment created", zap.String("payment_id", paymentID))
	return paymentID, nil
}

func (r *paymentRepository) UpdatePaymentStatus(ctx context.Context, paymentID string, status models.PaymentStatus) error {
	query := `UPDATE payments SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(ctx, query, status, time.Now(), paymentID)
	if err != nil {
		r.logger.Error("Failed to update payment status", zap.String("payment_id", paymentID), zap.String("status", string(status)), zap.Error(err))
		return fmt.Errorf("error updating payment status: %w", err)
	}

	r.logger.Info("Payment status updated", zap.String("payment_id", paymentID), zap.String("status", string(status)))
	return nil
}

func (r *paymentRepository) GetActivePayments(ctx context.Context, userID string) ([]*models.Payment, error) {
	query := `SELECT id, from_user_id, to_user_id, amount, currency, status, created_at, updated_at 
			  FROM payments WHERE from_user_id = $1 AND status = 'PENDING' OR from_user_id = $1 AND status = 'FAILED'`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		r.logger.Error("Failed to fetch payment history", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("error fetching payment history: %w", err)
	}
	defer rows.Close()

	var payments []*models.Payment
	for rows.Next() {
		var payment models.Payment
		if err := rows.Scan(
			&payment.ID,
			&payment.FromUserID,
			&payment.ToUserID,
			&payment.Amount,
			&payment.Currency,
			&payment.Status,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		); err != nil {
			r.logger.Error("Failed to scan payment history row", zap.Error(err))
			return nil, fmt.Errorf("error scanning payment history: %w", err)
		}
		payments = append(payments, &payment)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error occurred during rows iteration", zap.Error(err))
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return payments, nil
}
