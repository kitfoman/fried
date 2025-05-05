package server

import (
	"time"

	"github.com/NVIDIA/go-dcgm/pkg/dcgm"
)

// DiagnosticLevel represents the diagnostic intensity level
type DiagnosticLevel int

const (
	// DiagBasic is a quick basic diagnostic
	DiagBasic DiagnosticLevel = iota + 1
	// DiagNormal is a standard diagnostic
	DiagNormal
	// DiagExtensive is a comprehensive diagnostic
	DiagExtensive
)

// JobStatus represents the status of a diagnostic job
type JobStatus string

const (
	// StatusPending indicates the job is queued but not started
	StatusPending JobStatus = "pending"
	// StatusRunning indicates the job is currently running
	StatusRunning JobStatus = "running"
	// StatusCompleted indicates the job completed successfully
	StatusCompleted JobStatus = "completed"
	// StatusFailed indicates the job failed
	StatusFailed JobStatus = "failed"
)

// ScheduleRequest represents a request to schedule a diagnostic job
type ScheduleRequest struct {
	GPUIDs []int           `json:"gpuIds"`
	Level  DiagnosticLevel `json:"level"`
}

// ScheduleResponse contains the ID of the scheduled job
type ScheduleResponse struct {
	JobID string `json:"jobId"`
}

// StatusRequest represents a request to get the status of a diagnostic job
type StatusRequest struct {
	JobID string `json:"jobId"`
}

// StatusResponse contains the current status and results of a diagnostic job
type StatusResponse struct {
	JobID     string         `json:"jobId"`
	Status    JobStatus      `json:"status"`
	StartTime time.Time      `json:"startTime,omitempty"`
	EndTime   time.Time      `json:"endTime,omitempty"`
	Results   dcgm.DiagResults `json:"results,omitempty"`
	Error     string         `json:"error,omitempty"`
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Error string `json:"error"`
}