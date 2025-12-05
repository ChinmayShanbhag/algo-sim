package cache

import (
	"sync"
	"time"
)

// FIFOCache implements First In First Out cache eviction policy
// When cache is full, evicts the oldest inserted item (regardless of usage)
type FIFOCache struct {
	mu       sync.RWMutex
	capacity int
	cache    map[string]*fifoItem // key -> item
	queue    []string             // insertion order queue
	history  []AccessEvent
}

// fifoItem represents an item in the FIFO cache
type fifoItem struct {
	key        string
	value      string
	insertTime time.Time
	accessTime time.Time
}

// NewFIFOCache creates a new FIFO cache with given capacity
func NewFIFOCache(capacity int) *FIFOCache {
	return &FIFOCache{
		capacity: capacity,
		cache:    make(map[string]*fifoItem),
		queue:    make([]string, 0, capacity),
		history:  []AccessEvent{},
	}
}

// Get retrieves a value from cache (does NOT affect eviction order)
func (c *FIFOCache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.cache[key]
	if !exists {
		c.recordEvent("GET", key, "", false, "")
		return "", false
	}

	item.accessTime = time.Now()
	c.recordEvent("GET", key, item.value, true, "")
	return item.value, true
}

// Put adds or updates a key-value pair in cache
// If cache is full, evicts the first inserted item
func (c *FIFOCache) Put(key, value string) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if key already exists
	if item, exists := c.cache[key]; exists {
		item.value = value
		item.accessTime = time.Now()
		c.recordEvent("PUT", key, value, true, "")
		return ""
	}

	// Check if cache is full
	evictedKey := ""
	if len(c.cache) >= c.capacity {
		// Evict first item in queue (FIFO)
		evictedKey = c.queue[0]
		c.queue = c.queue[1:]
		delete(c.cache, evictedKey)
	}

	// Add new item
	now := time.Now()
	item := &fifoItem{
		key:        key,
		value:      value,
		insertTime: now,
		accessTime: now,
	}
	c.cache[key] = item
	c.queue = append(c.queue, key)

	c.recordEvent("PUT", key, value, false, evictedKey)
	return evictedKey
}

// GetState returns current cache state for visualization
func (c *FIFOCache) GetState() CacheState {
	c.mu.RLock()
	defer c.mu.RUnlock()

	items := make([]CacheItem, 0, len(c.queue))
	for position, key := range c.queue {
		item := c.cache[key]
		items = append(items, CacheItem{
			Key:        item.key,
			Value:      item.value,
			InsertTime: item.insertTime,
			AccessTime: item.accessTime,
			Position:   position,
		})
	}

	return CacheState{
		Algorithm: "FIFO (First In First Out)",
		Capacity:  c.capacity,
		Size:      len(c.cache),
		Items:     items,
		History:   c.getRecentHistory(20),
	}
}

// Reset clears the cache
func (c *FIFOCache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*fifoItem)
	c.queue = make([]string, 0, c.capacity)
	c.history = []AccessEvent{}
}

// recordEvent adds an event to history (must be called with lock held)
func (c *FIFOCache) recordEvent(operation, key, value string, hit bool, evictedKey string) {
	event := AccessEvent{
		Operation:   operation,
		Key:         key,
		Value:       value,
		Hit:         hit,
		EvictedKey:  evictedKey,
		Timestamp:   time.Now(),
		CacheSize:   len(c.cache),
		CacheItems:  c.getCurrentKeys(),
	}
	c.history = append(c.history, event)
}

// getCurrentKeys returns current cache keys in FIFO order (must be called with lock held)
func (c *FIFOCache) getCurrentKeys() []string {
	keys := make([]string, len(c.queue))
	copy(keys, c.queue)
	return keys
}

// getRecentHistory returns last N events (must be called with lock held)
func (c *FIFOCache) getRecentHistory(n int) []AccessEvent {
	start := len(c.history) - n
	if start < 0 {
		start = 0
	}
	return c.history[start:]
}

