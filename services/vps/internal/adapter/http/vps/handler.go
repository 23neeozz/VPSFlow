package vps

import (
	"strings"

	"github.com/vpsflow/vpsflow/libs/go/errors"
	"github.com/vpsflow/vpsflow/libs/go/httpx"
	"github.com/vpsflow/vpsflow/libs/go/httpx/middleware"
	"github.com/vpsflow/vpsflow/services/vps/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *usecase.Service
}

func NewHandler(service *usecase.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(api fiber.Router, auth fiber.Handler) {
	user := api.Group("/vps", auth, tenantMiddleware)
	user.Get("/", h.listMine)
	user.Get("/:id", h.getMine)
	user.Post("/:id/start", h.start)
	user.Post("/:id/stop", h.stop)
	user.Post("/:id/console", h.console)

	admin := api.Group("/admin/vps", auth, tenantMiddleware)
	admin.Post("/", h.create)
	admin.Get("/", h.listAll)
	admin.Get("/:id", h.getAdmin)
	admin.Put("/:id/assign", h.assign)
	admin.Delete("/:id", h.delete)
}

func tenantMiddleware(c *fiber.Ctx) error {
	tenantID := strings.TrimSpace(c.Get("X-Tenant-ID"))
	if tenantID == "" {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "X-Tenant-ID header is required"))
	}
	c.Locals("tenant_id", tenantID)
	return c.Next()
}

func (h *Handler) create(c *fiber.Ctx) error {
	ctx, ok := middleware.Context(c)
	if !ok {
		return httpx.WriteError(c, fiber.ErrUnauthorized)
	}
	tenantID, _ := c.Locals("tenant_id").(string)
	var req struct {
		Name        string `json:"name"`
		OwnerEmail  string `json:"owner_email"`
		OwnerUserID string `json:"owner_user_id"`
		VCPUs       int    `json:"vcpus"`
		MemoryMB    int    `json:"memory_mb"`
		DiskGB      int    `json:"disk_gb"`
		ImageRef    string `json:"image_ref"`
	}
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}
	vps, err := h.service.CreateVPS(c.UserContext(), usecase.CreateVPSRequest{
		TenantID: tenantID, ActorUserID: ctx.UserID, Authorization: c.Get("Authorization"),
		Name: req.Name, OwnerEmail: req.OwnerEmail, OwnerUserID: req.OwnerUserID,
		VCPUs: req.VCPUs, MemoryMB: req.MemoryMB, DiskGB: req.DiskGB, ImageRef: req.ImageRef,
		IdempotencyKey: c.Get("Idempotency-Key"),
	})
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.Status(fiber.StatusAccepted).JSON(vps)
}

func (h *Handler) listAll(c *fiber.Ctx) error {
	ctx, ok := middleware.Context(c)
	if !ok {
		return httpx.WriteError(c, fiber.ErrUnauthorized)
	}
	tenantID, _ := c.Locals("tenant_id").(string)
	items, err := h.service.ListAdmin(c.UserContext(), tenantID, ctx.UserID, c.Get("Authorization"))
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(fiber.Map{"vps_instances": items})
}

func (h *Handler) assign(c *fiber.Ctx) error {
	ctx, ok := middleware.Context(c)
	if !ok {
		return httpx.WriteError(c, fiber.ErrUnauthorized)
	}
	tenantID, _ := c.Locals("tenant_id").(string)
	var req struct {
		OwnerEmail  string `json:"owner_email"`
		OwnerUserID string `json:"owner_user_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}
	vps, err := h.service.AssignVPS(c.UserContext(), usecase.AssignVPSRequest{
		TenantID: tenantID, ActorUserID: ctx.UserID, Authorization: c.Get("Authorization"),
		VPSID: c.Params("id"), OwnerEmail: req.OwnerEmail, OwnerUserID: req.OwnerUserID,
	})
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(vps)
}

func (h *Handler) delete(c *fiber.Ctx) error {
	ctx, ok := middleware.Context(c)
	if !ok {
		return httpx.WriteError(c, fiber.ErrUnauthorized)
	}
	tenantID, _ := c.Locals("tenant_id").(string)
	if err := h.service.DeleteVPS(c.UserContext(), tenantID, ctx.UserID, c.Params("id"), c.Get("Authorization")); err != nil {
		return httpx.WriteError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) listMine(c *fiber.Ctx) error {
	ctx, ok := middleware.Context(c)
	if !ok {
		return httpx.WriteError(c, fiber.ErrUnauthorized)
	}
	tenantID, _ := c.Locals("tenant_id").(string)
	items, err := h.service.ListMine(c.UserContext(), tenantID, ctx.UserID, c.Get("Authorization"))
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(fiber.Map{"vps_instances": items})
}

func (h *Handler) getMine(c *fiber.Ctx) error {
	ctx, ok := middleware.Context(c)
	if !ok {
		return httpx.WriteError(c, fiber.ErrUnauthorized)
	}
	tenantID, _ := c.Locals("tenant_id").(string)
	vps, err := h.service.GetMine(c.UserContext(), tenantID, ctx.UserID, c.Params("id"), c.Get("Authorization"))
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(vps)
}

func (h *Handler) getAdmin(c *fiber.Ctx) error {
	ctx, ok := middleware.Context(c)
	if !ok {
		return httpx.WriteError(c, fiber.ErrUnauthorized)
	}
	tenantID, _ := c.Locals("tenant_id").(string)
	vps, err := h.service.GetAdmin(c.UserContext(), tenantID, ctx.UserID, c.Params("id"), c.Get("Authorization"))
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(vps)
}

func (h *Handler) start(c *fiber.Ctx) error {
	return h.lifecycle(c, "start")
}

func (h *Handler) stop(c *fiber.Ctx) error {
	return h.lifecycle(c, "stop")
}

func (h *Handler) lifecycle(c *fiber.Ctx, action string) error {
	ctx, ok := middleware.Context(c)
	if !ok {
		return httpx.WriteError(c, fiber.ErrUnauthorized)
	}
	tenantID, _ := c.Locals("tenant_id").(string)
	var vps any
	var err error
	if action == "start" {
		vps, err = h.service.StartVPS(c.UserContext(), tenantID, ctx.UserID, c.Params("id"), c.Get("Authorization"))
	} else {
		vps, err = h.service.StopVPS(c.UserContext(), tenantID, ctx.UserID, c.Params("id"), c.Get("Authorization"))
	}
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(vps)
}

func (h *Handler) console(c *fiber.Ctx) error {
	ctx, ok := middleware.Context(c)
	if !ok {
		return httpx.WriteError(c, fiber.ErrUnauthorized)
	}
	tenantID, _ := c.Locals("tenant_id").(string)
	session, err := h.service.OpenConsole(c.UserContext(), tenantID, ctx.UserID, c.Params("id"), c.Get("Authorization"))
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(session)
}
