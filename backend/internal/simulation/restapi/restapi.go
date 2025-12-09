package restapi

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// RESTAPISimulator simulates a REST API server with CRUD operations
type RESTAPISimulator struct {
	mu             sync.RWMutex
	resources      map[string]map[int]*Resource
	requestHistory []RequestHistory
	totalRequests  int
	nextID         map[string]int
}

// NewSimulator creates a new REST API simulator with sample data
func NewSimulator() *RESTAPISimulator {
	sim := &RESTAPISimulator{
		resources:      make(map[string]map[int]*Resource),
		requestHistory: []RequestHistory{},
		nextID:         make(map[string]int),
	}
	
	// Initialize with sample data
	sim.seedData()
	
	return sim
}

// seedData populates the API with sample resources
func (s *RESTAPISimulator) seedData() {
	now := time.Now()
	
	// Create users
	s.resources["users"] = make(map[int]*Resource)
	s.nextID["users"] = 1
	
	users := []map[string]interface{}{
		{"name": "Alice Johnson", "email": "alice@example.com", "role": "admin"},
		{"name": "Bob Smith", "email": "bob@example.com", "role": "user"},
		{"name": "Charlie Brown", "email": "charlie@example.com", "role": "user"},
	}
	
	for _, userData := range users {
		id := s.nextID["users"]
		s.resources["users"][id] = &Resource{
			ID:        id,
			Type:      "users",
			Data:      userData,
			CreatedAt: now,
			UpdatedAt: now,
		}
		s.nextID["users"]++
	}
	
	// Create posts
	s.resources["posts"] = make(map[int]*Resource)
	s.nextID["posts"] = 1
	
	posts := []map[string]interface{}{
		{"title": "Introduction to REST APIs", "body": "REST is an architectural style...", "userId": 1},
		{"title": "Understanding HTTP Methods", "body": "GET, POST, PUT, DELETE...", "userId": 1},
		{"title": "Building Scalable APIs", "body": "Best practices for API design...", "userId": 2},
	}
	
	for _, postData := range posts {
		id := s.nextID["posts"]
		s.resources["posts"][id] = &Resource{
			ID:        id,
			Type:      "posts",
			Data:      postData,
			CreatedAt: now,
			UpdatedAt: now,
		}
		s.nextID["posts"]++
	}
	
	// Create comments
	s.resources["comments"] = make(map[int]*Resource)
	s.nextID["comments"] = 1
	
	comments := []map[string]interface{}{
		{"text": "Great article!", "postId": 1, "userId": 2},
		{"text": "Very helpful, thanks!", "postId": 1, "userId": 3},
		{"text": "Looking forward to more content", "postId": 2, "userId": 3},
	}
	
	for _, commentData := range comments {
		id := s.nextID["comments"]
		s.resources["comments"][id] = &Resource{
			ID:        id,
			Type:      "comments",
			Data:      commentData,
			CreatedAt: now,
			UpdatedAt: now,
		}
		s.nextID["comments"]++
	}
}

// HandleRequest processes a REST API request and returns a response
func (s *RESTAPISimulator) HandleRequest(req APIRequest) APIResponse {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.totalRequests++
	startTime := time.Now()
	
	var response APIResponse
	
	// Parse the path
	pathParts := strings.Split(strings.Trim(req.Path, "/"), "/")
	
	// Route the request
	if len(pathParts) == 0 || pathParts[0] == "" {
		response = s.handleRoot(req)
	} else {
		resourceType := pathParts[0]
		
		if len(pathParts) == 1 {
			// Collection endpoint: /users, /posts, etc.
			response = s.handleCollection(req, resourceType)
		} else if len(pathParts) == 2 {
			// Resource endpoint: /users/1, /posts/2, etc.
			response = s.handleResource(req, resourceType, pathParts[1])
		} else {
			response = s.createErrorResponse(404, "Not Found", "Invalid path")
		}
	}
	
	// Calculate latency (simulate network delay)
	latency := int(time.Since(startTime).Milliseconds())
	if latency < 10 {
		latency = 10 + (s.totalRequests % 20) // Simulate 10-30ms latency
	}
	response.Latency = latency
	response.Timestamp = time.Now()
	
	// Add standard headers
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}
	response.Headers["Content-Type"] = "application/json"
	response.Headers["X-Request-ID"] = fmt.Sprintf("req-%d", s.totalRequests)
	
	// Record in history
	s.requestHistory = append(s.requestHistory, RequestHistory{
		Request:   req,
		Response:  response,
		Timestamp: time.Now(),
	})
	
	return response
}

// handleRoot handles requests to the root path
func (s *RESTAPISimulator) handleRoot(req APIRequest) APIResponse {
	if req.Method != GET {
		return s.createErrorResponse(405, "Method Not Allowed", "Only GET is allowed on root")
	}
	
	return APIResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body: map[string]interface{}{
			"message": "Welcome to the REST API Simulator",
			"version": "1.0.0",
			"endpoints": []string{
				"GET /users",
				"GET /users/:id",
				"POST /users",
				"PUT /users/:id",
				"PATCH /users/:id",
				"DELETE /users/:id",
				"GET /posts",
				"GET /posts/:id",
				"POST /posts",
				"PUT /posts/:id",
				"PATCH /posts/:id",
				"DELETE /posts/:id",
				"GET /comments",
				"GET /comments/:id",
				"POST /comments",
				"PUT /comments/:id",
				"PATCH /comments/:id",
				"DELETE /comments/:id",
			},
		},
	}
}

// handleCollection handles requests to collection endpoints
func (s *RESTAPISimulator) handleCollection(req APIRequest, resourceType string) APIResponse {
	switch req.Method {
	case GET:
		return s.handleGetCollection(req, resourceType)
	case POST:
		return s.handlePostCollection(req, resourceType)
	default:
		return s.createErrorResponse(405, "Method Not Allowed", fmt.Sprintf("Method %s not allowed on collections", req.Method))
	}
}

// handleGetCollection handles GET requests to collections
func (s *RESTAPISimulator) handleGetCollection(req APIRequest, resourceType string) APIResponse {
	collection, exists := s.resources[resourceType]
	if !exists {
		return s.createErrorResponse(404, "Not Found", fmt.Sprintf("Resource type '%s' not found", resourceType))
	}
	
	// Convert map to slice
	items := []interface{}{}
	for _, resource := range collection {
		items = append(items, map[string]interface{}{
			"id":        resource.ID,
			"data":      resource.Data,
			"createdAt": resource.CreatedAt,
			"updatedAt": resource.UpdatedAt,
		})
	}
	
	return APIResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body: map[string]interface{}{
			"data":  items,
			"count": len(items),
		},
	}
}

// handlePostCollection handles POST requests to collections
func (s *RESTAPISimulator) handlePostCollection(req APIRequest, resourceType string) APIResponse {
	if req.Body == nil || len(req.Body) == 0 {
		return s.createErrorResponse(400, "Bad Request", "Request body is required")
	}
	
	// Ensure resource type exists
	if _, exists := s.resources[resourceType]; !exists {
		s.resources[resourceType] = make(map[int]*Resource)
		s.nextID[resourceType] = 1
	}
	
	// Create new resource
	id := s.nextID[resourceType]
	now := time.Now()
	
	resource := &Resource{
		ID:        id,
		Type:      resourceType,
		Data:      req.Body,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	s.resources[resourceType][id] = resource
	s.nextID[resourceType]++
	
	return APIResponse{
		StatusCode: 201,
		StatusText: "Created",
		Headers: map[string]string{
			"Location": fmt.Sprintf("/%s/%d", resourceType, id),
		},
		Body: map[string]interface{}{
			"id":        resource.ID,
			"data":      resource.Data,
			"createdAt": resource.CreatedAt,
			"updatedAt": resource.UpdatedAt,
		},
	}
}

// handleResource handles requests to individual resource endpoints
func (s *RESTAPISimulator) handleResource(req APIRequest, resourceType, idStr string) APIResponse {
	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil {
		return s.createErrorResponse(400, "Bad Request", "Invalid resource ID")
	}
	
	switch req.Method {
	case GET:
		return s.handleGetResource(resourceType, id)
	case PUT:
		return s.handlePutResource(req, resourceType, id)
	case PATCH:
		return s.handlePatchResource(req, resourceType, id)
	case DELETE:
		return s.handleDeleteResource(resourceType, id)
	default:
		return s.createErrorResponse(405, "Method Not Allowed", fmt.Sprintf("Method %s not allowed on resources", req.Method))
	}
}

// handleGetResource handles GET requests to individual resources
func (s *RESTAPISimulator) handleGetResource(resourceType string, id int) APIResponse {
	collection, exists := s.resources[resourceType]
	if !exists {
		return s.createErrorResponse(404, "Not Found", fmt.Sprintf("Resource type '%s' not found", resourceType))
	}
	
	resource, exists := collection[id]
	if !exists {
		return s.createErrorResponse(404, "Not Found", fmt.Sprintf("Resource with ID %d not found", id))
	}
	
	return APIResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body: map[string]interface{}{
			"id":        resource.ID,
			"data":      resource.Data,
			"createdAt": resource.CreatedAt,
			"updatedAt": resource.UpdatedAt,
		},
	}
}

// handlePutResource handles PUT requests (full update)
func (s *RESTAPISimulator) handlePutResource(req APIRequest, resourceType string, id int) APIResponse {
	if req.Body == nil || len(req.Body) == 0 {
		return s.createErrorResponse(400, "Bad Request", "Request body is required")
	}
	
	collection, exists := s.resources[resourceType]
	if !exists {
		return s.createErrorResponse(404, "Not Found", fmt.Sprintf("Resource type '%s' not found", resourceType))
	}
	
	resource, exists := collection[id]
	if !exists {
		return s.createErrorResponse(404, "Not Found", fmt.Sprintf("Resource with ID %d not found", id))
	}
	
	// Full replacement
	resource.Data = req.Body
	resource.UpdatedAt = time.Now()
	
	return APIResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body: map[string]interface{}{
			"id":        resource.ID,
			"data":      resource.Data,
			"createdAt": resource.CreatedAt,
			"updatedAt": resource.UpdatedAt,
		},
	}
}

// handlePatchResource handles PATCH requests (partial update)
func (s *RESTAPISimulator) handlePatchResource(req APIRequest, resourceType string, id int) APIResponse {
	if req.Body == nil || len(req.Body) == 0 {
		return s.createErrorResponse(400, "Bad Request", "Request body is required")
	}
	
	collection, exists := s.resources[resourceType]
	if !exists {
		return s.createErrorResponse(404, "Not Found", fmt.Sprintf("Resource type '%s' not found", resourceType))
	}
	
	resource, exists := collection[id]
	if !exists {
		return s.createErrorResponse(404, "Not Found", fmt.Sprintf("Resource with ID %d not found", id))
	}
	
	// Partial update - merge fields
	for key, value := range req.Body {
		resource.Data[key] = value
	}
	resource.UpdatedAt = time.Now()
	
	return APIResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body: map[string]interface{}{
			"id":        resource.ID,
			"data":      resource.Data,
			"createdAt": resource.CreatedAt,
			"updatedAt": resource.UpdatedAt,
		},
	}
}

// handleDeleteResource handles DELETE requests
func (s *RESTAPISimulator) handleDeleteResource(resourceType string, id int) APIResponse {
	collection, exists := s.resources[resourceType]
	if !exists {
		return s.createErrorResponse(404, "Not Found", fmt.Sprintf("Resource type '%s' not found", resourceType))
	}
	
	_, exists = collection[id]
	if !exists {
		return s.createErrorResponse(404, "Not Found", fmt.Sprintf("Resource with ID %d not found", id))
	}
	
	delete(collection, id)
	
	return APIResponse{
		StatusCode: 204,
		StatusText: "No Content",
		Body:       nil,
	}
}

// createErrorResponse creates an error response
func (s *RESTAPISimulator) createErrorResponse(statusCode int, statusText, message string) APIResponse {
	return APIResponse{
		StatusCode: statusCode,
		StatusText: statusText,
		Body: map[string]interface{}{
			"error": map[string]interface{}{
				"code":    statusCode,
				"message": message,
			},
		},
	}
}

// GetState returns the current state of the REST API simulator
func (s *RESTAPISimulator) GetState() RESTAPIState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return RESTAPIState{
		Resources:      s.resources,
		RequestHistory: s.requestHistory,
		TotalRequests:  s.totalRequests,
		NextID:         s.nextID,
	}
}

// Reset resets the simulator to initial state
func (s *RESTAPISimulator) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.resources = make(map[string]map[int]*Resource)
	s.requestHistory = []RequestHistory{}
	s.totalRequests = 0
	s.nextID = make(map[string]int)
	
	s.seedData()
}

