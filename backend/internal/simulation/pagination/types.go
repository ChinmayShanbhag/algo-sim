package pagination

import "time"

// Item represents a single data item in the list
type Item struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

// PageRequest represents a request for a specific page
type PageRequest struct {
	PageNumber int       `json:"pageNumber"`
	PageSize   int       `json:"pageSize"`
	Timestamp  time.Time `json:"timestamp"`
}

// PageResponse represents the response for a page
type PageResponse struct {
	Items       []Item `json:"items"`
	PageNumber  int    `json:"pageNumber"`
	PageSize    int    `json:"pageSize"`
	TotalItems  int    `json:"totalItems"`
	TotalPages  int    `json:"totalPages"`
	HasNextPage bool   `json:"hasNextPage"`
	HasPrevPage bool   `json:"hasPrevPage"`
}

// PaginationStats tracks pagination metrics
type PaginationStats struct {
	TotalAPICalls    int     `json:"totalApiCalls"`
	TotalItemsFetched int    `json:"totalItemsFetched"`
	AverageLatencyMs float64 `json:"averageLatencyMs"`
	LastPageViewed   int     `json:"lastPageViewed"`
	RecentRequests   []PageRequest `json:"recentRequests"`
}

// VirtualizationStats tracks virtualization metrics
type VirtualizationStats struct {
	TotalAPICalls      int     `json:"totalApiCalls"`
	ItemsInMemory      int     `json:"itemsInMemory"`
	ItemsRendered      int     `json:"itemsRendered"`
	MemoryUsageKB      int     `json:"memoryUsageKB"`
	RenderTimeMs       float64 `json:"renderTimeMs"`
	ScrollPosition     int     `json:"scrollPosition"`
	VisibleRange       string  `json:"visibleRange"`
}

// SimulationState represents the complete state
type SimulationState struct {
	// Dataset
	TotalItems int    `json:"totalItems"`
	Items      []Item `json:"items"`
	
	// Pagination state
	PaginationEnabled    bool             `json:"paginationEnabled"`
	PaginationStats      PaginationStats  `json:"paginationStats"`
	CurrentPage          int              `json:"currentPage"`
	PageSize             int              `json:"pageSize"`
	
	// Virtualization state
	VirtualizationEnabled bool                `json:"virtualizationEnabled"`
	VirtualizationStats   VirtualizationStats `json:"virtualizationStats"`
	AllDataLoaded         bool                `json:"allDataLoaded"`
}

