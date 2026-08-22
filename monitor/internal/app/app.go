package app

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/adapters/parser"
	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/config"
	grpcserver "github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/server/grpc"
	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/service"
	"go.uber.org/zap"
)

type App struct {
	cfg        *config.Config
	logger     *zap.Logger
	grpcServer *grpcserver.Server
}

func NewApp(cfg *config.Config, logger *zap.Logger) (*App, error) {
	productParser := &parser.FakeParser{}

	monitorService := service.NewMonitorService(productParser)
	monitorHandler := grpcserver.NewMonitorHandler(monitorService)

	server := grpcserver.NewServer(cfg.GRPC.Port, monitorHandler, logger)

	return &App{
		cfg:        cfg,
		logger:     logger,
		grpcServer: server,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.logger.Info(
		"application started",
		zap.String("app", a.cfg.App.Name),
		zap.String("version", a.cfg.App.Version),
	)

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

	return application.Run(ctx)
}
