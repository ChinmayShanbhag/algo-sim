package atomic_commit

import (
	"net/http"
	
	"sds/internal/session"
)

var sessionManager *session.Manager

// SetupRoutes registers all atomic commit (2PC and 3PC) related endpoints
// This function is called from the main API routes setup
func SetupRoutes(sm *session.Manager) {
	sessionManager = sm
	
	// Two-Phase Commit endpoints
	http.HandleFunc("/api/atomic-commit/2pc/state", GetState)
	http.HandleFunc("/api/atomic-commit/2pc/start-transaction", StartTransaction)
	http.HandleFunc("/api/atomic-commit/2pc/reset", ResetSystem)
	http.HandleFunc("/api/atomic-commit/2pc/set-participant-vote", SetParticipantVote)
	http.HandleFunc("/api/atomic-commit/2pc/simulate-failure", SimulateFailure)
	
	// Three-Phase Commit endpoints
	http.HandleFunc("/api/atomic-commit/3pc/state", GetState3PC)
	http.HandleFunc("/api/atomic-commit/3pc/start-transaction", StartTransaction3PC)
	http.HandleFunc("/api/atomic-commit/3pc/reset", ResetSystem3PC)
	http.HandleFunc("/api/atomic-commit/3pc/set-participant-vote", SetParticipantVote3PC)
	http.HandleFunc("/api/atomic-commit/3pc/simulate-failure", SimulateFailure3PC)
}

// Helper function to extract session ID from request
func getSessionID(r *http.Request) string {
	// Try header first
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID != "" {
		return sessionID
	}
	
	// Fallback to query parameter
	sessionID = r.URL.Query().Get("session_id")
	if sessionID != "" {
		return sessionID
	}
	
	// Default session for backward compatibility
	return "default"
}

