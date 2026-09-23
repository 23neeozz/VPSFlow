package httpx

import (
	"testing"

	"github.com/bosscloud/bosscloud/libs/go/errors"
	"github.com/gofiber/fiber/v2"
)

func TestStatusFromCode(t *testing.T) {
	cases := []struct {
		code   errors.Code
		status int
	}{
		{errors.CodeNotFound, fiber.StatusNotFound},
		{errors.CodeUnauthorized, fiber.StatusUnauthorized},
		{errors.CodeInternal, fiber.StatusInternalServerError},
	}
	for _, tc := range cases {
		if got := statusFromCode(tc.code); got != tc.status {
			t.Fatalf("code %s: expected %d got %d", tc.code, tc.status, got)
		}
	}
}
