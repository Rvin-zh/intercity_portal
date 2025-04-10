#!/usr/bin/env node

/**
 * Database Preservation Tool for Secure Sign In
 * 
 * This script backs up the database before building and restores it afterward
 * to ensure data is not lost during rebuilds.
 */

const fs = require('fs');
const path = require('path');
const os = require('os');

// Define paths
const HOME_DIR = os.homedir();
const DB_PATH = path.join(HOME_DIR, '.securesignin', 'securesignin.db');
const BACKUP_DIR = path.join(HOME_DIR, '.config', 'secure-sign-in-app', 'backups');
const TEMP_BACKUP_PATH = path.join(BACKUP_DIR, 'pre-build-backup.db');

// Process arguments
const args = process.argv.slice(2);
const operation = args[0] || 'backup'; // Default to backup

// Ensure backup directory exists
function ensureBackupDirExists() {
  if (!fs.existsSync(BACKUP_DIR)) {
    fs.mkdirSync(BACKUP_DIR, { recursive: true });
    console.log(`Created backup directory: ${BACKUP_DIR}`);
  }
}

// Backup the database before build
function backupDatabase() {
  if (!fs.existsSync(DB_PATH)) {
    console.log(`No database found at ${DB_PATH}. Nothing to backup.`);
    return false;
  }

  ensureBackupDirExists();

  try {
    fs.copyFileSync(DB_PATH, TEMP_BACKUP_PATH);
    console.log(`Successfully backed up database to ${TEMP_BACKUP_PATH}`);
    return true;
  } catch (error) {
    console.error(`Error backing up database: ${error.message}`);
    return false;
  }
}

// Restore the database after build
function restoreDatabase() {
  if (!fs.existsSync(TEMP_BACKUP_PATH)) {
    console.log(`No backup found at ${TEMP_BACKUP_PATH}. Nothing to restore.`);
    return false;
  }

  // Ensure the target directory exists
  const dbDir = path.dirname(DB_PATH);
  if (!fs.existsSync(dbDir)) {
    fs.mkdirSync(dbDir, { recursive: true });
    console.log(`Created database directory: ${dbDir}`);
  }

  try {
    fs.copyFileSync(TEMP_BACKUP_PATH, DB_PATH);
    console.log(`Successfully restored database to ${DB_PATH}`);
    
    // Make a timestamp backup as well
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const timestampBackup = path.join(BACKUP_DIR, `backup-${timestamp}.db`);
    fs.copyFileSync(TEMP_BACKUP_PATH, timestampBackup);
    console.log(`Created timestamped backup at ${timestampBackup}`);
    
    return true;
  } catch (error) {
    console.error(`Error restoring database: ${error.message}`);
    return false;
  }
}

// Main function
function main() {
  if (operation === 'backup') {
    console.log('=== Backing up database before build ===');
    backupDatabase();
  } else if (operation === 'restore') {
    console.log('=== Restoring database after build ===');
    restoreDatabase();
  } else {
    console.error(`Unknown operation: ${operation}`);
    console.log('Usage: node preserve-data.js [backup|restore]');
    process.exit(1);
  }
}

// Run the main function
main(); 