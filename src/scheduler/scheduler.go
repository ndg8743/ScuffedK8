package scheduler

import (
	"errors"
	"sort"
)

// ---- Core Types ----

// Resources represents a collection of named resources with integer quantities.
// Basically just a map of resource names to amounts - like "cpu": 1000, "memory": 4096
type Resources map[string]int64

// Clone creates a deep copy of the Resources map to avoid mutation of the original.
// Need this because Go maps are reference types and we don't want to mess up the original
func (r Resources) Clone() Resources {
	out := make(Resources, len(r))
	for k, v := range r {
		out[k] = v
	}
	return out
}

// Add increases the resource quantities by the amounts in the provided Resources.
// Used when we allocate resources to a node
func (r Resources) Add(x Resources) {
	for k, v := range x {
		r[k] += v
	}
}

// Sub decreases the resource quantities by the amounts in the provided Resources.
// Used when we deallocate resources from a node
func (r Resources) Sub(x Resources) {
	for k, v := range x {
		r[k] -= v
	}
}

// Fits checks if the current Resources can satisfy the requested Resources.
// Returns true if we have enough of everything the pod needs
func (r Resources) Fits(need Resources) bool {
	for k, v := range need {
		if r[k] < v {
			return false
		}
	}
	return true
}

// Node represents a compute node in the cluster with its capacity, current allocation,
// and metadata for scheduling decisions.
type Node struct {
	ID        string    // Unique identifier for the node
	Capacity  Resources // Total available resources on the node
	Allocated Resources // Currently allocated resources
	Up        bool      // Whether the node is available for scheduling
}

// Free calculates the remaining available resources on the node by subtracting
// allocated resources from total capacity.
func (n Node) Free() Resources {
	f := n.Capacity.Clone()
	f.Sub(n.Allocated)
	return f
}

// Pod represents a workload unit that needs to be scheduled onto a node.
type Pod struct {
	ID       string    // Unique identifier for the pod
	Request  Resources // Resource requirements for the pod
	Priority int64     // Scheduling priority (higher values = higher priority)
}

// Result contains the outcome of a scheduling operation, including successful
// assignments, pods that couldn't be scheduled, and the updated node states.
type Result struct {
	Assignments map[string]string // Mapping of podID -> nodeID for successful assignments
	Pending     []string          // List of pod IDs that couldn't be scheduled
	Nodes       []Node            // Updated node states with new allocations
}

// ---- Main Scheduler Function ----

// Assign schedules a list of pods onto available nodes using a simple greedy algorithm.
// Returns a Result with successful assignments, pending pods, and updated node states.
//
// The algorithm is pretty simple:
//  1. Check inputs and make copies so we don't mess up the originals
//  2. Sort pods by resource requirements (biggest first)
//  3. For each pod, find the first node that can fit it
//  4. Update node allocations and track assignments
//
// Parameters:
//   - pods: List of pods to be scheduled
//   - nodes: List of available nodes in the cluster
//
// Returns:
//   - Result: Contains assignments, pending pods, and updated node states

func Assign(pods []Pod, nodes []Node) (Result, error) {
	// Check if we have any nodes to work with
	if len(nodes) == 0 {
		return Result{}, errors.New("no nodes available for scheduling")
	}

	// Make copies of nodes so we don't mess up the originals
	cnodes := make([]Node, len(nodes))
	for i := range nodes {
		cnodes[i] = nodes[i]
		if cnodes[i].Allocated == nil {
			cnodes[i].Allocated = make(Resources)
		}
		if cnodes[i].Capacity == nil {
			cnodes[i].Capacity = make(Resources)
		}
	}

	// Make a copy of pods and sort them by resource requirements
	cpods := make([]Pod, len(pods))
	copy(cpods, pods)

	// Sort pods by dominant resource (biggest first), then by priority (highest first)
	sort.Slice(cpods, func(i, j int) bool {
		// Calculate dominant resource for each pod
		domI := dominantResource(cpods[i].Request)
		domJ := dominantResource(cpods[j].Request)

		// Sort by dominant resource first (biggest first)
		if domI != domJ {
			return domI > domJ
		}

		// Then by priority (highest first)
		if cpods[i].Priority != cpods[j].Priority {
			return cpods[i].Priority > cpods[j].Priority
		}

		// Finally by ID for stability
		return cpods[i].ID < cpods[j].ID
	})

	// Set up tracking for results
	assignments := make(map[string]string, len(cpods))
	var pending []string

	// Main scheduling loop - try to fit each pod somewhere
	for _, pod := range cpods {
		assigned := false

		// Find the first node that can fit this pod
		for i := range cnodes {
			// Skip nodes that are down
			if !cnodes[i].Up {
				continue
			}

			// Check if node has enough resources
			if cnodes[i].Free().Fits(pod.Request) {
				// Assign pod to this node
				cnodes[i].Allocated.Add(pod.Request)
				assignments[pod.ID] = cnodes[i].ID
				assigned = true
				break
			}
		}

		// If no node could fit this pod, mark it as pending
		if !assigned {
			pending = append(pending, pod.ID)
		}
	}

	return Result{
		Assignments: assignments,
		Pending:     pending,
		Nodes:       cnodes,
	}, nil
}

// dominantResource calculates the largest resource requirement for a pod.
// Used for sorting pods by their resource needs - basically finds the biggest resource request
func dominantResource(r Resources) int64 {
	var max int64
	for _, v := range r {
		if v > max {
			max = v
		}
	}
	return max
}
