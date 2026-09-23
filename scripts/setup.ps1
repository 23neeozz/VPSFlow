# VPSFlow Development Setup Script (Windows)

param(
    [switch]$SkipInfra
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot

Write-Host "VPSFlow Development Setup" -ForegroundColor Cyan
Write-Host "===========================" -ForegroundColor Cyan

# Check Go
$goCmd = Get-Command go -ErrorAction SilentlyContinue
if (-not $goCmd) {
    $goPath = "C:\Program Files\Go\bin\go.exe"
    if (Test-Path $goPath) {
        $env:PATH = "C:\Program Files\Go\bin;$env:PATH"
    } else {
        Write-Host "ERROR: Go 1.23+ is required. Install from https://go.dev/dl/" -ForegroundColor Red
        exit 1
    }
}

Write-Host "Go version: $(go version)" -ForegroundColor Green

# Copy env file
if (-not (Test-Path "$Root\.env")) {
    Copy-Item "$Root\.env.example" "$Root\.env"
    Write-Host "Created .env from .env.example" -ForegroundColor Green
}

# Tidy all modules
$modules = @(
    "libs\go\agentprotocol",
    "libs\go\config",
    "libs\go\errors",
    "libs\go\ids",
    "libs\go\observability",
    "libs\go\httpx",
    "libs\go\security",
    "services\gateway",
    "services\auth",
    "services\iam",
    "services\tenant",
    "services\cluster",
    "services\agent-control",
    "services\vm",
    "services\console",
    "agents\hypervisor-agent"
)

foreach ($mod in $modules) {
    Write-Host "Tidying $mod..." -ForegroundColor Yellow
    Push-Location "$Root\$mod"
    go mod tidy
    Pop-Location
}

# Run tests
Write-Host "Running tests..." -ForegroundColor Yellow
foreach ($mod in $modules) {
    Push-Location "$Root\$mod"
    go test -count=1 ./...
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Tests failed in $mod" -ForegroundColor Red
        exit 1
    }
    Pop-Location
}

Write-Host "All tests passed!" -ForegroundColor Green

# Start infrastructure
if (-not $SkipInfra) {
    $dockerCmd = Get-Command docker -ErrorAction SilentlyContinue
    if ($dockerCmd) {
        Write-Host "Starting infrastructure..." -ForegroundColor Yellow
        docker compose -f "$Root\platform\docker\docker-compose.yml" up -d
        Write-Host "Infrastructure started." -ForegroundColor Green
    } else {
        Write-Host "Docker not found. Skipping infrastructure startup." -ForegroundColor Yellow
        Write-Host "Create PostgreSQL databases: vpsflow_cluster, vpsflow_vm, vpsflow_agent_control, vpsflow_console" -ForegroundColor Yellow
    }
}

Write-Host "Setup complete! Run services:" -ForegroundColor Cyan
Write-Host "  Terminal 1:  cd services\auth; go run .\cmd\auth" -ForegroundColor White
Write-Host "  Terminal 2:  cd services\iam; go run .\cmd\iam" -ForegroundColor White
Write-Host "  Terminal 3:  cd services\tenant; go run .\cmd\tenant" -ForegroundColor White
Write-Host "  Terminal 4:  cd services\cluster; go run .\cmd\cluster" -ForegroundColor White
Write-Host "  Terminal 5:  cd services\agent-control; go run .\cmd\agent-control" -ForegroundColor White
Write-Host "  Terminal 6:  cd services\vm; go run .\cmd\vm" -ForegroundColor White
Write-Host "  Terminal 7:  cd services\console; go run .\cmd\console" -ForegroundColor White
Write-Host "  Terminal 8:  cd services\gateway; go run .\cmd\gateway" -ForegroundColor White
Write-Host "  Terminal 9:  cd agents\hypervisor-agent; go run .\cmd\hypervisor-agent" -ForegroundColor White
