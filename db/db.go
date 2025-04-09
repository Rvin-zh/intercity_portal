package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

var DB *sql.DB

// InitializeDB sets up the SQLite database connection
func InitializeDB() error {
	var err error

	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Join(homeDir, ".config", "secure-sign-in-app")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set database path
	dbPath := filepath.Join(configDir, "securesignin.db")
	log.Printf("Initializing SQLite database at: %s", dbPath)

	// Add foreign_keys=on for SQLite
	DB, err = sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Set pragmatic WAL mode for better concurrency
	_, err = DB.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		log.Printf("Warning: Failed to set WAL mode for SQLite: %v", err)
	}

	// Required for SQLite with WAL mode
	DB.SetMaxOpenConns(1)

	// Check connection
	pingErr := DB.Ping()
	if pingErr != nil {
		DB.Close() // Ensure connection is closed if ping fails
		return fmt.Errorf("failed to connect to database: %w", pingErr)
	}

	log.Println("Database connection successful. Initializing schema...")
	err = initializeSchema()
	if err != nil {
		DB.Close()
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// initializeSchema creates the necessary tables if they don't exist
func initializeSchema() error {
	usersTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TEXT DEFAULT CURRENT_TIMESTAMP
	);
	`

	loginHistoryTableSQL := `
	CREATE TABLE IF NOT EXISTS login_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		login_time TEXT DEFAULT CURRENT_TIMESTAMP,
		ip_address TEXT,
		success BOOLEAN,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`

	_, err := DB.Exec(usersTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	_, err = DB.Exec(loginHistoryTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create login_history table: %w", err)
	}

	// Add unique index separately for better cross-DB compatibility
	usernameIndexSQL := `CREATE UNIQUE INDEX IF NOT EXISTS idx_username ON users(username);`
	_, err = DB.Exec(usernameIndexSQL)
	if err != nil {
		log.Printf("Warning: Could not ensure username index (might already exist): %v", err)
	}

	return nil
}

// AddUser inserts a new user into the database
func AddUser(username, passwordHash string) (int64, error) {
	query := "INSERT INTO users (username, password_hash) VALUES (?, ?)"
	result, err := DB.Exec(query, username, passwordHash)
	if err != nil {
		return 0, fmt.Errorf("error inserting user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting last insert ID: %w", err)
	}
	return id, nil
}

// GetUserByUsername retrieves a user by their username
func GetUserByUsername(username string) (*sql.Row, error) {
	query := "SELECT id, username, password_hash FROM users WHERE username = ?"
	row := DB.QueryRow(query, username)
	return row, nil
}

// GetUserByID retrieves a user by their ID
func GetUserByID(userID int) (*sql.Row, error) {
	query := "SELECT id, username, password_hash FROM users WHERE id = ?"
	row := DB.QueryRow(query, userID)
	return row, nil
}

// UpdateUserPassword updates a user's password hash
func UpdateUserPassword(userID int, newPasswordHash string) error {
	query := "UPDATE users SET password_hash = ? WHERE id = ?"
	_, err := DB.Exec(query, newPasswordHash, userID)
	if err != nil {
		return fmt.Errorf("error updating password for user ID %d: %w", userID, err)
	}
	return nil
}

// LogLoginAttempt records a login attempt
func LogLoginAttempt(userID int, ipAddress string, success bool) error {
	query := "INSERT INTO login_history (user_id, ip_address, success) VALUES (?, ?, ?)"
	_, err := DB.Exec(query, userID, ipAddress, success)
	if err != nil {
		return fmt.Errorf("error logging login attempt for user ID %d: %w", userID, err)
	}
	return nil
}

// GetAllUsers retrieves all users
func GetAllUsers() (*sql.Rows, error) {
	rows, err := DB.Query("SELECT id, username, created_at FROM users ORDER BY username")
	if err != nil {
		return nil, fmt.Errorf("error retrieving all users: %w", err)
	}
	return rows, nil
}

// GetLoginHistory retrieves all login history records
func GetLoginHistory() (*sql.Rows, error) {
	query := `
	SELECT lh.id, u.username, lh.login_time, lh.ip_address, lh.success
	FROM login_history lh
	JOIN users u ON lh.user_id = u.id
	ORDER BY lh.login_time DESC
	LIMIT 100
	`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error retrieving login history: %w", err)
	}
	return rows, nil
}
