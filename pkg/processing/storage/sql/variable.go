package sql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

type variableRepo struct {
	q querier
}

func (r *variableRepo) Create(ctx context.Context, v *storage.Variable) error {
	_, err := r.q.ExecContext(ctx,
		`INSERT INTO variables (key_id, process_instance_key, scope_key, name, value)
		 VALUES (?, ?, ?, ?, ?)`,
		v.Key, v.ProcessInstanceKey, v.ScopeKey, v.Name, v.Value,
	)
	return err
}

func (r *variableRepo) Update(ctx context.Context, key uint64, value []byte) error {
	_, err := r.q.ExecContext(ctx,
		`UPDATE variables SET value = ? WHERE key_id = ?`, value, key)
	return err
}

func (r *variableRepo) FindByScope(ctx context.Context, scopeKey uint64) ([]*storage.Variable, error) {
	rows, err := r.q.QueryContext(ctx,
		`SELECT key_id, process_instance_key, scope_key, name, value
		 FROM variables WHERE scope_key = ?`, scopeKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*storage.Variable
	for rows.Next() {
		v := &storage.Variable{}
		if err := rows.Scan(&v.Key, &v.ProcessInstanceKey, &v.ScopeKey, &v.Name, &v.Value); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func (r *variableRepo) FindByName(ctx context.Context, scopeKey uint64, name string) (*storage.Variable, error) {
	v := &storage.Variable{}
	err := r.q.QueryRowContext(ctx,
		`SELECT key_id, process_instance_key, scope_key, name, value
		 FROM variables WHERE scope_key = ? AND name = ?`, scopeKey, name).
		Scan(&v.Key, &v.ProcessInstanceKey, &v.ScopeKey, &v.Name, &v.Value)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (r *variableRepo) CopyToScope(ctx context.Context, sourceScopeKey, targetScopeKey uint64, names []string) error {
	vars, err := r.FindByScope(ctx, sourceScopeKey)
	if err != nil {
		return err
	}
	for _, v := range vars {
		if len(names) > 0 && !contains(names, v.Name) {
			continue
		}
		_, err := r.q.ExecContext(ctx,
			`INSERT INTO variables (key_id, process_instance_key, scope_key, name, value)
			 VALUES (?, ?, ?, ?, ?)
			 ON CONFLICT(scope_key, name) DO UPDATE SET value = excluded.value`,
			v.Key, v.ProcessInstanceKey, targetScopeKey, v.Name, v.Value,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *variableRepo) PropagateToParent(ctx context.Context, childScopeKey, parentScopeKey uint64, names []string) error {
	return r.CopyToScope(ctx, childScopeKey, parentScopeKey, names)
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
