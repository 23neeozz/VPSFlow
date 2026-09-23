package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	apperrors "github.com/vpsflow/vpsflow/libs/go/errors"
	"github.com/vpsflow/vpsflow/services/vm/internal/domain"
	vmmigrations "github.com/vpsflow/vpsflow/services/vm/migrations"
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
	source, err := iofs.New(vmmigrations.FS, ".")
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

func (s *Store) CreateVM(ctx context.Context, vm domain.VirtualMachine) error {
	now := time.Now().UTC()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO virtual_machines (vm_id, tenant_id, project_id, name, status, hypervisor_id, domain_name, vcpus, memory_mb, disk_gb, image_ref, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`, vm.ID, vm.TenantID, vm.ProjectID, vm.Name, vm.Status, nullString(vm.HypervisorID), nullString(vm.DomainName),
		vm.VCPUs, vm.MemoryMB, vm.DiskGB, vm.ImageRef, now, now)
	if err != nil {
		if isUniqueViolation(err) {
			return apperrors.New(apperrors.CodeConflict, "vm name already exists in tenant")
		}
		return apperrors.Wrap(apperrors.CodeInternal, "failed to create vm", err)
	}
	return nil
}

func (s *Store) GetVM(ctx context.Context, tenantID, vmID string) (*domain.VirtualMachine, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT vm_id, tenant_id, project_id, name, status, COALESCE(hypervisor_id,''), COALESCE(domain_name,''),
			vcpus, memory_mb, disk_gb, image_ref, error_message, created_at, updated_at, deleted_at
		FROM virtual_machines WHERE vm_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, vmID, tenantID)
	return scanVM(row)
}

func (s *Store) ListVMs(ctx context.Context, tenantID string) ([]domain.VirtualMachine, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT vm_id, tenant_id, project_id, name, status, COALESCE(hypervisor_id,''), COALESCE(domain_name,''),
			vcpus, memory_mb, disk_gb, image_ref, error_message, created_at, updated_at, deleted_at
		FROM virtual_machines WHERE tenant_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC
	`, tenantID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list vms", err)
	}
	defer rows.Close()
	result := make([]domain.VirtualMachine, 0)
	for rows.Next() {
		vm, err := scanVM(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *vm)
	}
	return result, rows.Err()
}

func (s *Store) ListPendingVMs(ctx context.Context) ([]domain.VirtualMachine, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT vm_id, tenant_id, project_id, name, status, COALESCE(hypervisor_id,''), COALESCE(domain_name,''),
			vcpus, memory_mb, disk_gb, image_ref, error_message, created_at, updated_at, deleted_at
		FROM virtual_machines WHERE status = $1 AND deleted_at IS NULL ORDER BY created_at ASC
	`, domain.StatusPending)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list pending vms", err)
	}
	defer rows.Close()
	result := make([]domain.VirtualMachine, 0)
	for rows.Next() {
		vm, err := scanVM(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *vm)
	}
	return result, rows.Err()
}

func (s *Store) UpdateVMStatus(ctx context.Context, vmID, status, hypervisorID, domainName, errorMessage string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE virtual_machines SET status = $2,
			hypervisor_id = COALESCE(NULLIF($3,''), hypervisor_id),
			domain_name = COALESCE(NULLIF($4,''), domain_name),
			error_message = $5, updated_at = NOW()
		WHERE vm_id = $1
	`, vmID, status, hypervisorID, domainName, errorMessage)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to update vm status", err)
	}
	return nil
}

func (s *Store) SoftDeleteVM(ctx context.Context, vmID string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE virtual_machines SET status = $2, deleted_at = NOW(), updated_at = NOW() WHERE vm_id = $1
	`, vmID, domain.StatusDeleted)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to delete vm", err)
	}
	return nil
}

func (s *Store) GetOperationByIdempotency(ctx context.Context, tenantID, key string) (*domain.Operation, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT operation_id, vm_id, tenant_id, operation_type, idempotency_key, status, COALESCE(command_id,''), created_at
		FROM vm_operations WHERE tenant_id = $1 AND idempotency_key = $2
	`, tenantID, key)
	var op domain.Operation
	if err := row.Scan(&op.ID, &op.VMID, &op.TenantID, &op.OperationType, &op.IdempotencyKey, &op.Status, &op.CommandID, &op.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to get operation", err)
	}
	return &op, nil
}

func (s *Store) CreateOperation(ctx context.Context, op domain.Operation) error {
	now := time.Now().UTC()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO vm_operations (operation_id, vm_id, tenant_id, operation_type, idempotency_key, status, command_id, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`, op.ID, op.VMID, op.TenantID, op.OperationType, op.IdempotencyKey, op.Status, nullString(op.CommandID), now, now)
	if err != nil {
		if isUniqueViolation(err) {
			return apperrors.New(apperrors.CodeConflict, "duplicate idempotency key")
		}
		return apperrors.Wrap(apperrors.CodeInternal, "failed to create operation", err)
	}
	return nil
}

func (s *Store) UpdateOperation(ctx context.Context, opID, status, commandID string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE vm_operations SET status = $2, command_id = COALESCE(NULLIF($3,''), command_id), updated_at = NOW()
		WHERE operation_id = $1
	`, opID, status, commandID)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to update operation", err)
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanVM(row rowScanner) (*domain.VirtualMachine, error) {
	var vm domain.VirtualMachine
	var deletedAt sql.NullTime
	if err := row.Scan(&vm.ID, &vm.TenantID, &vm.ProjectID, &vm.Name, &vm.Status, &vm.HypervisorID, &vm.DomainName,
		&vm.VCPUs, &vm.MemoryMB, &vm.DiskGB, &vm.ImageRef, &vm.ErrorMessage, &vm.CreatedAt, &vm.UpdatedAt, &deletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(apperrors.CodeNotFound, "virtual machine not found")
		}
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan vm", err)
	}
	if deletedAt.Valid {
		t := deletedAt.Time
		vm.DeletedAt = &t
	}
	return &vm, nil
}

func nullString(v string) sql.NullString {
	if v == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: v, Valid: true}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
