@echo off
echo Cerrando servicios BossCloud...
taskkill /FI "WINDOWTITLE eq BossCloud - auth :8081*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq BossCloud - iam :8082*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq BossCloud - tenant :8083*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq BossCloud - cluster :8084*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq BossCloud - agent-control :8086*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq BossCloud - vm :8085*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq BossCloud - console :8087*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq BossCloud - hypervisor-agent*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq BossCloud - gateway :8080*" /F >nul 2>&1
echo Hecho. Cierra manualmente cualquier ventana que quede abierta.
pause
