package main

import (
	"context"
	"fmt"
	"log"

	"github.com/lemavisaitov/lk-api/config"
	"github.com/lemavisaitov/lk-api/internal/app"
	"github.com/lemavisaitov/lk-api/internal/cache"
	"github.com/lemavisaitov/lk-api/internal/handler"
	"github.com/lemavisaitov/lk-api/internal/logger"
	"github.com/lemavisaitov/lk-api/internal/metrics"
	"github.com/lemavisaitov/lk-api/internal/repository"
	"github.com/lemavisaitov/lk-api/internal/storage"
	"github.com/lemavisaitov/lk-api/internal/usecase"
	"github.com/lemavisaitov/lk-api/migrations"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	if err := logger.Init(cfg.LogLevel); err != nil {
		log.Fatal(err)
	}

	connStr := cfg.GetDBConnStr()
	logger.Info("connecting to database",
		zap.String("connection string", connStr),
	)

	ctx := context.Background()
	withTimeout, cancel := context.WithTimeout(ctx, cfg.DBConnTimeout)
	defer cancel()

	if err := migrations.Migrate(connStr); err != nil {
		logger.Fatal("error while migrating database",
			zap.Error(errors.Wrap(err, "")),
		)
	}
	pool, err := storage.GetConnect(withTimeout, connStr)
	if err != nil {
		logger.Fatal("error while connecting to storage",
			zap.Error(errors.Wrap(err, "")),
		)
	}
	userRepo := repository.NewUserProvider(pool)
	cacheProvider, err := cache.NewDecorator(userRepo, cfg.CacheCleanupInterval, cfg.CacheTTL)
	if err != nil {
		logger.Fatal("error while initializing cache",
			zap.Error(errors.Wrap(err, "")),
		)
	}
	defer cacheProvider.Close()
	defer pool.Close()

	userUC := usecase.NewUserProvider(cacheProvider)
	handle := handler.New(userUC)
	router := app.GetRouter(handle)

	metrics.InitMetrics(cfg.MetricsAddress, cacheProvider)

	if err := router.Run(fmt.Sprintf(":%s", cfg.AppAddress)); err != nil {
		logger.Fatal("error while starting server",
			zap.Error(errors.Wrap(err, "")),
		)
	}
}
