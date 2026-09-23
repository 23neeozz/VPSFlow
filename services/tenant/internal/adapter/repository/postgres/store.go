package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	apperrors "github.com/bosscloud/bosscloud/libs/go/errors"
	"github.com/bosscloud/bosscloud/libs/go/ids"
	"github.com/bosscloud/bosscloud/services/tenant/internal/domain"
	tenantmigrations "github.com/bosscloud/bosscloud/services/tenant/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func Up(databaseURL string) error {
	source, err := iofs.New(tenantmigrations.FS, ".")
	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

func (s *Store) CreateOrganization(ctx context.Context, org domain.Organization, ownerID, ownerEmail string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to begin transaction", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	_, err = tx.Exec(ctx, `
		INSERT INTO organizations (org_id, name, slug, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, org.ID, org.Name, org.Slug, now, now)
	if err != nil {
		if isUniqueViolation(err) {
			return apperrors.New(apperrors.CodeConflict, "organization slug already exists")
		}
		return apperrors.Wrap(apperrors.CodeInternal, "failed to create organization", err)
	}

	membershipID, err := ids.New("mem")
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to generate membership id", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO memberships (membership_id, org_id, user_id, email, role, status, created_at)
		VALUES ($1, $2, $3, $4, $5, 'active', $6)
	`, membershipID, org.ID, ownerID, ownerEmail, "owner", now)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to create membership", err)
	}

	return tx.Commit(ctx)
}

func (s *Store) ListOrganizations(ctx context.Context, userID string) ([]domain.Organization, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT o.org_id, o.name, o.slug, o.created_at
		FROM organizations o
		INNER JOIN memberships m ON m.org_id = o.org_id
		WHERE m.user_id = $1 AND m.status = 'active'
		ORDER BY o.created_at DESC
	`, userID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list organizations", err)
	}
	defer rows.Close()

	orgs := make([]domain.Organization, 0)
	for rows.Next() {
		var org domain.Organization
		if err := rows.Scan(&org.ID, &org.Name, &org.Slug, &org.CreatedAt); err != nil {
			return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan organization", err)
		}
		orgs = append(orgs, org)
	}
	return orgs, nil
}

func (s *Store) GetOrganization(ctx context.Context, orgID string) (*domain.Organization, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT org_id, name, slug, created_at FROM organizations WHERE org_id = $1
	`, orgID)
	var org domain.Organization
	if err := row.Scan(&org.ID, &org.Name, &org.Slug, &org.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(apperrors.CodeNotFound, "organization not found")
		}
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to get organization", err)
	}
	return &org, nil
}

func (s *Store) ListMembers(ctx context.Context, orgID string) ([]domain.Member, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT user_id, email, role, status FROM memberships WHERE org_id = $1 ORDER BY created_at
	`, orgID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list members", err)
	}
	defer rows.Close()

	members := make([]domain.Member, 0)
	for rows.Next() {
		var member domain.Member
		if err := rows.Scan(&member.UserID, &member.Email, &member.Role, &member.Status); err != nil {
			return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan member", err)
		}
		members = append(members, member)
	}
	return members, nil
}

func (s *Store) InviteMember(ctx context.Context, orgID, email, role string) error {
	invitationID, err := ids.New("inv")
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to generate invitation id", err)
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO invitations (invitation_id, org_id, email, role, status, created_at)
		VALUES ($1, $2, $3, $4, 'invited', $5)
	`, invitationID, orgID, email, role, time.Now().UTC())
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to create invitation", err)
	}
	return nil
}

func (s *Store) CreateProject(ctx context.Context, project domain.Project) error {
	now := time.Now().UTC()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO projects (project_id, org_id, name, created_at)
		VALUES ($1, $2, $3, $4)
	`, project.ID, project.OrgID, project.Name, now)
	if err != nil {
		if isUniqueViolation(err) {
			return apperrors.New(apperrors.CodeConflict, "project name already exists")
		}
		return apperrors.Wrap(apperrors.CodeInternal, "failed to create project", err)
	}
	project.CreatedAt = now
	return nil
}

func (s *Store) ListProjects(ctx context.Context, orgID string) ([]domain.Project, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT project_id, org_id, name, created_at FROM projects WHERE org_id = $1 ORDER BY created_at DESC
	`, orgID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list projects", err)
	}
	defer rows.Close()

	projects := make([]domain.Project, 0)
	for rows.Next() {
		var project domain.Project
		if err := rows.Scan(&project.ID, &project.OrgID, &project.Name, &project.CreatedAt); err != nil {
			return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan project", err)
		}
		projects = append(projects, project)
	}
	return projects, nil
}

func (s *Store) UserBelongsToOrg(ctx context.Context, orgID, userID string) (bool, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM memberships WHERE org_id = $1 AND user_id = $2 AND status = 'active'
	`, orgID, userID).Scan(&count)
	if err != nil {
		return false, apperrors.Wrap(apperrors.CodeInternal, "failed to check membership", err)
	}
	return count > 0, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
