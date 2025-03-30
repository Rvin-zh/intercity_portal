#!/bin/bash

# This script runs the Docker containers with improved networking

# Ensure the script stops on errors
set -e

# Step 1: Clean up any old containers
echo "Cleaning up any old containers..."
sudo docker compose down || true

# Step 2: Make sure we have vendor directory
if [ ! -d "vendor" ]; then
  echo "Vendor directory not found, creating it..."
  ./prepare-build.sh
fi

# Step 3: Build and run the containers
echo "Building and running the containers..."
sudo docker compose up --build $@

# Note: if you want to run in detached mode, use ./run-docker.sh -d
# To view logs in detached mode, use sudo docker compose logs -f