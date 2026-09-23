package internalapi

import (
	"github.com/vpsflow/vpsflow/libs/go/errors"
	"github.com/vpsflow/vpsflow/libs/go/httpx"
	"github.com/vpsflow/vpsflow/services/agent-control/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *usecase.Service
}

func NewHandler(service *usecase.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(api fiber.Router, internalKey string) {
	group := api.Group("/internal/agent-control", internalAPIKey(internalKey))
	group.Post("/commands", h.dispatch)
	group.Get("/commands/:id", h.getCommand)
}

func internalAPIKey(expected string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if expected == "" || c.Get("X-Internal-API-Key") != expected {
			return httpx.WriteError(c, errors.New(errors.CodeForbidden, "invalid internal api key"))
		}
		return c.Next()
	}
}

type dispatchRequest struct {
	HypervisorID string `json:"hypervisor_id"`
	TenantID     string `json:"tenant_id"`
	Type         string `json:"type"`
	Payload      any    `json:"payload"`
	Wait         bool   `json:"wait"`
}

func (h *Handler) dispatch(c *fiber.Ctx) error {
	var req dispatchRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}
	resp, err := h.service.DispatchCommand(c.UserContext(), usecase.DispatchRequest{
		HypervisorID: req.HypervisorID,
		TenantID:     req.TenantID,
		Type:         req.Type,
		Payload:      req.Payload,
		Wait:         req.Wait,
	})
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.Status(fiber.StatusAccepted).JSON(resp)
}

func (h *Handler) getCommand(c *fiber.Ctx) error {
	cmd, err := h.service.GetCommand(c.UserContext(), c.Params("id"))
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(cmd)
}
