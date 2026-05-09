@echo off
setlocal

echo Starting Game Backend...
echo.

echo Running go mod tidy...
go mod tidy

echo.
echo Building...
go build -o bin\game_backend.exe cmd\game_backend\main.go

if %errorlevel% neq 0 (
    echo Build failed!
    exit /b 1
)

echo.
echo Starting servers...
bin\game_backend.exe --config .\configs
