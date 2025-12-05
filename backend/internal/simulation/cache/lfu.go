package cache

import (
	"sort"
	"sync"
	"time"
)

// LFUCache implements Least Frequently Used cache eviction policy
// When cache is full, evicts the item with lowest access frequency
type LFUCache struct {
	mu        sync.RWMutex
	capacity  int
	cache     map[string]*lfuItem // key -> item
	freqMap   map[int]map[string]bool // frequency -> set of keys
	minFreq   int
	history   []AccessEvent
}

// lfuItem represents an item in the LFU cache
type lfuItem struct {
	key        string
	value      string
	frequency  int
	accessTime time.Time
}

// NewLFUCache creates a new LFU cache with given capacity
func NewLFUCache(capacity int) *LFUCache {
	return &LFUCache{
		capacity: capacity,
		cache:    make(map[string]*lfuItem),
		freqMap:  make(map[int]map[string]bool),
		minFreq:  0,
		history:  []AccessEvent{},
	}
}

// Get retrieves a value from cache and increments its frequency
func (c *LFUCache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.cache[key]
	if !exists {
		c.recordEvent("GET", key, "", false, "")
		return "", false
	}

	// Update frequency
	c.incrementFrequency(item)
	item.accessTime = time.Now()

	c.recordEvent("GET", key, item.value, true, "")
	return item.value, true
}

// Put adds or updates a key-value pair in cache
// If cache is full, evicts the least frequently used item
func (c *LFUCache) Put(key, value string) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if key already exists
	if item, exists := c.cache[key]; exists {
		item.value = value
		c.incrementFrequency(item)
		item.accessTime = time.Now()
		c.recordEvent("PUT", key, value, true, "")
		return ""
	}

	// Check if cache is full
	evictedKey := ""
	if len(c.cache) >= c.capacity {
		// Evict least frequently used
		evictedKey = c.evictLFU()
	}

	// Add new item with frequency 1
	item := &lfuItem{
		key:        key,
		value:      value,
		frequency:  1,
		accessTime: time.Now(),
	}
	c.cache[key] = item

	// Add to frequency map
	if c.freqMap[1] == nil {
		c.freqMap[1] = make(map[string]bool)
	}
	c.freqMap[1][key] = true
	c.minFreq = 1

	c.recordEvent("PUT", key, value, false, evictedKey)
	return evictedKey
}

// incrementFrequency increases the frequency of an item
func (c *LFUCache) incrementFrequency(item *lfuItem) {
	oldFreq := item.frequency
	newFreq := oldFreq + 1

	// Remove from old frequency bucket
	delete(c.freqMap[oldFreq], item.key)
	if len(c.freqMap[oldFreq]) == 0 {
		delete(c.freqMap, oldFreq)
		if c.minFreq == oldFreq {
			c.minFreq = newFreq
		}
	}

	// Add to new frequency bucket
	if c.freqMap[newFreq] == nil {
		c.freqMap[newFreq] = make(map[string]bool)
	}
	c.freqMap[newFreq][item.key] = true
	item.frequency = newFreq
}

// evictLFU removes the least frequently used item
func (c *LFUCache) evictLFU() string {
	// Get any key from the minimum frequency bucket
	keysAtMinFreq := c.freqMap[c.minFreq]
	var keyToEvict string
	var oldestTime time.Time

	// Among items with same frequency, evict the oldest
	for key := range keysAtMinFreq {
		item := c.cache[key]
		if keyToEvict == "" || item.accessTime.Before(oldestTime) {
			keyToEvict = key
			oldestTime = item.accessTime
		}
	}

	// Remove from cache and frequency map
	delete(c.cache, keyToEvict)
	delete(c.freqMap[c.minFreq], keyToEvict)
	if len(c.freqMap[c.minFreq]) == 0 {
		delete(c.freqMap, c.minFreq)
	}

	return keyToEvict
}

// GetState returns current cache state for visualization
func (c *LFUCache) GetState() CacheState {
	c.mu.RLock()
	defer c.mu.RUnlock()

	items := make([]CacheItem, 0, len(c.cache))
	for _, item := range c.cache {
		items = append(items, CacheItem{
			Key:        item.key,
			Value:      item.value,
			Frequency:  item.frequency,
			AccessTime: item.accessTime,
		})
	}

	// Sort items: first by frequency (descending), then by access time (most recent first)
	// This shows most frequently used items at top, with LRU tie-breaking
	sort.Slice(items, func(i, j int) bool {
		if items[i].Frequency != items[j].Frequency {
			return items[i].Frequency > items[j].Frequency // Higher frequency first
		}
		return items[i].AccessTime.After(items[j].AccessTime) // More recent first for ties
	})

	return CacheState{
		Algorithm: "LFU (Least Frequently Used)",
		Capacity:  c.capacity,
		Size:      len(c.cache),
		Items:     items,
		History:   c.getRecentHistory(20),
	}
}

// Reset clears the cache
func (c *LFUCache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*lfuItem)
	c.freqMap = make(map[int]map[string]bool)
	c.minFreq = 0
	c.history = []AccessEvent{}
}

// recordEvent adds an event to history (must be called with lock held)
func (c *LFUCache) recordEvent(operation, key, value string, hit bool, evictedKey string) {
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

// getCurrentKeys returns current cache keys (must be called with lock held)
func (c *LFUCache) getCurrentKeys() []string {
	keys := make([]string, 0, len(c.cache))
	for key := range c.cache {
		keys = append(keys, key)
	}
	return keys
}

// getRecentHistory returns last N events (must be called with lock held)
func (c *LFUCache) getRecentHistory(n int) []AccessEvent {
	start := len(c.history) - n
	if start < 0 {
		start = 0
	}
	return c.history[start:]
}

