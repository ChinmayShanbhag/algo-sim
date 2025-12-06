package tcpudp

import (
	"encoding/json"
	"net/http"
	"strconv"

	"sds/internal/session"
)

var sessionManager *session.Manager

// getSessionID extracts the session ID from the request header
func getSessionID(r *http.Request) string {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID != "" {
		return sessionID
	}
	sessionID = r.URL.Query().Get("session_id")
	if sessionID != "" {
		return sessionID
	}
	return "default"
}

// setCORSHeaders sets CORS headers for cross-origin requests
func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Session-ID")
}

// GetState returns the current state of the TCP/UDP simulation
// GET /api/tcpudp/state
func GetState(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	state := userState.TCPUDPSimulator.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// SendTCPPacket sends a packet using TCP
// POST /api/tcpudp/send-tcp?data=...
func SendTCPPacket(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	data := r.URL.Query().Get("data")
	if data == "" {
		data = "TCP Data"
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	userState.TCPUDPSimulator.SendTCPPacket(data)
	state := userState.TCPUDPSimulator.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// SendUDPPacket sends a packet using UDP
// POST /api/tcpudp/send-udp?data=...
func SendUDPPacket(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	data := r.URL.Query().Get("data")
	if data == "" {
		data = "UDP Data"
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	userState.TCPUDPSimulator.SendUDPPacket(data)
	state := userState.TCPUDPSimulator.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// ProcessPackets processes packets for both protocols
// POST /api/tcpudp/process
func ProcessPackets(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	userState.TCPUDPSimulator.ProcessTCPPackets()
	userState.TCPUDPSimulator.ProcessUDPPackets()
	state := userState.TCPUDPSimulator.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// SetPacketLossRate sets the packet loss rate
// POST /api/tcpudp/set-loss-rate?rate=...
func SetPacketLossRate(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	rateStr := r.URL.Query().Get("rate")
	rate, err := strconv.ParseFloat(rateStr, 64)
	if err != nil {
		http.Error(w, "invalid rate parameter", http.StatusBadRequest)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	userState.TCPUDPSimulator.SetPacketLossRate(rate)
	state := userState.TCPUDPSimulator.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// Reset resets the simulation
// POST /api/tcpudp/reset
func Reset(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	userState.TCPUDPSimulator.Reset()
	state := userState.TCPUDPSimulator.GetState()

	responseJSON, err := json.Marshal(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// EstablishConnection establishes TCP connection (3-way handshake)
// POST /api/tcpudp/establish-connection
func EstablishConnection(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	steps := userState.TCPUDPSimulator.EstablishTCPConnection()
	state := userState.TCPUDPSimulator.GetState()

	response := map[string]interface{}{
		"state": state,
		"steps": steps,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// CloseConnection closes TCP connection
// POST /api/tcpudp/close-connection
func CloseConnection(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionID := getSessionID(r)
	userState := sessionManager.GetOrCreate(sessionID)

	steps := userState.TCPUDPSimulator.CloseTCPConnection()
	state := userState.TCPUDPSimulator.GetState()

	response := map[string]interface{}{
		"state": state,
		"steps": steps,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// SetupRoutes registers all TCP/UDP endpoints
func SetupRoutes(sm *session.Manager) {
	sessionManager = sm

	http.HandleFunc("/api/tcpudp/state", GetState)
	http.HandleFunc("/api/tcpudp/send-tcp", SendTCPPacket)
	http.HandleFunc("/api/tcpudp/send-udp", SendUDPPacket)
	http.HandleFunc("/api/tcpudp/process", ProcessPackets)
	http.HandleFunc("/api/tcpudp/set-loss-rate", SetPacketLossRate)
	http.HandleFunc("/api/tcpudp/reset", Reset)
	http.HandleFunc("/api/tcpudp/establish-connection", EstablishConnection)
	http.HandleFunc("/api/tcpudp/close-connection", CloseConnection)
}

