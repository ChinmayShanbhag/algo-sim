package rate_limiting

import (
	"sync"
	"time"
)

// RateLimiter is the interface that all rate limiting algorithms must implement
type RateLimiter interface {
	// AllowRequest checks if a request should be allowed
	// Returns true if allowed, false if rate limited
	AllowRequest() bool
	
	// GetState returns the current state of the rate limiter for visualization
	GetState() interface{}
	
	// Reset resets the rate limiter to initial state
	Reset()
	
	// GetName returns the name of the algorithm
	GetName() string
}

// RequestLog represents a single request for logging-based algorithms
type RequestLog struct {
	Timestamp time.Time `json:"timestamp"`
	Allowed   bool      `json:"allowed"`
}

// FixedWindowCounter implements rate limiting with fixed time windows
// Simple but can allow 2x limit at window boundaries
type FixedWindowCounter struct {
	mu            sync.RWMutex
	limit         int       // Max requests per window
	windowSize    time.Duration
	counter       int       // Current count in window
	windowStart   time.Time // Start of current window
	requestHistory []RequestLog `json:"requestHistory"`
}

// NewFixedWindowCounter creates a new fixed window counter rate limiter
func NewFixedWindowCounter(limit int, windowSize time.Duration) *FixedWindowCounter {
	return &FixedWindowCounter{
		limit:         limit,
		windowSize:    windowSize,
		counter:       0,
		windowStart:   time.Now(),
		requestHistory: []RequestLog{},
	}
}

func (f *FixedWindowCounter) AllowRequest() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	now := time.Now()
	
	// Check if we need to reset the window
	if now.Sub(f.windowStart) >= f.windowSize {
		f.counter = 0
		f.windowStart = now
	}
	
	allowed := f.counter < f.limit
	
	if allowed {
		f.counter++
	}
	
	// Record request
	f.requestHistory = append(f.requestHistory, RequestLog{
		Timestamp: now,
		Allowed:   allowed,
	})
	
	return allowed
}

func (f *FixedWindowCounter) GetState() interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	return map[string]interface{}{
		"algorithm":      "Fixed Window Counter",
		"limit":          f.limit,
		"windowSize":     f.windowSize.Seconds(),
		"currentCount":   f.counter,
		"windowStart":    f.windowStart,
		"windowEnd":      f.windowStart.Add(f.windowSize),
		"requestHistory": f.requestHistory,
	}
}

func (f *FixedWindowCounter) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	f.counter = 0
	f.windowStart = time.Now()
	f.requestHistory = []RequestLog{}
}

func (f *FixedWindowCounter) GetName() string {
	return "Fixed Window Counter"
}

// SlidingLog implements rate limiting by keeping a log of all requests
// Accurate but memory-intensive
type SlidingLog struct {
	mu            sync.RWMutex
	limit         int
	windowSize    time.Duration
	requestLog    []time.Time // Timestamps of allowed requests
	requestHistory []RequestLog
}

func NewSlidingLog(limit int, windowSize time.Duration) *SlidingLog {
	return &SlidingLog{
		limit:         limit,
		windowSize:    windowSize,
		requestLog:    []time.Time{},
		requestHistory: []RequestLog{},
	}
}

func (s *SlidingLog) AllowRequest() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	windowStart := now.Add(-s.windowSize)
	
	// Remove old requests outside the window
	validRequests := []time.Time{}
	for _, timestamp := range s.requestLog {
		if timestamp.After(windowStart) {
			validRequests = append(validRequests, timestamp)
		}
	}
	s.requestLog = validRequests
	
	allowed := len(s.requestLog) < s.limit
	
	if allowed {
		s.requestLog = append(s.requestLog, now)
	}
	
	// Record request
	s.requestHistory = append(s.requestHistory, RequestLog{
		Timestamp: now,
		Allowed:   allowed,
	})
	
	return allowed
}

func (s *SlidingLog) GetState() interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	now := time.Now()
	windowStart := now.Add(-s.windowSize)
	
	return map[string]interface{}{
		"algorithm":      "Sliding Log",
		"limit":          s.limit,
		"windowSize":     s.windowSize.Seconds(),
		"currentCount":   len(s.requestLog),
		"windowStart":    windowStart,
		"windowEnd":      now,
		"requestLog":     s.requestLog,
		"requestHistory": s.requestHistory,
	}
}

func (s *SlidingLog) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.requestLog = []time.Time{}
	s.requestHistory = []RequestLog{}
}

func (s *SlidingLog) GetName() string {
	return "Sliding Log"
}

// SlidingWindowCounter combines fixed window and sliding log approaches
// More memory efficient than sliding log, more accurate than fixed window
type SlidingWindowCounter struct {
	mu            sync.RWMutex
	limit         int
	windowSize    time.Duration
	prevCounter   int       // Count from previous window
	currCounter   int       // Count in current window
	currWindowStart time.Time
	requestHistory []RequestLog
}

func NewSlidingWindowCounter(limit int, windowSize time.Duration) *SlidingWindowCounter {
	return &SlidingWindowCounter{
		limit:          limit,
		windowSize:     windowSize,
		prevCounter:    0,
		currCounter:    0,
		currWindowStart: time.Now(),
		requestHistory: []RequestLog{},
	}
}

func (sw *SlidingWindowCounter) AllowRequest() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	
	now := time.Now()
	
	// Check if we need to move to next window
	if now.Sub(sw.currWindowStart) >= sw.windowSize {
		sw.prevCounter = sw.currCounter
		sw.currCounter = 0
		sw.currWindowStart = now
	}
	
	// Calculate weighted count
	elapsed := now.Sub(sw.currWindowStart)
	prevWeight := 1.0 - (float64(elapsed) / float64(sw.windowSize))
	estimatedCount := float64(sw.prevCounter)*prevWeight + float64(sw.currCounter)
	
	allowed := estimatedCount < float64(sw.limit)
	
	if allowed {
		sw.currCounter++
	}
	
	// Record request
	sw.requestHistory = append(sw.requestHistory, RequestLog{
		Timestamp: now,
		Allowed:   allowed,
	})
	
	return allowed
}

func (sw *SlidingWindowCounter) GetState() interface{} {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	
	now := time.Now()
	elapsed := now.Sub(sw.currWindowStart)
	prevWeight := 1.0 - (float64(elapsed) / float64(sw.windowSize))
	estimatedCount := float64(sw.prevCounter)*prevWeight + float64(sw.currCounter)
	
	return map[string]interface{}{
		"algorithm":      "Sliding Window Counter",
		"limit":          sw.limit,
		"windowSize":     sw.windowSize.Seconds(),
		"prevCount":      sw.prevCounter,
		"currentCount":   sw.currCounter,
		"estimatedCount": estimatedCount,
		"windowStart":    sw.currWindowStart,
		"requestHistory": sw.requestHistory,
	}
}

func (sw *SlidingWindowCounter) Reset() {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	
	sw.prevCounter = 0
	sw.currCounter = 0
	sw.currWindowStart = time.Now()
	sw.requestHistory = []RequestLog{}
}

func (sw *SlidingWindowCounter) GetName() string {
	return "Sliding Window Counter"
}

// TokenBucket allows bursts of traffic up to bucket capacity
// Tokens are added at a fixed rate
type TokenBucket struct {
	mu            sync.RWMutex
	capacity      int       // Max tokens in bucket
	refillRate    float64   // Tokens per second
	tokens        float64   // Current tokens
	lastRefill    time.Time
	requestHistory []RequestLog
}

func NewTokenBucket(capacity int, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:      capacity,
		refillRate:    refillRate,
		tokens:        float64(capacity), // Start with full bucket
		lastRefill:    time.Now(),
		requestHistory: []RequestLog{},
	}
}

func (tb *TokenBucket) AllowRequest() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	
	now := time.Now()
	
	// Refill tokens based on time elapsed
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens = min(float64(tb.capacity), tb.tokens+(elapsed*tb.refillRate))
	tb.lastRefill = now
	
	allowed := tb.tokens >= 1.0
	
	if allowed {
		tb.tokens -= 1.0
	}
	
	// Record request
	tb.requestHistory = append(tb.requestHistory, RequestLog{
		Timestamp: now,
		Allowed:   allowed,
	})
	
	return allowed
}

func (tb *TokenBucket) GetState() interface{} {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	
	return map[string]interface{}{
		"algorithm":      "Token Bucket",
		"capacity":       tb.capacity,
		"refillRate":     tb.refillRate,
		"currentTokens":  tb.tokens,
		"lastRefill":     tb.lastRefill,
		"requestHistory": tb.requestHistory,
	}
}

func (tb *TokenBucket) Reset() {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	
	tb.tokens = float64(tb.capacity)
	tb.lastRefill = time.Now()
	tb.requestHistory = []RequestLog{}
}

func (tb *TokenBucket) GetName() string {
	return "Token Bucket"
}

// LeakyBucket processes requests at a fixed rate
// Excess requests are queued (or rejected if queue is full)
type LeakyBucket struct {
	mu            sync.RWMutex
	capacity      int       // Max queue size
	processRate   float64   // Requests processed per second
	queue         int       // Current queue size
	lastProcess   time.Time
	requestHistory []RequestLog
}

func NewLeakyBucket(capacity int, processRate float64) *LeakyBucket {
	return &LeakyBucket{
		capacity:      capacity,
		processRate:   processRate,
		queue:         0,
		lastProcess:   time.Now(),
		requestHistory: []RequestLog{},
	}
}

func (lb *LeakyBucket) AllowRequest() bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	now := time.Now()
	
	// Process (leak) requests based on time elapsed
	elapsed := now.Sub(lb.lastProcess).Seconds()
	processed := int(elapsed * lb.processRate)
	lb.queue = max(0, lb.queue-processed)
	lb.lastProcess = now
	
	allowed := lb.queue < lb.capacity
	
	if allowed {
		lb.queue++
	}
	
	// Record request
	lb.requestHistory = append(lb.requestHistory, RequestLog{
		Timestamp: now,
		Allowed:   allowed,
	})
	
	return allowed
}

func (lb *LeakyBucket) GetState() interface{} {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	
	return map[string]interface{}{
		"algorithm":      "Leaky Bucket",
		"capacity":       lb.capacity,
		"processRate":    lb.processRate,
		"currentQueue":   lb.queue,
		"lastProcess":    lb.lastProcess,
		"requestHistory": lb.requestHistory,
	}
}

func (lb *LeakyBucket) Reset() {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	lb.queue = 0
	lb.lastProcess = time.Now()
	lb.requestHistory = []RequestLog{}
}

func (lb *LeakyBucket) GetName() string {
	return "Leaky Bucket"
}

// Helper functions
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

