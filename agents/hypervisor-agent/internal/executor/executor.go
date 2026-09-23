package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/vpsflow/vpsflow/libs/go/agentprotocol"
)

// DomainExecutor runs libvirt domain operations.
type DomainExecutor interface {
	CreateDomain(ctx context.Context, payload agentprotocol.CreateDomainPayload) error
	StartDomain(ctx context.Context, domainName string) error
	StopDomain(ctx context.Context, domainName string) error
	DestroyDomain(ctx context.Context, domainName string) error
	Capacity(ctx context.Context) (agentprotocol.NodeCapacity, error)
}

func New(mode, storageDir string) DomainExecutor {
	if strings.EqualFold(mode, "simulate") {
		return NewSimulateExecutor(storageDir)
	}
	return NewVirshExecutor()
}

// SimulateExecutor keeps domain state in memory for development without libvirt.
type SimulateExecutor struct {
	mu      sync.Mutex
	domains map[string]string
	dir     string
}

func NewSimulateExecutor(dir string) *SimulateExecutor {
	if dir == "" {
		dir = os.TempDir()
	}
	return &SimulateExecutor{domains: make(map[string]string), dir: dir}
}

func (e *SimulateExecutor) CreateDomain(ctx context.Context, payload agentprotocol.CreateDomainPayload) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if payload.Name == "" {
		return fmt.Errorf("domain name is required")
	}
	e.domains[payload.Name] = "running"
	disk := filepath.Join(e.dir, payload.Name+".qcow2")
	return os.WriteFile(disk, []byte("vpsflow-simulated-disk"), 0o644)
}

func (e *SimulateExecutor) StartDomain(ctx context.Context, domainName string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if _, ok := e.domains[domainName]; !ok {
		return fmt.Errorf("domain %s not found", domainName)
	}
	e.domains[domainName] = "running"
	return nil
}

func (e *SimulateExecutor) StopDomain(ctx context.Context, domainName string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if _, ok := e.domains[domainName]; !ok {
		return fmt.Errorf("domain %s not found", domainName)
	}
	e.domains[domainName] = "stopped"
	return nil
}

func (e *SimulateExecutor) DestroyDomain(ctx context.Context, domainName string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.domains, domainName)
	_ = os.Remove(filepath.Join(e.dir, domainName+".qcow2"))
	return nil
}

func (e *SimulateExecutor) Capacity(ctx context.Context) (agentprotocol.NodeCapacity, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	running := 0
	for _, state := range e.domains {
		if state == "running" {
			running++
		}
	}
	return agentprotocol.NodeCapacity{
		CPUCores: 8, MemoryBytes: 32 * 1024 * 1024 * 1024, StorageBytes: 500 * 1024 * 1024 * 1024, RunningVMs: running,
	}, nil
}

// VirshExecutor shells out to virsh for production KVM nodes.
type VirshExecutor struct{}

func NewVirshExecutor() *VirshExecutor {
	return &VirshExecutor{}
}

func (e *VirshExecutor) CreateDomain(ctx context.Context, payload agentprotocol.CreateDomainPayload) error {
	xml := buildDomainXML(payload)
	tmp, err := os.CreateTemp("", "vpsflow-domain-*.xml")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(xml); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return runVirsh(ctx, "define", tmp.Name())
}

func (e *VirshExecutor) StartDomain(ctx context.Context, domainName string) error {
	return runVirsh(ctx, "start", domainName)
}

func (e *VirshExecutor) StopDomain(ctx context.Context, domainName string) error {
	return runVirsh(ctx, "shutdown", domainName)
}

func (e *VirshExecutor) DestroyDomain(ctx context.Context, domainName string) error {
	_ = runVirsh(ctx, "destroy", domainName)
	return runVirsh(ctx, "undefine", domainName)
}

func (e *VirshExecutor) Capacity(ctx context.Context) (agentprotocol.NodeCapacity, error) {
	out, err := exec.CommandContext(ctx, "virsh", "nodeinfo").CombinedOutput()
	if err != nil {
		return agentprotocol.NodeCapacity{}, fmt.Errorf("virsh nodeinfo failed: %s", strings.TrimSpace(string(out)))
	}
	cores := 4
	memGB := int64(16)
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "CPU(s):") {
			fmt.Sscanf(line, "CPU(s):%d", &cores)
		}
		if strings.Contains(line, "Memory size:") {
			var kb int64
			fmt.Sscanf(line, "Memory size:%d KiB", &kb)
			memGB = kb * 1024
		}
	}
	domains, _ := exec.CommandContext(ctx, "virsh", "list", "--name").CombinedOutput()
	running := 0
	for _, name := range strings.Split(string(domains), "\n") {
		if strings.TrimSpace(name) != "" {
			running++
		}
	}
	return agentprotocol.NodeCapacity{
		CPUCores: cores, MemoryBytes: memGB, StorageBytes: 1024 * 1024 * 1024 * 1024, RunningVMs: running,
	}, nil
}

func runVirsh(ctx context.Context, args ...string) error {
	out, err := exec.CommandContext(ctx, "virsh", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("virsh %s failed: %s", strings.Join(args, " "), strings.TrimSpace(string(out)))
	}
	return nil
}

func buildDomainXML(payload agentprotocol.CreateDomainPayload) string {
	memKiB := payload.MemoryMB * 1024
	diskPath := payload.ImagePath
	if diskPath == "" {
		diskPath = fmt.Sprintf("/var/lib/vpsflow/images/%s.qcow2", payload.Name)
	}
	return fmt.Sprintf(`<domain type='kvm'>
  <name>%s</name>
  <memory unit='KiB'>%d</memory>
  <vcpu>%d</vcpu>
  <os><type arch='x86_64' machine='pc'>hvm</type></os>
  <devices>
    <disk type='file' device='disk'>
      <source file='%s'/>
      <target dev='vda' bus='virtio'/>
    </disk>
    <interface type='network'>
      <source network='default'/>
      <model type='virtio'/>
    </interface>
    <console type='pty'/>
  </devices>
</domain>`, payload.Name, memKiB, payload.VCPUs, diskPath)
}

func DecodeCreatePayload(raw []byte) (agentprotocol.CreateDomainPayload, error) {
	var payload agentprotocol.CreateDomainPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return payload, err
	}
	return payload, nil
}

func DecodeDomainAction(raw []byte) (agentprotocol.DomainActionPayload, error) {
	var payload agentprotocol.DomainActionPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return payload, err
	}
	return payload, nil
}
