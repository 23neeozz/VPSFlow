package postgres

import (
	"context"
	"errors"
	"time"

	apperrors "github.com/vpsflow/vpsflow/libs/go/errors"
	"github.com/vpsflow/vpsflow/services/vps/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Create(ctx context.Context, vps domain.Instance) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO vps_instances (
			vps_id, tenant_id, vm_id, name, owner_user_id, owner_email,
			assigned_by_user_id, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`, vps.ID, vps.TenantID, vps.VMID, vps.Name, vps.OwnerUserID, vps.OwnerEmail,
		vps.AssignedByUserID, vps.CreatedAt, vps.UpdatedAt)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to create vps", err)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, tenantID, vpsID string) (*domain.Instance, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT vps_id, tenant_id, vm_id, name, owner_user_id, owner_email,
		       assigned_by_user_id, created_at, updated_at, deleted_at
		FROM vps_instances
		WHERE tenant_id = $1 AND vps_id = $2 AND deleted_at IS NULL
	`, tenantID, vpsID)
	return scanInstance(row)
}

func (s *Store) ListByTenant(ctx context.Context, tenantID string) ([]domain.Instance, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT vps_id, tenant_id, vm_id, name, owner_user_id, owner_email,
		       assigned_by_user_id, created_at, updated_at, deleted_at
		FROM vps_instances
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, tenantID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list vps", err)
	}
	defer rows.Close()
	return scanInstances(rows)
}

func (s *Store) ListByOwner(ctx context.Context, tenantID, ownerUserID string) ([]domain.Instance, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT vps_id, tenant_id, vm_id, name, owner_user_id, owner_email,
		       assigned_by_user_id, created_at, updated_at, deleted_at
		FROM vps_instances
		WHERE tenant_id = $1 AND owner_user_id = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, tenantID, ownerUserID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list owner vps", err)
	}
	defer rows.Close()
	return scanInstances(rows)
}

func (s *Store) UpdateOwner(ctx context.Context, tenantID, vpsID, ownerUserID, ownerEmail string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE vps_instances
		SET owner_user_id = $3, owner_email = $4, updated_at = $5
		WHERE tenant_id = $1 AND vps_id = $2 AND deleted_at IS NULL
	`, tenantID, vpsID, ownerUserID, ownerEmail, time.Now().UTC())
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to update vps owner", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.New(apperrors.CodeNotFound, "vps not found")
	}
	return nil
}

func (s *Store) SoftDelete(ctx context.Context, tenantID, vpsID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE vps_instances SET deleted_at = $3, updated_at = $3
		WHERE tenant_id = $1 AND vps_id = $2 AND deleted_at IS NULL
	`, tenantID, vpsID, time.Now().UTC())
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to delete vps", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.New(apperrors.CodeNotFound, "vps not found")
	}
	return nil
}

func scanInstance(row pgx.Row) (*domain.Instance, error) {
	var v domain.Instance
	var deletedAt *time.Time
	err := row.Scan(
		&v.ID, &v.TenantID, &v.VMID, &v.Name, &v.OwnerUserID, &v.OwnerEmail,
		&v.AssignedByUserID, &v.CreatedAt, &v.UpdatedAt, &deletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.New(apperrors.CodeNotFound, "vps not found")
	}
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan vps", err)
	}
	v.DeletedAt = deletedAt
	return &v, nil
}

func scanInstances(rows pgx.Rows) ([]domain.Instance, error) {
	out := make([]domain.Instance, 0)
	for rows.Next() {
		var v domain.Instance
		var deletedAt *time.Time
		if err := rows.Scan(
			&v.ID, &v.TenantID, &v.VMID, &v.Name, &v.OwnerUserID, &v.OwnerEmail,
			&v.AssignedByUserID, &v.CreatedAt, &v.UpdatedAt, &deletedAt,
		); err != nil {
			return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan vps row", err)
		}
		v.DeletedAt = deletedAt
		out = append(out, v)
	}
	return out, nil
}
