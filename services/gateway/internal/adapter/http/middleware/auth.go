package middleware

import (
	"strings"

	"github.com/bosscloud/bosscloud/libs/go/errors"
	"github.com/bosscloud/bosscloud/libs/go/httpx"
	"github.com/bosscloud/bosscloud/libs/go/security/jwt"
	"github.com/gofiber/fiber/v2"
)

// JWT validates Bearer access tokens and injects claims into request context.
func JWT(jwtService *jwt.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := bearerToken(c.Get("Authorization"))
		if token == "" {
			return httpx.WriteError(c, errors.New(errors.CodeUnauthorized, "missing bearer token"))
		}

		claims, err := jwtService.ParseAccessToken(token)
		if err != nil {
			return httpx.WriteError(c, err)
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("session_id", claims.SessionID)
		c.Locals("email", claims.Email)
		return c.Next()
	}
}

func bearerToken(header string) string {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
