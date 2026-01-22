package handlers

import (
	"net/http"
	"strconv"

	"github.com/dvloznov/finance-tracker/internal/api/middleware"
	"github.com/dvloznov/finance-tracker/internal/jobs"
	"github.com/rs/zerolog"
)

// JobsHandler handles job-related endpoints.
type JobsHandler struct {
	store jobs.JobStore
	log   zerolog.Logger
}

// NewJobsHandler creates a new jobs handler.
func NewJobsHandler(store jobs.JobStore, log zerolog.Logger) *JobsHandler {
	return &JobsHandler{
		store: store,
		log:   log,
	}
}

// GetJob handles GET /api/jobs/{id}
func (h *JobsHandler) GetJob(w http.ResponseWriter, r *http.Request, jobID string) {
	ctx := r.Context()

	job, err := h.store.GetJob(ctx, jobID)
	if err != nil {
		h.log.Error().Err(err).Str("job_id", jobID).Msg("Failed to get job")
		middleware.WriteError(w, http.StatusNotFound, "Job not found")
		return
	}

	middleware.WriteJSON(w, http.StatusOK, job)
}

// ListJobs handles GET /api/jobs
func (h *JobsHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	query := r.URL.Query()
	filter := jobs.JobFilter{
		DocumentID: query.Get("document_id"),
		Status:     jobs.JobStatus(query.Get("status")),
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	jobsList, err := h.store.ListJobs(ctx, filter)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to list jobs")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to list jobs")
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"jobs":  jobsList,
		"count": len(jobsList),
	})
}
