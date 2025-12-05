package bloomfilter

import (
	"encoding/json"
	"net/http"

	"sds/internal/session"
)

var sessionManager *session.Manager

// getSessionID extracts the session ID from the request header
func getSessionID(r *http.Request) string {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID != "" {
		return sessionID
	}
	sessionID = r.URL.Query().Get("session_id")
	if sessionID != "" {
		return sessionID
	}
	return "default"
}

// setCORSHeaders sets CORS headers for cross-origin requests
func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Session-ID")
}

// GetState returns the current state of the Bloom Filter
// GET /api/bloomfilter/state
func GetState(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	state := userState.BloomFilter.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// AddItem adds an item to the Bloom Filter
// POST /api/bloomfilter/add?item=...
func AddItem(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	item := r.URL.Query().Get("item")
	if item == "" {
		http.Error(w, "item parameter is required", http.StatusBadRequest)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	userState.BloomFilter.Add(item)
	state := userState.BloomFilter.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// CheckItem checks if an item might be in the Bloom Filter
// GET /api/bloomfilter/check?item=...
func CheckItem(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	item := r.URL.Query().Get("item")
	if item == "" {
		http.Error(w, "item parameter is required", http.StatusBadRequest)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	result := userState.BloomFilter.Check(item)

	responseJSON, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// Reset resets the Bloom Filter
// POST /api/bloomfilter/reset
func Reset(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	userState.BloomFilter.Reset()
	state := userState.BloomFilter.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// SetupRoutes registers all Bloom Filter endpoints
func SetupRoutes(sm *session.Manager) {
	sessionManager = sm

	http.HandleFunc("/api/bloomfilter/state", GetState)
	http.HandleFunc("/api/bloomfilter/add", AddItem)
	http.HandleFunc("/api/bloomfilter/check", CheckItem)
	http.HandleFunc("/api/bloomfilter/reset", Reset)
}

