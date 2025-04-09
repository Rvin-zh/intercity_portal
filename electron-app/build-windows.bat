@echo off
REM Windows Build Script for Secure Sign In Application
REM This script automatically installs dependencies and builds the application

echo Starting Windows build process...

REM Check if running from the electron-app directory
for %%I in (.) do set CURRENT_DIR=%%~nxI
if not "%CURRENT_DIR%"=="electron-app" (
    echo This script must be run from the electron-app directory.
    echo Please cd into the electron-app directory and try again.
    exit /b 1
)

REM Function to install Chocolatey package manager
:install_chocolatey
echo Checking for Chocolatey package manager...
where choco >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo Installing Chocolatey package manager...
    @powershell -NoProfile -ExecutionPolicy Bypass -Command "[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))"
    if %ERRORLEVEL% neq 0 (
        echo Failed to install Chocolatey. Please install it manually.
        exit /b 1
    )
    REM Refresh environment variables to use choco
    call RefreshEnv.cmd
) else (
    echo Chocolatey is already installed.
)

REM Check for Node.js and install if needed
where node >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo Node.js not found. Installing...
    choco install nodejs -y
    if %ERRORLEVEL% neq 0 (
        echo Failed to install Node.js. Please install it manually.
        exit /b 1
    )
    REM Refresh environment variables to use node
    call RefreshEnv.cmd
) else (
    echo Node.js is already installed.
)

REM Check for npm and install if needed
where npm >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo npm not found. It should have been installed with Node.js.
    echo Please check your Node.js installation.
    exit /b 1
)

REM Check for Go and install if needed
where go >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo Go not found. Installing...
    choco install golang -y
    if %ERRORLEVEL% neq 0 (
        echo Failed to install Go. Please install it manually.
        exit /b 1
    )
    REM Refresh environment variables to use go
    call RefreshEnv.cmd
) else (
    echo Go is already installed.
)

REM Display versions
echo Node.js version:
node --version
echo npm version:
npm --version
echo Go version:
go version

REM Install Node.js dependencies
echo Installing Node.js dependencies...
call npm install
if %ERRORLEVEL% neq 0 (
    echo Failed to install Node.js dependencies.
    exit /b 1
)
echo Dependencies installed successfully.

REM Check if the backend build script exists and make it executable
if not exist build-backend.bat (
    echo Creating build-backend.bat script...
    (
        echo @echo off
        echo REM Build the Go backend for Windows
        echo echo Building Go backend for Windows...
        echo cd ..
        echo go build -o main.exe .
        echo if %%ERRORLEVEL%% neq 0 (
        echo     echo Failed to build Go backend.
        echo     exit /b 1
        echo ^)
        echo echo Go backend built successfully.
        echo cd electron-app
    ) > build-backend.bat
)

REM Build the Go backend
echo Building Go backend for Windows...
call build-backend.bat
if %ERRORLEVEL% neq 0 (
    echo Failed to build Go backend.
    exit /b 1
)
echo Go backend built successfully.

REM Package the application for Windows
echo Packaging the application for Windows...
call npm run package-win
if %ERRORLEVEL% neq 0 (
    echo Failed to package the application.
    exit /b 1
)

echo Application packaged successfully.
echo The packaged application can be found in the dist/ directory.

REM List the contents of the dist directory
echo Contents of the dist directory:
dir dist

echo Build process completed successfully!
echo Press any key to exit...
pause > nul 