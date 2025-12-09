package dns

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// DNSSimulator simulates DNS resolution with cache and recursive queries
type DNSSimulator struct {
	mu                    sync.RWMutex
	cache                 map[string]*CacheEntry
	servers               []DNSServer
	queryHistory          []ResolutionResult
	totalQueries          int
	cacheHits             int
	cacheMisses           int
	expiredCacheHits      int
	totalCacheLatency     int
	totalRecursiveLatency int
}

// NewSimulator creates a new DNS simulator with default configuration
func NewSimulator() *DNSSimulator {
	sim := &DNSSimulator{
		cache:        make(map[string]*CacheEntry),
		queryHistory: []ResolutionResult{},
		servers: []DNSServer{
			{
				Name:        "Local Resolver",
				Type:        "resolver",
				Description: "Your ISP's DNS resolver",
				Latency:     5, // 5ms to local resolver
			},
			{
				Name:        "Root Server",
				Type:        "root",
				Description: "One of 13 root name servers",
				Latency:     50, // 50ms to root server
			},
			{
				Name:        "TLD Server",
				Type:        "tld",
				Description: "Top-Level Domain server (.com, .org, etc.)",
				Latency:     40, // 40ms to TLD server
			},
			{
				Name:        "Authoritative Server",
				Type:        "authoritative",
				Description: "Domain's authoritative name server",
				Latency:     45, // 45ms to authoritative server
			},
		},
	}

	// Pre-populate cache with some common domains
	sim.seedCache()

	return sim
}

// seedCache pre-populates the cache with common domains using real DNS resolution
func (s *DNSSimulator) seedCache() {
	now := time.Now()

	domains := []struct {
		domain string
		ttl    int
		age    time.Duration // How long ago it was cached
	}{
		{"google.com", 300, 30 * time.Second},
		{"github.com", 60, 25 * time.Second}, // About to expire
		{"amazon.com", 600, 10 * time.Second},
		{"facebook.com", 3600, 1800 * time.Second}, // Long TTL
		{"cloudflare.com", 30, 60 * time.Second},   // Already expired
	}

	for _, d := range domains {
		// Resolve the domain to get real IP
		ip := s.resolveRealDNS(d.domain)

		cachedAt := now.Add(-d.age)
		expiresAt := cachedAt.Add(time.Duration(d.ttl) * time.Second)

		s.cache[d.domain] = &CacheEntry{
			Domain:     d.domain,
			IPAddress:  ip,
			RecordType: RecordTypeA,
			TTL:        d.ttl,
			CachedAt:   cachedAt,
			ExpiresAt:  expiresAt,
			IsExpired:  now.After(expiresAt),
		}
	}
}

// resolveRealDNS performs actual DNS resolution to get real IP addresses
func (s *DNSSimulator) resolveRealDNS(domain string) string {
	// Use Go's net.LookupIP to get real DNS resolution
	ips, err := net.LookupIP(domain)
	if err != nil || len(ips) == 0 {
		// Fallback to a placeholder if resolution fails
		return "0.0.0.0"
	}

	// Return the first IPv4 address found
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4.String()
		}
	}

	// If no IPv4, return first IP as string
	return ips[0].String()
}

// ResolveDomain resolves a domain name, using cache if available and valid
func (s *DNSSimulator) ResolveDomain(domain string) ResolutionResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	s.totalQueries++

	// Check cache first
	if entry, exists := s.cache[domain]; exists {
		// Check if cache entry is expired
		if now.After(entry.ExpiresAt) {
			entry.IsExpired = true
			s.expiredCacheHits++

			// Cache expired - must do recursive query
			result := s.performRecursiveQuery(domain, true)
			result.CacheExpired = true
			result.Timestamp = now

			// Update cache with fresh data
			s.updateCache(domain, result.IPAddress, 300) // Default 5 min TTL

			s.queryHistory = append(s.queryHistory, result)
			return result
		}

		// Cache hit with valid entry
		s.cacheHits++
		s.totalCacheLatency += 1 // 1ms for cache lookup

		result := ResolutionResult{
			Domain:         domain,
			IPAddress:      entry.IPAddress,
			Method:         "cache",
			Success:        true,
			TotalLatency:   1, // Near-instant cache lookup
			NetworkLatency: 0, // No network I/O
			CacheHit:       true,
			CacheExpired:   false,
			Steps:          []QueryStep{},
			Timestamp:      now,
			TTL:            int(entry.ExpiresAt.Sub(now).Seconds()),
		}

		s.queryHistory = append(s.queryHistory, result)
		return result
	}

	// Cache miss - perform recursive query
	s.cacheMisses++
	result := s.performRecursiveQuery(domain, false)
	result.Timestamp = now

	// Cache the result
	s.updateCache(domain, result.IPAddress, result.TTL)

	s.queryHistory = append(s.queryHistory, result)
	return result
}

// performRecursiveQuery simulates a full recursive DNS query
func (s *DNSSimulator) performRecursiveQuery(domain string, cacheExpired bool) ResolutionResult {
	steps := []QueryStep{}
	totalLatency := 0
	stepNum := 1
	now := time.Now()

	// Step 1: Client -> Local Resolver
	resolverLatency := s.servers[0].Latency
	steps = append(steps, QueryStep{
		StepNumber:  stepNum,
		FromServer:  "Client",
		ToServer:    s.servers[0].Name,
		Query:       fmt.Sprintf("What is the IP for %s?", domain),
		Response:    "Let me find out for you...",
		Latency:     resolverLatency,
		Timestamp:   now.Add(time.Duration(totalLatency) * time.Millisecond),
		Description: "Client sends query to local DNS resolver",
	})
	totalLatency += resolverLatency
	stepNum++

	// Step 2: Resolver -> Root Server
	rootLatency := s.servers[1].Latency
	steps = append(steps, QueryStep{
		StepNumber:  stepNum,
		FromServer:  s.servers[0].Name,
		ToServer:    s.servers[1].Name,
		Query:       fmt.Sprintf("Where can I find %s?", domain),
		Response:    "Try the .com TLD server",
		Latency:     rootLatency,
		Timestamp:   now.Add(time.Duration(totalLatency) * time.Millisecond),
		Description: "Resolver queries root server for TLD information",
	})
	totalLatency += rootLatency
	stepNum++

	// Step 3: Resolver -> TLD Server
	tldLatency := s.servers[2].Latency
	steps = append(steps, QueryStep{
		StepNumber:  stepNum,
		FromServer:  s.servers[0].Name,
		ToServer:    s.servers[2].Name,
		Query:       fmt.Sprintf("Where is the authoritative server for %s?", domain),
		Response:    "Try ns1.example.com",
		Latency:     tldLatency,
		Timestamp:   now.Add(time.Duration(totalLatency) * time.Millisecond),
		Description: "Resolver queries TLD server for authoritative nameserver",
	})
	totalLatency += tldLatency
	stepNum++

	// Step 4: Resolver -> Authoritative Server
	authLatency := s.servers[3].Latency
	ip := s.generateIPAddress(domain)
	steps = append(steps, QueryStep{
		StepNumber:  stepNum,
		FromServer:  s.servers[0].Name,
		ToServer:    s.servers[3].Name,
		Query:       fmt.Sprintf("What is the IP address for %s?", domain),
		Response:    fmt.Sprintf("IP: %s (TTL: 300s)", ip),
		Latency:     authLatency,
		Timestamp:   now.Add(time.Duration(totalLatency) * time.Millisecond),
		Description: "Resolver queries authoritative server for final answer",
	})
	totalLatency += authLatency
	stepNum++

	// Step 5: Resolver -> Client
	finalLatency := 5 // 5ms to return to client
	steps = append(steps, QueryStep{
		StepNumber:  stepNum,
		FromServer:  s.servers[0].Name,
		ToServer:    "Client",
		Query:       "",
		Response:    fmt.Sprintf("%s resolves to %s", domain, ip),
		Latency:     finalLatency,
		Timestamp:   now.Add(time.Duration(totalLatency) * time.Millisecond),
		Description: "Resolver returns final answer to client",
	})
	totalLatency += finalLatency

	s.totalRecursiveLatency += totalLatency

	return ResolutionResult{
		Domain:         domain,
		IPAddress:      ip,
		Method:         "recursive",
		Success:        true,
		TotalLatency:   totalLatency,
		NetworkLatency: totalLatency, // All time is network I/O
		CacheHit:       false,
		CacheExpired:   cacheExpired,
		Steps:          steps,
		TTL:            300, // Default 5 minute TTL
	}
}

// generateIPAddress performs real DNS resolution to get authentic IP addresses
func (s *DNSSimulator) generateIPAddress(domain string) string {
	return s.resolveRealDNS(domain)
}

// updateCache updates or adds an entry to the cache
func (s *DNSSimulator) updateCache(domain, ip string, ttl int) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(ttl) * time.Second)

	s.cache[domain] = &CacheEntry{
		Domain:     domain,
		IPAddress:  ip,
		RecordType: RecordTypeA,
		TTL:        ttl,
		CachedAt:   now,
		ExpiresAt:  expiresAt,
		IsExpired:  false,
	}
}

// ClearCache clears all entries from the DNS cache
func (s *DNSSimulator) ClearCache() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache = make(map[string]*CacheEntry)
}

// GetState returns the current state of the DNS system
func (s *DNSSimulator) GetState() DNSSystemState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Update expired status for all cache entries
	now := time.Now()
	for _, entry := range s.cache {
		entry.IsExpired = now.After(entry.ExpiresAt)
	}

	// Calculate averages
	avgCacheLatency := 0.0
	if s.cacheHits > 0 {
		avgCacheLatency = float64(s.totalCacheLatency) / float64(s.cacheHits)
	}

	avgRecursiveLatency := 0.0
	recursiveQueries := s.cacheMisses + s.expiredCacheHits
	if recursiveQueries > 0 {
		avgRecursiveLatency = float64(s.totalRecursiveLatency) / float64(recursiveQueries)
	}

	return DNSSystemState{
		Cache:               s.cache,
		Servers:             s.servers,
		QueryHistory:        s.queryHistory,
		TotalQueries:        s.totalQueries,
		CacheHits:           s.cacheHits,
		CacheMisses:         s.cacheMisses,
		ExpiredCacheHits:    s.expiredCacheHits,
		AvgCacheLatency:     avgCacheLatency,
		AvgRecursiveLatency: avgRecursiveLatency,
	}
}

// Reset resets the simulator to initial state
func (s *DNSSimulator) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache = make(map[string]*CacheEntry)
	s.queryHistory = []ResolutionResult{}
	s.totalQueries = 0
	s.cacheHits = 0
	s.cacheMisses = 0
	s.expiredCacheHits = 0
	s.totalCacheLatency = 0
	s.totalRecursiveLatency = 0

	// Re-seed cache
	s.seedCache()
}
