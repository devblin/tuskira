package handler

import (
	"net/http"

	"github.com/devblin/tuskira/internal/middleware"
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, ah *AuthHandler, nh *NotificationHandler, th *TemplateHandler, jwtSecret string) {
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	auth := e.Group("/api/v1/auth")
	auth.POST("/register", ah.Register)
	auth.POST("/login", ah.Login)

	api := e.Group("/api/v1", middleware.JWTMiddleware(jwtSecret))

	api.POST("/notifications", nh.Send)
	api.GET("/notifications/sent", nh.ListSent)
	api.GET("/notifications/scheduled", nh.GetPendingScheduled)
	api.GET("/notifications/:id", nh.GetByID)
	api.GET("/notifications", nh.ListByRecipient)
	api.POST("/notifications/:id/send", nh.TriggerSend)
	api.PATCH("/notifications/:id/schedule", nh.UpdateSchedule)
	api.POST("/notifications/:id/cancel", nh.CancelScheduled)

	api.POST("/templates", th.Create)
	api.GET("/templates/:id", th.GetByID)
	api.GET("/templates", th.List)
}
