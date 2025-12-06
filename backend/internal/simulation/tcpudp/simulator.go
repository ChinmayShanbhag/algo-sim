package tcpudp

import (
	"math/rand"
	"sync"
	"time"
)

// Simulator simulates TCP and UDP packet transmission
type Simulator struct {
	mu sync.RWMutex
	
	// TCP state
	tcpPackets    []Packet
	tcpConnection ConnectionState
	tcpSeqNumber  int
	tcpNextID     int
	
	// UDP state
	udpPackets    []Packet
	udpSeqNumber  int
	udpNextID     int
	
	// Simulation parameters
	packetLossRate float64 // Probability of packet loss (0.0 - 1.0)
	latencyMs      int     // Simulated network latency in ms
	
	// Statistics
	tcpStats TCPStats
	udpStats UDPStats
}

// NewSimulator creates a new TCP/UDP simulator
func NewSimulator() *Simulator {
	return &Simulator{
		tcpPackets:     []Packet{},
		udpPackets:     []Packet{},
		tcpConnection:  ConnectionState{},
		tcpSeqNumber:   1000,
		udpSeqNumber:   1000,
		tcpNextID:      1,
		udpNextID:      1,
		packetLossRate: 0.2, // 20% packet loss rate for demonstration
		latencyMs:      100,  // 100ms latency
		tcpStats:       TCPStats{},
		udpStats:       UDPStats{},
	}
}

// SetPacketLossRate sets the packet loss rate (0.0 - 1.0)
func (s *Simulator) SetPacketLossRate(rate float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	s.packetLossRate = rate
}

// SendTCPPacket sends a packet using TCP (with reliability)
// Returns the packet in Pending state
func (s *Simulator) SendTCPPacket(data string) Packet {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	packet := Packet{
		ID:         s.tcpNextID,
		SeqNumber:  s.tcpSeqNumber,
		Data:       data,
		Status:     PacketPending,
		Protocol:   "TCP",
		SentTime:   time.Now(),
		IsLost:     false,
		RetryCount: 0,
	}
	
	s.tcpNextID++
	s.tcpSeqNumber++
	s.tcpPackets = append(s.tcpPackets, packet)
	s.tcpStats.PacketsSent++
	
	// Keep only last 10 packets for visualization
	if len(s.tcpPackets) > 10 {
		s.tcpPackets = s.tcpPackets[len(s.tcpPackets)-10:]
	}
	
	return packet
}

// SendUDPPacket sends a packet using UDP (no reliability)
// Returns the packet in Pending state
func (s *Simulator) SendUDPPacket(data string) Packet {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	packet := Packet{
		ID:        s.udpNextID,
		SeqNumber: s.udpSeqNumber,
		Data:      data,
		Status:    PacketPending,
		Protocol:  "UDP",
		SentTime:  time.Now(),
		IsLost:    false,
	}
	
	s.udpNextID++
	s.udpSeqNumber++
	s.udpPackets = append(s.udpPackets, packet)
	s.udpStats.PacketsSent++
	
	// Keep only last 10 packets for visualization
	if len(s.udpPackets) > 10 {
		s.udpPackets = s.udpPackets[len(s.udpPackets)-10:]
	}
	
	return packet
}

// ProcessTCPPackets simulates COMPLETE TCP packet transmission with ACKs and retries
// Processes each packet to completion in a single call
// TCP retries up to 3 times before giving up (in real world, this would be more)
func (s *Simulator) ProcessTCPPackets() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	const maxRetries = 3 // TCP will retry up to 3 times
	
	for i := range s.tcpPackets {
		packet := &s.tcpPackets[i]
		
		// Skip already completed packets
		if packet.Status == PacketAcknowledged {
			continue
		}
		
		// Process pending packet to completion
		if packet.Status == PacketPending {
			// Step 1: Send packet - check if lost
			isLost := rand.Float64() < s.packetLossRate
			
			// TCP will retry on loss
			for isLost && packet.RetryCount < maxRetries {
				packet.RetryCount++
				s.tcpStats.PacketsRetried++
				
				// Retry - check if retry succeeds
				isLost = rand.Float64() < s.packetLossRate
			}
			
			if isLost {
				// All retries exhausted - TCP gives up (rare but possible)
				packet.Status = PacketLost
				packet.IsLost = true
				s.tcpStats.PacketsLost++
			} else {
				// Successful delivery (either first attempt or after retries)
				if packet.RetryCount > 0 {
					packet.IsLost = true // Mark that it had losses (for visualization)
				}
				
				// Step 2: Packet delivered
				// Step 3: Receiver sends ACK back
				packet.Status = PacketAcknowledged
				now := time.Now()
				packet.AckTime = &now
				
				// Update statistics
				s.tcpStats.PacketsDelivered++
				s.tcpStats.Acknowledgments++
			}
		}
	}
	
	// Calculate delivery rate
	if s.tcpStats.PacketsSent > 0 {
		s.tcpStats.DeliveryRate = float64(s.tcpStats.PacketsDelivered) / float64(s.tcpStats.PacketsSent) * 100
	}
}

// ProcessUDPPackets simulates COMPLETE UDP packet transmission
// Processes each packet to completion (delivered or lost) in a single call
func (s *Simulator) ProcessUDPPackets() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for i := range s.udpPackets {
		packet := &s.udpPackets[i]
		
		// Skip already processed packets
		if packet.Status == PacketDelivered || packet.Status == PacketLost {
			continue
		}
		
		// Process pending packet to completion
		if packet.Status == PacketPending {
			// Check if packet is lost (random based on loss rate)
			isLost := rand.Float64() < s.packetLossRate
			
			if isLost {
				// Packet lost - UDP does NOT retry
				packet.Status = PacketLost
				packet.IsLost = true
				s.udpStats.PacketsLost++
			} else {
				// Packet delivered successfully
				packet.Status = PacketDelivered
				s.udpStats.PacketsDelivered++
			}
		}
	}
	
	// Calculate delivery rate (UDP varies based on network conditions)
	if s.udpStats.PacketsSent > 0 {
		s.udpStats.DeliveryRate = float64(s.udpStats.PacketsDelivered) / float64(s.udpStats.PacketsSent) * 100
	}
}

// EstablishTCPConnection simulates TCP 3-way handshake
func (s *Simulator) EstablishTCPConnection() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	steps := []string{}
	
	if !s.tcpConnection.IsConnected {
		// SYN
		s.tcpConnection.SynSent = true
		steps = append(steps, "SYN sent")
		
		// SYN-ACK
		s.tcpConnection.SynAckReceived = true
		steps = append(steps, "SYN-ACK received")
		
		// ACK
		s.tcpConnection.AckSent = true
		s.tcpConnection.IsConnected = true
		steps = append(steps, "ACK sent - Connection established")
	}
	
	return steps
}

// CloseTCPConnection simulates TCP connection termination
func (s *Simulator) CloseTCPConnection() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	steps := []string{}
	
	if s.tcpConnection.IsConnected {
		s.tcpConnection.FinSent = true
		steps = append(steps, "FIN sent")
		
		s.tcpConnection.IsConnected = false
		s.tcpConnection.Closed = true
		steps = append(steps, "Connection closed")
	}
	
	return steps
}

// GetState returns the current simulation state
func (s *Simulator) GetState() SimulationState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Calculate average latency (simulated)
	s.tcpStats.AverageLatencyMs = float64(s.latencyMs)
	s.udpStats.AverageLatencyMs = float64(s.latencyMs) * 0.8 // UDP is typically faster
	
	return SimulationState{
		TCPPackets:    s.tcpPackets,
		TCPConnection: s.tcpConnection,
		TCPSeqNumber:  s.tcpSeqNumber,
		TCPInFlight:   s.countInFlight(s.tcpPackets),
		UDPPackets:    s.udpPackets,
		UDPSeqNumber:  s.udpSeqNumber,
		Stats: Statistics{
			TCP: s.tcpStats,
			UDP: s.udpStats,
		},
	}
}

// countInFlight counts packets currently in transit
func (s *Simulator) countInFlight(packets []Packet) int {
	count := 0
	for _, p := range packets {
		if p.Status == PacketInTransit {
			count++
		}
	}
	return count
}

// Reset resets the simulation
func (s *Simulator) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.tcpPackets = []Packet{}
	s.udpPackets = []Packet{}
	s.tcpConnection = ConnectionState{}
	s.tcpSeqNumber = 1000
	s.udpSeqNumber = 1000
	s.tcpNextID = 1
	s.udpNextID = 1
	s.tcpStats = TCPStats{}
	s.udpStats = UDPStats{}
}
