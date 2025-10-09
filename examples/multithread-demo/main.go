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

func runMultithreadExample() {
   // Creating a channel
   podQueue := make(chan int)

   // Creating 10.000 workers to execute the task
   for i := 0; i < 10000; i++ {
	  go runNode(i, podQueue)
   }

   // In real K8s, these would be containers (nginx, postgres, etc.)
   numPods := 100000
   for i := 0; i < numPods; i++ {
	  podQueue <- i // Scheduler assigns pod to available node
   }

   // Keep main alive to see results
   time.Sleep(5 * time.Second)
}