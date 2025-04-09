#!/bin/bash

# Linux Build Script for Secure Sign In Application
# This script automatically installs dependencies and builds the application for Linux

set -e  # Exit on any error

echo "Starting Linux build process..."

# Check if running from the electron-app directory
if [[ "$(basename "$(pwd)")" != "electron-app" ]]; then
  echo "This script must be run from the electron-app directory."
  echo "Please cd into the electron-app directory and try again."
  exit 1
fi

# Function to install Node.js and npm on Linux
install_nodejs_linux() {
  echo "Installing Node.js and npm..."
  
  # Check if we can use apt (Debian/Ubuntu)
  if command -v apt &> /dev/null; then
    sudo apt update
    sudo apt install -y nodejs npm
  # Check if we can use dnf (Fedora)
  elif command -v dnf &> /dev/null; then
    sudo dnf install -y nodejs npm
  # Check if we can use yum (CentOS/RHEL)
  elif command -v yum &> /dev/null; then
    sudo yum install -y nodejs npm
  # Check if we can use pacman (Arch)
  elif command -v pacman &> /dev/null; then
    sudo pacman -Sy nodejs npm
  else
    echo "Couldn't detect package manager. Please install Node.js manually."
    exit 1
  fi
}

# Function to install Go on Linux
install_go_linux() {
  echo "Installing Go..."
  
  # Check if we can use apt (Debian/Ubuntu)
  if command -v apt &> /dev/null; then
    sudo apt update
    sudo apt install -y golang
  # Check if we can use dnf (Fedora)
  elif command -v dnf &> /dev/null; then
    sudo dnf install -y golang
  # Check if we can use yum (CentOS/RHEL)
  elif command -v yum &> /dev/null; then
    sudo yum install -y golang
  # Check if we can use pacman (Arch)
  elif command -v pacman &> /dev/null; then
    sudo pacman -Sy go
  else
    echo "Couldn't detect package manager. Please install Go manually."
    exit 1
  fi
}

# Check if Node.js and npm are installed, install if not
if ! command -v node &> /dev/null || ! command -v npm &> /dev/null; then
  echo "Node.js or npm not found. Installing..."
  install_nodejs_linux
fi

# Check if Go is installed, install if not
if ! command -v go &> /dev/null; then
  echo "Go not found. Installing..."
  install_go_linux
fi

# Display versions
echo "Node.js version: $(node --version)"
echo "npm version: $(npm --version)"
echo "Go version: $(go version)"

# Install dependencies
echo "Installing Node.js dependencies..."
npm install
if [ $? -ne 0 ]; then
  echo "Failed to install Node.js dependencies."
  exit 1
fi
echo "Dependencies installed successfully."

# Build the Go backend
echo "Building Go backend for Linux..."
chmod +x build-backend.sh
./build-backend.sh
if [ $? -ne 0 ]; then
  echo "Failed to build Go backend."
  exit 1
fi
echo "Go backend built successfully."

# Package the application for Linux
echo "Packaging the application for Linux..."
npm run package-linux
if [ $? -ne 0 ]; then
  echo "Failed to package the application."
  exit 1
fi

echo "Application packaged successfully."
echo "The packaged application can be found in the dist/ directory."

# List the contents of the dist directory
echo "Contents of the dist directory:"
ls -la dist/

echo "Linux build process completed successfully!" 