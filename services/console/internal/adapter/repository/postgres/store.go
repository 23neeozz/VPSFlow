package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	apperrors "github.com/bosscloud/bosscloud/libs/go/errors"
	"github.com/bosscloud/bosscloud/services/console/internal/domain"
	consolemigrations "github.com/bosscloud/bosscloud/services/console/migrations"
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
	source, err := iofs.New(consolemigrations.FS, ".")
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

func (s *Store) CreateSession(ctx context.Context, session domain.Session, tokenHash string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO console_sessions (session_id, tenant_id, vm_id, user_id, token_hash, status, expires_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, session.ID, session.TenantID, session.VMID, session.UserID, tokenHash, session.Status, session.ExpiresAt, session.CreatedAt)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to create console session", err)
	}
	return nil
}

func (s *Store) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT session_id, tenant_id, vm_id, user_id, status, expires_at, created_at
		FROM console_sessions WHERE token_hash = $1 AND revoked_at IS NULL
	`, tokenHash)
	var session domain.Session
	if err := row.Scan(&session.ID, &session.TenantID, &session.VMID, &session.UserID, &session.Status, &session.ExpiresAt, &session.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(apperrors.CodeNotFound, "console session not found")
		}
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to get console session", err)
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, apperrors.New(apperrors.CodeUnauthorized, "console session expired")
	}
	return &session, nil
}

func (s *Store) RevokeSession(ctx context.Context, sessionID string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE console_sessions SET status = $2, revoked_at = NOW() WHERE session_id = $1
	`, sessionID, domain.SessionRevoked)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to revoke console session", err)
	}
	return nil
}
