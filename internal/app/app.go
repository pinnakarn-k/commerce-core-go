package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/pinnakarn-k/commerce-core-go/docs"

	"github.com/gin-gonic/gin"
	"github.com/pinnakarn-k/commerce-core-go/internal/config"
	"github.com/pinnakarn-k/commerce-core-go/internal/platform/database/postgres"
	"github.com/pinnakarn-k/commerce-core-go/internal/platform/httpmiddleware"
	"github.com/pinnakarn-k/commerce-core-go/internal/platform/logger"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	grpcserver "github.com/pinnakarn-k/commerce-core-go/internal/grpc"
	googlegrpc "google.golang.org/grpc"
)

type App struct {
	cfg        config.Config
	logger     *slog.Logger
	postgres   *postgres.Postgres
	httpServer *http.Server
	grpcServer *googlegrpc.Server
}

func New(cfg config.Config) (*App, error) {
	logger := logger.New(cfg.Env).With(
		"service", cfg.Service,
		"env", cfg.Env,
	)

	db, err := postgres.New(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("init postgres: %w", err)
	}

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(httpmiddleware.RequestID())
	router.Use(gin.Recovery())
	router.Use(httpmiddleware.Logger(logger))

	api := router.Group("/api")
	v1 := api.Group("/v1")

	appInstance := &App{
		cfg:      cfg,
		logger:   logger,
		postgres: db,
	}

	if err := appInstance.registerModules(v1, db); err != nil {
		db.Close()
		return nil, err
	}

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.AppPort),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	appInstance.httpServer = server

	return appInstance, nil
}

func (a *App) Handler() http.Handler {
	return a.httpServer.Handler
}

func (a *App) Run() error {
	errCh := make(chan error, 1)

	go func() {
		a.logger.Info("server listening", "addr", a.httpServer.Addr)

		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	go func() {
		addr := fmt.Sprintf(":%d", a.cfg.GrpcPort)

		a.logger.Info("grpc server listening", "addr", addr)

		if err := grpcserver.Run(a.grpcServer, addr); err != nil {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case err := <-errCh:
		a.shutdown()
		return err

	case sig := <-quit:
		a.logger.Info("received signal", "signal", sig.String())

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		a.logger.Info("shutting down server")

		if err := a.httpServer.Shutdown(ctx); err != nil {
			a.shutdown()
			return err
		}

		a.shutdown()
		a.logger.Info("server stopped gracefully")
		return nil
	}
}

func (a *App) shutdown() {
	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}

	if a.postgres != nil {
		a.postgres.Close()
	}
}
