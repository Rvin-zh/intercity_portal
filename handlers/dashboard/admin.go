package dashboard

import (
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"SecureSignIn/db"
	"SecureSignIn/handlers/templates"
	"SecureSignIn/models"
)

// AdminDashboardHandler - Handler for admin dashboard
func AdminDashboardHandler(c echo.Context) error {
	// Make sure user is logged in and has admin role
	usernameCookie, err := c.Cookie("username")
	if err != nil || usernameCookie.Value == "" {
		log.Printf("Admin dashboard access attempted without valid session")
		return c.Redirect(http.StatusSeeOther, "/login?error=You must be logged in to access this page")
	}

	username := usernameCookie.Value
	
	// Check role
	roleCookie, err := c.Cookie("user_role")
	if err != nil || roleCookie.Value != "Admin" {
		log.Printf("Admin dashboard access attempted by non-admin user: %s", username)
		return c.Redirect(http.StatusSeeOther, "/dashboard?error=You do not have permission to access the admin dashboard")
	}

	log.Printf("Admin dashboard accessed by user: %s", username)

	// Prepare page data
	data := models.PageData{
		Title:        "Admin Dashboard",
		ActivePage:   "admin",
		IsLoggedIn:   true,
		Username:     username,
		UserRole:     "Admin",
		Success:      c.QueryParam("success"),
		Error:        c.QueryParam("error"),
	}

	return templates.RenderTemplate(c, "admin_dashboard.html", data)
}

// AdminUsersHandler - Handler for admin user management
func AdminUsersHandler(c echo.Context) error {
	// Make sure user is logged in and has admin role
	usernameCookie, err := c.Cookie("username")
	if err != nil || usernameCookie.Value == "" {
		log.Printf("Admin users page access attempted without valid session")
		return c.Redirect(http.StatusSeeOther, "/login?error=You must be logged in to access this page")
	}

	username := usernameCookie.Value
	
	// Check role
	roleCookie, err := c.Cookie("user_role")
	if err != nil || roleCookie.Value != "Admin" {
		log.Printf("Admin users page access attempted by non-admin user: %s", username)
		return c.Redirect(http.StatusSeeOther, "/dashboard?error=You do not have permission to access the admin users page")
	}

	// Fetch all users
	allUsers, err := db.GetAllUsers()
	if err != nil {
		log.Printf("Error getting all users: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve users",
		})
	}
	defer allUsers.Close()

	// Convert rows to array of maps
	var users []map[string]interface{}
	for allUsers.Next() {
		var id int64
		var username, email, passwordHash, createdAt, dateOfBirth, socialSecurity, role string
		
		if err := allUsers.Scan(&id, &username, &email, &passwordHash, &createdAt, &dateOfBirth, &socialSecurity, &role); err != nil {
			log.Printf("Error scanning user row: %v", err)
			continue
		}
		
		// Don't include password hash in the response
		users = append(users, map[string]interface{}{
			"id":         id,
			"username":   username,
			"email":      email,
			"created_at": createdAt,
			"role":       role,
		})
	}

	return c.JSON(http.StatusOK, users)
}

// AdminCreateUserHandler - Handler for creating new users
func AdminCreateUserHandler(c echo.Context) error {
	// Make sure user is logged in and has admin role
	usernameCookie, err := c.Cookie("username")
	if err != nil || usernameCookie.Value == "" {
		log.Printf("Admin create user attempted without valid session")
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "You must be logged in to perform this action",
		})
	}
	
	// Check role
	roleCookie, err := c.Cookie("user_role")
	if err != nil || roleCookie.Value != "Admin" {
		log.Printf("Admin create user attempted by non-admin user: %s", usernameCookie.Value)
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "You do not have permission to perform this action",
		})
	}

	// Parse form data
	username := c.FormValue("username")
	email := c.FormValue("email")
	password := c.FormValue("password")
	role := c.FormValue("role")

	// Basic validation
	if username == "" || email == "" || password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Username, email, and password are required",
		})
	}

	// Check valid role
	if role != "Operator" && role != "Manager" && role != "Accountant" && role != "Admin" {
		role = "Operator" // Default to Operator if invalid role
	}

	// Create user
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create user",
		})
	}
	
	// Add user to database
	userID, err := db.AddUser(username, string(hashedPassword), "", "", email, role)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create user",
		})
	}
	
	log.Printf("User created successfully by admin. ID: %d, Username: %s, Role: %s", userID, username, role)
	
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "User created successfully",
		"user_id": userID,
	})
}

// AdminUpdateUserHandler - Handler for updating a user's role
func AdminUpdateUserHandler(c echo.Context) error {
	// Ensure request body contains id and role
	var req struct {
		ID   int64  `json:"id"`
		Role string `json:"role"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	// Validate role
	switch req.Role {
	case "Operator", "Manager", "Accountant", "Admin":
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid role"})
	}
	// Update role in database
	if err := db.UpdateUserRole(req.ID, req.Role); err != nil {
		log.Printf("Error updating role for user %d: %v", req.ID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user role"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Role updated successfully"})
}

// AdminDeleteUserHandler - Handler for deleting a user
func AdminDeleteUserHandler(c echo.Context) error {
	idParam := c.Param("id")
	userID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}
	if err := db.DeleteUser(userID); err != nil {
		log.Printf("Error deleting user %d: %v", userID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "User deleted successfully"})
} 