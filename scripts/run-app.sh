#!/bin/bash
# Run script for Secure Sign In application

# Set correct database path
export SQLITE_DB_PATH="$HOME/.securesignin/securesignin.db"

# Run database setup script if it exists
if [ -f "./scripts/linux-db-setup.sh" ]; then
  ./scripts/linux-db-setup.sh
fi

# Run the application
if [ -f "./Secure Sign In-1.0.0.AppImage" ]; then
  echo "Starting Secure Sign In application..."
  ./Secure\ Sign\ In-1.0.0.AppImage
else
  echo "Error: Application not found."
  echo "Please ensure you're running this script from the application directory."
  exit 1
fi
