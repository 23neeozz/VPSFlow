package iam

import (
	"github.com/vpsflow/vpsflow/libs/go/errors"
	"github.com/vpsflow/vpsflow/libs/go/httpx"
	"github.com/vpsflow/vpsflow/libs/go/httpx/middleware"
	"github.com/vpsflow/vpsflow/services/iam/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

// Handler exposes IAM HTTP endpoints.
type Handler struct {
	service *usecase.Service
}

// NewHandler creates an IAM HTTP handler.
func NewHandler(service *usecase.Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers IAM routes with authentication middleware.
func (h *Handler) RegisterRoutes(api fiber.Router, auth fiber.Handler, tenantAuth fiber.Handler) {
	api.Get("/iam/permissions", auth, h.listPermissions)
	api.Post("/iam/policy/evaluate", auth, h.evaluatePolicy)

	tenant := api.Group("/iam", tenantAuth)
	tenant.Get("/roles", h.listRoles)
	tenant.Post("/roles", h.createRole)
	tenant.Put("/users/:user_id/roles", h.assignUserRoles)
	tenant.Post("/bootstrap-owner", h.bootstrapOwner)
}

func (h *Handler) listPermissions(c *fiber.Ctx) error {
	permissions, err := h.service.ListPermissions(c.UserContext())
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(fiber.Map{"data": permissions})
}

func (h *Handler) listRoles(c *fiber.Ctx) error {
	ctx, ok := middleware.Context(c)
	if !ok {
		return httpx.WriteError(c, fiber.ErrUnauthorized)
	}
	roles, err := h.service.ListRoles(c.UserContext(), ctx.TenantID)
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(fiber.Map{"data": roles})
}

type createRoleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

func (h *Handler) createRole(c *fiber.Ctx) error {
	ctx, ok := middleware.Context(c)
	if !ok {
		return httpx.WriteError(c, fiber.ErrUnauthorized)
	}

	var req createRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}

	role, err := h.service.CreateRole(c.UserContext(), ctx.TenantID, req.Name, req.Description, req.Permissions)
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(role)
}

type assignRolesRequest struct {
	RoleIDs []string `json:"role_ids"`
}

func (h *Handler) assignUserRoles(c *fiber.Ctx) error {
	ctx, ok := middleware.Context(c)
	if !ok {
		return httpx.WriteError(c, fiber.ErrUnauthorized)
	}

	var req assignRolesRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}

	if err := h.service.AssignUserRoles(c.UserContext(), ctx.TenantID, c.Params("user_id"), req.RoleIDs); err != nil {
		return httpx.WriteError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

type evaluatePolicyRequest struct {
	ActorID    string `json:"actor_id"`
	TenantID   string `json:"tenant_id"`
	Permission string `json:"permission"`
	Resource   string `json:"resource"`
}

func (h *Handler) evaluatePolicy(c *fiber.Ctx) error {
	var req evaluatePolicyRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}

	decision, err := h.service.EvaluatePolicy(c.UserContext(), req.TenantID, req.ActorID, req.Permission)
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(decision)
}

type bootstrapOwnerRequest struct {
	UserID string `json:"user_id"`
}

func (h *Handler) bootstrapOwner(c *fiber.Ctx) error {
	ctx, ok := middleware.Context(c)
	if !ok {
		return httpx.WriteError(c, fiber.ErrUnauthorized)
	}

	var req bootstrapOwnerRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}
	if req.UserID == "" {
		req.UserID = ctx.UserID
	}

	if err := h.service.AssignOwner(c.UserContext(), ctx.TenantID, req.UserID); err != nil {
		return httpx.WriteError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}
