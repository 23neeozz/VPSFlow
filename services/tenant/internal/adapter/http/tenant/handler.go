package tenant

import (
	"github.com/vpsflow/vpsflow/libs/go/errors"
	"github.com/vpsflow/vpsflow/libs/go/httpx"
	"github.com/vpsflow/vpsflow/libs/go/httpx/middleware"
	"github.com/vpsflow/vpsflow/services/tenant/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *usecase.Service
}

func NewHandler(service *usecase.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(api fiber.Router, auth fiber.Handler) {
	api.Get("/organizations", auth, h.listOrganizations)
	api.Post("/organizations", auth, h.createOrganization)
	api.Get("/organizations/:org_id", auth, h.getOrganization)
	api.Get("/organizations/:org_id/members", auth, h.listMembers)
	api.Post("/organizations/:org_id/members", auth, h.inviteMember)
	api.Get("/organizations/:org_id/projects", auth, h.listProjects)
	api.Post("/organizations/:org_id/projects", auth, h.createProject)
}

func (h *Handler) createOrganization(c *fiber.Ctx) error {
	ctx, ok := middleware.Context(c)
	if !ok {
		return httpx.WriteError(c, fiber.ErrUnauthorized)
	}
	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}
	org, err := h.service.CreateOrganization(c.UserContext(), ctx.UserID, ctx.Email, req.Name, req.Slug, c.Get("Authorization"))
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(org)
}

func (h *Handler) listOrganizations(c *fiber.Ctx) error {
	ctx, ok := middleware.Context(c)
	if !ok {
		return httpx.WriteError(c, fiber.ErrUnauthorized)
	}
	orgs, err := h.service.ListOrganizations(c.UserContext(), ctx.UserID)
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(fiber.Map{"data": orgs})
}

func (h *Handler) getOrganization(c *fiber.Ctx) error {
	ctx, ok := middleware.Context(c)
	if !ok {
		return httpx.WriteError(c, fiber.ErrUnauthorized)
	}
	org, err := h.service.GetOrganization(c.UserContext(), c.Params("org_id"), ctx.UserID)
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(org)
}

func (h *Handler) listMembers(c *fiber.Ctx) error {
	ctx, ok := middleware.Context(c)
	if !ok {
		return httpx.WriteError(c, fiber.ErrUnauthorized)
	}
	members, err := h.service.ListMembers(c.UserContext(), c.Params("org_id"), ctx.UserID)
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(fiber.Map{"data": members})
}

func (h *Handler) inviteMember(c *fiber.Ctx) error {
	ctx, ok := middleware.Context(c)
	if !ok {
		return httpx.WriteError(c, fiber.ErrUnauthorized)
	}
	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}
	if err := h.service.InviteMember(c.UserContext(), c.Params("org_id"), ctx.UserID, req.Email, req.Role); err != nil {
		return httpx.WriteError(c, err)
	}
	return c.SendStatus(fiber.StatusCreated)
}

func (h *Handler) createProject(c *fiber.Ctx) error {
	ctx, ok := middleware.Context(c)
	if !ok {
		return httpx.WriteError(c, fiber.ErrUnauthorized)
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}
	project, err := h.service.CreateProject(c.UserContext(), c.Params("org_id"), ctx.UserID, req.Name)
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(project)
}

func (h *Handler) listProjects(c *fiber.Ctx) error {
	ctx, ok := middleware.Context(c)
	if !ok {
		return httpx.WriteError(c, fiber.ErrUnauthorized)
	}
	projects, err := h.service.ListProjects(c.UserContext(), c.Params("org_id"), ctx.UserID)
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(fiber.Map{"data": projects})
}
