#!/bin/bash

# This script prepares the Go modules for Docker build to avoid network issues
echo "Setting up Go vendor directory..."

# Ensure the script fails on errors
set -e

# Disable Go checksum database to avoid network issues
export GOSUMDB=off
export GOPROXY=direct

# Download Go modules
echo "Downloading Go modules..."
go mod download || true

# Create vendor directory
echo "Creating vendor directory..."
go mod vendor || true

# Check if vendor directory was created
if [ -d "vendor" ]; then
  echo "Vendor directory created successfully"
  ls -la vendor/
else
  echo "Failed to create vendor directory, but continuing anyway"
fi

echo "Build preparation completed"
echo ""
echo "INSTRUCTIONS FOR DOCKER BUILD:"
echo "-------------------------------"
echo "1. Run this script to create the vendor directory with all dependencies"
echo "2. Use 'sudo docker compose up --build' to build and run the containers"
echo "3. If you still have networking issues, try these solutions:"
echo "   - Add DNS config: 'dns: [8.8.8.8, 8.8.4.4]' to the app service in docker-compose.yml"
echo "   - Set GOPROXY=off and GOSUMDB=off in the Dockerfile"
echo "   - Use '-mod=vendor' flag with 'go build' in the Dockerfile"
echo "   - Add 'nameserver 8.8.8.8' and 'nameserver 8.8.4.4' to /etc/resolv.conf in the container"
echo ""