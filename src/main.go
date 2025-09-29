// Package main implements a simple HTTP server for testing Kubernetes scheduler functionality.
// Basically just a basic server with some endpoints to check if everything is working
// and serves as a foundation for more complex scheduling operations.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const (
	// DefaultPort is the default port the server listens on
	DefaultPort = ":8080"
	// ReadTimeout is the maximum duration for reading the entire request
	ReadTimeout = 15 * time.Second
	// WriteTimeout is the maximum duration before timing out writes of the response
	WriteTimeout = 15 * time.Second
)

// WebSocket upgrader for node connections
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// Track active node connections
var nodeConnections = make(map[string]*websocket.Conn)

// Send message to a specific node
func sendToNode(nodeID string, message interface{}) error {
	conn, exists := nodeConnections[nodeID]
	if !exists {
		return fmt.Errorf("node %s not connected", nodeID)
	}
	return conn.WriteJSON(message)
}

// healthResponse represents the structure of a health check response
type healthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
}

// rootHandler handles requests to the root endpoint and provides basic service information
func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "ScuffedK8\n")
	fmt.Fprintf(w, "Status: Running\n")
}

// healthHandler handles health check requests and returns service status
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := healthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Service:   "scuffedk8-scheduler",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding health response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// WebSocket handler for node connections
func nodeWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Get node ID from query parameter
	nodeID := r.URL.Query().Get("node_id")
	if nodeID == "" {
		log.Printf("No node_id provided")
		return
	}

	// Store connection
	nodeConnections[nodeID] = conn
	log.Printf("Node %s connected", nodeID)

	// Keep connection alive and handle messages
	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Node %s disconnected: %v", nodeID, err)
			delete(nodeConnections, nodeID)
			break
		}
		log.Printf("[WS] %s -> server (type=%d): %s", nodeID, messageType, string(data))
	}
}

// nodesStatusHandler returns the current state of connected nodes
func nodesStatusHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[MUX] Route: %s %s", r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "application/json")

	type nodeStatus struct {
		NodeID    string `json:"node_id"`
		Connected bool   `json:"connected"`
	}

	nodes := []nodeStatus{}
	for id := range nodeConnections {
		nodes = append(nodes, nodeStatus{NodeID: id, Connected: true})
	}

	response := map[string]interface{}{
		"total_nodes": len(nodes),
		"nodes":       nodes,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// setupRoutes configures HTTP routes and their handlers
func setupRoutes() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", rootHandler)
	r.HandleFunc("/health", healthHandler)
	r.HandleFunc("/test", healthHandler)
	r.HandleFunc("/nodes", nodesStatusHandler)     // Node status endpoint
	r.HandleFunc("/ws/node", nodeWebSocketHandler) // WebSocket endpoint for nodes

	return r
}

// createServer creates and configures an HTTP server with proper timeouts
func createServer(port string) *http.Server {
	return &http.Server{
		Addr:         port,
		Handler:      setupRoutes(),
		ReadTimeout:  ReadTimeout,
		WriteTimeout: WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}
}

func main() {
	server := createServer(DefaultPort)

	log.Printf("Starting ScuffedK8 Scheduler Service on: http://localhost%s/", DefaultPort)
	log.Printf("Health check available at: http://localhost%s/health", DefaultPort)
	log.Printf("Legacy test endpoint available at: http://localhost%s/test", DefaultPort)

	// TODO: Consider using gorilla/mux for more advanced routing features
	// Example: r := mux.NewRouter(); r.HandleFunc("/api/v1/schedule", scheduleHandler).Methods("POST")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
