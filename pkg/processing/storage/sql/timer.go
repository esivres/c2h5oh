package sql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

type timerRepo struct {
	q querier
}

func (r *timerRepo) Create(ctx context.Context, timer *storage.Timer) error {
	_, err := r.q.ExecContext(ctx,
		`INSERT INTO timers (key_id, process_instance_key, element_instance_key, process_definition_key, state, due_date, repetitions, cycle_duration_ns, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		timer.Key, timer.ProcessInstanceKey, timer.ElementInstanceKey, timer.ProcessDefinitionKey,
		timer.State, timer.DueDate.UnixMilli(), timer.Repetitions, int64(timer.CycleDuration), timer.CreatedAt.UnixMilli(),
	)
	return err
}

func (r *timerRepo) GetByKey(ctx context.Context, key uint64) (*storage.Timer, error) {
	timer := &storage.Timer{}
	var dueDate, createdAt, cycleDurNs int64
	err := r.q.QueryRowContext(ctx,
		`SELECT key_id, process_instance_key, element_instance_key, process_definition_key, state, due_date, repetitions, cycle_duration_ns, created_at
		 FROM timers WHERE key_id = ?`, key).
		Scan(&timer.Key, &timer.ProcessInstanceKey, &timer.ElementInstanceKey, &timer.ProcessDefinitionKey,
			&timer.State, &dueDate, &timer.Repetitions, &cycleDurNs, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	timer.DueDate = time.UnixMilli(dueDate)
	timer.CycleDuration = time.Duration(cycleDurNs)
	timer.CreatedAt = time.UnixMilli(createdAt)
	return timer, nil
}

func (r *timerRepo) Trigger(ctx context.Context, key uint64) error {
	_, err := r.q.ExecContext(ctx,
		`UPDATE timers SET state = ? WHERE key_id = ?`, storage.TimerTriggered, key)
	return err
}

func (r *timerRepo) Cancel(ctx context.Context, key uint64) error {
	_, err := r.q.ExecContext(ctx,
		`UPDATE timers SET state = ? WHERE key_id = ?`, storage.TimerCanceled, key)
	return err
}

func (r *timerRepo) FindDue(ctx context.Context, now time.Time, limit int) ([]*storage.Timer, error) {
	rows, err := r.q.QueryContext(ctx,
		`SELECT key_id, process_instance_key, element_instance_key, process_definition_key, state, due_date, repetitions, cycle_duration_ns, created_at
		 FROM timers WHERE state = ? AND due_date <= ? ORDER BY due_date LIMIT ?`,
		storage.TimerCreated, now.UnixMilli(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanRows(rows)
}

func (r *timerRepo) FindByProcessInstance(ctx context.Context, piKey uint64) ([]*storage.Timer, error) {
	rows, err := r.q.QueryContext(ctx,
		`SELECT key_id, process_instance_key, element_instance_key, process_definition_key, state, due_date, repetitions, cycle_duration_ns, created_at
		 FROM timers WHERE process_instance_key = ?`, piKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanRows(rows)
}

func (*timerRepo) scanRows(rows *sql.Rows) ([]*storage.Timer, error) {
	var result []*storage.Timer
	for rows.Next() {
		timer := &storage.Timer{}
		var dueDate, createdAt, cycleDurNs int64
		if err := rows.Scan(&timer.Key, &timer.ProcessInstanceKey, &timer.ElementInstanceKey, &timer.ProcessDefinitionKey,
			&timer.State, &dueDate, &timer.Repetitions, &cycleDurNs, &createdAt); err != nil {
			return nil, err
		}
		timer.DueDate = time.UnixMilli(dueDate)
		timer.CycleDuration = time.Duration(cycleDurNs)
		timer.CreatedAt = time.UnixMilli(createdAt)
		result = append(result, timer)
	}
	return result, rows.Err()
}
