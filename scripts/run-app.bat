@echo off
setlocal enabledelayedexpansion

:: Run script for Secure Sign In application

:: Set correct database path
set "SQLITE_DB_PATH=%USERPROFILE%\.securesignin\securesignin.db"

:: Run database setup script if it exists
if exist ".\scripts\windows-db-setup.bat" (
  call .\scripts\windows-db-setup.bat
)

:: Run the application
if exist "Secure Sign In.exe" (
  echo Starting Secure Sign In application...
  start "" "Secure Sign In.exe"
) else (
  echo Error: Application not found.
  echo Please ensure you're running this script from the application directory.
  exit /b 1
)
