package port

import (
	"context"

	"github.com/bosscloud/bosscloud/services/vps/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, vps domain.Instance) error
	Get(ctx context.Context, tenantID, vpsID string) (*domain.Instance, error)
	ListByTenant(ctx context.Context, tenantID string) ([]domain.Instance, error)
	ListByOwner(ctx context.Context, tenantID, ownerUserID string) ([]domain.Instance, error)
	UpdateOwner(ctx context.Context, tenantID, vpsID, ownerUserID, ownerEmail string) error
	SoftDelete(ctx context.Context, tenantID, vpsID string) error
}

type VMClient interface {
	CreateVM(ctx context.Context, tenantID, authorization, idempotencyKey string, req CreateVMRequest) (*VMInfo, error)
	GetVM(ctx context.Context, tenantID, vmID, authorization string) (*VMInfo, error)
	StartVM(ctx context.Context, tenantID, vmID, authorization string) (*VMInfo, error)
	StopVM(ctx context.Context, tenantID, vmID, authorization string) (*VMInfo, error)
	DeleteVM(ctx context.Context, tenantID, vmID, authorization string) error
}

type CreateVMRequest struct {
	Name     string
	VCPUs    int
	MemoryMB int
	DiskGB   int
	ImageRef string
}

type VMInfo struct {
	ID           string
	Status       string
	VCPUs        int
	MemoryMB     int
	DiskGB       int
	ErrorMessage string
}

type ConsoleClient interface {
	CreateSession(ctx context.Context, tenantID, vmID, authorization string) (*domain.ConsoleSession, error)
}

type TenantClient interface {
	GetMemberRole(ctx context.Context, tenantID, userID, authorization string) (string, error)
	ResolveMemberByEmail(ctx context.Context, tenantID, email, authorization string) (userID string, err error)
}
