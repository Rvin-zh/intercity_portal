@echo off
setlocal enabledelayedexpansion

echo === Secure Sign In Database Setup ===

:: Set paths
set "APP_CONFIG_DIR=%USERPROFILE%\.config\secure-sign-in-app"
set "USER_HOME_DIR=%USERPROFILE%\.securesignin"
set "DB_PATH=%USER_HOME_DIR%\securesignin.db"
set "KEY_PATH=%USER_HOME_DIR%\encryption.key"
set "BACKUP_DIR=%APP_CONFIG_DIR%\backups"

:: Create all necessary directories
echo Creating application directories...
if not exist "%APP_CONFIG_DIR%" mkdir "%APP_CONFIG_DIR%"
if not exist "%USER_HOME_DIR%" mkdir "%USER_HOME_DIR%"
if not exist "%BACKUP_DIR%" mkdir "%BACKUP_DIR%"

:: Check for existing encryption key
if not exist "%KEY_PATH%" (
    echo No encryption key found, creating placeholder for app to use
    :: Use PowerShell to create a random key file since CMD doesn't have good tools for this
    powershell -Command "$bytes = New-Object byte[] 32; (New-Object System.Security.Cryptography.RNGCryptoServiceProvider).GetBytes($bytes); [System.IO.File]::WriteAllBytes('%KEY_PATH%', $bytes)"
    
    if errorlevel 1 (
        echo Failed to create key file. Please check permissions.
        exit /b 1
    )
)

echo Database setup complete. Your database will be stored at: %DB_PATH%
