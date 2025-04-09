@echo off
REM Build the Go backend for Windows

REM Get the absolute path of the project root directory
cd ..
set PROJECT_ROOT=%CD%
echo Project root directory: %PROJECT_ROOT%

echo Building Go backend for Windows...
go build -o main.exe .
if %ERRORLEVEL% neq 0 (
    echo Failed to build Go backend.
    exit /b 1
)
echo Go backend built successfully.

REM Return to electron-app directory
cd electron-app 