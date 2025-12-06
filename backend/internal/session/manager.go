package session

import (
	"sync"
	"time"

	"sds/internal/simulation/bloomfilter"
	"sds/internal/simulation/cache"
	"sds/internal/simulation/cdc"
	"sds/internal/simulation/mapreduce"
	"sds/internal/simulation/pagination"
	"sds/internal/simulation/raft"
	"sds/internal/simulation/rate_limiting"
	"sds/internal/simulation/tcpudp"
	"sds/internal/simulation/three_phase_commit"
	"sds/internal/simulation/two_phase_commit"
)

// State holds all simulation states for a single user session
// Each user gets their own isolated copy of all simulations
type State struct {
	// Raft consensus simulation
	RaftCluster *raft.Cluster

	// Two-Phase Commit simulation
	TwoPCCoordinator  *two_phase_commit.Coordinator
	TwoPCParticipants []*two_phase_commit.Participant
	TwoPCTransaction  *two_phase_commit.Transaction

	// Three-Phase Commit simulation
	ThreePCCoordinator  *three_phase_commit.Coordinator
	ThreePCParticipants []*three_phase_commit.Participant
	ThreePCTransaction  *three_phase_commit.Transaction

	// Rate Limiting simulations (all 5 algorithms)
	FixedWindow    *rate_limiting.FixedWindowCounter
	SlidingLog     *rate_limiting.SlidingLog
	SlidingWindow  *rate_limiting.SlidingWindowCounter
	TokenBucket    *rate_limiting.TokenBucket
	LeakyBucket    *rate_limiting.LeakyBucket

	// Cache Eviction simulations (3 algorithms)
	LRUCache  *cache.LRUCache
	LFUCache  *cache.LFUCache
	FIFOCache *cache.FIFOCache

	// MapReduce simulation
	MapReduceJob *mapreduce.Job

	// CDC (Change Data Capture) simulation
	CDCSystem *cdc.CDCSystem

	// Bloom Filter simulation
	BloomFilter *bloomfilter.BloomFilter

	// TCP/UDP simulation
	TCPUDPSimulator *tcpudp.Simulator

	// Pagination simulation
	PaginationSimulator *pagination.Simulator

	// Metadata
	LastAccessed time.Time
	CreatedAt    time.Time
}

// Manager manages all user sessions
// Uses in-memory storage for now, can be swapped with Redis later
type Manager struct {
	sessions map[string]*State
	mu       sync.RWMutex
}

// NewManager creates a new session manager
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*State),
	}
}

// GetOrCreate retrieves an existing session or creates a new one
// This is the main entry point for all API handlers
func (m *Manager) GetOrCreate(sessionID string) *State {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if session already exists
	if state, exists := m.sessions[sessionID]; exists {
		state.LastAccessed = time.Now()
		return state
	}

	// Create new isolated state for this user
	now := time.Now()
	m.sessions[sessionID] = &State{
		// Initialize Raft cluster with 5 nodes
		RaftCluster: raft.NewCluster(5),

		// Initialize 2PC with 4 participants (original configuration)
		TwoPCCoordinator:  two_phase_commit.NewCoordinator(4),
		TwoPCParticipants: make([]*two_phase_commit.Participant, 4),

		// Initialize 3PC with 4 participants (original configuration)
		ThreePCCoordinator:  three_phase_commit.NewCoordinator(4),
		ThreePCParticipants: make([]*three_phase_commit.Participant, 4),

		// Initialize all rate limiting algorithms (10 requests per 60 seconds)
		FixedWindow:   rate_limiting.NewFixedWindowCounter(10, 60*time.Second),
		SlidingLog:    rate_limiting.NewSlidingLog(10, 60*time.Second),
		SlidingWindow: rate_limiting.NewSlidingWindowCounter(10, 60*time.Second),
		TokenBucket:   rate_limiting.NewTokenBucket(10, 10.0/60.0), // 10 tokens, refill at 1 token per 6 seconds
		LeakyBucket:   rate_limiting.NewLeakyBucket(10, 10.0/60.0), // 10 capacity, process at 1 request per 6 seconds

		// Initialize cache eviction algorithms (capacity of 5 items each)
		LRUCache:  cache.NewLRUCache(5),
		LFUCache:  cache.NewLFUCache(5),
		FIFOCache: cache.NewFIFOCache(5),

		// Initialize MapReduce job with sample word count data
		MapReduceJob: mapreduce.NewJob(
			"word-count-1",
			[]string{
				"hello world hello",
				"world of mapreduce",
				"hello mapreduce world",
				"distributed computing rocks",
			},
			2, // 2 mappers
			2, // 2 reducers
		),

		// Initialize CDC system with sample database
		CDCSystem: cdc.NewCDCSystem(),

		// Initialize Bloom Filter (size: 32 bits, 3 hash functions)
		BloomFilter: bloomfilter.NewBloomFilter(32, 3),

		// Initialize TCP/UDP simulator
		TCPUDPSimulator: tcpudp.NewSimulator(),

		// Initialize Pagination simulator
		PaginationSimulator: pagination.NewSimulator(),

		// Set timestamps
		LastAccessed: now,
		CreatedAt:    now,
	}

	// Initialize 2PC participants
	for i := 0; i < 4; i++ {
		m.sessions[sessionID].TwoPCParticipants[i] = two_phase_commit.NewParticipant(i + 1)
	}

	// Initialize 3PC participants
	for i := 0; i < 4; i++ {
		m.sessions[sessionID].ThreePCParticipants[i] = three_phase_commit.NewParticipant(i + 1)
	}

	return m.sessions[sessionID]
}

// Get retrieves an existing session (returns nil if not found)
func (m *Manager) Get(sessionID string) *State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[sessionID]
}

// Delete removes a session
func (m *Manager) Delete(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
}

// CleanupExpired removes sessions that haven't been accessed recently
// This prevents memory leaks from abandoned sessions
func (m *Manager) CleanupExpired(maxAge time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	deletedCount := 0

	for id, state := range m.sessions {
		if now.Sub(state.LastAccessed) > maxAge {
			delete(m.sessions, id)
			deletedCount++
		}
	}

	return deletedCount
}

// Count returns the number of active sessions
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

