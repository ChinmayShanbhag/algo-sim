# System Design Simulator (SDS)

A learning-focused interactive simulator for exploring distributed systems concepts, trade-offs, and behaviors through step-by-step visualizations.

## 🎯 Overview

SDS provides hands-on, visual simulations of complex distributed systems concepts. Each simulation runs in an isolated user session, allowing you to experiment freely without affecting others.

## 🏗️ Architecture

- **Backend**: Go-based simulation engine with multi-user session management
- **Frontend**: Vanilla JavaScript + D3.js for interactive visualizations
- **Session Isolation**: Each user gets their own isolated simulation state
- **Real-time Updates**: Live state synchronization between backend and frontend

## 📂 Project Structure

```
.
├── backend/
│   ├── cmd/server/          # Application entry point
│   ├── internal/
│   │   ├── api/            # HTTP handlers & routes
│   │   ├── simulation/     # Core simulation algorithms
│   │   └── session/        # Multi-user session management
│   └── go.mod
├── frontend/
│   ├── index.html          # Landing page
│   ├── session.js          # Session management
│   └── visualizations/     # Individual feature simulations
└── README.md
```

## 🚀 Getting Started

### Prerequisites
- Go 1.21 or higher
- Python 3 (for serving frontend)
- Modern web browser with JavaScript enabled

### 1. Start the Backend (Port 8080)
```bash
cd backend
go run cmd/server/main.go
```

### 2. Start the Frontend (Port 8000)
```bash
cd frontend
python -m http.server 8000
```

### 3. Open in Browser
Navigate to `http://localhost:8000`

## 🎮 Available Simulations

### 🗳️ Consensus Algorithms
- **Raft Consensus**: Leader election and log replication with step-by-step visualization

### 🔒 Atomic Commit Protocols
- **Two-Phase Commit (2PC)**: Classic atomic commit with prepare and commit phases
- **Three-Phase Commit (3PC)**: Non-blocking atomic commit with pre-commit phase

### 📊 Data Structures
- **Bloom Filter**: Probabilistic set membership testing with hash function visualization

### 💾 Caching & Eviction
- **Cache Eviction Policies**: Compare LRU, LFU, and FIFO strategies side-by-side

### 🚦 Rate Limiting
- **5 Rate Limiting Algorithms**:
  - Fixed Window Counter
  - Sliding Log
  - Sliding Window Counter
  - Token Bucket
  - Leaky Bucket

### ⚙️ Distributed Processing
- **MapReduce**: Map, shuffle, and reduce phases with parallel execution flow

### 🔄 Data Replication
- **Change Data Capture (CDC)**: Real-time data replication from database through Kafka to derived systems

### 🌐 Network Protocols
- **TCP vs UDP**: Side-by-side comparison of reliable TCP vs fast UDP protocol behavior

### ⚡ Frontend Performance
- **Pagination vs Virtualization**: Compare traditional pagination with efficient list virtualization

## 🎨 Features

- **Dark/Light Theme**: Toggle between themes with persistent preference
- **Multi-User Support**: Each user gets isolated simulation state via session IDs
- **Step-by-Step Execution**: Watch algorithms execute one step at a time
- **Reset Functionality**: Reset any simulation to initial state
- **Real-time Visualization**: Live updates using D3.js
- **Responsive Design**: Works on desktop and tablet devices

## 🛠️ Development

### Adding a New Simulation

1. **Backend**: Create simulation logic in `internal/simulation/<feature>/`
2. **API Layer**: Add routes in `internal/api/<feature>/routes.go`
3. **Session State**: Update `internal/session/manager.go` to include new feature
4. **Frontend**: Create folder `frontend/visualizations/<feature>/`
5. **Link**: Add simulation card to `frontend/index.html`

See `.codebase-structure.json` for detailed patterns and conventions.

### Code Patterns

- **Backend**: Thread-safe simulations with `sync.RWMutex`, JSON-serializable state
- **Frontend**: Session-aware API calls, D3.js visualizations, theme support
- **API**: RESTful endpoints with CORS enabled for local development

## 📚 Learning Concepts

Each simulation is designed to help you understand:
- How distributed systems handle coordination and failures
- Trade-offs between different algorithmic approaches
- Performance characteristics of various strategies
- Real-world implications of design decisions

## 🤝 Contributing

When adding new features:
1. Ensure thread-safety in backend simulations
2. Add session support for multi-user isolation
3. Include reset functionality
4. Support both dark and light themes
5. Update README files with new endpoints and features

## 📝 License

This is an educational project for learning distributed systems concepts.
