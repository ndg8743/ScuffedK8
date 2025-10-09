package multithreaded

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// JobScheduler manages distributed job execution across multiple nodes
type JobScheduler struct {
	nodeManager *NodeManager
	workerPool  *WorkerPool
	jobs        map[string]*ScheduledJob
	mu          sync.RWMutex
}

// ScheduledJob represents a job being processed by the scheduler
type ScheduledJob struct {
	ID          string                    `json:"id"`
	Type        string                    `json:"type"`
	Status      string                    `json:"status"`
	CreatedAt   time.Time                 `json:"created_at"`
	StartedAt   *time.Time                `json:"started_at,omitempty"`
	CompletedAt *time.Time                `json:"completed_at,omitempty"`
	Progress    float64                   `json:"progress"`
	Subtasks    map[string]*SubTask       `json:"subtasks"`
	Results     map[string]*SubTaskResult `json:"results"`
	FinalResult interface{}               `json:"final_result,omitempty"`
	Error       string                    `json:"error,omitempty"`
	NodeIDs     []string                  `json:"node_ids"`
	mu          sync.RWMutex
}

// JobSchedulerStats holds statistics about job processing
type JobSchedulerStats struct {
	TotalJobs      int64         `json:"total_jobs"`
	CompletedJobs  int64         `json:"completed_jobs"`
	FailedJobs     int64         `json:"failed_jobs"`
	ActiveJobs     int           `json:"active_jobs"`
	AverageLatency time.Duration `json:"average_latency"`
}

// NewJobScheduler creates a new job scheduler
func NewJobScheduler(nodeManager *NodeManager, workerPool *WorkerPool) *JobScheduler {
	return &JobScheduler{
		nodeManager: nodeManager,
		workerPool:  workerPool,
		jobs:        make(map[string]*ScheduledJob),
	}
}

// SubmitJob submits a new job for distributed processing
func (js *JobScheduler) SubmitJob(job *Job) (*JobResult, error) {
	log.Printf("Submitting job %s of type %s", job.ID, job.Type)

	// Create scheduled job
	now := time.Now()
	scheduledJob := &ScheduledJob{
		ID:        job.ID,
		Type:      job.Type,
		Status:    "submitted",
		CreatedAt: job.CreatedAt,
		StartedAt: &now,
		Progress:  0.0,
		Subtasks:  make(map[string]*SubTask),
		Results:   make(map[string]*SubTaskResult),
	}

	// Store the job
	js.mu.Lock()
	js.jobs[job.ID] = scheduledJob
	js.mu.Unlock()

	// Process the job based on its type
	switch job.Type {
	case "pi_calculation":
		return js.processPiCalculationJob(job, scheduledJob)
	default:
		return nil, fmt.Errorf("unsupported job type: %s", job.Type)
	}
}

// processPiCalculationJob handles distributed pi calculation
func (js *JobScheduler) processPiCalculationJob(job *Job, scheduledJob *ScheduledJob) (*JobResult, error) {
	// Extract job parameters
	iterations, ok := job.Data.(map[string]interface{})["iterations"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid iterations parameter")
	}

	parallelism := job.Parallelism
	if parallelism <= 0 {
		parallelism = 1
	}

	// Get available nodes
	availableNodes := js.nodeManager.GetAvailableNodes()
	if len(availableNodes) == 0 {
		// Fall back to local worker pool
		return js.processJobLocally(job)
	}

	// Limit parallelism to available nodes
	if parallelism > len(availableNodes) {
		parallelism = len(availableNodes)
	}

	// Split the work into subtasks
	subtasks := js.splitPiCalculationJob(job.ID, int(iterations), parallelism)
	
	// Update scheduled job
	scheduledJob.mu.Lock()
	for _, subtask := range subtasks {
		scheduledJob.Subtasks[subtask.ID] = subtask
	}
	scheduledJob.Status = "running"
	scheduledJob.mu.Unlock()

	// Distribute subtasks to nodes
	nodeIndex := 0
	for _, subtask := range subtasks {
		nodeID := availableNodes[nodeIndex].ID
		subtask.NodeID = nodeID
		
		if err := js.nodeManager.AssignTaskToNode(nodeID, *subtask); err != nil {
			log.Printf("Failed to assign subtask %s to node %s: %v", subtask.ID, nodeID, err)
			continue
		}

		scheduledJob.NodeIDs = append(scheduledJob.NodeIDs, nodeID)
		nodeIndex = (nodeIndex + 1) % len(availableNodes)
	}

	// Wait for completion or timeout
	return js.waitForJobCompletion(scheduledJob, 60*time.Second)
}

// splitPiCalculationJob splits a pi calculation job into subtasks
func (js *JobScheduler) splitPiCalculationJob(jobID string, totalIterations int, parallelism int) []*SubTask {
	iterationsPerTask := totalIterations / parallelism
	subtasks := make([]*SubTask, parallelism)

	for i := 0; i < parallelism; i++ {
		start := i * iterationsPerTask
		end := start + iterationsPerTask
		if i == parallelism-1 {
			end = totalIterations // Handle remainder in last task
		}

		subtasks[i] = &SubTask{
			ID:       fmt.Sprintf("%s-subtask-%d", jobID, i),
			ParentID: jobID,
			Data: map[string]interface{}{
				"start_iteration": start,
				"end_iteration":   end,
				"task_type":      "pi_calculation_range",
			},
		}
	}

	return subtasks
}

// processJobLocally processes a job using the local worker pool
func (js *JobScheduler) processJobLocally(job *Job) (*JobResult, error) {
	log.Printf("Processing job %s locally using worker pool", job.ID)

	// Convert to worker pool job format
	workerJob := Job{
		ID:          job.ID,
		Type:        "distributed_pi",
		Data:        job.Data,
		CreatedAt:   job.CreatedAt,
		Priority:    1,
		Parallelism: job.Parallelism,
	}

	// Submit to worker pool
	if err := js.workerPool.SubmitJob(workerJob); err != nil {
		return nil, fmt.Errorf("failed to submit job to worker pool: %v", err)
	}

	// For local processing, we need to implement a result collection mechanism
	// This is a simplified version - in practice you'd want proper async handling
	return &JobResult{
		JobID:     job.ID,
		Status:    "completed",
		Result:    "Local processing initiated",
		Duration:  0,
		NodesUsed: []string{"local"},
	}, nil
}

// waitForJobCompletion waits for a distributed job to complete
func (js *JobScheduler) waitForJobCompletion(job *ScheduledJob, timeout time.Duration) (*JobResult, error) {
	start := time.Now()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	timeoutCh := time.After(timeout)

	for {
		select {
		case <-timeoutCh:
			js.markJobFailed(job, "timeout waiting for completion")
			return &JobResult{
				JobID:     job.ID,
				Status:    "timeout",
				Duration:  time.Since(start),
				NodesUsed: job.NodeIDs,
			}, fmt.Errorf("job %s timed out", job.ID)

		case <-ticker.C:
			if js.checkJobCompletion(job) {
				result := js.aggregateJobResults(job)
				js.markJobCompleted(job, result)
				
				return &JobResult{
					JobID:     job.ID,
					Status:    "completed",
					Result:    result,
					Duration:  time.Since(start),
					NodesUsed: job.NodeIDs,
				}, nil
			}
		}
	}
}

// checkJobCompletion checks if all subtasks have completed
func (js *JobScheduler) checkJobCompletion(job *ScheduledJob) bool {
	job.mu.RLock()
	defer job.mu.RUnlock()

	return len(job.Results) == len(job.Subtasks)
}

// aggregateJobResults combines results from all subtasks
func (js *JobScheduler) aggregateJobResults(job *ScheduledJob) interface{} {
	job.mu.RLock()
	defer job.mu.RUnlock()

	switch job.Type {
	case "pi_calculation":
		totalPi := 0.0
		successCount := 0

		for _, result := range job.Results {
			if result.Success {
				if piValue, ok := result.Result.(float64); ok {
					totalPi += piValue
					successCount++
				}
			}
		}

		return map[string]interface{}{
			"pi":                totalPi,
			"successful_tasks":  successCount,
			"total_tasks":      len(job.Subtasks),
		}
	}

	return nil
}

// markJobCompleted marks a job as completed
func (js *JobScheduler) markJobCompleted(job *ScheduledJob, result interface{}) {
	now := time.Now()
	job.mu.Lock()
	job.Status = "completed"
	job.CompletedAt = &now
	job.Progress = 1.0
	job.FinalResult = result
	job.mu.Unlock()

	log.Printf("Job %s completed successfully", job.ID)
}

// markJobFailed marks a job as failed
func (js *JobScheduler) markJobFailed(job *ScheduledJob, errorMsg string) {
	now := time.Now()
	job.mu.Lock()
	job.Status = "failed"
	job.CompletedAt = &now
	job.Error = errorMsg
	job.mu.Unlock()

	log.Printf("Job %s failed: %s", job.ID, errorMsg)
}

// RecordSubtaskResult records a result from a completed subtask
func (js *JobScheduler) RecordSubtaskResult(result *SubTaskResult) {
	js.mu.RLock()
	job, exists := js.jobs[result.ParentID]
	js.mu.RUnlock()

	if !exists {
		log.Printf("Received result for unknown job: %s", result.ParentID)
		return
	}

	job.mu.Lock()
	job.Results[result.SubTaskID] = result
	
	// Update progress
	if len(job.Subtasks) > 0 {
		job.Progress = float64(len(job.Results)) / float64(len(job.Subtasks))
	}
	job.mu.Unlock()

	log.Printf("Recorded result for subtask %s of job %s", result.SubTaskID, result.ParentID)
}

// GetJobStatus returns the current status of a job
func (js *JobScheduler) GetJobStatus(jobID string) *JobStatus {
	js.mu.RLock()
	job, exists := js.jobs[jobID]
	js.mu.RUnlock()

	if !exists {
		return nil
	}

	job.mu.RLock()
	defer job.mu.RUnlock()

	status := &JobStatus{
		JobID:     job.ID,
		Status:    job.Status,
		Progress:  job.Progress,
		StartedAt: job.CreatedAt,
	}

	if job.CompletedAt != nil {
		status.CompletedAt = job.CompletedAt
	}

	if job.FinalResult != nil {
		status.Result = job.FinalResult
	}

	return status
}

// GetStats returns scheduler statistics
func (js *JobScheduler) GetStats() JobSchedulerStats {
	js.mu.RLock()
	defer js.mu.RUnlock()

	var completed, failed, active int64

	for _, job := range js.jobs {
		job.mu.RLock()
		switch job.Status {
		case "completed":
			completed++
		case "failed":
			failed++
		case "running", "submitted":
			active++
		}
		job.mu.RUnlock()
	}

	return JobSchedulerStats{
		TotalJobs:     int64(len(js.jobs)),
		CompletedJobs: completed,
		FailedJobs:    failed,
		ActiveJobs:    int(active),
	}
}