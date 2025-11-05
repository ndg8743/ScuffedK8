// Package types defines the core data structures used throughout ScuffedK8
package types

import (
	"time"
)

// Resources represents a collection of named resources with integer quantities
// Example: {"cpu": 1000, "memory": 4096}
type Resources map[string]int64

// Clone creates a deep copy of the Resources map
func (r Resources) Clone() Resources {
	out := make(Resources, len(r))
	for k, v := range r {
		out[k] = v
	}
	return out
}

// Add increases resource quantities by the amounts in the provided Resources
func (r Resources) Add(x Resources) {
	for k, v := range x {
		r[k] += v
	}
}

// Sub decreases resource quantities by the amounts in the provided Resources
func (r Resources) Sub(x Resources) {
	for k, v := range x {
		r[k] -= v
	}
}

// Fits checks if current Resources can satisfy the requested Resources
func (r Resources) Fits(need Resources) bool {
	for k, v := range need {
		if r[k] < v {
			return false
		}
	}
	return true
}

// Node represents a compute node in the cluster
type Node struct {
	ID        string    `json:"id"`
	Capacity  Resources `json:"capacity"`
	Allocated Resources `json:"allocated"`
	Up        bool      `json:"up"`
	LastSeen  time.Time `json:"last_seen"`
}

// Free calculates remaining available resources on the node
func (n Node) Free() Resources {
	f := n.Capacity.Clone()
	f.Sub(n.Allocated)
	return f
}

// Pod represents a workload unit to be scheduled
type Pod struct {
	ID       string    `json:"id"`
	Request  Resources `json:"request"`
	Priority int64     `json:"priority"`
	NodeID   string    `json:"node_id,omitempty"`
	Status   string    `json:"status"`
}

// ScheduleResult contains the outcome of a scheduling operation
type ScheduleResult struct {
	Assignments map[string]string `json:"assignments"` // podID -> nodeID
	Pending     []string          `json:"pending"`     // pods that couldn't be scheduled
	Nodes       []Node            `json:"nodes"`       // updated node states
}

// NodeMessage represents communication between components
type NodeMessage struct {
	Type      string      `json:"type"`
	NodeID    string      `json:"node_id"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// HeartbeatData represents node heartbeat information
type HeartbeatData struct {
	NodeID    string    `json:"node_id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Capacity  Resources `json:"capacity,omitempty"`
	Allocated Resources `json:"allocated,omitempty"`
}

// HealthResponse represents API health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
}

// NodeStatus represents the status of a connected node
type NodeStatus struct {
	NodeID    string    `json:"node_id"`
	Connected bool      `json:"connected"`
	Status    string    `json:"status"`
	LastSeen  time.Time `json:"last_seen"`
	Capacity  Resources `json:"capacity"`
	Allocated Resources `json:"allocated"`
}

// WorkloadRequest represents a CPU workload request
type WorkloadRequest struct {
	Iterations  int `json:"iterations"`
	Parallelism int `json:"parallelism"`
}

// WorkloadResponse represents the result of a workload execution
type WorkloadResponse struct {
	Workload   string  `json:"workload"`
	Iterations int     `json:"iterations"`
	Result     float64 `json:"result"`
	DurationMS int64   `json:"duration_ms"`
	CPUUsed    bool    `json:"cpu_used"`
}