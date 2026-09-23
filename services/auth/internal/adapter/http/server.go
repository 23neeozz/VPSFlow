package http

import (
	"log/slog"
	"net/http"
	"time"

	authhandler "github.com/bosscloud/bosscloud/services/auth/internal/adapter/http/auth"
	"github.com/bosscloud/bosscloud/services/auth/internal/config"
	"github.com/bosscloud/bosscloud/services/auth/internal/domain/health"
	healthhandler "github.com/bosscloud/bosscloud/services/auth/internal/adapter/http/health"
	"github.com/bosscloud/bosscloud/services/auth/internal/usecase"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// Server wraps the auth HTTP application.
type Server struct {
	app *fiber.App
	cfg config.Config
	log *slog.Logger
}

// NewServer constructs the auth HTTP server.
func NewServer(cfg config.Config, log *slog.Logger, authService *usecase.Service) *Server {
	app := fiber.New(fiber.Config{
		AppName:      cfg.ServiceName,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		ErrorHandler: defaultErrorHandler,
	})

	app.Use(recover.New())
	app.Use(requestIDMiddleware())
	app.Use(accessLogMiddleware(log))
	app.Use(tracingMiddleware(cfg.ServiceName))

	checker := health.NewChecker(cfg.ServiceName, cfg.ServiceVersion)
	healthhandler.NewHandler(checker).RegisterRoutes(app)

	app.Get("/metrics", adaptorHandler(promhttp.Handler()))

	api := app.Group("/api/v1")
	authhandler.NewHandler(authService).RegisterRoutes(api)

	return &Server{app: app, cfg: cfg, log: log}
}

// Listen starts the HTTP server.
func (s *Server) Listen() error {
	s.log.Info("starting http server", slog.String("addr", s.cfg.HTTPAddr))
	return s.app.Listen(s.cfg.HTTPAddr)
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}

// App exposes the underlying Fiber app for testing.
func (s *Server) App() *fiber.App {
	return s.app
}

func requestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Set("X-Request-ID", requestID)
		c.Locals("request_id", requestID)
		return c.Next()
	}
}

func accessLogMiddleware(log *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		requestID, _ := c.Locals("request_id").(string)
		log.Info("http request",
			slog.String("request_id", requestID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("duration", time.Since(start)),
		)
		return err
	}
}

func tracingMiddleware(serviceName string) fiber.Handler {
	tracer := otel.Tracer(serviceName)
	return func(c *fiber.Ctx) error {
		ctx, span := tracer.Start(c.UserContext(), c.Method()+" "+c.Path())
		defer span.End()
		c.SetUserContext(ctx)
		err := c.Next()
		span.SetAttributes(
			attribute.String("http.request.method", c.Method()),
			attribute.String("url.path", c.Path()),
			attribute.Int("http.response.status_code", c.Response().StatusCode()),
		)
		if err != nil {
			span.RecordError(err)
		}
		return err
	}
}

func defaultErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{
		"code":     "internal_error",
		"message":  err.Error(),
		"trace_id": c.Get("X-Request-ID"),
	})
}

func adaptorHandler(h http.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fasthttpadaptor.NewFastHTTPHandler(h)(c.Context())
		return nil
	}
}
