package agentcontrolclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/vpsflow/vpsflow/libs/go/agentprotocol"
	apperrors "github.com/vpsflow/vpsflow/libs/go/errors"
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
		httpClient: &http.Client{Timeout: 3 * time.Minute},
	}
}

func (c *HTTPClient) DispatchCreateDomain(ctx context.Context, hypervisorID, tenantID string, payload any, wait bool) (string, string, string, error) {
	return c.dispatch(ctx, hypervisorID, tenantID, agentprotocol.CommandCreateDomain, payload, wait)
}

func (c *HTTPClient) DispatchStartDomain(ctx context.Context, hypervisorID, tenantID, domainName string, wait bool) (string, string, string, error) {
	return c.dispatch(ctx, hypervisorID, tenantID, agentprotocol.CommandStartDomain, agentprotocol.DomainActionPayload{DomainName: domainName}, wait)
}

func (c *HTTPClient) DispatchStopDomain(ctx context.Context, hypervisorID, tenantID, domainName string, wait bool) (string, string, string, error) {
	return c.dispatch(ctx, hypervisorID, tenantID, agentprotocol.CommandStopDomain, agentprotocol.DomainActionPayload{DomainName: domainName}, wait)
}

func (c *HTTPClient) DispatchDestroyDomain(ctx context.Context, hypervisorID, tenantID, domainName string, wait bool) (string, string, string, error) {
	return c.dispatch(ctx, hypervisorID, tenantID, agentprotocol.CommandDestroyDomain, agentprotocol.DomainActionPayload{DomainName: domainName}, wait)
}

func (c *HTTPClient) dispatch(ctx context.Context, hypervisorID, tenantID, commandType string, payload any, wait bool) (string, string, string, error) {
	body, _ := json.Marshal(map[string]any{
		"hypervisor_id": hypervisorID,
		"tenant_id":     tenantID,
		"type":          commandType,
		"payload":       payload,
		"wait":          wait,
	})
	var resp struct {
		CommandID string `json:"command_id"`
		Status    string `json:"status"`
		Error     string `json:"error"`
	}
	if err := c.post(ctx, "/api/v1/internal/agent-control/commands", body, &resp); err != nil {
		return "", "", "", err
	}
	return resp.CommandID, resp.Status, resp.Error, nil
}

func (c *HTTPClient) post(ctx context.Context, path string, body []byte, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to create agent-control request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-API-Key", c.apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeServiceUnavailable, "agent-control service unavailable", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(resp.Body)
		return apperrors.New(apperrors.CodeInternal, fmt.Sprintf("agent-control request failed: %s", string(payload)))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
