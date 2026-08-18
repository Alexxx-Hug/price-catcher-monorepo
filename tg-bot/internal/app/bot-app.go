package app

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/adapters/parser"
	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/adapters/productstore"
	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/adapters/telegram"
	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/config"
	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/providers"
	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	cfg                  *config.Config
	logger               *zap.Logger
	bot                  *telegram.Bot
	kafkaProvider        *providers.KafkaProvider
	productStoreProvider *productstore.Client
	productStoreConn     *grpc.ClientConn
}

func NewApp(cfg *config.Config, logger *zap.Logger) (*App, error) {
	productParser := &parser.ProductParser{}
	kafkaProvider, err := providers.NewKafkaProvider(cfg.Kafka)
	if err != nil {
		return nil, fmt.Errorf("init kafka provider: %w", err)
	}

	conn, err := grpc.NewClient(
		cfg.ProductStoreGRPCConfig.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect product-store grpc: %w", err)
	}

	productStoreClient := productstore.NewClient(
		conn,
		cfg.ProductStoreGRPCConfig.Timeout,
	)

	botUseCase := service.NewBotUseCase(productParser, kafkaProvider.UserActionProducer, productStoreClient)

	bot, err := telegram.NewBot(cfg.Telegram.Token, botUseCase, logger)
	if err != nil {
		_ = conn.Close()
		_ = kafkaProvider.Close()
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}

	return &App{
		cfg:                  cfg,
		logger:               logger,
		bot:                  bot,
		kafkaProvider:        kafkaProvider,
		productStoreProvider: productStoreClient,
		productStoreConn:     conn,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.logger.Info(
		"application started",
		zap.String("app", a.cfg.App.Name),
		zap.String("version", a.cfg.App.Version),
	)

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
