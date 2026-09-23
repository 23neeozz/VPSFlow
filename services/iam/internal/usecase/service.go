package usecase

import (
	"context"
	"strings"

	"github.com/bosscloud/bosscloud/libs/go/errors"
	"github.com/bosscloud/bosscloud/libs/go/ids"
	"github.com/bosscloud/bosscloud/services/iam/internal/domain"
	"github.com/bosscloud/bosscloud/services/iam/internal/port"
)

// Service orchestrates IAM operations.
type Service struct {
	repo port.Repository
}

// NewService creates an IAM service.
func NewService(repo port.Repository) *Service {
	return &Service{repo: repo}
}

// ListPermissions returns the global permission catalog.
func (s *Service) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	return s.repo.ListPermissions(ctx)
}

// ListRoles returns roles for a tenant, bootstrapping system roles if needed.
func (s *Service) ListRoles(ctx context.Context, tenantID string) ([]domain.Role, error) {
	if tenantID == "" {
		return nil, errors.New(errors.CodeValidation, "tenant_id is required")
	}

	if err := s.ensureTenantBootstrapped(ctx, tenantID); err != nil {
		return nil, err
	}

	return s.repo.ListRoles(ctx, tenantID)
}

// CreateRole creates a custom tenant role.
func (s *Service) CreateRole(ctx context.Context, tenantID, name, description string, permissions []string) (*domain.Role, error) {
	if tenantID == "" || strings.TrimSpace(name) == "" {
		return nil, errors.New(errors.CodeValidation, "tenant_id and name are required")
	}
	if len(permissions) == 0 {
		return nil, errors.New(errors.CodeValidation, "permissions are required")
	}

	if err := s.ensureTenantBootstrapped(ctx, tenantID); err != nil {
		return nil, err
	}

	roleID, err := ids.New("rol")
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to generate role id", err)
	}

	role := domain.Role{
		ID:          roleID,
		TenantID:    tenantID,
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		IsSystem:    false,
		Permissions: permissions,
	}

	if err := s.repo.CreateRole(ctx, role); err != nil {
		return nil, err
	}

	return &role, nil
}

// AssignUserRoles replaces user role assignments within a tenant.
func (s *Service) AssignUserRoles(ctx context.Context, tenantID, userID string, roleIDs []string) error {
	if tenantID == "" || userID == "" {
		return errors.New(errors.CodeValidation, "tenant_id and user_id are required")
	}
	if len(roleIDs) == 0 {
		return errors.New(errors.CodeValidation, "role_ids are required")
	}

	if err := s.ensureTenantBootstrapped(ctx, tenantID); err != nil {
		return err
	}

	return s.repo.AssignUserRoles(ctx, tenantID, userID, roleIDs)
}

// EvaluatePolicy checks whether a user has a permission within a tenant.
func (s *Service) EvaluatePolicy(ctx context.Context, tenantID, userID, permission string) (domain.PolicyDecision, error) {
	if tenantID == "" || userID == "" || permission == "" {
		return domain.PolicyDecision{}, errors.New(errors.CodeValidation, "tenant_id, user_id and permission are required")
	}

	if err := s.ensureTenantBootstrapped(ctx, tenantID); err != nil {
		return domain.PolicyDecision{}, err
	}

	permissions, err := s.repo.UserPermissions(ctx, tenantID, userID)
	if err != nil {
		return domain.PolicyDecision{}, err
	}

	for _, p := range permissions {
		if p == permission {
			return domain.PolicyDecision{Allowed: true, Reason: "permission granted"}, nil
		}
	}

	return domain.PolicyDecision{Allowed: false, Reason: "permission denied"}, nil
}

// AssignOwner bootstraps tenant roles and assigns the owner role to a user.
func (s *Service) AssignOwner(ctx context.Context, tenantID, userID string) error {
	if tenantID == "" || userID == "" {
		return errors.New(errors.CodeValidation, "tenant_id and user_id are required")
	}
	if err := s.ensureTenantBootstrapped(ctx, tenantID); err != nil {
		return err
	}

	roles, err := s.repo.ListRoles(ctx, tenantID)
	if err != nil {
		return err
	}

	for _, role := range roles {
		if role.Name == "owner" {
			return s.repo.AssignUserRoles(ctx, tenantID, userID, []string{role.ID})
		}
	}

	return errors.New(errors.CodeInternal, "owner role not found")
}

func (s *Service) ensureTenantBootstrapped(ctx context.Context, tenantID string) error {
	hasRoles, err := s.repo.TenantHasRoles(ctx, tenantID)
	if err != nil {
		return err
	}
	if hasRoles {
		return nil
	}
	return s.repo.BootstrapTenantRoles(ctx, tenantID, systemRoles(tenantID))
}

func systemRoles(tenantID string) []domain.Role {
	allPermissions := []string{
		"vm.create", "vm.read", "vm.update", "vm.delete", "vm.console",
		"network.create", "network.read", "network.update", "network.delete",
		"storage.create", "storage.read", "storage.update", "storage.delete",
		"tenant.manage", "tenant.billing", "tenant.members",
		"iam.roles.manage", "iam.users.manage", "audit.read", "admin.hypervisors.manage",
	}

	return []domain.Role{
		{TenantID: tenantID, Name: "owner", Description: "Full tenant access", IsSystem: true, Permissions: allPermissions},
		{
			TenantID: tenantID, Name: "admin", Description: "Tenant administrator", IsSystem: true,
			Permissions: filterOut(allPermissions, "tenant.manage"),
		},
		{
			TenantID: tenantID, Name: "operator", Description: "Infrastructure operator", IsSystem: true,
			Permissions: []string{
				"vm.create", "vm.read", "vm.update", "vm.delete", "vm.console",
				"network.create", "network.read", "network.update", "network.delete",
				"storage.create", "storage.read", "storage.update", "storage.delete",
			},
		},
		{
			TenantID: tenantID, Name: "viewer", Description: "Read-only access", IsSystem: true,
			Permissions: []string{"vm.read", "network.read", "storage.read", "audit.read"},
		},
		{
			TenantID: tenantID, Name: "billing", Description: "Billing and audit access", IsSystem: true,
			Permissions: []string{"tenant.billing", "audit.read"},
		},
	}
}

func filterOut(items []string, skip ...string) []string {
	skipSet := make(map[string]struct{}, len(skip))
	for _, s := range skip {
		skipSet[s] = struct{}{}
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, found := skipSet[item]; !found {
			result = append(result, item)
		}
	}
	return result
}
