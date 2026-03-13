package sql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

type incidentRepo struct {
	q querier
}

func (r *incidentRepo) Create(ctx context.Context, inc *storage.Incident) error {
	_, err := r.q.ExecContext(ctx,
		`INSERT INTO incidents (key_id, process_instance_key, element_instance_key, job_key, type, state, error_message, created_at, resolved_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		inc.Key, inc.ProcessInstanceKey, inc.ElementInstanceKey, inc.JobKey,
		inc.Type, inc.State, inc.ErrorMessage, inc.CreatedAt.UnixMilli(), inc.ResolvedAt.UnixMilli(),
	)
	return err
}

func (r *incidentRepo) GetByKey(ctx context.Context, key uint64) (*storage.Incident, error) {
	inc := &storage.Incident{}
	var createdAt, resolvedAt int64
	err := r.q.QueryRowContext(ctx,
		`SELECT key_id, process_instance_key, element_instance_key, job_key, type, state, error_message, created_at, resolved_at
		 FROM incidents WHERE key_id = ?`, key).
		Scan(&inc.Key, &inc.ProcessInstanceKey, &inc.ElementInstanceKey, &inc.JobKey,
			&inc.Type, &inc.State, &inc.ErrorMessage, &createdAt, &resolvedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	inc.CreatedAt = time.UnixMilli(createdAt)
	if resolvedAt > 0 {
		inc.ResolvedAt = time.UnixMilli(resolvedAt)
	}
	return inc, nil
}

func (r *incidentRepo) Resolve(ctx context.Context, key uint64) error {
	_, err := r.q.ExecContext(ctx,
		`UPDATE incidents SET state = ?, resolved_at = ? WHERE key_id = ?`,
		storage.IncidentResolved, time.Now().UnixMilli(), key)
	return err
}

func (r *incidentRepo) FindByProcessInstance(ctx context.Context, piKey uint64) ([]*storage.Incident, error) {
	rows, err := r.q.QueryContext(ctx,
		`SELECT key_id, process_instance_key, element_instance_key, job_key, type, state, error_message, created_at, resolved_at
		 FROM incidents WHERE process_instance_key = ?`, piKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanRows(rows)
}

func (r *incidentRepo) FindUnresolved(ctx context.Context, limit int) ([]*storage.Incident, error) {
	rows, err := r.q.QueryContext(ctx,
		`SELECT key_id, process_instance_key, element_instance_key, job_key, type, state, error_message, created_at, resolved_at
		 FROM incidents WHERE state = ? LIMIT ?`, storage.IncidentCreated, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanRows(rows)
}

func (r *incidentRepo) scanRows(rows *sql.Rows) ([]*storage.Incident, error) {
	var result []*storage.Incident
	for rows.Next() {
		inc := &storage.Incident{}
		var createdAt, resolvedAt int64
		if err := rows.Scan(&inc.Key, &inc.ProcessInstanceKey, &inc.ElementInstanceKey, &inc.JobKey,
			&inc.Type, &inc.State, &inc.ErrorMessage, &createdAt, &resolvedAt); err != nil {
			return nil, err
		}
		inc.CreatedAt = time.UnixMilli(createdAt)
		if resolvedAt > 0 {
			inc.ResolvedAt = time.UnixMilli(resolvedAt)
		}
		result = append(result, inc)
	}
	return result, rows.Err()
}
