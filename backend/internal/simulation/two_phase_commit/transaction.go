package two_phase_commit

import (
	"fmt"
	"time"
)

// TransactionState represents the current state of a transaction
type TransactionState string

const (
	TxStateInitiated TransactionState = "initiated" // Transaction started
	TxStatePreparing TransactionState = "preparing" // Sending prepare requests
	TxStateAborted   TransactionState = "aborted"   // Transaction aborted
	TxStateCommitted TransactionState = "committed" // Transaction committed
)

// Transaction represents a distributed transaction in 2PC
type Transaction struct {
	ID          string           `json:"id"`
	State       TransactionState `json:"state"`
	Data        string           `json:"data"`        // The data being committed
	StartTime   time.Time        `json:"startTime"`
	EndTime     *time.Time       `json:"endTime,omitempty"`
	YesVotes    int              `json:"yesVotes"`    // Count of YES votes
	NoVotes     int              `json:"noVotes"`     // Count of NO votes
	TotalVotes  int              `json:"totalVotes"`  // Total participants
}

// NewTransaction creates a new transaction
// Parameters:
//   - id: Unique identifier for the transaction
//   - data: The data/operation to be committed
//   - participantCount: Number of participants in this transaction
// Returns a pointer to a new Transaction
func NewTransaction(id string, data string, participantCount int) *Transaction {
	return &Transaction{
		ID:         id,
		State:      TxStateInitiated,
		Data:       data,
		StartTime:  time.Now(),
		YesVotes:   0,
		NoVotes:    0,
		TotalVotes: participantCount,
	}
}

// RecordVote records a vote from a participant
// Parameters:
//   - vote: The vote (YES or NO) from a participant
func (t *Transaction) RecordVote(vote VoteResponse) {
	if vote == VoteYes {
		t.YesVotes++
	} else {
		t.NoVotes++
	}
}

// CanCommit checks if the transaction can be committed
// In 2PC, ALL participants must vote YES for commit to proceed
// Returns true if all votes are YES, false otherwise
func (t *Transaction) CanCommit() bool {
	// All participants must vote YES
	return t.YesVotes == t.TotalVotes && t.NoVotes == 0
}

// HasAllVotes checks if all participants have voted
func (t *Transaction) HasAllVotes() bool {
	return (t.YesVotes + t.NoVotes) == t.TotalVotes
}

// Commit marks the transaction as committed
func (t *Transaction) Commit() {
	t.State = TxStateCommitted
	now := time.Now()
	t.EndTime = &now
}

// Abort marks the transaction as aborted
func (t *Transaction) Abort() {
	t.State = TxStateAborted
	now := time.Now()
	t.EndTime = &now
}

// GetResult returns a human-readable result of the transaction
func (t *Transaction) GetResult() string {
	switch t.State {
	case TxStateCommitted:
		return fmt.Sprintf("Transaction %s COMMITTED successfully (All %d participants voted YES)", t.ID, t.TotalVotes)
	case TxStateAborted:
		return fmt.Sprintf("Transaction %s ABORTED (%d YES, %d NO votes)", t.ID, t.YesVotes, t.NoVotes)
	case TxStatePreparing:
		return fmt.Sprintf("Transaction %s in progress (%d/%d votes received)", t.ID, t.YesVotes+t.NoVotes, t.TotalVotes)
	default:
		return fmt.Sprintf("Transaction %s initiated", t.ID)
	}
}

