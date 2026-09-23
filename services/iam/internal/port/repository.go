package port

import (
	"context"

	"github.com/bosscloud/bosscloud/services/iam/internal/domain"
)

// Repository persists IAM data.
type Repository interface {
	ListPermissions(ctx context.Context) ([]domain.Permission, error)
	ListRoles(ctx context.Context, tenantID string) ([]domain.Role, error)
	CreateRole(ctx context.Context, role domain.Role) error
	AssignUserRoles(ctx context.Context, tenantID, userID string, roleIDs []string) error
	UserPermissions(ctx context.Context, tenantID, userID string) ([]string, error)
	TenantHasRoles(ctx context.Context, tenantID string) (bool, error)
	BootstrapTenantRoles(ctx context.Context, tenantID string, roles []domain.Role) error
}
