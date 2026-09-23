package cluster

import (
	"github.com/vpsflow/vpsflow/libs/go/errors"
	"github.com/vpsflow/vpsflow/libs/go/httpx"
	"github.com/vpsflow/vpsflow/services/cluster/internal/domain"
	"github.com/vpsflow/vpsflow/services/cluster/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *usecase.Service
}

func NewHandler(service *usecase.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(api fiber.Router, jwtMiddleware fiber.Handler, internalKey string) {
	cluster := api.Group("/hypervisors", jwtMiddleware)
	cluster.Get("/", h.listHypervisors)
	cluster.Get("/:id", h.getHypervisor)

	internal := api.Group("/internal/cluster", internalAPIKey(internalKey))
	internal.Post("/hypervisors/register", h.internalRegister)
	internal.Post("/hypervisors/:id/heartbeat", h.internalHeartbeat)
	internal.Post("/placement/select", h.internalSelect)
	internal.Post("/hypervisors/:id/release-capacity", h.internalReleaseCapacity)
	internal.Post("/nodes/drain-siblings", h.internalDrainSiblings)
}

func internalAPIKey(expected string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if expected == "" || c.Get("X-Internal-API-Key") != expected {
			return httpx.WriteError(c, errors.New(errors.CodeForbidden, "invalid internal api key"))
		}
		return c.Next()
	}
}

func (h *Handler) listHypervisors(c *fiber.Ctx) error {
	onlineOnly := c.Query("online_only", "false") == "true"
	clusterID := c.Query("cluster_id", "")
	hvs, err := h.service.ListHypervisors(c.UserContext(), clusterID, onlineOnly)
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(fiber.Map{"hypervisors": hvs})
}

func (h *Handler) getHypervisor(c *fiber.Ctx) error {
	hv, err := h.service.GetHypervisor(c.UserContext(), c.Params("id"))
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(hv)
}

type registerRequest struct {
	HypervisorID string              `json:"hypervisor_id"`
	NodeName     string              `json:"node_name"`
	AgentVersion string              `json:"agent_version"`
	Capacity     domain.NodeCapacity `json:"capacity"`
}

func (h *Handler) internalRegister(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}
	var hv *domain.Hypervisor
	var err error
	if req.HypervisorID != "" {
		hv, err = h.service.RegisterHypervisorWithID(c.UserContext(), req.HypervisorID, req.NodeName, req.AgentVersion, req.Capacity)
	} else {
		hv, err = h.service.RegisterHypervisor(c.UserContext(), req.NodeName, req.AgentVersion, req.Capacity)
	}
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(hv)
}

type heartbeatRequest struct {
	Capacity domain.NodeCapacity `json:"capacity"`
}

func (h *Handler) internalHeartbeat(c *fiber.Ctx) error {
	var req heartbeatRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}
	if err := h.service.Heartbeat(c.UserContext(), c.Params("id"), req.Capacity); err != nil {
		return httpx.WriteError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

type placementRequest struct {
	VCPUs     int    `json:"vcpus"`
	MemoryMB  int    `json:"memory_mb"`
	DiskGB    int    `json:"disk_gb"`
	ClusterID string `json:"cluster_id"`
}

func (h *Handler) internalSelect(c *fiber.Ctx) error {
	var req placementRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}
	hv, err := h.service.SelectHypervisor(c.UserContext(), usecase.PlacementRequest{
		VCPUs: req.VCPUs, MemoryMB: req.MemoryMB, DiskGB: req.DiskGB, ClusterID: req.ClusterID,
	})
	if err != nil {
		return httpx.WriteError(c, err)
	}
	return c.JSON(hv)
}

type releaseRequest struct {
	VCPUs    int `json:"vcpus"`
	MemoryMB int `json:"memory_mb"`
	DiskGB   int `json:"disk_gb"`
}

func (h *Handler) internalReleaseCapacity(c *fiber.Ctx) error {
	var req releaseRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}
	if err := h.service.ReleaseCapacity(c.UserContext(), c.Params("id"), req.VCPUs, req.MemoryMB, req.DiskGB); err != nil {
		return httpx.WriteError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

type drainSiblingsRequest struct {
	NodeName     string `json:"node_name"`
	HypervisorID string `json:"hypervisor_id"`
}

func (h *Handler) internalDrainSiblings(c *fiber.Ctx) error {
	var req drainSiblingsRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}
	if err := h.service.DrainSiblingHypervisors(c.UserContext(), req.NodeName, req.HypervisorID); err != nil {
		return httpx.WriteError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
