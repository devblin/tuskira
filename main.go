package main

import (
	"context"
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
	"github.com/devblin/tuskira/pkg/database"
	"github.com/devblin/tuskira/pkg/queue"
	"github.com/devblin/tuskira/pkg/scheduler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg := config.Load()
	db := database.Init(cfg)

	// Queue and scheduler (provider chosen via config)
	q, err := queue.New(queue.Config{Provider: cfg.QueueProvider, EventKey: cfg.EventKey})
	if err != nil {
		log.Fatalf("failed to init queue: %v", err)
	}
	sched, err := scheduler.New(scheduler.Config{Provider: cfg.SchedulerProvider, EventKey: cfg.EventKey})
	if err != nil {
		log.Fatalf("failed to init scheduler: %v", err)
	}

	// Notification channel providers
	registry := provider.NewRegistry()
	registry.Register(provider.NewEmailProvider())
	registry.Register(provider.NewSlackProvider())
	registry.Register(provider.NewInAppProvider())

	// Repositories
	notifRepo := repository.NewNotificationRepository(db)
	tmplRepo := repository.NewTemplateRepository(db)

	// Services
	notifSvc := service.NewNotificationService(notifRepo, registry, q, sched)
	tmplSvc := service.NewTemplateService(tmplRepo)

	// HTTP server
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	nh := handler.NewNotificationHandler(notifSvc)
	th := handler.NewTemplateHandler(tmplSvc)
	handler.RegisterRoutes(e, nh, th)

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
	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
	log.Println("server stopped")
}
