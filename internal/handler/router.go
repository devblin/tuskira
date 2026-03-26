package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, nh *NotificationHandler, th *TemplateHandler) {
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	api := e.Group("/api/v1")

	api.POST("/notifications", nh.Send)
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
