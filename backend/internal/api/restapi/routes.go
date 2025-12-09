package restapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
	
	"sds/internal/session"
	"sds/internal/simulation/restapi"
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
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Session-ID")
}

// GetState returns the current state of the REST API simulator
// GET /api/restapi/state
func GetState(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Get user's session
	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)
	
	state := userState.RESTAPISimulator.GetState()
	
	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// HandleAPIRequest handles a REST API request (both simulated and real external APIs)
// POST /api/restapi/request
// Body: APIRequest with optional "mode" field
func HandleAPIRequest(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Get user's session
	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)
	
	// Parse request body
	var apiReq struct {
		restapi.APIRequest
		Mode string `json:"mode"` // "simulated" or "real"
	}
	if err := json.NewDecoder(r.Body).Decode(&apiReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	var apiResp restapi.APIResponse
	
	// Check if this is a real external API call or simulated
	if apiReq.Mode == "real" || strings.HasPrefix(apiReq.Path, "http://") || strings.HasPrefix(apiReq.Path, "https://") {
		// Make real external API call
		apiResp = makeRealAPIRequest(apiReq.APIRequest)
	} else {
		// Handle simulated API request
		apiResp = userState.RESTAPISimulator.HandleRequest(apiReq.APIRequest)
	}
	
	// Return response and updated state
	response := map[string]interface{}{
		"response": apiResp,
		"state":    userState.RESTAPISimulator.GetState(),
	}
	
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// makeRealAPIRequest makes an actual HTTP request to an external API
func makeRealAPIRequest(apiReq restapi.APIRequest) restapi.APIResponse {
	startTime := time.Now()
	
	// Prepare request body
	var bodyReader io.Reader
	if apiReq.Body != nil && len(apiReq.Body) > 0 {
		bodyJSON, err := json.Marshal(apiReq.Body)
		if err != nil {
			return createErrorResponse(400, "Bad Request", "Invalid request body")
		}
		bodyReader = bytes.NewReader(bodyJSON)
	}
	
	// Create HTTP request
	req, err := http.NewRequest(string(apiReq.Method), apiReq.Path, bodyReader)
	if err != nil {
		return createErrorResponse(400, "Bad Request", err.Error())
	}
	
	// Add headers
	for key, value := range apiReq.Headers {
		req.Header.Set(key, value)
	}
	
	// Add query parameters
	if len(apiReq.Query) > 0 {
		q := req.URL.Query()
		for key, value := range apiReq.Query {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}
	
	// Make the request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return createErrorResponse(0, "Request Failed", err.Error())
	}
	defer resp.Body.Close()
	
	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return createErrorResponse(0, "Response Read Failed", err.Error())
	}
	
	// Parse response body as JSON if possible
	var bodyData interface{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &bodyData); err != nil {
			// If not JSON, return as string
			bodyData = string(respBody)
		}
	}
	
	// Build response headers map
	responseHeaders := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			responseHeaders[key] = values[0]
		}
	}
	
	// Calculate latency
	latency := int(time.Since(startTime).Milliseconds())
	
	return restapi.APIResponse{
		StatusCode: resp.StatusCode,
		StatusText: resp.Status,
		Headers:    responseHeaders,
		Body:       bodyData,
		Latency:    latency,
		Timestamp:  time.Now(),
	}
}

// createErrorResponse creates an error response
func createErrorResponse(statusCode int, statusText, message string) restapi.APIResponse {
	return restapi.APIResponse{
		StatusCode: statusCode,
		StatusText: statusText,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body: map[string]interface{}{
			"error": map[string]interface{}{
				"code":    statusCode,
				"message": message,
			},
		},
		Latency:   0,
		Timestamp: time.Now(),
	}
}

// Reset resets the REST API simulator to initial state
// POST /api/restapi/reset
func Reset(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Get user's session
	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)
	
	userState.RESTAPISimulator.Reset()
	
	state := userState.RESTAPISimulator.GetState()
	
	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// SetupRoutes registers all REST API endpoints
func SetupRoutes(sm *session.Manager) {
	sessionManager = sm
	
	http.HandleFunc("/api/restapi/state", GetState)
	http.HandleFunc("/api/restapi/request", HandleAPIRequest)
	http.HandleFunc("/api/restapi/reset", Reset)
}

