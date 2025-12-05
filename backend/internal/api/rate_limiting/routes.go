package rate_limiting

import (
	"encoding/json"
	"net/http"
	"strconv"
	
	"sds/internal/session"
)

var sessionManager *session.Manager

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

// setCORSHeaders sets CORS headers for cross-origin requests
func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Session-ID")
}

// GetAllStates returns the current state of all rate limiters
// GET /api/rate-limiting/state
func GetAllStates(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Get user's session
	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)
	
	response := map[string]interface{}{
		"fixedWindow":   userState.FixedWindow.GetState(),
		"slidingLog":    userState.SlidingLog.GetState(),
		"slidingWindow": userState.SlidingWindow.GetState(),
		"tokenBucket":   userState.TokenBucket.GetState(),
		"leakyBucket":   userState.LeakyBucket.GetState(),
	}
	
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// SendRequest sends a single request to all rate limiters
// POST /api/rate-limiting/send-request
func SendRequest(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Get user's session
	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)
	
	// Send request to all rate limiters
	results := map[string]interface{}{
		"fixedWindow":   userState.FixedWindow.AllowRequest(),
		"slidingLog":    userState.SlidingLog.AllowRequest(),
		"slidingWindow": userState.SlidingWindow.AllowRequest(),
		"tokenBucket":   userState.TokenBucket.AllowRequest(),
		"leakyBucket":   userState.LeakyBucket.AllowRequest(),
	}
	
	// Get updated states
	response := map[string]interface{}{
		"results": results,
		"states": map[string]interface{}{
			"fixedWindow":   userState.FixedWindow.GetState(),
			"slidingLog":    userState.SlidingLog.GetState(),
			"slidingWindow": userState.SlidingWindow.GetState(),
			"tokenBucket":   userState.TokenBucket.GetState(),
			"leakyBucket":   userState.LeakyBucket.GetState(),
		},
	}
	
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// SendBurstRequests sends multiple requests at once
// POST /api/rate-limiting/send-burst?count=<number>
func SendBurstRequests(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Get user's session
	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)
	
	// Get count from query parameter
	countStr := r.URL.Query().Get("count")
	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 {
		count = 20 // Default to 20
	}
	
	// Send multiple requests to all rate limiters
	allResults := make([]map[string]interface{}, count)
	
	for i := 0; i < count; i++ {
		allResults[i] = map[string]interface{}{
			"fixedWindow":   userState.FixedWindow.AllowRequest(),
			"slidingLog":    userState.SlidingLog.AllowRequest(),
			"slidingWindow": userState.SlidingWindow.AllowRequest(),
			"tokenBucket":   userState.TokenBucket.AllowRequest(),
			"leakyBucket":   userState.LeakyBucket.AllowRequest(),
		}
	}
	
	// Get final states
	response := map[string]interface{}{
		"count":   count,
		"results": allResults,
		"states": map[string]interface{}{
			"fixedWindow":   userState.FixedWindow.GetState(),
			"slidingLog":    userState.SlidingLog.GetState(),
			"slidingWindow": userState.SlidingWindow.GetState(),
			"tokenBucket":   userState.TokenBucket.GetState(),
			"leakyBucket":   userState.LeakyBucket.GetState(),
		},
	}
	
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// ResetAll resets all rate limiters
// POST /api/rate-limiting/reset
func ResetAll(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Get user's session
	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)
	
	userState.FixedWindow.Reset()
	userState.SlidingLog.Reset()
	userState.SlidingWindow.Reset()
	userState.TokenBucket.Reset()
	userState.LeakyBucket.Reset()
	
	// Return new states
	response := map[string]interface{}{
		"fixedWindow":   userState.FixedWindow.GetState(),
		"slidingLog":    userState.SlidingLog.GetState(),
		"slidingWindow": userState.SlidingWindow.GetState(),
		"tokenBucket":   userState.TokenBucket.GetState(),
		"leakyBucket":   userState.LeakyBucket.GetState(),
	}
	
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// SetupRoutes registers all rate limiting endpoints
func SetupRoutes(sm *session.Manager) {
	sessionManager = sm
	
	http.HandleFunc("/api/rate-limiting/state", GetAllStates)
	http.HandleFunc("/api/rate-limiting/send-request", SendRequest)
	http.HandleFunc("/api/rate-limiting/send-burst", SendBurstRequests)
	http.HandleFunc("/api/rate-limiting/reset", ResetAll)
}

