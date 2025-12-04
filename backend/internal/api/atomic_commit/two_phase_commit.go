package atomic_commit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	
	"sds/internal/simulation/two_phase_commit"
)

// Global coordinator instance for the simulation
var coordinator *two_phase_commit.Coordinator

// init initializes the coordinator with 4 participants
// This function runs automatically when the package is imported
func init() {
	coordinator = two_phase_commit.NewCoordinator(4)
}

// setCORSHeaders sets CORS headers for cross-origin requests
// This allows the frontend (running on port 8000) to communicate with the backend (port 8080)
func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// GetState returns the current state of the coordinator and participants
// GET /api/atomic-commit/2pc/state
func GetState(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Create response with coordinator state
	response := map[string]interface{}{
		"coordinator":  map[string]interface{}{
			"state":    coordinator.State,
			"isFailed": coordinator.IsFailed,
		},
		"participants": coordinator.Participants,
		"transaction":  coordinator.Transaction,
	}
	
	// Convert to JSON
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// StartTransaction initiates a new 2PC transaction
// POST /api/atomic-commit/2pc/start-transaction?data=<transaction_data>
func StartTransaction(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Get transaction data from query parameter
	data := r.URL.Query().Get("data")
	if data == "" {
		data = "Sample Transaction"  // Default data
	}
	
	// Generate transaction ID
	transactionID := fmt.Sprintf("TX-%d", len(coordinator.ProtocolSteps)+1)
	
	// Start the transaction and get protocol steps
	steps, err := coordinator.StartTransaction(transactionID, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return the protocol steps and final state
	response := map[string]interface{}{
		"coordinator":    map[string]interface{}{
			"state":    coordinator.State,
			"isFailed": coordinator.IsFailed,
		},
		"participants":   coordinator.Participants,
		"transaction":    coordinator.Transaction,
		"protocolSteps":  steps,
	}
	
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// ResetSystem resets the coordinator and all participants
// POST /api/atomic-commit/2pc/reset
func ResetSystem(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Reset the coordinator
	coordinator.Reset()
	
	// Return the new state
	response := map[string]interface{}{
		"coordinator":  map[string]interface{}{
			"state":    coordinator.State,
			"isFailed": coordinator.IsFailed,
		},
		"participants": coordinator.Participants,
		"transaction":  nil,
	}
	
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// SetParticipantVote sets whether a participant will vote YES or NO
// POST /api/atomic-commit/2pc/set-participant-vote?participantId=<id>&canCommit=<true|false>
func SetParticipantVote(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Get participant ID
	participantIDStr := r.URL.Query().Get("participantId")
	participantID, err := strconv.Atoi(participantIDStr)
	if err != nil {
		http.Error(w, "Invalid participantId parameter", http.StatusBadRequest)
		return
	}
	
	// Get canCommit flag
	canCommitStr := r.URL.Query().Get("canCommit")
	canCommit := canCommitStr == "true"
	
	// Set the participant's vote
	err = coordinator.SetParticipantCanCommit(participantID, canCommit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Return updated state
	response := map[string]interface{}{
		"coordinator":  map[string]interface{}{
			"state":    coordinator.State,
			"isFailed": coordinator.IsFailed,
		},
		"participants": coordinator.Participants,
		"transaction":  coordinator.Transaction,
	}
	
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// SimulateFailure simulates a coordinator or participant failure
// POST /api/atomic-commit/2pc/simulate-failure?nodeType=<coordinator|participant>&nodeId=<id>&failed=<true|false>
func SimulateFailure(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Get node type
	nodeType := r.URL.Query().Get("nodeType")
	failedStr := r.URL.Query().Get("failed")
	failed := failedStr == "true"
	
	if nodeType == "coordinator" {
		coordinator.SetCoordinatorFailed(failed)
	} else if nodeType == "participant" {
		nodeIDStr := r.URL.Query().Get("nodeId")
		nodeID, err := strconv.Atoi(nodeIDStr)
		if err != nil {
			http.Error(w, "Invalid nodeId parameter", http.StatusBadRequest)
			return
		}
		
		err = coordinator.SetParticipantFailed(nodeID, failed)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "Invalid nodeType parameter (must be 'coordinator' or 'participant')", http.StatusBadRequest)
		return
	}
	
	// Return updated state
	response := map[string]interface{}{
		"coordinator":  map[string]interface{}{
			"state":    coordinator.State,
			"isFailed": coordinator.IsFailed,
		},
		"participants": coordinator.Participants,
		"transaction":  coordinator.Transaction,
	}
	
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

