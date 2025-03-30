package main

import (
        "fmt"
        "html/template"
        "net/http"
        "path/filepath"
        "strings"
        
        "go-transportation-portal/db"
)

// Template cache
var templates = make(map[string]*template.Template)

// Load templates on init
func init() {
        templatesDir := "templates"
        layouts, err := filepath.Glob(templatesDir + "/base.html")
        if err != nil {
                panic(err)
        }

        includes, err := filepath.Glob(templatesDir + "/*.html")
        if err != nil {
                panic(err)
        }

        // Generate our templates map from our layouts/ and includes/ directories
        for _, include := range includes {
                // Skip the base layout since we'll use it as our template base
                if include == templatesDir+"/base.html" {
                        continue
                }
                
                files := []string{include}
                files = append(files, layouts...)
                
                fileName := filepath.Base(include)
                templates[fileName] = template.Must(template.ParseFiles(files...))
        }
}

// Render a template given a model
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
        // Ensure the template exists in the map
        t, ok := templates[tmpl]
        if !ok {
                http.Error(w, fmt.Sprintf("The template %s does not exist.", tmpl), http.StatusInternalServerError)
                return
        }

        // Execute the template
        err := t.ExecuteTemplate(w, "base", data)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
        }
}

// Page data struct
type PageData struct {
        Title      string
        Error      string
        Success    string
        ActivePage string
        Users      []map[string]interface{}
        LoginLogs  []map[string]interface{}
}

// Index handler - Home page
func indexHandler(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/" {
                http.NotFound(w, r)
                return
        }

        data := PageData{
                Title:      "Transportation Portal - Home",
                ActivePage: "home",
        }

        // Check for success message
        if successMsg := r.URL.Query().Get("success"); successMsg != "" {
                data.Success = successMsg
        }

        // Get user data from database
        users, err := db.GetAllUsers()
        if err != nil {
                // Log the error but don't use fallback data
                fmt.Printf("Error fetching users: %v\n", err)
                // Return empty array instead of fallback data
                users = []map[string]interface{}{}
                // Add a message in the UI
                data.Error = "Unable to fetch user data from the database. Please try again later."
        }
        data.Users = users

        // Get login logs from database
        logs, err := db.GetRecentLoginLogs(10)
        if err != nil {
                // Log the error but don't use fallback data
                fmt.Printf("Error fetching login logs: %v\n", err)
                // Return empty array instead of fallback data
                logs = []map[string]interface{}{}
                // Add additional message only if we don't already have an error
                if data.Error == "" {
                        data.Error = "Unable to fetch login history. Please try again later."
                }
        }
        data.LoginLogs = logs

        renderTemplate(w, "index.html", data)
}

// Login handler - Render login page
func loginHandler(w http.ResponseWriter, r *http.Request) {
        data := PageData{
                Title:      "Transportation Portal - Sign In",
                ActivePage: "login",
        }

        // Check if error message is passed
        if errMsg := r.URL.Query().Get("error"); errMsg != "" {
                data.Error = errMsg
        }

        // Check if success message is passed
        if successMsg := r.URL.Query().Get("success"); successMsg != "" {
                data.Success = successMsg
        }

        renderTemplate(w, "login.html", data)
}

// Auth handler - Process login form
func basicAuthHandler(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
        }

        // Parse form
        if err := r.ParseForm(); err != nil {
                http.Redirect(w, r, "/login?error=Error processing form", http.StatusSeeOther)
                return
        }

        username := strings.TrimSpace(r.FormValue("username"))
        password := strings.TrimSpace(r.FormValue("password"))

        // Validate input
        if username == "" || password == "" {
                http.Redirect(w, r, "/login?error=Username and password are required", http.StatusSeeOther)
                return
        }

        // Validate user against database
        valid, err := db.ValidateUser(username, password)
        
        // Get IP address for logging
        ipAddress := r.RemoteAddr
        if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
                ipAddress = ip
        }
        
        // Log this login attempt
        logErr := db.LogLoginAttempt(username, valid && err == nil, ipAddress)
        if logErr != nil {
                fmt.Printf("Failed to log login attempt: %v\n", logErr)
        }

        if err != nil {
                fmt.Printf("Error during login validation: %v\n", err)
                http.Redirect(w, r, "/login?error=System error, please try again", http.StatusSeeOther)
                return
        }

        if !valid {
                http.Redirect(w, r, "/login?error=Invalid username or password", http.StatusSeeOther)
                return
        }

        // Successful login
        http.Redirect(w, r, "/?success=Successfully logged in", http.StatusSeeOther)
}

// Forgot password handler
func forgotHandler(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" {
                // Parse form
                if err := r.ParseForm(); err != nil {
                        http.Redirect(w, r, "/forgot?error=Error processing form", http.StatusSeeOther)
                        return
                }

                email := strings.TrimSpace(r.FormValue("email"))
                if email == "" {
                        http.Redirect(w, r, "/forgot?error=Email is required", http.StatusSeeOther)
                        return
                }

                // In a real app, you would send a password reset email
                // For demo purposes, just return success
                data := PageData{
                        Title:   "Transportation Portal - Forgot Password",
                        Success: "Password reset instructions sent to your email",
                }
                renderTemplate(w, "forgot.html", data)
        } else {
                data := PageData{
                        Title: "Transportation Portal - Forgot Password",
                }
                
                // Check if error message is passed
                if errMsg := r.URL.Query().Get("error"); errMsg != "" {
                        data.Error = errMsg
                }
                
                renderTemplate(w, "forgot.html", data)
        }
}

// Reset password handler
func resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
        // In a real app, this would validate a reset token and allow setting a new password
        http.Redirect(w, r, "/login?error=Reset functionality not implemented in demo", http.StatusSeeOther)
}

// Register handler - For user registration
func basicRegisterHandler(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" {
                // Parse form
                if err := r.ParseForm(); err != nil {
                        http.Redirect(w, r, "/register?error=Error processing form", http.StatusSeeOther)
                        return
                }

                username := strings.TrimSpace(r.FormValue("username"))
                password := strings.TrimSpace(r.FormValue("password"))
                confirmPassword := strings.TrimSpace(r.FormValue("confirmPassword"))

                // Validate input
                if username == "" || password == "" || confirmPassword == "" {
                        http.Redirect(w, r, "/register?error=All fields are required", http.StatusSeeOther)
                        return
                }

                if password != confirmPassword {
                        http.Redirect(w, r, "/register?error=Passwords do not match", http.StatusSeeOther)
                        return
                }

                // Check if user already exists
                exists, err := db.UserExists(username)
                if err != nil {
                        fmt.Printf("Error checking if user exists: %v\n", err)
                        http.Redirect(w, r, "/register?error=System error, please try again", http.StatusSeeOther)
                        return
                }
                
                if exists {
                        http.Redirect(w, r, "/register?error=Username already exists", http.StatusSeeOther)
                        return
                }
                
                // Create the user in the database
                err = db.CreateUser(username, password)
                if err != nil {
                        fmt.Printf("Error creating user: %v\n", err)
                        http.Redirect(w, r, "/register?error=Failed to create user: "+err.Error(), http.StatusSeeOther)
                        return
                }
                
                // Log the successful registration
                ipAddress := r.RemoteAddr
                if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
                        ipAddress = ip
                }
                db.LogLoginAttempt(username, true, ipAddress)
                
                http.Redirect(w, r, "/login?success=Registration successful! Please log in", http.StatusSeeOther)
        } else {
                data := PageData{
                        Title:      "Transportation Portal - Register",
                        ActivePage: "register",
                }

                // Check if error or success message is passed
                if errMsg := r.URL.Query().Get("error"); errMsg != "" {
                        data.Error = errMsg
                }
                if successMsg := r.URL.Query().Get("success"); successMsg != "" {
                        data.Success = successMsg
                }

                renderTemplate(w, "register.html", data)
        }
}
