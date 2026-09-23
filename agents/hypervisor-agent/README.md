# Hypervisor Agent

VPSFlow agent that runs on KVM hypervisor nodes and executes libvirt operations under command from agent-control.

## Modes

- `simulate` (default for local dev): in-memory domain simulation, no libvirt required
- `virsh`: production mode using virsh/libvirt on Linux

## Run (development)

```powershell
cd agents\hypervisor-agent
$env:AGENT_CONTROL_SERVICE_URL="http://127.0.0.1:8086"
$env:AGENT_EXECUTOR_MODE="simulate"
go run .\cmd\hypervisor-agent
```

## Environment

| Variable | Default | Description |
|----------|---------|-------------|
| `AGENT_CONTROL_SERVICE_URL` | `http://127.0.0.1:8086` | agent-control base URL |
| `AGENT_EXECUTOR_MODE` | `simulate` | `simulate` or `virsh` |
| `AGENT_NODE_NAME` | hostname | node display name |
| `AGENT_VERSION` | `0.1.0` | reported agent version |
