# Backend Server

Go-based simulation engine with multi-user session management for the System Design Simulator.

## 🚀 Running the Server

```bash
go run cmd/server/main.go
```

Server will start on **port 8080**.

**Note:** Manual restart required after code changes (Ctrl+C to stop, then run again).

## 🏗️ Architecture

### Core Components

- **`cmd/server/main.go`**: Entry point, initializes session manager and routes
- **`internal/api/`**: HTTP handlers with CORS and session extraction
- **`internal/simulation/`**: Core simulation algorithms and business logic
- **`internal/session/`**: Multi-user session management with isolated state

### Session Management

- Each user gets isolated simulation state via `X-Session-ID` header
- Sessions stored in-memory (can swap to Redis)
- Auto-cleanup of inactive sessions (1 hour timeout)
- Fallback to "default" session for backward compatibility

## 📡 API Endpoints

### Health Check
- `GET /api/health` - Server health check

### Consensus Algorithms

#### Raft
- `GET /api/consensus/raft/state` - Get current cluster state
- `POST /api/consensus/raft/election?nodeId=<id>` - Start election from specific node
- `POST /api/consensus/raft/set-leader?nodeId=<id>` - Directly set a node as leader
- `POST /api/consensus/raft/reset` - Reset cluster to initial state

### Atomic Commit Protocols

#### Two-Phase Commit (2PC)
- `GET /api/atomic-commit/2pc/state` - Get coordinator and participant states
- `POST /api/atomic-commit/2pc/start` - Start new 2PC transaction
  - Body: `{"transactionId": "tx-1", "data": "update account balance"}`
- `POST /api/atomic-commit/2pc/participant/fail?participantId=<id>` - Simulate participant failure
- `POST /api/atomic-commit/2pc/participant/recover?participantId=<id>` - Recover failed participant
- `POST /api/atomic-commit/2pc/coordinator/fail` - Simulate coordinator failure
- `POST /api/atomic-commit/2pc/coordinator/recover` - Recover failed coordinator
- `POST /api/atomic-commit/2pc/reset` - Reset to initial state

#### Three-Phase Commit (3PC)
- `GET /api/atomic-commit/3pc/state` - Get coordinator and participant states
- `POST /api/atomic-commit/3pc/start` - Start new 3PC transaction
  - Body: `{"transactionId": "tx-1", "data": "update account balance"}`
- `POST /api/atomic-commit/3pc/participant/fail?participantId=<id>` - Simulate participant failure
- `POST /api/atomic-commit/3pc/participant/recover?participantId=<id>` - Recover failed participant
- `POST /api/atomic-commit/3pc/coordinator/fail` - Simulate coordinator failure
- `POST /api/atomic-commit/3pc/coordinator/recover` - Recover failed coordinator
- `POST /api/atomic-commit/3pc/reset` - Reset to initial state

### Rate Limiting
- `GET /api/rate-limiting/state` - Get state of all 5 rate limiters
- `POST /api/rate-limiting/send-request` - Send single request to all limiters
- `POST /api/rate-limiting/send-burst?count=<n>` - Send burst of n requests
- `POST /api/rate-limiting/reset` - Reset all rate limiters

### Cache Eviction
- `GET /api/cache/state` - Get state of all 3 cache implementations
- `POST /api/cache/put` - Put item in all caches
  - Body: `{"key": "A", "value": "Data A"}`
- `POST /api/cache/get` - Get item from all caches
  - Body: `{"key": "A"}`
- `POST /api/cache/reset` - Reset all caches

### MapReduce
- `GET /api/mapreduce/state` - Get current job state
- `POST /api/mapreduce/start` - Start MapReduce job
  - Body: `{"jobId": "word-count-1", "input": ["line1", "line2"], "mappers": 2, "reducers": 2}`
- `POST /api/mapreduce/reset` - Reset job

### Change Data Capture (CDC)
- `GET /api/cdc/state` - Get CDC system state
- `POST /api/cdc/insert` - Insert row in database
  - Body: `{"table": "users", "data": {"id": 1, "name": "Alice"}}`
- `POST /api/cdc/update` - Update row in database
  - Body: `{"table": "users", "id": 1, "data": {"name": "Alice Updated"}}`
- `POST /api/cdc/delete` - Delete row from database
  - Body: `{"table": "users", "id": 1}`
- `POST /api/cdc/reset` - Reset CDC system

### Bloom Filter
- `GET /api/bloomfilter/state` - Get Bloom filter state
- `POST /api/bloomfilter/add` - Add element
  - Body: `{"element": "alice@example.com"}`
- `POST /api/bloomfilter/check` - Check if element exists
  - Body: `{"element": "alice@example.com"}`
- `POST /api/bloomfilter/reset` - Reset filter

### TCP/UDP Simulation
- `GET /api/tcpudp/state` - Get simulation state
- `POST /api/tcpudp/send-tcp` - Send TCP packet
  - Body: `{"message": "Hello", "packetLoss": 0.1}`
- `POST /api/tcpudp/send-udp` - Send UDP packet
  - Body: `{"message": "Hello", "packetLoss": 0.1}`
- `POST /api/tcpudp/reset` - Reset simulation

### Pagination
- `GET /api/pagination/state` - Get simulation state
- `POST /api/pagination/load-page` - Load page with pagination
  - Body: `{"pageNumber": 1, "pageSize": 20}`
- `POST /api/pagination/load-virtual` - Load items with virtualization
  - Body: `{"startIndex": 0, "endIndex": 50}`
- `POST /api/pagination/reset` - Reset simulation

## 🔧 Adding New Features

### 1. Create Simulation Logic
```
internal/simulation/<feature>/
├── types.go          # Data structures
└── <algorithm>.go    # Core implementation
```

**Pattern:**
- Thread-safe with `sync.RWMutex`
- JSON-serializable state
- Methods: `GetState()`, `Reset()`, `<Action>()`

### 2. Create API Layer
```
internal/api/<feature>/
└── routes.go         # HTTP handlers
```

**Pattern:**
```go
func SetupRoutes(sm *session.Manager) {
    sessionManager = sm
    http.HandleFunc("/api/<feature>/state", GetState)
    http.HandleFunc("/api/<feature>/action", DoAction)
    http.HandleFunc("/api/<feature>/reset", Reset)
}

func getSessionID(r *http.Request) string {
    // Try header → query param → "default"
}

func setCORSHeaders(w http.ResponseWriter) {
    // Allow localhost:8000
}
```

### 3. Update Session Manager
Add to `internal/session/manager.go`:
- Add field to `State` struct
- Initialize in `GetOrCreate()` method

### 4. Register Routes
Add to `internal/api/routes.go`:
```go
import "sds/internal/api/<feature>"

func SetupRoutes(sessionManager *session.Manager) {
    // ... existing routes ...
    <feature>.SetupRoutes(sessionManager)
}
```

## 🔐 CORS Configuration

CORS enabled for `http://localhost:8000` with:
- Methods: GET, POST, OPTIONS
- Headers: Content-Type, X-Session-ID

## 📦 Dependencies

- Standard library only (no external dependencies)
- Go module: `sds`

## 🧪 Testing Multi-User Sessions

Use different `X-Session-ID` header values to simulate multiple users:

```bash
# User 1
curl -H "X-Session-ID: user1" http://localhost:8080/api/consensus/raft/state

# User 2
curl -H "X-Session-ID: user2" http://localhost:8080/api/consensus/raft/state
```

Each user will have completely isolated simulation state.

## 🔍 Session Cleanup

Background task runs every 10 minutes to remove sessions inactive for more than 1 hour. Logs cleanup activity to console.
