package console

import (
	"strings"

	"github.com/vpsflow/vpsflow/libs/go/errors"
	"github.com/vpsflow/vpsflow/libs/go/httpx"
	"github.com/vpsflow/vpsflow/libs/go/httpx/middleware"
	"github.com/vpsflow/vpsflow/services/console/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *usecase.Service
}

func NewHandler(service *usecase.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(api fiber.Router, auth fiber.Handler) {
	group := api.Group("/console", auth, tenantMiddleware)
	group.Post("/sessions", h.createSession)
	group.Delete("/sessions/:id", h.revokeSession)
	api.Get("/console/ws", h.validateWS)
}

func tenantMiddleware(c *fiber.Ctx) error {
	tenantID := strings.TrimSpace(c.Get("X-Tenant-ID"))
	if tenantID == "" {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "X-Tenant-ID header is required"))
	}
	c.Locals("tenant_id", tenantID)
	return c.Next()
}

func (h *Handler) createSession(c *fiber.Ctx) error {
	ctx, ok := middleware.Context(c)
	if !ok {
		return httpx.WriteError(c, fiber.ErrUnauthorized)
	}
	tenantID, _ := c.Locals("tenant_id").(string)
	var req struct {
		VMID string `json:"vm_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}
	token, err := h.service.CreateSession(c.UserContext(), tenantID, req.VMID, ctx.UserID, c.Get("Authorization"))
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(token)
}

func (h *Handler) revokeSession(c *fiber.Ctx) error {
	if err := h.service.RevokeSession(c.UserContext(), c.Params("id")); err != nil {
		return httpx.WriteError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) validateWS(c *fiber.Ctx) error {
	session, err := h.service.ValidateToken(c.UserContext(), c.Query("token"))
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(fiber.Map{
		"session_id": session.ID,
		"vm_id":      session.VMID,
		"tenant_id":  session.TenantID,
		"status":     "ready",
		"message":    "console proxy handshake accepted",
	})
}
