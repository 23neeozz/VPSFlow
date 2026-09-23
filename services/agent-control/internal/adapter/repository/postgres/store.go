package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	apperrors "github.com/vpsflow/vpsflow/libs/go/errors"
	"github.com/vpsflow/vpsflow/services/agent-control/internal/domain"
	agentmigrations "github.com/vpsflow/vpsflow/services/agent-control/migrations"
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
	source, err := iofs.New(agentmigrations.FS, ".")
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

func (s *Store) UpsertSession(ctx context.Context, session domain.Session) error {
	now := time.Now().UTC()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO agent_sessions (session_id, hypervisor_id, agent_version, node_name, status, last_heartbeat_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (hypervisor_id) DO UPDATE SET
			session_id = EXCLUDED.session_id,
			agent_version = EXCLUDED.agent_version,
			node_name = EXCLUDED.node_name,
			status = EXCLUDED.status,
			last_heartbeat_at = EXCLUDED.last_heartbeat_at,
			updated_at = EXCLUDED.updated_at
	`, session.ID, session.HypervisorID, session.AgentVersion, session.NodeName, session.Status, now, now, now)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to upsert session", err)
	}
	return nil
}

func (s *Store) GetSessionByHypervisor(ctx context.Context, hypervisorID string) (*domain.Session, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT session_id, hypervisor_id, agent_version, node_name, status, last_heartbeat_at, created_at
		FROM agent_sessions WHERE hypervisor_id = $1
	`, hypervisorID)
	var session domain.Session
	if err := row.Scan(&session.ID, &session.HypervisorID, &session.AgentVersion, &session.NodeName, &session.Status, &session.LastHeartbeatAt, &session.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(apperrors.CodeNotFound, "agent session not found")
		}
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to get session", err)
	}
	return &session, nil
}

func (s *Store) TouchSession(ctx context.Context, sessionID string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE agent_sessions SET last_heartbeat_at = NOW(), updated_at = NOW() WHERE session_id = $1
	`, sessionID)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to touch session", err)
	}
	return nil
}

func (s *Store) CreateCommand(ctx context.Context, cmd domain.Command) error {
	now := time.Now().UTC()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO agent_commands (command_id, hypervisor_id, tenant_id, command_type, payload, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, cmd.ID, cmd.HypervisorID, cmd.TenantID, cmd.Type, cmd.Payload, cmd.Status, now, now)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to create command", err)
	}
	return nil
}

func (s *Store) GetCommand(ctx context.Context, commandID string) (*domain.Command, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT command_id, hypervisor_id, tenant_id, command_type, payload, status, result, error_message, created_at, updated_at, completed_at
		FROM agent_commands WHERE command_id = $1
	`, commandID)
	return scanCommand(row)
}

func (s *Store) ClaimNextCommand(ctx context.Context, hypervisorID string) (*domain.Command, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to begin transaction", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
		SELECT command_id, hypervisor_id, tenant_id, command_type, payload, status, result, error_message, created_at, updated_at, completed_at
		FROM agent_commands
		WHERE hypervisor_id = $1 AND status = $2
		ORDER BY created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`, hypervisorID, domain.CommandPending)

	cmd, err := scanCommand(row)
	if err != nil {
		if apperrors.IsCode(err, apperrors.CodeNotFound) {
			return nil, nil
		}
		return nil, err
	}

	_, err = tx.Exec(ctx, `UPDATE agent_commands SET status = $2, updated_at = NOW() WHERE command_id = $1`, cmd.ID, domain.CommandRunning)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to mark command running", err)
	}
	cmd.Status = domain.CommandRunning
	if err := tx.Commit(ctx); err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to commit command claim", err)
	}
	return cmd, nil
}

func (s *Store) CompleteCommand(ctx context.Context, commandID, status string, result []byte, errorMessage string) error {
	now := time.Now().UTC()
	tag, err := s.pool.Exec(ctx, `
		UPDATE agent_commands SET status = $2, result = $3, error_message = $4, updated_at = $5, completed_at = $5
		WHERE command_id = $1
	`, commandID, status, result, errorMessage, now)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to complete command", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.New(apperrors.CodeNotFound, "command not found")
	}
	return nil
}

func (s *Store) WaitForCommandCompletion(ctx context.Context, commandID string, timeout time.Duration) (*domain.Command, error) {
	deadline := time.Now().Add(timeout)
	for {
		cmd, err := s.GetCommand(ctx, commandID)
		if err != nil {
			return nil, err
		}
		if cmd.Status == domain.CommandCompleted || cmd.Status == domain.CommandFailed {
			return cmd, nil
		}
		if time.Now().After(deadline) {
			return cmd, apperrors.New(apperrors.CodeServiceUnavailable, "command execution timed out")
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
	}
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanCommand(row rowScanner) (*domain.Command, error) {
	var cmd domain.Command
	var result []byte
	var completedAt sql.NullTime
	if err := row.Scan(&cmd.ID, &cmd.HypervisorID, &cmd.TenantID, &cmd.Type, &cmd.Payload, &cmd.Status, &result, &cmd.ErrorMessage, &cmd.CreatedAt, &cmd.UpdatedAt, &completedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(apperrors.CodeNotFound, "command not found")
		}
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan command", err)
	}
	cmd.Result = result
	if completedAt.Valid {
		t := completedAt.Time
		cmd.CompletedAt = &t
	}
	return &cmd, nil
}
