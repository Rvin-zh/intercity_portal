# Secure Sign In Desktop Application

This is a desktop version of the Secure Sign In web application, packaged using Electron.

## Prerequisites

Before building or running the application, ensure you have the following installed:

- Node.js (v14 or higher)
- npm (v6 or higher)
- Go (v1.16 or higher) - required for building the backend

For detailed installation instructions for these prerequisites, see [PREREQUISITES.md](PREREQUISITES.md).

## Development

### Setup

1. Install dependencies:

```bash
npm install
```

2. Build the Go backend for both Linux and Windows:

```bash
# First make the build script executable
chmod +x build-backend.sh

# Then run it
./build-backend.sh
```

3. Start the application in development mode:

```bash
npm start
```

### Building

To build the application for different platforms:

#### For Linux:

```bash
npm run package-linux
```

This will create both AppImage and .deb packages in the `dist` directory.

#### For Windows:

```bash
npm run package-win
```

This will create both a portable executable and an NSIS installer in the `dist` directory.

#### For both platforms:

```bash
npm run package-all
```

## Application Structure

- `src/main.js`: Main Electron process that starts the application
- `backend/`: Contains the Go backend application (included from parent directory)

## How It Works

The application embeds the Go backend executable and runs it as a child process when the Electron app starts. The frontend is then loaded in an Electron BrowserWindow, connecting to the backend running on localhost:8080.

## Troubleshooting

- **Backend fails to start**: Check the application logs. You can enable extra logging by setting the environment variable `DEBUG=1`.
- **Window shows blank screen**: The backend may not have started successfully. Check the application logs.
