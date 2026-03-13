package sql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// querier abstracts *sql.DB and *sql.Tx — both can ExecContext/QueryContext.
type querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Store is the SQL implementation of storage.Store.
type Store struct {
	db *sql.DB
}

// NewStore creates a new SQL store.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// txStore wraps a transaction and provides repository access within it.
type txStore struct {
	tx *sql.Tx
}

func (s *Store) ProcessDefinitions() storage.ProcessDefinitionRepository {
	return &processDefinitionRepo{q: s.db}
}

func (s *Store) ProcessInstances() storage.ProcessInstanceRepository {
	return &processInstanceRepo{q: s.db}
}

func (s *Store) Variables() storage.VariableRepository {
	return &variableRepo{q: s.db}
}

func (s *Store) Jobs() storage.JobRepository {
	return &jobRepo{q: s.db}
}

func (s *Store) Timers() storage.TimerRepository {
	return &timerRepo{q: s.db}
}

func (s *Store) MessageSubscriptions() storage.MessageSubscriptionRepository {
	return &messageSubscriptionRepo{q: s.db}
}

func (s *Store) Incidents() storage.IncidentRepository {
	return &incidentRepo{q: s.db}
}

func (s *Store) Forms() storage.FormRepository {
	return &formRepo{q: s.db}
}

// GetMaxKeyId returns the maximum key_id across all tables.
// Used to restore the key generator sequence after restart.
func (s *Store) GetMaxKeyId(ctx context.Context) (uint64, error) {
	var maxKey uint64
	tables := []string{
		"process_definitions",
		"process_instances",
		"element_instances",
		"variables",
		"jobs",
		"timers",
		"message_subscriptions",
		"message_buffer",
		"incidents",
		"forms",
	}
	for _, table := range tables {
		var k sql.NullInt64
		// table names are hardcoded constants, not user input
		err := s.db.QueryRowContext(ctx, "SELECT MAX(key_id) FROM "+table).Scan(&k)
		if err != nil {
			return 0, err
		}
		if k.Valid && uint64(k.Int64) > maxKey {
			maxKey = uint64(k.Int64)
		}
	}
	return maxKey, nil
}

// Execute runs a handler within a database transaction.
// All repository operations inside the handler use the same transaction.
// Commits on success, rolls back on error.
func (s *Store) Execute(ctx context.Context, h storage.Handler) ([]intent.Intent, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}

	ts := &txStore{tx: tx}
	intents, err := h.Handle(ctx, ts)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return intents, nil
}

// txStore implements storage.Store using the transaction.
func (ts *txStore) ProcessDefinitions() storage.ProcessDefinitionRepository {
	return &processDefinitionRepo{q: ts.tx}
}

func (ts *txStore) ProcessInstances() storage.ProcessInstanceRepository {
	return &processInstanceRepo{q: ts.tx}
}

func (ts *txStore) Variables() storage.VariableRepository {
	return &variableRepo{q: ts.tx}
}

func (ts *txStore) Jobs() storage.JobRepository {
	return &jobRepo{q: ts.tx}
}

func (ts *txStore) Timers() storage.TimerRepository {
	return &timerRepo{q: ts.tx}
}

func (ts *txStore) MessageSubscriptions() storage.MessageSubscriptionRepository {
	return &messageSubscriptionRepo{q: ts.tx}
}

func (ts *txStore) Incidents() storage.IncidentRepository {
	return &incidentRepo{q: ts.tx}
}

func (ts *txStore) Forms() storage.FormRepository {
	return &formRepo{q: ts.tx}
}

func (ts *txStore) Execute(ctx context.Context, h storage.Handler) ([]intent.Intent, error) {
	// Already in a transaction — just delegate (nested call).
	return h.Handle(ctx, ts)
}
