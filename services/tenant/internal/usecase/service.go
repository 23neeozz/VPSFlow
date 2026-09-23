package usecase

import (
	"context"
	"regexp"
	"strings"

	"github.com/vpsflow/vpsflow/libs/go/errors"
	"github.com/vpsflow/vpsflow/libs/go/ids"
	"github.com/vpsflow/vpsflow/services/tenant/internal/domain"
	"github.com/vpsflow/vpsflow/services/tenant/internal/port"
)

type Service struct {
	repo      port.Repository
	iamClient port.IAMClient
}

func NewService(repo port.Repository, iamClient port.IAMClient) *Service {
	return &Service{repo: repo, iamClient: iamClient}
}

func (s *Service) CreateOrganization(ctx context.Context, userID, email, name, slug, authorization string) (*domain.Organization, error) {
	name = strings.TrimSpace(name)
	if userID == "" || name == "" {
		return nil, errors.New(errors.CodeValidation, "name is required")
	}
	if slug == "" {
		slug = slugify(name)
	}

	orgID, err := ids.New("org")
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to generate org id", err)
	}

	org := domain.Organization{ID: orgID, Name: name, Slug: slug}
	if err := s.repo.CreateOrganization(ctx, org, userID, email); err != nil {
		return nil, err
	}

	if err := s.iamClient.BootstrapOwner(ctx, orgID, userID, authorization); err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to bootstrap iam owner role", err)
	}

	created, err := s.repo.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *Service) ListOrganizations(ctx context.Context, userID string) ([]domain.Organization, error) {
	if userID == "" {
		return nil, errors.New(errors.CodeValidation, "user_id is required")
	}
	return s.repo.ListOrganizations(ctx, userID)
}

func (s *Service) GetOrganization(ctx context.Context, orgID, userID string) (*domain.Organization, error) {
	if err := s.ensureMembership(ctx, orgID, userID); err != nil {
		return nil, err
	}
	return s.repo.GetOrganization(ctx, orgID)
}

func (s *Service) ListMembers(ctx context.Context, orgID, userID string) ([]domain.Member, error) {
	if err := s.ensureMembership(ctx, orgID, userID); err != nil {
		return nil, err
	}
	return s.repo.ListMembers(ctx, orgID)
}

func (s *Service) InviteMember(ctx context.Context, orgID, userID, email, role string) error {
	if err := s.ensureMembership(ctx, orgID, userID); err != nil {
		return err
	}
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || role == "" {
		return errors.New(errors.CodeValidation, "email and role are required")
	}
	return s.repo.InviteMember(ctx, orgID, email, role)
}

func (s *Service) CreateProject(ctx context.Context, orgID, userID, name string) (*domain.Project, error) {
	if err := s.ensureMembership(ctx, orgID, userID); err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New(errors.CodeValidation, "name is required")
	}

	projectID, err := ids.New("prj")
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to generate project id", err)
	}

	project := domain.Project{ID: projectID, OrgID: orgID, Name: name}
	if err := s.repo.CreateProject(ctx, project); err != nil {
		return nil, err
	}
	return &project, nil
}

func (s *Service) ListProjects(ctx context.Context, orgID, userID string) ([]domain.Project, error) {
	if err := s.ensureMembership(ctx, orgID, userID); err != nil {
		return nil, err
	}
	return s.repo.ListProjects(ctx, orgID)
}

func (s *Service) ensureMembership(ctx context.Context, orgID, userID string) error {
	ok, err := s.repo.UserBelongsToOrg(ctx, orgID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New(errors.CodeForbidden, "organization access denied")
	}
	return nil
}

var slugPattern = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(value string) string {
	slug := strings.ToLower(strings.TrimSpace(value))
	slug = slugPattern.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}
