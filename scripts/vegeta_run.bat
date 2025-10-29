@echo off

REM Используем host.docker.internal для доступа к сервису на хосте из контейнера Docker
set TARGET_URL=http://host.docker.internal:8080/order/b563feb7b2b84b6test
set RATE=1000
set DURATION=20s

echo [INFO] start vegeta test...
echo [INFO] cURL: %TARGET_URL%
echo [INFO] Rate: %RATE% RPS, Duration: %DURATION%

REM Передаём команду GET напрямую в vegeta attack через stdin
echo GET %TARGET_URL% | docker run --rm -i peterevans/vegeta vegeta attack -duration=%DURATION% -rate=%RATE% | docker run --rm -i peterevans/vegeta vegeta report

if %errorlevel% equ 0 (
    echo [INFO] vegeta test completed successfully.
) else (
    echo [ERROR] ERROR vegeta TEST.
    exit /b %errorlevel%
)
