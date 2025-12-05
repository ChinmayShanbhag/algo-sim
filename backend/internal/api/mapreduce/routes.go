package mapreduce

import (
	"encoding/json"
	"net/http"

	"sds/internal/session"
)

var sessionManager *session.Manager

// Helper function to extract session ID from request
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

// GetJobState returns the current state of the MapReduce job
// GET /api/mapreduce/state
func GetJobState(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	state := userState.MapReduceJob.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// StartJob starts the MapReduce job
// POST /api/mapreduce/start
func StartJob(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	userState.MapReduceJob.Start()
	state := userState.MapReduceJob.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// ExecuteMapPhase executes the map phase
// POST /api/mapreduce/execute-map
func ExecuteMapPhase(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	userState.MapReduceJob.ExecuteMapPhase()
	state := userState.MapReduceJob.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// ExecuteShufflePhase executes the shuffle phase
// POST /api/mapreduce/execute-shuffle
func ExecuteShufflePhase(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	userState.MapReduceJob.ExecuteShufflePhase()
	state := userState.MapReduceJob.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// ExecuteReducePhase executes the reduce phase
// POST /api/mapreduce/execute-reduce
func ExecuteReducePhase(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	userState.MapReduceJob.ExecuteReducePhase()
	state := userState.MapReduceJob.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// ResetJob resets the MapReduce job
// POST /api/mapreduce/reset
func ResetJob(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	userState.MapReduceJob.Reset()
	state := userState.MapReduceJob.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// SetupRoutes registers all MapReduce endpoints
func SetupRoutes(sm *session.Manager) {
	sessionManager = sm

	http.HandleFunc("/api/mapreduce/state", GetJobState)
	http.HandleFunc("/api/mapreduce/start", StartJob)
	http.HandleFunc("/api/mapreduce/execute-map", ExecuteMapPhase)
	http.HandleFunc("/api/mapreduce/execute-shuffle", ExecuteShufflePhase)
	http.HandleFunc("/api/mapreduce/execute-reduce", ExecuteReducePhase)
	http.HandleFunc("/api/mapreduce/reset", ResetJob)
}

