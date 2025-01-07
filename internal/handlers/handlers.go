package handlers

import (
	"context"
	"fmt"
	"gitlab.crja72.ru/gospec/go8/payment/internal/models"

	"gitlab.crja72.ru/gospec/go8/payment/internal/payment-service/proto"
	"gitlab.crja72.ru/gospec/go8/payment/internal/service"
	"go.uber.org/zap"
)

// PaymentHandler структура для ручек оплаты
type PaymentHandler struct {
	proto.UnimplementedPaymentServiceServer
	service *service.PaymentService
	logger  *zap.Logger
}

// NewPaymentHandler создание экземпляра ручек оплаты
func NewPaymentHandler(service *service.PaymentService, logger *zap.Logger) *PaymentHandler {
	return &PaymentHandler{service: service, logger: logger}
}

// GetPaymentLink ручка получение ссылки на оплату
func (h *PaymentHandler) GetPaymentLink(ctx context.Context, req *proto.GetPaymentLinkRequest) (*proto.GetPaymentLinkResponse, error) {
	paymentLink, err := h.service.GetPaymentLink(ctx, req.PaymentId)

	if err != nil {
		return nil, fmt.Errorf("error generating link for payment: %w", err)
	}

	return &proto.GetPaymentLinkResponse{
		PaymentLink: paymentLink,
	}, nil
}

// GetPayment ручка получения статуса оплаты
func (h *PaymentHandler) GetPayment(ctx context.Context, req *proto.GetPaymentRequest) (*proto.GetPaymentResponse, error) {
	paymentStatus, err := h.service.GetPayment(ctx, req.PaymentId)

	if err != nil {
		return nil, fmt.Errorf("error paying for payment: %w", err)
	}

	return &proto.GetPaymentResponse{
		Status: paymentStatus,
	}, nil
}

// CreatePayment Ручка создания оплаты
func (h *PaymentHandler) CreatePayment(ctx context.Context, req *proto.CreatePaymentRequest) (*proto.CreatePaymentResponse, error) {
	paymentID, err := h.service.CreatePayment(ctx, req.FromUserId, req.ToUserId, float64(req.Amount), req.Currency)
	if err != nil {
		return nil, fmt.Errorf("error creating payment: %w", err)
	}

	return &proto.CreatePaymentResponse{
		PaymentId: paymentID,
	}, nil
}

// RefundPayment Ручка создания возврата оплаты
func (h *PaymentHandler) RefundPayment(ctx context.Context, req *proto.RefundPaymentRequest) (*proto.RefundPaymentResponse, error) {
	err := h.service.RefundPayment(ctx, req.PaymentId)
	if err != nil {
		return nil, fmt.Errorf("error refunding payment: %w", err)
	}

	return &proto.RefundPaymentResponse{
		Status: string(models.StatusRefunded),
	}, nil
}

// GetPaymentByID Ручка получения данных оплаты по id
func (h *PaymentHandler) GetPaymentByID(ctx context.Context, req *proto.GetPaymentByIDRequest) (*proto.GetPaymentByIDResponse, error) {
	payment, err := h.service.GetPaymentByID(ctx, req.PaymentId)
	if err != nil {
		return nil, fmt.Errorf("error getting payment: %w", err)
	}

	return &proto.GetPaymentByIDResponse{
		Id:         payment.ID,
		FromUserId: payment.FromUserID,
		ToUserId:   payment.ToUserID,
		Amount:     float32(payment.Amount),
		Currency:   payment.Currency,
		Status:     string(payment.Status),
		CreatedAt:  payment.CreatedAt.String(),
		UpdatedAt:  payment.UpdatedAt.String(),
	}, nil
}

// GetPaymentHistory получение истории оплат
func (h *PaymentHandler) GetPaymentHistory(ctx context.Context, req *proto.GetPaymentHistoryRequest) (*proto.GetPaymentHistoryResponse, error) {
	payments, err := h.service.GetPaymentHistory(ctx, req.FromUserId, int(req.Page), int(req.Limit))
	if err != nil {
		return nil, fmt.Errorf("error getting payment history: %w", err)
	}

	var protoPayments []*proto.Payment
	for _, payment := range payments {
		protoPayments = append(protoPayments, &proto.Payment{
			Id:         payment.ID,
			FromUserId: payment.FromUserID,
			ToUserId:   payment.ToUserID,
			Amount:     float32(payment.Amount),
			Currency:   payment.Currency,
			Status:     string(payment.Status),
			CreatedAt:  payment.CreatedAt.String(),
			UpdatedAt:  payment.UpdatedAt.String(),
		})
	}

	return &proto.GetPaymentHistoryResponse{
		Payment: protoPayments,
	}, nil
}

// GetActivePayments получение активных счетов оплаты
func (h *PaymentHandler) GetActivePayments(ctx context.Context, req *proto.GetActivePaymentsRequest) (*proto.GetActivePaymentsResponse, error) {
	payments, err := h.service.GetActivePayments(ctx, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("error getting active payments: %w", err)
	}

	var protoPayments []*proto.Payment
	for _, payment := range payments {
		protoPayments = append(protoPayments, &proto.Payment{
			Id:         payment.ID,
			FromUserId: payment.FromUserID,
			ToUserId:   payment.ToUserID,
			Amount:     float32(payment.Amount),
			Currency:   payment.Currency,
			Status:     string(payment.Status),
			CreatedAt:  payment.CreatedAt.String(),
			UpdatedAt:  payment.UpdatedAt.String(),
		})
	}

	return &proto.GetActivePaymentsResponse{
		Payments: protoPayments,
	}, nil
}
