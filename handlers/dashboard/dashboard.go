package dashboard

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"SecureSignIn/db"
	"SecureSignIn/handlers"
	"SecureSignIn/handlers/templates"
	"SecureSignIn/models"
)

// Handler - For logged in users
func DashboardHandler(c echo.Context) error {
	// Make sure user is logged in first by checking for username cookie
	usernameCookie, err := c.Cookie("username")
	if err != nil || usernameCookie.Value == "" {
		log.Printf("Dashboard access attempted without valid session")
		return c.Redirect(http.StatusSeeOther, "/login?error=You must be logged in to access this page")
	}

	username := usernameCookie.Value
	log.Printf("Dashboard accessed by user: %s", username)

	// Get the user role
	roleCookie, err := c.Cookie("user_role")
	var userRole string
	if err != nil || roleCookie.Value == "" {
		// If role cookie not found, default to Operator
		userRole = "Operator"
		log.Printf("Warning: No role cookie found for user %s, defaulting to Operator", username)
	} else {
		userRole = roleCookie.Value
	}

	// Fetch all users
	allUsers, err := db.GetAllUsers()
	if err != nil {
		log.Printf("Error getting all users: %v", err)
		return templates.RenderTemplate(c, "dashboard.html", models.PageData{
			Title:      "Dashboard",
			Error:      "Error retrieving user data",
			ActivePage: "dashboard",
			IsLoggedIn: true,
			Username:   username,
			UserRole:   userRole,
		})
	}
	userMaps, err := handlers.RowsToMap(allUsers)
	if err != nil {
		log.Printf("Error converting user rows to map: %v", err)
	}

	// Fetch login history
	loginHistory, err := db.GetLoginHistory()
	if err != nil {
		log.Printf("Error getting login history: %v", err)
		return templates.RenderTemplate(c, "dashboard.html", models.PageData{
			Title:      "Dashboard",
			Error:      "Error retrieving login history",
			ActivePage: "dashboard",
			IsLoggedIn: true,
			Username:   username,
			UserRole:   userRole,
		})
	}
	historyMaps, err := handlers.RowsToMap(loginHistory)
	if err != nil {
		log.Printf("Error converting login history rows to map: %v", err)
	}

	// Get the user ID for security question check
	var userID int64
	for _, user := range userMaps {
		if user["username"] == username {
			userID, _ = user["id"].(int64)
			break
		}
	}

	// Check if user has security question
	hasSecurityQ := false
	if userID > 0 {
		hasSecurityQ, err = db.HasSecurityQuestion(userID)
		if err != nil {
			log.Printf("Error checking security question: %v", err)
		}
	}

	// Prepare common data
	data := models.PageData{
		Title:        "Dashboard",
		ActivePage:   "dashboard",
		Users:        userMaps,
		LoginLogs:    historyMaps,
		IsLoggedIn:   true,
		Username:     username,
		UserRole:     userRole,
		HasSecurityQ: hasSecurityQ,
		Success:      c.QueryParam("success"),
		Error:        c.QueryParam("error"),
	}

	// Render appropriate dashboard based on role
	switch userRole {
	case "Operator":
		return templates.RenderTemplate(c, "operator_dashboard.html", data)
	case "Manager":
		return templates.RenderTemplate(c, "manager_dashboard.html", data)
	case "Accountant":
		return templates.RenderTemplate(c, "accountant_dashboard.html", data)
	case "Admin":
		return templates.RenderTemplate(c, "admin_dashboard.html", data)
	default:
		// Default to operator dashboard if role is unknown
		log.Printf("Warning: Unknown role '%s' for user %s, defaulting to Operator dashboard", userRole, username)
		return templates.RenderTemplate(c, "operator_dashboard.html", data)
	}
}

// IndexHandler - Redirects appropriately
func IndexHandler(c echo.Context) error {
	data := models.PageData{
		Title:      "Home",
		ActivePage: "home",
	}
	return templates.RenderTemplate(c, "index.html", data)
}

// HealthCheckHandler - Health check endpoint
func HealthCheckHandler(c echo.Context) error {
	if err := db.DB.Ping(); err != nil {
		log.Printf("Health check failed: DB ping error: %v", err)
		return c.String(http.StatusServiceUnavailable, "Database connection failed")
	}
	return c.String(http.StatusOK, "OK")
} 