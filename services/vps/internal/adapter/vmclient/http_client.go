package vmclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	apperrors "github.com/vpsflow/vpsflow/libs/go/errors"
	"github.com/vpsflow/vpsflow/services/vps/internal/port"
)

type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *HTTPClient) CreateVM(ctx context.Context, tenantID, authorization, idempotencyKey string, req port.CreateVMRequest) (*port.VMInfo, error) {
	body, _ := json.Marshal(map[string]any{
		"name":      req.Name,
		"vcpus":     req.VCPUs,
		"memory_mb": req.MemoryMB,
		"disk_gb":   req.DiskGB,
		"image_ref": req.ImageRef,
	})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/virtual-machines", bytes.NewReader(body))
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to create vm request", err)
	}
	c.setHeaders(httpReq, tenantID, authorization, idempotencyKey)
	return c.doVM(httpReq)
}

func (c *HTTPClient) GetVM(ctx context.Context, tenantID, vmID, authorization string) (*port.VMInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/virtual-machines/%s", c.baseURL, vmID), nil)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to create vm request", err)
	}
	c.setHeaders(httpReq, tenantID, authorization, "")
	return c.doVM(httpReq)
}

func (c *HTTPClient) StartVM(ctx context.Context, tenantID, vmID, authorization string) (*port.VMInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/api/v1/virtual-machines/%s/start", c.baseURL, vmID), nil)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to create vm request", err)
	}
	c.setHeaders(httpReq, tenantID, authorization, "")
	return c.doVM(httpReq)
}

func (c *HTTPClient) StopVM(ctx context.Context, tenantID, vmID, authorization string) (*port.VMInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/api/v1/virtual-machines/%s/stop", c.baseURL, vmID), nil)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to create vm request", err)
	}
	c.setHeaders(httpReq, tenantID, authorization, "")
	return c.doVM(httpReq)
}

func (c *HTTPClient) DeleteVM(ctx context.Context, tenantID, vmID, authorization string) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/api/v1/virtual-machines/%s", c.baseURL, vmID), nil)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to create vm request", err)
	}
	c.setHeaders(httpReq, tenantID, authorization, "")
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeServiceUnavailable, "vm service unavailable", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 && resp.StatusCode != http.StatusNoContent {
		return decodeAPIError(resp)
	}
	return nil
}

func (c *HTTPClient) setHeaders(req *http.Request, tenantID, authorization, idempotencyKey string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authorization)
	req.Header.Set("X-Tenant-ID", tenantID)
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
}

func (c *HTTPClient) doVM(req *http.Request) (*port.VMInfo, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeServiceUnavailable, "vm service unavailable", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, apperrors.New(apperrors.CodeNotFound, "virtual machine not found")
	}
	if resp.StatusCode >= 300 {
		return nil, decodeAPIError(resp)
	}
	var vm struct {
		ID           string `json:"id"`
		Status       string `json:"status"`
		VCPUs        int    `json:"vcpus"`
		MemoryMB     int    `json:"memory_mb"`
		DiskGB       int    `json:"disk_gb"`
		ErrorMessage string `json:"error_message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&vm); err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to decode vm response", err)
	}
	return &port.VMInfo{
		ID: vm.ID, Status: vm.Status, VCPUs: vm.VCPUs,
		MemoryMB: vm.MemoryMB, DiskGB: vm.DiskGB, ErrorMessage: vm.ErrorMessage,
	}, nil
}

func decodeAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	var apiErr struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(body, &apiErr)
	if apiErr.Message == "" {
		apiErr.Message = string(body)
	}
	code := apperrors.CodeInternal
	switch resp.StatusCode {
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		code = apperrors.CodeValidation
	case http.StatusUnauthorized:
		code = apperrors.CodeUnauthorized
	case http.StatusForbidden:
		code = apperrors.CodeForbidden
	case http.StatusNotFound:
		code = apperrors.CodeNotFound
	case http.StatusConflict:
		code = apperrors.CodeConflict
	}
	return apperrors.New(code, apiErr.Message)
}
