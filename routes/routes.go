package routes

import (
	"github.com/labstack/echo/v4"

	"SecureSignIn/handlers/auth"
	"SecureSignIn/handlers/dashboard"
	"SecureSignIn/handlers/middleware"
	"SecureSignIn/handlers/templates"
)

// RegisterRoutes registers all application routes
func RegisterRoutes(e *echo.Echo) {
	// Initialize templates
	templates.InitTemplates()
	
	// Middleware
	e.Use(middleware.LogAndRecover)

	// Static files
	e.Static("/static", "static")

	// Basic routes
	e.GET("/", dashboard.IndexHandler)
	e.GET("/dashboard", dashboard.DashboardHandler)
	e.GET("/health", dashboard.HealthCheckHandler)
	
	// Auth routes
	e.GET("/login", auth.LoginHandler)
	e.POST("/auth", auth.BasicAuthHandler)
	e.GET("/logout", auth.LogoutHandler)
	
	// Registration routes
	e.GET("/register", auth.RegisterHandler)
	e.POST("/register", auth.BasicRegisterHandler)
	
	// Password reset routes
	e.GET("/forgot", auth.ForgotHandler)
	e.POST("/forgot", auth.ForgotHandler)
	e.GET("/reset/:token", auth.ShowResetFormHandler)
	e.POST("/reset/:token", auth.HandleResetPasswordHandler)
	e.GET("/security-reset", auth.SecurityQuestionResetHandler)
	e.POST("/security-reset", auth.SecurityQuestionResetHandler)
	e.GET("/setup-security", auth.SetupSecurityQuestionHandler)
	e.POST("/setup-security", auth.SetupSecurityQuestionHandler)
} 