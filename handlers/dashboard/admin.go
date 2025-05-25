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

	// Parse JSON request
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	// Basic validation
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Username, email, and password are required",
		})
	}

	// Check valid role
	if req.Role != "Operator" && req.Role != "Manager" && req.Role != "Accountant" && req.Role != "Admin" {
		req.Role = "Operator" // Default to Operator if invalid role
	}

	// Create user
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create user",
		})
	}
	
	// Add user to database
	userID, err := db.AddUser(req.Username, string(hashedPassword), "", "", req.Email, req.Role)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create user",
		})
	}
	
	log.Printf("User created successfully by admin. ID: %d, Username: %s, Role: %s", userID, req.Username, req.Role)
	
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

// AdminUpdatePasswordHandler - Handler for updating a user's password
func AdminUpdatePasswordHandler(c echo.Context) error {
	// Ensure request body contains id and password
	var req struct {
		ID       int64  `json:"id"`
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	
	// Validate password
	if len(req.Password) < 6 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Password must be at least 6 characters"})
	}
	
	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update password"})
	}
	
	// Update password in database
	if err := db.UpdateUserPassword(int(req.ID), string(hashedPassword)); err != nil {
		log.Printf("Error updating password for user %d: %v", req.ID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update password"})
	}
	
	log.Printf("Password updated successfully for user ID: %d", req.ID)
	return c.JSON(http.StatusOK, map[string]string{"message": "Password updated successfully"})
}

// AdminUpdateUsernameHandler - Handler for updating a user's username
func AdminUpdateUsernameHandler(c echo.Context) error {
	// Ensure request body contains id and username
	var req struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	
	// Validate username
	if len(req.Username) < 3 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username must be at least 3 characters"})
	}
	
	// Update username in database
	if err := db.UpdateUsername(req.ID, req.Username); err != nil {
		log.Printf("Error updating username for user %d: %v", req.ID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	
	log.Printf("Username updated successfully for user ID: %d to '%s'", req.ID, req.Username)
	return c.JSON(http.StatusOK, map[string]string{"message": "Username updated successfully"})
}

// --- Vehicle Management Handlers ---

// AdminVehiclesHandler - Handler for getting all vehicles
func AdminVehiclesHandler(c echo.Context) error {
	// Make sure user is logged in and has admin role
	usernameCookie, err := c.Cookie("username")
	if err != nil || usernameCookie.Value == "" {
		log.Printf("Admin vehicles page access attempted without valid session")
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "You must be logged in to access this page",
		})
	}
	
	// Check role
	roleCookie, err := c.Cookie("user_role")
	if err != nil || roleCookie.Value != "Admin" {
		log.Printf("Admin vehicles page access attempted by non-admin user: %s", usernameCookie.Value)
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "You do not have permission to access the admin vehicles page",
		})
	}

	// Fetch all vehicles
	allVehicles, err := db.GetAllVehicles()
	if err != nil {
		log.Printf("Error getting all vehicles: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve vehicles",
		})
	}
	defer allVehicles.Close()

	// Convert rows to array of maps
	var vehicles []map[string]interface{}
	for allVehicles.Next() {
		var id int64
		var vehicleNumber, vehicleType, status, lastMaintenance, nextMaintenance, createdAt, notes string
		var capacity int
		
		if err := allVehicles.Scan(&id, &vehicleNumber, &vehicleType, &capacity, &status, 
								 &lastMaintenance, &nextMaintenance, &createdAt, &notes); err != nil {
			log.Printf("Error scanning vehicle row: %v", err)
			continue
		}
		
		vehicles = append(vehicles, map[string]interface{}{
			"id":                  id,
			"vehicle_number":      vehicleNumber,
			"type":                vehicleType,
			"capacity":            capacity,
			"status":              status,
			"last_maintenance":    lastMaintenance,
			"next_maintenance":    nextMaintenance,
			"created_at":          createdAt,
			"notes":               notes,
		})
	}

	return c.JSON(http.StatusOK, vehicles)
}

// AdminCreateVehicleHandler - Handler for creating a new vehicle
func AdminCreateVehicleHandler(c echo.Context) error {
	// Make sure user is logged in and has admin role
	usernameCookie, err := c.Cookie("username")
	if err != nil || usernameCookie.Value == "" {
		log.Printf("Admin create vehicle attempted without valid session")
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "You must be logged in to perform this action",
		})
	}
	
	// Check role
	roleCookie, err := c.Cookie("user_role")
	if err != nil || roleCookie.Value != "Admin" {
		log.Printf("Admin create vehicle attempted by non-admin user: %s", usernameCookie.Value)
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "You do not have permission to perform this action",
		})
	}

	// Parse JSON request
	var req struct {
		VehicleNumber    string `json:"vehicle_number"`
		Type             string `json:"type"`
		Capacity         int    `json:"capacity"`
		Status           string `json:"status"`
		LastMaintenance  string `json:"last_maintenance"`
		NextMaintenance  string `json:"next_maintenance"`
		Notes            string `json:"notes"`
	}
	
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	// Basic validation
	if req.VehicleNumber == "" || req.Type == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Vehicle number and type are required",
		})
	}

	// Set default status if not provided
	if req.Status == "" {
		req.Status = "Active"
	}

	// Add vehicle to database
	vehicleID, err := db.AddVehicle(req.VehicleNumber, req.Type, req.Capacity, req.Status, 
								  req.LastMaintenance, req.NextMaintenance, req.Notes)
	if err != nil {
		log.Printf("Error creating vehicle: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create vehicle: " + err.Error(),
		})
	}
	
	log.Printf("Vehicle created successfully by admin. ID: %d, Vehicle Number: %s", vehicleID, req.VehicleNumber)
	
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Vehicle created successfully",
		"vehicle_id": vehicleID,
	})
}

// AdminUpdateVehicleHandler - Handler for updating a vehicle
func AdminUpdateVehicleHandler(c echo.Context) error {
	// Ensure request body contains required fields
	var req struct {
		ID               int64  `json:"id"`
		VehicleNumber    string `json:"vehicle_number"`
		Type             string `json:"type"`
		Capacity         int    `json:"capacity"`
		Status           string `json:"status"`
		LastMaintenance  string `json:"last_maintenance"`
		NextMaintenance  string `json:"next_maintenance"`
		Notes            string `json:"notes"`
	}
	
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	
	// Validate required fields
	if req.VehicleNumber == "" || req.Type == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Vehicle number and type are required"})
	}

	// Update vehicle in database
	err := db.UpdateVehicle(req.ID, req.VehicleNumber, req.Type, req.Capacity, req.Status, 
						  req.LastMaintenance, req.NextMaintenance, req.Notes)
	if err != nil {
		log.Printf("Error updating vehicle %d: %v", req.ID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	
	log.Printf("Vehicle updated successfully. ID: %d, Vehicle Number: %s", req.ID, req.VehicleNumber)
	return c.JSON(http.StatusOK, map[string]string{"message": "Vehicle updated successfully"})
}

// AdminDeleteVehicleHandler - Handler for deleting a vehicle
func AdminDeleteVehicleHandler(c echo.Context) error {
	idParam := c.Param("id")
	vehicleID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid vehicle ID"})
	}
	
	if err := db.DeleteVehicle(vehicleID); err != nil {
		log.Printf("Error deleting vehicle %d: %v", vehicleID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete vehicle"})
	}
	
	log.Printf("Vehicle deleted successfully. ID: %d", vehicleID)
	return c.JSON(http.StatusOK, map[string]string{"message": "Vehicle deleted successfully"})
} 