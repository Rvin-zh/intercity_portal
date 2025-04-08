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
        docker compose up -d --build
        echo "Application restarted successfully."
        ;;
    "logs")
        check_docker_compose
        echo "Showing application logs..."
        docker compose logs -f app
        ;;
    "help"|"--help"|"-h")
        show_help
        ;;
    "start"|"")
        check_docker
        check_docker_compose
        check_running
        echo "Starting application..."
        docker compose up -d --build
        echo "Application started successfully."
        echo "Access the application at http://localhost:8080"
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