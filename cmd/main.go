package main

import (
	"context"
	"fmt"
	"github.com/lemavisaitov/lk-api/config"
	"github.com/lemavisaitov/lk-api/internal/app"
	"github.com/lemavisaitov/lk-api/internal/cache"
	"github.com/lemavisaitov/lk-api/internal/handler"
	"github.com/lemavisaitov/lk-api/internal/repository"
	"github.com/lemavisaitov/lk-api/internal/storage"
	"github.com/lemavisaitov/lk-api/internal/usecase"
	"github.com/lemavisaitov/lk-api/migrations"
	"log"
	"log/slog"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	connStr := cfg.GetDBConnStr()
	slog.Info(fmt.Sprintf("Connecting to database: %s", connStr))

	ctx := context.Background()
	withTimeout, cancel := context.WithTimeout(ctx, cfg.DBConnTimeout)
	defer cancel()

	if err := migrations.Migrate(connStr); err != nil {
		log.Fatal(err)
	}
	pool, err := storage.GetConnect(withTimeout, connStr)
	if err != nil {
		log.Fatal(err)
	}
	userRepo := repository.NewUserProvider(pool)
	cacheProvider, err := cache.NewDecorator(userRepo, cfg.CacheInterval, cfg.CacheTTL)
	if err != nil {
		log.Fatal(err)
	}
	defer cacheProvider.Close()

	userUC := usecase.NewUserProvider(cacheProvider)
	handle := handler.New(userUC)
	router := app.GetRouter(handle)

	if err := router.Run(fmt.Sprintf(":%s", cfg.AppAddress)); err != nil {
		log.Fatal(err)
	}
}
