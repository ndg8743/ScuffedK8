package scheduler
// Package scheduler implements the pod scheduling algorithm
package scheduler

import (
	"errors"
	"log"
	"sort"

	"scuffedk8/pkg/types"
)

// Scheduler handles pod scheduling decisions
type Scheduler struct {
	nodes map[string]*types.Node
}

// New creates a new Scheduler instance
func New() *Scheduler {
	return &Scheduler{
		nodes: make(map[string]*types.Node),
	}
}

// RegisterNode adds a node to the scheduler
func (s *Scheduler) RegisterNode(node *types.Node) {
	s.nodes[node.ID] = node
	log.Printf("[Scheduler] Registered node %s with capacity: %v", node.ID, node.Capacity)
}

// UnregisterNode removes a node from the scheduler
func (s *Scheduler) Unregister(nodeID string) {
	delete(s.nodes, nodeID)
	log.Printf("[Scheduler] Unregistered node %s", nodeID)
}

// Assign schedules pods onto available nodes using a greedy algorithm
// Algorithm:
//  1. Sort pods by dominant resource (largest first), then by priority
//  2. For each pod, find first node with sufficient resources
//  3. Allocate pod to node and update node state
func (s *Scheduler) Assign(pods []types.Pod) (types.ScheduleResult, error) {
	if len(s.nodes) == 0 {
		return types.ScheduleResult{}, errors.New("no nodes available for scheduling")
	}

	// Convert map to slice and make copies
	nodesList := make([]types.Node, 0, len(s.nodes))
	for _, node := range s.nodes {
		nodeCopy := *node
		if nodeCopy.Allocated == nil {
			nodeCopy.Allocated = make(types.Resources)
		}
		nodesList = append(nodesList, nodeCopy)
	}

	// Sort pods by resource requirements
	cpods := make([]types.Pod, len(pods))
	copy(cpods, pods)
	
	sort.Slice(cpods, func(i, j int) bool {
		domI := dominantResource(cpods[i].Request)
		domJ := dominantResource(cpods[j].Request)

		if domI != domJ {
			return domI > domJ
		}

		if cpods[i].Priority != cpods[j].Priority {
			return cpods[i].Priority > cpods[j].Priority
		}

		return cpods[i].ID < cpods[j].ID
	})

	// Schedule pods
	assignments := make(map[string]string)
	var pending []string

	for _, pod := range cpods {
		assigned := false

		for i := range nodesList {
			if !nodesList[i].Up {
				continue
			}

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

	// Update actual node states
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

// GetNodes returns all registered nodes
func (s *Scheduler) GetNodes() []types.Node {
	nodes := make([]types.Node, 0, len(s.nodes))
	for _, node := range s.nodes {
		nodes = append(nodes, *node)
	}
	return nodes
}

// dominantResource finds the largest resource requirement
func dominantResource(r types.Resources) int64 {
	var max int64
	for _, v := range r {
		if v > max {
			max = v
		}
	}
	return max
}