package db

import "time"

type Job struct {
	ID          string
	Type        string
	Payload     string
	Status      string
	Attempts    int
	MaxAttempts int
	Error       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}