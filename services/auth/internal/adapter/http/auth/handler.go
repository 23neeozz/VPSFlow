package auth

import (
	"strings"

	"github.com/vpsflow/vpsflow/libs/go/errors"
	"github.com/vpsflow/vpsflow/libs/go/httpx"
	"github.com/vpsflow/vpsflow/services/auth/internal/domain"
	"github.com/vpsflow/vpsflow/services/auth/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

// Handler exposes authentication HTTP endpoints.
type Handler struct {
	service *usecase.Service
}

// NewHandler creates an auth HTTP handler.
func NewHandler(service *usecase.Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers auth routes under /api/v1/auth.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	auth := router.Group("/auth")
	auth.Post("/register", h.register)
	auth.Post("/login", h.login)
	auth.Post("/refresh", h.refresh)
	auth.Post("/logout", h.logout)
	auth.Post("/mfa/verify", h.verifyMFA)
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type mfaVerifyRequest struct {
	ChallengeID string `json:"challenge_id"`
	Code        string `json:"code"`
}

type authTokensResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	TokenType    string `json:"token_type"`
	MFARequired  bool   `json:"mfa_required,omitempty"`
	ChallengeID  string `json:"challenge_id,omitempty"`
}

func (h *Handler) register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}

	tokens, err := h.service.Register(c.UserContext(), req.Email, req.Name, req.Password, clientMetadata(c))
	if err != nil {
		return httpx.WriteError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(toResponse(tokens))
}

func (h *Handler) login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}

	tokens, err := h.service.Login(c.UserContext(), req.Email, req.Password, clientMetadata(c))
	if err != nil {
		return httpx.WriteError(c, err)
	}

	return c.JSON(toResponse(tokens))
}

func (h *Handler) refresh(c *fiber.Ctx) error {
	var req refreshRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}

	tokens, err := h.service.Refresh(c.UserContext(), req.RefreshToken)
	if err != nil {
		return httpx.WriteError(c, err)
	}

	return c.JSON(toResponse(tokens))
}

func (h *Handler) logout(c *fiber.Ctx) error {
	accessToken := bearerToken(c.Get("Authorization"))
	if accessToken == "" {
		return httpx.WriteError(c, errors.New(errors.CodeUnauthorized, "missing bearer token"))
	}

	if err := h.service.Logout(c.UserContext(), accessToken); err != nil {
		return httpx.WriteError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) verifyMFA(c *fiber.Ctx) error {
	var req mfaVerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.WriteError(c, errors.New(errors.CodeValidation, "invalid request body"))
	}

	tokens, err := h.service.VerifyMFA(c.UserContext(), req.ChallengeID, req.Code)
	if err != nil {
		return httpx.WriteError(c, err)
	}

	return c.JSON(toResponse(tokens))
}

func clientMetadata(c *fiber.Ctx) usecase.ClientMetadata {
	return usecase.ClientMetadata{
		IPAddress:         c.IP(),
		UserAgent:         c.Get("User-Agent"),
		DeviceFingerprint: c.Get("X-Device-Fingerprint"),
	}
}

func bearerToken(header string) string {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func toResponse(tokens *domain.AuthTokens) authTokensResponse {
	return authTokensResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		TokenType:    tokens.TokenType,
		MFARequired:  tokens.MFARequired,
		ChallengeID:  tokens.ChallengeID,
	}
}
