package sql

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

type processDefinitionRepo struct {
	q querier
}

func (r *processDefinitionRepo) Create(ctx context.Context, def *storage.ProcessDefinition) error {
	_, err := r.q.ExecContext(ctx,
		`INSERT INTO process_definitions (key_id, bpmn_process_id, name, version, content_hash, content, deployed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		def.Key, def.BpmnProcessId, def.Name, def.Version, def.ContentHash, def.Content, def.DeployedAt.UnixMilli(),
	)
	return err
}

func (r *processDefinitionRepo) FindByKey(ctx context.Context, key uint64) (*storage.ProcessDefinition, error) {
	return r.scanOne(r.q.QueryRowContext(ctx,
		`SELECT key_id, bpmn_process_id, name, version, content_hash, content, deployed_at
		 FROM process_definitions WHERE key_id = ?`, key))
}

func (r *processDefinitionRepo) FindLatestByProcessId(ctx context.Context, bpmnProcessId string) (*storage.ProcessDefinition, error) {
	return r.scanOne(r.q.QueryRowContext(ctx,
		`SELECT key_id, bpmn_process_id, name, version, content_hash, content, deployed_at
		 FROM process_definitions WHERE bpmn_process_id = ? AND deleted_at = 0 ORDER BY version DESC LIMIT 1`, bpmnProcessId))
}

func (r *processDefinitionRepo) FindByProcessIdAndVersion(ctx context.Context, bpmnProcessId string, version uint64) (*storage.ProcessDefinition, error) {
	return r.scanOne(r.q.QueryRowContext(ctx,
		`SELECT key_id, bpmn_process_id, name, version, content_hash, content, deployed_at
		 FROM process_definitions WHERE bpmn_process_id = ? AND version = ?`, bpmnProcessId, version))
}

func (r *processDefinitionRepo) FindByContentHash(ctx context.Context, bpmnProcessId string, hash []byte) (*storage.ProcessDefinition, error) {
	rows, err := r.q.QueryContext(ctx,
		`SELECT key_id, bpmn_process_id, name, version, content_hash, content, deployed_at
		 FROM process_definitions WHERE bpmn_process_id = ? AND content_hash = ?`, bpmnProcessId, hash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		def, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		if bytes.Equal(def.ContentHash, hash) {
			return def, nil
		}
	}
	return nil, rows.Err()
}

func (r *processDefinitionRepo) GetLastVersion(ctx context.Context, bpmnProcessId string) (uint64, error) {
	var version sql.NullInt64
	err := r.q.QueryRowContext(ctx,
		`SELECT MAX(version) FROM process_definitions WHERE bpmn_process_id = ?`, bpmnProcessId).Scan(&version)
	if err != nil {
		return 0, err
	}
	if !version.Valid {
		return 0, nil
	}
	return uint64(version.Int64), nil
}

func (r *processDefinitionRepo) Delete(ctx context.Context, key uint64) error {
	_, err := r.q.ExecContext(ctx,
		`UPDATE process_definitions SET deleted_at = ? WHERE key_id = ?`,
		time.Now().UnixMilli(), key)
	return err
}

func (*processDefinitionRepo) scanOne(row *sql.Row) (*storage.ProcessDefinition, error) {
	def := &storage.ProcessDefinition{}
	var deployedAt int64
	err := row.Scan(&def.Key, &def.BpmnProcessId, &def.Name, &def.Version, &def.ContentHash, &def.Content, &deployedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	def.DeployedAt = time.UnixMilli(deployedAt)
	return def, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func (*processDefinitionRepo) scanRow(row rowScanner) (*storage.ProcessDefinition, error) {
	def := &storage.ProcessDefinition{}
	var deployedAt int64
	err := row.Scan(&def.Key, &def.BpmnProcessId, &def.Name, &def.Version, &def.ContentHash, &def.Content, &deployedAt)
	if err != nil {
		return nil, err
	}
	def.DeployedAt = time.UnixMilli(deployedAt)
	return def, nil
}
