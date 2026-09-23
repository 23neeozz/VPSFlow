package iam

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPClient calls the IAM service over HTTP.
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPClient creates an IAM HTTP client.
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// BootstrapOwner assigns the owner role for a tenant.
func (c *HTTPClient) BootstrapOwner(ctx context.Context, tenantID, userID, authorization string) error {
	body, err := json.Marshal(map[string]string{"user_id": userID})
	if err != nil {
		return fmt.Errorf("marshal bootstrap request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/iam/bootstrap-owner", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create bootstrap request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authorization)
	req.Header.Set("X-Tenant-ID", tenantID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call iam bootstrap: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("iam bootstrap failed: status=%d body=%s", resp.StatusCode, string(payload))
	}
	return nil
}
