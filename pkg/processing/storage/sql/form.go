package sql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

type formRepo struct {
	q querier
}

func (r *formRepo) Create(ctx context.Context, form *storage.FormDefinition) error {
	_, err := r.q.ExecContext(ctx,
		`INSERT INTO forms (key_id, form_id, version, content, content_hash, deployed_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		form.Key, form.FormId, form.Version, form.Content, form.ContentHash, form.DeployedAt.UnixMilli(),
	)
	return err
}

func (r *formRepo) FindByKey(ctx context.Context, key uint64) (*storage.FormDefinition, error) {
	return r.scanOne(ctx,
		`SELECT key_id, form_id, version, content, content_hash, deployed_at
		 FROM forms WHERE key_id = ?`, key)
}

func (r *formRepo) FindLatestByFormId(ctx context.Context, formId string) (*storage.FormDefinition, error) {
	return r.scanOne(ctx,
		`SELECT key_id, form_id, version, content, content_hash, deployed_at
		 FROM forms WHERE form_id = ? AND deleted_at = 0 ORDER BY version DESC LIMIT 1`, formId)
}

func (r *formRepo) FindByContentHash(ctx context.Context, formId string, hash []byte) (*storage.FormDefinition, error) {
	return r.scanOne(ctx,
		`SELECT key_id, form_id, version, content, content_hash, deployed_at
		 FROM forms WHERE form_id = ? AND content_hash = ? LIMIT 1`, formId, hash)
}

func (r *formRepo) Delete(ctx context.Context, key uint64) error {
	_, err := r.q.ExecContext(ctx,
		`UPDATE forms SET deleted_at = ? WHERE key_id = ?`,
		time.Now().UnixMilli(), key)
	return err
}

func (r *formRepo) scanOne(ctx context.Context, query string, args ...any) (*storage.FormDefinition, error) {
	form := &storage.FormDefinition{}
	var deployedAt int64
	err := r.q.QueryRowContext(ctx, query, args...).
		Scan(&form.Key, &form.FormId, &form.Version, &form.Content, &form.ContentHash, &deployedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	form.DeployedAt = time.UnixMilli(deployedAt)
	return form, nil
}
