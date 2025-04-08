package main

import (
	"log"
	"net/http"
	"time"

	"SecureSignIn/db"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Create a new Echo instance
	e := echo.New()
	e.Debug = true

	// Add middleware
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "time=${time_rfc3339} method=${method}, uri=${uri}, status=${status}, latency=${latency}, error=${error}\n",
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	// Add detailed request logging middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			log.Printf("Detailed request info - Method: %s, Path: %s, RemoteAddr: %s, Headers: %v",
				req.Method, req.URL.Path, req.RemoteAddr, req.Header)

			err := next(c)

			if err != nil {
				log.Printf("Error handling request: %v", err)
			}

			return err
		}
	})

	e.Use(logAndRecover)

	// Serve static files
	e.Static("/static", "static")

	// CORS middleware with more detailed configuration
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	// Simple test endpoint
	e.GET("/test", func(c echo.Context) error {
		log.Printf("Test endpoint hit from %s", c.RealIP())
		return c.String(200, "Test endpoint working!")
	})

	// Initialize database
	if err := db.InitializeDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Routes
	e.GET("/", indexHandler)
	e.GET("/dashboard", dashboardHandler)
	e.GET("/login", loginHandler)
	e.POST("/auth", basicAuthHandler)
	e.GET("/forgot", forgotHandler)
	e.POST("/forgot", forgotHandler)
	e.GET("/reset/:token", showResetFormHandler)
	e.POST("/reset/:token", handleResetPasswordHandler)
	e.GET("/health", healthCheckHandler)
	e.GET("/register", registerHandler)
	e.POST("/register", basicRegisterHandler)
	e.GET("/logout", logoutHandler)

	// Create custom server with timeouts
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server
	log.Printf("Starting server on IPv4 http://0.0.0.0:8080")
	log.Printf("Starting server on IPv6 http://[::]:8080")
	if err := e.StartServer(server); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
