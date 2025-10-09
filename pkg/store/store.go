package store
// Package store provides in-memory storage for cluster state
package store

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	
	"scuffedk8/pkg/scheduler"
	"scuffedk8/pkg/types"
)

// Store manages all cluster state
type Store struct {
	nodes       map[string]*types.Node
	connections map[string]*websocket.Conn
	scheduler   *scheduler.Scheduler
	mu          sync.RWMutex
}

// New creates a new Store instance
func New() *Store {
	return &Store{
		nodes:       make(map[string]*types.Node),
		connections: make(map[string]*websocket.Conn),
		scheduler:   scheduler.New(),
	}
}

// RegisterNode adds a node to the store and scheduler
func (s *Store) RegisterNode(node *types.Node) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nodes[node.ID] = node
	s.scheduler.RegisterNode(node)
}

// UnregisterNode removes a node from the store
func (s *Store) UnregisterNode(nodeID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.nodes, nodeID)
	s.scheduler.UnregisterNode(nodeID)
}

// AddNodeConnection stores a WebSocket connection for a node
func (s *Store) AddNodeConnection(nodeID string, conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.connections[nodeID] = conn
}

// RemoveNodeConnection removes a node's WebSocket connection
func (s *Store) RemoveNodeConnection(nodeID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.connections, nodeID)
}

// GetNodeConnection retrieves a node's WebSocket connection
func (s *Store) GetNodeConnection(nodeID string) (*websocket.Conn, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conn, exists := s.connections[nodeID]
	return conn, exists
}

// UpdateNodeLastSeen updates the last seen timestamp for a node
func (s *Store) UpdateNodeLastSeen(nodeID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if node, exists := s.nodes[nodeID]; exists {
		node.LastSeen = time.Now()
	}
}

// GetNodeStatuses returns status information for all nodes
func (s *Store) GetNodeStatuses() []types.NodeStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	statuses := make([]types.NodeStatus, 0, len(s.nodes))
	for _, node := range s.nodes {
		status := types.NodeStatus{
			NodeID:    node.ID,
			Connected: s.connections[node.ID] != nil,
			Status:    getNodeStatus(node),
			LastSeen:  node.LastSeen,
			Capacity:  node.Capacity,
			Allocated: node.Allocated,
		}
		statuses = append(statuses, status)
	}

	return statuses
}

// GetNodeCount returns the number of registered nodes
func (s *Store) GetNodeCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.nodes)
}

// SchedulePods schedules pods using the scheduler
func (s *Store) SchedulePods(pods []types.Pod) (types.ScheduleResult, error) {
	return s.scheduler.Assign(pods)
}

// GetScheduler returns the scheduler instance
func (s *Store) GetScheduler() *scheduler.Scheduler {
	return s.scheduler
}

// SendToNode sends a message to a specific node via WebSocket
func (s *Store) SendToNode(nodeID string, message interface{}) error {
	s.mu.RLock()
	conn, exists := s.connections[nodeID]
	s.mu.RUnlock()

	if !exists {
		return nil
	}

	return conn.WriteJSON(message)
}

// getNodeStatus determines the status string for a node
func getNodeStatus(node *types.Node) string {
	if !node.Up {
		return "down"
	}

	if time.Since(node.LastSeen) > 30*time.Second {
		return "unreachable"
	}

	return "ready"
}