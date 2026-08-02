package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/config"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/db"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/delivery"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/repository/postgres"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/usecase"

	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("cannot init logger: %s", err)
	}
	defer logger.Sync()

	cfg := config.MustLoad()
	logger.Info("config loaded", zap.String("app", cfg.App.Name), zap.String("version", cfg.App.Version))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := db.NewPostgresPool(ctx, cfg.DB.GetDSN())
	if err != nil {
		logger.Fatal("cannot connect to database", zap.Error(err))
	}
	defer pool.Close()

	logger.Info("successfully connected to database") // на проде нахуй не надо, от обратного идем - приложение не упало (не отработал логгер на 37 строке) - коннект к базе есть

	productRepo := postgres.NewProductRepository(pool)
	productUseCase := usecase.NewProductUseCase(productRepo, logger)

	r := delivery.NewRouter(productUseCase, logger)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      r.GetEngine(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("server is running", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-stopChan
	logger.Info("shutting down gracefully...", zap.String("signal", sig.String()))

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server exited cleanly")
}
