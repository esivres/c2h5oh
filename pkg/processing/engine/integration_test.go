package engine

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/behavior"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
	sqlstore "github.com/esivres/c2h5oh/pkg/processing/storage/sql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	// Use file-based temp DB to avoid SQLite in-memory isolation issues
	tmpFile := t.TempDir() + "/test.db"
	db, err := sql.Open("sqlite", tmpFile)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	require.NoError(t, sqlstore.Migrate(context.Background(), db))
	return db
}

func TestIntegration_DeployAndQuery(t *testing.T) {
	db := openTestDB(t)
	store := sqlstore.NewStore(db)
	registry := behavior.DefaultRegistry()

	proc := NewProcessor(Config{
		PartitionId: 1,
		Store:       store,
		Registry:    registry,
		QueueSize:   16,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go proc.Run(ctx)

	content := []byte(`<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL">
		<bpmn:process id="order-process" name="Order Process" isExecutable="true"/>
	</bpmn:definitions>`)
	hash := sha256.Sum256(content)

	proc.Submit(&intent.DeployProcessIntent{
		Header:        intent.Header{Origin: intent.External},
		BpmnProcessId: "order-process",
		Name:          "Order Process",
		Content:       content,
		ContentHash:   hash[:],
	})

	// Wait for processing
	require.Eventually(t, func() bool {
		def, _ := store.ProcessDefinitions().FindLatestByProcessId(context.Background(), "order-process")
		return def != nil
	}, time.Second, 10*time.Millisecond)

	// Verify
	def, err := store.ProcessDefinitions().FindLatestByProcessId(ctx, "order-process")
	require.NoError(t, err)
	require.NotNil(t, def)
	assert.Equal(t, "order-process", def.BpmnProcessId)
	assert.Equal(t, "Order Process", def.Name)
	assert.Equal(t, uint64(1), def.Version)
	assert.Equal(t, uint8(1), PartitionOf(def.Key))
}

func TestIntegration_DeployTwoVersions(t *testing.T) {
	db := openTestDB(t)
	store := sqlstore.NewStore(db)
	registry := behavior.DefaultRegistry()

	proc := NewProcessor(Config{
		PartitionId: 2,
		Store:       store,
		Registry:    registry,
		QueueSize:   16,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go proc.Run(ctx)

	// Deploy v1
	content1 := []byte("<v1/>")
	hash1 := sha256.Sum256(content1)
	proc.Submit(&intent.DeployProcessIntent{
		Header:        intent.Header{Origin: intent.External},
		BpmnProcessId: "proc",
		Name:          "Proc",
		Content:       content1,
		ContentHash:   hash1[:],
	})

	require.Eventually(t, func() bool {
		def, _ := store.ProcessDefinitions().FindLatestByProcessId(context.Background(), "proc")
		return def != nil && def.Version == 1
	}, time.Second, 10*time.Millisecond)

	// Deploy v2
	content2 := []byte("<v2/>")
	hash2 := sha256.Sum256(content2)
	proc.Submit(&intent.DeployProcessIntent{
		Header:        intent.Header{Origin: intent.External},
		BpmnProcessId: "proc",
		Name:          "Proc",
		Content:       content2,
		ContentHash:   hash2[:],
	})

	require.Eventually(t, func() bool {
		def, _ := store.ProcessDefinitions().FindLatestByProcessId(context.Background(), "proc")
		return def != nil && def.Version == 2
	}, time.Second, 10*time.Millisecond)

	// Deploy v1 again — idempotent
	proc.Submit(&intent.DeployProcessIntent{
		Header:        intent.Header{Origin: intent.External},
		BpmnProcessId: "proc",
		Name:          "Proc",
		Content:       content1,
		ContentHash:   hash1[:],
	})

	// Give it time to process
	time.Sleep(50 * time.Millisecond)

	// Still version 2
	def, err := store.ProcessDefinitions().FindLatestByProcessId(ctx, "proc")
	require.NoError(t, err)
	assert.Equal(t, uint64(2), def.Version)

	// Partition prefix is correct
	assert.Equal(t, uint8(2), PartitionOf(def.Key))
}

func TestIntegration_ProcessorAssignsKeys(t *testing.T) {
	db := openTestDB(t)
	store := sqlstore.NewStore(db)

	// Custom behavior that records the key
	var assignedKey uint64
	registry := behavior.NewRegistry()
	registry.Register(intent.DeployProcess, behavior.Typed(
		func(_ context.Context, s storage.Store, i *intent.DeployProcessIntent) ([]intent.Intent, error) {
			assignedKey = i.Key
			hash := sha256.Sum256(i.Content)
			return nil, s.ProcessDefinitions().Create(context.Background(), &storage.ProcessDefinition{
				Key: i.Key, BpmnProcessId: i.BpmnProcessId, Version: 1,
				ContentHash: hash[:], Content: i.Content, DeployedAt: time.Now(),
			})
		},
	))

	proc := NewProcessor(Config{
		PartitionId: 5,
		Store:       store,
		Registry:    registry,
		QueueSize:   16,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go proc.Run(ctx)

	proc.Submit(&intent.DeployProcessIntent{
		Header:        intent.Header{Origin: intent.External},
		BpmnProcessId: "test",
		Content:       []byte("<test/>"),
	})

	require.Eventually(t, func() bool {
		return assignedKey != 0
	}, time.Second, 10*time.Millisecond)

	assert.Equal(t, uint8(5), PartitionOf(assignedKey))
}
