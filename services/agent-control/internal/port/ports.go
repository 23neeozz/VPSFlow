package port

import (
	"context"
	"time"

	"github.com/vpsflow/vpsflow/services/agent-control/internal/domain"
)

type Repository interface {
	UpsertSession(ctx context.Context, session domain.Session) error
	GetSessionByHypervisor(ctx context.Context, hypervisorID string) (*domain.Session, error)
	TouchSession(ctx context.Context, sessionID string) error
	CreateCommand(ctx context.Context, cmd domain.Command) error
	GetCommand(ctx context.Context, commandID string) (*domain.Command, error)
	ClaimNextCommand(ctx context.Context, hypervisorID string) (*domain.Command, error)
	CompleteCommand(ctx context.Context, commandID, status string, result []byte, errorMessage string) error
	WaitForCommandCompletion(ctx context.Context, commandID string, timeout time.Duration) (*domain.Command, error)
}

type ClusterClient interface {
	RegisterHypervisor(ctx context.Context, hypervisorID, nodeName, agentVersion string, cpuCores int, memoryBytes, storageBytes int64, runningVMs int) error
	Heartbeat(ctx context.Context, hypervisorID string, cpuCores int, memoryBytes, storageBytes int64, runningVMs int) error
	DrainSiblingHypervisors(ctx context.Context, nodeName, keepHypervisorID string) error
}
