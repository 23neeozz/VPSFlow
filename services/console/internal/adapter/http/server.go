package http

import (
	"log/slog"
	"net/http"
	"time"

	consolehandler "github.com/bosscloud/bosscloud/services/console/internal/adapter/http/console"
	healthhandler "github.com/bosscloud/bosscloud/services/console/internal/adapter/http/health"
	"github.com/bosscloud/bosscloud/services/console/internal/config"
	"github.com/bosscloud/bosscloud/services/console/internal/domain/health"
	"github.com/bosscloud/bosscloud/services/console/internal/usecase"
	"github.com/bosscloud/bosscloud/libs/go/httpx/middleware"
	"github.com/bosscloud/bosscloud/libs/go/security/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type Server struct {
	app *fiber.App
	cfg config.Config
	log *slog.Logger
}

func NewServer(cfg config.Config, log *slog.Logger, consoleService *usecase.Service, jwtService *jwt.Service) *Server {
	app := fiber.New(fiber.Config{
		AppName:      cfg.ServiceName,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		ErrorHandler: defaultErrorHandler,
	})
	app.Use(recover.New())
	app.Use(requestIDMiddleware())
	app.Use(accessLogMiddleware(log))

	healthhandler.NewHandler(health.NewChecker(cfg.ServiceName, cfg.ServiceVersion)).RegisterRoutes(app)
	app.Get("/metrics", adaptorHandler(promhttp.Handler()))

	api := app.Group("/api/v1")
	consolehandler.NewHandler(consoleService).RegisterRoutes(api, middleware.JWT(jwtService, false))

	return &Server{app: app, cfg: cfg, log: log}
}

func (s *Server) Listen() error {
	s.log.Info("starting http server", slog.String("addr", s.cfg.HTTPAddr))
	return s.app.Listen(s.cfg.HTTPAddr)
}

func (s *Server) Shutdown() error { return s.app.Shutdown() }

func requestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Set("X-Request-ID", requestID)
		return c.Next()
	}
}

func accessLogMiddleware(log *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		log.Info("http request", slog.String("method", c.Method()), slog.String("path", c.Path()), slog.Int("status", c.Response().StatusCode()), slog.Duration("duration", time.Since(start)))
		return err
	}
}

func defaultErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{"code": "internal_error", "message": err.Error(), "trace_id": c.Get("X-Request-ID")})
}

func adaptorHandler(h http.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fasthttpadaptor.NewFastHTTPHandler(h)(c.Context())
		return nil
	}
}
