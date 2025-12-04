package api

import (
	"net/http"
	
	"sds/internal/api/atomic_commit"
	"sds/internal/api/consensus"
	"sds/internal/api/rate_limiting"
)

func SetupRoutes() {
	// Health check endpoint
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	// Consensus algorithm endpoints
	consensus.SetupRoutes()
	
	// Atomic commit (2PC and 3PC) endpoints
	atomic_commit.SetupRoutes()
	
	// Rate limiting endpoints
	rate_limiting.SetupRoutes()
}

