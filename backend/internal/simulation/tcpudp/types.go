package tcpudp

import "time"

// PacketStatus represents the status of a packet
type PacketStatus string

const (
	PacketPending     PacketStatus = "pending"
	PacketInTransit   PacketStatus = "in_transit"
	PacketDelivered   PacketStatus = "delivered"
	PacketLost        PacketStatus = "lost"
	PacketAcknowledged PacketStatus = "acknowledged"
)

// Packet represents a data packet
type Packet struct {
	ID          int          `json:"id"`
	SeqNumber   int          `json:"seqNumber"`
	Data        string       `json:"data"`
	Status      PacketStatus `json:"status"`
	Protocol    string       `json:"protocol"` // "TCP" or "UDP"
	SentTime    time.Time    `json:"sentTime"`
	AckTime     *time.Time   `json:"ackTime,omitempty"`
	IsLost      bool         `json:"isLost"`
	RetryCount  int          `json:"retryCount"`
}

// ConnectionState represents TCP connection state
type ConnectionState struct {
	IsConnected    bool   `json:"isConnected"`
	SynSent        bool   `json:"synSent"`
	SynAckReceived bool   `json:"synAckReceived"`
	AckSent        bool   `json:"ackSent"`
	FinSent        bool   `json:"finSent"`
	Closed         bool   `json:"closed"`
}

// SimulationState represents the complete state of both protocols
type SimulationState struct {
	// TCP State
	TCPPackets      []Packet        `json:"tcpPackets"`
	TCPConnection   ConnectionState `json:"tcpConnection"`
	TCPSeqNumber    int             `json:"tcpSeqNumber"`
	TCPInFlight     int             `json:"tcpInFlight"`
	
	// UDP State
	UDPPackets      []Packet        `json:"udpPackets"`
	UDPSeqNumber    int             `json:"udpSeqNumber"`
	
	// Statistics
	Stats           Statistics      `json:"stats"`
}

// Statistics tracks metrics for both protocols
type Statistics struct {
	TCP TCPStats `json:"tcp"`
	UDP UDPStats `json:"udp"`
}

// TCPStats tracks TCP-specific metrics
type TCPStats struct {
	PacketsSent       int     `json:"packetsSent"`
	PacketsDelivered  int     `json:"packetsDelivered"`
	PacketsRetried    int     `json:"packetsRetried"`
	PacketsLost       int     `json:"packetsLost"`        // Packets that failed after max retries
	Acknowledgments   int     `json:"acknowledgments"`
	AverageLatencyMs  float64 `json:"averageLatencyMs"`
	DeliveryRate      float64 `json:"deliveryRate"`
}

// UDPStats tracks UDP-specific metrics
type UDPStats struct {
	PacketsSent      int     `json:"packetsSent"`
	PacketsDelivered int     `json:"packetsDelivered"`
	PacketsLost      int     `json:"packetsLost"`
	AverageLatencyMs float64 `json:"averageLatencyMs"`
	DeliveryRate     float64 `json:"deliveryRate"`
}

