package db

import (
        "database/sql"
        "fmt"
        "log"
        "os"
        "time"

        _ "github.com/lib/pq"
)

var db *sql.DB

// InitializeDB initializes the database connection and creates tables if they don't exist
func InitializeDB() error {
        var err error
        
        // Get PostgreSQL connection details from environment variables
        // Use Replit-provided environment variables if available, otherwise fall back to Docker ones
        host := getEnv("PGHOST", getEnv("DB_HOST", "localhost"))
        port := getEnv("PGPORT", getEnv("DB_PORT", "5432"))
        user := getEnv("PGUSER", getEnv("DB_USER", "postgres"))
        password := getEnv("PGPASSWORD", getEnv("DB_PASSWORD", "postgres"))
        dbname := getEnv("PGDATABASE", getEnv("DB_NAME", "postgres"))
        
        // Log connection attempt for debugging
        log.Printf("Connecting to PostgreSQL at %s:%s database=%s user=%s", host, port, dbname, user)
        
        // Construct PostgreSQL connection string
        // Determine SSL mode based on environment
        // Use sslmode=require for Replit's managed PostgreSQL, but sslmode=disable for Docker
        sslMode := "disable"
        if os.Getenv("REPL_ID") != "" || os.Getenv("PGHOST") != "" {
                sslMode = "require"
        }
        
        connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
                host, port, user, password, dbname, sslMode)
        
        // Open the database connection
        db, err = sql.Open("postgres", connStr)
        if err != nil {
                return fmt.Errorf("failed to open database: %v", err)
        }

        // Test the connection
        err = db.Ping()
        if err != nil {
                return fmt.Errorf("failed to connect to database: %v", err)
        }

        // Create users table if it doesn't exist
        createUsersTable := `
        CREATE TABLE IF NOT EXISTS users (
                id SERIAL PRIMARY KEY,
                username TEXT NOT NULL UNIQUE,
                password TEXT NOT NULL,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );`

        _, err = db.Exec(createUsersTable)
        if err != nil {
                return fmt.Errorf("failed to create users table: %v", err)
        }

        // Create login_logs table if it doesn't exist
        createLoginLogsTable := `
        CREATE TABLE IF NOT EXISTS login_logs (
                id SERIAL PRIMARY KEY,
                username TEXT NOT NULL,
                login_time TIMESTAMP NOT NULL,
                success BOOLEAN NOT NULL,
                ip_address TEXT
        );`

        _, err = db.Exec(createLoginLogsTable)
        if err != nil {
                return fmt.Errorf("failed to create login_logs table: %v", err)
        }

        // Insert a default admin user if no users exist
        var count int
        err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
        if err != nil {
                return fmt.Errorf("failed to check users count: %v", err)
        }

        if count == 0 {
                // In a real application, you would hash the password
                _, err = db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", "admin", "password")
                if err != nil {
                        return fmt.Errorf("failed to insert default user: %v", err)
                }
                log.Println("Default admin user created")
        }

        log.Println("Database initialized successfully")
        return nil
}

// Helper function to get environment variables with a default value
func getEnv(key, defaultValue string) string {
        value := os.Getenv(key)
        if value == "" {
                return defaultValue
        }
        return value
}

// ValidateUser checks if the username and password match a record in the database
func ValidateUser(username, password string) (bool, error) {
        var storedPassword string
        
        // Query the database for the user
        err := db.QueryRow("SELECT password FROM users WHERE username = $1", username).Scan(&storedPassword)
        if err != nil {
                if err == sql.ErrNoRows {
                        return false, nil // User not found
                }
                return false, fmt.Errorf("database error during validation: %v", err)
        }
        
        // In a real application, you would use proper password hashing (like bcrypt)
        return storedPassword == password, nil
}

// LogLoginAttempt records a login attempt in the database
func LogLoginAttempt(username string, success bool, ipAddress string) error {
        _, err := db.Exec(
                "INSERT INTO login_logs (username, login_time, success, ip_address) VALUES ($1, $2, $3, $4)",
                username,
                time.Now(),
                success,
                ipAddress,
        )
        if err != nil {
                return fmt.Errorf("failed to log login attempt: %v", err)
        }
        return nil
}

// UserExists checks if a username already exists in the database
func UserExists(username string) (bool, error) {
        var count int
        err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", username).Scan(&count)
        if err != nil {
                return false, fmt.Errorf("failed to check if user exists: %v", err)
        }
        return count > 0, nil
}

// CreateUser adds a new user to the database
func CreateUser(username, password string) error {
        // Check if user already exists
        exists, err := UserExists(username)
        if err != nil {
                return err
        }
        if exists {
                return fmt.Errorf("username already exists")
        }

        // In a real application, you would hash the password
        _, err = db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", username, password)
        if err != nil {
                return fmt.Errorf("failed to create user: %v", err)
        }
        return nil
}

// GetAllUsers returns all users from the database
func GetAllUsers() ([]map[string]interface{}, error) {
        rows, err := db.Query("SELECT id, username, created_at FROM users")
        if err != nil {
                return nil, fmt.Errorf("failed to query users: %v", err)
        }
        defer rows.Close()

        var users []map[string]interface{}
        for rows.Next() {
                var id int
                var username string
                var createdAt time.Time
                
                err := rows.Scan(&id, &username, &createdAt)
                if err != nil {
                        return nil, fmt.Errorf("failed to scan user row: %v", err)
                }
                
                user := map[string]interface{}{
                        "id":         id,
                        "username":   username,
                        "created_at": createdAt,
                }
                users = append(users, user)
        }

        if err = rows.Err(); err != nil {
                return nil, fmt.Errorf("error iterating user rows: %v", err)
        }

        return users, nil
}

// GetRecentLoginLogs returns the most recent login logs
func GetRecentLoginLogs(limit int) ([]map[string]interface{}, error) {
        rows, err := db.Query("SELECT id, username, login_time, success, ip_address FROM login_logs ORDER BY login_time DESC LIMIT $1", limit)
        if err != nil {
                return nil, fmt.Errorf("failed to query login logs: %v", err)
        }
        defer rows.Close()

        var logs []map[string]interface{}
        for rows.Next() {
                var id int
                var username string
                var loginTime time.Time
                var success bool
                var ipAddress string
                
                err := rows.Scan(&id, &username, &loginTime, &success, &ipAddress)
                if err != nil {
                        return nil, fmt.Errorf("failed to scan log row: %v", err)
                }
                
                log := map[string]interface{}{
                        "id":         id,
                        "username":   username,
                        "login_time": loginTime,
                        "success":    success,
                        "ip_address": ipAddress,
                }
                logs = append(logs, log)
        }

        if err = rows.Err(); err != nil {
                return nil, fmt.Errorf("error iterating log rows: %v", err)
        }

        return logs, nil
}

// CloseDB closes the database connection
func CloseDB() {
        if db != nil {
                db.Close()
        }
}