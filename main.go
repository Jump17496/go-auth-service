package main

import (
	"log"
	"net/http"

	"go-auth-service/internal/config"
	"go-auth-service/internal/database"
	"go-auth-service/internal/handlers"
	"go-auth-service/internal/middleware"

	"github.com/gorilla/mux"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Initialize(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg.JWTSecret)

	// Setup routes
	router := mux.NewRouter()

	// CORS middleware
	router.Use(middleware.CORS)

	// Public routes
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/auth/refresh", authHandler.RefreshToken).Methods("POST", "OPTIONS")

	// Protected routes
	protected := router.PathPrefix("/api/auth").Subrouter()
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	protected.HandleFunc("/user", authHandler.GetUser).Methods("GET", "OPTIONS")

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
