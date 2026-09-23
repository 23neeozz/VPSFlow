package middleware

import (
	"strings"

	"github.com/bosscloud/bosscloud/libs/go/errors"
	"github.com/bosscloud/bosscloud/libs/go/httpx"
	"github.com/bosscloud/bosscloud/libs/go/security/jwt"
	"github.com/gofiber/fiber/v2"
)

// JWTContext holds authenticated request context.
type JWTContext struct {
	UserID    string
	SessionID string
	Email     string
	TenantID  string
}

const jwtContextKey = "jwt_context"

// JWT validates Bearer tokens and optional X-Tenant-ID header.
func JWT(jwtService *jwt.Service, requireTenant bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := bearerToken(c.Get("Authorization"))
		if token == "" {
			return httpx.WriteError(c, errors.New(errors.CodeUnauthorized, "missing bearer token"))
		}

		claims, err := jwtService.ParseAccessToken(token)
		if err != nil {
			return httpx.WriteError(c, err)
		}

		tenantID := strings.TrimSpace(c.Get("X-Tenant-ID"))
		if requireTenant && tenantID == "" {
			return httpx.WriteError(c, errors.New(errors.CodeValidation, "X-Tenant-ID header is required"))
		}

		ctx := JWTContext{
			UserID:    claims.UserID,
			SessionID: claims.SessionID,
			Email:     claims.Email,
			TenantID:  tenantID,
		}
		c.Locals(jwtContextKey, ctx)
		c.Locals("user_id", ctx.UserID)
		c.Locals("tenant_id", ctx.TenantID)
		return c.Next()
	}
}

// Context extracts JWT context from the request.
func Context(c *fiber.Ctx) (JWTContext, bool) {
	value, ok := c.Locals(jwtContextKey).(JWTContext)
	return value, ok
}

func bearerToken(header string) string {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
