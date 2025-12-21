package jobs

import (
	"context"
	"time"
)

// JobType represents the type of job to be executed.
type JobType string

const (
	// JobTypeParseDocument represents a document parsing job.
	JobTypeParseDocument JobType = "parse_document"
)

// JobStatus represents the current status of a job.
type JobStatus string

const (
	// JobStatusPending indicates the job is waiting to be processed.
	JobStatusPending JobStatus = "pending"
	// JobStatusRunning indicates the job is currently being processed.
	JobStatusRunning JobStatus = "running"
	// JobStatusCompleted indicates the job completed successfully.
	JobStatusCompleted JobStatus = "completed"
	// JobStatusFailed indicates the job failed.
	JobStatusFailed JobStatus = "failed"
	// JobStatusRetrying indicates the job failed and is being retried.
	JobStatusRetrying JobStatus = "retrying"
)

// ParseDocumentJob represents a job to parse a document from GCS.
type ParseDocumentJob struct {
	// JobID is the unique identifier for this job.
	JobID string `json:"job_id"`

	// DocumentID is the ID of the document in BigQuery.
	DocumentID string `json:"document_id"`

	// GCSURI is the GCS URI of the document to parse.
	GCSURI string `json:"gcs_uri"`

	// ParsingRunID is the ID of the parsing run in BigQuery.
	ParsingRunID string `json:"parsing_run_id,omitempty"`

	// Status is the current status of the job.
	Status JobStatus `json:"status"`

	// CreatedAt is when the job was created.
	CreatedAt time.Time `json:"created_at"`

	// StartedAt is when the job started processing.
	StartedAt *time.Time `json:"started_at,omitempty"`

	// CompletedAt is when the job completed (success or failure).
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Error contains error details if the job failed.
	Error string `json:"error,omitempty"`

	// RetryCount is the number of times this job has been retried.
	RetryCount int `json:"retry_count"`

	// MaxRetries is the maximum number of retries allowed.
	MaxRetries int `json:"max_retries"`
}

// Job is a generic interface for all job types.
type Job interface {
	// GetID returns the unique job identifier.
	GetID() string

	// GetType returns the job type.
	GetType() JobType

	// GetStatus returns the current job status.
	GetStatus() JobStatus
}

// GetID implements the Job interface.
func (j *ParseDocumentJob) GetID() string {
	return j.JobID
}

// GetType implements the Job interface.
func (j *ParseDocumentJob) GetType() JobType {
	return JobTypeParseDocument
}

// GetStatus implements the Job interface.
func (j *ParseDocumentJob) GetStatus() JobStatus {
	return j.Status
}

// Publisher defines the interface for publishing jobs to a queue.
// This abstraction allows for different queue implementations (in-memory, Cloud Tasks, Pub/Sub).
type Publisher interface {
	// PublishParseDocument publishes a document parsing job.
	PublishParseDocument(ctx context.Context, job *ParseDocumentJob) error

	// Close closes the publisher and releases resources.
	Close() error
}

// Consumer defines the interface for consuming jobs from a queue.
// This abstraction allows for different queue implementations (in-memory, Cloud Tasks, Pub/Sub).
type Consumer interface {
	// Start begins consuming jobs from the queue.
	// The handler function is called for each job received.
	Start(ctx context.Context, handler JobHandler) error

	// Stop stops consuming jobs and waits for in-flight jobs to complete.
	Stop(ctx context.Context) error
}

// JobHandler is a function that processes a job.
// It should return an error if the job failed and should be retried.
type JobHandler func(ctx context.Context, job Job) error

// JobStore defines the interface for storing and retrieving job status.
// This allows tracking job execution across service restarts.
type JobStore interface {
	// SaveJob saves or updates a job's state.
	SaveJob(ctx context.Context, job *ParseDocumentJob) error

	// GetJob retrieves a job by ID.
	GetJob(ctx context.Context, jobID string) (*ParseDocumentJob, error)

	// ListJobs retrieves jobs with optional filtering.
	ListJobs(ctx context.Context, filter JobFilter) ([]*ParseDocumentJob, error)

	// UpdateJobStatus updates the status of a job.
	UpdateJobStatus(ctx context.Context, jobID string, status JobStatus, errorMsg string) error
}

// JobFilter defines filtering criteria for listing jobs.
type JobFilter struct {
	// DocumentID filters jobs by document ID.
	DocumentID string

	// Status filters jobs by status.
	Status JobStatus

	// Limit limits the number of results.
	Limit int

	// Offset for pagination.
	Offset int
}
