package bpmn_asserts

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

// ProcessDefinitionAssert provides fluent assertions on a deployed process definition.
// This is the Go equivalent of the Java ProcessAssert.
type ProcessDefinitionAssert struct {
	t                    testing.TB
	processDefinitionKey int64
	stream               *RecordStream
	ctx                  context.Context
}

// ForProcessDefinition creates a new ProcessDefinitionAssert for the given process definition key.
func ForProcessDefinition(t testing.TB, stream *RecordStream, processDefinitionKey int64) *ProcessDefinitionAssert {
	t.Helper()
	return &ProcessDefinitionAssert{
		t:                    t,
		processDefinitionKey: processDefinitionKey,
		stream:               stream,
		ctx:                  context.Background(),
	}
}

// WithContext sets the timeout context for waiting assertions.
func (a *ProcessDefinitionAssert) WithContext(ctx context.Context) *ProcessDefinitionAssert {
	a.ctx = ctx
	return a
}

// HasBpmnProcessID asserts that the process definition has the specified BPMN process ID.
func (a *ProcessDefinitionAssert) HasBpmnProcessID(expectedBpmnProcessID string) *ProcessDefinitionAssert {
	a.t.Helper()
	v := a.getProcessValue()
	if v.BpmnProcessID != expectedBpmnProcessID {
		a.t.Fatalf("expected process definition %d to have BPMN process ID %q, but was %q",
			a.processDefinitionKey, expectedBpmnProcessID, v.BpmnProcessID)
	}
	return a
}

// HasVersion asserts that the process definition has the specified version.
func (a *ProcessDefinitionAssert) HasVersion(expectedVersion int32) *ProcessDefinitionAssert {
	a.t.Helper()
	v := a.getProcessValue()
	if v.Version != expectedVersion {
		a.t.Fatalf("expected process definition %d to have version %d, but was %d",
			a.processDefinitionKey, expectedVersion, v.Version)
	}
	return a
}

// HasResourceName asserts that the process definition has the specified resource name.
func (a *ProcessDefinitionAssert) HasResourceName(expectedResourceName string) *ProcessDefinitionAssert {
	a.t.Helper()
	v := a.getProcessValue()
	if v.ResourceName != expectedResourceName {
		a.t.Fatalf("expected process definition %d to have resource name %q, but was %q",
			a.processDefinitionKey, expectedResourceName, v.ResourceName)
	}
	return a
}

// HasAnyInstances asserts that at least one process instance has been started from this definition.
func (a *ProcessDefinitionAssert) HasAnyInstances() *ProcessDefinitionAssert {
	a.t.Helper()
	count := a.instanceCount()
	if count == 0 {
		a.t.Fatalf("expected process definition %d to have instances, but found none",
			a.processDefinitionKey)
	}
	return a
}

// HasNoInstances asserts that no process instances have been started from this definition (point-in-time).
func (a *ProcessDefinitionAssert) HasNoInstances() *ProcessDefinitionAssert {
	a.t.Helper()
	count := a.instanceCount()
	if count > 0 {
		a.t.Fatalf("expected process definition %d to have no instances, but found %d",
			a.processDefinitionKey, count)
	}
	return a
}

// HasInstances asserts that exactly the specified number of process instances have been started.
func (a *ProcessDefinitionAssert) HasInstances(expectedCount int) *ProcessDefinitionAssert {
	a.t.Helper()
	count := a.instanceCount()
	if count != expectedCount {
		a.t.Fatalf("expected process definition %d to have %d instance(s), but found %d",
			a.processDefinitionKey, expectedCount, count)
	}
	return a
}

func (a *ProcessDefinitionAssert) getProcessValue() ProcessDefinitionValue {
	a.t.Helper()
	rec, err := a.stream.WaitFor(a.ctx, func(r Record) bool {
		return r.ValueType == ValueTypeProcess && r.Intent == IntentCreated && r.Key == a.processDefinitionKey
	})
	if err != nil {
		a.t.Fatalf("expected to find CREATED record for process definition %d, but timed out waiting",
			a.processDefinitionKey)
	}
	var v ProcessDefinitionValue
	if err := json.Unmarshal(rec.Value, &v); err != nil {
		a.t.Fatalf("failed to unmarshal process definition value: %v", err)
	}
	return v
}

func (a *ProcessDefinitionAssert) instanceCount() int {
	instanceKeys := make(map[int64]struct{})

	records := a.stream.Filter(func(r Record) bool {
		if r.ValueType != ValueTypeProcessInstance || r.RecordType != RecordTypeEvent {
			return false
		}
		if r.Intent != IntentElementActivated {
			return false
		}
		var v ProcessInstanceValue
		if json.Unmarshal(r.Value, &v) != nil {
			return false
		}
		return v.ProcessDefinitionKey == a.processDefinitionKey && v.BpmnElementType == "PROCESS"
	})

	for _, r := range records {
		var v ProcessInstanceValue
		if json.Unmarshal(r.Value, &v) == nil {
			instanceKeys[v.ProcessInstanceKey] = struct{}{}
		}
	}
	return len(instanceKeys)
}

// String returns a debug description.
func (a *ProcessDefinitionAssert) String() string {
	return fmt.Sprintf("ProcessDefinitionAssert{key=%d}", a.processDefinitionKey)
}
