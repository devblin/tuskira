package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devblin/tuskira/internal/config"
	"github.com/devblin/tuskira/internal/handler"
	"github.com/devblin/tuskira/internal/provider"
	"github.com/devblin/tuskira/internal/repository"
	"github.com/devblin/tuskira/internal/service"
	"github.com/devblin/tuskira/internal/sse"
	"github.com/devblin/tuskira/migrations"
	"github.com/devblin/tuskira/pkg/database"
	"github.com/devblin/tuskira/pkg/scheduler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg := config.Load()
	db, err := database.Init(cfg)
	if err != nil {
		log.Fatalf("failed to init database: %v", err)
	}

	if err := migrations.RunGormMigrations(db); err != nil {
		log.Fatalf("failed to run gorm migrations: %v", err)
	}

	if err := database.SeedDefaultUser(db); err != nil {
		log.Fatalf("failed to seed default user: %v", err)
	}

	// Set up pgx pool and run River migrations
	pgxPool, err := database.NewPgxPool(context.Background(), cfg)
	if err != nil {
		log.Fatalf("failed to create pgx pool: %v", err)
	}
	defer pgxPool.Close()

	if err := migrations.RunRiverMigrations(context.Background(), pgxPool); err != nil {
		log.Fatalf("failed to run river migrations: %v", err)
	}

	// Scheduler
	sched, err := scheduler.New(scheduler.Config{Provider: cfg.SchedulerProvider, Pool: pgxPool})
	if err != nil {
		log.Fatalf("failed to init scheduler: %v", err)
	}

	// SSE Hub for in-app notifications
	hub := sse.NewHub()

	// Notification channel providers (always registered, config checked at send time)
	registry := provider.NewRegistry()
	registry.Register(provider.NewEmailProvider())
	registry.Register(provider.NewSlackProvider())
	registry.Register(provider.NewInAppProvider(hub))

	// Repositories
	notifRepo := repository.NewNotificationRepository(db)
	tmplRepo := repository.NewTemplateRepository(db)
	userRepo := repository.NewUserRepository(db)
	channelConfigRepo := repository.NewChannelConfigRepository(db)

	// JWT config
	jwtExpiry, err := time.ParseDuration(cfg.JWTExpiry)
	if err != nil {
		log.Fatalf("invalid JWT_EXPIRY: %v", err)
	}

	// Services
	notifSvc := service.NewNotificationService(notifRepo, channelConfigRepo, registry, sched)
	tmplSvc := service.NewTemplateService(tmplRepo)
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, jwtExpiry)
	channelConfigSvc := service.NewChannelConfigService(channelConfigRepo)

	// Set up River worker handlers and start processing
	sched.SetHandler(func(ctx context.Context, externalID string, payload []byte) error {
		var data struct {
			NotificationID uint `json:"notification_id"`
		}
		if err := json.Unmarshal(payload, &data); err != nil {
			return fmt.Errorf("failed to unmarshal scheduled job payload: %w", err)
		}
		_, err := notifSvc.SendByID(data.NotificationID)
		return err
	})

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	if err := sched.Start(workerCtx); err != nil {
		log.Fatalf("failed to start scheduler workers: %v", err)
	}

	// HTTP server
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	ah := handler.NewAuthHandler(authSvc)
	nh := handler.NewNotificationHandler(notifSvc)
	th := handler.NewTemplateHandler(tmplSvc)
	ch := handler.NewChannelConfigHandler(channelConfigSvc)
	sh := handler.NewSSEHandler(hub, channelConfigSvc, notifRepo)
	handler.RegisterRoutes(e, ah, nh, th, ch, sh, cfg.JWTSecret)
	e.File("/web", "web/index.html")
	e.File("/web/", "web/index.html")
	e.Static("/web", "web")

	// Graceful shutdown
	go func() {
		if err := e.Start(fmt.Sprintf(":%s", cfg.ServerPort)); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Stop River workers gracefully
	workerCancel()
	if err := sched.Stop(ctx); err != nil {
		log.Printf("scheduler worker stop error: %v", err)
	}

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
	log.Println("server stopped")
}
