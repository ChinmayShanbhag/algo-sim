package graphql

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
	
	"sds/internal/session"
	"sds/internal/simulation/graphql"
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

// GetState returns the current state of the GraphQL simulator
// GET /api/graphql/state
func GetState(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)
	
	state := userState.GraphQLSimulator.GetState()
	
	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// ExecuteQuery executes a GraphQL query (simulated or real external API)
// POST /api/graphql/query
// Body: { "query": "...", "variables": {...}, "endpoint": "https://..." (optional) }
func ExecuteQuery(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)
	
	// Parse request body
	var req struct {
		graphql.GraphQLRequest
		Endpoint string `json:"endpoint"` // Optional: for external GraphQL APIs
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	var response graphql.GraphQLResponse
	var latency int
	
	// Check if this is an external GraphQL API call
	if req.Endpoint != "" && (strings.HasPrefix(req.Endpoint, "http://") || strings.HasPrefix(req.Endpoint, "https://")) {
		// Make real external GraphQL API call
		response, latency = makeExternalGraphQLRequest(req.Endpoint, req.GraphQLRequest)
	} else {
		// Execute simulated GraphQL query
		response, latency = userState.GraphQLSimulator.ExecuteQuery(req.GraphQLRequest)
	}
	
	// Return response and updated state
	result := map[string]interface{}{
		"response": response,
		"latency":  latency,
		"state":    userState.GraphQLSimulator.GetState(),
	}
	
	responseJSON, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// makeExternalGraphQLRequest makes an actual HTTP request to an external GraphQL API
func makeExternalGraphQLRequest(endpoint string, req graphql.GraphQLRequest) (graphql.GraphQLResponse, int) {
	startTime := time.Now()
	
	// Prepare request body
	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return graphql.GraphQLResponse{
			Errors: []graphql.GraphQLError{
				{Message: "Failed to marshal request: " + err.Error()},
			},
		}, 0
	}
	
	// Create HTTP request
	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewReader(bodyJSON))
	if err != nil {
		return graphql.GraphQLResponse{
			Errors: []graphql.GraphQLError{
				{Message: "Failed to create request: " + err.Error()},
			},
		}, 0
	}
	
	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	
	// Make the request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	resp, err := client.Do(httpReq)
	if err != nil {
		return graphql.GraphQLResponse{
			Errors: []graphql.GraphQLError{
				{Message: "Request failed: " + err.Error()},
			},
		}, 0
	}
	defer resp.Body.Close()
	
	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return graphql.GraphQLResponse{
			Errors: []graphql.GraphQLError{
				{Message: "Failed to read response: " + err.Error()},
			},
		}, 0
	}
	
	// Parse GraphQL response
	var gqlResp graphql.GraphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		// If not valid GraphQL response, return as error
		return graphql.GraphQLResponse{
			Errors: []graphql.GraphQLError{
				{Message: "Invalid GraphQL response: " + err.Error()},
			},
		}, 0
	}
	
	// Calculate latency
	latency := int(time.Since(startTime).Milliseconds())
	
	return gqlResp, latency
}

// Reset resets the GraphQL simulator to initial state
// POST /api/graphql/reset
func Reset(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)
	
	userState.GraphQLSimulator.Reset()
	
	state := userState.GraphQLSimulator.GetState()
	
	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// SetupRoutes registers all GraphQL endpoints
func SetupRoutes(sm *session.Manager) {
	sessionManager = sm
	
	http.HandleFunc("/api/graphql/state", GetState)
	http.HandleFunc("/api/graphql/query", ExecuteQuery)
	http.HandleFunc("/api/graphql/reset", Reset)
}

