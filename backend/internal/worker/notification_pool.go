package worker

import (
	"fmt"
	"sync"
	"time"

	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
)

// NotificationJob represents a task to be processed
type NotificationJob struct {
	Notification *domain.Notification
	DestEmail    string // In real app, we'd need email. We'll simulate it.
}

// WorkerPool manages a set of workers to process jobs
type WorkerPool struct {
	JobQueue    chan NotificationJob
	WorkerCount int
	wg          sync.WaitGroup
	quit        chan bool
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workerCount int, bufferSize int) *WorkerPool {
	return &WorkerPool{
		JobQueue:    make(chan NotificationJob, bufferSize),
		WorkerCount: workerCount,
		quit:        make(chan bool),
	}
}

// Start spins up the workers
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.WorkerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i + 1)
	}
	fmt.Printf("WorkerPool started with %d workers\n", wp.WorkerCount)
}

// Stop waits for all jobs to finish and stops workers
func (wp *WorkerPool) Stop() {
	fmt.Println("WorkerPool stopping...")
	close(wp.JobQueue) // Close channel to signal no more jobs
	wp.wg.Wait()       // Wait for all workers to finish processing
	fmt.Println("WorkerPool stopped")
}

// Submit adds a job to the queue
func (wp *WorkerPool) Submit(job NotificationJob) {
	select {
	case wp.JobQueue <- job:
		// Job submitted successfully
	default:
		// Queue is full, handle accordingly (drop or log)
		fmt.Printf("WorkerPool queue full, dropping notification for user %s\n", job.Notification.UserID)
	}
}

// worker represents a single worker routine
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	fmt.Printf("Worker %d started\n", id)

	for job := range wp.JobQueue {
		wp.process(id, job)
	}

	fmt.Printf("Worker %d stopped\n", id)
}

// process simulates sending a notification (e.g., email or push)
func (wp *WorkerPool) process(workerID int, job NotificationJob) {
	fmt.Printf("[Worker %d] Processing notification: %s\n", workerID, job.Notification.Title)

	// Simulate IO latency (e.g., connecting to SMTP server)
	time.Sleep(2 * time.Second)

	fmt.Printf("[Worker %d] SENT notification to %s (User: %s)\n", workerID, job.DestEmail, job.Notification.UserID)
}
