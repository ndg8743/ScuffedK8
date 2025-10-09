package main

import (
	"fmt"
	"time"
)

// Node represents a worker computer/machine
func runNode(nodeID int, podQueue chan int) {
	for podID := range podQueue {
		// Simulate node executing a pod (application)
		time.Sleep(2 * time.Second)
		fmt.Printf("Node %d executed Pod %d\n", nodeID, podID)
	}
}

func main() {
	fmt.Println("=== Multithread Demo: 10,000 Nodes executing 100,000 Pods ===\n")

	// Create a channel for distributing pods to nodes
	podQueue := make(chan int)

	// Create 10,000 NODES (worker computers)
	// In real K8s, this would be like 10,000 servers/VMs
	numNodes := 10000
	fmt.Printf("Starting %d worker nodes...\n", numNodes)
	for i := 0; i < numNodes; i++ {
		go runNode(i, podQueue)
	}

	// Create 100,000 PODS (applications/tasks to run)
	// In real K8s, these would be containers (nginx, postgres, etc.)
	numPods := 100000
	fmt.Printf("Distributing %d pods to nodes...\n\n", numPods)
	for i := 0; i < numPods; i++ {
		podQueue <- i // Scheduler assigns pod to available node
	}

	// Keep main alive to see results
	fmt.Println("\nWaiting for pods to complete...")
	time.Sleep(10 * time.Second)
	fmt.Println("Done!")
}
