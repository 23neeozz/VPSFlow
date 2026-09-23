package agentprotocol

// Command types executed by hypervisor agents.
const (
	CommandCreateDomain  = "create_domain"
	CommandStartDomain   = "start_domain"
	CommandStopDomain    = "stop_domain"
	CommandDestroyDomain = "destroy_domain"
	CommandDefineDomain  = "define_domain"
)

// CommandStatus values for agent command lifecycle.
const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

// NodeCapacity describes hypervisor resources.
type NodeCapacity struct {
	CPUCores     int   `json:"cpu_cores"`
	MemoryBytes  int64 `json:"memory_bytes"`
	StorageBytes int64 `json:"storage_bytes"`
	RunningVMs   int   `json:"running_vms"`
}

// RegisterAgentRequest enrolls an agent.
type RegisterAgentRequest struct {
	HypervisorID  string       `json:"hypervisor_id"`
	AgentVersion  string       `json:"agent_version"`
	NodeName      string       `json:"node_name"`
	Capacity      NodeCapacity `json:"capacity"`
}

// RegisterAgentResponse confirms registration.
type RegisterAgentResponse struct {
	SessionID              string `json:"session_id"`
	HeartbeatIntervalSec   int64  `json:"heartbeat_interval_seconds"`
	CommandPollIntervalSec int64  `json:"command_poll_interval_seconds"`
}

// HeartbeatRequest updates agent liveness and capacity.
type HeartbeatRequest struct {
	HypervisorID string       `json:"hypervisor_id"`
	SessionID    string       `json:"session_id"`
	Capacity     NodeCapacity `json:"capacity"`
}

// Command describes work for an agent.
type Command struct {
	CommandID    string `json:"command_id"`
	HypervisorID string `json:"hypervisor_id"`
	Type         string `json:"type"`
	Payload      []byte `json:"payload"`
	TenantID     string `json:"tenant_id"`
}

// CommandResult reports command completion.
type CommandResult struct {
	CommandID    string `json:"command_id"`
	Status       string `json:"status"`
	Result       []byte `json:"result,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// CreateDomainPayload defines a VM domain to create.
type CreateDomainPayload struct {
	VMID       string `json:"vm_id"`
	Name       string `json:"name"`
	VCPUs      int    `json:"vcpus"`
	MemoryMB   int    `json:"memory_mb"`
	DiskGB     int    `json:"disk_gb"`
	ImagePath  string `json:"image_path,omitempty"`
	CloudInit  string `json:"cloud_init,omitempty"`
}

// DomainActionPayload targets an existing domain by name.
type DomainActionPayload struct {
	DomainName string `json:"domain_name"`
}
