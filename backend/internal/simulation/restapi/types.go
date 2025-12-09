package restapi

import "time"

// HTTPMethod represents HTTP request methods
type HTTPMethod string

const (
	GET    HTTPMethod = "GET"
	POST   HTTPMethod = "POST"
	PUT    HTTPMethod = "PUT"
	PATCH  HTTPMethod = "PATCH"
	DELETE HTTPMethod = "DELETE"
)

// Resource represents a REST resource in the mock API
type Resource struct {
	ID        int                    `json:"id"`
	Type      string                 `json:"type"` // "user", "post", "comment", etc.
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
}

// APIRequest represents an HTTP request made to the REST API
type APIRequest struct {
	Method  HTTPMethod             `json:"method"`
	Path    string                 `json:"path"`
	Headers map[string]string      `json:"headers"`
	Body    map[string]interface{} `json:"body,omitempty"`
	Query   map[string]string      `json:"query,omitempty"`
}

// APIResponse represents an HTTP response from the REST API
type APIResponse struct {
	StatusCode int                    `json:"statusCode"`
	StatusText string                 `json:"statusText"`
	Headers    map[string]string      `json:"headers"`
	Body       interface{}            `json:"body"`
	Latency    int                    `json:"latency"` // Milliseconds
	Timestamp  time.Time              `json:"timestamp"`
}

// RequestHistory represents a single request-response pair
type RequestHistory struct {
	Request   APIRequest  `json:"request"`
	Response  APIResponse `json:"response"`
	Timestamp time.Time   `json:"timestamp"`
}

// RESTAPIState represents the complete state of the REST API simulator
type RESTAPIState struct {
	Resources      map[string]map[int]*Resource `json:"resources"` // resourceType -> id -> Resource
	RequestHistory []RequestHistory             `json:"requestHistory"`
	TotalRequests  int                          `json:"totalRequests"`
	NextID         map[string]int               `json:"nextId"` // Next ID for each resource type
}

// StatusCodeInfo provides information about HTTP status codes
type StatusCodeInfo struct {
	Code        int    `json:"code"`
	Text        string `json:"text"`
	Description string `json:"description"`
	Category    string `json:"category"` // "success", "redirect", "client_error", "server_error"
}

