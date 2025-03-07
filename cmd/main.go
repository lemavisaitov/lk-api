package main

import (
	"context"
	"github.com/lemavisaitov/lk-api/internal/app"
	"github.com/lemavisaitov/lk-api/internal/handler"
	"github.com/lemavisaitov/lk-api/internal/storage"
	"log"
	"time"
)

func main() {
	connStr := "host=postgres port=5432 user=postgres database=postgres password=postgres sslmode=disable"

	ctx := context.Background()
	withTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	s, err := storage.GetConnect(withTimeout, connStr)
	if err != nil {
		log.Fatal(err)
	}

	h := handler.New(s)
	router := app.GetRouter(h)

	err = router.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
