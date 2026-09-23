package usecase

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/vpsflow/vpsflow/libs/go/agentprotocol"
	"github.com/vpsflow/vpsflow/libs/go/errors"
	"github.com/vpsflow/vpsflow/libs/go/ids"
	"github.com/vpsflow/vpsflow/services/agent-control/internal/domain"
	"github.com/vpsflow/vpsflow/services/agent-control/internal/port"
)

type Service struct {
	repo          port.Repository
	clusterClient port.ClusterClient
	cmdTimeout    time.Duration
}

func NewService(repo port.Repository, clusterClient port.ClusterClient, cmdTimeout time.Duration) *Service {
	if cmdTimeout <= 0 {
		cmdTimeout = 2 * time.Minute
	}
	return &Service{repo: repo, clusterClient: clusterClient, cmdTimeout: cmdTimeout}
}

func (s *Service) RegisterAgent(ctx context.Context, req agentprotocol.RegisterAgentRequest) (*agentprotocol.RegisterAgentResponse, error) {
	if strings.TrimSpace(req.HypervisorID) == "" || strings.TrimSpace(req.NodeName) == "" {
		return nil, errors.New(errors.CodeValidation, "hypervisor_id and node_name are required")
	}

	sessionID, err := ids.New("ses")
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to generate session id", err)
	}

	session := domain.Session{
		ID:           sessionID,
		HypervisorID: req.HypervisorID,
		AgentVersion: req.AgentVersion,
		NodeName:     req.NodeName,
		Status:       domain.SessionActive,
	}
	if err := s.repo.UpsertSession(ctx, session); err != nil {
		return nil, err
	}

	if err := s.clusterClient.RegisterHypervisor(ctx, req.HypervisorID, req.NodeName, req.AgentVersion,
		req.Capacity.CPUCores, req.Capacity.MemoryBytes, req.Capacity.StorageBytes, req.Capacity.RunningVMs); err != nil {
		return nil, err
	}
	if err := s.clusterClient.DrainSiblingHypervisors(ctx, req.NodeName, req.HypervisorID); err != nil {
		return nil, err
	}

	return &agentprotocol.RegisterAgentResponse{
		SessionID:              sessionID,
		HeartbeatIntervalSec:   30,
		CommandPollIntervalSec: 2,
	}, nil
}

func (s *Service) Heartbeat(ctx context.Context, req agentprotocol.HeartbeatRequest) error {
	if req.HypervisorID == "" || req.SessionID == "" {
		return errors.New(errors.CodeValidation, "hypervisor_id and session_id are required")
	}
	session, err := s.repo.GetSessionByHypervisor(ctx, req.HypervisorID)
	if err != nil {
		return err
	}
	if session.ID != req.SessionID {
		return errors.New(errors.CodeUnauthorized, "invalid session")
	}
	if err := s.repo.TouchSession(ctx, req.SessionID); err != nil {
		return err
	}
	return s.clusterClient.Heartbeat(ctx, req.HypervisorID, req.Capacity.CPUCores, req.Capacity.MemoryBytes, req.Capacity.StorageBytes, req.Capacity.RunningVMs)
}

func (s *Service) PollCommand(ctx context.Context, hypervisorID, sessionID string) (*agentprotocol.Command, error) {
	if hypervisorID == "" || sessionID == "" {
		return nil, errors.New(errors.CodeValidation, "hypervisor_id and session_id are required")
	}
	session, err := s.repo.GetSessionByHypervisor(ctx, hypervisorID)
	if err != nil {
		return nil, err
	}
	if session.ID != sessionID {
		return nil, errors.New(errors.CodeUnauthorized, "invalid session")
	}

	cmd, err := s.repo.ClaimNextCommand(ctx, hypervisorID)
	if err != nil {
		return nil, err
	}
	if cmd == nil {
		return nil, nil
	}
	return &agentprotocol.Command{
		CommandID: cmd.ID, HypervisorID: cmd.HypervisorID, Type: cmd.Type, Payload: cmd.Payload, TenantID: cmd.TenantID,
	}, nil
}

func (s *Service) CompleteCommand(ctx context.Context, result agentprotocol.CommandResult) error {
	if result.CommandID == "" {
		return errors.New(errors.CodeValidation, "command_id is required")
	}
	status := domain.CommandCompleted
	if result.Status == agentprotocol.StatusFailed || result.ErrorMessage != "" {
		status = domain.CommandFailed
	}
	return s.repo.CompleteCommand(ctx, result.CommandID, status, result.Result, result.ErrorMessage)
}

type DispatchRequest struct {
	HypervisorID string
	TenantID     string
	Type         string
	Payload      any
	Wait         bool
}

type DispatchResponse struct {
	CommandID string
	Status    string
	Result    []byte
	Error     string
}

func (s *Service) DispatchCommand(ctx context.Context, req DispatchRequest) (*DispatchResponse, error) {
	if req.HypervisorID == "" || req.Type == "" {
		return nil, errors.New(errors.CodeValidation, "hypervisor_id and type are required")
	}

	commandID, err := ids.New("cmd")
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to generate command id", err)
	}

	var payload []byte
	switch v := req.Payload.(type) {
	case []byte:
		payload = v
	case nil:
		payload = []byte{}
	default:
		payload, err = json.Marshal(v)
		if err != nil {
			return nil, errors.Wrap(errors.CodeValidation, "invalid command payload", err)
		}
	}

	cmd := domain.Command{
		ID: commandID, HypervisorID: req.HypervisorID, TenantID: req.TenantID,
		Type: req.Type, Payload: payload, Status: domain.CommandPending,
	}
	if err := s.repo.CreateCommand(ctx, cmd); err != nil {
		return nil, err
	}

	resp := &DispatchResponse{CommandID: commandID, Status: domain.CommandPending}
	if !req.Wait {
		return resp, nil
	}

	finished, err := s.repo.WaitForCommandCompletion(ctx, commandID, s.cmdTimeout)
	if err != nil {
		return resp, err
	}
	resp.Status = finished.Status
	resp.Result = finished.Result
	resp.Error = finished.ErrorMessage
	return resp, nil
}

func (s *Service) GetCommand(ctx context.Context, commandID string) (*domain.Command, error) {
	return s.repo.GetCommand(ctx, commandID)
}
