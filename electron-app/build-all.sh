#!/bin/bash

# Master build script for the Secure Sign In Desktop Application
# This script builds the Go backend and packages the Electron app for Linux and Windows

set -e  # Exit on any error

# Display current working directory
echo "Current directory: $(pwd)"

# Check if running from the electron-app directory
if [[ "$(basename "$(pwd)")" != "electron-app" ]]; then
  echo "This script must be run from the electron-app directory."
  echo "Please cd into the electron-app directory and try again."
  exit 1
fi

# Check if Node.js and npm are installed
if ! command -v node &> /dev/null; then
  echo "Node.js is not installed. Please install Node.js and try again."
  exit 1
fi

if ! command -v npm &> /dev/null; then
  echo "npm is not installed. Please install npm and try again."
  exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
  echo "Go is not installed. Please install Go and try again."
  exit 1
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
echo "Building Go backend for Linux and Windows..."
chmod +x build-backend.sh
./build-backend.sh
if [ $? -ne 0 ]; then
  echo "Failed to build Go backend."
  exit 1
fi
echo "Go backend built successfully."

# Check which platforms to build for
BUILD_LINUX=1
BUILD_WINDOWS=1

# Check if Wine is installed if building for Windows
if [ $BUILD_WINDOWS -eq 1 ] && ! command -v wine &> /dev/null; then
  echo "Warning: Wine is not installed. Windows builds might fail."
  echo "Do you want to continue anyway? (y/n)"
  read -r response
  if [[ "$response" != "y" ]]; then
    echo "Aborting Windows build."
    BUILD_WINDOWS=0
  fi
fi

# Package the application
echo "Packaging the application..."

if [ $BUILD_LINUX -eq 1 ] && [ $BUILD_WINDOWS -eq 1 ]; then
  # Build for both platforms
  echo "Building for both Linux and Windows..."
  npm run package-all
elif [ $BUILD_LINUX -eq 1 ]; then
  # Build for Linux only
  echo "Building for Linux only..."
  npm run package-linux
elif [ $BUILD_WINDOWS -eq 1 ]; then
  # Build for Windows only
  echo "Building for Windows only..."
  npm run package-win
else
  echo "No platforms selected for building. Exiting."
  exit 1
fi

if [ $? -ne 0 ]; then
  echo "Failed to package the application."
  exit 1
fi

echo "Application packaged successfully."
echo "The packaged applications can be found in the dist/ directory."

# List the contents of the dist directory
echo "Contents of the dist directory:"
ls -la dist/

echo "Build process completed successfully!" 