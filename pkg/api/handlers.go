// Package api provides HTTP handlers for the API server
package api

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	
	"scuffedk8/pkg/store"
	"scuffedk8/pkg/types"
)

// Server represents the API server with its dependencies
type Server struct {
	store    *store.Store
	upgrader websocket.Upgrader
}

// NewServer creates a new API server
func NewServer(store *store.Store) *Server {
	return &Server{
		store: store,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
	}
}

// RootHandler handles requests to the root endpoint
func (s *Server) RootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "ScuffedK8 API Server\n")
	fmt.Fprintf(w, "Status: Running\n")
	fmt.Fprintf(w, "Connected Nodes: %d\n", s.store.GetNodeCount())
}

// HealthHandler returns health status
func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := types.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Service:   "scuffedk8-apiserver",
		Version:   "1.0.0",
	}

	json.NewEncoder(w).Encode(response)
}

// NodesHandler returns status of all connected nodes
func (s *Server) NodesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	statuses := s.store.GetNodeStatuses()

	response := map[string]interface{}{
		"total_nodes": len(statuses),
		"nodes":       statuses,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// WorkloadHandler simulates CPU-intensive workload
func (s *Server) WorkloadHandler(w http.ResponseWriter, r *http.Request) {
	iterStr := r.URL.Query().Get("iterations")
	iterations := 10000000
	if iterStr != "" {
		if i, err := strconv.Atoi(iterStr); err == nil && i > 0 {
			iterations = i
		}
	}

	start := time.Now()
	result := calculatePi(iterations)
	duration := time.Since(start)

	w.Header().Set("Content-Type", "application/json")
	response := types.WorkloadResponse{
		Workload:   "calculate_pi",
		Iterations: iterations,
		Result:     result,
		DurationMS: duration.Milliseconds(),
		CPUUsed:    true,
	}
	
	json.NewEncoder(w).Encode(response)
}

// NodeWebSocketHandler handles WebSocket connections from nodes (kubelet)
func (s *Server) NodeWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	nodeID := r.URL.Query().Get("node_id")
	if nodeID == "" {
		log.Printf("No node_id provided")
		return
	}

	// Register node
	node := &types.Node{
		ID: nodeID,
		Capacity: types.Resources{
			"cpu":    1000,
			"memory": 4096,
		},
		Allocated: make(types.Resources),
		Up:        true,
		LastSeen:  time.Now(),
	}

	s.store.AddNodeConnection(nodeID, conn)
	s.store.RegisterNode(node)
	log.Printf("[API] Node %s connected via WebSocket", nodeID)

	// Handle messages from node
	for {
		var msg types.NodeMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("[API] Node %s disconnected: %v", nodeID, err)
			s.store.RemoveNodeConnection(nodeID)
			s.store.UnregisterNode(nodeID)
			break
		}

		s.handleNodeMessage(nodeID, msg)
	}
}

// handleNodeMessage processes messages from nodes
func (s *Server) handleNodeMessage(nodeID string, msg types.NodeMessage) {
	switch msg.Type {
	case "heartbeat":
		s.store.UpdateNodeLastSeen(nodeID)
		log.Printf("[API] Heartbeat from node %s", nodeID)
	case "status":
		log.Printf("[API] Status update from node %s: %+v", nodeID, msg.Data)
	default:
		log.Printf("[API] Unknown message type from node %s: %s", nodeID, msg.Type)
	}
}

// SetupRoutes configures all HTTP routes
func (s *Server) SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", s.RootHandler).Methods("GET")
	r.HandleFunc("/health", s.HealthHandler).Methods("GET")
	r.HandleFunc("/nodes", s.NodesHandler).Methods("GET")
	r.HandleFunc("/workload", s.WorkloadHandler).Methods("GET")
	r.HandleFunc("/ws/node", s.NodeWebSocketHandler)

	return r
}

// calculatePi approximates the value of Pi using the Leibniz series:
//   π = 4 × (1 - 1/3 + 1/5 - 1/7 + 1/9 - ...)
func calculatePi(iterations int) float64 {
	pi := 0.0
	for i := 0; i < iterations; i++ {
		// Each term is (-1)^i divided by (2i + 1)
		// math.Pow is used here to generate the alternating sign
		pi += math.Pow(-1, float64(i)) / (2*float64(i) + 1)
	}
	return pi * 4
}