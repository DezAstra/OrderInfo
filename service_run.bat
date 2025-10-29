@echo off
REM run_service.bat - Скрипт для запуска order-display-service с предварительной настройкой среды

REM 1. Остановка локального PostgreSQL (если запущен как служба Windows)
echo [INFO] Остановка локальной службы PostgreSQL ...
net stop postgresql-x64-17 >nul 2>&1
if %errorlevel% equ 0 (
    echo [INFO] Служба PostgreSQL остановлена.
) else (
    echo [INFO] Служба PostgreSQL не запущена или не найдена. Продолжаем...
)

REM 2. Запуск Docker Compose (контейнеры order-postgres и nats-streaming)
echo [INFO] Запуск Docker Compose (order-postgres, nats-streaming)...
docker-compose up -d
if %errorlevel% neq 0 (
    echo [ERROR] Не удалось запустить docker-compose. Проверьте, запущен ли Docker Desktop.
    pause
    exit /b 1
)
echo [INFO] Docker Compose запущен в фоновом режиме.

REM 3. Ожидание запуска контейнеров
echo [INFO] Ожидание 10 секунд для полного запуска контейнеров...
timeout /t 10 /nobreak >nul

REM 4. Запуск сервиса 
echo [INFO] Запуск order-display-service...
go run main.go
if %errorlevel% neq 0 (
    echo [ERROR] Ошибка при запуске order-display-service.
    REM Останавливаем контейнеры перед выходом из-за ошибки
    docker-compose down
    pause
    exit /b %errorlevel%
)

REM 5. Остановка контейнеров после завершения работы сервиса
echo [INFO] Остановка контейнеров Docker Compose...
docker-compose down
echo [INFO] Контейнеры остановлены.

echo [INFO] Работа скрипта завершена.
pause
