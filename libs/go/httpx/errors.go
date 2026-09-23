package httpx

import (
	"github.com/bosscloud/bosscloud/libs/go/errors"
	"github.com/gofiber/fiber/v2"
)

// ErrorResponse is the canonical REST error envelope.
type ErrorResponse struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
	TraceID string         `json:"trace_id,omitempty"`
}

// WriteError maps AppError to HTTP status and JSON response.
func WriteError(c *fiber.Ctx, err error) error {
	traceID := c.Get("X-Request-ID")

	if appErr, ok := errors.AsAppError(err); ok {
		status := statusFromCode(appErr.Code)
		return c.Status(status).JSON(ErrorResponse{
			Code:    string(appErr.Code),
			Message: appErr.Message,
			Details: appErr.Details,
			TraceID: traceID,
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
		Code:    string(errors.CodeInternal),
		Message: "An unexpected error occurred",
		TraceID: traceID,
	})
}

func statusFromCode(code errors.Code) int {
	switch code {
	case errors.CodeValidation:
		return fiber.StatusBadRequest
	case errors.CodeUnauthorized:
		return fiber.StatusUnauthorized
	case errors.CodeForbidden:
		return fiber.StatusForbidden
	case errors.CodeNotFound:
		return fiber.StatusNotFound
	case errors.CodeConflict:
		return fiber.StatusConflict
	case errors.CodeRateLimited:
		return fiber.StatusTooManyRequests
	case errors.CodeServiceUnavailable:
		return fiber.StatusServiceUnavailable
	default:
		return fiber.StatusInternalServerError
	}
}
