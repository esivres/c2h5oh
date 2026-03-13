package engine

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/esivres/c2h5oh/pkg/bpmn_asserts"
	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/behavior"
	"github.com/esivres/c2h5oh/pkg/processing/export"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	sqlstore "github.com/esivres/c2h5oh/pkg/processing/storage/sql"

	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

// TestE2E_ServiceTask builds a BPMN process with bpmn_model builder,
// deploys it, starts an instance, and verifies the flow using bpmn_asserts.
//
// Process: start → serviceTask("test-worker") → end
func TestE2E_ServiceTask(t *testing.T) {
	// 1. Build BPMN model using builder
	bmi := bpmn_model.CreateExecutableProcess("test-process").
		StartEvent("start").
		ServiceTask("task1").Name("My Service Task").ZeebeJobType("test-worker").
		EndEvent("end").
		Done()

	bpmnXml, err := bmi.GetDocument().WriteToBytes()
	require.NoError(t, err)

	// 2. Setup engine with SQL store + RecordStream exporter
	db := openTestDB(t)
	store := sqlstore.NewStore(db)
	registry := behavior.DefaultRegistry()
	stream := bpmn_asserts.NewRecordStream()
	exporter := export.NewRecordStreamExporter(stream)

	proc := NewProcessor(Config{
		PartitionId: 1,
		Store:       store,
		Registry:    registry,
		Exporter:    exporter,
		QueueSize:   64,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go proc.Run(ctx)

	// 3. Deploy process
	contentHash := sha256.Sum256(bpmnXml)
	proc.Submit(&intent.DeployProcessIntent{
		Header:        intent.Header{Origin: intent.External},
		BpmnProcessId: "test-process",
		Name:          "Test Process",
		Content:       bpmnXml,
		ContentHash:   contentHash[:],
	})

	// Wait for deployment
	require.Eventually(t, func() bool {
		def, _ := store.ProcessDefinitions().FindLatestByProcessId(context.Background(), "test-process")
		return def != nil
	}, 5*time.Second, 10*time.Millisecond)

	// 4. Start process instance
	proc.Submit(&intent.CreateProcessInstanceIntent{
		Header:        intent.Header{Origin: intent.External},
		BpmnProcessId: "test-process",
	})

	// 5. Wait for service task job to be created
	rec, err := bpmn_asserts.Events().
		ValueType(bpmn_asserts.ValueTypeJob).
		Intent(bpmn_asserts.IntentCreated).
		WaitOnCtx(stream, ctx)
	require.NoError(t, err)

	jobValue := bpmn_asserts.MustParseValue[bpmn_asserts.JobValue](rec)
	require.Equal(t, "test-worker", jobValue.Type)

	// 6. Complete the job (simulate worker)
	proc.Submit(&intent.CompleteJobIntent{
		Header: intent.Header{
			Origin:             intent.External,
			ProcessInstanceKey: uint64(jobValue.ProcessInstanceKey),
		},
		JobKey: uint64(rec.Key),
	})

	// 7. Wait for process instance completion
	_, err = bpmn_asserts.Events().
		ValueType(bpmn_asserts.ValueTypeProcessInstance).
		Intent(bpmn_asserts.IntentCompleted).
		WaitOnCtx(stream, ctx)
	require.NoError(t, err)

	// 8. Verify the full flow via RecordStream
	// start event was activated
	startEvents := bpmn_asserts.Events().
		ValueType(bpmn_asserts.ValueTypeProcessInstance).
		Intent(bpmn_asserts.IntentElementActivated).
		ElementID("start").
		FindAll(stream)
	require.GreaterOrEqual(t, len(startEvents), 1, "start event should have been activated")

	// service task was activated
	taskEvents := bpmn_asserts.Events().
		ValueType(bpmn_asserts.ValueTypeProcessInstance).
		Intent(bpmn_asserts.IntentElementActivated).
		ElementID("task1").
		FindAll(stream)
	require.GreaterOrEqual(t, len(taskEvents), 1, "service task should have been activated")

	// end event was activated
	endEvents := bpmn_asserts.Events().
		ValueType(bpmn_asserts.ValueTypeProcessInstance).
		Intent(bpmn_asserts.IntentElementActivated).
		ElementID("end").
		FindAll(stream)
	require.GreaterOrEqual(t, len(endEvents), 1, "end event should have been activated")

	// Print compact log for debugging
	t.Log(stream.PrintCompact())
}

// TestE2E_ServiceTask_ActivateComplete tests the full job lifecycle:
// deploy → start → serviceTask → activate job → complete job → end
func TestE2E_ServiceTask_ActivateComplete(t *testing.T) {
	bmi := bpmn_model.CreateExecutableProcess("test-process").
		StartEvent("start").
		ServiceTask("task1").Name("My Service Task").ZeebeJobType("test-worker").
		EndEvent("end").
		Done()

	bpmnXml, err := bmi.GetDocument().WriteToBytes()
	require.NoError(t, err)

	db := openTestDB(t)
	store := sqlstore.NewStore(db)
	registry := behavior.DefaultRegistry()
	stream := bpmn_asserts.NewRecordStream()
	exporter := export.NewRecordStreamExporter(stream)

	proc := NewProcessor(Config{
		PartitionId: 1,
		Store:       store,
		Registry:    registry,
		Exporter:    exporter,
		QueueSize:   64,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go proc.Run(ctx)

	// Deploy
	contentHash := sha256.Sum256(bpmnXml)
	proc.Submit(&intent.DeployProcessIntent{
		Header:        intent.Header{Origin: intent.External},
		BpmnProcessId: "test-process",
		Name:          "Test Process",
		Content:       bpmnXml,
		ContentHash:   contentHash[:],
	})

	require.Eventually(t, func() bool {
		def, _ := store.ProcessDefinitions().FindLatestByProcessId(context.Background(), "test-process")
		return def != nil
	}, 5*time.Second, 10*time.Millisecond)

	// Start process
	proc.Submit(&intent.CreateProcessInstanceIntent{
		Header:        intent.Header{Origin: intent.External},
		BpmnProcessId: "test-process",
	})

	// Wait for job created
	rec, err := bpmn_asserts.Events().
		ValueType(bpmn_asserts.ValueTypeJob).
		Intent(bpmn_asserts.IntentCreated).
		WaitOnCtx(stream, ctx)
	require.NoError(t, err)

	jobValue := bpmn_asserts.MustParseValue[bpmn_asserts.JobValue](rec)
	require.Equal(t, "test-worker", jobValue.Type)

	// Activate job (simulate worker picking up)
	proc.Submit(&intent.ActivateJobIntent{
		Header: intent.Header{
			Origin:             intent.External,
			ProcessInstanceKey: uint64(jobValue.ProcessInstanceKey),
		},
		JobKey:  uint64(rec.Key),
		Worker:  "my-worker",
		Timeout: 30 * time.Second,
	})

	// Wait for job activated event
	_, err = bpmn_asserts.Events().
		ValueType(bpmn_asserts.ValueTypeJob).
		Intent(bpmn_asserts.IntentActivated).
		WaitOnCtx(stream, ctx)
	require.NoError(t, err)

	// Complete the job
	proc.Submit(&intent.CompleteJobIntent{
		Header: intent.Header{
			Origin:             intent.External,
			ProcessInstanceKey: uint64(jobValue.ProcessInstanceKey),
		},
		JobKey: uint64(rec.Key),
	})

	// Wait for process completion
	_, err = bpmn_asserts.Events().
		ValueType(bpmn_asserts.ValueTypeProcessInstance).
		Intent(bpmn_asserts.IntentCompleted).
		WaitOnCtx(stream, ctx)
	require.NoError(t, err)

	// Verify job went through activated state in DB
	job, err := store.Jobs().GetByKey(context.Background(), uint64(rec.Key))
	require.NoError(t, err)
	require.Equal(t, "my-worker", job.Worker)

	t.Log(stream.PrintCompact())
}

// TestE2E_CancelProcessInstance tests canceling a process while a service task is active.
// deploy → start → serviceTask → cancel → verify termination
func TestE2E_CancelProcessInstance(t *testing.T) {
	bmi := bpmn_model.CreateExecutableProcess("test-process").
		StartEvent("start").
		ServiceTask("task1").Name("My Service Task").ZeebeJobType("test-worker").
		EndEvent("end").
		Done()

	bpmnXml, err := bmi.GetDocument().WriteToBytes()
	require.NoError(t, err)

	db := openTestDB(t)
	store := sqlstore.NewStore(db)
	registry := behavior.DefaultRegistry()
	stream := bpmn_asserts.NewRecordStream()
	exporter := export.NewRecordStreamExporter(stream)

	proc := NewProcessor(Config{
		PartitionId: 1,
		Store:       store,
		Registry:    registry,
		Exporter:    exporter,
		QueueSize:   64,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go proc.Run(ctx)

	// Deploy
	contentHash := sha256.Sum256(bpmnXml)
	proc.Submit(&intent.DeployProcessIntent{
		Header:        intent.Header{Origin: intent.External},
		BpmnProcessId: "test-process",
		Name:          "Test Process",
		Content:       bpmnXml,
		ContentHash:   contentHash[:],
	})

	require.Eventually(t, func() bool {
		def, _ := store.ProcessDefinitions().FindLatestByProcessId(context.Background(), "test-process")
		return def != nil
	}, 5*time.Second, 10*time.Millisecond)

	// Start process
	proc.Submit(&intent.CreateProcessInstanceIntent{
		Header:        intent.Header{Origin: intent.External},
		BpmnProcessId: "test-process",
	})

	// Wait for job created (service task is active)
	rec, err := bpmn_asserts.Events().
		ValueType(bpmn_asserts.ValueTypeJob).
		Intent(bpmn_asserts.IntentCreated).
		WaitOnCtx(stream, ctx)
	require.NoError(t, err)

	jobValue := bpmn_asserts.MustParseValue[bpmn_asserts.JobValue](rec)
	piKey := uint64(jobValue.ProcessInstanceKey)

	// Cancel the process instance
	proc.Submit(&intent.CancelProcessInstanceIntent{
		Header: intent.Header{
			Origin:             intent.External,
			ProcessInstanceKey: piKey,
		},
	})

	// Wait for process termination event
	_, err = bpmn_asserts.Events().
		ValueType(bpmn_asserts.ValueTypeProcessInstance).
		Intent(bpmn_asserts.IntentCanceled).
		WaitOnCtx(stream, ctx)
	require.NoError(t, err)

	// Verify process instance is terminated in DB
	require.Eventually(t, func() bool {
		pi, _ := store.ProcessInstances().GetInstance(context.Background(), piKey)
		return pi != nil && pi.State == 3 // ProcessInstanceTerminated
	}, 2*time.Second, 10*time.Millisecond)

	t.Log(stream.PrintCompact())
}

// TestE2E_ServiceTask_Timeout tests job timeout with retry.
// deploy → start → serviceTask → activate → timeout → re-activate → complete → end
func TestE2E_ServiceTask_Timeout(t *testing.T) {
	bmi := bpmn_model.CreateExecutableProcess("test-process").
		StartEvent("start").
		ServiceTask("task1").Name("My Service Task").ZeebeJobType("test-worker").
		EndEvent("end").
		Done()

	bpmnXml, err := bmi.GetDocument().WriteToBytes()
	require.NoError(t, err)

	db := openTestDB(t)
	store := sqlstore.NewStore(db)
	registry := behavior.DefaultRegistry()
	stream := bpmn_asserts.NewRecordStream()
	exporter := export.NewRecordStreamExporter(stream)

	proc := NewProcessor(Config{
		PartitionId: 1,
		Store:       store,
		Registry:    registry,
		Exporter:    exporter,
		QueueSize:   64,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go proc.Run(ctx)

	// Deploy
	contentHash := sha256.Sum256(bpmnXml)
	proc.Submit(&intent.DeployProcessIntent{
		Header:        intent.Header{Origin: intent.External},
		BpmnProcessId: "test-process",
		Name:          "Test Process",
		Content:       bpmnXml,
		ContentHash:   contentHash[:],
	})

	require.Eventually(t, func() bool {
		def, _ := store.ProcessDefinitions().FindLatestByProcessId(context.Background(), "test-process")
		return def != nil
	}, 5*time.Second, 10*time.Millisecond)

	// Start process
	proc.Submit(&intent.CreateProcessInstanceIntent{
		Header:        intent.Header{Origin: intent.External},
		BpmnProcessId: "test-process",
	})

	// Wait for job created
	rec, err := bpmn_asserts.Events().
		ValueType(bpmn_asserts.ValueTypeJob).
		Intent(bpmn_asserts.IntentCreated).
		WaitOnCtx(stream, ctx)
	require.NoError(t, err)

	jobValue := bpmn_asserts.MustParseValue[bpmn_asserts.JobValue](rec)
	piKey := uint64(jobValue.ProcessInstanceKey)
	jobKey := uint64(rec.Key)

	// Activate job
	proc.Submit(&intent.ActivateJobIntent{
		Header:  intent.Header{Origin: intent.External, ProcessInstanceKey: piKey},
		JobKey:  jobKey,
		Worker:  "worker-1",
		Timeout: time.Millisecond, // very short timeout
	})

	// Wait for activation
	_, err = bpmn_asserts.Events().
		ValueType(bpmn_asserts.ValueTypeJob).
		Intent(bpmn_asserts.IntentActivated).
		WaitOnCtx(stream, ctx)
	require.NoError(t, err)

	// Simulate timeout
	proc.Submit(&intent.TimeOutJobIntent{
		Header: intent.Header{Origin: intent.Internal, ProcessInstanceKey: piKey},
		JobKey: jobKey,
	})

	// Wait for timeout event
	_, err = bpmn_asserts.Events().
		ValueType(bpmn_asserts.ValueTypeJob).
		Intent(bpmn_asserts.IntentTimedOut).
		WaitOnCtx(stream, ctx)
	require.NoError(t, err)

	// Job should have retries decremented (3 → 2), state=Failed
	job, err := store.Jobs().GetByKey(context.Background(), jobKey)
	require.NoError(t, err)
	require.Equal(t, 2, job.Retries)

	// Now complete the job directly (simulate worker completing after fail)
	// First need to re-complete via CompleteJob
	proc.Submit(&intent.CompleteJobIntent{
		Header: intent.Header{Origin: intent.External, ProcessInstanceKey: piKey},
		JobKey: jobKey,
	})

	// Wait for process completion
	_, err = bpmn_asserts.Events().
		ValueType(bpmn_asserts.ValueTypeProcessInstance).
		Intent(bpmn_asserts.IntentCompleted).
		WaitOnCtx(stream, ctx)
	require.NoError(t, err)

	t.Log(stream.PrintCompact())
}
