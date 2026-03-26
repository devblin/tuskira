package handler

import (
	"encoding/json"
	"net/http"

	"github.com/devblin/tuskira/internal/model"
	"github.com/devblin/tuskira/internal/service"
	"github.com/labstack/echo/v4"
)

type ChannelConfigHandler struct {
	svc *service.ChannelConfigService
}

func NewChannelConfigHandler(svc *service.ChannelConfigService) *ChannelConfigHandler {
	return &ChannelConfigHandler{svc: svc}
}

type UpsertChannelConfigRequest struct {
	Channel model.Channel   `json:"channel" validate:"required"`
	Enabled bool            `json:"enabled"`
	Config  json.RawMessage `json:"config" validate:"required"`
}

func (h *ChannelConfigHandler) Upsert(c echo.Context) error {
	var req UpsertChannelConfigRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	cfg := &model.ChannelConfig{
		Channel: req.Channel,
		Enabled: req.Enabled,
		Config:  model.ChannelConfigData(req.Config),
	}

	if err := h.svc.Upsert(cfg); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	updated, err := h.svc.GetByChannel(req.Channel)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, updated)
}

func (h *ChannelConfigHandler) GetByChannel(c echo.Context) error {
	channel := model.Channel(c.Param("channel"))
	cfg, err := h.svc.GetByChannel(channel)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "channel config not found"})
	}
	return c.JSON(http.StatusOK, cfg)
}

func (h *ChannelConfigHandler) List(c echo.Context) error {
	configs, err := h.svc.List()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, configs)
}

func (h *ChannelConfigHandler) Delete(c echo.Context) error {
	channel := model.Channel(c.Param("channel"))
	if err := h.svc.Delete(channel); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusNoContent, nil)
}
