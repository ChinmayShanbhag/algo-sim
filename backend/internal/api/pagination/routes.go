package pagination

import (
	"encoding/json"
	"net/http"
	"strconv"

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

// GetState returns the current state
// GET /api/pagination/state
func GetState(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	state := userState.PaginationSimulator.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// GetPage returns a specific page (for pagination)
// GET /api/pagination/get-page?page=1
func GetPage(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	pageResponse := userState.PaginationSimulator.GetPage(page)

	responseJSON, err := json.Marshal(pageResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// LoadAllData loads all data at once (for virtualization)
// GET /api/pagination/load-all
func LoadAllData(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	items := userState.PaginationSimulator.LoadAllData()

	responseJSON, err := json.Marshal(map[string]interface{}{
		"items": items,
		"total": len(items),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// UpdateVirtualizationView updates the virtualization viewport
// POST /api/pagination/update-viewport?scroll=0&height=600
func UpdateVirtualizationView(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	scrollStr := r.URL.Query().Get("scroll")
	heightStr := r.URL.Query().Get("height")
	
	scroll, _ := strconv.Atoi(scrollStr)
	height, _ := strconv.Atoi(heightStr)
	
	if height == 0 {
		height = 600 // Default viewport height
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	userState.PaginationSimulator.UpdateVirtualizationView(scroll, height)
	state := userState.PaginationSimulator.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// Reset resets the simulation
// POST /api/pagination/reset
func Reset(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	userState.PaginationSimulator.Reset()
	state := userState.PaginationSimulator.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// SetupRoutes registers all pagination endpoints
func SetupRoutes(sm *session.Manager) {
	sessionManager = sm

	http.HandleFunc("/api/pagination/state", GetState)
	http.HandleFunc("/api/pagination/get-page", GetPage)
	http.HandleFunc("/api/pagination/load-all", LoadAllData)
	http.HandleFunc("/api/pagination/update-viewport", UpdateVirtualizationView)
	http.HandleFunc("/api/pagination/reset", Reset)
}

