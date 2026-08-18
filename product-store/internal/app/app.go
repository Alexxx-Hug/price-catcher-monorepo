package app

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"

	consumeradapter "github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/adapters/consumer"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/config"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/db"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/providers"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/repository/postgres"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/scheduler"
	grpcserver "github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/server/grpc"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/server/http"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/usecase"

	"go.uber.org/zap"
)

type App struct {
	cfg                 *config.Config
	logger              *zap.Logger
	server              *http.Server
	priceCheckScheduler *scheduler.PriceCheckScheduler
	readinessChecker    *http.ReadinessChecker
	productChecked      *consumeradapter.ProductCheckedConsumer
	userAction          *consumeradapter.UserActionConsumer
	grpcServer          *grpcserver.Server
}

func NewApp(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*App, func(), error) {
	pool, err := db.NewPostgresPool(ctx, cfg.DB.GetDSN())
	if err != nil {
		return nil, nil, fmt.Errorf("connect to database: %w", err)
	}

	kafkaProvider, err := providers.NewKafkaProvider(cfg.Kafka)
	if err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("init kafka provider: %w", err)
	}

	subRepo := postgres.NewSubscriptionRepository(pool)
	productRepo := postgres.NewProductRepository(pool)
	productUseCase := usecase.NewProductUseCase(productRepo, subRepo, logger, kafkaProvider.PriceCheckTaskProducer, kafkaProvider.PriceChangedProducer)
	subscriptionUseCase := usecase.NewSubcriptionUseCase(subRepo)
	subscriptionHandler := grpcserver.NewSubscriptionHandler(subscriptionUseCase)
	grpcServer := grpcserver.NewServer(
		cfg.GRPC.Port,
		subscriptionHandler,
		logger,
	)

	userActionUseCase := usecase.NewUserActionUseCase(productUseCase, subscriptionUseCase, logger)
	productCheckedConsumer := consumeradapter.NewProductCheckedConsumer(
		cfg.Kafka.BrokerList(),
		cfg.Kafka.ProductCheckedTopic,
		cfg.Kafka.GroupID,
		productUseCase,
		kafkaProvider.DeadLetterProducer,
		logger,
	)
	userActionConsumer := consumeradapter.NewUserActionConsumer(
		cfg.Kafka.BrokerList(),
		cfg.Kafka.UserActionsTopic,
		cfg.Kafka.GroupID,
		userActionUseCase,
		logger,
	)
	priceCheckScheduler := scheduler.NewPriceCheckScheduler(
		productUseCase,
		cfg.Monitoring.PriceCheckInterval,
		cfg.Monitoring.PriceCheckBatchLimit,
		logger,
	)

	readinessChecker := http.NewReadinessChecker(
		pool.Ping,
		cfg.HTTP.ReadinessCheckInterval,
		cfg.HTTP.ReadinessCheckTimeout,
		logger,
	)

	httpRouter := http.NewRouter(readinessChecker)
	httpServer := http.NewServer(cfg.HTTP.Port, httpRouter.GetEngine(), logger)

	cleanup := func() {
		_ = productCheckedConsumer.Close()
		_ = userActionConsumer.Close()
		pool.Close()
		_ = kafkaProvider.Close()
	}

	return &App{
		cfg:                 cfg,
		logger:              logger,
		server:              httpServer,
		priceCheckScheduler: priceCheckScheduler,
		readinessChecker:    readinessChecker,
		productChecked:      productCheckedConsumer,
		userAction:          userActionConsumer,
		grpcServer:          grpcServer,
	}, cleanup, nil
}

func (a *App) Run(ctx context.Context) error {
	a.logger.Info(
		"application started",
		zap.String("app", a.cfg.App.Name),
		zap.String("version", a.cfg.App.Version),
	)

	go func() {
		if err := a.priceCheckScheduler.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			a.logger.Error("price check scheduler stopped", zap.Error(err))
		}
	}()

	go func() {
		if err := a.readinessChecker.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			a.logger.Error("readiness checker stopped", zap.Error(err))
		}
	}()

	go func() {
		if err := a.productChecked.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			a.logger.Error("product checked consumer stopped", zap.Error(err))
		}
	}()

	go func() {
		if err := a.userAction.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			a.logger.Error("user action consumer stopped", zap.Error(err))
		}
	}()

	go func() {
		if err := a.grpcServer.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			a.logger.Error("grpc server stopped", zap.Error(err))
		}
	}()

	if err := a.server.Run(ctx); err != nil {
		return fmt.Errorf("run server: %w", err)
	}

	return nil
}

func Run() error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer logger.Sync()

	cfg := config.MustLoad()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application, cleanup, err := NewApp(ctx, cfg, logger)
	if err != nil {
		return err
	}
	defer cleanup()

	return application.Run(ctx)
}
