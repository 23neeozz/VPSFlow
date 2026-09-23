package vm

import (
	"strings"

	"github.com/vpsflow/vpsflow/libs/go/errors"
	"github.com/vpsflow/vpsflow/libs/go/httpx"
	"github.com/vpsflow/vpsflow/services/vm/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *usecase.Service
}

func NewHandler(service *usecase.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(api fiber.Router, auth fiber.Handler) {
	vms := api.Group("/virtual-machines", auth, tenantMiddleware)
	vms.Post("/", h.create)
	vms.Get("/", h.list)
	vms.Get("/:id", h.get)
	vms.Post("/:id/start", h.start)
	vms.Post("/:id/stop", h.stop)
	vms.Delete("/:id", h.delete)
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
	tenantID, _ := c.Locals("tenant_id").(string)
	var req struct {
		Name      string `json:"name"`
		ProjectID string `json:"project_id"`
		VCPUs     int    `json:"vcpus"`
		MemoryMB  int    `json:"memory_mb"`
		DiskGB    int    `json:"disk_gb"`
		ImageRef  string `json:"image_ref"`
	}
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}
	vm, err := h.service.CreateVM(c.UserContext(), usecase.CreateVMRequest{
		TenantID: tenantID, ProjectID: req.ProjectID, Name: req.Name,
		VCPUs: req.VCPUs, MemoryMB: req.MemoryMB, DiskGB: req.DiskGB, ImageRef: req.ImageRef,
		IdempotencyKey: c.Get("Idempotency-Key"),
	})
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.Status(fiber.StatusAccepted).JSON(vm)
}

func (h *Handler) list(c *fiber.Ctx) error {
	tenantID, _ := c.Locals("tenant_id").(string)
	vms, err := h.service.ListVMs(c.UserContext(), tenantID)
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(fiber.Map{"virtual_machines": vms})
}

func (h *Handler) get(c *fiber.Ctx) error {
	tenantID, _ := c.Locals("tenant_id").(string)
	vm, err := h.service.GetVM(c.UserContext(), tenantID, c.Params("id"))
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(vm)
}

func (h *Handler) start(c *fiber.Ctx) error {
	tenantID, _ := c.Locals("tenant_id").(string)
	vm, err := h.service.StartVM(c.UserContext(), tenantID, c.Params("id"))
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(vm)
}

func (h *Handler) stop(c *fiber.Ctx) error {
	tenantID, _ := c.Locals("tenant_id").(string)
	vm, err := h.service.StopVM(c.UserContext(), tenantID, c.Params("id"))
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(vm)
}

func (h *Handler) delete(c *fiber.Ctx) error {
	tenantID, _ := c.Locals("tenant_id").(string)
	if err := h.service.DeleteVM(c.UserContext(), tenantID, c.Params("id")); err != nil {
		return httpx.WriteError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
