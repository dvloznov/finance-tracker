package inmemory

import (
	"context"
	"fmt"
	"sync"

	"github.com/dvloznov/finance-tracker/internal/jobs"
)

// Store is an in-memory implementation of JobStore.
// It stores jobs in memory and is safe for concurrent use.
// Data is lost on service restart - for persistence, use a database-backed store.
type Store struct {
	mu   sync.RWMutex
	jobs map[string]*jobs.ParseDocumentJob
}

// NewStore creates a new in-memory job store.
func NewStore() *Store {
	return &Store{
		jobs: make(map[string]*jobs.ParseDocumentJob),
	}
}

// SaveJob implements the JobStore interface.
// It saves or updates a job in memory.
func (s *Store) SaveJob(ctx context.Context, job *jobs.ParseDocumentJob) error {
	if job.JobID == "" {
		return fmt.Errorf("job ID is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a copy to avoid external modifications
	jobCopy := *job
	s.jobs[job.JobID] = &jobCopy

	return nil
}

// GetJob implements the JobStore interface.
// It retrieves a job by ID from memory.
func (s *Store) GetJob(ctx context.Context, jobID string) (*jobs.ParseDocumentJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	// Return a copy to avoid external modifications
	jobCopy := *job
	return &jobCopy, nil
}

// ListJobs implements the JobStore interface.
// It retrieves jobs with optional filtering from memory.
func (s *Store) ListJobs(ctx context.Context, filter jobs.JobFilter) ([]*jobs.ParseDocumentJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*jobs.ParseDocumentJob

	for _, job := range s.jobs {
		// Apply filters
		if filter.DocumentID != "" && job.DocumentID != filter.DocumentID {
			continue
		}
		if filter.Status != "" && job.Status != filter.Status {
			continue
		}

		// Create a copy to avoid external modifications
		jobCopy := *job
		result = append(result, &jobCopy)
	}

	// Apply limit and offset
	if filter.Offset > 0 {
		if filter.Offset >= len(result) {
			return []*jobs.ParseDocumentJob{}, nil
		}
		result = result[filter.Offset:]
	}

	if filter.Limit > 0 && filter.Limit < len(result) {
		result = result[:filter.Limit]
	}

	return result, nil
}

// UpdateJobStatus implements the JobStore interface.
// It updates the status of a job in memory.
func (s *Store) UpdateJobStatus(ctx context.Context, jobID string, status jobs.JobStatus, errorMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	job.Status = status
	if errorMsg != "" {
		job.Error = errorMsg
	}

	return nil
}

// Ensure Store implements JobStore interface.
var _ jobs.JobStore = (*Store)(nil)
