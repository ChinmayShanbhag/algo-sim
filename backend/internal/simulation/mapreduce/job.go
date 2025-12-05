package mapreduce

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Job represents a MapReduce job
type Job struct {
	mu           sync.RWMutex
	jobID        string
	status       JobStatus
	inputData    []string
	numMappers   int
	numReducers  int
	mapTasks     []MapTask
	shuffleData  []ShufflePartition
	reduceTasks  []ReduceTask
	finalOutput  []KeyValue
	startTime    time.Time
	endTime      time.Time
}

// NewJob creates a new MapReduce job
func NewJob(jobID string, inputData []string, numMappers, numReducers int) *Job {
	return &Job{
		jobID:       jobID,
		status:      StatusIdle,
		inputData:   inputData,
		numMappers:  numMappers,
		numReducers: numReducers,
		mapTasks:    []MapTask{},
		shuffleData: []ShufflePartition{},
		reduceTasks: []ReduceTask{},
		finalOutput: []KeyValue{},
	}
}

// Start begins the MapReduce job execution
func (j *Job) Start() {
	j.mu.Lock()
	defer j.mu.Unlock()

	if j.status != StatusIdle {
		return
	}

	j.status = StatusMapping
	j.startTime = time.Now()
	j.createMapTasks()
}

// createMapTasks divides input data among mappers
func (j *Job) createMapTasks() {
	j.mapTasks = make([]MapTask, 0)
	
	// Distribute input data across mappers
	for i, data := range j.inputData {
		workerID := i % j.numMappers
		task := MapTask{
			WorkerID:    workerID,
			InputData:   data,
			OutputPairs: []KeyValue{},
			Status:      "pending",
		}
		j.mapTasks = append(j.mapTasks, task)
	}
}

// ExecuteMapPhase simulates the map phase
func (j *Job) ExecuteMapPhase() {
	j.mu.Lock()
	defer j.mu.Unlock()

	if j.status != StatusMapping {
		return
	}

	// Execute map function on each task (word count example)
	for i := range j.mapTasks {
		j.mapTasks[i].Status = "running"
		j.mapTasks[i].StartTime = time.Now()
		
		// Map function: emit (word, 1) for each word
		words := strings.Fields(j.mapTasks[i].InputData)
		for _, word := range words {
			word = strings.ToLower(strings.Trim(word, ".,!?;:"))
			if word != "" {
				j.mapTasks[i].OutputPairs = append(j.mapTasks[i].OutputPairs, KeyValue{
					Key:   word,
					Value: "1",
				})
			}
		}
		
		j.mapTasks[i].Status = "completed"
		j.mapTasks[i].EndTime = time.Now()
	}

	j.status = StatusShuffling
}

// ExecuteShufflePhase simulates the shuffle and sort phase
func (j *Job) ExecuteShufflePhase() {
	j.mu.Lock()
	defer j.mu.Unlock()

	if j.status != StatusShuffling {
		return
	}

	// Collect all key-value pairs from map tasks
	allPairs := make(map[string][]string)
	for _, task := range j.mapTasks {
		for _, pair := range task.OutputPairs {
			allPairs[pair.Key] = append(allPairs[pair.Key], pair.Value)
		}
	}

	// Partition keys across reducers and sort
	keys := make([]string, 0, len(allPairs))
	for key := range allPairs {
		keys = append(keys, key)
	}
	sort.Strings(keys) // Sort keys

	j.shuffleData = make([]ShufflePartition, 0)
	for _, key := range keys {
		// Hash key to determine reducer (simple modulo)
		reducerID := hashKey(key) % j.numReducers
		partition := ShufflePartition{
			ReducerID: reducerID,
			Key:       key,
			Values:    allPairs[key],
		}
		j.shuffleData = append(j.shuffleData, partition)
	}

	j.status = StatusReducing
}

// ExecuteReducePhase simulates the reduce phase
func (j *Job) ExecuteReducePhase() {
	j.mu.Lock()
	defer j.mu.Unlock()

	if j.status != StatusReducing {
		return
	}

	// Create reduce tasks from shuffle data
	j.reduceTasks = make([]ReduceTask, 0)
	for _, partition := range j.shuffleData {
		task := ReduceTask{
			WorkerID:  partition.ReducerID,
			Key:       partition.Key,
			Values:    partition.Values,
			Status:    "running",
			StartTime: time.Now(),
		}

		// Reduce function: sum all values for word count
		count := 0
		for range task.Values {
			count++ // Each value is "1"
		}
		task.Result = fmt.Sprintf("%d", count)
		task.Status = "completed"
		task.EndTime = time.Now()

		j.reduceTasks = append(j.reduceTasks, task)
	}

	// Collect final output
	j.finalOutput = make([]KeyValue, 0)
	for _, task := range j.reduceTasks {
		j.finalOutput = append(j.finalOutput, KeyValue{
			Key:   task.Key,
			Value: task.Result,
		})
	}

	j.status = StatusCompleted
	j.endTime = time.Now()
}

// GetState returns the current state of the job
func (j *Job) GetState() JobState {
	j.mu.RLock()
	defer j.mu.RUnlock()

	// Calculate progress
	progress := 0
	switch j.status {
	case StatusIdle:
		progress = 0
	case StatusMapping:
		completed := 0
		for _, task := range j.mapTasks {
			if task.Status == "completed" {
				completed++
			}
		}
		if len(j.mapTasks) > 0 {
			progress = (completed * 33) / len(j.mapTasks)
		}
	case StatusShuffling:
		progress = 33
	case StatusReducing:
		completed := 0
		for _, task := range j.reduceTasks {
			if task.Status == "completed" {
				completed++
			}
		}
		if len(j.reduceTasks) > 0 {
			progress = 33 + (completed * 34) / len(j.reduceTasks)
		}
	case StatusCompleted:
		progress = 100
	}

	currentStage := ""
	switch j.status {
	case StatusIdle:
		currentStage = "Ready to start"
	case StatusMapping:
		currentStage = "Map Phase: Processing input data"
	case StatusShuffling:
		currentStage = "Shuffle Phase: Partitioning and sorting"
	case StatusReducing:
		currentStage = "Reduce Phase: Aggregating results"
	case StatusCompleted:
		currentStage = "Completed"
	}

	return JobState{
		JobID:        j.jobID,
		Status:       j.status,
		InputData:    j.inputData,
		NumMappers:   j.numMappers,
		NumReducers:  j.numReducers,
		MapTasks:     j.mapTasks,
		ShuffleData:  j.shuffleData,
		ReduceTasks:  j.reduceTasks,
		FinalOutput:  j.finalOutput,
		CurrentStage: currentStage,
		Progress:     progress,
		StartTime:    j.startTime,
		EndTime:      j.endTime,
	}
}

// Reset resets the job to initial state
func (j *Job) Reset() {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.status = StatusIdle
	j.mapTasks = []MapTask{}
	j.shuffleData = []ShufflePartition{}
	j.reduceTasks = []ReduceTask{}
	j.finalOutput = []KeyValue{}
	j.startTime = time.Time{}
	j.endTime = time.Time{}
}

// hashKey returns a simple hash of the key for partitioning
func hashKey(key string) int {
	hash := 0
	for _, char := range key {
		hash = hash*31 + int(char)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

