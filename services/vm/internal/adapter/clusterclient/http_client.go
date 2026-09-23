package clusterclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	apperrors "github.com/bosscloud/bosscloud/libs/go/errors"
)

type HTTPClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewHTTPClient(baseURL, apiKey string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *HTTPClient) SelectHypervisor(ctx context.Context, vcpus, memoryMB, diskGB int, clusterID string) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"vcpus": vcpus, "memory_mb": memoryMB, "disk_gb": diskGB, "cluster_id": clusterID,
	})
	var resp struct {
		ID string `json:"id"`
	}
	if err := c.post(ctx, "/api/v1/internal/cluster/placement/select", body, &resp); err != nil {
		return "", err
	}
	if resp.ID == "" {
		return "", apperrors.New(apperrors.CodeInternal, "cluster returned empty hypervisor id")
	}
	return resp.ID, nil
}

func (c *HTTPClient) ReleaseCapacity(ctx context.Context, hypervisorID string, vcpus, memoryMB, diskGB int) error {
	body, _ := json.Marshal(map[string]any{
		"vcpus": vcpus, "memory_mb": memoryMB, "disk_gb": diskGB,
	})
	return c.post(ctx, fmt.Sprintf("/api/v1/internal/cluster/hypervisors/%s/release-capacity", hypervisorID), body, nil)
}

func (c *HTTPClient) post(ctx context.Context, path string, body []byte, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to create cluster request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-API-Key", c.apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeServiceUnavailable, "cluster service unavailable", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(resp.Body)
		return apperrors.New(apperrors.CodeInternal, fmt.Sprintf("cluster request failed: %s", string(payload)))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
