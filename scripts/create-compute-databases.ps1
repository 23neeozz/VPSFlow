# Create BossCloud PostgreSQL databases (Windows)

$psql = "C:\Program Files\PostgreSQL\18\bin\psql.exe"
if (-not (Test-Path $psql)) {
    Write-Host "psql not found at $psql. Adjust path or add PostgreSQL bin to PATH." -ForegroundColor Red
    exit 1
}

$databases = @(
    "bosscloud_cluster",
    "bosscloud_vm",
    "bosscloud_agent_control",
    "bosscloud_console"
)

foreach ($db in $databases) {
    Write-Host "Creating database $db..." -ForegroundColor Yellow
    & $psql -U bosscloud -h localhost -c "CREATE DATABASE $db OWNER bosscloud;" 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  OK" -ForegroundColor Green
    } else {
        Write-Host "  Skipped (may already exist)" -ForegroundColor DarkYellow
    }
}

Write-Host "Done." -ForegroundColor Cyan
