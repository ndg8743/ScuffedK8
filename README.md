# ScuffedK8

Simplified Kubernetes in Go for learning container orchestration.

## Quick Start

```bash
# Start scheduler
go run src/main.go

# Test scheduler algorithm
go run examples/basic_usage.go

# Test node communication
go run examples/ws_demo.go
```

## API Endpoints

| Endpoint | Description | Example |
|----------|-------------|---------|
| `GET /` | Server status | `curl localhost:8080/` |
| `GET /health` | Health check JSON | `curl localhost:8080/health` |
| `GET /nodes` | Connected nodes | `curl localhost:8080/nodes` |
| `GET /workload?iterations=N` | CPU-intensive pi calculation | `curl localhost:8080/workload?iterations=10000000` |
| `WS /ws/node?node_id=X` | Node WebSocket connection | See `ws_demo.go` |

## Testing

```bash
./test_endpoints.sh
```

## Architecture

- `src/main.go` - HTTP server + WebSocket handler (like kube-apiserver)
- `src/scheduler/scheduler.go` - Pod scheduling algorithm (like kube-scheduler)
- `examples/basic_usage.go` - Scheduler demo
- `examples/ws_demo.go` - Node communication demo (like kubelet)

## Features

- Greedy pod scheduler with resource constraints
- WebSocket node heartbeats
- Real CPU workload simulation
- In-memory state management
