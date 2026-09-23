package controlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bosscloud/bosscloud/libs/go/agentprotocol"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{Timeout: 35 * time.Second},
	}
}

func (c *Client) Register(ctx context.Context, req agentprotocol.RegisterAgentRequest) (*agentprotocol.RegisterAgentResponse, error) {
	body, _ := json.Marshal(req)
	var resp agentprotocol.RegisterAgentResponse
	if err := c.post(ctx, "/agent/v1/register", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Heartbeat(ctx context.Context, req agentprotocol.HeartbeatRequest) error {
	body, _ := json.Marshal(req)
	return c.post(ctx, "/agent/v1/heartbeat", body, nil)
}

func (c *Client) PollCommand(ctx context.Context, hypervisorID, sessionID string) (*agentprotocol.Command, error) {
	url := fmt.Sprintf("%s/agent/v1/commands/poll?hypervisor_id=%s&session_id=%s", c.baseURL, hypervisorID, sessionID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}
	if resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("poll failed: %s", string(payload))
	}
	var cmd agentprotocol.Command
	if err := json.NewDecoder(resp.Body).Decode(&cmd); err != nil {
		return nil, err
	}
	return &cmd, nil
}

func (c *Client) CompleteCommand(ctx context.Context, result agentprotocol.CommandResult) error {
	body, _ := json.Marshal(result)
	return c.post(ctx, fmt.Sprintf("/agent/v1/commands/%s/complete", result.CommandID), body, nil)
}

func (c *Client) post(ctx context.Context, path string, body []byte, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed: %s", string(payload))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
