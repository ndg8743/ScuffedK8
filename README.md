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

main.go - API Server (Running on :8080)

HTTP endpoints: /, /health, /nodes, /workload
    WebSocket endpoint: /ws/node for node connections
    CPU workload simulation (Pi calculation)
    Connection tracking for nodes
scheduler.go - Greedy Scheduler Algorithm
    Pod scheduling logic (greedy algorithm)
    Resource-based pod placement
    Priority-based scheduling
basic_usage.go - Scheduler Test
    Demonstrates scheduling 3 pods to 2 nodes
    Shows resource allocation and assignments
ws_demo.go - Node Simulation
    Simulates 2 worker nodes
    WebSocket connections to API server
    Sends heartbeat messages every 2 seconds
