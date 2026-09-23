package agent

import (
	"time"

	"github.com/vpsflow/vpsflow/libs/go/agentprotocol"
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

func (h *Handler) RegisterRoutes(app fiber.Router) {
	agent := app.Group("/agent/v1")
	agent.Post("/register", h.register)
	agent.Post("/heartbeat", h.heartbeat)
	agent.Get("/commands/poll", h.pollCommand)
	agent.Post("/commands/:id/complete", h.completeCommand)
}

func (h *Handler) register(c *fiber.Ctx) error {
	var req agentprotocol.RegisterAgentRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}
	resp, err := h.service.RegisterAgent(c.UserContext(), req)
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *Handler) heartbeat(c *fiber.Ctx) error {
	var req agentprotocol.HeartbeatRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}
	if err := h.service.Heartbeat(c.UserContext(), req); err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(fiber.Map{"acknowledged": true})
}

func (h *Handler) pollCommand(c *fiber.Ctx) error {
	hypervisorID := c.Query("hypervisor_id")
	sessionID := c.Query("session_id")
	timeout := 25 * time.Second
	deadline := time.Now().Add(timeout)
	for {
		cmd, err := h.service.PollCommand(c.UserContext(), hypervisorID, sessionID)
		if err != nil {
			return httpx.WriteError(c, err)
		}
		if cmd != nil {
			return c.JSON(cmd)
		}
		if time.Now().After(deadline) {
			return c.SendStatus(fiber.StatusNoContent)
		}
		select {
		case <-c.UserContext().Done():
			return c.SendStatus(fiber.StatusNoContent)
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func (h *Handler) completeCommand(c *fiber.Ctx) error {
	var req agentprotocol.CommandResult
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}
	req.CommandID = c.Params("id")
	if err := h.service.CompleteCommand(c.UserContext(), req); err != nil {
		return httpx.WriteError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
