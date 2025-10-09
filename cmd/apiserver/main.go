package apiserver
// Package main implements the ScuffedK8 API Server
// This is like kube-apiserver - the central management point for the cluster
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"scuffedk8/pkg/api"
	"scuffedk8/pkg/store"
)

const (
	DefaultPort  = ":8080"
	ReadTimeout  = 15 * time.Second
	WriteTimeout = 15 * time.Second
)

func main() {
	log.Printf("Starting ScuffedK8 API Server...")

	// Initialize store (holds all cluster state)
	clusterStore := store.New()

	// Initialize API server
	apiServer := api.NewServer(clusterStore)

	// Setup HTTP server
	server := &http.Server{
		Addr:         DefaultPort,
		Handler:      apiServer.SetupRoutes(),
		ReadTimeout:  ReadTimeout,
		WriteTimeout: WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("API Server listening on http://localhost%s", DefaultPort)
		log.Printf("Available endpoints:")
		log.Printf("  GET  /              - Server status")
		log.Printf("  GET  /health        - Health check")
		log.Printf("  GET  /nodes         - Node status")
		log.Printf("  GET  /workload      - CPU workload test")
		log.Printf("  WS   /ws/node       - Node WebSocket connection")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down API server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("API server stopped")
}