package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	apperrors "github.com/bosscloud/bosscloud/libs/go/errors"
	"github.com/bosscloud/bosscloud/services/cluster/internal/domain"
	clustermigrations "github.com/bosscloud/bosscloud/services/cluster/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
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
	source, err := iofs.New(clustermigrations.FS, ".")
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

func (s *Store) UpsertHypervisor(ctx context.Context, hv domain.Hypervisor) error {
	now := time.Now().UTC()
	clusterID := sql.NullString{}
	if hv.ClusterID != "" {
		clusterID = sql.NullString{String: hv.ClusterID, Valid: true}
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to begin transaction", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO hypervisors (hypervisor_id, cluster_id, node_name, agent_version, status, maintenance_mode, last_heartbeat_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (hypervisor_id) DO UPDATE SET
			cluster_id = COALESCE(EXCLUDED.cluster_id, hypervisors.cluster_id),
			node_name = EXCLUDED.node_name,
			agent_version = EXCLUDED.agent_version,
			status = EXCLUDED.status,
			last_heartbeat_at = EXCLUDED.last_heartbeat_at,
			updated_at = EXCLUDED.updated_at
	`, hv.ID, clusterID, hv.NodeName, hv.AgentVersion, hv.Status, hv.MaintenanceMode, hv.LastHeartbeatAt, now, now)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to upsert hypervisor", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO node_capacity (hypervisor_id, cpu_cores, memory_bytes, storage_bytes, running_vms, allocated_cpu, allocated_memory_bytes, allocated_storage_bytes, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (hypervisor_id) DO UPDATE SET
			cpu_cores = EXCLUDED.cpu_cores,
			memory_bytes = EXCLUDED.memory_bytes,
			storage_bytes = EXCLUDED.storage_bytes,
			running_vms = EXCLUDED.running_vms,
			updated_at = EXCLUDED.updated_at
	`, hv.ID, hv.Capacity.CPUCores, hv.Capacity.MemoryBytes, hv.Capacity.StorageBytes, hv.Capacity.RunningVMs,
		hv.Capacity.AllocatedCPU, hv.Capacity.AllocatedMemoryBytes, hv.Capacity.AllocatedStorageBytes, now)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to upsert node capacity", err)
	}

	return tx.Commit(ctx)
}

func (s *Store) UpdateHypervisorCapacity(ctx context.Context, hypervisorID string, capacity domain.NodeCapacity, status string, heartbeatAt bool) error {
	now := time.Now().UTC()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to begin transaction", err)
	}
	defer tx.Rollback(ctx)

	if heartbeatAt {
		_, err = tx.Exec(ctx, `
			UPDATE hypervisors SET status = $2, last_heartbeat_at = $3, updated_at = $3 WHERE hypervisor_id = $1
		`, hypervisorID, status, now)
	} else {
		_, err = tx.Exec(ctx, `
			UPDATE hypervisors SET status = $2, updated_at = $3 WHERE hypervisor_id = $1
		`, hypervisorID, status, now)
	}
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to update hypervisor status", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE node_capacity SET
			cpu_cores = $2, memory_bytes = $3, storage_bytes = $4, running_vms = $5, updated_at = $6
		WHERE hypervisor_id = $1
	`, hypervisorID, capacity.CPUCores, capacity.MemoryBytes, capacity.StorageBytes, capacity.RunningVMs, now)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to update node capacity", err)
	}

	return tx.Commit(ctx)
}

func (s *Store) GetHypervisor(ctx context.Context, hypervisorID string) (*domain.Hypervisor, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT h.hypervisor_id, COALESCE(h.cluster_id, ''), h.node_name, h.agent_version, h.status,
			h.maintenance_mode, h.last_heartbeat_at, h.created_at, h.updated_at,
			c.cpu_cores, c.memory_bytes, c.storage_bytes, c.running_vms,
			c.allocated_cpu, c.allocated_memory_bytes, c.allocated_storage_bytes
		FROM hypervisors h
		JOIN node_capacity c ON c.hypervisor_id = h.hypervisor_id
		WHERE h.hypervisor_id = $1
	`, hypervisorID)
	return scanHypervisor(row)
}

func (s *Store) ListHypervisors(ctx context.Context, clusterID string, onlineOnly bool) ([]domain.Hypervisor, error) {
	query := `
		SELECT h.hypervisor_id, COALESCE(h.cluster_id, ''), h.node_name, h.agent_version, h.status,
			h.maintenance_mode, h.last_heartbeat_at, h.created_at, h.updated_at,
			c.cpu_cores, c.memory_bytes, c.storage_bytes, c.running_vms,
			c.allocated_cpu, c.allocated_memory_bytes, c.allocated_storage_bytes
		FROM hypervisors h
		JOIN node_capacity c ON c.hypervisor_id = h.hypervisor_id
		WHERE 1=1
	`
	args := []any{}
	argN := 1
	if clusterID != "" {
		query += fmt.Sprintf(" AND h.cluster_id = $%d", argN)
		args = append(args, clusterID)
		argN++
	}
	if onlineOnly {
		query += fmt.Sprintf(" AND h.status = $%d AND h.maintenance_mode = FALSE", argN)
		args = append(args, domain.HypervisorStatusOnline)
	}
	query += " ORDER BY h.node_name ASC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list hypervisors", err)
	}
	defer rows.Close()

	result := make([]domain.Hypervisor, 0)
	for rows.Next() {
		hv, err := scanHypervisor(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *hv)
	}
	return result, rows.Err()
}

func (s *Store) MarkStaleHypervisorsOffline(ctx context.Context, staleBefore time.Time) (int64, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE hypervisors SET status = $1, updated_at = NOW()
		WHERE status = $2 AND (last_heartbeat_at IS NULL OR last_heartbeat_at < $3)
	`, domain.HypervisorStatusOffline, domain.HypervisorStatusOnline, staleBefore)
	if err != nil {
		return 0, apperrors.Wrap(apperrors.CodeInternal, "failed to mark stale hypervisors", err)
	}
	return tag.RowsAffected(), nil
}

func (s *Store) ListClusters(ctx context.Context) ([]domain.Cluster, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT cluster_id, name, description, created_at, updated_at FROM clusters ORDER BY name
	`)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list clusters", err)
	}
	defer rows.Close()

	result := make([]domain.Cluster, 0)
	for rows.Next() {
		var c domain.Cluster
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan cluster", err)
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (s *Store) ReserveCapacity(ctx context.Context, hypervisorID string, cpu int, memoryBytes, storageBytes int64) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE node_capacity SET
			allocated_cpu = allocated_cpu + $2,
			allocated_memory_bytes = allocated_memory_bytes + $3,
			allocated_storage_bytes = allocated_storage_bytes + $4,
			updated_at = NOW()
		WHERE hypervisor_id = $1
			AND (cpu_cores - allocated_cpu) >= $2
			AND (memory_bytes - allocated_memory_bytes) >= $3
			AND (storage_bytes - allocated_storage_bytes) >= $4
	`, hypervisorID, cpu, memoryBytes, storageBytes)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to reserve capacity", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.New(apperrors.CodeConflict, "insufficient hypervisor capacity")
	}
	return nil
}

func (s *Store) ReleaseCapacity(ctx context.Context, hypervisorID string, cpu int, memoryBytes, storageBytes int64) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE node_capacity SET
			allocated_cpu = GREATEST(0, allocated_cpu - $2),
			allocated_memory_bytes = GREATEST(0, allocated_memory_bytes - $3),
			allocated_storage_bytes = GREATEST(0, allocated_storage_bytes - $4),
			updated_at = NOW()
		WHERE hypervisor_id = $1
	`, hypervisorID, cpu, memoryBytes, storageBytes)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to release capacity", err)
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanHypervisor(row rowScanner) (*domain.Hypervisor, error) {
	var hv domain.Hypervisor
	var lastHeartbeat sql.NullTime
	if err := row.Scan(
		&hv.ID, &hv.ClusterID, &hv.NodeName, &hv.AgentVersion, &hv.Status,
		&hv.MaintenanceMode, &lastHeartbeat, &hv.CreatedAt, &hv.UpdatedAt,
		&hv.Capacity.CPUCores, &hv.Capacity.MemoryBytes, &hv.Capacity.StorageBytes, &hv.Capacity.RunningVMs,
		&hv.Capacity.AllocatedCPU, &hv.Capacity.AllocatedMemoryBytes, &hv.Capacity.AllocatedStorageBytes,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(apperrors.CodeNotFound, "hypervisor not found")
		}
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan hypervisor", err)
	}
	if lastHeartbeat.Valid {
		t := lastHeartbeat.Time
		hv.LastHeartbeatAt = &t
	}
	return &hv, nil
}

func (s *Store) MarkSiblingHypervisorsOffline(ctx context.Context, nodeName, keepHypervisorID string) (int64, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE hypervisors SET status = $1, updated_at = NOW()
		WHERE node_name = $2 AND hypervisor_id <> $3 AND status = $4
	`, domain.HypervisorStatusOffline, nodeName, keepHypervisorID, domain.HypervisorStatusOnline)
	if err != nil {
		return 0, apperrors.Wrap(apperrors.CodeInternal, "failed to drain sibling hypervisors", err)
	}
	return tag.RowsAffected(), nil
}