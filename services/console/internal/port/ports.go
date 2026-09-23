package port

import (
	"context"

	"github.com/bosscloud/bosscloud/services/console/internal/domain"
)

type Repository interface {
	CreateSession(ctx context.Context, session domain.Session, tokenHash string) error
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error)
	RevokeSession(ctx context.Context, sessionID string) error
}

type VMClient interface {
	GetVM(ctx context.Context, tenantID, vmID, authorization string) (status, hypervisorID, domainName string, err error)
}
