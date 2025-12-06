package pagination

import (
	"fmt"
	"sync"
	"time"
)

// Simulator simulates pagination vs virtualization
type Simulator struct {
	mu sync.RWMutex
	
	// Dataset - simulating 10,000 items
	totalItems int
	items      []Item
	
	// Pagination state
	paginationStats PaginationStats
	currentPage     int
	pageSize        int
	
	// Virtualization state
	virtualizationStats VirtualizationStats
	allDataLoaded       bool
}

// NewSimulator creates a new pagination simulator
func NewSimulator() *Simulator {
	totalItems := 10000
	pageSize := 20
	
	// Generate sample dataset
	items := make([]Item, totalItems)
	for i := 0; i < totalItems; i++ {
		items[i] = Item{
			ID:          i + 1,
			Title:       fmt.Sprintf("Item %d", i+1),
			Description: fmt.Sprintf("Description for item %d with some sample text to simulate real data", i+1),
			Timestamp:   time.Now().Add(time.Duration(-i) * time.Hour),
		}
	}
	
	return &Simulator{
		totalItems: totalItems,
		items:      items,
		currentPage: 1,
		pageSize:    pageSize,
		paginationStats: PaginationStats{
			RecentRequests: []PageRequest{},
		},
		virtualizationStats: VirtualizationStats{
			ItemsRendered: 0,
		},
		allDataLoaded: false,
	}
}

// GetPage returns a specific page of items (for pagination)
func (s *Simulator) GetPage(pageNumber int) PageResponse {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Simulate API call
	s.paginationStats.TotalAPICalls++
	s.currentPage = pageNumber
	
	// Calculate pagination
	totalPages := (s.totalItems + s.pageSize - 1) / s.pageSize
	start := (pageNumber - 1) * s.pageSize
	end := start + s.pageSize
	
	if start < 0 {
		start = 0
	}
	if end > s.totalItems {
		end = s.totalItems
	}
	
	pageItems := s.items[start:end]
	s.paginationStats.TotalItemsFetched += len(pageItems)
	s.paginationStats.LastPageViewed = pageNumber
	
	// Track recent request
	request := PageRequest{
		PageNumber: pageNumber,
		PageSize:   s.pageSize,
		Timestamp:  time.Now(),
	}
	s.paginationStats.RecentRequests = append(s.paginationStats.RecentRequests, request)
	
	// Keep only last 10 requests
	if len(s.paginationStats.RecentRequests) > 10 {
		s.paginationStats.RecentRequests = s.paginationStats.RecentRequests[len(s.paginationStats.RecentRequests)-10:]
	}
	
	// Simulate latency (50-150ms)
	s.paginationStats.AverageLatencyMs = 100
	
	return PageResponse{
		Items:       pageItems,
		PageNumber:  pageNumber,
		PageSize:    s.pageSize,
		TotalItems:  s.totalItems,
		TotalPages:  totalPages,
		HasNextPage: pageNumber < totalPages,
		HasPrevPage: pageNumber > 1,
	}
}

// LoadAllData simulates loading all data at once (for virtualization initial load)
func (s *Simulator) LoadAllData() []Item {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.allDataLoaded = true
	s.virtualizationStats.TotalAPICalls = 1 // Only one API call
	s.virtualizationStats.ItemsInMemory = s.totalItems
	
	// Simulate memory usage: ~0.5KB per item (realistic for JSON data)
	s.virtualizationStats.MemoryUsageKB = s.totalItems / 2
	
	// Simulate initial render time (proportional to items)
	// Without virtualization, rendering 10k items would be slow
	s.virtualizationStats.RenderTimeMs = float64(s.totalItems) * 0.1 // 1000ms for 10k items
	
	return s.items
}

// UpdateVirtualizationView simulates updating the rendered viewport
func (s *Simulator) UpdateVirtualizationView(scrollPosition int, viewportHeight int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Calculate visible items (74px per item: 30px content + 24px padding + 12px margin + 8px buffer)
	itemHeight := 74 // Match frontend itemHeight
	visibleItems := (viewportHeight + itemHeight - 1) / itemHeight // Ceiling division
	startIndex := scrollPosition / itemHeight
	endIndex := startIndex + visibleItems
	
	if endIndex > s.totalItems {
		endIndex = s.totalItems
	}
	
	// Ensure we show at least the range that's actually visible
	actualStartIndex := startIndex + 1 // 1-indexed
	actualEndIndex := endIndex
	if actualEndIndex < actualStartIndex {
		actualEndIndex = actualStartIndex
	}
	
	s.virtualizationStats.ScrollPosition = scrollPosition
	s.virtualizationStats.ItemsRendered = endIndex - startIndex
	s.virtualizationStats.VisibleRange = fmt.Sprintf("%d-%d", actualStartIndex, actualEndIndex)
	
	// With virtualization, render time is constant (only renders visible items)
	s.virtualizationStats.RenderTimeMs = float64(s.virtualizationStats.ItemsRendered) * 0.1 // ~0.5-0.8ms for 5-8 visible items
}

// GetState returns the current simulation state
func (s *Simulator) GetState() SimulationState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return SimulationState{
		TotalItems:            s.totalItems,
		Items:                 s.items, // Only included for initial virtualization load
		PaginationEnabled:     true,
		PaginationStats:       s.paginationStats,
		CurrentPage:           s.currentPage,
		PageSize:              s.pageSize,
		VirtualizationEnabled: true,
		VirtualizationStats:   s.virtualizationStats,
		AllDataLoaded:         s.allDataLoaded,
	}
}

// Reset resets the simulation
func (s *Simulator) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.currentPage = 1
	s.paginationStats = PaginationStats{
		RecentRequests: []PageRequest{},
	}
	s.virtualizationStats = VirtualizationStats{
		ItemsRendered: 0,
	}
	s.allDataLoaded = false
}

