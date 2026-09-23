package health

import (
	"net/http"

	domainhealth "github.com/vpsflow/vpsflow/services/console/internal/domain/health"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	checker *domainhealth.Checker
}

func NewHandler(checker *domainhealth.Checker) *Handler {
	return &Handler{checker: checker}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).JSON(h.checker.Liveness())
	})
	app.Get("/readyz", func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).JSON(h.checker.Readiness())
	})
}
