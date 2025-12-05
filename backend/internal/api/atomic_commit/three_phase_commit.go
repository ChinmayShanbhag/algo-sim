package atomic_commit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// GetState3PC returns the current state of the 3PC coordinator and participants
// GET /api/atomic-commit/3pc/state
func GetState3PC(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get user's session
	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)
	coordinator3PC := userState.ThreePCCoordinator

	// Create response with coordinator state
	response := map[string]interface{}{
		"coordinator": map[string]interface{}{
			"state":    coordinator3PC.State,
			"isFailed": coordinator3PC.IsFailed,
		},
		"participants": coordinator3PC.Participants,
		"transaction":  coordinator3PC.Transaction,
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

// StartTransaction3PC initiates a new 3PC transaction
// POST /api/atomic-commit/3pc/start-transaction?data=<transaction_data>
func StartTransaction3PC(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get user's session
	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)
	coordinator3PC := userState.ThreePCCoordinator

	// Get transaction data from query parameter
	data := r.URL.Query().Get("data")
	if data == "" {
		data = "Sample Transaction" // Default data
	}

	// Generate transaction ID
	transactionID := fmt.Sprintf("TX-%d", len(coordinator3PC.ProtocolSteps)+1)

	// Start the transaction and get protocol steps
	steps, err := coordinator3PC.StartTransaction(transactionID, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the protocol steps and final state
	response := map[string]interface{}{
		"coordinator": map[string]interface{}{
			"state":    coordinator3PC.State,
			"isFailed": coordinator3PC.IsFailed,
		},
		"participants":  coordinator3PC.Participants,
		"transaction":   coordinator3PC.Transaction,
		"protocolSteps": steps,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// ResetSystem3PC resets the 3PC coordinator and all participants
// POST /api/atomic-commit/3pc/reset
func ResetSystem3PC(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get user's session
	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)
	coordinator3PC := userState.ThreePCCoordinator

	// Reset the coordinator
	coordinator3PC.Reset()

	// Return the new state
	response := map[string]interface{}{
		"coordinator": map[string]interface{}{
			"state":    coordinator3PC.State,
			"isFailed": coordinator3PC.IsFailed,
		},
		"participants": coordinator3PC.Participants,
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

// SetParticipantVote3PC sets whether a participant will vote YES or NO
// POST /api/atomic-commit/3pc/set-participant-vote?participantId=<id>&canCommit=<true|false>
func SetParticipantVote3PC(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get user's session
	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)
	coordinator3PC := userState.ThreePCCoordinator

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
	err = coordinator3PC.SetParticipantCanCommit(participantID, canCommit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return updated state
	response := map[string]interface{}{
		"coordinator": map[string]interface{}{
			"state":    coordinator3PC.State,
			"isFailed": coordinator3PC.IsFailed,
		},
		"participants": coordinator3PC.Participants,
		"transaction":  coordinator3PC.Transaction,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// SimulateFailure3PC simulates a coordinator or participant failure
// POST /api/atomic-commit/3pc/simulate-failure?nodeType=<coordinator|participant>&nodeId=<id>&failed=<true|false>
func SimulateFailure3PC(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get user's session
	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)
	coordinator3PC := userState.ThreePCCoordinator

	// Get node type
	nodeType := r.URL.Query().Get("nodeType")
	failedStr := r.URL.Query().Get("failed")
	failed := failedStr == "true"

	if nodeType == "coordinator" {
		coordinator3PC.SetCoordinatorFailed(failed)
	} else if nodeType == "participant" {
		nodeIDStr := r.URL.Query().Get("nodeId")
		nodeID, err := strconv.Atoi(nodeIDStr)
		if err != nil {
			http.Error(w, "Invalid nodeId parameter", http.StatusBadRequest)
			return
		}

		err = coordinator3PC.SetParticipantFailed(nodeID, failed)
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
		"coordinator": map[string]interface{}{
			"state":    coordinator3PC.State,
			"isFailed": coordinator3PC.IsFailed,
		},
		"participants": coordinator3PC.Participants,
		"transaction":  coordinator3PC.Transaction,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

