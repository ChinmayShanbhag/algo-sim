package graphql

import "time"

// GraphQLRequest represents a GraphQL query/mutation request
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
	Operation string                 `json:"operationName,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data       interface{}            `json:"data,omitempty"`
	Errors     []GraphQLError         `json:"errors,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                 `json:"message"`
	Locations  []ErrorLocation        `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// ErrorLocation represents the location of an error in the query
type ErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// User represents a user in the system
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// Post represents a blog post
type Post struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	AuthorID int    `json:"authorId"`
	Author   *User  `json:"author,omitempty"`
}

// Comment represents a comment on a post
type Comment struct {
	ID       int    `json:"id"`
	Text     string `json:"text"`
	PostID   int    `json:"postId"`
	AuthorID int    `json:"authorId"`
	Author   *User  `json:"author,omitempty"`
}

// QueryHistory represents a single GraphQL query execution
type QueryHistory struct {
	Request   GraphQLRequest  `json:"request"`
	Response  GraphQLResponse `json:"response"`
	Latency   int             `json:"latency"` // Milliseconds
	Timestamp time.Time       `json:"timestamp"`
	IsError   bool            `json:"isError"`
}

// GraphQLState represents the complete state of the GraphQL simulator
type GraphQLState struct {
	Users         map[int]*User    `json:"users"`
	Posts         map[int]*Post    `json:"posts"`
	Comments      map[int]*Comment `json:"comments"`
	QueryHistory  []QueryHistory   `json:"queryHistory"`
	TotalQueries  int              `json:"totalQueries"`
	NextUserID    int              `json:"nextUserId"`
	NextPostID    int              `json:"nextPostId"`
	NextCommentID int              `json:"nextCommentId"`
}
