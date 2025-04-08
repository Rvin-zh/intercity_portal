package main

import (
	"SecureSignIn/db"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// Middleware to recover from panics and log errors
func logAndRecover(handler echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic: %v\nStack trace:\n%s", err, string(debug.Stack()))
				httpError := echo.NewHTTPError(http.StatusInternalServerError, "An unexpected error occurred. Please try again later.")
				c.Error(httpError)
			}
		}()

		log.Printf("Request: %s %s from %s", c.Request().Method, c.Request().URL.Path, c.RealIP())
		err := handler(c)
		if err != nil {
			log.Printf("Handler error for %s %s: %v", c.Request().Method, c.Request().URL.Path, err)
			return err
		}
		log.Printf("Response %d sent for: %s %s", c.Response().Status, c.Request().Method, c.Request().URL.Path)
		return nil
	}
}

// Template cache
var templates = make(map[string]*template.Template)

// Load templates on init
func init() {
	templatesDir := "templates"
	log.Printf("Loading templates from: %s", templatesDir)

	layouts, err := filepath.Glob(filepath.Join(templatesDir, "base.html"))
	if err != nil || len(layouts) == 0 {
		log.Fatalf("Error loading base template: %v (found: %d)", err, len(layouts))
	}
	log.Printf("Found base template: %v", layouts)

	includes, err := filepath.Glob(filepath.Join(templatesDir, "*.html"))
	if err != nil {
		log.Fatalf("Error finding template includes: %v", err)
	}
	log.Printf("Found templates: %v", includes)

	funcMap := template.FuncMap{
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
	}

	for _, include := range includes {
		if filepath.Base(include) == "base.html" {
			continue
		}

		files := append([]string{include}, layouts...)
		fileName := filepath.Base(include)
		log.Printf("Loading template: %s with files: %v", fileName, files)
		templates[fileName] = template.Must(template.New(fileName).Funcs(funcMap).ParseFiles(files...))
	}

	log.Printf("Templates loaded successfully. Count: %d", len(templates))
	if _, ok := templates["login.html"]; !ok {
		log.Fatalf("FATAL: login.html template not loaded correctly.")
	}
}

// Render a template given a model
func renderTemplate(c echo.Context, tmpl string, data interface{}) error {
	log.Printf("Attempting to render template: %s", tmpl)

	t, ok := templates[tmpl]
	if !ok {
		log.Printf("Template %s does not exist in map. Available templates: %v", tmpl, templates)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Template %s not found.", tmpl))
	}

	var buf strings.Builder
	err := t.ExecuteTemplate(&buf, "base", data)
	if err != nil {
		log.Printf("Error executing template %s: %v", tmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error rendering page.")
	}

	return c.HTML(http.StatusOK, buf.String())
}

// Page data struct
type PageData struct {
	Title      string
	Error      string
	Success    string
	ActivePage string
	Users      []map[string]interface{}
	LoginLogs  []map[string]interface{}
	IsLoggedIn bool
	Username   string
	ResetToken string
}

// --- Password Reset Token Store (In-Memory - Demo Only) ---
type ResetTokenInfo struct {
	UserID int
	Expiry time.Time
}

var (
	resetTokens = make(map[string]ResetTokenInfo)
	tokenMutex  sync.RWMutex
)

const resetTokenValidity = 15 * time.Minute // Token valid for 15 minutes

// generateResetToken creates a secure random token.
func generateResetToken() (string, error) {
	b := make([]byte, 16) // 128 bits
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// storeResetToken stores a token for a user.
func storeResetToken(userID int) (string, error) {
	token, err := generateResetToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate reset token: %w", err)
	}
	tokenMutex.Lock()
	defer tokenMutex.Unlock()
	resetTokens[token] = ResetTokenInfo{
		UserID: userID,
		Expiry: time.Now().Add(resetTokenValidity),
	}
	log.Printf("Stored reset token for user ID %d (expires %v)", userID, resetTokens[token].Expiry)
	return token, nil
}

// validateResetToken checks if a token is valid and returns the user ID.
func validateResetToken(token string) (int, bool) {
	tokenMutex.RLock()
	defer tokenMutex.RUnlock()
	info, exists := resetTokens[token]
	valid := exists && !time.Now().After(info.Expiry)

	if exists && !valid { // Token expired, remove it
		tokenMutex.Lock()
		delete(resetTokens, token)
		tokenMutex.Unlock()
		log.Printf("Reset token %s expired and removed.", token)
	}

	if valid {
		log.Printf("Validated reset token %s for user ID %d.", token, info.UserID)
	} else {
		log.Printf("Reset token %s invalid or not found.", token)
	}
	return info.UserID, valid
}

// invalidateResetToken removes a token from the store.
func invalidateResetToken(token string) {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()
	delete(resetTokens, token)
	log.Printf("Invalidated reset token %s.", token)
}

// --- End Token Store ---

// --- Handlers ---

// Index handler - Redirects appropriately
func indexHandler(c echo.Context) error {
	data := PageData{
		Title:      "Home",
		ActivePage: "home",
	}
	return renderTemplate(c, "index.html", data)
}

// Helper to convert *sql.Rows to []map[string]interface{}
func rowsToMap(rows *sql.Rows) ([]map[string]interface{}, error) {
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		rowMap := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			rowMap[colName] = *val
		}
		results = append(results, rowMap)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}
	return results, nil
}

// Dashboard handler - For logged in users (Simplified, assumes auth middleware sets user)
func dashboardHandler(c echo.Context) error {
	userRows, err := db.GetAllUsers()
	if err != nil {
		log.Printf("Error getting users: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve user data.")
	}
	usersData, err := rowsToMap(userRows)
	if err != nil {
		log.Printf("Error mapping users: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process user data.")
	}

	logRows, err := db.GetLoginHistory()
	if err != nil {
		log.Printf("Error getting login history: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve login history.")
	}
	loginData, err := rowsToMap(logRows)
	if err != nil {
		log.Printf("Error mapping login history: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process login history.")
	}

	data := PageData{
		Title:      "Dashboard",
		ActivePage: "dashboard",
		IsLoggedIn: true,
		Users:      usersData,
		LoginLogs:  loginData,
	}
	return renderTemplate(c, "dashboard.html", data)
}

// Login handler - Render login page
func loginHandler(c echo.Context) error {
	data := PageData{
		Title:      "Login",
		ActivePage: "login",
		Success:    c.QueryParam("success"),
		Error:      c.QueryParam("error"),
	}
	return renderTemplate(c, "login.html", data)
}

// Auth handler - Process login form
func basicAuthHandler(c echo.Context) error {
	username := strings.TrimSpace(c.FormValue("username"))
	password := strings.TrimSpace(c.FormValue("password"))
	ipAddress := c.RealIP()

	userRow, err := db.GetUserByUsername(username)
	if err != nil {
		log.Printf("Error querying user %s: %v", username, err)
		db.LogLoginAttempt(0, ipAddress, false)
		return c.Redirect(http.StatusSeeOther, "/login?error=Invalid credentials")
	}

	var userID int
	var storedUsername string
	var passwordHash string

	err = userRow.Scan(&userID, &storedUsername, &passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Login attempt failed: User %s not found", username)
			db.LogLoginAttempt(0, ipAddress, false)
			return c.Redirect(http.StatusSeeOther, "/login?error=Invalid credentials")
		}
		log.Printf("Error scanning user row for %s: %v", username, err)
		db.LogLoginAttempt(0, ipAddress, false)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error processing login.")
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		log.Printf("Login attempt failed: Invalid password for user %s (ID: %d)", username, userID)
		db.LogLoginAttempt(userID, ipAddress, false)
		return c.Redirect(http.StatusSeeOther, "/login?error=Invalid credentials")
	}

	log.Printf("User %s (ID: %d) logged in successfully from %s", username, userID, ipAddress)
	db.LogLoginAttempt(userID, ipAddress, true)

	return c.Redirect(http.StatusSeeOther, "/dashboard?success=Successfully logged in&user="+username)
}

// Register handler - Render registration page
func registerHandler(c echo.Context) error {
	data := PageData{
		Title:      "Register",
		ActivePage: "register",
	}
	return renderTemplate(c, "register.html", data)
}

// Basic Register Handler - Process registration form
func basicRegisterHandler(c echo.Context) error {
	username := strings.TrimSpace(c.FormValue("username"))
	password := c.FormValue("password")
	ipAddress := c.RealIP()

	if username == "" || password == "" {
		return renderTemplate(c, "register.html", PageData{Title: "Register", Error: "Username and password cannot be empty.", ActivePage: "register"})
	}
	if len(password) < 8 {
		return renderTemplate(c, "register.html", PageData{Title: "Register", Error: "Password must be at least 8 characters long.", ActivePage: "register"})
	}

	// Check if user exists
	row, err := db.GetUserByUsername(username)
	// Need to scan the row to trigger the actual error (sql.ErrNoRows)
	var userID int
	var storedUsername string
	var passwordHash string
	scanErr := row.Scan(&userID, &storedUsername, &passwordHash)

	if scanErr == nil {
		// User exists
		return renderTemplate(c, "register.html", PageData{Title: "Register", Error: "Username already taken.", ActivePage: "register"})
	}
	if scanErr != sql.ErrNoRows {
		log.Printf("Error checking username %s existence: %v", username, scanErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error checking username.")
	}

	// User doesn't exist, continue with registration
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error processing registration.")
	}

	newUserID, err := db.AddUser(username, string(hashedPassword))
	if err != nil {
		log.Printf("Error creating user %s: %v", username, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user.")
	}

	log.Printf("User %s (ID: %d) registered successfully from %s", username, newUserID, ipAddress)
	return c.Redirect(http.StatusSeeOther, "/login?success=Registration successful! Please log in.")
}

// Forgot Password handler - Render/Process forgot password form
func forgotHandler(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		emailOrUsername := c.FormValue("email")

		userRow, err := db.GetUserByUsername(emailOrUsername)
		if err != nil {
			log.Printf("Password reset request failed: Error finding user %s: %v", emailOrUsername, err)
			return renderTemplate(c, "forgot.html", PageData{Title: "Forgot Password", Success: "If an account exists for that username, a password reset link has been simulated.", ActivePage: "forgot"})
		}

		var userID int
		var storedUsername string
		var passwordHash string
		err = userRow.Scan(&userID, &storedUsername, &passwordHash)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("Password reset request: User %s not found", emailOrUsername)
				return renderTemplate(c, "forgot.html", PageData{Title: "Forgot Password", Success: "If an account exists for that username, a password reset link has been simulated.", ActivePage: "forgot"})
			}
			log.Printf("Password reset request failed: Error scanning user %s: %v", emailOrUsername, err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error processing request.")
		}

		token, err := storeResetToken(userID)
		if err != nil {
			log.Printf("Error storing reset token for user ID %d: %v", userID, err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error processing request.")
		}

		resetLink := fmt.Sprintf("%s/reset/%s", c.Request().Host, token)
		log.Printf("Password reset requested for user %s (ID: %d). Simulated email sent. Reset link: http://%s", storedUsername, userID, resetLink)

		return renderTemplate(c, "forgot.html", PageData{Title: "Forgot Password", Success: "Password reset simulation complete. Check logs for the reset link.", ActivePage: "forgot"})
	}

	data := PageData{
		Title:      "Forgot Password",
		ActivePage: "forgot",
	}
	return renderTemplate(c, "forgot.html", data)
}

// Show Reset Password Form handler
func showResetFormHandler(c echo.Context) error {
	token := c.Param("token")
	userID, valid := validateResetToken(token)

	if !valid {
		log.Printf("Invalid or expired reset token presented: %s", token)
		return c.Redirect(http.StatusSeeOther, "/forgot?error=Invalid or expired reset link.")
	}

	log.Printf("Showing password reset form for valid token %s (User ID: %d)", token, userID)
	data := PageData{
		Title:      "Reset Password",
		ActivePage: "reset",
		ResetToken: token,
	}
	return renderTemplate(c, "reset_password.html", data)
}

// Handle Reset Password Submission handler
func handleResetPasswordHandler(c echo.Context) error {
	token := c.Param("token")
	newPassword := c.FormValue("password")
	confirmPassword := c.FormValue("confirm_password")

	userID, valid := validateResetToken(token)
	if !valid {
		log.Printf("Password reset attempt with invalid/expired token: %s", token)
		return c.Redirect(http.StatusSeeOther, "/forgot?error=Invalid or expired reset link.")
	}

	if newPassword == "" || newPassword != confirmPassword {
		log.Printf("Password reset failed for token %s: Passwords do not match or are empty.", token)
		data := PageData{
			Title:      "Reset Password",
			Error:      "Passwords do not match or are empty.",
			ResetToken: token,
		}
		return renderTemplate(c, "reset_password.html", data)
	}
	if len(newPassword) < 8 {
		data := PageData{
			Title:      "Reset Password",
			Error:      "Password must be at least 8 characters long.",
			ResetToken: token,
		}
		return renderTemplate(c, "reset_password.html", data)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing new password for user ID %d: %v", userID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error processing password reset.")
	}

	err = db.UpdateUserPassword(userID, string(hashedPassword))
	if err != nil {
		log.Printf("Error updating password in DB for user ID %d: %v", userID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update password.")
	}

	invalidateResetToken(token)

	log.Printf("Password successfully reset for user ID %d using token %s", userID, token)
	return c.Redirect(http.StatusSeeOther, "/login?success=Password successfully reset. Please log in.")
}

// Health Check handler
func healthCheckHandler(c echo.Context) error {
	if err := db.DB.Ping(); err != nil {
		log.Printf("Health check failed: DB ping error: %v", err)
		return c.String(http.StatusServiceUnavailable, "Database connection failed")
	}
	return c.String(http.StatusOK, "OK")
}

// Logout handler (simple redirect for demo)
func logoutHandler(c echo.Context) error {
	log.Printf("User logged out.")
	return c.Redirect(http.StatusSeeOther, "/login?success=Successfully logged out.")
}

// Simple validation helpers (replace with a proper validation library if needed)
func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

// --- End Handlers ---
