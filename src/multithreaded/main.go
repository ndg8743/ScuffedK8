package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"scuffedk8/src/multithreaded"
)

const (
	DefaultPort = ":8082"
	WorkerCount = 4
	JobBufferSize = 1000
	ReadTimeout = 15 * time.Second
	WriteTimeout = 15 * time.Second
)

func main() {
	log.Printf("Starting ScuffedK8 Multi-Threaded Scheduler...")

	// Initialize components
	nodeManager := multithreaded.NewNodeManager()
	workerPool := multithreaded.NewWorkerPool(WorkerCount, JobBufferSize)
	jobScheduler := multithreaded.NewJobScheduler(nodeManager, workerPool)

	// Start worker pool
	workerPool.Start()

	// Create API server with all components
	apiServer := &APIServer{
		NodeManager:  nodeManager,
		WorkerPool:   workerPool, 
		JobScheduler: jobScheduler,
	}

	// Setup HTTP server
	server := &http.Server{
		Addr:         DefaultPort,
		Handler:      apiServer.SetupRoutes(),
		ReadTimeout:  ReadTimeout,
		WriteTimeout: WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on http://localhost%s", DefaultPort)
		log.Printf("Multi-threaded endpoints:")
		log.Printf("  GET  /                    - Server status") 
		log.Printf("  GET  /health              - Health check")
		log.Printf("  GET  /nodes               - Connected nodes")
		log.Printf("  GET  /workload?iterations=N&parallelism=P - Distributed workload")
		log.Printf("  GET  /jobs/{id}/status    - Job status")
		log.Printf("  GET  /stats/workers       - Worker pool stats")
		log.Printf("  WS   /ws/node?node_id=X   - Node WebSocket")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Shutdown worker pool
	if err := workerPool.Shutdown(10 * time.Second); err != nil {
		log.Printf("Worker pool forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped")
}

// Temporary APIServer struct - this should be moved to api package
type APIServer struct {
	NodeManager  interface{}
	WorkerPool   interface{}
	JobScheduler interface{}
}

func (api *APIServer) SetupRoutes() http.Handler {
	// This is a placeholder - the real implementation would be in the api package
	mux := http.NewServeMux()
	
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ScuffedK8 Multi-Threaded Scheduler Running"))
	})
	
	return mux
}