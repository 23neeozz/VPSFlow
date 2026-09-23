package domain

import "time"

const (
	SessionActive  = "active"
	SessionExpired = "expired"
	SessionRevoked = "revoked"
)

type Session struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	VMID      string    `json:"vm_id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type SessionToken struct {
	SessionID string    `json:"session_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	ProxyURL  string    `json:"proxy_url"`
}
