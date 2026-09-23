package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bosscloud/bosscloud/agents/hypervisor-agent/internal/agent"
	libconfig "github.com/bosscloud/bosscloud/libs/go/config"
)

func main() {
	for _, file := range []string{"../../.env", ".env"} {
		_ = libconfig.NewLoader(file).Load()
	}

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	controlURL := libconfig.String("AGENT_CONTROL_SERVICE_URL", "http://127.0.0.1:8086")
	mode := libconfig.String("AGENT_EXECUTOR_MODE", "simulate")
	storageDir := libconfig.String("AGENT_SIMULATE_STORAGE_DIR", "")
	version := libconfig.String("AGENT_VERSION", "0.1.0")
	nodeName := libconfig.String("AGENT_NODE_NAME", "")
	hypervisorID := libconfig.String("AGENT_HYPERVISOR_ID", "")
	stateFile := libconfig.String("AGENT_STATE_FILE", "agent-state.json")

	runtime, err := agent.NewRuntime(controlURL, mode, storageDir, version, nodeName, hypervisorID, stateFile, log)
	if err != nil {
		log.Error("failed to initialize agent", slog.String("error", err.Error()))
		os.Exit(1)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Info("starting hypervisor agent", slog.String("control_url", controlURL), slog.String("mode", mode))
	if err := runtime.Run(ctx); err != nil && err != context.Canceled {
		log.Error("agent stopped with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
