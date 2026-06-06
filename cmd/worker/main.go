package main

import (
	"log"
	"math"
	"os"
	"time"
	"fmt"

	"github.com/tarunbtw/job-queue/internal/db"
	"github.com/tarunbtw/job-queue/internal/queue"
)

type Worker struct {
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
	w := &Worker{db: database, queue: q}

	log.Println("worker started, waiting for jobs...")
	w.run()
}

func (w *Worker) run() {
	for {
		jobID, err := w.queue.Pop()
		if err != nil {
			log.Println("queue error:", err)
			time.Sleep(2 * time.Second)
			continue
		}
		go w.process(jobID)
	}
}

func (w *Worker) process(jobID string) {
	job, err := w.db.GetJob(jobID)
	if err != nil {
		log.Println("job not found:", jobID)
		return
	}

	log.Printf("processing job %s type=%s attempt=%d\n", job.ID, job.Type, job.Attempts+1)

	w.db.UpdateJobStatus(job.ID, "processing", "")
	w.db.IncrementAttempts(job.ID)

	err = handle(job.Type, job.Payload)

	if err == nil {
		w.db.UpdateJobStatus(job.ID, "completed", "")
		log.Printf("job %s completed\n", job.ID)
		return
	}

	// job failed
	log.Printf("job %s failed: %s\n", job.ID, err.Error())

	job, _ = w.db.GetJob(job.ID)
	if job.Attempts >= job.MaxAttempts {
		w.db.UpdateJobStatus(job.ID, "failed", err.Error())
		log.Printf("job %s exhausted retries, moved to DLQ\n", job.ID)
		return
	}

	// exponential backoff before requeue
	delay := time.Duration(math.Pow(2, float64(job.Attempts))) * time.Second
	log.Printf("job %s retrying in %v\n", job.ID, delay)
	time.Sleep(delay)

	w.db.UpdateJobStatus(job.ID, "pending", "")
	w.queue.Push(job.ID)
}

// handle simulates different job types
func handle(jobType, payload string) error {
	switch jobType {
	case "email":
		log.Println("sending email with payload:", payload)
		time.Sleep(500 * time.Millisecond)
		return nil

	case "report":
		log.Println("generating report with payload:", payload)
		time.Sleep(1 * time.Second)
		return nil

	case "fail":
		// always fails — useful for testing DLQ
		return fmt.Errorf("intentional failure for testing")

	default:
		log.Println("unknown job type:", jobType)
		time.Sleep(200 * time.Millisecond)
		return nil
	}
}