package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type persistedState struct {
	HypervisorID string `json:"hypervisor_id"`
}

func resolveHypervisorID(envID, stateFile string) (string, error) {
	if envID != "" {
		return envID, nil
	}
	if stateFile == "" {
		stateFile = "agent-state.json"
	}
	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	var state persistedState
	if err := json.Unmarshal(data, &state); err != nil {
		return "", err
	}
	return state.HypervisorID, nil
}

func saveHypervisorID(stateFile, hypervisorID string) error {
	if stateFile == "" {
		stateFile = "agent-state.json"
	}
	if err := os.MkdirAll(filepath.Dir(absOrDot(stateFile)), 0o755); err != nil && filepath.Dir(stateFile) != "." {
		return err
	}
	payload, err := json.MarshalIndent(persistedState{HypervisorID: hypervisorID}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(stateFile, payload, 0o644)
}

func absOrDot(path string) string {
	if filepath.Dir(path) == "" {
		return "."
	}
	return filepath.Dir(path)
}
