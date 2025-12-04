package rate_limiting

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	
	"sds/internal/simulation/rate_limiting"
)

// Global rate limiters
var (
	fixedWindow       *rate_limiting.FixedWindowCounter
	slidingLog        *rate_limiting.SlidingLog
	slidingWindow     *rate_limiting.SlidingWindowCounter
	tokenBucket       *rate_limiting.TokenBucket
	leakyBucket       *rate_limiting.LeakyBucket
)

// init initializes all rate limiters with the same constraints:
// 10 requests per 60 seconds
func init() {
	limit := 10
	windowSize := 60 * time.Second
	
	fixedWindow = rate_limiting.NewFixedWindowCounter(limit, windowSize)
	slidingLog = rate_limiting.NewSlidingLog(limit, windowSize)
	slidingWindow = rate_limiting.NewSlidingWindowCounter(limit, windowSize)
	
	// Token bucket: capacity = limit, refill rate = limit/windowSize
	tokenBucket = rate_limiting.NewTokenBucket(limit, float64(limit)/windowSize.Seconds())
	
	// Leaky bucket: capacity = limit, process rate = limit/windowSize
	leakyBucket = rate_limiting.NewLeakyBucket(limit, float64(limit)/windowSize.Seconds())
}

// setCORSHeaders sets CORS headers for cross-origin requests
func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// GetAllStates returns the current state of all rate limiters
// GET /api/rate-limiting/state
func GetAllStates(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	response := map[string]interface{}{
		"fixedWindow":   fixedWindow.GetState(),
		"slidingLog":    slidingLog.GetState(),
		"slidingWindow": slidingWindow.GetState(),
		"tokenBucket":   tokenBucket.GetState(),
		"leakyBucket":   leakyBucket.GetState(),
	}
	
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// SendRequest sends a single request to all rate limiters
// POST /api/rate-limiting/send-request
func SendRequest(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Send request to all rate limiters
	results := map[string]interface{}{
		"fixedWindow":   fixedWindow.AllowRequest(),
		"slidingLog":    slidingLog.AllowRequest(),
		"slidingWindow": slidingWindow.AllowRequest(),
		"tokenBucket":   tokenBucket.AllowRequest(),
		"leakyBucket":   leakyBucket.AllowRequest(),
	}
	
	// Get updated states
	response := map[string]interface{}{
		"results": results,
		"states": map[string]interface{}{
			"fixedWindow":   fixedWindow.GetState(),
			"slidingLog":    slidingLog.GetState(),
			"slidingWindow": slidingWindow.GetState(),
			"tokenBucket":   tokenBucket.GetState(),
			"leakyBucket":   leakyBucket.GetState(),
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

// SendBurstRequests sends multiple requests at once
// POST /api/rate-limiting/send-burst?count=<number>
func SendBurstRequests(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Get count from query parameter
	countStr := r.URL.Query().Get("count")
	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 {
		count = 20 // Default to 20
	}
	
	// Send multiple requests to all rate limiters
	allResults := make([]map[string]interface{}, count)
	
	for i := 0; i < count; i++ {
		allResults[i] = map[string]interface{}{
			"fixedWindow":   fixedWindow.AllowRequest(),
			"slidingLog":    slidingLog.AllowRequest(),
			"slidingWindow": slidingWindow.AllowRequest(),
			"tokenBucket":   tokenBucket.AllowRequest(),
			"leakyBucket":   leakyBucket.AllowRequest(),
		}
	}
	
	// Get final states
	response := map[string]interface{}{
		"count":   count,
		"results": allResults,
		"states": map[string]interface{}{
			"fixedWindow":   fixedWindow.GetState(),
			"slidingLog":    slidingLog.GetState(),
			"slidingWindow": slidingWindow.GetState(),
			"tokenBucket":   tokenBucket.GetState(),
			"leakyBucket":   leakyBucket.GetState(),
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

// ResetAll resets all rate limiters
// POST /api/rate-limiting/reset
func ResetAll(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	fixedWindow.Reset()
	slidingLog.Reset()
	slidingWindow.Reset()
	tokenBucket.Reset()
	leakyBucket.Reset()
	
	// Return new states
	response := map[string]interface{}{
		"fixedWindow":   fixedWindow.GetState(),
		"slidingLog":    slidingLog.GetState(),
		"slidingWindow": slidingWindow.GetState(),
		"tokenBucket":   tokenBucket.GetState(),
		"leakyBucket":   leakyBucket.GetState(),
	}
	
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// SetupRoutes registers all rate limiting endpoints
func SetupRoutes() {
	http.HandleFunc("/api/rate-limiting/state", GetAllStates)
	http.HandleFunc("/api/rate-limiting/send-request", SendRequest)
	http.HandleFunc("/api/rate-limiting/send-burst", SendBurstRequests)
	http.HandleFunc("/api/rate-limiting/reset", ResetAll)
}

