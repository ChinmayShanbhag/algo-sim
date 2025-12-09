package dns

import "time"

// DNSRecordType represents the type of DNS record
type DNSRecordType string

const (
	RecordTypeA     DNSRecordType = "A"     // IPv4 address
	RecordTypeAAAA  DNSRecordType = "AAAA"  // IPv6 address
	RecordTypeCNAME DNSRecordType = "CNAME" // Canonical name
)

// CacheEntry represents a DNS record in the local cache
type CacheEntry struct {
	Domain     string        `json:"domain"`
	IPAddress  string        `json:"ipAddress"`
	RecordType DNSRecordType `json:"recordType"`
	TTL        int           `json:"ttl"`        // Time to live in seconds
	CachedAt   time.Time     `json:"cachedAt"`   // When it was cached
	ExpiresAt  time.Time     `json:"expiresAt"`  // When it expires
	IsExpired  bool          `json:"isExpired"`  // Whether it's currently expired
}

// DNSServer represents a server in the DNS hierarchy
type DNSServer struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "resolver", "root", "tld", "authoritative"
	Description string `json:"description"`
	Latency     int    `json:"latency"` // Latency in milliseconds
}

// QueryStep represents a single step in the recursive DNS query process
type QueryStep struct {
	StepNumber  int       `json:"stepNumber"`
	FromServer  string    `json:"fromServer"`
	ToServer    string    `json:"toServer"`
	Query       string    `json:"query"`
	Response    string    `json:"response"`
	Latency     int       `json:"latency"`     // Milliseconds for this step
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
}

// ResolutionResult represents the result of a DNS resolution
type ResolutionResult struct {
	Domain         string       `json:"domain"`
	IPAddress      string       `json:"ipAddress"`
	Method         string       `json:"method"`         // "cache" or "recursive"
	Success        bool         `json:"success"`
	TotalLatency   int          `json:"totalLatency"`   // Total time in milliseconds
	NetworkLatency int          `json:"networkLatency"` // Network I/O time
	CacheHit       bool         `json:"cacheHit"`
	CacheExpired   bool         `json:"cacheExpired"`
	Steps          []QueryStep  `json:"steps,omitempty"`
	Error          string       `json:"error,omitempty"`
	Timestamp      time.Time    `json:"timestamp"`
	TTL            int          `json:"ttl,omitempty"` // TTL of the resolved record
}

// DNSSystemState represents the complete state of the DNS simulation
type DNSSystemState struct {
	Cache           map[string]*CacheEntry `json:"cache"`
	Servers         []DNSServer            `json:"servers"`
	QueryHistory    []ResolutionResult     `json:"queryHistory"`
	TotalQueries    int                    `json:"totalQueries"`
	CacheHits       int                    `json:"cacheHits"`
	CacheMisses     int                    `json:"cacheMisses"`
	ExpiredCacheHits int                   `json:"expiredCacheHits"`
	AvgCacheLatency float64                `json:"avgCacheLatency"`
	AvgRecursiveLatency float64            `json:"avgRecursiveLatency"`
}

