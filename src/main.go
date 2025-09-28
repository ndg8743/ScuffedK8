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
)

const (
	// DefaultPort is the default port the server listens on
	DefaultPort = ":3000"
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

// setupRoutes configures HTTP routes and their handlers
func setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/test", healthHandler)

	return mux
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
