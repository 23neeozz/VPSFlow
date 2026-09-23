package health

import (
	"net/http"

	domainhealth "github.com/bosscloud/bosscloud/services/auth/internal/domain/health"
	"github.com/gofiber/fiber/v2"
)

// Handler exposes health endpoints.
type Handler struct {
	checker *domainhealth.Checker
}

// NewHandler creates a health HTTP handler.
func NewHandler(checker *domainhealth.Checker) *Handler {
	return &Handler{checker: checker}
}

// RegisterRoutes registers liveness and readiness routes.
func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Get("/healthz", h.liveness)
	app.Get("/readyz", h.readiness)
}

func (h *Handler) liveness(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(h.checker.Liveness())
}

func (h *Handler) readiness(c *fiber.Ctx) error {
	report := h.checker.Readiness()
	status := http.StatusOK
	if report.Status == domainhealth.StatusUnhealthy {
		status = http.StatusServiceUnavailable
	}
	return c.Status(status).JSON(report)
}
