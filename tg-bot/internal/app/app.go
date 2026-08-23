package app

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/adapters/consumer"
	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/adapters/monitor"
	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/adapters/productstore"
	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/adapters/telegram"
	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/config"
	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/providers"
	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	cfg                         *config.Config
	logger                      *zap.Logger
	bot                         *telegram.Bot
	kafkaProvider               *providers.KafkaProvider
	productPriceChangedConsumer *consumer.PriceChangedConsumer
	productStoreProvider        *productstore.Client
	productStoreConn            *grpc.ClientConn
	monitorConn                 *grpc.ClientConn
}

func NewApp(cfg *config.Config, logger *zap.Logger) (*App, error) {
	kafkaProvider, err := providers.NewKafkaProvider(cfg.Kafka)
	if err != nil {
		return nil, fmt.Errorf("init kafka provider: %w", err)
	}

	productStoreConn, err := grpc.NewClient(
		cfg.ProductStoreGRPCConfig.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		_ = kafkaProvider.Close()
		return nil, fmt.Errorf("connect product-store grpc: %w", err)
	}

	monitorConn, err := grpc.NewClient(
		cfg.MonitorGRPCConfig.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		_ = productStoreConn.Close()
		_ = kafkaProvider.Close()
		return nil, fmt.Errorf("connect monitor grpc: %w", err)
	}

	productStoreClient := productstore.NewClient(
		productStoreConn,
		cfg.ProductStoreGRPCConfig.Timeout,
	)

	monitorClient := monitor.NewClient(
		monitorConn,
		cfg.MonitorGRPCConfig.Timeout,
	)

	telegramAPI, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		_ = monitorConn.Close()
		_ = productStoreConn.Close()
		_ = kafkaProvider.Close()
		return nil, fmt.Errorf("create telegram api: %w", err)
	}

	notifier := telegram.NewNotifier(telegramAPI)
	botUseCase := service.NewBotUseCase(
		monitorClient,
		kafkaProvider.UserActionProducer,
		productStoreClient,
		notifier,
	)

	bot := telegram.NewBotWithAPI(telegramAPI, botUseCase, logger)

	productPriceChangedConsumer := consumer.NewPriceChangedConsumer(
		cfg.Kafka.ProductPriceChangedTopic,
		cfg.Kafka.BrokerList(),
		cfg.Kafka.GroupID,
		botUseCase,
		logger,
	)

	return &App{
		cfg:                         cfg,
		logger:                      logger,
		bot:                         bot,
		kafkaProvider:               kafkaProvider,
		productPriceChangedConsumer: productPriceChangedConsumer,
		productStoreProvider:        productStoreClient,
		productStoreConn:            productStoreConn,
		monitorConn:                 monitorConn,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.logger.Info(
		"application started",
		zap.String("app", a.cfg.App.Name),
		zap.String("version", a.cfg.App.Version),
	)

	go func() {
		if err := a.productPriceChangedConsumer.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			a.logger.Error("product price changed consumer stopped", zap.Error(err))
		}
	}()

	return a.bot.Start(ctx)
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
	defer application.Close()

	return application.Run(ctx)
}

func (a *App) Close() error {
	if a.productPriceChangedConsumer != nil {
		if err := a.productPriceChangedConsumer.Close(); err != nil {
			return fmt.Errorf("close product price changed consumer: %w", err)
		}
	}

	if a.monitorConn != nil {
		if err := a.monitorConn.Close(); err != nil {
			return fmt.Errorf("close monitor grpc connection: %w", err)
		}
	}

	if a.productStoreConn != nil {
		if err := a.productStoreConn.Close(); err != nil {
			return fmt.Errorf("close product-store grpc connection: %w", err)
		}
	}

	if a.kafkaProvider != nil {
		if err := a.kafkaProvider.Close(); err != nil {
			return fmt.Errorf("close kafka provider: %w", err)
		}
	}

	return nil
}
