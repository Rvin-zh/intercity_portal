package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// BackupDatabase creates a backup of the SQLite database file
func BackupDatabase(dbPath string) (string, error) {
	if dbPath == "" {
		dbPath = filepath.Join("data", "securesignin.db")
	}

	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return "", fmt.Errorf("database file not found at %s", dbPath)
	}

	// Create backup directory if it doesn't exist
	backupDir := filepath.Join(filepath.Dir(dbPath), "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Create backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupFilename := fmt.Sprintf("securesignin-%s.db.bak", timestamp)
	backupPath := filepath.Join(backupDir, backupFilename)

	// Copy database file to backup location
	srcFile, err := os.Open(dbPath)
	if err != nil {
		return "", fmt.Errorf("failed to open source database file: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer destFile.Close()

	// Perform the copy
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return "", fmt.Errorf("failed to copy database file: %w", err)
	}

	// Set file permissions to be readonly
	if err := os.Chmod(backupPath, 0444); err != nil {
		log.Printf("Warning: Failed to set backup file permissions: %v", err)
	}

	log.Printf("Successfully created database backup at %s", backupPath)
	return backupPath, nil
}

// ScheduleBackups sets up a ticker to perform regular database backups
func ScheduleBackups(dbPath string, intervalHours int) {
	if intervalHours <= 0 {
		intervalHours = 24 // Default to daily backups
	}

	// Set up a ticker to run at the specified interval
	ticker := time.NewTicker(time.Duration(intervalHours) * time.Hour)
	go func() {
		for range ticker.C {
			backupPath, err := BackupDatabase(dbPath)
			if err != nil {
				log.Printf("Scheduled backup failed: %v", err)
			} else {
				log.Printf("Scheduled backup completed: %s", backupPath)
			}

			// Clean up old backups (keep only the last 7)
			cleanupOldBackups(filepath.Join(filepath.Dir(dbPath), "backups"), 7)
		}
	}()

	log.Printf("Database backup scheduler started with %d hour interval", intervalHours)
}

// cleanupOldBackups removes backup files older than the most recent 'keep' files
func cleanupOldBackups(backupDir string, keep int) {
	files, err := filepath.Glob(filepath.Join(backupDir, "securesignin-*.db.bak"))
	if err != nil {
		log.Printf("Failed to list backup files: %v", err)
		return
	}

	// Return if we don't have more than 'keep' backups
	if len(files) <= keep {
		return
	}

	// Sort files by modification time (newest first)
	type fileInfo struct {
		path    string
		modTime time.Time
	}

	fileInfos := make([]fileInfo, 0, len(files))
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			log.Printf("Failed to stat file %s: %v", f, err)
			continue
		}
		fileInfos = append(fileInfos, fileInfo{path: f, modTime: info.ModTime()})
	}

	// Sort by modification time (newest first)
	for i := 0; i < len(fileInfos); i++ {
		for j := i + 1; j < len(fileInfos); j++ {
			if fileInfos[i].modTime.Before(fileInfos[j].modTime) {
				fileInfos[i], fileInfos[j] = fileInfos[j], fileInfos[i]
			}
		}
	}

	// Delete older backups
	for i := keep; i < len(fileInfos); i++ {
		if err := os.Remove(fileInfos[i].path); err != nil {
			log.Printf("Failed to remove old backup %s: %v", fileInfos[i].path, err)
		} else {
			log.Printf("Removed old backup: %s", fileInfos[i].path)
		}
	}
}
