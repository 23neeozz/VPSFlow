@echo off
echo Cerrando servicios VPSFlow...
taskkill /FI "WINDOWTITLE eq VPSFlow - auth :8081*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq VPSFlow - iam :8082*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq VPSFlow - tenant :8083*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq VPSFlow - cluster :8084*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq VPSFlow - agent-control :8086*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq VPSFlow - vm :8085*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq VPSFlow - console :8087*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq VPSFlow - hypervisor-agent*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq VPSFlow - gateway :8080*" /F >nul 2>&1
echo Hecho. Cierra manualmente cualquier ventana que quede abierta.
pause
