@echo off
setlocal EnableExtensions
title BossCloud - Launcher

set "ROOT=%~dp0.."
cd /d "%ROOT%"

echo ============================================
echo   BossCloud - Iniciando todos los servicios
echo ============================================
echo.
echo Raiz: %ROOT%
echo.

where go >nul 2>&1
if errorlevel 1 (
    if exist "C:\Program Files\Go\bin\go.exe" (
        set "PATH=C:\Program Files\Go\bin;%PATH%"
    ) else (
        echo ERROR: Go no encontrado.
        pause
        exit /b 1
    )
)

go version
echo.

if not exist "%ROOT%\.env" (
    echo AVISO: Falta .env - copia .env.example a .env
    echo.
)

echo [1/9] auth :8081
start "BossCloud-auth" cmd /k "cd /d ""%ROOT%\services\auth"" && title BossCloud - auth :8081 && go run .\cmd\auth"
timeout /t 2 /nobreak >nul

echo [2/9] iam :8082
start "BossCloud-iam" cmd /k "cd /d ""%ROOT%\services\iam"" && title BossCloud - iam :8082 && go run .\cmd\iam"
timeout /t 2 /nobreak >nul

echo [3/9] tenant :8083
start "BossCloud-tenant" cmd /k "cd /d ""%ROOT%\services\tenant"" && title BossCloud - tenant :8083 && go run .\cmd\tenant"
timeout /t 2 /nobreak >nul

echo [4/9] cluster :8084
start "BossCloud-cluster" cmd /k "cd /d ""%ROOT%\services\cluster"" && title BossCloud - cluster :8084 && go run .\cmd\cluster"
timeout /t 2 /nobreak >nul

echo [5/9] agent-control :8086
start "BossCloud-agent-control" cmd /k "cd /d ""%ROOT%\services\agent-control"" && title BossCloud - agent-control :8086 && go run .\cmd\agent-control"
timeout /t 2 /nobreak >nul

echo [6/9] vm :8085
start "BossCloud-vm" cmd /k "cd /d ""%ROOT%\services\vm"" && title BossCloud - vm :8085 && go run .\cmd\vm"
timeout /t 2 /nobreak >nul

echo [7/10] console :8087
start "BossCloud-console" cmd /k "cd /d ""%ROOT%\services\console"" && title BossCloud - console :8087 && go run .\cmd\console"
timeout /t 2 /nobreak >nul

echo [8/10] vps :8088
start "BossCloud-vps" cmd /k "cd /d ""%ROOT%\services\vps"" && title BossCloud - vps :8088 && go run .\cmd\vps"
timeout /t 2 /nobreak >nul

echo [9/10] hypervisor-agent
start "BossCloud-agent" cmd /k "cd /d ""%ROOT%\agents\hypervisor-agent"" && title BossCloud - hypervisor-agent && go run .\cmd\hypervisor-agent"
timeout /t 3 /nobreak >nul

echo [10/10] gateway :8080
start "BossCloud-gateway" cmd /k "cd /d ""%ROOT%\services\gateway"" && title BossCloud - gateway :8080 && go run .\cmd\gateway"

echo.
echo ============================================
echo   Listo - 10 ventanas abiertas
echo ============================================
echo Gateway: http://localhost:8080/healthz
echo Espera ~30s a que compile la primera vez.
echo.
pause
