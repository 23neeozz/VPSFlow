package port

import (
	"context"
	"time"

	"github.com/bosscloud/bosscloud/services/cluster/internal/domain"
)

type Repository interface {
	UpsertHypervisor(ctx context.Context, hv domain.Hypervisor) error
	UpdateHypervisorCapacity(ctx context.Context, hypervisorID string, capacity domain.NodeCapacity, status string, heartbeatAt bool) error
	GetHypervisor(ctx context.Context, hypervisorID string) (*domain.Hypervisor, error)
	ListHypervisors(ctx context.Context, clusterID string, onlineOnly bool) ([]domain.Hypervisor, error)
	MarkStaleHypervisorsOffline(ctx context.Context, staleBefore time.Time) (int64, error)
	ListClusters(ctx context.Context) ([]domain.Cluster, error)
	ReserveCapacity(ctx context.Context, hypervisorID string, cpu int, memoryBytes, storageBytes int64) error
	ReleaseCapacity(ctx context.Context, hypervisorID string, cpu int, memoryBytes, storageBytes int64) error
	MarkSiblingHypervisorsOffline(ctx context.Context, nodeName, keepHypervisorID string) (int64, error)
}
