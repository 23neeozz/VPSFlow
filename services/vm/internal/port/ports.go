package port

import (
	"context"

	"github.com/vpsflow/vpsflow/services/vm/internal/domain"
)

type Repository interface {
	CreateVM(ctx context.Context, vm domain.VirtualMachine) error
	GetVM(ctx context.Context, tenantID, vmID string) (*domain.VirtualMachine, error)
	ListVMs(ctx context.Context, tenantID string) ([]domain.VirtualMachine, error)
	UpdateVMStatus(ctx context.Context, vmID, status, hypervisorID, domainName, errorMessage string) error
	SoftDeleteVM(ctx context.Context, vmID string) error
	ListPendingVMs(ctx context.Context) ([]domain.VirtualMachine, error)
	GetOperationByIdempotency(ctx context.Context, tenantID, key string) (*domain.Operation, error)
	CreateOperation(ctx context.Context, op domain.Operation) error
	UpdateOperation(ctx context.Context, opID, status, commandID string) error
}

type ClusterClient interface {
	SelectHypervisor(ctx context.Context, vcpus, memoryMB, diskGB int, clusterID string) (hypervisorID string, err error)
	ReleaseCapacity(ctx context.Context, hypervisorID string, vcpus, memoryMB, diskGB int) error
}

type AgentControlClient interface {
	DispatchCreateDomain(ctx context.Context, hypervisorID, tenantID string, payload any, wait bool) (commandID, status, errMsg string, err error)
	DispatchStartDomain(ctx context.Context, hypervisorID, tenantID, domainName string, wait bool) (commandID, status, errMsg string, err error)
	DispatchStopDomain(ctx context.Context, hypervisorID, tenantID, domainName string, wait bool) (commandID, status, errMsg string, err error)
	DispatchDestroyDomain(ctx context.Context, hypervisorID, tenantID, domainName string, wait bool) (commandID, status, errMsg string, err error)
}
