package main

import (
	"fmt"
	"math/rand"
	"time"
)

// Pod represents an application/task that needs to run
type Pod struct {
	ID          string
	AppName     string
	CPURequired int
	MemRequired int
	WorkTime    time.Duration // How long the app takes to run
}

// Node represents a worker machine/computer
type Node struct {
	ID          int
	Name        string
	CPUCapacity int
	MemCapacity int
	tasks       chan Pod
}

// RunNode simulates a worker computer executing pods (applications)
func (n *Node) RunNode() {
	fmt.Printf("[Node %s] Started - Capacity: %d CPU, %d MB Memory\n",
		n.Name, n.CPUCapacity, n.MemCapacity)

	for pod := range n.tasks {
		// Check if node has enough resources
		if pod.CPURequired > n.CPUCapacity || pod.MemRequired > n.MemCapacity {
			fmt.Printf("[Node %s] ❌ Cannot run %s - insufficient resources\n",
				n.Name, pod.AppName)
			continue
		}

		// Execute the pod (application)
		fmt.Printf("[Node %s] ▶️  Running pod '%s' (%s) - Using %d CPU, %d MB\n",
			n.Name, pod.ID, pod.AppName, pod.CPURequired, pod.MemRequired)

		// Simulate application running
		time.Sleep(pod.WorkTime)

		fmt.Printf("[Node %s] ✅ Completed pod '%s' (%s)\n",
			n.Name, pod.ID, pod.AppName)
	}
}

func main() {
	fmt.Println("=== ScuffedK8 Node-Pod Demonstration ===\n")

	// Create a channel to distribute pods to nodes
	podQueue := make(chan Pod, 100)

	// Create 5 worker nodes (computers)
	numNodes := 5
	nodes := make([]*Node, numNodes)

	for i := 0; i < numNodes; i++ {
		nodes[i] = &Node{
			ID:          i,
			Name:        fmt.Sprintf("worker-%d", i+1),
			CPUCapacity: 500 + rand.Intn(500),   // Random capacity 500-1000
			MemCapacity: 2048 + rand.Intn(2048), // Random 2-4GB
			tasks:       podQueue,
		}

		// Start the node (worker computer)
		go nodes[i].RunNode()
	}

	fmt.Println()

	// Create various pods (applications to run)
	pods := []Pod{
		{ID: "pod-1", AppName: "test1", CPURequired: 100, MemRequired: 512, WorkTime: 2 * time.Second},
		{ID: "pod-2", AppName: "test2", CPURequired: 200, MemRequired: 1024, WorkTime: 3 * time.Second},
		{ID: "pod-3", AppName: "test2", CPURequired: 50, MemRequired: 256, WorkTime: 1 * time.Second},
		{ID: "pod-4", AppName: "test3", CPURequired: 150, MemRequired: 768, WorkTime: 2 * time.Second},
		{ID: "pod-5", AppName: "test4", CPURequired: 300, MemRequired: 1536, WorkTime: 4 * time.Second},
		{ID: "pod-6", AppName: "test5", CPURequired: 400, MemRequired: 2048, WorkTime: 3 * time.Second},
		{ID: "pod-7", AppName: "test6", CPURequired: 250, MemRequired: 1280, WorkTime: 2 * time.Second},
		{ID: "pod-8", AppName: "test7", CPURequired: 350, MemRequired: 2048, WorkTime: 5 * time.Second},
		{ID: "pod-9", AppName: "test8", CPURequired: 100, MemRequired: 512, WorkTime: 2 * time.Second},
		{ID: "pod-10", AppName: "test9", CPURequired: 80, MemRequired: 384, WorkTime: 1 * time.Second},
	}

	// Send all pods to the queue (scheduler distributes them to nodes)
	fmt.Println("📦 Scheduling pods to nodes...\n")
	go func() {
		for _, pod := range pods {
			podQueue <- pod
		}
		// Send a few more for demonstration
		for i := 11; i <= 20; i++ {
			podQueue <- Pod{
				ID:          fmt.Sprintf("pod-%d", i),
				AppName:     fmt.Sprintf("app-%d", i),
				CPURequired: 50 + rand.Intn(300),
				MemRequired: 256 + rand.Intn(1024),
				WorkTime:    time.Duration(1+rand.Intn(3)) * time.Second,
			}
		}
		close(podQueue)
	}()

	// Wait for all pods to complete
	time.Sleep(10 * time.Second)

	fmt.Println("\n=== All pods completed! ===")
}
