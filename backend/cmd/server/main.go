package main

import (
	"log"
	"net/http"
	"time"

	"sds/internal/api"
	"sds/internal/session"
)

func main() {
	// Initialize session manager
	sessionManager := session.NewManager()
	log.Println("✅ Session manager initialized")

	// Start background cleanup task
	// Removes sessions inactive for more than 1 hour, runs every 10 minutes
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			deleted := sessionManager.CleanupExpired(1 * time.Hour)
			if deleted > 0 {
				log.Printf("🧹 Cleaned up %d expired sessions (total active: %d)", deleted, sessionManager.Count())
			}
		}
	}()

	// Initialize API routes with session manager
	api.SetupRoutes(sessionManager)

	log.Println("🚀 SDS backend starting on :8080")
	log.Println("📊 Multi-user sessions enabled - each user gets isolated state")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
