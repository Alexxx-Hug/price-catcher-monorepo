package app

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/adapters/parser"
	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/adapters/telegram"
	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/config"
	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/providers"
	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/service"
	"go.uber.org/zap"
)

type App struct {
	cfg           *config.Config
	logger        *zap.Logger
	bot           *telegram.Bot
	kafkaProvider *providers.KafkaProvider
}

func NewApp(cfg *config.Config, logger *zap.Logger) (*App, error) {
	productParser := &parser.ProductParser{}
	kafkaProvider, err := providers.NewKafkaProvider(cfg.Kafka)
	if err != nil {
		return nil, fmt.Errorf("init kafka provider: %w", err)
	}

	botUseCase := service.NewBotUseCase(productParser, kafkaProvider.UserActionProducer)

	bot, err := telegram.NewBot(cfg.Telegram.Token, botUseCase, logger)
	if err != nil {
		_ = kafkaProvider.Close()
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}

	return &App{
		cfg:           cfg,
		logger:        logger,
		bot:           bot,
		kafkaProvider: kafkaProvider,
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

	return application.Run(ctx)
}
