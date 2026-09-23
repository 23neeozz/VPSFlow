package http

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	healthhandler "github.com/vpsflow/vpsflow/services/gateway/internal/adapter/http/health"
	gwmiddleware "github.com/vpsflow/vpsflow/services/gateway/internal/adapter/http/middleware"
	"github.com/vpsflow/vpsflow/services/gateway/internal/config"
	domainhealth "github.com/vpsflow/vpsflow/services/gateway/internal/domain/health"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vpsflow/vpsflow/libs/go/security/jwt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// Server wraps the Fiber HTTP application.
type Server struct {
	app *fiber.App
	cfg config.Config
	log *slog.Logger
}

// NewServer constructs the gateway HTTP server with middleware and routes.
func NewServer(cfg config.Config, log *slog.Logger) *Server {
	app := fiber.New(fiber.Config{
		AppName:      cfg.ServiceName,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		ErrorHandler: defaultErrorHandler,
	})

	app.Use(recover.New())
	app.Use(corsMiddleware())
	app.Use(requestIDMiddleware())
	app.Use(accessLogMiddleware(log))
	app.Use(tracingMiddleware(cfg.ServiceName))

	healthChecker := domainhealth.NewChecker(cfg.ServiceName, cfg.ServiceVersion)
	healthHandler := healthhandler.NewHandler(healthChecker)
	healthHandler.RegisterRoutes(app)

	app.Get("/metrics", adaptorHandler(promhttp.Handler()))

	app.Get("/api/v1", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": cfg.ServiceName,
			"version": cfg.ServiceVersion,
			"status":  "operational",
		})
	})

	if cfg.AuthServiceURL != "" {
		authBase := strings.TrimRight(cfg.AuthServiceURL, "/")
		forwardAuth := func(c *fiber.Ctx) error {
			return proxy.Do(c, authBase+c.OriginalURL())
		}
		app.All("/api/v1/auth", forwardAuth)
		app.All("/api/v1/auth/*", forwardAuth)
	}

	if cfg.IAMServiceURL != "" {
		iamBase := strings.TrimRight(cfg.IAMServiceURL, "/")
		forwardIAM := func(c *fiber.Ctx) error {
			return proxy.Do(c, iamBase+c.OriginalURL())
		}
		app.All("/api/v1/iam", forwardIAM)
		app.All("/api/v1/iam/*", forwardIAM)
	}

	if cfg.TenantServiceURL != "" {
		tenantBase := strings.TrimRight(cfg.TenantServiceURL, "/")
		forwardTenant := func(c *fiber.Ctx) error {
			return proxy.Do(c, tenantBase+c.OriginalURL())
		}
		app.All("/api/v1/organizations", forwardTenant)
		app.All("/api/v1/organizations/*", forwardTenant)
	}

	if cfg.ClusterServiceURL != "" {
		clusterBase := strings.TrimRight(cfg.ClusterServiceURL, "/")
		forwardCluster := func(c *fiber.Ctx) error {
			return proxy.Do(c, clusterBase+c.OriginalURL())
		}
		app.All("/api/v1/hypervisors", forwardCluster)
		app.All("/api/v1/hypervisors/*", forwardCluster)
	}

	if cfg.VmServiceURL != "" {
		vmBase := strings.TrimRight(cfg.VmServiceURL, "/")
		forwardVM := func(c *fiber.Ctx) error {
			return proxy.Do(c, vmBase+c.OriginalURL())
		}
		app.All("/api/v1/virtual-machines", forwardVM)
		app.All("/api/v1/virtual-machines/*", forwardVM)
	}

	if cfg.ConsoleServiceURL != "" {
		consoleBase := strings.TrimRight(cfg.ConsoleServiceURL, "/")
		forwardConsole := func(c *fiber.Ctx) error {
			return proxy.Do(c, consoleBase+c.OriginalURL())
		}
		app.All("/api/v1/console", forwardConsole)
		app.All("/api/v1/console/*", forwardConsole)
	}

	if cfg.VpsServiceURL != "" {
		vpsBase := strings.TrimRight(cfg.VpsServiceURL, "/")
		forwardVPS := func(c *fiber.Ctx) error {
			return proxy.Do(c, vpsBase+c.OriginalURL())
		}
		app.All("/api/v1/vps", forwardVPS)
		app.All("/api/v1/vps/*", forwardVPS)
		app.All("/api/v1/admin/vps", forwardVPS)
		app.All("/api/v1/admin/vps/*", forwardVPS)
	}

	if cfg.JWTSigningKey != "" {
		jwtService, err := jwt.NewService(jwt.Config{
			SigningKey: []byte(cfg.JWTSigningKey),
			Issuer:     cfg.JWTIssuer,
			Audience:   cfg.JWTAudience,
		})
		if err == nil {
			protected := app.Group("/api/v1/protected", gwmiddleware.JWT(jwtService))
			protected.Get("/me", func(c *fiber.Ctx) error {
				return c.JSON(fiber.Map{
					"user_id":    c.Locals("user_id"),
					"session_id": c.Locals("session_id"),
					"email":      c.Locals("email"),
				})
			})
		}
	}

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

func corsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")
		if origin != "" {
			c.Set("Access-Control-Allow-Origin", origin)
			c.Set("Vary", "Origin")
			c.Set("Access-Control-Allow-Credentials", "true")
		}
		c.Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Tenant-ID,X-Request-ID,Idempotency-Key")
		c.Set("Access-Control-Max-Age", "600")
		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusNoContent)
		}
		return c.Next()
	}
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
		spanName := c.Method() + " " + c.Path()
		ctx, span := tracer.Start(c.UserContext(), spanName)
		defer span.End()

		c.SetUserContext(ctx)
		err := c.Next()

		statusCode := c.Response().StatusCode()
		routePath := ""
		if route := c.Route(); route != nil {
			routePath = route.Path
		}

		span.SetAttributes(
			attribute.String("http.request.method", c.Method()),
			attribute.String("url.path", c.Path()),
			attribute.Int("http.response.status_code", statusCode),
			attribute.String("http.route", routePath),
			attribute.String("client.address", c.IP()),
			attribute.String("request.id", c.Get("X-Request-ID")),
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
