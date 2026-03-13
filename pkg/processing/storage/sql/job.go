package sql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

type jobRepo struct {
	q querier
}

func (r *jobRepo) Create(ctx context.Context, job *storage.Job) error {
	_, err := r.q.ExecContext(ctx,
		`INSERT INTO jobs (key_id, process_instance_key, element_instance_key, process_definition_key, type, state, retries, worker, deadline, error_message, error_code, variables, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		job.Key, job.ProcessInstanceKey, job.ElementInstanceKey, job.ProcessDefinitionKey,
		job.Type, job.State, job.Retries, job.Worker, job.Deadline.UnixMilli(),
		job.ErrorMessage, job.ErrorCode, job.Variables, job.CreatedAt.UnixMilli(),
	)
	return err
}

func (r *jobRepo) GetByKey(ctx context.Context, key uint64) (*storage.Job, error) {
	job := &storage.Job{}
	var deadline, createdAt int64
	err := r.q.QueryRowContext(ctx,
		`SELECT key_id, process_instance_key, element_instance_key, process_definition_key, type, state, retries, worker, deadline, error_message, error_code, variables, created_at
		 FROM jobs WHERE key_id = ?`, key).
		Scan(&job.Key, &job.ProcessInstanceKey, &job.ElementInstanceKey, &job.ProcessDefinitionKey,
			&job.Type, &job.State, &job.Retries, &job.Worker, &deadline,
			&job.ErrorMessage, &job.ErrorCode, &job.Variables, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	job.Deadline = time.UnixMilli(deadline)
	job.CreatedAt = time.UnixMilli(createdAt)
	return job, nil
}

func (r *jobRepo) Activate(ctx context.Context, key uint64, worker string, deadline time.Time) error {
	_, err := r.q.ExecContext(ctx,
		`UPDATE jobs SET state = ?, worker = ?, deadline = ? WHERE key_id = ?`,
		storage.JobActivated, worker, deadline.UnixMilli(), key)
	return err
}

func (r *jobRepo) Complete(ctx context.Context, key uint64, variables []byte) error {
	_, err := r.q.ExecContext(ctx,
		`UPDATE jobs SET state = ?, variables = ? WHERE key_id = ?`,
		storage.JobCompleted, variables, key)
	return err
}

func (r *jobRepo) Fail(ctx context.Context, key uint64, retries int, errorMessage string) error {
	_, err := r.q.ExecContext(ctx,
		`UPDATE jobs SET state = ?, retries = ?, error_message = ? WHERE key_id = ?`,
		storage.JobFailed, retries, errorMessage, key)
	return err
}

func (r *jobRepo) ThrowError(ctx context.Context, key uint64, errorCode string, errorMessage string) error {
	_, err := r.q.ExecContext(ctx,
		`UPDATE jobs SET state = ?, error_code = ?, error_message = ? WHERE key_id = ?`,
		storage.JobErrorThrown, errorCode, errorMessage, key)
	return err
}

func (r *jobRepo) UpdateRetries(ctx context.Context, key uint64, retries int) error {
	// Update retries; if job is in Failed state and new retries > 0, move to Created
	_, err := r.q.ExecContext(ctx,
		`UPDATE jobs SET retries = ?, state = CASE WHEN state = ? AND ? > 0 THEN ? ELSE state END WHERE key_id = ?`,
		retries, storage.JobFailed, retries, storage.JobCreated, key)
	return err
}

func (r *jobRepo) UpdateDeadline(ctx context.Context, key uint64, deadline time.Time) error {
	_, err := r.q.ExecContext(ctx,
		`UPDATE jobs SET deadline = ? WHERE key_id = ?`,
		deadline.UnixMilli(), key)
	return err
}

func (r *jobRepo) FindActivatable(ctx context.Context, jobType string, maxJobs int) ([]*storage.Job, error) {
	rows, err := r.q.QueryContext(ctx,
		`SELECT key_id, process_instance_key, element_instance_key, process_definition_key, type, state, retries, worker, deadline, error_message, error_code, variables, created_at
		 FROM jobs WHERE type = ? AND state = ? LIMIT ?`,
		jobType, storage.JobCreated, maxJobs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanRows(rows)
}

func (r *jobRepo) FindByProcessInstance(ctx context.Context, piKey uint64) ([]*storage.Job, error) {
	rows, err := r.q.QueryContext(ctx,
		`SELECT key_id, process_instance_key, element_instance_key, process_definition_key, type, state, retries, worker, deadline, error_message, error_code, variables, created_at
		 FROM jobs WHERE process_instance_key = ?`, piKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanRows(rows)
}

func (r *jobRepo) scanRows(rows *sql.Rows) ([]*storage.Job, error) {
	var result []*storage.Job
	for rows.Next() {
		job := &storage.Job{}
		var deadline, createdAt int64
		if err := rows.Scan(&job.Key, &job.ProcessInstanceKey, &job.ElementInstanceKey, &job.ProcessDefinitionKey,
			&job.Type, &job.State, &job.Retries, &job.Worker, &deadline,
			&job.ErrorMessage, &job.ErrorCode, &job.Variables, &createdAt); err != nil {
			return nil, err
		}
		job.Deadline = time.UnixMilli(deadline)
		job.CreatedAt = time.UnixMilli(createdAt)
		result = append(result, job)
	}
	return result, rows.Err()
}
