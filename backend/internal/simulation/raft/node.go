package raft

// NodeState represents the possible states of a Raft node
type NodeState string

const (
	StateFollower  NodeState = "follower"
	StateCandidate NodeState = "candidate"
	StateLeader    NodeState = "leader"
)

// Node represents a single Raft node in the cluster
type Node struct {
	ID          int       `json:"id"`
	State       NodeState `json:"state"`
	CurrentTerm int       `json:"currentTerm"`
	VotedFor    *int      `json:"votedFor"` // nil if hasn't voted this term
	LastHeartbeat int     `json:"lastHeartbeat"` // Simulated timestamp
}

// NewNode creates a new Raft node
func NewNode(id int) *Node {
	return &Node{
		ID:            id,
		State:         StateFollower,
		CurrentTerm:   0,
		VotedFor:      nil,
		LastHeartbeat: 0,
	}
}

