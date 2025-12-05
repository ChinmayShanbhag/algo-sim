package cache

import (
	"container/list"
	"sync"
	"time"
)

// LRUCache implements Least Recently Used cache eviction policy
// When cache is full, evicts the least recently accessed item
type LRUCache struct {
	mu       sync.RWMutex
	capacity int
	cache    map[string]*list.Element // key -> list element
	lruList  *list.List               // doubly linked list for LRU ordering
	history  []AccessEvent            // History of all operations
}

// cacheItem represents an item in the LRU cache
type cacheItem struct {
	key        string
	value      string
	accessTime time.Time
}

// NewLRUCache creates a new LRU cache with given capacity
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		lruList:  list.New(),
		history:  []AccessEvent{},
	}
}

// Get retrieves a value from cache and marks it as recently used
// Returns (value, hit) where hit indicates if key was found
func (c *LRUCache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	element, exists := c.cache[key]
	if !exists {
		c.recordEvent("GET", key, "", false, "")
		return "", false
	}

	// Move to front (most recently used)
	c.lruList.MoveToFront(element)
	item := element.Value.(*cacheItem)
	item.accessTime = time.Now()

	c.recordEvent("GET", key, item.value, true, "")
	return item.value, true
}

// Put adds or updates a key-value pair in cache
// If cache is full, evicts the least recently used item
func (c *LRUCache) Put(key, value string) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if key already exists
	if element, exists := c.cache[key]; exists {
		// Update existing item and move to front
		c.lruList.MoveToFront(element)
		item := element.Value.(*cacheItem)
		item.value = value
		item.accessTime = time.Now()
		c.recordEvent("PUT", key, value, true, "")
		return ""
	}

	// Check if cache is full
	evictedKey := ""
	if len(c.cache) >= c.capacity {
		// Evict least recently used (back of list)
		oldest := c.lruList.Back()
		if oldest != nil {
			item := oldest.Value.(*cacheItem)
			evictedKey = item.key
			delete(c.cache, item.key)
			c.lruList.Remove(oldest)
		}
	}

	// Add new item to front (most recently used)
	item := &cacheItem{
		key:        key,
		value:      value,
		accessTime: time.Now(),
	}
	element := c.lruList.PushFront(item)
	c.cache[key] = element

	c.recordEvent("PUT", key, value, false, evictedKey)
	return evictedKey
}

// GetState returns current cache state for visualization
func (c *LRUCache) GetState() CacheState {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Items are already in LRU order (front = most recent, back = least recent)
	items := make([]CacheItem, 0, c.lruList.Len())
	position := 0
	for e := c.lruList.Front(); e != nil; e = e.Next() {
		item := e.Value.(*cacheItem)
		items = append(items, CacheItem{
			Key:        item.key,
			Value:      item.value,
			AccessTime: item.accessTime,
			Position:   position,
		})
		position++
	}

	return CacheState{
		Algorithm: "LRU (Least Recently Used)",
		Capacity:  c.capacity,
		Size:      len(c.cache),
		Items:     items,
		History:   c.getRecentHistory(20),
	}
}

// Reset clears the cache
func (c *LRUCache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*list.Element)
	c.lruList = list.New()
	c.history = []AccessEvent{}
}

// recordEvent adds an event to history (must be called with lock held)
func (c *LRUCache) recordEvent(operation, key, value string, hit bool, evictedKey string) {
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
func (c *LRUCache) getCurrentKeys() []string {
	keys := make([]string, 0, len(c.cache))
	for e := c.lruList.Front(); e != nil; e = e.Next() {
		item := e.Value.(*cacheItem)
		keys = append(keys, item.key)
	}
	return keys
}

// getRecentHistory returns last N events (must be called with lock held)
func (c *LRUCache) getRecentHistory(n int) []AccessEvent {
	start := len(c.history) - n
	if start < 0 {
		start = 0
	}
	return c.history[start:]
}

