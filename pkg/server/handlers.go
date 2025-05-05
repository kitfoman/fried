package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/NVIDIA/go-dcgm/pkg/dcgm"
	"github.com/google/uuid"
)

// handleSchedule handles requests to schedule a new diagnostic job
func (s *Server) handleSchedule(w http.ResponseWriter, r *http.Request) {
	var req ScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	if req.Level < DiagBasic || req.Level > DiagExtensive {
		writeError(w, http.StatusBadRequest, "Invalid diagnostic level")
		return
	}

	jobID := uuid.New().String()

	job := &JobInfo{
		ID:        jobID,
		Status:    StatusPending,
		GPUIDs:    req.GPUIDs,
		Level:     req.Level,
		StartTime: time.Time{}, // Will be set when job starts
		EndTime:   time.Time{}, // Will be set when job completes
	}

	// Store the job
	s.mu.Lock()
	s.jobs[jobID] = job
	s.mu.Unlock()

	// Start the job asynchronously
	go s.runDiagnostic(job)

	// Return the job ID
	resp := ScheduleResponse{
		JobID: jobID,
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleStatus handles requests to get the status of a diagnostic job
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	jobID := r.URL.Query().Get("jobId")
	if jobID == "" {
		writeError(w, http.StatusBadRequest, "Job ID not provided")
		return
	}

	// Get the job
	s.mu.Lock()
	job, exists := s.jobs[jobID]
	s.mu.Unlock()

	if !exists {
		writeError(w, http.StatusNotFound, "Job not found")
		return
	}

	// Return the job status
	resp := StatusResponse{
		JobID:     job.ID,
		Status:    job.Status,
		StartTime: job.StartTime,
		EndTime:   job.EndTime,
		Results:   job.Results,
		Error:     job.Error,
	}

	writeJSON(w, http.StatusOK, resp)
}

// runDiagnostic runs a diagnostic job
func (s *Server) runDiagnostic(job *JobInfo) {
	// Update job status to running
	s.mu.Lock()
	job.Status = StatusRunning
	job.StartTime = time.Now()
	s.mu.Unlock()

	s.logger.Info("Running diagnostic",
		"jobID", job.ID,
		"gpuIDs", job.GPUIDs,
		"level", job.Level,
	)

	var (
		gpuGroup dcgm.GroupHandle
		err      error
	)

	// Initialize DCGM if not already initialized
	cleanup, err := dcgm.Init(dcgm.Embedded)
	if err != nil {
		s.updateJobWithError(job, fmt.Sprintf("Failed to initialize DCGM: %v", err))
		return
	}
	defer cleanup()

	// Create a GPU group
	gpuGroup, err = dcgm.CreateGroup("diagnostic-group")
	if err != nil {
		s.updateJobWithError(job, fmt.Sprintf("Failed to create GPU group: %v", err))
		return
	}
	defer dcgm.DestroyGroup(gpuGroup)

	// use all gpus if none are specified
	if len(job.GPUIDs) == 0 {
		gpuGroup = dcgm.GroupAllGPUs()
		s.logger.Info("no gpu ids specified. using all gpus")
	}

	// Add GPUs to the group
	for _, gpuID := range job.GPUIDs {
		if err := dcgm.AddToGroup(gpuGroup, uint(gpuID)); err != nil {
			s.updateJobWithError(job, fmt.Sprintf("Failed to add GPU %d to group: %v", gpuID, err))
			return
		}
	}

	// Map our diagnostic level to DCGM's diagnostic type
	var diagType dcgm.DiagType
	switch job.Level {
	case DiagBasic:
		diagType = dcgm.DiagQuick
	case DiagNormal:
		diagType = dcgm.DiagMedium
	case DiagExtensive:
		diagType = dcgm.DiagLong
	default:
		s.updateJobWithError(job, "Invalid diagnostic level")
		return
	}

	// Run diagnostic
	results, err := dcgm.RunDiag(diagType, gpuGroup)
	if err != nil {
		s.updateJobWithError(job, fmt.Sprintf("Diagnostic failed: %v", err))
		return
	}

	// Update job with results
	s.mu.Lock()
	job.Status = StatusCompleted
	job.EndTime = time.Now()
	job.Results = results
	s.mu.Unlock()

	s.logger.Info("Diagnostic job completed", "jobID", job.ID)
}

// updateJobWithError updates a job with an error status
func (s *Server) updateJobWithError(job *JobInfo, errMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job.Status = StatusFailed
	job.EndTime = time.Now()
	job.Error = errMsg

	s.logger.Error("Diagnostic job failed",
		"jobID", job.ID,
		"error", errMsg,
	)
}
