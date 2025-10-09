package multithreaded
package multithreaded

import (
	"context"
	"fmt"
	"log"
	"math"
	"runtime"
	"sync"
	"time"
)

// WorkerPool manages a pool of worker goroutines for parallel job processing
type WorkerPool struct {
	workers     int
	jobQueue    chan Job
	resultQueue chan JobResult
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	stats       WorkerPoolStats
	mu          sync.RWMutex
	startTime   time.Time
}

// WorkerPoolStats holds statistics about the worker pool performance
type WorkerPoolStats struct {
	TotalWorkers   int           `json:"total_workers"`
	ActiveWorkers  int           `json:"active_workers"`
	QueueSize      int           `json:"queue_size"`
	ProcessedJobs  int64         `json:"processed_jobs"`
	AverageLatency time.Duration `json:"average_latency"`
	TotalLatency   time.Duration
}

// Job represents a unit of work that can be processed by workers
type Job struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"`
	Data        interface{} `json:"data"`
	CreatedAt   time.Time   `json:"created_at"`
	Priority    int         `json:"priority"`
	Parallelism int         `json:"parallelism"`
}

// JobResult holds the outcome of a job execution
type JobResult struct {
	JobID     string        `json:"job_id"`
	Success   bool          `json:"success"`
	Result    interface{}   `json:"result"`
	Error     string        `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	WorkerID  int           `json:"worker_id"`
	Timestamp time.Time     `json:"timestamp"`
}

// SubTask represents a portion of work split from a main job
type SubTask struct {
	ID       string      `json:"id"`
	ParentID string      `json:"parent_id"`
	Data     interface{} `json:"data"`
	NodeID   string      `json:"node_id"`
}

// SubTaskResult holds the result from a subtask execution
type SubTaskResult struct {
	SubTaskID string        `json:"subtask_id"`
	ParentID  string        `json:"parent_id"`
	Success   bool          `json:"success"`
	Result    interface{}   `json:"result"`
	Error     string        `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	NodeID    string        `json:"node_id"`
}

// NewWorkerPool creates a new worker pool with the specified number of workers
func NewWorkerPool(workers int, jobBufferSize int) *WorkerPool {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		workers:     workers,
		jobQueue:    make(chan Job, jobBufferSize),
		resultQueue: make(chan JobResult, jobBufferSize),
		ctx:         ctx,
		cancel:      cancel,
		stats: WorkerPoolStats{
			TotalWorkers: workers,
		},
		startTime: time.Now(),
	}

	return pool
}

// Start initializes and starts all worker goroutines
func (wp *WorkerPool) Start() {
	log.Printf("Starting worker pool with %d workers", wp.workers)

	// Start worker goroutines
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}

	// Start result collector goroutine
	go wp.resultCollector()
}

// worker is the main goroutine function that processes jobs
func (wp *WorkerPool) worker(workerID int) {
	defer wp.wg.Done()
	
	log.Printf("Worker %d started", workerID)

	for {
		select {
		case <-wp.ctx.Done():
			log.Printf("Worker %d shutting down", workerID)
			return

		case job := <-wp.jobQueue:
			wp.incrementActiveWorkers()
			result := wp.processJob(job, workerID)
			wp.decrementActiveWorkers()

			// Send result to result queue
			select {
			case wp.resultQueue <- result:
			case <-wp.ctx.Done():
				return
			}
		}
	}
}

// processJob handles the actual job execution
func (wp *WorkerPool) processJob(job Job, workerID int) JobResult {
	start := time.Now()

	result := JobResult{
		JobID:     job.ID,
		WorkerID:  workerID,
		Timestamp: start,
	}

	switch job.Type {
	case "pi_calculation":
		piResult, err := wp.calculatePi(job.Data)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
		} else {
			result.Success = true
			result.Result = piResult
		}

	case "distributed_pi":
		// Handle distributed pi calculation
		distributedResult, err := wp.processDistributedPi(job)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
		} else {
			result.Success = true
			result.Result = distributedResult
		}

	default:
		result.Success = false
		result.Error = fmt.Sprintf("unknown job type: %s", job.Type)
	}

	result.Duration = time.Since(start)
	return result
}

// calculatePi performs pi calculation using Leibniz formula
func (wp *WorkerPool) calculatePi(data interface{}) (float64, error) {
	jobData, ok := data.(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid job data format")
	}

	iterations, ok := jobData["iterations"].(float64)
	if !ok {
		return 0, fmt.Errorf("iterations not found or invalid type")
	}

	pi := 0.0
	for i := 0; i < int(iterations); i++ {
		pi += math.Pow(-1, float64(i)) / (2*float64(i) + 1)
	}

	return pi * 4, nil
}

// processDistributedPi handles distributed pi calculation across multiple subtasks
func (wp *WorkerPool) processDistributedPi(job Job) (interface{}, error) {
	jobData, ok := job.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid job data format")
	}

	iterations, ok := jobData["iterations"].(float64)
	if !ok {
		return nil, fmt.Errorf("iterations not found")
	}

	parallelism := job.Parallelism
	if parallelism <= 0 {
		parallelism = runtime.NumCPU()
	}

	// Split work into subtasks
	iterationsPerTask := int(iterations) / parallelism
	results := make(chan float64, parallelism)
	errors := make(chan error, parallelism)

	var wg sync.WaitGroup

	// Start subtasks
	for i := 0; i < parallelism; i++ {
		wg.Add(1)
		go func(taskID int) {
			defer wg.Done()

			start := taskID * iterationsPerTask
			end := start + iterationsPerTask
			if taskID == parallelism-1 {
				end = int(iterations) // Handle remainder in last task
			}

			pi := 0.0
			for j := start; j < end; j++ {
				pi += math.Pow(-1, float64(j)) / (2*float64(j) + 1)
			}

			results <- pi * 4
		}(i)
	}

	// Wait for all subtasks to complete
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	// Collect results
	totalPi := 0.0
	completedTasks := 0

	for result := range results {
		totalPi += result
		completedTasks++
	}

	// Check for errors
	select {
	case err := <-errors:
		if err != nil {
			return nil, err
		}
	default:
	}

	return map[string]interface{}{
		"pi":               totalPi,
		"completed_tasks":  completedTasks,
		"parallelism_used": parallelism,
	}, nil
}

// resultCollector processes job results and updates statistics
func (wp *WorkerPool) resultCollector() {
	for {
		select {
		case <-wp.ctx.Done():
			return

		case result := <-wp.resultQueue:
			wp.updateStats(result)
			log.Printf("Job %s completed by worker %d in %v", 
				result.JobID, result.WorkerID, result.Duration)
		}
	}
}

// SubmitJob adds a new job to the job queue
func (wp *WorkerPool) SubmitJob(job Job) error {
	select {
	case wp.jobQueue <- job:
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is shutting down")
	default:
		return fmt.Errorf("job queue is full")
	}
}

// GetStats returns current worker pool statistics
func (wp *WorkerPool) GetStats() WorkerPoolStats {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	stats := wp.stats
	stats.QueueSize = len(wp.jobQueue)
	return stats
}

// GetActiveWorkerCount returns the number of currently active workers
func (wp *WorkerPool) GetActiveWorkerCount() int {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return wp.stats.ActiveWorkers
}

// GetStartTime returns when the worker pool was started
func (wp *WorkerPool) GetStartTime() time.Time {
	return wp.startTime
}

// updateStats updates internal statistics with job result
func (wp *WorkerPool) updateStats(result JobResult) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	wp.stats.ProcessedJobs++
	wp.stats.TotalLatency += result.Duration

	if wp.stats.ProcessedJobs > 0 {
		wp.stats.AverageLatency = wp.stats.TotalLatency / time.Duration(wp.stats.ProcessedJobs)
	}
}

// incrementActiveWorkers increases the active worker count
func (wp *WorkerPool) incrementActiveWorkers() {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	wp.stats.ActiveWorkers++
}

// decrementActiveWorkers decreases the active worker count
func (wp *WorkerPool) decrementActiveWorkers() {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	wp.stats.ActiveWorkers--
}

// Shutdown gracefully shuts down the worker pool
func (wp *WorkerPool) Shutdown(timeout time.Duration) error {
	log.Printf("Shutting down worker pool...")

	// Stop accepting new jobs
	wp.cancel()

	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("Worker pool shutdown complete")
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("worker pool shutdown timed out after %v", timeout)
	}
}