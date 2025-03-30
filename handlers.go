package main

import (
        "fmt"
        "html/template"
        "net/http"
        "path/filepath"
        "strings"
        "time"
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

        // Add dummy user data for the dashboard
        dummyUsers := []map[string]interface{}{
                {
                        "id":         1,
                        "username":   "admin",
                        "created_at": time.Now().Add(-24 * time.Hour),
                },
                {
                        "id":         2,
                        "username":   "user1",
                        "created_at": time.Now().Add(-12 * time.Hour),
                },
        }
        data.Users = dummyUsers

        // Add dummy login logs
        dummyLogs := []map[string]interface{}{
                {
                        "id":         1,
                        "username":   "admin",
                        "login_time": time.Now().Add(-1 * time.Hour),
                        "success":    true,
                        "ip_address": "127.0.0.1",
                },
                {
                        "id":         2,
                        "username":   "user1",
                        "login_time": time.Now().Add(-30 * time.Minute),
                        "success":    true,
                        "ip_address": "127.0.0.1",
                },
        }
        data.LoginLogs = dummyLogs

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

        // For demo purposes, accept any username/password
        // In a real app, you would validate against a database
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

                // In a real app, we would create the user in a database
                // For demo purposes, just redirect with success message
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
