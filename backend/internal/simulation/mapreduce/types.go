package mapreduce

import "time"

// JobStatus represents the current status of a MapReduce job
type JobStatus string

const (
	StatusIdle       JobStatus = "idle"
	StatusMapping    JobStatus = "mapping"
	StatusShuffling  JobStatus = "shuffling"
	StatusReducing   JobStatus = "reducing"
	StatusCompleted  JobStatus = "completed"
)

// KeyValue represents a key-value pair
type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// MapTask represents a map task on a worker
type MapTask struct {
	WorkerID    int         `json:"workerId"`
	InputData   string      `json:"inputData"`
	OutputPairs []KeyValue  `json:"outputPairs"`
	Status      string      `json:"status"` // "pending", "running", "completed"
	StartTime   time.Time   `json:"startTime"`
	EndTime     time.Time   `json:"endTime"`
}

// ShufflePartition represents data partitioned by key
type ShufflePartition struct {
	ReducerID int        `json:"reducerId"`
	Key       string     `json:"key"`
	Values    []string   `json:"values"`
}

// ReduceTask represents a reduce task on a worker
type ReduceTask struct {
	WorkerID   int       `json:"workerId"`
	Key        string    `json:"key"`
	Values     []string  `json:"values"`
	Result     string    `json:"result"`
	Status     string    `json:"status"` // "pending", "running", "completed"
	StartTime  time.Time `json:"startTime"`
	EndTime    time.Time `json:"endTime"`
}

// JobState represents the complete state of a MapReduce job
type JobState struct {
	JobID            string              `json:"jobId"`
	Status           JobStatus           `json:"status"`
	InputData        []string            `json:"inputData"`
	NumMappers       int                 `json:"numMappers"`
	NumReducers      int                 `json:"numReducers"`
	MapTasks         []MapTask           `json:"mapTasks"`
	ShuffleData      []ShufflePartition  `json:"shuffleData"`
	ReduceTasks      []ReduceTask        `json:"reduceTasks"`
	FinalOutput      []KeyValue          `json:"finalOutput"`
	CurrentStage     string              `json:"currentStage"`
	Progress         int                 `json:"progress"` // 0-100
	StartTime        time.Time           `json:"startTime"`
	EndTime          time.Time           `json:"endTime"`
}

