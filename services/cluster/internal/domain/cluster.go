package domain

import "time"

const (
	HypervisorStatusOnline  = "online"
	HypervisorStatusOffline = "offline"
	HypervisorStatusDraining = "draining"
)

type Hypervisor struct {
	ID              string        `json:"id"`
	ClusterID       string        `json:"cluster_id,omitempty"`
	NodeName        string        `json:"node_name"`
	AgentVersion    string        `json:"agent_version"`
	Status          string        `json:"status"`
	MaintenanceMode bool          `json:"maintenance_mode"`
	LastHeartbeatAt *time.Time    `json:"last_heartbeat_at,omitempty"`
	Capacity        NodeCapacity  `json:"capacity"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

type NodeCapacity struct {
	CPUCores              int   `json:"cpu_cores"`
	MemoryBytes           int64 `json:"memory_bytes"`
	StorageBytes          int64 `json:"storage_bytes"`
	RunningVMs            int   `json:"running_vms"`
	AllocatedCPU          int   `json:"allocated_cpu"`
	AllocatedMemoryBytes  int64 `json:"allocated_memory_bytes"`
	AllocatedStorageBytes int64 `json:"allocated_storage_bytes"`
}

type Cluster struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
