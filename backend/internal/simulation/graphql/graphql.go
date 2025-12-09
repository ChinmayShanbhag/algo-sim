package graphql

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// GraphQLSimulator simulates a GraphQL API server
type GraphQLSimulator struct {
	mu            sync.RWMutex
	users         map[int]*User
	posts         map[int]*Post
	comments      map[int]*Comment
	queryHistory  []QueryHistory
	totalQueries  int
	nextUserID    int
	nextPostID    int
	nextCommentID int
}

// NewSimulator creates a new GraphQL simulator with sample data
func NewSimulator() *GraphQLSimulator {
	sim := &GraphQLSimulator{
		users:         make(map[int]*User),
		posts:         make(map[int]*Post),
		comments:      make(map[int]*Comment),
		queryHistory:  []QueryHistory{},
		nextUserID:    1,
		nextPostID:    1,
		nextCommentID: 1,
	}
	
	sim.seedData()
	return sim
}

// seedData populates the simulator with sample data
func (s *GraphQLSimulator) seedData() {
	// Create users
	users := []User{
		{ID: 1, Name: "Alice Johnson", Email: "alice@example.com", Role: "admin"},
		{ID: 2, Name: "Bob Smith", Email: "bob@example.com", Role: "user"},
		{ID: 3, Name: "Charlie Brown", Email: "charlie@example.com", Role: "user"},
	}
	
	for _, user := range users {
		s.users[user.ID] = &User{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		}
	}
	s.nextUserID = 4
	
	// Create posts
	posts := []Post{
		{ID: 1, Title: "Introduction to GraphQL", Body: "GraphQL is a query language...", AuthorID: 1},
		{ID: 2, Title: "Understanding Queries", Body: "Queries in GraphQL...", AuthorID: 1},
		{ID: 3, Title: "Mutations Explained", Body: "Mutations allow you to...", AuthorID: 2},
	}
	
	for _, post := range posts {
		s.posts[post.ID] = &Post{
			ID:       post.ID,
			Title:    post.Title,
			Body:     post.Body,
			AuthorID: post.AuthorID,
		}
	}
	s.nextPostID = 4
	
	// Create comments
	comments := []Comment{
		{ID: 1, Text: "Great article!", PostID: 1, AuthorID: 2},
		{ID: 2, Text: "Very helpful, thanks!", PostID: 1, AuthorID: 3},
		{ID: 3, Text: "Looking forward to more", PostID: 2, AuthorID: 3},
	}
	
	for _, comment := range comments {
		s.comments[comment.ID] = &Comment{
			ID:       comment.ID,
			Text:     comment.Text,
			PostID:   comment.PostID,
			AuthorID: comment.AuthorID,
		}
	}
	s.nextCommentID = 4
}

// extractFields extracts the requested fields from a GraphQL query for a given type
func extractFields(query string, typeName string) []string {
	// Find the type block (e.g., "users { id name email }")
	pattern := regexp.MustCompile(typeName + `\s*\{([^}]+)\}`)
	matches := pattern.FindStringSubmatch(query)
	
	if len(matches) < 2 {
		// If no fields specified, return all common fields
		return []string{"id", "name", "email", "role"}
	}
	
	fieldsStr := matches[1]
	
	// Split by whitespace and newlines to get individual fields
	fields := []string{}
	for _, field := range regexp.MustCompile(`\s+`).Split(strings.TrimSpace(fieldsStr), -1) {
		field = strings.TrimSpace(field)
		if field != "" && !strings.Contains(field, "{") {
			fields = append(fields, field)
		}
	}
	
	return fields
}

// ExecuteQuery executes a GraphQL query
func (s *GraphQLSimulator) ExecuteQuery(req GraphQLRequest) (GraphQLResponse, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	startTime := time.Now()
	s.totalQueries++
	
	// Simple query parser (not a full GraphQL parser, just for simulation)
	query := strings.TrimSpace(req.Query)
	
	// Strip query/mutation keywords and operation names for easier parsing
	cleanQuery := regexp.MustCompile(`^(query|mutation)\s+\w*\s*(\([^)]*\))?\s*`).ReplaceAllString(query, "")
	
	var response GraphQLResponse
	var err error
	
	// Determine if it's a query or mutation based on original query
	if strings.Contains(strings.ToLower(query), "mutation") {
		response, err = s.executeMutationOperation(cleanQuery, req.Variables)
	} else {
		// Default to query for both explicit "query" and shorthand "{...}"
		response, err = s.executeQueryOperation(cleanQuery, req.Variables)
	}
	
	if err != nil {
		response.Errors = []GraphQLError{
			{Message: err.Error()},
		}
	}
	
	latency := int(time.Since(startTime).Milliseconds())
	if latency < 5 {
		latency = 5 + (s.totalQueries % 10) // Simulate 5-15ms latency
	}
	
	// Record in history
	s.queryHistory = append(s.queryHistory, QueryHistory{
		Request:   req,
		Response:  response,
		Latency:   latency,
		Timestamp: time.Now(),
		IsError:   len(response.Errors) > 0,
	})
	
	return response, latency
}

// executeQueryOperation handles query operations
func (s *GraphQLSimulator) executeQueryOperation(query string, variables map[string]interface{}) (GraphQLResponse, error) {
	data := make(map[string]interface{})
	
	// Parse for users query
	if strings.Contains(query, "users") {
		// Extract requested fields for users
		requestedFields := extractFields(query, "users")
		
		users := []map[string]interface{}{}
		for _, user := range s.users {
			userData := make(map[string]interface{})
			
			// Only include requested fields
			for _, field := range requestedFields {
				switch field {
				case "id":
					userData["id"] = user.ID
				case "name":
					userData["name"] = user.Name
				case "email":
					userData["email"] = user.Email
				case "role":
					userData["role"] = user.Role
				}
			}
			
			users = append(users, userData)
		}
		data["users"] = users
	}
	
	// Parse for user(id:) query
	if match := regexp.MustCompile(`user\s*\(\s*id\s*:\s*(\d+|\$\w+)\s*\)`).FindStringSubmatch(query); match != nil {
		idStr := match[1]
		var id int
		
		if strings.HasPrefix(idStr, "$") {
			// Variable reference
			varName := strings.TrimPrefix(idStr, "$")
			if val, ok := variables[varName]; ok {
				if idFloat, ok := val.(float64); ok {
					id = int(idFloat)
				} else if idInt, ok := val.(int); ok {
					id = idInt
				}
			}
		} else {
			id, _ = strconv.Atoi(idStr)
		}
		
		if user, exists := s.users[id]; exists {
			// Extract requested fields for user
			requestedFields := extractFields(query, "user")
			
			userData := make(map[string]interface{})
			for _, field := range requestedFields {
				switch field {
				case "id":
					userData["id"] = user.ID
				case "name":
					userData["name"] = user.Name
				case "email":
					userData["email"] = user.Email
				case "role":
					userData["role"] = user.Role
				}
			}
			
			data["user"] = userData
		} else {
			data["user"] = nil
		}
	}
	
	// Parse for posts query
	if strings.Contains(query, "posts") {
		// Extract requested fields for posts
		requestedFields := extractFields(query, "posts")
		
		posts := []map[string]interface{}{}
		for _, post := range s.posts {
			postData := make(map[string]interface{})
			
			// Only include requested fields
			for _, field := range requestedFields {
				switch field {
				case "id":
					postData["id"] = post.ID
				case "title":
					postData["title"] = post.Title
				case "body":
					postData["body"] = post.Body
				case "authorId":
					postData["authorId"] = post.AuthorID
				case "author":
					// Handle nested author fields
					if author, exists := s.users[post.AuthorID]; exists {
						authorFields := extractFields(query, "author")
						authorData := make(map[string]interface{})
						
						for _, authorField := range authorFields {
							switch authorField {
							case "id":
								authorData["id"] = author.ID
							case "name":
								authorData["name"] = author.Name
							case "email":
								authorData["email"] = author.Email
							case "role":
								authorData["role"] = author.Role
							}
						}
						
						postData["author"] = authorData
					}
				}
			}
			
			posts = append(posts, postData)
		}
		data["posts"] = posts
	}
	
	// Parse for post(id:) query
	if match := regexp.MustCompile(`post\s*\(\s*id\s*:\s*(\d+|\$\w+)\s*\)`).FindStringSubmatch(query); match != nil {
		idStr := match[1]
		var id int
		
		if strings.HasPrefix(idStr, "$") {
			varName := strings.TrimPrefix(idStr, "$")
			if val, ok := variables[varName]; ok {
				if idFloat, ok := val.(float64); ok {
					id = int(idFloat)
				}
			}
		} else {
			id, _ = strconv.Atoi(idStr)
		}
		
		if post, exists := s.posts[id]; exists {
			// Extract requested fields for post
			requestedFields := extractFields(query, "post")
			
			postData := make(map[string]interface{})
			for _, field := range requestedFields {
				switch field {
				case "id":
					postData["id"] = post.ID
				case "title":
					postData["title"] = post.Title
				case "body":
					postData["body"] = post.Body
				case "authorId":
					postData["authorId"] = post.AuthorID
				case "author":
					// Handle nested author fields
					if author, exists := s.users[post.AuthorID]; exists {
						authorFields := extractFields(query, "author")
						authorData := make(map[string]interface{})
						
						for _, authorField := range authorFields {
							switch authorField {
							case "id":
								authorData["id"] = author.ID
							case "name":
								authorData["name"] = author.Name
							case "email":
								authorData["email"] = author.Email
							case "role":
								authorData["role"] = author.Role
							}
						}
						
						postData["author"] = authorData
					}
				}
			}
			
			data["post"] = postData
		} else {
			data["post"] = nil
		}
	}
	
	return GraphQLResponse{Data: data}, nil
}

// executeMutationOperation handles mutation operations
func (s *GraphQLSimulator) executeMutationOperation(query string, variables map[string]interface{}) (GraphQLResponse, error) {
	data := make(map[string]interface{})
	
	// Parse for createUser mutation
	if strings.Contains(query, "createUser") {
		// Extract input from variables or inline
		var name, email, role string
		
		if input, ok := variables["input"].(map[string]interface{}); ok {
			name, _ = input["name"].(string)
			email, _ = input["email"].(string)
			role, _ = input["role"].(string)
		}
		
		if name == "" {
			return GraphQLResponse{}, fmt.Errorf("name is required")
		}
		
		user := &User{
			ID:    s.nextUserID,
			Name:  name,
			Email: email,
			Role:  role,
		}
		s.users[user.ID] = user
		s.nextUserID++
		
		// Extract requested fields
		requestedFields := extractFields(query, "createUser")
		userData := make(map[string]interface{})
		
		for _, field := range requestedFields {
			switch field {
			case "id":
				userData["id"] = user.ID
			case "name":
				userData["name"] = user.Name
			case "email":
				userData["email"] = user.Email
			case "role":
				userData["role"] = user.Role
			}
		}
		
		data["createUser"] = userData
	}
	
	// Parse for createPost mutation
	if strings.Contains(query, "createPost") {
		var title, body string
		var authorID int
		
		if input, ok := variables["input"].(map[string]interface{}); ok {
			title, _ = input["title"].(string)
			body, _ = input["body"].(string)
			if authorIDFloat, ok := input["authorId"].(float64); ok {
				authorID = int(authorIDFloat)
			}
		}
		
		if title == "" {
			return GraphQLResponse{}, fmt.Errorf("title is required")
		}
		
		post := &Post{
			ID:       s.nextPostID,
			Title:    title,
			Body:     body,
			AuthorID: authorID,
		}
		s.posts[post.ID] = post
		s.nextPostID++
		
		// Extract requested fields
		requestedFields := extractFields(query, "createPost")
		postData := make(map[string]interface{})
		
		for _, field := range requestedFields {
			switch field {
			case "id":
				postData["id"] = post.ID
			case "title":
				postData["title"] = post.Title
			case "body":
				postData["body"] = post.Body
			case "authorId":
				postData["authorId"] = post.AuthorID
			}
		}
		
		data["createPost"] = postData
	}
	
	// Parse for updateUser mutation
	if match := regexp.MustCompile(`updateUser\s*\(\s*id\s*:\s*(\d+|\$\w+)`).FindStringSubmatch(query); match != nil {
		idStr := match[1]
		var id int
		
		if strings.HasPrefix(idStr, "$") {
			varName := strings.TrimPrefix(idStr, "$")
			if val, ok := variables[varName]; ok {
				if idFloat, ok := val.(float64); ok {
					id = int(idFloat)
				}
			}
		} else {
			id, _ = strconv.Atoi(idStr)
		}
		
		if user, exists := s.users[id]; exists {
			if input, ok := variables["input"].(map[string]interface{}); ok {
				if name, ok := input["name"].(string); ok && name != "" {
					user.Name = name
				}
				if email, ok := input["email"].(string); ok && email != "" {
					user.Email = email
				}
				if role, ok := input["role"].(string); ok && role != "" {
					user.Role = role
				}
			}
			
			// Extract requested fields
			requestedFields := extractFields(query, "updateUser")
			userData := make(map[string]interface{})
			
			for _, field := range requestedFields {
				switch field {
				case "id":
					userData["id"] = user.ID
				case "name":
					userData["name"] = user.Name
				case "email":
					userData["email"] = user.Email
				case "role":
					userData["role"] = user.Role
				}
			}
			
			data["updateUser"] = userData
		} else {
			return GraphQLResponse{}, fmt.Errorf("user with id %d not found", id)
		}
	}
	
	// Parse for deleteUser mutation
	if match := regexp.MustCompile(`deleteUser\s*\(\s*id\s*:\s*(\d+|\$\w+)\s*\)`).FindStringSubmatch(query); match != nil {
		idStr := match[1]
		var id int
		
		if strings.HasPrefix(idStr, "$") {
			varName := strings.TrimPrefix(idStr, "$")
			if val, ok := variables[varName]; ok {
				if idFloat, ok := val.(float64); ok {
					id = int(idFloat)
				}
			}
		} else {
			id, _ = strconv.Atoi(idStr)
		}
		
		if _, exists := s.users[id]; exists {
			delete(s.users, id)
			data["deleteUser"] = map[string]interface{}{
				"success": true,
				"id":      id,
			}
		} else {
			return GraphQLResponse{}, fmt.Errorf("user with id %d not found", id)
		}
	}
	
	return GraphQLResponse{Data: data}, nil
}

// GetState returns the current state of the GraphQL simulator
func (s *GraphQLSimulator) GetState() GraphQLState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return GraphQLState{
		Users:         s.users,
		Posts:         s.posts,
		Comments:      s.comments,
		QueryHistory:  s.queryHistory,
		TotalQueries:  s.totalQueries,
		NextUserID:    s.nextUserID,
		NextPostID:    s.nextPostID,
		NextCommentID: s.nextCommentID,
	}
}

// Reset resets the simulator to initial state
func (s *GraphQLSimulator) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.users = make(map[int]*User)
	s.posts = make(map[int]*Post)
	s.comments = make(map[int]*Comment)
	s.queryHistory = []QueryHistory{}
	s.totalQueries = 0
	s.nextUserID = 1
	s.nextPostID = 1
	s.nextCommentID = 1
	
	s.seedData()
}

