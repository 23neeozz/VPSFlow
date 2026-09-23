package tenantclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	apperrors "github.com/vpsflow/vpsflow/libs/go/errors"
)

type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *HTTPClient) GetMemberRole(ctx context.Context, tenantID, userID, authorization string) (string, error) {
	members, err := c.listMembers(ctx, tenantID, authorization)
	if err != nil {
		return "", err
	}
	for _, m := range members {
		if m.UserID == userID {
			return m.Role, nil
		}
	}
	return "", apperrors.New(apperrors.CodeForbidden, "user is not a member of this organization")
}

func (c *HTTPClient) ResolveMemberByEmail(ctx context.Context, tenantID, email, authorization string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	members, err := c.listMembers(ctx, tenantID, authorization)
	if err != nil {
		return "", err
	}
	for _, m := range members {
		if strings.EqualFold(m.Email, email) {
			return m.UserID, nil
		}
	}
	return "", apperrors.New(apperrors.CodeNotFound, "no member found with that email in this organization")
}

func (c *HTTPClient) listMembers(ctx context.Context, tenantID, authorization string) ([]member, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/api/v1/organizations/%s/members", c.baseURL, tenantID), nil)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to create tenant request", err)
	}
	req.Header.Set("Authorization", authorization)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeServiceUnavailable, "tenant service unavailable", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, apperrors.New(apperrors.CodeForbidden, "cannot list organization members")
	}
	var payload struct {
		Data []member `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to decode members", err)
	}
	return payload.Data, nil
}

type member struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Status string `json:"status"`
}
