package main

import (
	"log"

	"github.com/pinnakarn-k/commerce-core-go/internal/app"
	"github.com/pinnakarn-k/commerce-core-go/internal/config"
)

// @title Commerce Core Go API
// @version 1.0
// @description Production-minded e-commerce backend API.
// @BasePath /api/v1
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.Load()

	app, err := app.New(cfg)
	if err != nil {
		return err
	}

	return app.Run()
}
