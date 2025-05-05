package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/NVIDIA/go-dcgm/pkg/dcgm"
	"github.com/gorilla/mux"
)

// Server represents the GPU diagnostic server
type Server struct {
	listenAddr string
	router     *mux.Router
	httpServer *http.Server
	jobs       map[string]*JobInfo
	mu         sync.Mutex
	logger     *slog.Logger
}

// JobInfo contains information about a diagnostic job
type JobInfo struct {
	ID        string
	Status    JobStatus
	GPUIDs    []int
	Level     DiagnosticLevel
	StartTime time.Time
	EndTime   time.Time
	Results   dcgm.DiagResults
	Error     string
}

// NewServer creates a new diagnostic server
func NewServer(listenAddr string, logger *slog.Logger) *Server {
	s := &Server{
		listenAddr: listenAddr,
		jobs:       make(map[string]*JobInfo),
		router:     mux.NewRouter(),
		logger:     logger,
	}

	// Set up routes
	s.routes()

	return s
}

// routes sets up the HTTP routes
func (s *Server) routes() {
	s.router.HandleFunc("/schedule", s.handleSchedule).Methods("POST")
	s.router.HandleFunc("/status", s.handleStatus).Methods("GET")
}

func (s *Server) Start(ctx context.Context) error {
	s.httpServer = &http.Server{
		Addr:    s.listenAddr,
		Handler: s.router,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Error shutting down server", "error", err)
		}
	}()

	s.logger.Info("Starting server", "address", s.listenAddr)
	if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %v", err)
	}

	return nil
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, err string) error {
	return writeJSON(w, status, ErrorResponse{Error: err})
}
