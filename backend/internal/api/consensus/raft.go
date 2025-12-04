package consensus

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sds/internal/simulation/raft"
)

var raftCluster *raft.Cluster

func init() {
	// Initialize a default cluster with 5 nodes (typical for Raft)
	raftCluster = raft.NewCluster(5)
}

// setCORSHeaders sets CORS headers for all responses
func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// GetRaftState returns the current state of the Raft cluster
func GetRaftState(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	state, err := raftCluster.GetState()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(state)
}

// StartElection triggers a leader election from a specific node (step by step)
func StartElection(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Get node ID from query parameter
	nodeIDStr := r.URL.Query().Get("nodeId")
	nodeID, err := strconv.Atoi(nodeIDStr)
	if err != nil {
		http.Error(w, "Invalid nodeId parameter", http.StatusBadRequest)
		return
	}
	
	// Start election step by step
	steps, err := raftCluster.StartElectionStepByStep(nodeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Get state with steps
	state, err := raftCluster.GetState()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Add steps to response
	type Response struct {
		Nodes        interface{} `json:"nodes"`
		ElectionSteps interface{} `json:"electionSteps"`
	}
	
	var clusterState map[string]interface{}
	json.Unmarshal(state, &clusterState)
	
	response := Response{
		Nodes:        clusterState["nodes"],
		ElectionSteps: steps,
	}
	
	responseJSON, _ := json.Marshal(response)
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// ResetCluster resets all nodes to initial state
func ResetCluster(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	raftCluster.Reset()
	
	state, err := raftCluster.GetState()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(state)
}

// SetLeader sets a specific node as leader
func SetLeader(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Get node ID from query parameter
	nodeIDStr := r.URL.Query().Get("nodeId")
	nodeID, err := strconv.Atoi(nodeIDStr)
	if err != nil {
		http.Error(w, "Invalid nodeId parameter", http.StatusBadRequest)
		return
	}
	
	err = raftCluster.SetLeader(nodeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	state, err := raftCluster.GetState()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(state)
}

