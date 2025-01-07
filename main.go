package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net"

	"gitlab.crja72.ru/gospec/go8/payment/internal/clients"
	paymentsDemon "gitlab.crja72.ru/gospec/go8/payment/internal/payment-demon"

	"gitlab.crja72.ru/gospec/go8/payment/internal/config"
	"gitlab.crja72.ru/gospec/go8/payment/internal/db"
	"gitlab.crja72.ru/gospec/go8/payment/internal/handlers"
	"gitlab.crja72.ru/gospec/go8/payment/internal/payment-service/proto"
	"gitlab.crja72.ru/gospec/go8/payment/internal/repository"
	"gitlab.crja72.ru/gospec/go8/payment/internal/service"
	"gitlab.crja72.ru/gospec/go8/payment/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

//go:embed migrations/*
var migrations embed.FS

func main() {
	cfg, err := config.LoadConfig() // создаем логгер
	if err != nil {
		panic(fmt.Errorf("Failed to load config: %v", err))
	}

	logger := utils.NewLogger(cfg)
	defer logger.Sync()

	ctx := context.WithValue(context.Background(), "logger", logger)

	dbConn, err := db.NewPostgres(ctx, cfg, logger) // создаем подключение к БД
	if err != nil {
		logger.Fatal("Failed to initialize PostgreSQL", zap.Error(err))
	}

	defer func() {
		if dbConn != nil {
			dbConn.Close()
			logger.Info("Database connection closed")
		}
	}()

	if err := db.MigratePostgres(ctx, dbConn, logger, migrations); err != nil { // выполняем миграции
		logger.Fatal("Failed to apply migrations", zap.Error(err))
	}

	const grpcServerAddress = "localhost:8888"

	authClient, err := clients.NewAuthClient(grpcServerAddress) // создаем клиент для авторизации
	if err != nil {
		log.Fatalf("Failed to create AuthClient: %v", err)
	}
	defer authClient.Close()

	rdb := db.InitRedis(cfg, logger)
	defer rdb.Close()

	paymentsQueue := db.NewPaymentsQueue() // создаем очередь

	converter := clients.NewForexClient(cfg)        // создаем клиент для конвертации
	paymentClient := clients.NewYooMoneyClient(cfg) // создаем клиент для платежей

	repo := repository.NewPaymentRepository(dbConn, logger, rdb)                            // создаем репозиторий
	svc := service.NewPaymentService(repo, logger, converter, paymentClient, paymentsQueue) // создаем сервис

	demon := paymentsDemon.NewPaymentDemon(*svc, repo, paymentClient, paymentsQueue, logger, authClient) // создаем демон
	go demon.Start(ctx)

	grpcServer := grpc.NewServer()                                 // создаем сервер
	paymentHandler := handlers.NewPaymentHandler(svc, logger)      // создаем обработчик
	proto.RegisterPaymentServiceServer(grpcServer, paymentHandler) // подключаем обработчик

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		logger.Fatal("Failed to start gRPC listener", zap.Error(err))
	}
	logger.Info(fmt.Sprintf("Starting gRPC server on port %d", cfg.Server.Port))
	if err := grpcServer.Serve(listener); err != nil {
		logger.Fatal("Failed to start gRPC server", zap.Error(err))
	}
}
