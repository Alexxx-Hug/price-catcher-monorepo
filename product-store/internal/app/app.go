package app

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/config"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/db"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/delivery"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/repository/postgres"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/server"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/usecase"

	"go.uber.org/zap"
)

type App struct {
	cfg    *config.Config
	logger *zap.Logger
	server *server.Server
}

func NewApp(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*App, func(), error) {
	pool, err := db.NewPostgresPool(ctx, cfg.DB.GetDSN())
	if err != nil {
		return nil, nil, fmt.Errorf("connect to database: %w", err)
	}

	productRepo := postgres.NewProductRepository(pool)
	productUseCase := usecase.NewProductUseCase(productRepo, logger)

	router := delivery.NewRouter(productUseCase, logger)
	httpServer := server.NewServer(cfg.HTTP.Port, router.GetEngine(), logger)

	cleanup := func() {
		pool.Close()
	}

	return &App{
		cfg:    cfg,
		logger: logger,
		server: httpServer,
	}, cleanup, nil
}

func (a *App) Run(ctx context.Context) error {
	a.logger.Info(
		"application started",
		zap.String("app", a.cfg.App.Name),
		zap.String("version", a.cfg.App.Version),
	)

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
