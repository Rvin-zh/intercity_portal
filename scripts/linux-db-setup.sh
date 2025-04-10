#!/bin/bash

echo "=== Secure Sign In Database Setup ==="

# Set paths
APP_CONFIG_DIR="$HOME/.config/secure-sign-in-app"
USER_HOME_DIR="$HOME/.securesignin"
DB_PATH="$USER_HOME_DIR/securesignin.db"
KEY_PATH="$USER_HOME_DIR/encryption.key"

# Create all necessary directories
echo "Creating application directories..."
mkdir -p "$APP_CONFIG_DIR"
mkdir -p "$USER_HOME_DIR"
mkdir -p "$APP_CONFIG_DIR/backups"

# Set directory permissions
chmod 755 "$APP_CONFIG_DIR"
chmod 755 "$USER_HOME_DIR"
chmod 755 "$APP_CONFIG_DIR/backups"

# Check for existing encryption key
if [ ! -f "$KEY_PATH" ]; then
  echo "No encryption key found, creating placeholder for app to use"
  dd if=/dev/urandom bs=32 count=1 of="$KEY_PATH" 2>/dev/null
  if [ $? -ne 0 ]; then
    echo "Failed to create key file. Please check permissions."
    exit 1
  fi
  chmod 600 "$KEY_PATH"
fi

echo "Database setup complete. Your database will be stored at: $DB_PATH"
