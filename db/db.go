package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var DB *sql.DB

// InitializeDB sets up the PostgreSQL database connection using environment variables.
func InitializeDB() error {
	var err error

	// PostgreSQL connection logic using environment variables from docker-compose
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		log.Println("Warning: Missing one or more PostgreSQL environment variables (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME). Defaulting to standard values.")
		dbHost = "db" // Default service name in docker-compose
		dbPort = "5432"
		dbUser = "postgres"
		dbPassword = "postgres"
		dbName = "transport_db" // Matching docker-compose
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

// initializeSchema creates the necessary tables and adds missing columns (PostgreSQL syntax).
func initializeSchema() error {
	// Create users table with all essential columns, including NOT NULL
	usersTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		password TEXT NOT NULL, -- Included directly
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := DB.Exec(usersTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Add optional/later columns individually using IF NOT EXISTS (PostgreSQL 9.6+)
	addColumnSQLs := []string{
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS date_of_birth DATE`,   // Nullable for now
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS social_security TEXT`, // Nullable for now
	}

	for _, sql := range addColumnSQLs {
		_, err = DB.Exec(sql)
		if err != nil {
			log.Printf("Warning: Failed to execute ALTER TABLE statement [%s]: %v", sql, err)
		}
	}

	// Create login_history table if it doesn't exist
	loginHistoryTableSQL := `
	CREATE TABLE IF NOT EXISTS login_history (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		login_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		ip_address VARCHAR(50),
		success BOOLEAN,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`
	_, err = DB.Exec(loginHistoryTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create login_history table: %w", err)
	}

	// Add unique index separately
	usernameIndexSQL := `CREATE UNIQUE INDEX IF NOT EXISTS idx_username ON users(username);`
	_, err = DB.Exec(usernameIndexSQL)
	if err != nil {
		log.Printf("Warning: Could not ensure username index (might already exist): %v", err)
	}

	return nil
}

// AddUser inserts a new user into the database (PostgreSQL).
func AddUser(username, passwordHash, dob, ssn string) (int64, error) {
	var userID int64
	query := "INSERT INTO users (username, password, date_of_birth, social_security) VALUES ($1, $2, $3, $4) RETURNING id"
	err := DB.QueryRow(query, username, passwordHash, dob, ssn).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("error inserting user and getting ID (postgres): %w", err)
	}
	return userID, nil
}

// GetUserByUsername retrieves a user by their username (PostgreSQL).
func GetUserByUsername(username string) (*sql.Row, error) {
	query := "SELECT id, username, date_of_birth, social_security, password FROM users WHERE username = $1"
	row := DB.QueryRow(query, username)
	return row, nil // Error checking deferred to Scan
}

// GetUserByID retrieves a user by their ID (PostgreSQL).
// Note: This func still returns only id, username, password. Update if dob/ssn needed.
func GetUserByID(userID int) (*sql.Row, error) {
	query := "SELECT id, username, password FROM users WHERE id = $1"
	row := DB.QueryRow(query, userID)
	return row, nil // Error checking deferred to Scan
}

// UpdateUserPassword updates a user's password hash (PostgreSQL).
// Note: userID param is int, but DB might store as int64. Ensure type consistency.
func UpdateUserPassword(userID int, newPasswordHash string) error {
	query := "UPDATE users SET password = $1 WHERE id = $2"
	_, err := DB.Exec(query, newPasswordHash, userID)
	if err != nil {
		return fmt.Errorf("error updating password for user ID %d: %w", userID, err)
	}
	return nil
}

// LogLoginAttempt records a login attempt (PostgreSQL).
func LogLoginAttempt(userID int64, ipAddress string, success bool) error {
	query := "INSERT INTO login_history (user_id, ip_address, success) VALUES ($1, $2, $3)"
	_, err := DB.Exec(query, userID, ipAddress, success)
	if err != nil {
		return fmt.Errorf("error logging login attempt for user ID %d: %w", userID, err)
	}
	return nil
}

// GetAllUsers retrieves all users (PostgreSQL).
func GetAllUsers() (*sql.Rows, error) {
	rows, err := DB.Query("SELECT id, username, created_at FROM users ORDER BY username")
	if err != nil {
		return nil, fmt.Errorf("error retrieving all users: %w", err)
	}
	return rows, nil
}

// GetLoginHistory retrieves all login history records (PostgreSQL).
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
