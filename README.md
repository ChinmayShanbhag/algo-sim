# System Design Simulator (SDS)

A learning-focused simulator for exploring distributed systems concepts, trade-offs, and behaviors.

## Architecture

- **Backend**: Go simulation engine
- **Frontend**: D3.js/JavaScript visualization

## Project Structure

```
.
├── backend/          # Go simulation engine
├── frontend/         # D3.js visualization client
└── README.md
```

## Getting Started

### Backend (Go)
```bash
cd backend
go run cmd/server/main.go
```

### Frontend
```bash
cd frontend
python -m http.server 8000
# Then open http://localhost:8000
```

## Concepts to Explore

- Consensus algorithms
- Replication strategies
- Load balancing
- Distributed caching
- Message queues
- CAP theorem trade-offs
- Network partitioning
- Consistency models

