// Package scheduler implements a greedy pod scheduling algorithm.
// Assigns pods to nodes based on resource availability using a first-fit approach.
package scheduler

import (
	"errors"
	"log"
	"sort"

	"scuffedk8/pkg/types"
)

// Scheduler manages pod scheduling decisions.
// Not thread-safe - access through Store for thread safety.
type Scheduler struct {
	nodes map[string]*types.Node // node ID -> node state
}

// New creates a new Scheduler instance.
func New() *Scheduler {
	return &Scheduler{
		nodes: make(map[string]*types.Node),
	}
}

// RegisterNode adds a node to the scheduler.
func (s *Scheduler) RegisterNode(node *types.Node) {
	s.nodes[node.ID] = node
	log.Printf("[Scheduler] Registered node %s with capacity: %v", node.ID, node.Capacity)
}

// Unregister removes a node from the scheduler.
// Note: Does not clean up existing allocations.
func (s *Scheduler) Unregister(nodeID string) {
	delete(s.nodes, nodeID)
	log.Printf("[Scheduler] Unregistered node %s", nodeID)
}

// Assign schedules pods to nodes using a greedy first-fit algorithm.
//
// Algorithm:
//  1. Sort pods by dominant resource (largest first), then priority
//  2. For each pod, find first node with sufficient resources
//  3. Allocate pod to node and update allocations
//
// Returns assignments (podID -> nodeID), pending pods, and updated nodes.
func (s *Scheduler) Assign(pods []types.Pod) (types.ScheduleResult, error) {
	if len(s.nodes) == 0 {
		return types.ScheduleResult{}, errors.New("no nodes available for scheduling")
	}

	// Copy nodes to avoid modifying originals during scheduling
	nodesList := make([]types.Node, 0, len(s.nodes))
	for _, node := range s.nodes {
		nodeCopy := *node
		if nodeCopy.Allocated == nil {
			nodeCopy.Allocated = make(types.Resources)
		}
		nodesList = append(nodesList, nodeCopy)
	}

	// Copy and sort pods: dominant resource DESC, priority DESC, ID ASC
	cpods := make([]types.Pod, len(pods))
	copy(cpods, pods)

	sort.Slice(cpods, func(i, j int) bool {
		domI := dominantResource(cpods[i].Request)
		domJ := dominantResource(cpods[j].Request)

		// Sort by dominant resource first (largest pods first reduces fragmentation)
		if domI != domJ {
			return domI > domJ
		}

		// Then by priority (higher priority first)
		if cpods[i].Priority != cpods[j].Priority {
			return cpods[i].Priority > cpods[j].Priority
		}

		// Finally by ID for determinism
		return cpods[i].ID < cpods[j].ID
	})

	// Assign pods to nodes (first-fit greedy algorithm)
	assignments := make(map[string]string)
	var pending []string

	for _, pod := range cpods {
		assigned := false

		// Try each node until we find one that fits
		for i := range nodesList {
			if !nodesList[i].Up {
				continue
			}

			// Check if node has enough free resources
			if nodesList[i].Free().Fits(pod.Request) {
				nodesList[i].Allocated.Add(pod.Request)
				assignments[pod.ID] = nodesList[i].ID
				assigned = true
				log.Printf("[Scheduler] Assigned pod %s to node %s", pod.ID, nodesList[i].ID)
				break
			}
		}

		if !assigned {
			pending = append(pending, pod.ID)
			log.Printf("[Scheduler] Pod %s pending - no suitable node found", pod.ID)
		}
	}

	// Commit allocations to actual node states
	for _, node := range nodesList {
		if s.nodes[node.ID] != nil {
			s.nodes[node.ID].Allocated = node.Allocated
		}
	}

	return types.ScheduleResult{
		Assignments: assignments,
		Pending:     pending,
		Nodes:       nodesList,
	}, nil
}

// GetNodes returns a copy of all registered nodes.
func (s *Scheduler) GetNodes() []types.Node {
	nodes := make([]types.Node, 0, len(s.nodes))
	for _, node := range s.nodes {
		nodes = append(nodes, *node)
	}
	return nodes
}

// dominantResource returns the largest resource value in the map.
// Used for sorting pods by their biggest resource requirement.
func dominantResource(r types.Resources) int64 {
	var max int64
	for _, v := range r {
		if v > max {
			max = v
		}
	}
	return max
}
