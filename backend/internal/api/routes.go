package api

import (
	"net/http"

	"sds/internal/api/atomic_commit"
	"sds/internal/api/bloomfilter"
	"sds/internal/api/cache"
	"sds/internal/api/cdc"
	"sds/internal/api/consensus"
	"sds/internal/api/dns"
	"sds/internal/api/graphql"
	"sds/internal/api/mapreduce"
	"sds/internal/api/pagination"
	"sds/internal/api/rate_limiting"
	"sds/internal/api/restapi"
	"sds/internal/api/tcpudp"
	"sds/internal/session"
)

func SetupRoutes(sessionManager *session.Manager) {
	// Health check endpoint
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Consensus algorithm endpoints
	consensus.SetupRoutes(sessionManager)

	// Atomic commit (2PC and 3PC) endpoints
	atomic_commit.SetupRoutes(sessionManager)

	// Rate limiting endpoints
	rate_limiting.SetupRoutes(sessionManager)

	// Cache eviction endpoints
	cache.SetupRoutes(sessionManager)

	// MapReduce endpoints
	mapreduce.SetupRoutes(sessionManager)

	// CDC (Change Data Capture) endpoints
	cdc.SetupRoutes(sessionManager)

	// Bloom Filter endpoints
	bloomfilter.SetupRoutes(sessionManager)

	// TCP/UDP endpoints
	tcpudp.SetupRoutes(sessionManager)

	// Pagination endpoints
	pagination.SetupRoutes(sessionManager)

	// DNS endpoints
	dns.SetupRoutes(sessionManager)

	// REST API endpoints
	restapi.SetupRoutes(sessionManager)
	
	// GraphQL endpoints
	graphql.SetupRoutes(sessionManager)
}
