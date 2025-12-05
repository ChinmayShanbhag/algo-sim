package consensus

import (
	"net/http"
	
	"sds/internal/session"
)

var sessionManager *session.Manager

func SetupRoutes(sm *session.Manager) {
	sessionManager = sm
	
	// Raft consensus endpoints
	http.HandleFunc("/api/consensus/raft/state", GetRaftState)
	http.HandleFunc("/api/consensus/raft/election", StartElection)
	http.HandleFunc("/api/consensus/raft/reset", ResetCluster)
	http.HandleFunc("/api/consensus/raft/set-leader", SetLeader)
}

