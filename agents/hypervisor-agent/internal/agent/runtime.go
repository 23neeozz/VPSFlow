package agent

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/bosscloud/bosscloud/agents/hypervisor-agent/internal/controlclient"
	"github.com/bosscloud/bosscloud/agents/hypervisor-agent/internal/executor"
	"github.com/bosscloud/bosscloud/libs/go/agentprotocol"
	"github.com/google/uuid"
)

type Runtime struct {
	client       *controlclient.Client
	exec         executor.DomainExecutor
	hypervisorID string
	sessionID    string
	version      string
	nodeName     string
	stateFile    string
	log          *slog.Logger
}

func NewRuntime(controlURL, mode, storageDir, version, nodeName, hypervisorID, stateFile string, log *slog.Logger) (*Runtime, error) {
	if nodeName == "" {
		nodeName, _ = osHostname()
	}

	resolvedID, err := resolveHypervisorID(hypervisorID, stateFile)
	if err != nil {
		return nil, err
	}
	if resolvedID == "" {
		resolvedID = "hyp_" + uuid.NewString()
	}

	return &Runtime{
		client:       controlclient.New(controlURL),
		exec:         executor.New(mode, storageDir),
		hypervisorID: resolvedID,
		version:      version,
		nodeName:     nodeName,
		stateFile:    stateFile,
		log:          log,
	}, nil
}

func (r *Runtime) Run(ctx context.Context) error {
	reg, err := r.registerWithRetry(ctx)
	if err != nil {
		return err
	}
	r.sessionID = reg.SessionID
	r.log.Info("agent registered", slog.String("hypervisor_id", r.hypervisorID), slog.String("session_id", r.sessionID))

	heartbeatInterval := time.Duration(reg.HeartbeatIntervalSec) * time.Second
	if heartbeatInterval <= 0 {
		heartbeatInterval = 30 * time.Second
	}

	loopCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, 2)
	go func() { errCh <- r.heartbeatLoop(loopCtx, heartbeatInterval) }()
	go func() { errCh <- r.pollLoop(loopCtx) }()

	select {
	case <-ctx.Done():
		cancel()
		return ctx.Err()
	case err := <-errCh:
		cancel()
		return err
	}
}

func (r *Runtime) heartbeatLoop(ctx context.Context, interval time.Duration) error {
	r.sendHeartbeat(ctx)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			r.sendHeartbeat(ctx)
		}
	}
}

func (r *Runtime) sendHeartbeat(ctx context.Context) {
	capacity, err := r.exec.Capacity(ctx)
	if err != nil {
		r.log.Error("capacity read failed", slog.String("error", err.Error()))
		return
	}
	if err := r.client.Heartbeat(ctx, agentprotocol.HeartbeatRequest{
		HypervisorID: r.hypervisorID, SessionID: r.sessionID, Capacity: capacity,
	}); err != nil {
		r.log.Error("heartbeat failed", slog.String("error", err.Error()))
	}
}

func (r *Runtime) pollLoop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		cmd, err := r.client.PollCommand(ctx, r.hypervisorID, r.sessionID)
		if err != nil {
			r.log.Error("poll failed", slog.String("error", err.Error()))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(2 * time.Second):
			}
			continue
		}
		if cmd != nil {
			r.executeCommand(ctx, cmd)
		}
	}
}

func (r *Runtime) registerWithRetry(ctx context.Context) (*agentprotocol.RegisterAgentResponse, error) {
	const maxBackoff = 30 * time.Second
	backoff := 2 * time.Second
	for attempt := 1; ; attempt++ {
		capacity, err := r.exec.Capacity(ctx)
		if err != nil {
			return nil, err
		}
		reg, err := r.client.Register(ctx, agentprotocol.RegisterAgentRequest{
			HypervisorID: r.hypervisorID,
			AgentVersion: r.version,
			NodeName:     r.nodeName,
			Capacity:     capacity,
		})
		if err == nil {
			if saveErr := saveHypervisorID(r.stateFile, r.hypervisorID); saveErr != nil {
				r.log.Warn("failed to persist hypervisor id", slog.String("error", saveErr.Error()))
			}
			return reg, nil
		}

		r.log.Warn("agent registration failed, retrying",
			slog.Int("attempt", attempt),
			slog.Duration("retry_in", backoff),
			slog.String("error", err.Error()),
		)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
		if backoff < maxBackoff {
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

func (r *Runtime) executeCommand(ctx context.Context, cmd *agentprotocol.Command) {
	result := agentprotocol.CommandResult{CommandID: cmd.CommandID, Status: agentprotocol.StatusCompleted}
	var err error

	switch cmd.Type {
	case agentprotocol.CommandCreateDomain:
		var payload agentprotocol.CreateDomainPayload
		err = json.Unmarshal(cmd.Payload, &payload)
		if err == nil {
			err = r.exec.CreateDomain(ctx, payload)
		}
	case agentprotocol.CommandStartDomain:
		payload, decodeErr := executor.DecodeDomainAction(cmd.Payload)
		err = decodeErr
		if err == nil {
			err = r.exec.StartDomain(ctx, payload.DomainName)
		}
	case agentprotocol.CommandStopDomain:
		payload, decodeErr := executor.DecodeDomainAction(cmd.Payload)
		err = decodeErr
		if err == nil {
			err = r.exec.StopDomain(ctx, payload.DomainName)
		}
	case agentprotocol.CommandDestroyDomain:
		payload, decodeErr := executor.DecodeDomainAction(cmd.Payload)
		err = decodeErr
		if err == nil {
			err = r.exec.DestroyDomain(ctx, payload.DomainName)
		}
	default:
		result.Status = agentprotocol.StatusFailed
		result.ErrorMessage = "unsupported command type: " + cmd.Type
	}

	if err != nil {
		result.Status = agentprotocol.StatusFailed
		result.ErrorMessage = err.Error()
	}
	if completeErr := r.client.CompleteCommand(ctx, result); completeErr != nil {
		r.log.Error("failed to report command result", slog.String("error", completeErr.Error()))
	}
}

func osHostname() (string, error) {
	return hostname()
}
