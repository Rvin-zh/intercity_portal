#!/bin/bash

# Function to check if Docker is installed
check_docker() {
    if ! command -v docker &> /dev/null; then
        echo "Docker is not installed. Please install Docker first."
        exit 1
    fi
}

# Function to check if Docker Compose is installed
check_docker_compose() {
    if ! command -v docker compose &> /dev/null; then
        echo "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
}

# Function to check if the application is already running
check_running() {
    if docker compose ps | grep -q "securesignin-app-1"; then
        echo "Application is already running. Use 'docker compose down' to stop it first."
        exit 1
    fi
}

# Function to display help
show_help() {
    echo "Usage: ./run.sh [command]"
    echo ""
    echo "Commands:"
    echo "  start     Start the application"
    echo "  stop      Stop the application"
    echo "  restart   Restart the application"
    echo "  logs      View application logs"
    echo "  setup     Set up shared database environment"
    echo "  help      Show this help message"
    echo ""
    echo "Default command is 'start'"
}

# Main script
case "$1" in
    "stop")
        check_docker_compose
        echo "Stopping application..."
        docker compose down
        echo "Application stopped successfully."
        ;;
    "restart")
        check_docker_compose
        echo "Restarting application..."
        docker compose down
        
        # Set up shared database environment
        if [ -f "./setup-shared-db.sh" ]; then
            echo "Setting up shared database environment..."
            ./setup-shared-db.sh
        fi
        
        docker compose up -d --build
        echo "Application restarted successfully."
        ;;
    "logs")
        check_docker_compose
        echo "Showing application logs..."
        docker compose logs -f app
        ;;
    "setup")
        if [ -f "./setup-shared-db.sh" ]; then
            echo "Setting up shared database environment..."
            ./setup-shared-db.sh
        else
            echo "Error: setup-shared-db.sh script not found."
            exit 1
        fi
        ;;
    "help"|"--help"|"-h")
        show_help
        ;;
    "start"|"")
        check_docker
        check_docker_compose
        check_running
        
        # Set up shared database environment
        if [ -f "./setup-shared-db.sh" ]; then
            echo "Setting up shared database environment..."
            ./setup-shared-db.sh
        fi
        
        echo "Starting application..."
        docker compose up -d --build
        echo "Waiting for application to become healthy..."
        # Wait up to 60 seconds for the app service to be healthy
        for i in {1..12}; do
            if docker compose ps app | grep -q '\(healthy\)'; then
                echo "Application is healthy."
                break
            fi
            echo "Still waiting for app service... ($i/12)"
            sleep 5
        done

        if ! docker compose ps app | grep -q '\(healthy\)'; then
            echo "Application did not become healthy after 60 seconds."
            docker compose logs app
            exit 1
        fi

        echo "Testing /health endpoint from within the container..."
        # Use docker compose exec to run the check inside the app container
        if docker compose exec app curl -f http://localhost:8080/health; then
            echo "Endpoint /health test successful (from container)."
        else
            echo "Endpoint /health test failed (from container)."
            docker compose logs app
            exit 1
        fi

        echo "Application started successfully."
        echo "Access the application at http://localhost:8080"
        ;;
    *)
        echo "Unknown command: $1"
        show_help
        exit 1
        ;;
esac 