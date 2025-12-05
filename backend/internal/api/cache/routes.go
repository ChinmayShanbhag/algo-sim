package cache

import (
	"encoding/json"
	"net/http"

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

// GetAllStates returns the current state of all cache algorithms
// GET /api/cache/state
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
		"lru":  userState.LRUCache.GetState(),
		"lfu":  userState.LFUCache.GetState(),
		"fifo": userState.FIFOCache.GetState(),
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// AccessCache performs a GET or PUT operation on all caches
// POST /api/cache/access?operation=<GET|PUT>&key=<key>&value=<value>
func AccessCache(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get user's session
	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	// Get parameters
	operation := r.URL.Query().Get("operation") // GET or PUT
	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")

	if operation == "" || key == "" {
		http.Error(w, "Missing required parameters: operation and key", http.StatusBadRequest)
		return
	}

	results := make(map[string]interface{})

	if operation == "GET" {
		// Perform GET on all caches
		lruValue, lruHit := userState.LRUCache.Get(key)
		lfuValue, lfuHit := userState.LFUCache.Get(key)
		fifoValue, fifoHit := userState.FIFOCache.Get(key)

		results["lru"] = map[string]interface{}{"hit": lruHit, "value": lruValue}
		results["lfu"] = map[string]interface{}{"hit": lfuHit, "value": lfuValue}
		results["fifo"] = map[string]interface{}{"hit": fifoHit, "value": fifoValue}
	} else if operation == "PUT" {
		if value == "" {
			http.Error(w, "Missing required parameter: value", http.StatusBadRequest)
			return
		}

		// Perform PUT on all caches
		lruEvicted := userState.LRUCache.Put(key, value)
		lfuEvicted := userState.LFUCache.Put(key, value)
		fifoEvicted := userState.FIFOCache.Put(key, value)

		results["lru"] = map[string]interface{}{"evicted": lruEvicted}
		results["lfu"] = map[string]interface{}{"evicted": lfuEvicted}
		results["fifo"] = map[string]interface{}{"evicted": fifoEvicted}
	} else {
		http.Error(w, "Invalid operation. Must be GET or PUT", http.StatusBadRequest)
		return
	}

	// Get updated states
	response := map[string]interface{}{
		"results": results,
		"states": map[string]interface{}{
			"lru":  userState.LRUCache.GetState(),
			"lfu":  userState.LFUCache.GetState(),
			"fifo": userState.FIFOCache.GetState(),
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

// ResetAll resets all cache algorithms
// POST /api/cache/reset
func ResetAll(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get user's session
	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	userState.LRUCache.Reset()
	userState.LFUCache.Reset()
	userState.FIFOCache.Reset()

	// Return new states
	response := map[string]interface{}{
		"lru":  userState.LRUCache.GetState(),
		"lfu":  userState.LFUCache.GetState(),
		"fifo": userState.FIFOCache.GetState(),
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// SetupRoutes registers all cache eviction endpoints
func SetupRoutes(sm *session.Manager) {
	sessionManager = sm

	http.HandleFunc("/api/cache/state", GetAllStates)
	http.HandleFunc("/api/cache/access", AccessCache)
	http.HandleFunc("/api/cache/reset", ResetAll)
}

