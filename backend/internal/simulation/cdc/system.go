package cdc

import (
	"fmt"
	"sync"
	"time"
)

// CDCSystem represents the complete Change Data Capture pipeline
type CDCSystem struct {
	mu sync.RWMutex
	
	// Primary Database
	database       map[int]DatabaseRecord
	nextID         int
	replicationLog []ChangeEvent
	lsn            int64 // Log Sequence Number
	
	// Kafka Broker
	kafkaMessages []KafkaMessage
	nextOffset    int64
	
	// Derived Systems
	searchIndex map[int]SearchIndexEntry
	cache       map[string]CacheEntry
	
	// Statistics
	totalChanges     int
	changesProcessed int
	messagesSent     int
	searchIndexed    int
	cacheUpdated     int
	startTime        time.Time
}

// NewCDCSystem creates a new CDC system with sample data
func NewCDCSystem() *CDCSystem {
	system := &CDCSystem{
		database:      make(map[int]DatabaseRecord),
		nextID:        1,
		replicationLog: []ChangeEvent{},
		lsn:           1000,
		kafkaMessages: []KafkaMessage{},
		nextOffset:    0,
		searchIndex:   make(map[int]SearchIndexEntry),
		cache:         make(map[string]CacheEntry),
		startTime:     time.Now(),
	}
	
	// Initialize with sample data
	system.initializeSampleData()
	
	return system
}

// initializeSampleData creates initial database records
func (s *CDCSystem) initializeSampleData() {
	sampleRecords := []DatabaseRecord{
		{ID: 1, Name: "Alice Johnson", Email: "alice@example.com", Status: "active", UpdatedAt: time.Now()},
		{ID: 2, Name: "Bob Smith", Email: "bob@example.com", Status: "active", UpdatedAt: time.Now()},
		{ID: 3, Name: "Carol White", Email: "carol@example.com", Status: "inactive", UpdatedAt: time.Now()},
	}
	
	for _, record := range sampleRecords {
		s.database[record.ID] = record
		s.nextID = record.ID + 1
	}
}

// InsertRecord inserts a new record into the database and captures the change
func (s *CDCSystem) InsertRecord(name, email, status string) ChangeEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	record := DatabaseRecord{
		ID:        s.nextID,
		Name:      name,
		Email:     email,
		Status:    status,
		UpdatedAt: time.Now(),
	}
	
	s.database[record.ID] = record
	s.nextID++
	
	// Capture change in replication log
	event := ChangeEvent{
		ID:        fmt.Sprintf("evt-%d", s.lsn),
		Operation: OpInsert,
		Table:     "users",
		Record:    record,
		Timestamp: time.Now(),
		LSN:       s.lsn,
	}
	
	s.lsn++
	s.replicationLog = append(s.replicationLog, event)
	s.totalChanges++
	
	// Keep only last 10 events in log
	if len(s.replicationLog) > 10 {
		s.replicationLog = s.replicationLog[len(s.replicationLog)-10:]
	}
	
	return event
}

// UpdateRecord updates an existing record and captures the change
func (s *CDCSystem) UpdateRecord(id int, name, email, status string) (ChangeEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	oldRecord, exists := s.database[id]
	if !exists {
		return ChangeEvent{}, fmt.Errorf("record with ID %d not found", id)
	}
	
	newRecord := DatabaseRecord{
		ID:        id,
		Name:      name,
		Email:     email,
		Status:    status,
		UpdatedAt: time.Now(),
	}
	
	s.database[id] = newRecord
	
	// Capture change in replication log
	event := ChangeEvent{
		ID:        fmt.Sprintf("evt-%d", s.lsn),
		Operation: OpUpdate,
		Table:     "users",
		Record:    newRecord,
		OldRecord: &oldRecord,
		Timestamp: time.Now(),
		LSN:       s.lsn,
	}
	
	s.lsn++
	s.replicationLog = append(s.replicationLog, event)
	s.totalChanges++
	
	// Keep only last 10 events in log
	if len(s.replicationLog) > 10 {
		s.replicationLog = s.replicationLog[len(s.replicationLog)-10:]
	}
	
	return event, nil
}

// DeleteRecord deletes a record and captures the change
func (s *CDCSystem) DeleteRecord(id int) (ChangeEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	record, exists := s.database[id]
	if !exists {
		return ChangeEvent{}, fmt.Errorf("record with ID %d not found", id)
	}
	
	delete(s.database, id)
	
	// Capture change in replication log
	event := ChangeEvent{
		ID:        fmt.Sprintf("evt-%d", s.lsn),
		Operation: OpDelete,
		Table:     "users",
		Record:    record,
		Timestamp: time.Now(),
		LSN:       s.lsn,
	}
	
	s.lsn++
	s.replicationLog = append(s.replicationLog, event)
	s.totalChanges++
	
	// Keep only last 10 events in log
	if len(s.replicationLog) > 10 {
		s.replicationLog = s.replicationLog[len(s.replicationLog)-10:]
	}
	
	return event, nil
}

// StreamToKafka streams pending changes to Kafka
func (s *CDCSystem) StreamToKafka() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Stream all events from replication log to Kafka
	for _, event := range s.replicationLog {
		// Check if already streamed
		alreadyStreamed := false
		for _, msg := range s.kafkaMessages {
			if msg.Event.ID == event.ID {
				alreadyStreamed = true
				break
			}
		}
		
		if !alreadyStreamed {
			message := KafkaMessage{
				ID:        fmt.Sprintf("msg-%d", s.nextOffset),
				Topic:     "db.changes.users",
				Partition: event.Record.ID % 3, // Simple partitioning
				Offset:    s.nextOffset,
				Event:     event,
				Status:    "pending",
				Timestamp: time.Now(),
			}
			
			s.kafkaMessages = append(s.kafkaMessages, message)
			s.nextOffset++
			s.messagesSent++
		}
	}
	
	// Keep only last 15 messages
	if len(s.kafkaMessages) > 15 {
		s.kafkaMessages = s.kafkaMessages[len(s.kafkaMessages)-15:]
	}
}

// ConsumeFromKafka processes Kafka messages and updates derived systems
func (s *CDCSystem) ConsumeFromKafka() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for i := range s.kafkaMessages {
		if s.kafkaMessages[i].Status == "pending" {
			s.kafkaMessages[i].Status = "processing"
			
			event := s.kafkaMessages[i].Event
			
			// Update Search Index
			s.updateSearchIndex(event)
			
			// Update Cache
			s.updateCache(event)
			
			s.kafkaMessages[i].Status = "delivered"
			s.changesProcessed++
		}
	}
}

// updateSearchIndex updates the search index based on the change event
func (s *CDCSystem) updateSearchIndex(event ChangeEvent) {
	switch event.Operation {
	case OpInsert, OpUpdate:
		entry := SearchIndexEntry{
			ID:        event.Record.ID,
			Name:      event.Record.Name,
			Email:     event.Record.Email,
			Status:    event.Record.Status,
			UpdatedAt: event.Record.UpdatedAt,
			Indexed:   true,
		}
		s.searchIndex[event.Record.ID] = entry
		s.searchIndexed++
		
	case OpDelete:
		delete(s.searchIndex, event.Record.ID)
		s.searchIndexed++
	}
}

// updateCache updates the cache based on the change event
func (s *CDCSystem) updateCache(event ChangeEvent) {
	key := fmt.Sprintf("user:%d", event.Record.ID)
	
	switch event.Operation {
	case OpInsert, OpUpdate:
		entry := CacheEntry{
			Key:       key,
			Value:     event.Record,
			TTL:       300, // 5 minutes
			UpdatedAt: time.Now(),
		}
		s.cache[key] = entry
		s.cacheUpdated++
		
	case OpDelete:
		delete(s.cache, key)
		s.cacheUpdated++
	}
}

// GetState returns the current state of the CDC system
func (s *CDCSystem) GetState() SystemState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Convert database map to slice
	dbRecords := make([]DatabaseRecord, 0, len(s.database))
	for _, record := range s.database {
		dbRecords = append(dbRecords, record)
	}
	
	// Convert search index map to slice
	searchIndexSlice := make([]SearchIndexEntry, 0, len(s.searchIndex))
	for _, entry := range s.searchIndex {
		searchIndexSlice = append(searchIndexSlice, entry)
	}
	
	// Convert cache map to slice
	cacheSlice := make([]CacheEntry, 0, len(s.cache))
	for _, entry := range s.cache {
		cacheSlice = append(cacheSlice, entry)
	}
	
	// Calculate statistics
	elapsed := time.Since(s.startTime).Seconds()
	throughput := 0.0
	if elapsed > 0 {
		throughput = float64(s.changesProcessed) / elapsed
	}
	
	avgLatency := 0.0
	if s.changesProcessed > 0 {
		avgLatency = 50.0 // Simulated latency in ms
	}
	
	stats := Statistics{
		TotalChanges:      s.totalChanges,
		ChangesProcessed:  s.changesProcessed,
		MessagesSent:      s.messagesSent,
		SearchIndexed:     s.searchIndexed,
		CacheUpdated:      s.cacheUpdated,
		AverageLatencyMs:  avgLatency,
		ThroughputPerSec:  throughput,
	}
	
	return SystemState{
		DatabaseRecords: dbRecords,
		ReplicationLog:  s.replicationLog,
		KafkaMessages:   s.kafkaMessages,
		SearchIndex:     searchIndexSlice,
		Cache:           cacheSlice,
		Stats:           stats,
	}
}

// Reset resets the CDC system to initial state
func (s *CDCSystem) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.database = make(map[int]DatabaseRecord)
	s.nextID = 1
	s.replicationLog = []ChangeEvent{}
	s.lsn = 1000
	s.kafkaMessages = []KafkaMessage{}
	s.nextOffset = 0
	s.searchIndex = make(map[int]SearchIndexEntry)
	s.cache = make(map[string]CacheEntry)
	s.totalChanges = 0
	s.changesProcessed = 0
	s.messagesSent = 0
	s.searchIndexed = 0
	s.cacheUpdated = 0
	s.startTime = time.Now()
	
	s.initializeSampleData()
}

