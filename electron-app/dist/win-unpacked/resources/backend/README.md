# Secure Sign In Application

A secure web application built with Go and Docker, featuring user authentication, registration, and a dashboard.

## Prerequisites

Before running the application, ensure you have the following installed:

- Docker (version 20.10.0 or higher)
- Docker Compose (version 2.0.0 or higher)

## Installation Instructions

### Linux

1. Install Docker:

```bash
# For Ubuntu/Debian
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# For Fedora
sudo dnf install docker
sudo systemctl start docker
sudo systemctl enable docker
```

2. Install Docker Compose:

```bash
# For Ubuntu/Debian
sudo apt-get update
sudo apt-get install docker-compose-plugin

# For Fedora
sudo dnf install docker-compose
```

3. Add your user to the docker group (to avoid using sudo):

```bash
sudo usermod -aG docker $USER
# Log out and log back in for the changes to take effect
```

### macOS

1. Install Docker Desktop for Mac:

   - Download from [Docker's official website](https://www.docker.com/products/docker-desktop)
   - Install the downloaded package
   - Start Docker Desktop from your Applications folder

2. Docker Compose comes included with Docker Desktop for Mac

### Windows

1. Install Docker Desktop for Windows:

   - Download from [Docker's official website](https://www.docker.com/products/docker-desktop)
   - Run the installer
   - Follow the installation wizard
   - Restart your computer if prompted

2. Docker Compose comes included with Docker Desktop for Windows

## Running the Application

### Using the Shell Script (Recommended)

1. Make the run script executable:

```bash
chmod +x run.sh
```

2. Start the application:

```bash
./run.sh start
```

3. Other available commands:

```bash
./run.sh stop      # Stop the application
./run.sh restart   # Restart the application
./run.sh logs      # View application logs
./run.sh help      # Show help message
```

### Manual Docker Compose Commands

If you prefer to use Docker Compose directly:

1. Start the application:

```bash
docker compose up -d
```

2. Stop the application:

```bash
docker compose down
```

3. View logs:

```bash
docker compose logs -f app
```

## Accessing the Application

Once the application is running, you can access it at:

- http://localhost:8080

## Features

- User registration and authentication
- Secure password handling
- Dashboard with user management
- Login history tracking
- Responsive design
- Health check endpoint

## Development

### Project Structure

```
.
├── Dockerfile
├── docker-compose.yml
├── main.go
├── handlers.go
├── db/
│   └── db.go
├── templates/
│   ├── base.html
│   ├── dashboard.html
│   ├── index.html
│   ├── login.html
│   ├── register.html
│   └── forgot.html
└── static/
    ├── css/
    │   ├── reset.css
    │   └── style.css
    ├── js/
    │   └── script.js
    └── images/
```

### Building from Source

1. Install Go (version 1.21 or higher)
2. Clone the repository
3. Install dependencies:

```bash
go mod download
```

4. Build the application:

```bash
go build -o main .
```

## Troubleshooting

### Common Issues

1. **Port Already in Use**

   - Ensure no other application is using port 8080
   - You can change the port in docker-compose.yml if needed

2. **Docker Permission Issues**

   - Make sure your user is in the docker group
   - Try running `sudo usermod -aG docker $USER` and log out/in

3. **Database Connection Issues**
   - Check if the database container is running: `docker compose ps`
   - View database logs: `docker compose logs db`

### Logs

To view detailed logs:

```bash
./run.sh logs
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Desktop Application

A desktop version of this application is now available for both Linux and Windows. The desktop application offers the same functionality as the web version but runs as a standalone application.

### Features

- Runs as a native desktop application
- No need for a separate web browser
- Same security and functionality as the web version
- Available for Linux and Windows

### Installation

For installation instructions, see the [Desktop Application Installation Guide](electron-app/INSTALLATION.md).

### Building from Source

If you want to build the desktop application from source, see the [Packaging and Distribution Guide](electron-app/PACKAGING.md).
