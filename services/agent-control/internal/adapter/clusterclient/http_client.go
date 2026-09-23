package clusterclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bosscloud/bosscloud/libs/go/agentprotocol"
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
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *HTTPClient) RegisterHypervisor(ctx context.Context, hypervisorID, nodeName, agentVersion string, cpuCores int, memoryBytes, storageBytes int64, runningVMs int) error {
	body, _ := json.Marshal(map[string]any{
		"hypervisor_id": hypervisorID,
		"node_name":     nodeName,
		"agent_version": agentVersion,
		"capacity": agentprotocol.NodeCapacity{
			CPUCores: cpuCores, MemoryBytes: memoryBytes, StorageBytes: storageBytes, RunningVMs: runningVMs,
		},
	})
	return c.post(ctx, "/api/v1/internal/cluster/hypervisors/register", body, nil)
}

func (c *HTTPClient) Heartbeat(ctx context.Context, hypervisorID string, cpuCores int, memoryBytes, storageBytes int64, runningVMs int) error {
	body, _ := json.Marshal(map[string]any{
		"capacity": agentprotocol.NodeCapacity{
			CPUCores: cpuCores, MemoryBytes: memoryBytes, StorageBytes: storageBytes, RunningVMs: runningVMs,
		},
	})
	return c.post(ctx, fmt.Sprintf("/api/v1/internal/cluster/hypervisors/%s/heartbeat", hypervisorID), body, nil)
}

func (c *HTTPClient) DrainSiblingHypervisors(ctx context.Context, nodeName, keepHypervisorID string) error {
	body, _ := json.Marshal(map[string]string{
		"node_name":     nodeName,
		"hypervisor_id": keepHypervisorID,
	})
	return c.post(ctx, "/api/v1/internal/cluster/nodes/drain-siblings", body, nil)
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
