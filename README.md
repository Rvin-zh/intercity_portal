# SecureSignIn

A secure authentication system built with Go and Electron, featuring modern security practices and a user-friendly interface.

## Features

- Secure user authentication
- Password recovery system
- Modern web interface
- Desktop application support via Electron
- Docker support for easy deployment
- Database integration

## Project Structure

```
.
├── db/                 # Database related code
├── electron-app/       # Electron desktop application
├── static/            # Static assets (CSS, images)
├── templates/         # HTML templates
├── utils/             # Utility functions
└── src/               # Source code
```

## Getting Started

### Prerequisites

- Go 1.16 or higher
- Node.js and npm (for Electron app)
- Docker (optional)

### Installation

1. Clone the repository:

```bash
git clone https://github.com/yourusername/SecureSignIn.git
cd SecureSignIn
```

2. Install Go dependencies:

```bash
go mod download
```

3. Install Electron app dependencies:

```bash
cd electron-app
npm install
```

### Running the Application

#### Web Version

```bash
go run main.go
```

#### Desktop Version

```bash
cd electron-app
npm start
```

#### Docker

```bash
docker-compose up
```

## Development

### Branch Strategy

- `main` - Production-ready code
- `develop` - Development branch
- `feature/*` - New features
- `bugfix/*` - Bug fixes
- `release/*` - Release preparation

### Contributing

1. Create a new branch from `develop`
2. Make your changes
3. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
