package db

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

func (d *DB) CreateJob(jobType, payload string) (*Job, error) {
	j := &Job{
		ID:          uuid.NewString(),
		Type:        jobType,
		Payload:     payload,
		Status:      "pending",
		Attempts:    0,
		MaxAttempts: 3,
	}
	_, err := d.Conn.Exec(
		`INSERT INTO jobs (id, type, payload, status, attempts, max_attempts)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		j.ID, j.Type, j.Payload, j.Status, j.Attempts, j.MaxAttempts,
	)
	return j, err
}

func (d *DB) GetJob(id string) (*Job, error) {
	j := &Job{}
	var errMsg sql.NullString
	err := d.Conn.QueryRow(
		`SELECT id, type, payload, status, attempts, max_attempts, error, created_at, updated_at
		 FROM jobs WHERE id = ?`, id,
	).Scan(&j.ID, &j.Type, &j.Payload, &j.Status, &j.Attempts, &j.MaxAttempts, &errMsg, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		return nil, err
	}
	j.Error = errMsg.String
	return j, nil
}

func (d *DB) GetAllJobs() ([]Job, error) {
	rows, err := d.Conn.Query(
		`SELECT id, type, payload, status, attempts, max_attempts, error, created_at, updated_at
		 FROM jobs ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var j Job
		var errMsg sql.NullString
		rows.Scan(&j.ID, &j.Type, &j.Payload, &j.Status, &j.Attempts, &j.MaxAttempts, &errMsg, &j.CreatedAt, &j.UpdatedAt)
		j.Error = errMsg.String
		jobs = append(jobs, j)
	}
	return jobs, nil
}

func (d *DB) GetFailedJobs() ([]Job, error) {
	rows, err := d.Conn.Query(
		`SELECT id, type, payload, status, attempts, max_attempts, error, created_at, updated_at
		 FROM jobs WHERE status = 'failed' ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var j Job
		var errMsg sql.NullString
		rows.Scan(&j.ID, &j.Type, &j.Payload, &j.Status, &j.Attempts, &j.MaxAttempts, &errMsg, &j.CreatedAt, &j.UpdatedAt)
		j.Error = errMsg.String
		jobs = append(jobs, j)
	}
	return jobs, nil
}

func (d *DB) UpdateJobStatus(id, status, errMsg string) error {
	_, err := d.Conn.Exec(
		`UPDATE jobs SET status = ?, error = ?, updated_at = ? WHERE id = ?`,
		status, errMsg, time.Now(), id,
	)
	return err
}

func (d *DB) IncrementAttempts(id string) error {
	_, err := d.Conn.Exec(
		`UPDATE jobs SET attempts = attempts + 1, updated_at = ? WHERE id = ?`,
		time.Now(), id,
	)
	return err
}

func (d *DB) ResetJob(id string) error {
	_, err := d.Conn.Exec(
		`UPDATE jobs SET status = 'pending', attempts = 0, error = '', updated_at = ? WHERE id = ?`,
		time.Now(), id,
	)
	return err
}