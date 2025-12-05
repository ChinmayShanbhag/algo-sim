package cdc

import "time"

// OperationType represents the type of database operation
type OperationType string

const (
	OpInsert OperationType = "INSERT"
	OpUpdate OperationType = "UPDATE"
	OpDelete OperationType = "DELETE"
)

// DatabaseRecord represents a record in the primary database
type DatabaseRecord struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ChangeEvent represents a change captured from the database
type ChangeEvent struct {
	ID          string         `json:"id"`
	Operation   OperationType  `json:"operation"`
	Table       string         `json:"table"`
	Record      DatabaseRecord `json:"record"`
	OldRecord   *DatabaseRecord `json:"oldRecord,omitempty"` // For updates
	Timestamp   time.Time      `json:"timestamp"`
	LSN         int64          `json:"lsn"` // Log Sequence Number
}

// KafkaMessage represents a message in the Kafka broker
type KafkaMessage struct {
	ID        string       `json:"id"`
	Topic     string       `json:"topic"`
	Partition int          `json:"partition"`
	Offset    int64        `json:"offset"`
	Event     ChangeEvent  `json:"event"`
	Status    string       `json:"status"` // "pending", "processing", "delivered"
	Timestamp time.Time    `json:"timestamp"`
}

// SearchIndexEntry represents an entry in the search index
type SearchIndexEntry struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updatedAt"`
	Indexed   bool      `json:"indexed"`
}

// CacheEntry represents an entry in the cache
type CacheEntry struct {
	Key       string         `json:"key"`
	Value     DatabaseRecord `json:"value"`
	TTL       int            `json:"ttl"` // seconds
	UpdatedAt time.Time      `json:"updatedAt"`
}

// SystemState represents the complete CDC system state
type SystemState struct {
	// Primary Database
	DatabaseRecords []DatabaseRecord `json:"databaseRecords"`
	ReplicationLog  []ChangeEvent    `json:"replicationLog"`
	
	// Kafka Broker
	KafkaMessages []KafkaMessage `json:"kafkaMessages"`
	
	// Derived Systems
	SearchIndex []SearchIndexEntry `json:"searchIndex"`
	Cache       []CacheEntry       `json:"cache"`
	
	// Statistics
	Stats Statistics `json:"stats"`
}

// Statistics tracks CDC pipeline metrics
type Statistics struct {
	TotalChanges      int     `json:"totalChanges"`
	ChangesProcessed  int     `json:"changesProcessed"`
	MessagesSent      int     `json:"messagesSent"`
	SearchIndexed     int     `json:"searchIndexed"`
	CacheUpdated      int     `json:"cacheUpdated"`
	AverageLatencyMs  float64 `json:"averageLatencyMs"`
	ThroughputPerSec  float64 `json:"throughputPerSec"`
}

