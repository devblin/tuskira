package handler

import (
	"net/http"
	"strconv"

	"github.com/devblin/tuskira/internal/model"
	"github.com/devblin/tuskira/internal/service"
	"github.com/labstack/echo/v4"
)

type TemplateHandler struct {
	svc *service.TemplateService
}

func NewTemplateHandler(svc *service.TemplateService) *TemplateHandler {
	return &TemplateHandler{svc: svc}
}

type CreateTemplateRequest struct {
	Name    string        `json:"name" validate:"required"`
	Channel model.Channel `json:"channel" validate:"required"`
	Subject string        `json:"subject"`
	Body    string        `json:"body" validate:"required"`
}

func (h *TemplateHandler) Create(c echo.Context) error {
	var req CreateTemplateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	t := &model.Template{
		Name:    req.Name,
		Channel: req.Channel,
		Subject: req.Subject,
		Body:    req.Body,
	}

	if err := h.svc.Create(t); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, t)
}

func (h *TemplateHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	t, err := h.svc.GetByID(uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "template not found"})
	}

	return c.JSON(http.StatusOK, t)
}

func (h *TemplateHandler) List(c echo.Context) error {
	templates, err := h.svc.List()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, templates)
}
