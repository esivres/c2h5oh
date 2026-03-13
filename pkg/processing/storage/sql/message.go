package sql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

type messageSubscriptionRepo struct {
	q querier
}

func (r *messageSubscriptionRepo) CreateSubscription(ctx context.Context, sub *storage.MessageSubscription) error {
	_, err := r.q.ExecContext(ctx,
		`INSERT INTO message_subscriptions (key_id, process_instance_key, element_instance_key, message_name, correlation_key, state, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		sub.Key, sub.ProcessInstanceKey, sub.ElementInstanceKey, sub.MessageName, sub.CorrelationKey, sub.State, sub.CreatedAt.UnixMilli(),
	)
	return err
}

func (r *messageSubscriptionRepo) Correlate(ctx context.Context, key uint64) error {
	_, err := r.q.ExecContext(ctx,
		`UPDATE message_subscriptions SET state = ? WHERE key_id = ?`,
		storage.MessageSubscriptionCorrelated, key)
	return err
}

func (r *messageSubscriptionRepo) CloseSubscription(ctx context.Context, key uint64) error {
	_, err := r.q.ExecContext(ctx,
		`UPDATE message_subscriptions SET state = ? WHERE key_id = ?`,
		storage.MessageSubscriptionClosed, key)
	return err
}

func (r *messageSubscriptionRepo) FindOpenSubscriptions(ctx context.Context, messageName string, correlationKey string) ([]*storage.MessageSubscription, error) {
	rows, err := r.q.QueryContext(ctx,
		`SELECT key_id, process_instance_key, element_instance_key, message_name, correlation_key, state, created_at
		 FROM message_subscriptions WHERE message_name = ? AND correlation_key = ? AND state = ?`,
		messageName, correlationKey, storage.MessageSubscriptionOpened)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanRows(rows)
}

func (r *messageSubscriptionRepo) FindByProcessInstance(ctx context.Context, piKey uint64) ([]*storage.MessageSubscription, error) {
	rows, err := r.q.QueryContext(ctx,
		`SELECT key_id, process_instance_key, element_instance_key, message_name, correlation_key, state, created_at
		 FROM message_subscriptions WHERE process_instance_key = ?`, piKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanRows(rows)
}

func (r *messageSubscriptionRepo) scanRows(rows *sql.Rows) ([]*storage.MessageSubscription, error) {
	var result []*storage.MessageSubscription
	for rows.Next() {
		sub := &storage.MessageSubscription{}
		var createdAt int64
		if err := rows.Scan(&sub.Key, &sub.ProcessInstanceKey, &sub.ElementInstanceKey,
			&sub.MessageName, &sub.CorrelationKey, &sub.State, &createdAt); err != nil {
			return nil, err
		}
		sub.CreatedAt = time.UnixMilli(createdAt)
		result = append(result, sub)
	}
	return result, rows.Err()
}

func (r *messageSubscriptionRepo) BufferMessage(ctx context.Context, msg *storage.MessageBuffer) error {
	_, err := r.q.ExecContext(ctx,
		`INSERT INTO message_buffer (key_id, message_name, correlation_key, variables, expires_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		msg.Key, msg.MessageName, msg.CorrelationKey, msg.Variables, msg.ExpiresAt.UnixMilli(), msg.CreatedAt.UnixMilli(),
	)
	return err
}

func (r *messageSubscriptionRepo) FindBufferedMessages(ctx context.Context, messageName string, correlationKey string) ([]*storage.MessageBuffer, error) {
	rows, err := r.q.QueryContext(ctx,
		`SELECT key_id, message_name, correlation_key, variables, expires_at, created_at
		 FROM message_buffer WHERE message_name = ? AND correlation_key = ? AND expires_at > ?`,
		messageName, correlationKey, time.Now().UnixMilli())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*storage.MessageBuffer
	for rows.Next() {
		msg := &storage.MessageBuffer{}
		var expiresAt, createdAt int64
		if err := rows.Scan(&msg.Key, &msg.MessageName, &msg.CorrelationKey, &msg.Variables, &expiresAt, &createdAt); err != nil {
			return nil, err
		}
		msg.ExpiresAt = time.UnixMilli(expiresAt)
		msg.CreatedAt = time.UnixMilli(createdAt)
		result = append(result, msg)
	}
	return result, rows.Err()
}

func (r *messageSubscriptionRepo) CleanExpired(ctx context.Context, now time.Time) (int64, error) {
	res, err := r.q.ExecContext(ctx,
		`DELETE FROM message_buffer WHERE expires_at <= ?`, now.UnixMilli())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// scanOne for message subscription
func (r *messageSubscriptionRepo) scanOne(row *sql.Row) (*storage.MessageSubscription, error) {
	sub := &storage.MessageSubscription{}
	var createdAt int64
	err := row.Scan(&sub.Key, &sub.ProcessInstanceKey, &sub.ElementInstanceKey,
		&sub.MessageName, &sub.CorrelationKey, &sub.State, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	sub.CreatedAt = time.UnixMilli(createdAt)
	return sub, nil
}
