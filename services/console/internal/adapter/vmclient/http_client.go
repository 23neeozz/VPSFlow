package vmclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	apperrors "github.com/bosscloud/bosscloud/libs/go/errors"
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

func (c *HTTPClient) GetVM(ctx context.Context, tenantID, vmID, authorization string) (string, string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/virtual-machines/%s", c.baseURL, vmID), nil)
	if err != nil {
		return "", "", "", apperrors.Wrap(apperrors.CodeInternal, "failed to create vm request", err)
	}
	req.Header.Set("Authorization", authorization)
	req.Header.Set("X-Tenant-ID", tenantID)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", "", apperrors.Wrap(apperrors.CodeServiceUnavailable, "vm service unavailable", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", "", "", apperrors.New(apperrors.CodeNotFound, "virtual machine not found")
	}
	if resp.StatusCode >= 300 {
		return "", "", "", apperrors.New(apperrors.CodeInternal, "vm request failed")
	}
	var vm struct {
		Status       string `json:"status"`
		HypervisorID string `json:"hypervisor_id"`
		DomainName   string `json:"domain_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&vm); err != nil {
		return "", "", "", apperrors.Wrap(apperrors.CodeInternal, "failed to decode vm response", err)
	}
	return vm.Status, vm.HypervisorID, vm.DomainName, nil
}
