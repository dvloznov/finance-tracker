package inmemory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dvloznov/finance-tracker/internal/jobs"
	"github.com/google/uuid"
)

// Queue is an in-memory implementation of job publisher and consumer.
// It uses Go channels for job distribution and is safe for concurrent use.
// This implementation is suitable for single-instance deployments and testing.
// For production multi-instance deployments, migrate to Cloud Tasks or Pub/Sub.
type Queue struct {
	jobChan   chan *jobs.ParseDocumentJob
	closeChan chan struct{}
	wg        sync.WaitGroup
	mu        sync.RWMutex
	store     jobs.JobStore
	closed    bool
}

// NewQueue creates a new in-memory job queue.
// bufferSize determines how many jobs can be queued before PublishParseDocument blocks.
func NewQueue(bufferSize int, store jobs.JobStore) *Queue {
	return &Queue{
		jobChan:   make(chan *jobs.ParseDocumentJob, bufferSize),
		closeChan: make(chan struct{}),
		store:     store,
	}
}

// PublishParseDocument implements the Publisher interface.
// It enqueues a document parsing job for asynchronous processing.
func (q *Queue) PublishParseDocument(ctx context.Context, job *jobs.ParseDocumentJob) error {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.closed {
		return fmt.Errorf("queue is closed")
	}

	// Generate job ID if not provided
	if job.JobID == "" {
		job.JobID = uuid.New().String()
	}

	// Set initial status and timestamp
	if job.Status == "" {
		job.Status = jobs.JobStatusPending
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	if job.MaxRetries == 0 {
		job.MaxRetries = 3 // Default retry count
	}

	// Save job to store
	if q.store != nil {
		if err := q.store.SaveJob(ctx, job); err != nil {
			return fmt.Errorf("failed to save job: %w", err)
		}
	}

	// Enqueue job with context cancellation support
	select {
	case q.jobChan <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-q.closeChan:
		return fmt.Errorf("queue is closed")
	}
}

// Start implements the Consumer interface.
// It starts consuming jobs from the queue and processes them using the provided handler.
// The handler is called concurrently for each job, up to workerCount workers.
func (q *Queue) Start(ctx context.Context, handler jobs.JobHandler) error {
	q.mu.RLock()
	if q.closed {
		q.mu.RUnlock()
		return fmt.Errorf("queue is closed")
	}
	q.mu.RUnlock()

	// Start worker goroutines
	workerCount := 5 // Configurable number of concurrent workers
	for i := 0; i < workerCount; i++ {
		q.wg.Add(1)
		go q.worker(ctx, handler)
	}

	return nil
}

// worker processes jobs from the queue.
func (q *Queue) worker(ctx context.Context, handler jobs.JobHandler) {
	defer q.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-q.closeChan:
			return
		case job := <-q.jobChan:
			if job == nil {
				return
			}

			q.processJob(ctx, job, handler)
		}
	}
}

// processJob executes a single job with retry logic.
func (q *Queue) processJob(ctx context.Context, job *jobs.ParseDocumentJob, handler jobs.JobHandler) {
	// Update job status to running
	job.Status = jobs.JobStatusRunning
	now := time.Now()
	job.StartedAt = &now

	if q.store != nil {
		_ = q.store.SaveJob(ctx, job)
	}

	// Execute the job handler
	err := handler(ctx, job)

	// Update job status based on result
	completedAt := time.Now()
	job.CompletedAt = &completedAt

	if err != nil {
		job.Error = err.Error()

		// Check if we should retry
		if job.RetryCount < job.MaxRetries {
			job.RetryCount++
			job.Status = jobs.JobStatusRetrying

			// Re-enqueue with exponential backoff
			backoff := time.Duration(job.RetryCount) * time.Second
			time.AfterFunc(backoff, func() {
				// Reset for retry
				job.Status = jobs.JobStatusPending
				job.StartedAt = nil
				job.CompletedAt = nil
				_ = q.PublishParseDocument(ctx, job)
			})
		} else {
			job.Status = jobs.JobStatusFailed
		}
	} else {
		job.Status = jobs.JobStatusCompleted
		job.Error = ""
	}

	if q.store != nil {
		_ = q.store.SaveJob(ctx, job)
	}
}

// Stop implements the Consumer interface.
// It stops the queue and waits for all in-flight jobs to complete.
func (q *Queue) Stop(ctx context.Context) error {
	q.mu.Lock()
	if q.closed {
		q.mu.Unlock()
		return nil
	}
	q.closed = true
	close(q.closeChan)
	q.mu.Unlock()

	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close implements the Publisher interface.
// It closes the queue and releases resources.
func (q *Queue) Close() error {
	return q.Stop(context.Background())
}

// Ensure Queue implements both Publisher and Consumer interfaces.
var _ jobs.Publisher = (*Queue)(nil)
var _ jobs.Consumer = (*Queue)(nil)
