@echo off

REM Используем host.docker.internal для доступа к сервису на хосте из контейнера Docker
set TARGET_URL=http://host.docker.internal:8080/order/b563feb7b2b84b6test
set THREADS=4
set CONNECTIONS=150
set DURATION=20s

echo [INFO] Starting wrk load test...
echo [INFO] Target URL: %TARGET_URL%
echo [INFO] Threads: %THREADS%, Connections: %CONNECTIONS%, Duration: %DURATION%

docker run --rm williamyeh/wrk -t%THREADS% -c%CONNECTIONS% -d%DURATION% %TARGET_URL%

if %errorlevel% equ 0 (
    echo [INFO] wrk load test completed successfully.
) else (
    echo [ERROR] Error occurred during wrk load test.
    exit /b %errorlevel%
)