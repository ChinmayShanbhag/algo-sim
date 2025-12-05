package cache

import "time"

// CacheItem represents a single item in the cache for visualization
type CacheItem struct {
	Key        string    `json:"key"`
	Value      string    `json:"value"`
	Frequency  int       `json:"frequency,omitempty"`  // For LFU
	InsertTime time.Time `json:"insertTime,omitempty"` // For FIFO
	AccessTime time.Time `json:"accessTime"`
	Position   int       `json:"position"` // Position in cache (for ordering)
}

// AccessEvent represents a cache access operation
type AccessEvent struct {
	Operation  string    `json:"operation"`  // "GET" or "PUT"
	Key        string    `json:"key"`
	Value      string    `json:"value"`
	Hit        bool      `json:"hit"`        // true if cache hit, false if miss
	EvictedKey string    `json:"evictedKey,omitempty"` // Key that was evicted (if any)
	Timestamp  time.Time `json:"timestamp"`
	CacheSize  int       `json:"cacheSize"`  // Cache size after operation
	CacheItems []string  `json:"cacheItems"` // Current cache keys
}

// CacheState represents the complete state of a cache for visualization
type CacheState struct {
	Algorithm string        `json:"algorithm"`
	Capacity  int           `json:"capacity"`
	Size      int           `json:"size"`
	Items     []CacheItem   `json:"items"`
	History   []AccessEvent `json:"history"`
}

