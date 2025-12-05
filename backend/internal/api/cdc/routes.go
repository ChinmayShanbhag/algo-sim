package cdc

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

// GetState returns the current state of the CDC system
// GET /api/cdc/state
func GetState(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	state := userState.CDCSystem.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// InsertRecord inserts a new record into the database
// POST /api/cdc/insert?name=...&email=...&status=...
func InsertRecord(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	name := r.URL.Query().Get("name")
	email := r.URL.Query().Get("email")
	status := r.URL.Query().Get("status")

	if name == "" || email == "" || status == "" {
		http.Error(w, "name, email, and status are required", http.StatusBadRequest)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	userState.CDCSystem.InsertRecord(name, email, status)
	state := userState.CDCSystem.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// UpdateRecord updates an existing record
// POST /api/cdc/update?id=...&name=...&email=...&status=...
func UpdateRecord(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	idStr := r.URL.Query().Get("id")
	name := r.URL.Query().Get("name")
	email := r.URL.Query().Get("email")
	status := r.URL.Query().Get("status")

	if idStr == "" || name == "" || email == "" || status == "" {
		http.Error(w, "id, name, email, and status are required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	_, err = userState.CDCSystem.UpdateRecord(id, name, email, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	state := userState.CDCSystem.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// DeleteRecord deletes a record
// POST /api/cdc/delete?id=...
func DeleteRecord(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	_, err = userState.CDCSystem.DeleteRecord(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	state := userState.CDCSystem.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// StreamToKafka streams changes to Kafka
// POST /api/cdc/stream-to-kafka
func StreamToKafka(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	userState.CDCSystem.StreamToKafka()
	state := userState.CDCSystem.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// ConsumeFromKafka processes Kafka messages
// POST /api/cdc/consume-from-kafka
func ConsumeFromKafka(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	userState.CDCSystem.ConsumeFromKafka()
	state := userState.CDCSystem.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// Reset resets the CDC system
// POST /api/cdc/reset
func Reset(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	userState.CDCSystem.Reset()
	state := userState.CDCSystem.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// SetupRoutes registers all CDC endpoints
func SetupRoutes(sm *session.Manager) {
	sessionManager = sm

	http.HandleFunc("/api/cdc/state", GetState)
	http.HandleFunc("/api/cdc/insert", InsertRecord)
	http.HandleFunc("/api/cdc/update", UpdateRecord)
	http.HandleFunc("/api/cdc/delete", DeleteRecord)
	http.HandleFunc("/api/cdc/stream-to-kafka", StreamToKafka)
	http.HandleFunc("/api/cdc/consume-from-kafka", ConsumeFromKafka)
	http.HandleFunc("/api/cdc/reset", Reset)
}

