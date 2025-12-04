package raft

import (
	"encoding/json"
	"fmt"
	"sync"
)

// ElectionStep represents a step in the election process
type ElectionStep struct {
	StepNumber  int      `json:"stepNumber"`
	Description string   `json:"description"`
	Action      string   `json:"action"`
	Votes       int      `json:"votes"`
	VotedNodes  []int    `json:"votedNodes"`
	FromNode    *int     `json:"fromNode,omitempty"` // Node sending message
	ToNode      *int     `json:"toNode,omitempty"`   // Node receiving message
	MessageType string   `json:"messageType,omitempty"` // "vote_request", "vote_response", "heartbeat"
}

// Cluster represents a Raft cluster with multiple nodes
type Cluster struct {
	mu           sync.RWMutex
	Nodes        []*Node        `json:"nodes"`
	ElectionSteps []ElectionStep `json:"electionSteps,omitempty"`
}

// NewCluster creates a new Raft cluster with the specified number of nodes
func NewCluster(nodeCount int) *Cluster {
	nodes := make([]*Node, nodeCount)
	for i := 0; i < nodeCount; i++ {
		nodes[i] = NewNode(i)
	}
	return &Cluster{
		Nodes: nodes,
	}
}

// GetState returns the current state of the cluster (thread-safe)
func (c *Cluster) GetState() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return json.Marshal(c)
}

// GetNode returns a specific node by ID (thread-safe)
func (c *Cluster) GetNode(id int) *Node {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if id >= 0 && id < len(c.Nodes) {
		return c.Nodes[id]
	}
	return nil
}

// StartElectionStepByStep simulates a node starting a leader election step by step
func (c *Cluster) StartElectionStepByStep(nodeID int) ([]ElectionStep, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Clear previous steps
	c.ElectionSteps = []ElectionStep{}
	
	// Validate node ID
	if nodeID < 0 || nodeID >= len(c.Nodes) {
		return nil, nil // Invalid node ID, ignore
	}
	
	candidate := c.Nodes[nodeID]
	majority := (len(c.Nodes) / 2) + 1
	
	// Step 1: Node becomes candidate
	candidate.State = StateCandidate
	candidate.CurrentTerm++
	candidate.VotedFor = &nodeID // Vote for itself
	
	c.ElectionSteps = append(c.ElectionSteps, ElectionStep{
		StepNumber:  1,
		Description: fmt.Sprintf("Node %d timeout: No heartbeat from leader. Becoming Candidate", nodeID),
		Action:      "increment_term_and_vote_self",
		Votes:       1,
		VotedNodes:  []int{nodeID},
		FromNode:    &nodeID,
	})
	
	// Step 2: Request votes from other nodes
	votes := 1
	votedNodes := []int{nodeID}
	
		for i, node := range c.Nodes {
		if i == nodeID {
			continue // Skip self
		}
		
		// Step: Send vote request
		stepNum := len(c.ElectionSteps) + 1
		targetNode := i // Copy to avoid pointer issues
		c.ElectionSteps = append(c.ElectionSteps, ElectionStep{
			StepNumber:  stepNum,
			Description: fmt.Sprintf("Vote request sent from Node %d to Node %d", nodeID, i),
			Action:      "vote_request_sent",
			Votes:       votes,
			VotedNodes:  append([]int{}, votedNodes...),
			FromNode:    &nodeID,
			ToNode:      &targetNode,
			MessageType: "vote_request",
		})
		
		// Node votes yes if:
		// 1. It hasn't voted this term, OR
		// 2. Candidate's term is higher
		if node.VotedFor == nil || candidate.CurrentTerm > node.CurrentTerm {
			node.VotedFor = &nodeID
			node.CurrentTerm = candidate.CurrentTerm
			node.State = StateFollower // Reset to follower
			votes++
			votedNodes = append(votedNodes, i)
			
			responseFrom := i // Copy to avoid pointer issues
			c.ElectionSteps = append(c.ElectionSteps, ElectionStep{
				StepNumber:  len(c.ElectionSteps) + 1,
				Description: fmt.Sprintf("Node %d votes YES for candidate", i),
				Action:      "vote_received",
				Votes:       votes,
				VotedNodes:  append([]int{}, votedNodes...),
				FromNode:    &responseFrom,
				ToNode:      &nodeID,
				MessageType: "vote_response",
			})
		} else {
			responseFrom := i // Copy to avoid pointer issues
			c.ElectionSteps = append(c.ElectionSteps, ElectionStep{
				StepNumber:  len(c.ElectionSteps) + 1,
				Description: fmt.Sprintf("Node %d votes NO (already voted this term)", i),
				Action:      "vote_rejected",
				Votes:       votes,
				VotedNodes:  append([]int{}, votedNodes...),
				FromNode:    &responseFrom,
				ToNode:      &nodeID,
				MessageType: "vote_response",
			})
		}
	}
	
	// Step 3: Check if majority achieved
	if votes >= majority {
		candidate.State = StateLeader
		// All other nodes become followers
		for i, node := range c.Nodes {
			if i != nodeID {
				node.State = StateFollower
				node.CurrentTerm = candidate.CurrentTerm
			}
		}
		
		c.ElectionSteps = append(c.ElectionSteps, ElectionStep{
			StepNumber:  len(c.ElectionSteps) + 1,
			Description: "Majority achieved! Candidate becomes Leader",
			Action:      "election_success",
			Votes:       votes,
			VotedNodes:  votedNodes,
		})
	} else {
		// Election failed, become follower
		candidate.State = StateFollower
		candidate.VotedFor = nil
		
		c.ElectionSteps = append(c.ElectionSteps, ElectionStep{
			StepNumber:  len(c.ElectionSteps) + 1,
			Description: "No majority. Election failed. Candidate becomes Follower",
			Action:      "election_failed",
			Votes:       votes,
			VotedNodes:  votedNodes,
		})
	}
	
	return c.ElectionSteps, nil
}

// GetStateAtStep returns the cluster state at a specific step of the election
func (c *Cluster) GetStateAtStep(stepNumber int) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if stepNumber < 0 || stepNumber > len(c.ElectionSteps) {
		return nil, fmt.Errorf("invalid step number")
	}
	
	// Create a copy of nodes for this step state
	type StepState struct {
		Nodes        []*Node        `json:"nodes"`
		CurrentStep  int            `json:"currentStep"`
		TotalSteps   int            `json:"totalSteps"`
		Step         *ElectionStep  `json:"step"`
	}
	
	// Apply steps up to stepNumber
	// For now, return the current state with step info
	// In a real implementation, we'd replay the steps
	
	stepState := StepState{
		Nodes:       c.Nodes,
		CurrentStep: stepNumber,
		TotalSteps:  len(c.ElectionSteps),
	}
	
	if stepNumber > 0 && stepNumber <= len(c.ElectionSteps) {
		stepState.Step = &c.ElectionSteps[stepNumber-1]
	}
	
	return json.Marshal(stepState)
}

// StartElection simulates a node starting a leader election (all at once, for backward compatibility)
func (c *Cluster) StartElection(nodeID int) error {
	_, err := c.StartElectionStepByStep(nodeID)
	return err
}

// Reset resets all nodes to initial follower state
func (c *Cluster) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for _, node := range c.Nodes {
		node.State = StateFollower
		node.CurrentTerm = 0
		node.VotedFor = nil
	}
	c.ElectionSteps = []ElectionStep{}
}

// SetLeader sets a specific node as leader (for simulation purposes)
func (c *Cluster) SetLeader(nodeID int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if nodeID < 0 || nodeID >= len(c.Nodes) {
		return fmt.Errorf("invalid node ID")
	}
	
	// Reset all nodes first
	for _, node := range c.Nodes {
		node.State = StateFollower
		node.CurrentTerm = 1
		node.VotedFor = nil
	}
	
	// Set the specified node as leader
	leader := c.Nodes[nodeID]
	leader.State = StateLeader
	leader.CurrentTerm = 1
	leader.VotedFor = &nodeID
	
	return nil
}

