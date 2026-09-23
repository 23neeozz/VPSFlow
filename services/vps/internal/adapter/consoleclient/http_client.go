package consoleclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	apperrors "github.com/bosscloud/bosscloud/libs/go/errors"
	"github.com/bosscloud/bosscloud/services/vps/internal/domain"
)

type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *HTTPClient) CreateSession(ctx context.Context, tenantID, vmID, authorization string) (*domain.ConsoleSession, error) {
	body, _ := json.Marshal(map[string]string{"vm_id": vmID})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/console/sessions", bytes.NewReader(body))
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to create console request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authorization)
	req.Header.Set("X-Tenant-ID", tenantID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeServiceUnavailable, "console service unavailable", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, apperrors.New(apperrors.CodeInternal, string(raw))
	}
	var session domain.ConsoleSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to decode console response", err)
	}
	return &session, nil
}
