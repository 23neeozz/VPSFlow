package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/bosscloud/bosscloud/libs/go/errors"
	"github.com/bosscloud/bosscloud/libs/go/ids"
	"github.com/bosscloud/bosscloud/services/vps/internal/domain"
	"github.com/bosscloud/bosscloud/services/vps/internal/port"
)

var adminRoles = map[string]bool{
	"owner": true,
	"admin": true,
}

type Service struct {
	repo    port.Repository
	vm      port.VMClient
	console port.ConsoleClient
	tenant  port.TenantClient
}

func NewService(repo port.Repository, vm port.VMClient, console port.ConsoleClient, tenant port.TenantClient) *Service {
	return &Service{repo: repo, vm: vm, console: console, tenant: tenant}
}

type CreateVPSRequest struct {
	TenantID       string
	ActorUserID    string
	Authorization  string
	Name           string
	OwnerEmail     string
	OwnerUserID    string
	VCPUs          int
	MemoryMB       int
	DiskGB         int
	ImageRef       string
	IdempotencyKey string
}

type AssignVPSRequest struct {
	TenantID      string
	ActorUserID   string
	Authorization string
	VPSID         string
	OwnerEmail    string
	OwnerUserID   string
}

func (s *Service) IsAdmin(ctx context.Context, tenantID, userID, authorization string) (bool, error) {
	role, err := s.tenant.GetMemberRole(ctx, tenantID, userID, authorization)
	if err != nil {
		return false, err
	}
	return adminRoles[role], nil
}

func (s *Service) CreateVPS(ctx context.Context, req CreateVPSRequest) (*domain.Instance, error) {
	ok, err := s.IsAdmin(ctx, req.TenantID, req.ActorUserID, req.Authorization)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New(errors.CodeForbidden, "only organization admins can create vps")
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return nil, errors.New(errors.CodeValidation, "name is required")
	}
	if req.VCPUs <= 0 || req.MemoryMB <= 0 || req.DiskGB <= 0 {
		return nil, errors.New(errors.CodeValidation, "vcpus, memory_mb and disk_gb must be positive")
	}

	ownerUserID := strings.TrimSpace(req.OwnerUserID)
	ownerEmail := strings.TrimSpace(req.OwnerEmail)
	if ownerUserID == "" && ownerEmail == "" {
		return nil, errors.New(errors.CodeValidation, "owner_email or owner_user_id is required")
	}
	if ownerUserID == "" {
		ownerUserID, err = s.tenant.ResolveMemberByEmail(ctx, req.TenantID, ownerEmail, req.Authorization)
		if err != nil {
			return nil, err
		}
	}
	if ownerEmail == "" {
		ownerEmail = ownerUserID
	}

	vm, err := s.vm.CreateVM(ctx, req.TenantID, req.Authorization, req.IdempotencyKey, port.CreateVMRequest{
		Name: req.Name, VCPUs: req.VCPUs, MemoryMB: req.MemoryMB, DiskGB: req.DiskGB, ImageRef: req.ImageRef,
	})
	if err != nil {
		return nil, err
	}

	vpsID, err := ids.New("vps")
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to generate vps id", err)
	}
	now := time.Now().UTC()
	vps := domain.Instance{
		ID: vpsID, TenantID: req.TenantID, VMID: vm.ID, Name: req.Name,
		OwnerUserID: ownerUserID, OwnerEmail: ownerEmail, AssignedByUserID: req.ActorUserID,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, vps); err != nil {
		return nil, err
	}
	return s.enrich(ctx, &vps, req.Authorization)
}

func (s *Service) AssignVPS(ctx context.Context, req AssignVPSRequest) (*domain.Instance, error) {
	ok, err := s.IsAdmin(ctx, req.TenantID, req.ActorUserID, req.Authorization)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New(errors.CodeForbidden, "only organization admins can assign vps")
	}

	ownerUserID := strings.TrimSpace(req.OwnerUserID)
	ownerEmail := strings.TrimSpace(req.OwnerEmail)
	if ownerUserID == "" && ownerEmail == "" {
		return nil, errors.New(errors.CodeValidation, "owner_email or owner_user_id is required")
	}
	if ownerUserID == "" {
		ownerUserID, err = s.tenant.ResolveMemberByEmail(ctx, req.TenantID, ownerEmail, req.Authorization)
		if err != nil {
			return nil, err
		}
	}
	if ownerEmail == "" {
		ownerEmail = ownerUserID
	}

	if err := s.repo.UpdateOwner(ctx, req.TenantID, req.VPSID, ownerUserID, ownerEmail); err != nil {
		return nil, err
	}
	vps, err := s.repo.Get(ctx, req.TenantID, req.VPSID)
	if err != nil {
		return nil, err
	}
	return s.enrich(ctx, vps, req.Authorization)
}

func (s *Service) ListAdmin(ctx context.Context, tenantID, actorUserID, authorization string) ([]domain.Instance, error) {
	ok, err := s.IsAdmin(ctx, tenantID, actorUserID, authorization)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New(errors.CodeForbidden, "only organization admins can list all vps")
	}
	items, err := s.repo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return s.enrichAll(ctx, items, authorization)
}

func (s *Service) ListMine(ctx context.Context, tenantID, userID, authorization string) ([]domain.Instance, error) {
	items, err := s.repo.ListByOwner(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	return s.enrichAll(ctx, items, authorization)
}

func (s *Service) GetMine(ctx context.Context, tenantID, userID, vpsID, authorization string) (*domain.Instance, error) {
	vps, err := s.repo.Get(ctx, tenantID, vpsID)
	if err != nil {
		return nil, err
	}
	if vps.OwnerUserID != userID {
		return nil, errors.New(errors.CodeForbidden, "vps not assigned to you")
	}
	return s.enrich(ctx, vps, authorization)
}

func (s *Service) GetAdmin(ctx context.Context, tenantID, actorUserID, vpsID, authorization string) (*domain.Instance, error) {
	ok, err := s.IsAdmin(ctx, tenantID, actorUserID, authorization)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New(errors.CodeForbidden, "only organization admins can view vps details")
	}
	vps, err := s.repo.Get(ctx, tenantID, vpsID)
	if err != nil {
		return nil, err
	}
	return s.enrich(ctx, vps, authorization)
}

func (s *Service) StartVPS(ctx context.Context, tenantID, actorUserID, vpsID, authorization string) (*domain.Instance, error) {
	vps, err := s.assertAccess(ctx, tenantID, actorUserID, vpsID, authorization)
	if err != nil {
		return nil, err
	}
	if _, err := s.vm.StartVM(ctx, tenantID, vps.VMID, authorization); err != nil {
		return nil, err
	}
	return s.enrich(ctx, vps, authorization)
}

func (s *Service) StopVPS(ctx context.Context, tenantID, actorUserID, vpsID, authorization string) (*domain.Instance, error) {
	vps, err := s.assertAccess(ctx, tenantID, actorUserID, vpsID, authorization)
	if err != nil {
		return nil, err
	}
	if _, err := s.vm.StopVM(ctx, tenantID, vps.VMID, authorization); err != nil {
		return nil, err
	}
	return s.enrich(ctx, vps, authorization)
}

func (s *Service) DeleteVPS(ctx context.Context, tenantID, actorUserID, vpsID, authorization string) error {
	ok, err := s.IsAdmin(ctx, tenantID, actorUserID, authorization)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New(errors.CodeForbidden, "only organization admins can delete vps")
	}
	vps, err := s.repo.Get(ctx, tenantID, vpsID)
	if err != nil {
		return err
	}
	if err := s.vm.DeleteVM(ctx, tenantID, vps.VMID, authorization); err != nil {
		return err
	}
	return s.repo.SoftDelete(ctx, tenantID, vpsID)
}

func (s *Service) OpenConsole(ctx context.Context, tenantID, actorUserID, vpsID, authorization string) (*domain.ConsoleSession, error) {
	vps, err := s.assertAccess(ctx, tenantID, actorUserID, vpsID, authorization)
	if err != nil {
		return nil, err
	}
	return s.console.CreateSession(ctx, tenantID, vps.VMID, authorization)
}

func (s *Service) assertAccess(ctx context.Context, tenantID, actorUserID, vpsID, authorization string) (*domain.Instance, error) {
	vps, err := s.repo.Get(ctx, tenantID, vpsID)
	if err != nil {
		return nil, err
	}
	if vps.OwnerUserID == actorUserID {
		return vps, nil
	}
	ok, err := s.IsAdmin(ctx, tenantID, actorUserID, authorization)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New(errors.CodeForbidden, "vps not assigned to you")
	}
	return vps, nil
}

func (s *Service) enrichAll(ctx context.Context, items []domain.Instance, authorization string) ([]domain.Instance, error) {
	out := make([]domain.Instance, 0, len(items))
	for i := range items {
		enriched, err := s.enrich(ctx, &items[i], authorization)
		if err != nil {
			return nil, err
		}
		out = append(out, *enriched)
	}
	return out, nil
}

func (s *Service) enrich(ctx context.Context, vps *domain.Instance, authorization string) (*domain.Instance, error) {
	vm, err := s.vm.GetVM(ctx, vps.TenantID, vps.VMID, authorization)
	if err != nil {
		return vps, nil
	}
	vps.Status = vm.Status
	vps.VCPUs = vm.VCPUs
	vps.MemoryMB = vm.MemoryMB
	vps.DiskGB = vm.DiskGB
	vps.ErrorMessage = vm.ErrorMessage
	return vps, nil
}
