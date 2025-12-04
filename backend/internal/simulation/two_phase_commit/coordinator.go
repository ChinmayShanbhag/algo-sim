package two_phase_commit

import (
	"fmt"
	"sync"
)

// CoordinatorState represents the state of the coordinator
type CoordinatorState string

const (
	CoordStateIdle      CoordinatorState = "idle"      // No active transaction
	CoordStatePreparing CoordinatorState = "preparing" // Phase 1: Sending prepare requests
	CoordStateCommitting CoordinatorState = "committing" // Phase 2: Sending commit
	CoordStateAborting  CoordinatorState = "aborting"  // Phase 2: Sending abort
	CoordStateFailed    CoordinatorState = "failed"    // Coordinator has failed
)

// ProtocolStep represents a single step in the 2PC protocol
// This is used for step-by-step visualization
type ProtocolStep struct {
	StepNumber   int              `json:"stepNumber"`
	Description  string           `json:"description"`
	Action       string           `json:"action"`
	FromNode     *int             `json:"fromNode,omitempty"`     // -1 represents coordinator
	ToNode       *int             `json:"toNode,omitempty"`
	MessageType  string           `json:"messageType,omitempty"`  // "prepare", "vote", "commit", "abort"
	VoteResponse *VoteResponse    `json:"voteResponse,omitempty"` // YES or NO
	YesVotes     int              `json:"yesVotes"`
	NoVotes      int              `json:"noVotes"`
}

// Coordinator manages the Two-Phase Commit protocol
type Coordinator struct {
	mu            sync.RWMutex
	State         CoordinatorState `json:"state"`
	Participants  []*Participant   `json:"participants"`
	Transaction   *Transaction     `json:"transaction,omitempty"`
	ProtocolSteps []ProtocolStep   `json:"protocolSteps,omitempty"`
	IsFailed      bool             `json:"isFailed"`
}

// NewCoordinator creates a new coordinator with the specified number of participants
// Parameters:
//   - participantCount: Number of participant nodes to create
// Returns a pointer to a new Coordinator
func NewCoordinator(participantCount int) *Coordinator {
	participants := make([]*Participant, participantCount)
	for i := 0; i < participantCount; i++ {
		participants[i] = NewParticipant(i)
	}
	
	return &Coordinator{
		State:        CoordStateIdle,
		Participants: participants,
		IsFailed:     false,
	}
}

// StartTransaction initiates a new 2PC transaction with step-by-step tracking
// Parameters:
//   - transactionID: Unique ID for the transaction
//   - data: The data/operation to commit
// Returns the protocol steps for visualization
func (c *Coordinator) StartTransaction(transactionID string, data string) ([]ProtocolStep, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Check if coordinator has failed
	if c.IsFailed {
		return nil, fmt.Errorf("coordinator has failed")
	}
	
	// Reset protocol steps
	c.ProtocolSteps = []ProtocolStep{}
	
	// Create new transaction
	c.Transaction = NewTransaction(transactionID, data, len(c.Participants))
	c.State = CoordStatePreparing
	
	// Step 1: Coordinator initiates transaction
	coordinatorID := -1  // Use -1 to represent coordinator
	c.ProtocolSteps = append(c.ProtocolSteps, ProtocolStep{
		StepNumber:  1,
		Description: fmt.Sprintf("Coordinator initiates transaction '%s' with data: '%s'", transactionID, data),
		Action:      "transaction_initiated",
		FromNode:    &coordinatorID,
		YesVotes:    0,
		NoVotes:     0,
	})
	
	// Phase 1: PREPARE - Send prepare requests to all participants
	yesVotes := 0
	noVotes := 0
	
	for i, participant := range c.Participants {
		// Step: Send prepare request
		targetNode := i
		c.ProtocolSteps = append(c.ProtocolSteps, ProtocolStep{
			StepNumber:  len(c.ProtocolSteps) + 1,
			Description: fmt.Sprintf("Coordinator sends PREPARE request to Participant %d", i),
			Action:      "prepare_request_sent",
			FromNode:    &coordinatorID,
			ToNode:      &targetNode,
			MessageType: "prepare",
			YesVotes:    yesVotes,
			NoVotes:     noVotes,
		})
		
		// Participant votes
		vote := participant.Prepare(transactionID)
		c.Transaction.RecordVote(vote)
		
		if vote == VoteYes {
			yesVotes++
		} else {
			noVotes++
		}
		
		// Step: Receive vote response
		responseFrom := i
		c.ProtocolSteps = append(c.ProtocolSteps, ProtocolStep{
			StepNumber:   len(c.ProtocolSteps) + 1,
			Description:  fmt.Sprintf("Participant %d votes %s", i, vote),
			Action:       "vote_received",
			FromNode:     &responseFrom,
			ToNode:       &coordinatorID,
			MessageType:  "vote",
			VoteResponse: &vote,
			YesVotes:     yesVotes,
			NoVotes:      noVotes,
		})
	}
	
	// Phase 2: COMMIT or ABORT based on votes
	decision := c.Transaction.CanCommit()
	
	if decision {
		// All voted YES - COMMIT
		c.State = CoordStateCommitting
		c.Transaction.Commit()
		
		// Step: Coordinator decides to commit
		c.ProtocolSteps = append(c.ProtocolSteps, ProtocolStep{
			StepNumber:  len(c.ProtocolSteps) + 1,
			Description: fmt.Sprintf("Coordinator decides to COMMIT (All %d participants voted YES)", len(c.Participants)),
			Action:      "decision_commit",
			FromNode:    &coordinatorID,
			YesVotes:    yesVotes,
			NoVotes:     noVotes,
		})
		
		// Send commit to all participants
		for i, participant := range c.Participants {
			targetNode := i
			c.ProtocolSteps = append(c.ProtocolSteps, ProtocolStep{
				StepNumber:  len(c.ProtocolSteps) + 1,
				Description: fmt.Sprintf("Coordinator sends COMMIT to Participant %d", i),
				Action:      "commit_sent",
				FromNode:    &coordinatorID,
				ToNode:      &targetNode,
				MessageType: "commit",
				YesVotes:    yesVotes,
				NoVotes:     noVotes,
			})
			
			participant.Commit()
			
			// Acknowledgment
			responseFrom := i
			c.ProtocolSteps = append(c.ProtocolSteps, ProtocolStep{
				StepNumber:  len(c.ProtocolSteps) + 1,
				Description: fmt.Sprintf("Participant %d acknowledges COMMIT", i),
				Action:      "commit_ack",
				FromNode:    &responseFrom,
				ToNode:      &coordinatorID,
				MessageType: "ack",
				YesVotes:    yesVotes,
				NoVotes:     noVotes,
			})
		}
		
		// Final step
		c.ProtocolSteps = append(c.ProtocolSteps, ProtocolStep{
			StepNumber:  len(c.ProtocolSteps) + 1,
			Description: fmt.Sprintf("Transaction '%s' COMMITTED successfully!", transactionID),
			Action:      "transaction_committed",
			YesVotes:    yesVotes,
			NoVotes:     noVotes,
		})
		
		c.State = CoordStateIdle
	} else {
		// At least one voted NO - ABORT
		c.State = CoordStateAborting
		c.Transaction.Abort()
		
		// Step: Coordinator decides to abort
		c.ProtocolSteps = append(c.ProtocolSteps, ProtocolStep{
			StepNumber:  len(c.ProtocolSteps) + 1,
			Description: fmt.Sprintf("Coordinator decides to ABORT (%d YES, %d NO votes)", yesVotes, noVotes),
			Action:      "decision_abort",
			FromNode:    &coordinatorID,
			YesVotes:    yesVotes,
			NoVotes:     noVotes,
		})
		
		// Send abort to all participants
		for i, participant := range c.Participants {
			targetNode := i
			c.ProtocolSteps = append(c.ProtocolSteps, ProtocolStep{
				StepNumber:  len(c.ProtocolSteps) + 1,
				Description: fmt.Sprintf("Coordinator sends ABORT to Participant %d", i),
				Action:      "abort_sent",
				FromNode:    &coordinatorID,
				ToNode:      &targetNode,
				MessageType: "abort",
				YesVotes:    yesVotes,
				NoVotes:     noVotes,
			})
			
			participant.Abort()
			
			// Acknowledgment
			responseFrom := i
			c.ProtocolSteps = append(c.ProtocolSteps, ProtocolStep{
				StepNumber:  len(c.ProtocolSteps) + 1,
				Description: fmt.Sprintf("Participant %d acknowledges ABORT", i),
				Action:      "abort_ack",
				FromNode:    &responseFrom,
				ToNode:      &coordinatorID,
				MessageType: "ack",
				YesVotes:    yesVotes,
				NoVotes:     noVotes,
			})
		}
		
		// Final step
		c.ProtocolSteps = append(c.ProtocolSteps, ProtocolStep{
			StepNumber:  len(c.ProtocolSteps) + 1,
			Description: fmt.Sprintf("Transaction '%s' ABORTED", transactionID),
			Action:      "transaction_aborted",
			YesVotes:    yesVotes,
			NoVotes:     noVotes,
		})
		
		c.State = CoordStateIdle
	}
	
	return c.ProtocolSteps, nil
}

// Reset resets the coordinator and all participants to initial state
func (c *Coordinator) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.State = CoordStateIdle
	c.Transaction = nil
	c.ProtocolSteps = []ProtocolStep{}
	c.IsFailed = false
	
	for _, participant := range c.Participants {
		participant.Reset()
	}
}

// SetParticipantCanCommit sets whether a specific participant can commit
// Used for simulating scenarios where a participant votes NO
func (c *Coordinator) SetParticipantCanCommit(participantID int, canCommit bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if participantID < 0 || participantID >= len(c.Participants) {
		return fmt.Errorf("invalid participant ID: %d", participantID)
	}
	
	c.Participants[participantID].SetCanCommit(canCommit)
	return nil
}

// SetParticipantFailed simulates a participant failure
func (c *Coordinator) SetParticipantFailed(participantID int, failed bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if participantID < 0 || participantID >= len(c.Participants) {
		return fmt.Errorf("invalid participant ID: %d", participantID)
	}
	
	c.Participants[participantID].SetFailed(failed)
	return nil
}

// SetCoordinatorFailed simulates coordinator failure
func (c *Coordinator) SetCoordinatorFailed(failed bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.IsFailed = failed
	if failed {
		c.State = CoordStateFailed
	} else if c.State == CoordStateFailed {
		c.State = CoordStateIdle
	}
}

