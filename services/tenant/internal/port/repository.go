package port

import (
	"context"

	"github.com/bosscloud/bosscloud/services/tenant/internal/domain"
)

type Repository interface {
	CreateOrganization(ctx context.Context, org domain.Organization, ownerID, ownerEmail string) error
	ListOrganizations(ctx context.Context, userID string) ([]domain.Organization, error)
	GetOrganization(ctx context.Context, orgID string) (*domain.Organization, error)
	ListMembers(ctx context.Context, orgID string) ([]domain.Member, error)
	InviteMember(ctx context.Context, orgID, email, role string) error
	CreateProject(ctx context.Context, project domain.Project) error
	ListProjects(ctx context.Context, orgID string) ([]domain.Project, error)
	UserBelongsToOrg(ctx context.Context, orgID, userID string) (bool, error)
}
