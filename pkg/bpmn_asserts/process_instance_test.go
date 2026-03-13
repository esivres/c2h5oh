package bpmn_asserts

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 100*time.Millisecond)
}

// --- Lifecycle ---

func TestProcessInstanceAssert_IsStarted(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "testProcess", "PROCESS"),
	)

	ctx, cancel := newTestContext()
	defer cancel()

	ForProcessInstance(t, stream, 100).WithContext(ctx).IsStarted()
}

func TestProcessInstanceAssert_IsStarted_WrongKey(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 200, "testProcess", "PROCESS"),
	)

	fc := &fatalCatcher{TB: t}
	ctx, cancel := newTestContext()
	defer cancel()

	catchFatal(func() {
		ForProcessInstance(fc, stream, 100).WithContext(ctx).IsStarted()
	})
	assert.True(t, fc.fataled, "should fail when process instance key doesn't match")
}

func TestProcessInstanceAssert_IsCompleted(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "testProcess", "PROCESS"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 2, 100, "testProcess", "PROCESS"),
	)

	ctx, cancel := newTestContext()
	defer cancel()

	ForProcessInstance(t, stream, 100).WithContext(ctx).IsCompleted()
}

func TestProcessInstanceAssert_IsNotCompleted(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "testProcess", "PROCESS"),
	)

	ForProcessInstance(t, stream, 100).IsNotCompleted()
}

func TestProcessInstanceAssert_IsNotCompleted_Fails(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "testProcess", "PROCESS"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 2, 100, "testProcess", "PROCESS"),
	)

	fc := &fatalCatcher{TB: t}
	catchFatal(func() {
		ForProcessInstance(fc, stream, 100).IsNotCompleted()
	})
	assert.True(t, fc.fataled)
}

func TestProcessInstanceAssert_IsTerminated(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "testProcess", "PROCESS"),
		makeProcessInstanceRecord("ELEMENT_TERMINATED", 2, 100, "testProcess", "PROCESS"),
	)

	ctx, cancel := newTestContext()
	defer cancel()

	ForProcessInstance(t, stream, 100).WithContext(ctx).IsTerminated()
}

func TestProcessInstanceAssert_IsNotTerminated(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "testProcess", "PROCESS"),
	)

	ForProcessInstance(t, stream, 100).IsNotTerminated()
}

func TestProcessInstanceAssert_IsActive(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "testProcess", "PROCESS"),
	)

	ctx, cancel := newTestContext()
	defer cancel()

	ForProcessInstance(t, stream, 100).WithContext(ctx).IsActive()
}

func TestProcessInstanceAssert_IsActive_FailsWhenCompleted(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "testProcess", "PROCESS"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 2, 100, "testProcess", "PROCESS"),
	)

	fc := &fatalCatcher{TB: t}
	ctx, cancel := newTestContext()
	defer cancel()

	catchFatal(func() {
		ForProcessInstance(fc, stream, 100).WithContext(ctx).IsActive()
	})
	assert.True(t, fc.fataled)
	assert.Contains(t, fc.fatalMsg, "already completed")
}

// --- Element assertions ---

func TestProcessInstanceAssert_HasPassedElement(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "start", "START_EVENT"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 2, 100, "start", "START_EVENT"),
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 3, 100, "task1", "SERVICE_TASK"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 4, 100, "task1", "SERVICE_TASK"),
	)

	ctx, cancel := newTestContext()
	defer cancel()

	ForProcessInstance(t, stream, 100).WithContext(ctx).
		HasPassedElement("start").
		HasPassedElement("task1")
}

func TestProcessInstanceAssert_HasPassedElement_Fails(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "start", "START_EVENT"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 2, 100, "start", "START_EVENT"),
	)

	fc := &fatalCatcher{TB: t}
	ctx, cancel := newTestContext()
	defer cancel()

	catchFatal(func() {
		ForProcessInstance(fc, stream, 100).WithContext(ctx).HasPassedElement("task1")
	})
	assert.True(t, fc.fataled)
}

func TestProcessInstanceAssert_HasPassedElementTimes(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 1, 100, "loopTask", "SERVICE_TASK"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 2, 100, "loopTask", "SERVICE_TASK"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 3, 100, "loopTask", "SERVICE_TASK"),
	)

	ForProcessInstance(t, stream, 100).HasPassedElementTimes("loopTask", 3)
}

func TestProcessInstanceAssert_HasPassedElementTimes_Fails(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 1, 100, "loopTask", "SERVICE_TASK"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 2, 100, "loopTask", "SERVICE_TASK"),
	)

	fc := &fatalCatcher{TB: t}
	catchFatal(func() {
		ForProcessInstance(fc, stream, 100).HasPassedElementTimes("loopTask", 3)
	})
	assert.True(t, fc.fataled)
}

func TestProcessInstanceAssert_HasNotPassedElement(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 1, 100, "start", "START_EVENT"),
	)

	ForProcessInstance(t, stream, 100).HasNotPassedElement("task1")
}

func TestProcessInstanceAssert_HasNotPassedElement_Fails(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 1, 100, "task1", "SERVICE_TASK"),
	)

	fc := &fatalCatcher{TB: t}
	catchFatal(func() {
		ForProcessInstance(fc, stream, 100).HasNotPassedElement("task1")
	})
	assert.True(t, fc.fataled)
}

func TestProcessInstanceAssert_HasPassedElementsInOrder(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 1, 100, "start", "START_EVENT"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 2, 100, "task1", "SERVICE_TASK"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 3, 100, "task2", "SERVICE_TASK"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 4, 100, "end", "END_EVENT"),
	)

	ctx, cancel := newTestContext()
	defer cancel()

	ForProcessInstance(t, stream, 100).WithContext(ctx).
		HasPassedElementsInOrder("start", "task1", "task2", "end")
}

func TestProcessInstanceAssert_HasPassedElementsInOrder_Subsequence(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 1, 100, "start", "START_EVENT"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 2, 100, "task1", "SERVICE_TASK"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 3, 100, "task2", "SERVICE_TASK"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 4, 100, "end", "END_EVENT"),
	)

	ctx, cancel := newTestContext()
	defer cancel()

	// Should work with a subsequence
	ForProcessInstance(t, stream, 100).WithContext(ctx).
		HasPassedElementsInOrder("start", "end")
}

func TestProcessInstanceAssert_HasPassedElementsInOrder_WrongOrder(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 1, 100, "start", "START_EVENT"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 2, 100, "task1", "SERVICE_TASK"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 3, 100, "end", "END_EVENT"),
	)

	fc := &fatalCatcher{TB: t}
	ctx, cancel := newTestContext()
	defer cancel()

	catchFatal(func() {
		ForProcessInstance(fc, stream, 100).WithContext(ctx).
			HasPassedElementsInOrder("end", "start")
	})
	assert.True(t, fc.fataled)
}

// --- Wait state assertions ---

func TestProcessInstanceAssert_IsWaitingAtElements(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "task1", "SERVICE_TASK"),
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 2, 100, "task2", "SERVICE_TASK"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 1, 100, "task1", "SERVICE_TASK"),
	)

	ForProcessInstance(t, stream, 100).IsWaitingAtElements("task2")
}

func TestProcessInstanceAssert_IsWaitingAtElements_Fails(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "task1", "SERVICE_TASK"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 1, 100, "task1", "SERVICE_TASK"),
	)

	fc := &fatalCatcher{TB: t}
	catchFatal(func() {
		ForProcessInstance(fc, stream, 100).IsWaitingAtElements("task1")
	})
	assert.True(t, fc.fataled)
}

func TestProcessInstanceAssert_IsNotWaitingAtElements(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "task1", "SERVICE_TASK"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 1, 100, "task1", "SERVICE_TASK"),
	)

	ForProcessInstance(t, stream, 100).IsNotWaitingAtElements("task1")
}

func TestProcessInstanceAssert_IsWaitingExactlyAtElements(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "task1", "SERVICE_TASK"),
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 2, 100, "task2", "SERVICE_TASK"),
	)

	ForProcessInstance(t, stream, 100).IsWaitingExactlyAtElements("task1", "task2")
}

func TestProcessInstanceAssert_IsWaitingExactlyAtElements_Extra(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "task1", "SERVICE_TASK"),
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 2, 100, "task2", "SERVICE_TASK"),
	)

	fc := &fatalCatcher{TB: t}
	catchFatal(func() {
		ForProcessInstance(fc, stream, 100).IsWaitingExactlyAtElements("task1")
	})
	assert.True(t, fc.fataled)
	assert.Contains(t, fc.fatalMsg, "task2")
}

// --- Variable assertions ---

func TestProcessInstanceAssert_HasVariable(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeVariableRecord("CREATED", 1, 100, "result", `"hello"`),
	)

	ctx, cancel := newTestContext()
	defer cancel()

	ForProcessInstance(t, stream, 100).WithContext(ctx).HasVariable("result")
}

func TestProcessInstanceAssert_HasVariableWithValue(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeVariableRecord("CREATED", 1, 100, "count", "42"),
	)

	ForProcessInstance(t, stream, 100).HasVariableWithValue("count", 42)
}

func TestProcessInstanceAssert_HasVariableWithValue_String(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeVariableRecord("CREATED", 1, 100, "name", `"hello"`),
	)

	ForProcessInstance(t, stream, 100).HasVariableWithValue("name", "hello")
}

func TestProcessInstanceAssert_HasVariableWithValue_WrongValue(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeVariableRecord("CREATED", 1, 100, "count", "42"),
	)

	fc := &fatalCatcher{TB: t}
	catchFatal(func() {
		ForProcessInstance(fc, stream, 100).HasVariableWithValue("count", 99)
	})
	assert.True(t, fc.fataled)
}

func TestProcessInstanceAssert_HasVariableWithValue_MissingVar(t *testing.T) {
	stream := NewRecordStream()

	fc := &fatalCatcher{TB: t}
	catchFatal(func() {
		ForProcessInstance(fc, stream, 100).HasVariableWithValue("missing", "x")
	})
	assert.True(t, fc.fataled)
	assert.Contains(t, fc.fatalMsg, "does not exist")
}

func TestProcessInstanceAssert_HasVariableWithValue_Updated(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeVariableRecord("CREATED", 1, 100, "count", "1"),
		makeVariableRecord("UPDATED", 2, 100, "count", "2"),
	)

	// Should use the latest value
	ForProcessInstance(t, stream, 100).HasVariableWithValue("count", 2)
}

func TestProcessInstanceAssert_ExtractVariables(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeVariableRecord("CREATED", 1, 100, "a", `"hello"`),
		makeVariableRecord("CREATED", 2, 100, "b", "42"),
	)

	vars := ForProcessInstance(t, stream, 100).ExtractVariables()
	require.Len(t, vars, 2)
	assert.Equal(t, `"hello"`, vars["a"])
	assert.Equal(t, "42", vars["b"])
}

func TestProcessInstanceAssert_ExtractingVariablesAssert(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeVariableRecord("CREATED", 1, 100, "a", `"hello"`),
		makeVariableRecord("CREATED", 2, 100, "b", "42"),
	)

	ForProcessInstance(t, stream, 100).ExtractingVariablesAssert().
		ContainsVariable("a").
		ContainsVariable("b").
		DoesNotContainVariable("c").
		HasSize(2)
}

// --- Incident assertions ---

func TestProcessInstanceAssert_HasAnyIncidents(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeIncidentRecord("CREATED", 1, 100, "JOB_NO_RETRIES", "no retries", "task1", 50),
	)

	ctx, cancel := newTestContext()
	defer cancel()

	ForProcessInstance(t, stream, 100).WithContext(ctx).HasAnyIncidents()
}

func TestProcessInstanceAssert_HasNoIncidents(t *testing.T) {
	stream := NewRecordStream()

	ForProcessInstance(t, stream, 100).HasNoIncidents()
}

func TestProcessInstanceAssert_HasNoIncidents_Fails(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeIncidentRecord("CREATED", 1, 100, "JOB_NO_RETRIES", "no retries", "task1", 50),
	)

	fc := &fatalCatcher{TB: t}
	catchFatal(func() {
		ForProcessInstance(fc, stream, 100).HasNoIncidents()
	})
	assert.True(t, fc.fataled)
}

func TestProcessInstanceAssert_ExtractLatestIncident(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeIncidentRecord("CREATED", 1, 100, "JOB_NO_RETRIES", "first", "task1", 50),
		makeIncidentRecord("CREATED", 2, 100, "CONDITION_ERROR", "second", "task2", 60),
	)

	v := ForProcessInstance(t, stream, 100).ExtractLatestIncident()
	assert.Equal(t, "CONDITION_ERROR", v.ErrorType)
	assert.Equal(t, "second", v.ErrorMessage)
}

func TestProcessInstanceAssert_ExtractingLatestIncident(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeIncidentRecord("CREATED", 1, 100, "JOB_NO_RETRIES", "error msg", "task1", 50),
	)

	ctx, cancel := newTestContext()
	defer cancel()

	ForProcessInstance(t, stream, 100).WithContext(ctx).
		ExtractingLatestIncident().
		HasErrorType("JOB_NO_RETRIES").
		HasErrorMessage("error msg").
		OccurredOnElement("task1")
}

// --- Message assertions ---

func TestProcessInstanceAssert_IsWaitingForMessages(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeMessageSubscriptionRecord("CREATED", 1, 100, 50, "paymentReceived"),
	)

	ForProcessInstance(t, stream, 100).IsWaitingForMessages("paymentReceived")
}

func TestProcessInstanceAssert_IsWaitingForMessages_Correlated(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeMessageSubscriptionRecord("CREATED", 1, 100, 50, "paymentReceived"),
		makeMessageSubscriptionRecord("CORRELATED", 2, 100, 50, "paymentReceived"),
	)

	// Once correlated, should NOT be waiting
	fc := &fatalCatcher{TB: t}
	catchFatal(func() {
		ForProcessInstance(fc, stream, 100).IsWaitingForMessages("paymentReceived")
	})
	assert.True(t, fc.fataled)
}

func TestProcessInstanceAssert_IsNotWaitingForMessages(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeMessageSubscriptionRecord("CREATED", 1, 100, 50, "paymentReceived"),
		makeMessageSubscriptionRecord("CORRELATED", 2, 100, 50, "paymentReceived"),
	)

	ForProcessInstance(t, stream, 100).IsNotWaitingForMessages("paymentReceived")
}

func TestProcessInstanceAssert_HasCorrelatedMessageByName(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeMessageSubscriptionRecord("CORRELATED", 1, 100, 50, "paymentReceived"),
	)

	ForProcessInstance(t, stream, 100).HasCorrelatedMessageByName("paymentReceived", 1)
}

func TestProcessInstanceAssert_HasCorrelatedMessageByCorrelationKey(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeMessageSubscriptionRecord("CORRELATED", 1, 100, 50, "paymentReceived"),
	)

	ForProcessInstance(t, stream, 100).HasCorrelatedMessageByCorrelationKey("corr-123", 1)
}

// --- Called process assertions ---

func TestProcessInstanceAssert_HasCalledProcess(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeChildProcessRecord("ELEMENT_ACTIVATED", 1, 200, 100, "childProcess"),
	)

	ForProcessInstance(t, stream, 100).HasCalledProcess()
}

func TestProcessInstanceAssert_HasCalledProcess_ByID(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeChildProcessRecord("ELEMENT_ACTIVATED", 1, 200, 100, "childProcess"),
	)

	ForProcessInstance(t, stream, 100).HasCalledProcess("childProcess")
}

func TestProcessInstanceAssert_HasNotCalledProcess(t *testing.T) {
	stream := NewRecordStream()

	ForProcessInstance(t, stream, 100).HasNotCalledProcess()
}

func TestProcessInstanceAssert_ExtractingLatestCalledProcess(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeChildProcessRecord("ELEMENT_ACTIVATED", 1, 200, 100, "childProcess"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 2, 200, "childProcess", "PROCESS"),
	)

	ctx, cancel := newTestContext()
	defer cancel()

	ForProcessInstance(t, stream, 100).WithContext(ctx).
		ExtractingLatestCalledProcess("childProcess").
		IsCompleted()
}

// --- Chaining ---

func TestProcessInstanceAssert_FullChain(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "testProcess", "PROCESS"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 2, 100, "start", "START_EVENT"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 3, 100, "task1", "SERVICE_TASK"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 4, 100, "end", "END_EVENT"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 5, 100, "testProcess", "PROCESS"),
		makeVariableRecord("CREATED", 6, 100, "result", `"ok"`),
	)

	ctx, cancel := newTestContext()
	defer cancel()

	ForProcessInstance(t, stream, 100).WithContext(ctx).
		IsStarted().
		IsCompleted().
		HasPassedElement("task1").
		HasPassedElementsInOrder("start", "task1", "end").
		HasNotPassedElement("unknownTask").
		HasNoIncidents().
		HasVariableWithValue("result", "ok")
}
