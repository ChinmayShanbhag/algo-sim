# Backend Server

## Running the Server

```bash
go run cmd/server/main.go
```

**Note:** You'll need to manually restart the server after code changes (Ctrl+C to stop, then run again).

## API Endpoints

- `GET /api/health` - Health check
- `GET /api/consensus/raft/state` - Get current Raft cluster state
- `POST /api/consensus/raft/election?nodeId=X` - Start election from node X
- `POST /api/consensus/raft/reset` - Reset cluster to initial state

