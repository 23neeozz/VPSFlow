package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	apperrors "github.com/vpsflow/vpsflow/libs/go/errors"
	"github.com/vpsflow/vpsflow/libs/go/ids"
	iammigrations "github.com/vpsflow/vpsflow/services/iam/migrations"
	"github.com/vpsflow/vpsflow/services/iam/internal/domain"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Store provides PostgreSQL-backed IAM persistence.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore creates an IAM store.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Up applies embedded migrations.
func Up(databaseURL string) error {
	source, err := iofs.New(iammigrations.FS, ".")
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

func (s *Store) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	rows, err := s.pool.Query(ctx, `SELECT name, description, resource FROM permissions ORDER BY name`)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list permissions", err)
	}
	defer rows.Close()

	permissions := make([]domain.Permission, 0)
	for rows.Next() {
		var permission domain.Permission
		if err := rows.Scan(&permission.Name, &permission.Description, &permission.Resource); err != nil {
			return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan permission", err)
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
}

func (s *Store) ListRoles(ctx context.Context, tenantID string) ([]domain.Role, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT r.role_id, r.tenant_id, r.name, r.description, r.is_system, r.created_at
		FROM roles r
		WHERE r.tenant_id = $1
		ORDER BY r.name
	`, tenantID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list roles", err)
	}
	defer rows.Close()

	roles := make([]domain.Role, 0)
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.TenantID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt); err != nil {
			return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan role", err)
		}
		perms, err := s.rolePermissions(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		role.Permissions = perms
		roles = append(roles, role)
	}
	return roles, nil
}

func (s *Store) CreateRole(ctx context.Context, role domain.Role) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to begin transaction", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	_, err = tx.Exec(ctx, `
		INSERT INTO roles (role_id, tenant_id, name, description, is_system, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, role.ID, role.TenantID, role.Name, role.Description, role.IsSystem, now)
	if err != nil {
		if isUniqueViolation(err) {
			return apperrors.New(apperrors.CodeConflict, "role name already exists")
		}
		return apperrors.Wrap(apperrors.CodeInternal, "failed to create role", err)
	}

	for _, permission := range role.Permissions {
		if _, err := tx.Exec(ctx, `
			INSERT INTO role_permissions (role_id, permission_name) VALUES ($1, $2)
		`, role.ID, permission); err != nil {
			return apperrors.Wrap(apperrors.CodeInternal, "failed to assign role permission", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to commit role creation", err)
	}
	role.CreatedAt = now
	return nil
}

func (s *Store) AssignUserRoles(ctx context.Context, tenantID, userID string, roleIDs []string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to begin transaction", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM user_roles WHERE tenant_id = $1 AND user_id = $2`, tenantID, userID); err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to clear user roles", err)
	}

	for _, roleID := range roleIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO user_roles (user_id, tenant_id, role_id, assigned_at)
			VALUES ($1, $2, $3, $4)
		`, userID, tenantID, roleID, time.Now().UTC()); err != nil {
			return apperrors.Wrap(apperrors.CodeInternal, "failed to assign user role", err)
		}
	}

	return tx.Commit(ctx)
}

func (s *Store) UserPermissions(ctx context.Context, tenantID, userID string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT rp.permission_name
		FROM user_roles ur
		INNER JOIN role_permissions rp ON rp.role_id = ur.role_id
		WHERE ur.tenant_id = $1 AND ur.user_id = $2
		ORDER BY rp.permission_name
	`, tenantID, userID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to load user permissions", err)
	}
	defer rows.Close()

	permissions := make([]string, 0)
	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan permission", err)
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
}

func (s *Store) TenantHasRoles(ctx context.Context, tenantID string) (bool, error) {
	var count int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM roles WHERE tenant_id = $1`, tenantID).Scan(&count); err != nil {
		return false, apperrors.Wrap(apperrors.CodeInternal, "failed to count tenant roles", err)
	}
	return count > 0, nil
}

func (s *Store) BootstrapTenantRoles(ctx context.Context, tenantID string, roles []domain.Role) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to begin transaction", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	for _, role := range roles {
		roleID, err := ids.New("rol")
		if err != nil {
			return apperrors.Wrap(apperrors.CodeInternal, "failed to generate role id", err)
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO roles (role_id, tenant_id, name, description, is_system, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, roleID, tenantID, role.Name, role.Description, role.IsSystem, now)
		if err != nil {
			return apperrors.Wrap(apperrors.CodeInternal, "failed to bootstrap role", err)
		}
		for _, permission := range role.Permissions {
			if _, err := tx.Exec(ctx, `
				INSERT INTO role_permissions (role_id, permission_name) VALUES ($1, $2)
			`, roleID, permission); err != nil {
				return apperrors.Wrap(apperrors.CodeInternal, "failed to bootstrap role permission", err)
			}
		}
	}

	return tx.Commit(ctx)
}

func (s *Store) rolePermissions(ctx context.Context, roleID string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT permission_name FROM role_permissions WHERE role_id = $1 ORDER BY permission_name
	`, roleID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list role permissions", err)
	}
	defer rows.Close()

	permissions := make([]string, 0)
	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan role permission", err)
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
