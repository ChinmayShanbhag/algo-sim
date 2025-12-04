package two_phase_commit

// ParticipantState represents the current state of a participant node
type ParticipantState string

const (
	StateIdle      ParticipantState = "idle"       // Not involved in any transaction
	StatePrepared  ParticipantState = "prepared"   // Voted YES, ready to commit
	StateAborted   ParticipantState = "aborted"    // Transaction aborted
	StateCommitted ParticipantState = "committed"  // Transaction committed
	StateFailed    ParticipantState = "failed"     // Node has failed
)

// VoteResponse represents a participant's vote
type VoteResponse string

const (
	VoteYes VoteResponse = "YES"  // Participant is ready to commit
	VoteNo  VoteResponse = "NO"   // Participant cannot commit
)

// Participant represents a node participating in a 2PC transaction
type Participant struct {
	ID              int              `json:"id"`
	State           ParticipantState `json:"state"`
	Vote            *VoteResponse    `json:"vote,omitempty"`           // The vote this participant cast
	TransactionID   *string          `json:"transactionId,omitempty"`  // Current transaction ID
	CanCommit       bool             `json:"canCommit"`                // Whether this participant can commit
	IsFailed        bool             `json:"isFailed"`                 // Simulated failure
}

// NewParticipant creates a new participant node
// Parameters:
//   - id: Unique identifier for this participant
// Returns a pointer to a new Participant initialized to idle state
func NewParticipant(id int) *Participant {
	return &Participant{
		ID:        id,
		State:     StateIdle,
		Vote:      nil,
		CanCommit: true,  // By default, participants can commit
		IsFailed:  false,
	}
}

// Prepare handles the prepare phase request from coordinator
// This is Phase 1 of 2PC where the participant votes YES or NO
// Parameters:
//   - transactionID: The ID of the transaction to prepare
// Returns the participant's vote (YES if can commit, NO otherwise)
func (p *Participant) Prepare(transactionID string) VoteResponse {
	// If node has failed, it cannot respond
	if p.IsFailed {
		return VoteNo
	}
	
	// Store the transaction ID
	p.TransactionID = &transactionID
	
	// Vote based on whether we can commit
	var vote VoteResponse
	if p.CanCommit {
		vote = VoteYes
		p.State = StatePrepared  // Move to prepared state
	} else {
		vote = VoteNo
		p.State = StateAborted   // Cannot commit, abort
	}
	
	p.Vote = &vote
	return vote
}

// Commit executes the commit phase
// This is Phase 2 of 2PC when coordinator decides to commit
func (p *Participant) Commit() {
	// Only commit if we're in prepared state
	if p.State == StatePrepared && !p.IsFailed {
		p.State = StateCommitted
	}
}

// Abort executes the abort phase
// This is Phase 2 of 2PC when coordinator decides to abort
func (p *Participant) Abort() {
	// Can abort from any state except committed
	if p.State != StateCommitted && !p.IsFailed {
		p.State = StateAborted
	}
}

// Reset resets the participant to initial state
func (p *Participant) Reset() {
	p.State = StateIdle
	p.Vote = nil
	p.TransactionID = nil
	p.CanCommit = true
	p.IsFailed = false
}

// SetCanCommit sets whether this participant can commit
// Used for simulating scenarios where a participant votes NO
func (p *Participant) SetCanCommit(canCommit bool) {
	p.CanCommit = canCommit
}

// SetFailed simulates a node failure
func (p *Participant) SetFailed(failed bool) {
	p.IsFailed = failed
	if failed {
		p.State = StateFailed
	} else if p.State == StateFailed {
		p.State = StateIdle
	}
}

