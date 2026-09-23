package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/vpsflow/vpsflow/libs/go/agentprotocol"
	"github.com/vpsflow/vpsflow/libs/go/errors"
	"github.com/vpsflow/vpsflow/libs/go/ids"
	"github.com/vpsflow/vpsflow/services/vm/internal/domain"
	"github.com/vpsflow/vpsflow/services/vm/internal/port"
)

type Service struct {
	repo         port.Repository
	cluster      port.ClusterClient
	agentControl port.AgentControlClient
}

func NewService(repo port.Repository, cluster port.ClusterClient, agentControl port.AgentControlClient) *Service {
	return &Service{repo: repo, cluster: cluster, agentControl: agentControl}
}

type CreateVMRequest struct {
	TenantID       string
	ProjectID      string
	Name           string
	VCPUs          int
	MemoryMB       int
	DiskGB         int
	ImageRef       string
	IdempotencyKey string
}

func (s *Service) CreateVM(ctx context.Context, req CreateVMRequest) (*domain.VirtualMachine, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.TenantID == "" || req.Name == "" {
		return nil, errors.New(errors.CodeValidation, "tenant_id and name are required")
	}
	if req.VCPUs <= 0 || req.MemoryMB <= 0 || req.DiskGB <= 0 {
		return nil, errors.New(errors.CodeValidation, "vcpus, memory_mb and disk_gb must be positive")
	}
	if req.IdempotencyKey != "" {
		if existing, err := s.repo.GetOperationByIdempotency(ctx, req.TenantID, req.IdempotencyKey); err != nil {
			return nil, err
		} else if existing != nil {
			vm, err := s.repo.GetVM(ctx, req.TenantID, existing.VMID)
			if err != nil {
				return nil, err
			}
			if vm.Status == domain.StatusPending {
				go s.provisionVM(context.WithoutCancel(ctx), vm.ID, req.TenantID)
			}
			return vm, nil
		}
	}

	vmID, err := ids.New("vm")
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to generate vm id", err)
	}

	domainName := fmt.Sprintf("bc-%s", vmID)
	vm := domain.VirtualMachine{
		ID: vmID, TenantID: req.TenantID, ProjectID: req.ProjectID, Name: req.Name,
		Status: domain.StatusPending, DomainName: domainName,
		VCPUs: req.VCPUs, MemoryMB: req.MemoryMB, DiskGB: req.DiskGB, ImageRef: req.ImageRef,
	}
	if err := s.repo.CreateVM(ctx, vm); err != nil {
		return nil, err
	}

	if req.IdempotencyKey != "" {
		opID, err := ids.New("vop")
		if err != nil {
			return nil, errors.Wrap(errors.CodeInternal, "failed to generate operation id", err)
		}
		_ = s.repo.CreateOperation(ctx, domain.Operation{
			ID: opID, VMID: vmID, TenantID: req.TenantID, OperationType: "create",
			IdempotencyKey: req.IdempotencyKey, Status: "accepted",
		})
	}

	go s.provisionVM(context.WithoutCancel(ctx), vmID, req.TenantID)
	return s.repo.GetVM(ctx, req.TenantID, vmID)
}

// ReconcilePending relaunches provisioning for VMs left in pending (e.g. after a crash).
func (s *Service) ReconcilePending(ctx context.Context) {
	vms, err := s.repo.ListPendingVMs(ctx)
	if err != nil {
		slog.Error("failed to list pending vms for reconciliation", slog.String("error", err.Error()))
		return
	}
	for _, vm := range vms {
		slog.Info("reconciling pending vm", slog.String("vm_id", vm.ID), slog.String("tenant_id", vm.TenantID))
		go s.provisionVM(context.WithoutCancel(ctx), vm.ID, vm.TenantID)
	}
}

func (s *Service) provisionVM(ctx context.Context, vmID, tenantID string) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	slog.Info("provisioning vm started", slog.String("vm_id", vmID), slog.String("tenant_id", tenantID))

	vm, err := s.repo.GetVM(ctx, tenantID, vmID)
	if err != nil {
		slog.Error("provisioning vm: get vm failed", slog.String("vm_id", vmID), slog.String("error", err.Error()))
		return
	}
	if err := s.repo.UpdateVMStatus(ctx, vmID, domain.StatusProvisioning, "", vm.DomainName, ""); err != nil {
		slog.Error("provisioning vm: failed to set provisioning", slog.String("vm_id", vmID), slog.String("error", err.Error()))
		return
	}

	hvID, err := s.cluster.SelectHypervisor(ctx, vm.VCPUs, vm.MemoryMB, vm.DiskGB, "")
	if err != nil {
		slog.Error("provisioning vm: select hypervisor failed", slog.String("vm_id", vmID), slog.String("error", err.Error()))
		_ = s.repo.UpdateVMStatus(ctx, vmID, domain.StatusError, "", vm.DomainName, err.Error())
		return
	}
	_ = s.repo.UpdateVMStatus(ctx, vmID, domain.StatusProvisioning, hvID, vm.DomainName, "")

	payload := agentprotocol.CreateDomainPayload{
		VMID: vmID, Name: vm.DomainName, VCPUs: vm.VCPUs, MemoryMB: vm.MemoryMB, DiskGB: vm.DiskGB, ImagePath: vm.ImageRef,
	}
	cmdID, status, errMsg, err := s.agentControl.DispatchCreateDomain(ctx, hvID, tenantID, payload, true)
	if err != nil {
		slog.Error("provisioning vm: dispatch failed", slog.String("vm_id", vmID), slog.String("error", err.Error()))
		_ = s.cluster.ReleaseCapacity(ctx, hvID, vm.VCPUs, vm.MemoryMB, vm.DiskGB)
		_ = s.repo.UpdateVMStatus(ctx, vmID, domain.StatusError, hvID, vm.DomainName, err.Error())
		return
	}
	if status == "failed" || errMsg != "" {
		slog.Error("provisioning vm: agent command failed", slog.String("vm_id", vmID), slog.String("status", status), slog.String("error", errMsg))
		_ = s.cluster.ReleaseCapacity(ctx, hvID, vm.VCPUs, vm.MemoryMB, vm.DiskGB)
		_ = s.repo.UpdateVMStatus(ctx, vmID, domain.StatusError, hvID, vm.DomainName, errMsg)
		return
	}
	if status != "completed" {
		msg := "provisioning timed out or agent did not complete the command"
		if err != nil {
			msg = err.Error()
		} else if errMsg != "" {
			msg = errMsg
		}
		slog.Error("provisioning vm: incomplete", slog.String("vm_id", vmID), slog.String("status", status), slog.String("error", msg))
		_ = s.cluster.ReleaseCapacity(ctx, hvID, vm.VCPUs, vm.MemoryMB, vm.DiskGB)
		_ = s.repo.UpdateVMStatus(ctx, vmID, domain.StatusError, hvID, vm.DomainName, msg)
		return
	}
	slog.Info("provisioning vm completed", slog.String("vm_id", vmID), slog.String("hypervisor_id", hvID))
	_ = s.repo.UpdateVMStatus(ctx, vmID, domain.StatusRunning, hvID, vm.DomainName, "")
	_ = cmdID
}

func (s *Service) GetVM(ctx context.Context, tenantID, vmID string) (*domain.VirtualMachine, error) {
	return s.repo.GetVM(ctx, tenantID, vmID)
}

func (s *Service) ListVMs(ctx context.Context, tenantID string) ([]domain.VirtualMachine, error) {
	if tenantID == "" {
		return nil, errors.New(errors.CodeValidation, "tenant_id is required")
	}
	return s.repo.ListVMs(ctx, tenantID)
}

func (s *Service) StartVM(ctx context.Context, tenantID, vmID string) (*domain.VirtualMachine, error) {
	vm, err := s.repo.GetVM(ctx, tenantID, vmID)
	if err != nil {
		return nil, err
	}
	if vm.Status == domain.StatusRunning {
		return vm, nil
	}
	if vm.HypervisorID == "" || vm.DomainName == "" {
		return nil, errors.New(errors.CodeConflict, "vm is not provisioned")
	}
	_ = s.repo.UpdateVMStatus(ctx, vmID, domain.StatusStarting, vm.HypervisorID, vm.DomainName, "")
	_, status, errMsg, err := s.agentControl.DispatchStartDomain(ctx, vm.HypervisorID, tenantID, vm.DomainName, true)
	if err != nil || status == "failed" || errMsg != "" {
		msg := errMsg
		if err != nil {
			msg = err.Error()
		}
		_ = s.repo.UpdateVMStatus(ctx, vmID, domain.StatusError, vm.HypervisorID, vm.DomainName, msg)
		return s.repo.GetVM(ctx, tenantID, vmID)
	}
	_ = s.repo.UpdateVMStatus(ctx, vmID, domain.StatusRunning, vm.HypervisorID, vm.DomainName, "")
	return s.repo.GetVM(ctx, tenantID, vmID)
}

func (s *Service) StopVM(ctx context.Context, tenantID, vmID string) (*domain.VirtualMachine, error) {
	vm, err := s.repo.GetVM(ctx, tenantID, vmID)
	if err != nil {
		return nil, err
	}
	if vm.Status == domain.StatusStopped {
		return vm, nil
	}
	if vm.HypervisorID == "" || vm.DomainName == "" {
		return nil, errors.New(errors.CodeConflict, "vm is not provisioned")
	}
	_ = s.repo.UpdateVMStatus(ctx, vmID, domain.StatusStopping, vm.HypervisorID, vm.DomainName, "")
	_, status, errMsg, err := s.agentControl.DispatchStopDomain(ctx, vm.HypervisorID, tenantID, vm.DomainName, true)
	if err != nil || status == "failed" || errMsg != "" {
		msg := errMsg
		if err != nil {
			msg = err.Error()
		}
		_ = s.repo.UpdateVMStatus(ctx, vmID, domain.StatusError, vm.HypervisorID, vm.DomainName, msg)
		return s.repo.GetVM(ctx, tenantID, vmID)
	}
	_ = s.repo.UpdateVMStatus(ctx, vmID, domain.StatusStopped, vm.HypervisorID, vm.DomainName, "")
	return s.repo.GetVM(ctx, tenantID, vmID)
}

func (s *Service) DeleteVM(ctx context.Context, tenantID, vmID string) error {
	vm, err := s.repo.GetVM(ctx, tenantID, vmID)
	if err != nil {
		return err
	}

	switch vm.Status {
	case domain.StatusPending, domain.StatusProvisioning, domain.StatusError:
		if vm.HypervisorID != "" {
			_ = s.cluster.ReleaseCapacity(ctx, vm.HypervisorID, vm.VCPUs, vm.MemoryMB, vm.DiskGB)
		}
		return s.repo.SoftDeleteVM(ctx, vmID)
	}

	_ = s.repo.UpdateVMStatus(ctx, vmID, domain.StatusDeleting, vm.HypervisorID, vm.DomainName, "")
	if vm.HypervisorID != "" && vm.DomainName != "" {
		_, _, errMsg, err := s.agentControl.DispatchDestroyDomain(ctx, vm.HypervisorID, tenantID, vm.DomainName, true)
		if err != nil {
			_ = s.repo.UpdateVMStatus(ctx, vmID, domain.StatusError, vm.HypervisorID, vm.DomainName, err.Error())
			return err
		}
		if errMsg != "" {
			_ = s.repo.UpdateVMStatus(ctx, vmID, domain.StatusError, vm.HypervisorID, vm.DomainName, errMsg)
			return errors.New(errors.CodeInternal, errMsg)
		}
		_ = s.cluster.ReleaseCapacity(ctx, vm.HypervisorID, vm.VCPUs, vm.MemoryMB, vm.DiskGB)
	}
	return s.repo.SoftDeleteVM(ctx, vmID)
}
