package sql

import (
	"context"
	"database/sql"
	"testing"

	"github.com/esivres/c2h5oh/pkg/processing/storage"

	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

func TestSQLStore_Conformance(t *testing.T) {
	db, err := sql.Open("sqlite", t.TempDir()+"/conformance.db")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	require.NoError(t, Migrate(context.Background(), db))

	store := NewStore(db)
	storage.RunConformanceSuite(t, store)
}
