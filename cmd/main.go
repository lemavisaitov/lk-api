package main

import (
	"context"
	"fmt"
	"github.com/lemavisaitov/lk-api/internal/app"
	"github.com/lemavisaitov/lk-api/internal/cache"
	"github.com/lemavisaitov/lk-api/internal/config"
	"github.com/lemavisaitov/lk-api/internal/handler"
	"github.com/lemavisaitov/lk-api/internal/repository"
	"github.com/lemavisaitov/lk-api/internal/storage"
	"github.com/lemavisaitov/lk-api/internal/usecase"
	"github.com/lemavisaitov/lk-api/migrations"
	"log"
)

func main() {
	cfg := config.MustLoad()

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, "5432")

	ctx := context.Background()
	withTimeout, cancel := context.WithTimeout(ctx, cfg.DBConnTimeout)
	defer cancel()

	pool, err := storage.GetConnect(withTimeout, connStr)
	if err != nil {
		log.Fatal(err)
	}
	if err := migrations.Migrate(connStr); err != nil {
		log.Fatal(err)
	}
	userRepo := repository.NewUserProvider(pool)
	cacheProvider := cache.NewDecorator(userRepo, cfg.CacheInterval, cfg.CacheTTL)
	userUC := usecase.NewUserProvider(cacheProvider)
	handle := handler.New(userUC)
	router := app.GetRouter(handle)

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
