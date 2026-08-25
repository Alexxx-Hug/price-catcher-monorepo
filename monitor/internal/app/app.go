package app

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/adapters/consumer"
	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/adapters/parser"
	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/config"
	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/metrics"
	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/providers"
	grpcserver "github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/server/grpc"
	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/service"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type App struct {
	cfg                *config.Config
	logger             *zap.Logger
	grpcServer         *grpcserver.Server
	priceCheckConsumer *consumer.PriceCheckConsumer
	kafkaProvider      *providers.KafkaProvider
	metricsRegistry    *prometheus.Registry
}

func NewApp(cfg *config.Config, logger *zap.Logger) (*App, error) {
	productParser := parser.NewWildberriesParser()
	appMetrics, metricsRegistry := metrics.New(cfg.App.Name)

	kafkaProvider, err := providers.NewKafkaProvider(cfg.Kafka, appMetrics)
	if err != nil {
		return nil, fmt.Errorf("init kafka provider: %w", err)
	}

	monitorService := service.NewMonitorService(productParser, kafkaProvider.ProductCheckedProducer)
	monitorHandler := grpcserver.NewMonitorHandler(monitorService)

	priceCheckConsumer := consumer.NewPriceCheckConsumer(
		cfg.Kafka.BrokerList(),
		cfg.Kafka.TaskCheckPricesTopic,
		cfg.Kafka.GroupID,
		monitorService,
		logger,
		appMetrics,
	)

	server := grpcserver.NewServer(cfg.GRPC.Port, monitorHandler, logger)

	return &App{
		cfg:                cfg,
		logger:             logger,
		grpcServer:         server,
		kafkaProvider:      kafkaProvider,
		priceCheckConsumer: priceCheckConsumer,
		metricsRegistry:    metricsRegistry,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.logger.Info(
		"application started",
		zap.String("app", a.cfg.App.Name),
		zap.String("version", a.cfg.App.Version),
	)

	go func() {
		if err := a.priceCheckConsumer.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			a.logger.Error("price check consumer stopped", zap.Error(err))
		}
	}()

	go func() {
		if err := metrics.RunServer(ctx, a.cfg.Metrics.Port, a.metricsRegistry, a.logger); err != nil && !errors.Is(err, context.Canceled) {
			a.logger.Error("metrics server stopped", zap.Error(err))
		}
	}()

	if err := a.grpcServer.Run(ctx); err != nil {
		return fmt.Errorf("run grpc server: %w", err)
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

	application, err := NewApp(cfg, logger)
	if err != nil {
		return err
	}

	defer application.kafkaProvider.Close()
	defer application.priceCheckConsumer.Close()

	return application.Run(ctx)
}
