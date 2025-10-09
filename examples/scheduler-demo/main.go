package main

import (
	"fmt"
	"log"

	"scuffedk8/src/scheduler"
)

func main() {
	// Create some pods to schedule
	pods := []scheduler.Pod{
		{
			ID:       "web-pod-1",
			Request:  scheduler.Resources{"cpu": 100, "memory": 512},
			Priority: 10,
		},
		{
			ID:       "db-pod-1",
			Request:  scheduler.Resources{"cpu": 200, "memory": 1024},
			Priority: 5,
		},
		{
			ID:       "cache-pod-1",
			Request:  scheduler.Resources{"cpu": 50, "memory": 256},
			Priority: 1,
		},
	}

	// Create some nodes in the cluster
	nodes := []scheduler.Node{
		{
			ID:       "worker-1",
			Capacity: scheduler.Resources{"cpu": 1000, "memory": 4096},
			Up:       true,
		},
		{
			ID:       "worker-2",
			Capacity: scheduler.Resources{"cpu": 500, "memory": 2048},
			Up:       true,
		},
	}

	// Schedule pods using the simple greedy algorithm
	result, err := scheduler.Assign(pods, nodes)
	if err != nil {
		log.Fatalf("Scheduling failed: %v", err)
	}

	// Print results
	fmt.Printf("Scheduling Results:\n")
	fmt.Printf("==================\n")

	fmt.Printf("Successful Assignments:\n")
	for podID, nodeID := range result.Assignments {
		fmt.Printf("  %s -> %s\n", podID, nodeID)
	}

	if len(result.Pending) > 0 {
		fmt.Printf("\nPending Pods (couldn't be scheduled):\n")
		for _, podID := range result.Pending {
			fmt.Printf("  %s\n", podID)
		}
	}

	fmt.Printf("\nUpdated Node States:\n")
	for _, node := range result.Nodes {
		free := node.Free()
		fmt.Printf("  %s: %d CPU, %d Memory free\n",
			node.ID, free["cpu"], free["memory"])
	}
}
