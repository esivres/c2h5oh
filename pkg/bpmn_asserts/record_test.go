package bpmn_asserts

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- RecordStream ---

func TestRecordStream_Add_And_Filter(t *testing.T) {
	stream := NewRecordStream()

	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "start", "START_EVENT"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 2, 100, "start", "START_EVENT"),
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 3, 100, "task1", "SERVICE_TASK"),
		makeJobRecord("CREATED", 10, "myJob", "task1", 100, 3),
	)

	all := stream.Filter(func(r Record) bool { return true })
	assert.Len(t, all, 4)

	piOnly := stream.Filter(func(r Record) bool { return r.ValueType == ValueTypeProcessInstance })
	assert.Len(t, piOnly, 3)

	jobOnly := stream.Filter(func(r Record) bool { return r.ValueType == ValueTypeJob })
	assert.Len(t, jobOnly, 1)
	assert.Equal(t, IntentCreated, jobOnly[0].Intent)
}

func TestRecordStream_WaitFor_ExistingRecord(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 1, 100, "start", "START_EVENT"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	rec, err := stream.WaitFor(ctx, func(r Record) bool {
		return r.Intent == IntentElementCompleted
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), rec.Key)
}

func TestRecordStream_WaitFor_FutureRecord(t *testing.T) {
	stream := NewRecordStream()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan Record, 1)
	go func() {
		rec, err := stream.WaitFor(ctx, func(r Record) bool {
			return r.Intent == IntentElementCompleted
		})
		assert.NoError(t, err)
		done <- rec
	}()

	// Add record after short delay
	time.Sleep(50 * time.Millisecond)
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 5, 100, "end", "END_EVENT"),
	)

	select {
	case rec := <-done:
		assert.Equal(t, int64(5), rec.Key)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for WaitFor to return")
	}
}

func TestRecordStream_WaitFor_Timeout(t *testing.T) {
	stream := NewRecordStream()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := stream.WaitFor(ctx, func(r Record) bool {
		return r.Intent == "NEVER_HAPPENS"
	})
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestRecordStream_WaitFor_CleansUpWaiter(t *testing.T) {
	stream := NewRecordStream()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, _ = stream.WaitFor(ctx, func(r Record) bool { return false })

	stream.mu.Lock()
	assert.Empty(t, stream.waiters, "waiters should be cleaned up after timeout")
	stream.mu.Unlock()
}

func TestRecordStream_Reset(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "start", "START_EVENT"),
	)

	assert.Len(t, stream.Filter(func(r Record) bool { return true }), 1)

	stream.Reset()
	assert.Empty(t, stream.Filter(func(r Record) bool { return true }))
}

func TestRecordStream_Add_InvalidJSON(t *testing.T) {
	stream := NewRecordStream()
	stream.Add(json.RawMessage(`not valid json`))
	assert.Empty(t, stream.Filter(func(r Record) bool { return true }))
}

// --- ParseValue ---

func TestParseValue(t *testing.T) {
	rec := makeJobRecord("CREATED", 10, "myJob", "task1", 100, 3)
	v, err := ParseValue[JobValue](rec)
	require.NoError(t, err)
	assert.Equal(t, "myJob", v.Type)
	assert.Equal(t, "task1", v.ElementID)
	assert.Equal(t, int64(100), v.ProcessInstanceKey)
	assert.Equal(t, int32(3), v.Retries)
}

func TestParseValue_ProcessInstance(t *testing.T) {
	rec := makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "start", "START_EVENT")
	v, err := ParseValue[ProcessInstanceValue](rec)
	require.NoError(t, err)
	assert.Equal(t, int64(100), v.ProcessInstanceKey)
	assert.Equal(t, "start", v.ElementID)
	assert.Equal(t, "START_EVENT", v.BpmnElementType)
}

func TestTryParseValue(t *testing.T) {
	rec := makeVariableRecord("CREATED", 1, 100, "myVar", `"hello"`)
	v, ok := TryParseValue[VariableValue](rec)
	assert.True(t, ok)
	assert.Equal(t, "myVar", v.Name)
	assert.Equal(t, `"hello"`, v.Value)
}

func TestTryParseValue_WrongType(t *testing.T) {
	rec := makeRecord("JOB", "EVENT", "CREATED", 1, map[string]string{"bad": "data"})
	v, ok := TryParseValue[IncidentValue](rec)
	// Should succeed since JSON is flexible, but fields won't match
	assert.True(t, ok)
	assert.Empty(t, v.ErrorType)
}

func TestMustParseValue_Success(t *testing.T) {
	rec := makeIncidentRecord("CREATED", 1, 100, "JOB_NO_RETRIES", "no retries", "task1", 50)
	v := MustParseValue[IncidentValue](rec)
	assert.Equal(t, "JOB_NO_RETRIES", v.ErrorType)
	assert.Equal(t, "no retries", v.ErrorMessage)
}

func TestMustParseValue_Panic(t *testing.T) {
	rec := Record{Value: json.RawMessage(`{invalid`)}
	assert.Panics(t, func() {
		MustParseValue[IncidentValue](rec)
	})
}
