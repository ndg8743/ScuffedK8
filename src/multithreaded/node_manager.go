package multithreaded

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// NodeManager handles all node connections and communication
type NodeManager struct {
	nodes       map[string]*ManagedNode
	mu          sync.RWMutex
	connections map[string]*websocket.Conn
	heartbeatInterval time.Duration
	nodeTimeout       time.Duration
}

// ManagedNode represents a node with its state and capabilities
type ManagedNode struct {
	ID         string            `json:"id"`
	Status     string            `json:"status"`
	LastSeen   time.Time         `json:"last_seen"`
	Capacity   map[string]int64  `json:"capacity"`
	Allocated  map[string]int64  `json:"allocated"`
	Connection *websocket.Conn   `json:"-"`
	TaskQueue  chan SubTask      `json:"-"`
	mu         sync.RWMutex
}

// NodeStatus represents the status information for a node
type NodeStatus struct {
	NodeID    string            `json:"node_id"`
	Connected bool              `json:"connected"`
	Status    string            `json:"status"`
	LastSeen  time.Time         `json:"last_seen"`
	Capacity  map[string]int64  `json:"capacity"`
	Allocated map[string]int64  `json:"allocated"`
	QueueSize int               `json:"queue_size"`
}

// NodeMessage represents communication between scheduler and nodes
type NodeMessage struct {
	Type      string      `json:"type"`
	NodeID    string      `json:"node_id"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// HeartbeatMessage represents a node heartbeat
type HeartbeatMessage struct {
	NodeID    string            `json:"node_id"`
	Status    string            `json:"status"`
	Capacity  map[string]int64  `json:"capacity"`
	Allocated map[string]int64  `json:"allocated"`
	Timestamp time.Time         `json:"timestamp"`
}

// TaskAssignmentMessage represents a task assignment to a node
type TaskAssignmentMessage struct {
	TaskID   string      `json:"task_id"`
	TaskType string      `json:"task_type"`
	Data     interface{} `json:"data"`
	Priority int         `json:"priority"`
}

// TaskResultMessage represents a task result from a node
type TaskResultMessage struct {
	TaskID    string        `json:"task_id"`
	Success   bool          `json:"success"`
	Result    interface{}   `json:"result"`
	Error     string        `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	NodeID    string        `json:"node_id"`
}

// NewNodeManager creates a new node manager
func NewNodeManager() *NodeManager {
	return &NodeManager{
		nodes:             make(map[string]*ManagedNode),
		connections:       make(map[string]*websocket.Conn),
		heartbeatInterval: 5 * time.Second,
		nodeTimeout:       30 * time.Second,
	}
}

// HandleNodeConnection manages a WebSocket connection from a node
func (nm *NodeManager) HandleNodeConnection(node *Node) {
	nodeID := node.ID

	// Create managed node
	managedNode := &ManagedNode{
		ID:         nodeID,
		Status:     "connected",
		LastSeen:   time.Now(),
		Capacity:   node.Capacity,
		Allocated:  make(map[string]int64),
		Connection: node.Connection,
		TaskQueue:  make(chan SubTask, 100),
	}

	// Register the node
	nm.mu.Lock()
	nm.nodes[nodeID] = managedNode
	nm.connections[nodeID] = node.Connection
	nm.mu.Unlock()

	log.Printf("Node %s registered successfully", nodeID)

	// Start node communication handlers
	go nm.handleNodeMessages(managedNode)
	go nm.handleNodeTasks(managedNode)

	// Send welcome message
	welcomeMsg := NodeMessage{
		Type:      "welcome",
		NodeID:    nodeID,
		Data:      map[string]string{"status": "connected"},
		Timestamp: time.Now(),
	}
	nm.sendMessageToNode(nodeID, welcomeMsg)

	// Start heartbeat monitoring
	go nm.monitorNodeHeartbeat(nodeID)
}

// handleNodeMessages processes incoming messages from a node
func (nm *NodeManager) handleNodeMessages(node *ManagedNode) {
	defer func() {
		nm.removeNode(node.ID)
		node.Connection.Close()
	}()

	for {
		var message NodeMessage
		err := node.Connection.ReadJSON(&message)
		if err != nil {
			log.Printf("Error reading message from node %s: %v", node.ID, err)
			break
		}

		// Update last seen
		node.mu.Lock()
		node.LastSeen = time.Now()
		node.mu.Unlock()

		// Process different message types
		switch message.Type {
		case "heartbeat":
			nm.handleHeartbeat(node, message)
		case "task_result":
			nm.handleTaskResult(node, message)
		case "status_update":
			nm.handleStatusUpdate(node, message)
		default:
			log.Printf("Unknown message type from node %s: %s", node.ID, message.Type)
		}
	}
}

// handleNodeTasks sends tasks from the queue to the node
func (nm *NodeManager) handleNodeTasks(node *ManagedNode) {
	for {
		select {
		case task := <-node.TaskQueue:
			taskMsg := TaskAssignmentMessage{
				TaskID:   task.ID,
				TaskType: "subtask",
				Data:     task.Data,
				Priority: 1,
			}

			message := NodeMessage{
				Type:      "task_assignment",
				NodeID:    node.ID,
				Data:      taskMsg,
				Timestamp: time.Now(),
			}

			if err := nm.sendMessageToNode(node.ID, message); err != nil {
				log.Printf("Failed to send task to node %s: %v", node.ID, err)
				// Requeue the task or handle failure
				continue
			}

			log.Printf("Sent task %s to node %s", task.ID, node.ID)
		}
	}
}

// handleHeartbeat processes heartbeat messages from nodes
func (nm *NodeManager) handleHeartbeat(node *ManagedNode, message NodeMessage) {
	var heartbeat HeartbeatMessage
	data, _ := json.Marshal(message.Data)
	json.Unmarshal(data, &heartbeat)

	node.mu.Lock()
	node.Status = heartbeat.Status
	if heartbeat.Capacity != nil {
		node.Capacity = heartbeat.Capacity
	}
	if heartbeat.Allocated != nil {
		node.Allocated = heartbeat.Allocated
	}
	node.mu.Unlock()

	log.Printf("Received heartbeat from node %s, status: %s", node.ID, heartbeat.Status)
}

// handleTaskResult processes task completion results from nodes
func (nm *NodeManager) handleTaskResult(node *ManagedNode, message NodeMessage) {
	var result TaskResultMessage
	data, _ := json.Marshal(message.Data)
	json.Unmarshal(data, &result)

	log.Printf("Received task result from node %s: task=%s, success=%t", 
		node.ID, result.TaskID, result.Success)

	// Here you would typically forward the result to the job scheduler
	// or update some central state about task completion
}

// handleStatusUpdate processes status updates from nodes
func (nm *NodeManager) handleStatusUpdate(node *ManagedNode, message NodeMessage) {
	log.Printf("Status update from node %s: %+v", node.ID, message.Data)
	
	node.mu.Lock()
	// Update node status based on the message data
	node.mu.Unlock()
}

// monitorNodeHeartbeat monitors if a node is sending regular heartbeats
func (nm *NodeManager) monitorNodeHeartbeat(nodeID string) {
	ticker := time.NewTicker(nm.nodeTimeout)
	defer ticker.Stop()

	for {
		<-ticker.C

		nm.mu.RLock()
		node, exists := nm.nodes[nodeID]
		nm.mu.RUnlock()

		if !exists {
			return // Node was removed
		}

		node.mu.RLock()
		lastSeen := node.LastSeen
		node.mu.RUnlock()

		if time.Since(lastSeen) > nm.nodeTimeout {
			log.Printf("Node %s timed out, removing", nodeID)
			nm.removeNode(nodeID)
			return
		}
	}
}

// sendMessageToNode sends a message to a specific node
func (nm *NodeManager) sendMessageToNode(nodeID string, message NodeMessage) error {
	nm.mu.RLock()
	conn, exists := nm.connections[nodeID]
	nm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("node %s not connected", nodeID)
	}

	return conn.WriteJSON(message)
}

// AssignTaskToNode assigns a subtask to a specific node
func (nm *NodeManager) AssignTaskToNode(nodeID string, task SubTask) error {
	nm.mu.RLock()
	node, exists := nm.nodes[nodeID]
	nm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	select {
	case node.TaskQueue <- task:
		return nil
	default:
		return fmt.Errorf("node %s task queue is full", nodeID)
	}
}

// GetAllNodeStatuses returns status information for all nodes
func (nm *NodeManager) GetAllNodeStatuses() []NodeStatus {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	statuses := make([]NodeStatus, 0, len(nm.nodes))
	for _, node := range nm.nodes {
		node.mu.RLock()
		status := NodeStatus{
			NodeID:    node.ID,
			Connected: true,
			Status:    node.Status,
			LastSeen:  node.LastSeen,
			Capacity:  node.Capacity,
			Allocated: node.Allocated,
			QueueSize: len(node.TaskQueue),
		}
		node.mu.RUnlock()
		statuses = append(statuses, status)
	}

	return statuses
}

// GetConnectedNodeCount returns the number of connected nodes
func (nm *NodeManager) GetConnectedNodeCount() int {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	return len(nm.nodes)
}

// GetAvailableNodes returns nodes that are available for task assignment
func (nm *NodeManager) GetAvailableNodes() []*ManagedNode {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	var available []*ManagedNode
	for _, node := range nm.nodes {
		node.mu.RLock()
		if node.Status == "connected" && len(node.TaskQueue) < cap(node.TaskQueue) {
			available = append(available, node)
		}
		node.mu.RUnlock()
	}

	return available
}

// removeNode removes a node from management
func (nm *NodeManager) removeNode(nodeID string) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	delete(nm.nodes, nodeID)
	delete(nm.connections, nodeID)
	log.Printf("Node %s removed from management", nodeID)
}

// BroadcastMessage sends a message to all connected nodes
func (nm *NodeManager) BroadcastMessage(message NodeMessage) {
	nm.mu.RLock()
	nodeIDs := make([]string, 0, len(nm.nodes))
	for nodeID := range nm.nodes {
		nodeIDs = append(nodeIDs, nodeID)
	}
	nm.mu.RUnlock()

	for _, nodeID := range nodeIDs {
		if err := nm.sendMessageToNode(nodeID, message); err != nil {
			log.Printf("Failed to send broadcast message to node %s: %v", nodeID, err)
		}
	}
}