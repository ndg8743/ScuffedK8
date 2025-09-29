package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

// connectNode establishes a WebSocket connection to the ScuffedK8 scheduler
func connectNode(nodeID string) (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws/node", RawQuery: "node_id=" + nodeID}
	log.Printf("connecting node %s to %s", nodeID, u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// runNode simulates a worker node that connects to the scheduler and sends periodic heartbeat messages. 
// In real Kubernetes, nodes send status updates about their health, available resources, and running pods.
func runNode(nodeID string, stop <-chan struct{}) {
	c, err := connectNode(nodeID)
	if err != nil {
		log.Printf("%s connect error: %v", nodeID, err)
		return
	}
	defer c.Close()

	// Start a goroutine to read messages from the scheduler
	// This would handle scheduling requests, pod assignments, etc.
	go func() {
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Printf("%s read error: %v", nodeID, err)
				return
			}
			log.Printf("scheduler -> %s: %s", nodeID, string(msg))
		}
	}()

	// Send heartbeat messages every 2 seconds to show the node is alive
	// Real kubelets send node status updates with resource availability
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-stop:
			// Clean shutdown when receiving stop signal
			_ = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"))
			return
		case t := <-ticker.C:
			// Send heartbeat with current timestamp
			payload := fmt.Sprintf("{\"type\":\"heartbeat\",\"node_id\":\"%s\",\"ts\":\"%s\"}", nodeID, t.Format(time.RFC3339))
			if err := c.WriteMessage(websocket.TextMessage, []byte(payload)); err != nil {
				log.Printf("%s write error: %v", nodeID, err)
				return
			}
		}
	}
}

func main() {
	//shutdown handling
	stop := make(chan struct{})
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Printf("shutting down nodes...")
		close(stop)
	}()

	// Start two simulated worker nodes
	log.Printf("starting worker nodes...")
	go runNode("worker-1", stop)
	go runNode("worker-2", stop)

	// Wait for shutdown signal
	<-stop
	log.Printf("demo complete")
}
