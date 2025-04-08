# Intercity Portal

A management application for an intercity terminal, built with Go and featuring a desktop version using Electron. This project originated as a secure authentication system (SecureSignIn) and is currently under development.

**Current Status:** The application currently only features the basic login/authentication pages. Core management functionality is yet to be implemented.

## Features (Planned & Existing)

- Secure user authentication (Existing)
- Password recovery system (Existing)
- Modern web interface (Partially implemented)
- Desktop application support via Electron (Setup complete)
- Docker support for easy deployment (Existing)
- Database integration (Existing)
- Terminal Management Features (Planned)

## Project Structure

```
.
├── db/                 # Database related code
├── electron-app/       # Electron desktop application source and build info
├── static/            # Static assets (CSS, images)
├── templates/         # HTML templates
├── utils/             # Utility functions
└── src/               # Go backend source code
```

## Getting Started

### Prerequisites

- Go 1.16 or higher
- Node.js and npm (for Electron app)
- Docker (optional)

### Installation

1. Clone the repository:

```bash
git clone https://github.com/Rvin-zh/intercity_portal.git
cd intercity_portal
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

#### Web Version (Backend)

From the project root:

```bash
go run main.go
```

This will start the backend server. You can access the web interface (currently login page) via your browser, typically at `http://localhost:8080`.

#### Desktop Version (Electron)

For instructions on how to build and run the Electron desktop application, please refer to the specific README located in the `electron-app` directory:

[**Electron App README**](./electron-app/README.md)

#### Docker

From the project root:

```bash
docker-compose up --build
```

## Development

### Branch Strategy

- `main` - Relatively stable code, potentially deployable
- `develop` - Main development branch where features are merged
- `feature/*` - New features
- `bugfix/*` - Bug fixes
- `release/*` - Release preparation

### Contributing

1. Create a new branch from `develop`
2. Make your changes
3. Submit a pull request against the `develop` branch

## License

This project is licensed under the MIT License.
