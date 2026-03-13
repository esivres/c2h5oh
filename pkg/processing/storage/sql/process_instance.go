package sql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

type processInstanceRepo struct {
	q querier
}

func (r *processInstanceRepo) CreateInstance(ctx context.Context, pi *storage.ProcessInstance) error {
	_, err := r.q.ExecContext(ctx,
		`INSERT INTO process_instances (key_id, process_definition_key, bpmn_process_id, parent_key, parent_element_key, state, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		pi.Key, pi.ProcessDefinitionKey, pi.BpmnProcessId, pi.ParentKey, pi.ParentElementKey, pi.State, pi.CreatedAt.UnixMilli(),
	)
	return err
}

func (r *processInstanceRepo) GetInstance(ctx context.Context, key uint64) (*storage.ProcessInstance, error) {
	pi := &storage.ProcessInstance{}
	var createdAt int64
	err := r.q.QueryRowContext(ctx,
		`SELECT key_id, process_definition_key, bpmn_process_id, parent_key, parent_element_key, state, created_at
		 FROM process_instances WHERE key_id = ?`, key).
		Scan(&pi.Key, &pi.ProcessDefinitionKey, &pi.BpmnProcessId, &pi.ParentKey, &pi.ParentElementKey, &pi.State, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	pi.CreatedAt = time.UnixMilli(createdAt)
	return pi, nil
}

func (r *processInstanceRepo) UpdateInstanceState(ctx context.Context, key uint64, state storage.ProcessInstanceState) error {
	_, err := r.q.ExecContext(ctx,
		`UPDATE process_instances SET state = ? WHERE key_id = ?`, state, key)
	return err
}

func (r *processInstanceRepo) CreateElementInstance(ctx context.Context, ei *storage.ElementInstance) error {
	_, err := r.q.ExecContext(ctx,
		`INSERT INTO element_instances (key_id, process_instance_key, process_definition_key, element_id, element_type, flow_scope_key, state, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		ei.Key, ei.ProcessInstanceKey, ei.ProcessDefinitionKey, ei.ElementId, ei.ElementType, ei.FlowScopeKey, ei.State, ei.CreatedAt.UnixMilli(),
	)
	return err
}

func (r *processInstanceRepo) GetElementInstance(ctx context.Context, key uint64) (*storage.ElementInstance, error) {
	ei := &storage.ElementInstance{}
	var createdAt int64
	err := r.q.QueryRowContext(ctx,
		`SELECT key_id, process_instance_key, process_definition_key, element_id, element_type, flow_scope_key, state, created_at
		 FROM element_instances WHERE key_id = ?`, key).
		Scan(&ei.Key, &ei.ProcessInstanceKey, &ei.ProcessDefinitionKey, &ei.ElementId, &ei.ElementType, &ei.FlowScopeKey, &ei.State, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ei.CreatedAt = time.UnixMilli(createdAt)
	return ei, nil
}

func (r *processInstanceRepo) UpdateElementInstanceState(ctx context.Context, key uint64, state storage.ElementInstanceState) error {
	_, err := r.q.ExecContext(ctx,
		`UPDATE element_instances SET state = ? WHERE key_id = ?`, state, key)
	return err
}

func (r *processInstanceRepo) FindElementInstancesByProcessInstance(ctx context.Context, piKey uint64) ([]*storage.ElementInstance, error) {
	rows, err := r.q.QueryContext(ctx,
		`SELECT key_id, process_instance_key, process_definition_key, element_id, element_type, flow_scope_key, state, created_at
		 FROM element_instances WHERE process_instance_key = ?`, piKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*storage.ElementInstance
	for rows.Next() {
		ei := &storage.ElementInstance{}
		var createdAt int64
		if err := rows.Scan(&ei.Key, &ei.ProcessInstanceKey, &ei.ProcessDefinitionKey, &ei.ElementId, &ei.ElementType, &ei.FlowScopeKey, &ei.State, &createdAt); err != nil {
			return nil, err
		}
		ei.CreatedAt = time.UnixMilli(createdAt)
		result = append(result, ei)
	}
	return result, rows.Err()
}
