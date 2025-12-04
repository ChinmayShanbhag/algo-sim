package three_phase_commit

// ParticipantState represents the current state of a participant node
type ParticipantState string

const (
	StateIdle         ParticipantState = "idle"          // Not involved in any transaction
	StateUncertain    ParticipantState = "uncertain"     // Voted YES, waiting for pre-commit
	StatePreCommitted ParticipantState = "pre_committed" // Ready to commit (can timeout and commit)
	StateAborted      ParticipantState = "aborted"       // Transaction aborted
	StateCommitted    ParticipantState = "committed"     // Transaction committed
	StateFailed       ParticipantState = "failed"        // Node has failed
)

// VoteResponse represents a participant's vote
type VoteResponse string

const (
	VoteYes VoteResponse = "YES" // Participant is ready to commit
	VoteNo  VoteResponse = "NO"  // Participant cannot commit
)

// Participant represents a node participating in a 3PC transaction
type Participant struct {
	ID            int              `json:"id"`
	State         ParticipantState `json:"state"`
	Vote          *VoteResponse    `json:"vote,omitempty"`          // The vote this participant cast
	TransactionID *string          `json:"transactionId,omitempty"` // Current transaction ID
	CanCommit     bool             `json:"canCommit"`               // Whether this participant can commit
	IsFailed      bool             `json:"isFailed"`                // Simulated failure
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
		CanCommit: true, // By default, participants can commit
		IsFailed:  false,
	}
}

// CanCommitPhase handles the can-commit phase request from coordinator
// This is Phase 1 of 3PC where the participant votes YES or NO
// Parameters:
//   - transactionID: The ID of the transaction to prepare
// Returns the participant's vote (YES if can commit, NO otherwise)
func (p *Participant) CanCommitPhase(transactionID string) VoteResponse {
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
		p.State = StateUncertain // Move to uncertain state (waiting for pre-commit)
	} else {
		vote = VoteNo
		p.State = StateAborted // Cannot commit, abort
	}

	p.Vote = &vote
	return vote
}

// PreCommit executes the pre-commit phase
// This is Phase 2 of 3PC - participant is now ready to commit
// Key difference from 2PC: In this state, participant can timeout and commit autonomously
func (p *Participant) PreCommit() {
	// Only pre-commit if we're in uncertain state
	if p.State == StateUncertain && !p.IsFailed {
		p.State = StatePreCommitted
	}
}

// Commit executes the commit phase
// This is Phase 3 of 3PC when coordinator sends final commit
func (p *Participant) Commit() {
	// Can commit from pre-committed state
	if p.State == StatePreCommitted && !p.IsFailed {
		p.State = StateCommitted
	}
}

// Abort executes the abort phase
// Can abort from uncertain state (before pre-commit)
func (p *Participant) Abort() {
	// Can abort from uncertain state, but NOT from pre-committed
	// This is a key property of 3PC - once pre-committed, must commit
	if p.State == StateUncertain && !p.IsFailed {
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

