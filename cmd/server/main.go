package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/tarunbtw/job-queue/internal/db"
	"github.com/tarunbtw/job-queue/internal/queue"
)

type Server struct {
	db    *db.DB
	queue *queue.Queue
}

func main() {
	log.SetOutput(os.Stdout)

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "jobs.db"
	}

	database := db.New(dbPath)
	q := queue.New()
	s := &Server{db: database, queue: q}

	http.HandleFunc("/jobs", s.handleJobs)
	http.HandleFunc("/jobs/", s.handleJobActions)

	log.Println("server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (s *Server) handleJobs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.createJob(w, r)
	case http.MethodGet:
		s.listJobs(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) createJob(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Type    string         `json:"type"`
		Payload map[string]any `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Type == "" {
		http.Error(w, "invalid body, need {type, payload}", http.StatusBadRequest)
		return
	}

	payload, _ := json.Marshal(body.Payload)
	job, err := s.db.CreateJob(body.Type, string(payload))
	if err != nil {
		http.Error(w, "failed to create job", http.StatusInternalServerError)
		return
	}

	if err := s.queue.Push(job.ID); err != nil {
		http.Error(w, "failed to queue job", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(job)
}

func (s *Server) listJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := s.db.GetAllJobs()
	if err != nil {
		http.Error(w, "failed to fetch jobs", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func (s *Server) handleJobActions(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	// GET /jobs/:id
	if len(parts) == 2 && r.Method == http.MethodGet {
		job, err := s.db.GetJob(parts[1])
		if err != nil {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(job)
		return
	}

	// POST /jobs/:id/retry
	if len(parts) == 3 && parts[2] == "retry" && r.Method == http.MethodPost {
		if err := s.db.ResetJob(parts[1]); err != nil {
			http.Error(w, "failed to reset job", http.StatusInternalServerError)
			return
		}
		if err := s.queue.Push(parts[1]); err != nil {
			http.Error(w, "failed to requeue job", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "queued",
			"message": "job reset and requeued",
		})
		return
	}

	http.Error(w, "unknown route", http.StatusNotFound)
}