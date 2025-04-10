#!/bin/bash

# Function to check if Docker is installed
check_docker() {
    if ! command -v docker &> /dev/null; then
        echo "Docker is not installed. Falling back to standalone mode."
        return 1
    fi
    return 0
}

# Function to check if Docker Compose is installed
check_docker_compose() {
    if ! command -v docker compose &> /dev/null; then
        echo "Docker Compose is not installed. Falling back to standalone mode."
        return 1
    fi
    return 0
}

# Function to check if Docker daemon is running
check_docker_running() {
    if ! docker info &> /dev/null; then
        echo "Docker daemon is not running. Falling back to standalone mode."
        return 1
    fi
    return 0
}

# Function to check if the application is already running
check_running() {
    if docker compose ps 2>/dev/null | grep -q "securesignin-app-1"; then
        echo "Application is already running. Use 'docker compose down' to stop it first."
        exit 1
    fi
}

# Function to display help
show_help() {
    echo "Usage: ./run.sh [command]"
    echo ""
    echo "Commands:"
    echo "  start          Start the application (with Docker if available, standalone otherwise)"
    echo "  start-docker   Force start with Docker (fails if Docker is unavailable)"
    echo "  start-local    Force start locally without Docker"
    echo "  stop           Stop the application"
    echo "  restart        Restart the application"
    echo "  logs           View application logs"
    echo "  migrate        Migrate from PostgreSQL to SQLite (if PostgreSQL is running)"
    echo "  db-check       Check SQLite database integrity"
    echo "  db-repair      Attempt to repair SQLite database"
    echo "  db-backup      Create a backup of the SQLite database"
    echo "  help           Show this help message"
    echo ""
    echo "Default command is 'start'"
}

# Function to migrate from PostgreSQL to SQLite
migrate_to_sqlite() {
    echo "Building migration tool..."
    if ! go build -tags migration -o migrate_to_sqlite migrate_postgres_to_sqlite.go 2>/dev/null; then
        echo "Failed to build migration tool. File may have been removed after successful migration."
        exit 1
    fi
    
    echo "Running migration from PostgreSQL to SQLite..."
    ./migrate_to_sqlite
    
    if [ $? -eq 0 ]; then
        echo "Migration completed successfully!"
        echo "You can now run the application with SQLite."
    else
        echo "Migration failed. Please check the error messages."
        exit 1
    fi
}

# Function to check database integrity
check_database() {
    # Ensure the cmd directory exists
    mkdir -p cmd/dbcheck
    
    # Check if the database check tool exists
    if [ ! -f cmd/dbcheck/main.go ]; then
        echo "Database check tool not found. Creating..."
        ./setup.sh
    fi
    
    echo "Checking database integrity..."
    go run cmd/dbcheck/main.go check "$@"
}

# Function to repair database
repair_database() {
    # Ensure the cmd directory exists
    mkdir -p cmd/dbcheck
    
    # Check if the database check tool exists
    if [ ! -f cmd/dbcheck/main.go ]; then
        echo "Database repair tool not found. Creating..."
        ./setup.sh
    fi
    
    echo "Attempting to repair database..."
    go run cmd/dbcheck/main.go repair "$@"
}

# Function to backup database
backup_database() {
    # Ensure the cmd directory exists
    mkdir -p cmd/dbbackup
    
    # Check if the database backup tool exists
    if [ ! -f cmd/dbbackup/main.go ]; then
        echo "Database backup tool not found. Creating..."
        ./setup.sh
    fi
    
    echo "Creating database backup..."
    go run cmd/dbbackup/main.go "$@"
}

# Function to start the application in standalone mode
start_standalone() {
    echo "Starting application in standalone mode..."
    
    # Check if run_standalone.sh exists
    if [ ! -f run_standalone.sh ]; then
        echo "Creating standalone script..."
        cat > run_standalone.sh << 'EOL'
#!/bin/bash

# Run standalone script for SecureSignIn - suitable for AppImage packaging
# This script sets up the environment for the application to run without Docker

# Set up environment variables
APP_DIR="$(dirname "$(readlink -f "$0")")"
export SQLITE_DB_PATH="$HOME/.securesignin/securesignin.db"
DATA_DIR="$(dirname "$SQLITE_DB_PATH")"

# Create data directory if it doesn't exist
mkdir -p "$DATA_DIR"

# Check if database file exists
if [ ! -f "$SQLITE_DB_PATH" ]; then
  echo "Creating new database at $SQLITE_DB_PATH"
fi

# Run the application
echo "Starting SecureSignIn from $APP_DIR"
echo "Using database: $SQLITE_DB_PATH"

# Kill existing instances if any
pkill -f securesignin || true

# Start the application
exec "$APP_DIR/securesignin"
EOL
        chmod +x run_standalone.sh
    fi
    
    # Build the application if it doesn't exist
    if [ ! -f securesignin ]; then
        echo "Building application..."
        go build -o securesignin .
    fi
    
    # Run the standalone script
    ./run_standalone.sh
}

# Function to start with Docker
start_docker() {
    check_docker || return 1
    check_docker_compose || return 1
    check_docker_running || return 1
    check_running
    
    echo "Starting application with Docker..."
    
    # Try to build and run with Docker
    echo "Building and starting Docker container..."
    if ! docker compose up -d --build; then
        echo "Failed to start Docker container."
        return 1
    fi
    
    echo "Waiting for application to become healthy..."
    # Wait up to 60 seconds for the app service to be healthy
    for i in {1..12}; do
        if docker compose ps 2>/dev/null | grep -q '\(healthy\)'; then
            echo "Application is healthy."
            return 0
        fi
        echo "Still waiting for app service... ($i/12)"
        sleep 5
    done
    
    echo "Application did not become healthy after 60 seconds."
    docker compose logs app
    return 1
}

# Main script
case "$1" in
    "stop")
        if check_docker && check_docker_compose && check_docker_running; then
            echo "Stopping Docker containers..."
            docker compose down
            echo "Docker containers stopped successfully."
        fi
        
        # Also kill any standalone instances
        echo "Stopping any standalone instances..."
        pkill -f securesignin || true
        echo "Application stopped successfully."
        ;;
    "restart")
        $0 stop
        sleep 2
        $0 start
        ;;
    "logs")
        if check_docker && check_docker_compose && check_docker_running; then
            echo "Showing Docker logs..."
            docker compose logs -f app
        else
            echo "Docker not available. Check application log files in $HOME/.securesignin/"
        fi
        ;;
    "migrate")
        migrate_to_sqlite
        ;;
    "db-check")
        shift
        check_database "$@"
        ;;
    "db-repair")
        shift
        repair_database "$@"
        ;;
    "db-backup")
        shift
        backup_database "$@"
        ;;
    "start-docker")
        # Force Docker mode
        if start_docker; then
            echo "Application started successfully with Docker."
            echo "Access the application at http://localhost:8080"
        else
            echo "Failed to start with Docker."
            exit 1
        fi
        ;;
    "start-local")
        # Force standalone mode
        start_standalone
        ;;
    "help"|"--help"|"-h")
        show_help
        ;;
    "start"|"")
        # Try Docker first, fall back to standalone
        if start_docker; then
            echo "Application started successfully with Docker."
            echo "Access the application at http://localhost:8080"
        else
            echo "Falling back to standalone mode..."
            start_standalone
        fi
        ;;
    *)
        echo "Unknown command: $1"
        show_help
        exit 1
        ;;
esac

# Check your current user and groups
groups

# Check Docker socket permissions
ls -l /var/run/docker.sock 