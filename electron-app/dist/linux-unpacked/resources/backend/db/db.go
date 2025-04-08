package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"           // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

var DB *sql.DB

// InitializeDB sets up the database connection based on environment variables.
// It prioritizes SQLite if USE_SQLITE=1 is set, otherwise uses PostgreSQL.
func InitializeDB() error {
	var err error
	useSqlite := os.Getenv("USE_SQLITE") == "1"

	if useSqlite {
		sqlitePath := os.Getenv("SQLITE_PATH")
		if sqlitePath == "" {
			// Default to user config directory if path not specified
			configDir, err := os.UserConfigDir()
			if err != nil {
				log.Printf("Warning: Could not get user config dir: %v. Using current directory.", err)
				configDir = "."
			}
			appDir := filepath.Join(configDir, "SecureSignIn")
			if err := os.MkdirAll(appDir, 0750); err != nil {
				return fmt.Errorf("failed to create app config directory %s: %w", appDir, err)
			}
			sqlitePath = filepath.Join(appDir, "securesignin.db")
		}

		log.Printf("Initializing SQLite database at: %s", sqlitePath)
		// Add foreign_keys=on for SQLite
		DB, err = sql.Open("sqlite3", sqlitePath+"?_foreign_keys=on")
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

	} else {
		// PostgreSQL connection logic
		dbHost := os.Getenv("DB_HOST")
		dbPort := os.Getenv("DB_PORT")
		dbUser := os.Getenv("DB_USER")
		dbPassword := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")

		if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" {
			log.Println("Warning: Missing one or more PostgreSQL environment variables (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME). Attempting default connection.")
			dbHost = "localhost"
			dbPort = "5432"
			dbUser = "postgres"
			dbPassword = "postgres"
			dbName = "transport_db"
		}

		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbPassword, dbName)

		log.Printf("Connecting to PostgreSQL at %s:%s database=%s user=%s", dbHost, dbPort, dbName, dbUser)
		DB, err = sql.Open("postgres", dsn)
		if err != nil {
			return fmt.Errorf("failed to open PostgreSQL database: %w", err)
		}

		DB.SetMaxOpenConns(25)
		DB.SetMaxIdleConns(25)
		DB.SetConnMaxLifetime(5 * time.Minute)
	}

	// Check connection
	pingErr := DB.Ping()
	if pingErr != nil {
		DB.Close() // Ensure connection is closed if ping fails
		return fmt.Errorf("failed to connect to database: %w", pingErr)
	}

	log.Println("Database connection successful. Initializing schema...")
	err = initializeSchema(useSqlite)
	if err != nil {
		DB.Close()
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// initializeSchema creates the necessary tables if they don't exist.
func initializeSchema(isSqlite bool) error {
	var usersTableSQL string
	var loginHistoryTableSQL string

	// Adjust SQL based on the database type
	if isSqlite {
		// SQLite specific syntax
		usersTableSQL = `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		);
		`
		loginHistoryTableSQL = `
		CREATE TABLE IF NOT EXISTS login_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			login_time TEXT DEFAULT CURRENT_TIMESTAMP,
			ip_address TEXT,
			success BOOLEAN,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
		`
	} else {
		// PostgreSQL specific syntax
		usersTableSQL = `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
		`
		loginHistoryTableSQL = `
		CREATE TABLE IF NOT EXISTS login_history (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			login_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			ip_address VARCHAR(50),
			success BOOLEAN,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
		`
	}

	_, err := DB.Exec(usersTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	_, err = DB.Exec(loginHistoryTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create login_history table: %w", err)
	}

	// Add unique index separately for better cross-DB compatibility if needed,
	// though UNIQUE constraint in CREATE TABLE is usually sufficient.
	usernameIndexSQL := `CREATE UNIQUE INDEX IF NOT EXISTS idx_username ON users(username);`
	_, err = DB.Exec(usernameIndexSQL)
	if err != nil {
		log.Printf("Warning: Could not ensure username index (might already exist): %v", err)
	}

	return nil
}

// placeholderValue returns the correct placeholder ($N for postgres, ? for sqlite).
// Note: This simple version assumes only one placeholder per query for clarity.
// For complex queries, consider a more robust approach or a query builder library.
func placeholderValue(index int) string {
	if os.Getenv("USE_SQLITE") == "1" {
		return "?"
	}
	return fmt.Sprintf("$%d", index)
}

// AddUser inserts a new user into the database.
func AddUser(username, passwordHash string) (int64, error) {
	// Use generic placeholder ? which works for SQLite and often PostgreSQL drivers too
	// Or adjust based on DB type if necessary
	query := fmt.Sprintf("INSERT INTO users (username, password_hash) VALUES (%s, %s)", placeholderValue(1), placeholderValue(2))
	result, err := DB.Exec(query, username, passwordHash)
	if err != nil {
		return 0, fmt.Errorf("error inserting user: %w", err)
	}

	// LastInsertId might not be supported/reliable on PostgreSQL without RETURNING id
	id, err := result.LastInsertId()
	if err != nil {
		// Attempt to get rows affected as fallback or log warning
		rowsAffected, _ := result.RowsAffected()
		log.Printf("Could not get LastInsertId (RowsAffected: %d): %v", rowsAffected, err)
		// Depending on use case, you might return 0 or an error
		return 0, nil // Return 0 if ID retrieval isn't critical or supported
	}
	return id, nil
}

// GetUserByUsername retrieves a user by their username.
func GetUserByUsername(username string) (*sql.Row, error) {
	query := fmt.Sprintf("SELECT id, username, password_hash FROM users WHERE username = %s", placeholderValue(1))
	row := DB.QueryRow(query, username)
	// Error checking is deferred to Scan in the caller
	return row, nil
}

// GetUserByID retrieves a user by their ID.
func GetUserByID(userID int) (*sql.Row, error) {
	query := fmt.Sprintf("SELECT id, username, password_hash FROM users WHERE id = %s", placeholderValue(1))
	row := DB.QueryRow(query, userID)
	// Error checking deferred to Scan
	return row, nil
}

// UpdateUserPassword updates a user's password hash.
func UpdateUserPassword(userID int, newPasswordHash string) error {
	query := fmt.Sprintf("UPDATE users SET password_hash = %s WHERE id = %s", placeholderValue(1), placeholderValue(2))
	_, err := DB.Exec(query, newPasswordHash, userID)
	if err != nil {
		return fmt.Errorf("error updating password for user ID %d: %w", userID, err)
	}
	return nil
}

// LogLoginAttempt records a login attempt.
func LogLoginAttempt(userID int, ipAddress string, success bool) error {
	query := fmt.Sprintf("INSERT INTO login_history (user_id, ip_address, success) VALUES (%s, %s, %s)", placeholderValue(1), placeholderValue(2), placeholderValue(3))
	_, err := DB.Exec(query, userID, ipAddress, success)
	if err != nil {
		return fmt.Errorf("error logging login attempt for user ID %d: %w", userID, err)
	}
	return nil
}

// GetAllUsers retrieves all users.
func GetAllUsers() (*sql.Rows, error) {
	rows, err := DB.Query("SELECT id, username, created_at FROM users ORDER BY username")
	if err != nil {
		return nil, fmt.Errorf("error retrieving all users: %w", err)
	}
	return rows, nil
}

// GetLoginHistory retrieves all login history records.
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
