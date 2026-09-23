package usecase

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/vpsflow/vpsflow/libs/go/errors"
	"github.com/vpsflow/vpsflow/libs/go/ids"
	"github.com/vpsflow/vpsflow/services/cluster/internal/domain"
	"github.com/vpsflow/vpsflow/services/cluster/internal/port"
)

type Service struct {
	repo              port.Repository
	heartbeatStaleDur time.Duration
}

func NewService(repo port.Repository, heartbeatStaleDur time.Duration) *Service {
	if heartbeatStaleDur <= 0 {
		heartbeatStaleDur = 90 * time.Second
	}
	return &Service{repo: repo, heartbeatStaleDur: heartbeatStaleDur}
}

func (s *Service) RegisterHypervisor(ctx context.Context, nodeName, agentVersion string, capacity domain.NodeCapacity) (*domain.Hypervisor, error) {
	nodeName = strings.TrimSpace(nodeName)
	if nodeName == "" {
		return nil, errors.New(errors.CodeValidation, "node_name is required")
	}

	hvID, err := ids.New("hyp")
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to generate hypervisor id", err)
	}

	now := time.Now().UTC()
	hv := domain.Hypervisor{
		ID:              hvID,
		ClusterID:       "clu_default",
		NodeName:        nodeName,
		AgentVersion:    agentVersion,
		Status:          domain.HypervisorStatusOnline,
		MaintenanceMode: false,
		LastHeartbeatAt: &now,
		Capacity:        capacity,
	}
	if err := s.repo.UpsertHypervisor(ctx, hv); err != nil {
		return nil, err
	}
	return s.repo.GetHypervisor(ctx, hvID)
}

func (s *Service) RegisterHypervisorWithID(ctx context.Context, hypervisorID, nodeName, agentVersion string, capacity domain.NodeCapacity) (*domain.Hypervisor, error) {
	hypervisorID = strings.TrimSpace(hypervisorID)
	nodeName = strings.TrimSpace(nodeName)
	if hypervisorID == "" || nodeName == "" {
		return nil, errors.New(errors.CodeValidation, "hypervisor_id and node_name are required")
	}

	now := time.Now().UTC()
	hv := domain.Hypervisor{
		ID:              hypervisorID,
		ClusterID:       "clu_default",
		NodeName:        nodeName,
		AgentVersion:    agentVersion,
		Status:          domain.HypervisorStatusOnline,
		MaintenanceMode: false,
		LastHeartbeatAt: &now,
		Capacity:        capacity,
	}
	if err := s.repo.UpsertHypervisor(ctx, hv); err != nil {
		return nil, err
	}
	return s.repo.GetHypervisor(ctx, hypervisorID)
}

func (s *Service) Heartbeat(ctx context.Context, hypervisorID string, capacity domain.NodeCapacity) error {
	if hypervisorID == "" {
		return errors.New(errors.CodeValidation, "hypervisor_id is required")
	}
	return s.repo.UpdateHypervisorCapacity(ctx, hypervisorID, capacity, domain.HypervisorStatusOnline, true)
}

func (s *Service) GetHypervisor(ctx context.Context, hypervisorID string) (*domain.Hypervisor, error) {
	if hypervisorID == "" {
		return nil, errors.New(errors.CodeValidation, "hypervisor_id is required")
	}
	return s.repo.GetHypervisor(ctx, hypervisorID)
}

func (s *Service) ListHypervisors(ctx context.Context, clusterID string, onlineOnly bool) ([]domain.Hypervisor, error) {
	_, _ = s.repo.MarkStaleHypervisorsOffline(ctx, time.Now().UTC().Add(-s.heartbeatStaleDur))
	return s.repo.ListHypervisors(ctx, clusterID, onlineOnly)
}

func (s *Service) ListClusters(ctx context.Context) ([]domain.Cluster, error) {
	return s.repo.ListClusters(ctx)
}

type PlacementRequest struct {
	VCPUs      int
	MemoryMB   int
	DiskGB     int
	ClusterID  string
}

func (s *Service) SelectHypervisor(ctx context.Context, req PlacementRequest) (*domain.Hypervisor, error) {
	if req.VCPUs <= 0 || req.MemoryMB <= 0 || req.DiskGB <= 0 {
		return nil, errors.New(errors.CodeValidation, "vcpu, memory_mb and disk_gb must be positive")
	}

	_, _ = s.repo.MarkStaleHypervisorsOffline(ctx, time.Now().UTC().Add(-s.heartbeatStaleDur))
	hvs, err := s.repo.ListHypervisors(ctx, req.ClusterID, true)
	if err != nil {
		return nil, err
	}
	if len(hvs) == 0 {
		return nil, errors.New(errors.CodeNotFound, "no online hypervisors available")
	}

	memoryBytes := int64(req.MemoryMB) * 1024 * 1024
	storageBytes := int64(req.DiskGB) * 1024 * 1024 * 1024

	type candidate struct {
		hv    domain.Hypervisor
		score int64
	}
	candidates := make([]candidate, 0, len(hvs))
	for _, hv := range hvs {
		freeCPU := hv.Capacity.CPUCores - hv.Capacity.AllocatedCPU
		freeMem := hv.Capacity.MemoryBytes - hv.Capacity.AllocatedMemoryBytes
		freeDisk := hv.Capacity.StorageBytes - hv.Capacity.AllocatedStorageBytes
		if freeCPU < req.VCPUs || freeMem < memoryBytes || freeDisk < storageBytes {
			continue
		}
		candidates = append(candidates, candidate{hv: hv, score: freeMem + freeDisk})
	}
	if len(candidates) == 0 {
		return nil, errors.New(errors.CodeConflict, "no hypervisor has sufficient capacity")
	}

	sort.Slice(candidates, func(i, j int) bool {
		hi := heartbeatTime(candidates[i].hv)
		hj := heartbeatTime(candidates[j].hv)
		if !hi.Equal(hj) {
			return hi.After(hj)
		}
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		return candidates[i].hv.Capacity.RunningVMs < candidates[j].hv.Capacity.RunningVMs
	})

	selected := candidates[0].hv
	if err := s.repo.ReserveCapacity(ctx, selected.ID, req.VCPUs, memoryBytes, storageBytes); err != nil {
		return nil, err
	}
	return &selected, nil
}

func heartbeatTime(hv domain.Hypervisor) time.Time {
	if hv.LastHeartbeatAt != nil {
		return hv.LastHeartbeatAt.UTC()
	}
	return time.Time{}
}

func (s *Service) ReleaseCapacity(ctx context.Context, hypervisorID string, vcpus, memoryMB, diskGB int) error {
	memoryBytes := int64(memoryMB) * 1024 * 1024
	storageBytes := int64(diskGB) * 1024 * 1024 * 1024
	return s.repo.ReleaseCapacity(ctx, hypervisorID, vcpus, memoryBytes, storageBytes)
}

func (s *Service) DrainSiblingHypervisors(ctx context.Context, nodeName, keepHypervisorID string) error {
	nodeName = strings.TrimSpace(nodeName)
	if nodeName == "" || keepHypervisorID == "" {
		return errors.New(errors.CodeValidation, "node_name and hypervisor_id are required")
	}
	_, err := s.repo.MarkSiblingHypervisorsOffline(ctx, nodeName, keepHypervisorID)
	return err
}
