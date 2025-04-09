const { app, BrowserWindow, dialog, shell } = require('electron');
const path = require('path');
const { spawn } = require('child_process');
const os = require('os');
const fs = require('fs');

// Keep a global reference of the window object to avoid garbage collection
let mainWindow;
let backendProcess = null;
let portCheckInterval = null;
const BACKEND_PORT = 8080;

// Get path to backend executable
function getBackendPath() {
  const isPackaged = app.isPackaged;
  let backendExecutablePath;

  if (isPackaged) {
    // In packaged app (AppImage, etc.), resourcesPath points to the 'resources' dir
    // The backend files were copied into a 'backend' subdirectory within 'resources'
    const backendDir = path.join(process.resourcesPath, 'backend');
    const exeName = os.platform() === 'win32' ? 'main.exe' : 'main';
    backendExecutablePath = path.join(backendDir, exeName);
    console.log(`Packaged app detected. Resources path: ${process.resourcesPath}`);
  } else {
    // In development, backend is in the project root (two levels up from src)
    const baseDir = path.join(__dirname, '..', '..');
    const exeName = os.platform() === 'win32' ? 'main.exe' : 'main';
    backendExecutablePath = path.join(baseDir, exeName);
    console.log(`Development mode detected. Project path: ${baseDir}`);
  }

  console.log(`Resolved backend executable path: ${backendExecutablePath}`);
  return backendExecutablePath;
}

// Check if backend is running
async function checkBackendReachable(url) {
  return new Promise((resolve) => {
    const req = require('http').get(url, (res) => {
      res.resume();
      resolve(res.statusCode === 200);
    }).on('error', () => {
      resolve(false);
    });
    req.setTimeout(1000, () => {
      req.abort();
      resolve(false);
    });
  });
}

// Start the backend server
async function startBackend() {
  const backendPath = getBackendPath();
  const backendDirectory = path.dirname(backendPath);
  const sqliteDbPath = path.join(app.getPath('userData'), 'securesignin.db');
  const keyDirectory = path.join(backendDirectory, 'keys');

  console.log(`Backend executable path: ${backendPath}`);
  console.log(`Backend working directory: ${backendDirectory}`);
  console.log(`SQLite database path: ${sqliteDbPath}`);
  console.log(`Key directory path: ${keyDirectory}`);

  // Ensure key directory exists
  if (!fs.existsSync(keyDirectory)) {
    console.log('Creating key directory...');
    try {
      fs.mkdirSync(keyDirectory, { recursive: true });
      console.log('Key directory created successfully.');
    } catch (error) {
      console.error(`Error creating key directory: ${error.message}`);
    }
  }

  if (!fs.existsSync(backendPath)) {
    const errorMsg = `Backend executable not found at: ${backendPath}`;
    console.error(errorMsg);
    dialog.showErrorBox('Backend Error', errorMsg + 
      `\n\nWindows troubleshooting:\n1. Check if antivirus is blocking the executable\n2. Ensure you have proper permissions\n3. Try running as administrator`);
    app.quit();
    return false; // Indicate failure
  }

  // Check backend executable permissions
  try {
    fs.accessSync(backendPath, fs.constants.X_OK);
    console.log('Backend executable has execute permissions.');
  } catch (error) {
    console.warn(`Backend executable doesn't have execute permissions. Attempting to fix...`);
    if (os.platform() !== 'win32') { // Windows doesn't use the same permission model
      try {
        fs.chmodSync(backendPath, 0o755); // Set execute permissions
        console.log('Execute permissions set on backend executable.');
      } catch (chmodErr) {
        console.error(`Failed to set execute permissions: ${chmodErr.message}`);
      }
    }
  }

  // Check if backend is already running (less likely now, but good practice)
  if (await checkBackendReachable(`http://localhost:${BACKEND_PORT}/health`)) {
    console.log('Backend appears to be running already.');
    return true; // Indicate success (already running)
  }

  try {
    // For Windows, log the exact command we're about to execute
    if (os.platform() === 'win32') {
      console.log(`Starting Windows backend with: "${backendPath}" in directory: ${backendDirectory}`);
    }
    
    backendProcess = spawn(backendPath, [], {
      cwd: backendDirectory,
      detached: os.platform() !== 'win32',
      shell: os.platform() === 'win32',
      env: {
        ...process.env, // Inherit environment
        USE_SQLITE: '1', // Tell backend to use SQLite
        SQLITE_PATH: sqliteDbPath, // Provide path for the db file
        KEY_DIR: keyDirectory // Specify key directory path
      }
    });

    backendProcess.stdout.on('data', (data) => {
      console.log(`Backend stdout: ${data.toString().trim()}`);
    });
    backendProcess.stderr.on('data', (data) => {
      console.error(`Backend stderr: ${data.toString().trim()}`);
    });
    backendProcess.on('error', (err) => {
      const errorMsg = `Failed to start backend process: ${err}`;
      console.error(errorMsg);
      let detailedMsg = errorMsg;
      
      if (os.platform() === 'win32') {
        detailedMsg += `\n\nWindows troubleshooting:\n- Make sure no antivirus is blocking the app\n- Try running as administrator\n- Check Windows Defender settings\n- Verify the backend executable exists at: ${backendPath}`;
      }
      
      dialog.showErrorBox('Backend Error', detailedMsg);
      if (!app.isQuitting) app.quit();
    });
    backendProcess.on('close', (code) => {
      console.log(`Backend process exited with code ${code}`);
      backendProcess = null;
      if (code !== 0 && !app.isQuitting) {
        let errorMsg = `The backend server stopped unexpectedly (code: ${code}).`;
        
        if (os.platform() === 'win32') {
          errorMsg += `\n\nWindows troubleshooting:\n- Check if Windows Firewall is blocking network access\n- Verify SQLite database path is accessible: ${sqliteDbPath}\n- Run the app as administrator`;
        }
        
        dialog.showErrorBox('Backend Error', errorMsg);
        app.quit();
      }
    });

    console.log('Backend process started, waiting for it to become reachable...');
    return true; // Indicate success (started)

  } catch (error) {
    console.error('Error spawning backend process:', error);
    let errorMsg = `Error starting the backend server: ${error.message}`;
    
    if (os.platform() === 'win32') {
      errorMsg += `\n\nWindows troubleshooting:\n- Verify all DLLs are present in resources/backend\n- Check Windows permissions\n- Try running as administrator`;
    }
    
    dialog.showErrorBox('Backend Error', errorMsg);
    if (!app.isQuitting) app.quit();
    return false; // Indicate failure
  }
}

// Create the browser window
function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true
    },
    // Optional: Define icon path relative to build output
    // icon: path.join(__dirname, '..\/assets\/icon.png') // Adjust if you add an icon
  });

  mainWindow.loadURL(`http://localhost:${BACKEND_PORT}/login`);

  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    if (url.startsWith('http')) {
      shell.openExternal(url);
      return { action: 'deny' };
    }
    return { action: 'allow' };
  });

  mainWindow.on('closed', () => {
    mainWindow = null;
  });
}

// Wait for backend to be ready
async function waitForBackend() {
  console.log('Checking if backend is reachable...');
  const healthUrl = `http://localhost:${BACKEND_PORT}/health`;
  let backendReady = false;
  let attempts = 0;
  const maxAttempts = 30; // Wait up to 30 seconds

  while (!backendReady && attempts < maxAttempts) {
    attempts++;
    backendReady = await checkBackendReachable(healthUrl);
    if (backendReady) {
      console.log(`Backend is reachable after ${attempts} attempt(s).`);
      return true;
    } else {
      console.log(`Waiting for backend... Attempt ${attempts}/${maxAttempts}`);
      await new Promise(resolve => setTimeout(resolve, 1000));
    }
  }

  console.error('Backend did not become reachable.');
  dialog.showErrorBox('Backend Error', 'The backend service failed to start in time.');
  return false;
}

// Electron App Lifecycle
app.whenReady().then(async () => {
  const backendStarted = await startBackend();
  if (!backendStarted) {
    if (!app.isQuitting) app.quit();
    return;
  }

  const ready = await waitForBackend();
  if (ready) {
    createWindow();
  } else {
    if (!app.isQuitting) app.quit();
  }

  app.on('activate', () => {
    if (mainWindow === null) {
      // Ensure backend is ready before creating window on activate
      waitForBackend().then(ready => {
        if (ready) createWindow();
      });
    }
  });
});

app.on('before-quit', () => {
  app.isQuitting = true;
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

app.on('will-quit', () => {
  console.log('App is quitting, cleaning up backend...');
  if (backendProcess) {
    console.log('Terminating backend process');
    try {
      if (os.platform() === 'win32') {
        spawn('taskkill', ['/pid', backendProcess.pid, '/f', '/t']);
      } else {
        // Kill the process group to ensure child processes are terminated
        process.kill(-backendProcess.pid, 'SIGTERM');
      }
    } catch (e) {
      console.error('Failed to kill backend process:', e);
    }
    backendProcess = null;
  }
}); 