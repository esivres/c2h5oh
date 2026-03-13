package bpmn_asserts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// ProcessInstanceAssert provides fluent assertions on a Zeebe process instance.
// Positive assertions (IsCompleted, HasPassedElement, etc.) block until the condition
// is met or the context times out. Negative assertions (HasNotPassedElement, IsNotCompleted, etc.)
// check point-in-time — call them after a positive assertion that establishes a known state.
type ProcessInstanceAssert struct {
	t      testing.TB
	ctx    context.Context
	stream *RecordStream
	key    int64
}

// ForProcessInstance creates a new ProcessInstanceAssert for the given process instance key.
func ForProcessInstance(t testing.TB, stream *RecordStream, key int64) *ProcessInstanceAssert {
	t.Helper()
	return &ProcessInstanceAssert{
		t:      t,
		key:    key,
		stream: stream,
		ctx:    context.Background(),
	}
}

// WithContext sets the timeout context for waiting assertions.
func (a *ProcessInstanceAssert) WithContext(ctx context.Context) *ProcessInstanceAssert {
	a.ctx = ctx
	return a
}

// --- Lifecycle assertions ---

// IsStarted asserts that the process instance has been activated.
func (a *ProcessInstanceAssert) IsStarted() *ProcessInstanceAssert {
	a.t.Helper()
	a.waitForProcessIntent(IntentElementActivated, "expected process instance %d to be started")
	return a
}

// IsActive asserts that the process instance has been activated but not yet completed or terminated.
func (a *ProcessInstanceAssert) IsActive() *ProcessInstanceAssert {
	a.t.Helper()
	a.waitForProcessIntent(IntentElementActivated, "expected process instance %d to be active (started)")

	if a.hasProcessIntent(IntentElementCompleted) {
		a.t.Fatalf("expected process instance %d to be active, but it is already completed", a.key)
	}
	if a.hasProcessIntent(IntentElementTerminated) {
		a.t.Fatalf("expected process instance %d to be active, but it is already terminated", a.key)
	}
	return a
}

// IsCompleted asserts that the process instance has completed.
func (a *ProcessInstanceAssert) IsCompleted() *ProcessInstanceAssert {
	a.t.Helper()
	a.waitForProcessIntent(IntentElementCompleted, "expected process instance %d to be completed")
	return a
}

// IsNotCompleted asserts that the process instance has NOT completed (point-in-time).
func (a *ProcessInstanceAssert) IsNotCompleted() *ProcessInstanceAssert {
	a.t.Helper()
	if a.hasProcessIntent(IntentElementCompleted) {
		a.t.Fatalf("expected process instance %d NOT to be completed, but it is", a.key)
	}
	return a
}

// IsTerminated asserts that the process instance has been terminated.
func (a *ProcessInstanceAssert) IsTerminated() *ProcessInstanceAssert {
	a.t.Helper()
	a.waitForProcessIntent(IntentElementTerminated, "expected process instance %d to be terminated")
	return a
}

// IsNotTerminated asserts that the process instance has NOT been terminated (point-in-time).
func (a *ProcessInstanceAssert) IsNotTerminated() *ProcessInstanceAssert {
	a.t.Helper()
	if a.hasProcessIntent(IntentElementTerminated) {
		a.t.Fatalf("expected process instance %d NOT to be terminated, but it is", a.key)
	}
	return a
}

// --- Element assertions ---

// HasPassedElement asserts that the element with the given ID has been completed at least once.
func (a *ProcessInstanceAssert) HasPassedElement(elementID string) *ProcessInstanceAssert {
	a.t.Helper()
	_, err := a.stream.WaitFor(a.ctx, func(r Record) bool {
		return a.isProcessInstanceRecord(r) &&
			r.Intent == IntentElementCompleted &&
			a.valueElementID(r) == elementID
	})
	if err != nil {
		a.t.Fatalf("expected process instance %d to have passed element %q, but timed out waiting", a.key, elementID)
	}
	return a
}

// HasPassedElementTimes asserts that the element has been completed exactly n times (point-in-time).
func (a *ProcessInstanceAssert) HasPassedElementTimes(elementID string, times int) *ProcessInstanceAssert {
	a.t.Helper()
	count := a.countCompletedElement(elementID)
	if count != times {
		a.t.Fatalf("expected process instance %d to have passed element %q %d time(s), but was %d",
			a.key, elementID, times, count)
	}
	return a
}

// HasNotPassedElement asserts that the element has NOT been completed (point-in-time).
func (a *ProcessInstanceAssert) HasNotPassedElement(elementID string) *ProcessInstanceAssert {
	a.t.Helper()
	count := a.countCompletedElement(elementID)
	if count > 0 {
		a.t.Fatalf("expected process instance %d NOT to have passed element %q, but it was passed %d time(s)",
			a.key, elementID, count)
	}
	return a
}

// HasPassedElementsInOrder asserts that the given elements were completed in the specified order.
// Waits for the last element, then verifies order among all completed elements.
func (a *ProcessInstanceAssert) HasPassedElementsInOrder(elementIDs ...string) *ProcessInstanceAssert {
	a.t.Helper()
	if len(elementIDs) == 0 {
		return a
	}

	a.HasPassedElement(elementIDs[len(elementIDs)-1])

	completedIDs := a.completedElementIDs()
	idx := 0
	for _, completed := range completedIDs {
		if idx < len(elementIDs) && completed == elementIDs[idx] {
			idx++
		}
	}
	if idx < len(elementIDs) {
		a.t.Fatalf("expected process instance %d to have passed elements in order %v, "+
			"but only matched %d of %d in completed sequence %v",
			a.key, elementIDs, idx, len(elementIDs), completedIDs)
	}
	return a
}

// IsWaitingAtElements asserts that the process instance is currently waiting at ALL specified elements.
// An element is "waiting" if it has been activated but not yet completed or terminated.
func (a *ProcessInstanceAssert) IsWaitingAtElements(elementIDs ...string) *ProcessInstanceAssert {
	a.t.Helper()
	waiting := a.elementsInWaitState()
	for _, id := range elementIDs {
		if _, ok := waiting[id]; !ok {
			a.t.Fatalf("expected process instance %d to be waiting at element %q, "+
				"but current wait state is %v", a.key, id, setKeys(waiting))
		}
	}
	return a
}

// IsNotWaitingAtElements asserts that the process instance is NOT waiting at any of the specified elements.
func (a *ProcessInstanceAssert) IsNotWaitingAtElements(elementIDs ...string) *ProcessInstanceAssert {
	a.t.Helper()
	waiting := a.elementsInWaitState()
	for _, id := range elementIDs {
		if _, ok := waiting[id]; ok {
			a.t.Fatalf("expected process instance %d NOT to be waiting at element %q, but it is",
				a.key, id)
		}
	}
	return a
}

// IsWaitingExactlyAtElements asserts that the process instance is waiting at exactly the specified elements.
func (a *ProcessInstanceAssert) IsWaitingExactlyAtElements(elementIDs ...string) *ProcessInstanceAssert {
	a.t.Helper()
	waiting := a.elementsInWaitState()
	expected := make(map[string]struct{}, len(elementIDs))
	for _, id := range elementIDs {
		expected[id] = struct{}{}
	}

	for id := range expected {
		if _, ok := waiting[id]; !ok {
			a.t.Fatalf("expected process instance %d to be waiting at element %q, "+
				"but current wait state is %v", a.key, id, setKeys(waiting))
		}
	}
	for id := range waiting {
		if _, ok := expected[id]; !ok {
			a.t.Fatalf("expected process instance %d to be waiting exactly at %v, "+
				"but also waiting at %q", a.key, elementIDs, id)
		}
	}
	return a
}

// --- Variable assertions ---

// HasVariable asserts that a variable with the given name exists for this process instance.
func (a *ProcessInstanceAssert) HasVariable(name string) *ProcessInstanceAssert {
	a.t.Helper()
	_, err := a.stream.WaitFor(a.ctx, func(r Record) bool {
		if r.ValueType != ValueTypeVariable || r.RecordType != RecordTypeEvent {
			return false
		}
		var v VariableValue
		if json.Unmarshal(r.Value, &v) != nil {
			return false
		}
		return v.ProcessInstanceKey == a.key && v.Name == name
	})
	if err != nil {
		a.t.Fatalf("expected process instance %d to have variable %q, but timed out waiting", a.key, name)
	}
	return a
}

// HasVariableWithValue asserts that a variable exists with the given name and value.
// The expectedValue is JSON-compared: strings are compared as JSON strings,
// other types are marshaled to JSON first.
func (a *ProcessInstanceAssert) HasVariableWithValue(name string, expectedValue any) *ProcessInstanceAssert {
	a.t.Helper()
	vars := a.processInstanceVariables()
	actualJSON, ok := vars[name]
	if !ok {
		a.t.Fatalf("expected process instance %d to have variable %q, but it does not exist. "+
			"Available variables: %v", a.key, name, variableNames(vars))
		return a
	}

	expectedJSON, err := json.Marshal(expectedValue)
	if err != nil {
		a.t.Fatalf("failed to marshal expected value for variable %q: %v", name, err)
		return a
	}

	if !jsonEqual(actualJSON, string(expectedJSON)) {
		a.t.Fatalf("expected process instance %d variable %q to have value %s, but was %s",
			a.key, name, string(expectedJSON), actualJSON)
	}
	return a
}

// --- Incident assertions ---

// HasAnyIncidents asserts that at least one incident has been created for this process instance.
func (a *ProcessInstanceAssert) HasAnyIncidents() *ProcessInstanceAssert {
	a.t.Helper()
	_, err := a.stream.WaitFor(a.ctx, func(r Record) bool {
		if r.ValueType != ValueTypeIncident || r.Intent != IntentCreated {
			return false
		}
		var v IncidentValue
		if json.Unmarshal(r.Value, &v) != nil {
			return false
		}
		return v.ProcessInstanceKey == a.key
	})
	if err != nil {
		a.t.Fatalf("expected process instance %d to have incidents, but timed out waiting", a.key)
	}
	return a
}

// HasNoIncidents asserts that no incidents have been created for this process instance (point-in-time).
func (a *ProcessInstanceAssert) HasNoIncidents() *ProcessInstanceAssert {
	a.t.Helper()
	incidents := a.stream.Filter(func(r Record) bool {
		if r.ValueType != ValueTypeIncident || r.Intent != IntentCreated {
			return false
		}
		var v IncidentValue
		if json.Unmarshal(r.Value, &v) != nil {
			return false
		}
		return v.ProcessInstanceKey == a.key
	})
	if len(incidents) > 0 {
		var v IncidentValue
		_ = json.Unmarshal(incidents[0].Value, &v)
		a.t.Fatalf("expected process instance %d to have no incidents, but found %d (first: %s on element %q)",
			a.key, len(incidents), v.ErrorType, v.ElementID)
	}
	return a
}

// --- Message assertions ---

// IsWaitingForMessages asserts that the process instance has open message subscriptions
// for ALL specified message names (point-in-time).
func (a *ProcessInstanceAssert) IsWaitingForMessages(messageNames ...string) *ProcessInstanceAssert {
	a.t.Helper()
	openSubs := a.openMessageSubscriptions()
	for _, name := range messageNames {
		if _, ok := openSubs[name]; !ok {
			a.t.Fatalf("expected process instance %d to be waiting for message %q, "+
				"but open subscriptions are %v", a.key, name, setKeys(openSubs))
		}
	}
	return a
}

// IsNotWaitingForMessages asserts that the process instance does NOT have open message subscriptions
// for any of the specified message names (point-in-time).
func (a *ProcessInstanceAssert) IsNotWaitingForMessages(messageNames ...string) *ProcessInstanceAssert {
	a.t.Helper()
	openSubs := a.openMessageSubscriptions()
	for _, name := range messageNames {
		if _, ok := openSubs[name]; ok {
			a.t.Fatalf("expected process instance %d NOT to be waiting for message %q, but it is", a.key, name)
		}
	}
	return a
}

// HasCorrelatedMessageByName asserts that a message with the given name has been correlated
// to this process instance the specified number of times (point-in-time).
func (a *ProcessInstanceAssert) HasCorrelatedMessageByName(messageName string, times int) *ProcessInstanceAssert {
	a.t.Helper()
	count := a.countCorrelatedMessages(func(v ProcessMessageSubscriptionValue) bool {
		return v.MessageName == messageName
	})
	if count != times {
		a.t.Fatalf("expected process instance %d to have correlated message %q %d time(s), but was %d",
			a.key, messageName, times, count)
	}
	return a
}

// HasCorrelatedMessageByCorrelationKey asserts that a message with the given correlation key
// has been correlated to this process instance the specified number of times (point-in-time).
func (a *ProcessInstanceAssert) HasCorrelatedMessageByCorrelationKey(correlationKey string, times int) *ProcessInstanceAssert {
	a.t.Helper()
	count := a.countCorrelatedMessages(func(v ProcessMessageSubscriptionValue) bool {
		return v.CorrelationKey == correlationKey
	})
	if count != times {
		a.t.Fatalf("expected process instance %d to have correlated message with key %q %d time(s), but was %d",
			a.key, correlationKey, times, count)
	}
	return a
}

// --- Called process assertions ---

// HasCalledProcess asserts that this process instance has started a child process via call activity.
// If processID is provided, checks for a specific BPMN process ID.
func (a *ProcessInstanceAssert) HasCalledProcess(processID ...string) *ProcessInstanceAssert {
	a.t.Helper()
	records := a.calledProcessRecords()
	if len(processID) > 0 {
		found := false
		for _, r := range records {
			var v ProcessInstanceValue
			if json.Unmarshal(r.Value, &v) == nil && v.BpmnProcessID == processID[0] {
				found = true
				break
			}
		}
		if !found {
			a.t.Fatalf("expected process instance %d to have called process %q, but it did not",
				a.key, processID[0])
		}
	} else if len(records) == 0 {
		a.t.Fatalf("expected process instance %d to have called a process, but it did not", a.key)
	}
	return a
}

// HasNotCalledProcess asserts that this process instance has NOT started any child process (point-in-time).
func (a *ProcessInstanceAssert) HasNotCalledProcess(processID ...string) *ProcessInstanceAssert {
	a.t.Helper()
	records := a.calledProcessRecords()
	if len(processID) > 0 {
		for _, r := range records {
			var v ProcessInstanceValue
			if json.Unmarshal(r.Value, &v) == nil && v.BpmnProcessID == processID[0] {
				a.t.Fatalf("expected process instance %d NOT to have called process %q, but it did",
					a.key, processID[0])
			}
		}
	} else if len(records) > 0 {
		a.t.Fatalf("expected process instance %d NOT to have called any process, but it called %d",
			a.key, len(records))
	}
	return a
}

// ExtractingLatestCalledProcess returns a new ProcessInstanceAssert for the most recently
// started child process. If processID is provided, filters by BPMN process ID.
func (a *ProcessInstanceAssert) ExtractingLatestCalledProcess(processID ...string) *ProcessInstanceAssert {
	a.t.Helper()
	records := a.calledProcessRecords()
	if len(processID) > 0 {
		var filtered []Record
		for _, r := range records {
			var v ProcessInstanceValue
			if json.Unmarshal(r.Value, &v) == nil && v.BpmnProcessID == processID[0] {
				filtered = append(filtered, r)
			}
		}
		records = filtered
	}
	if len(records) == 0 {
		if len(processID) > 0 {
			a.t.Fatalf("expected process instance %d to have called process %q, but found none", a.key, processID[0])
		} else {
			a.t.Fatalf("expected process instance %d to have called a process, but found none", a.key)
		}
	}
	latest := records[len(records)-1]
	var v ProcessInstanceValue
	_ = json.Unmarshal(latest.Value, &v)
	return ForProcessInstance(a.t, a.stream, v.ProcessInstanceKey).WithContext(a.ctx)
}

// --- Helpers ---

func (a *ProcessInstanceAssert) isProcessInstanceRecord(r Record) bool {
	if r.ValueType != ValueTypeProcessInstance || r.RecordType != RecordTypeEvent {
		return false
	}
	var v ProcessInstanceValue
	if json.Unmarshal(r.Value, &v) != nil {
		return false
	}
	return v.ProcessInstanceKey == a.key
}

func (*ProcessInstanceAssert) valueElementID(r Record) string {
	var v ProcessInstanceValue
	_ = json.Unmarshal(r.Value, &v)
	return v.ElementID
}

func (*ProcessInstanceAssert) valueBpmnElementType(r Record) string {
	var v ProcessInstanceValue
	_ = json.Unmarshal(r.Value, &v)
	return v.BpmnElementType
}

func (a *ProcessInstanceAssert) waitForProcessIntent(intent ZeebeIntent, msgFmt string) {
	a.t.Helper()
	_, err := a.stream.WaitFor(a.ctx, func(r Record) bool {
		return a.isProcessInstanceRecord(r) &&
			r.Intent == intent &&
			a.valueBpmnElementType(r) == BpmnElementTypeProcess
	})
	if err != nil {
		a.t.Fatalf(msgFmt+", but timed out waiting", a.key)
	}
}

func (a *ProcessInstanceAssert) hasProcessIntent(intent ZeebeIntent) bool {
	records := a.stream.Filter(func(r Record) bool {
		return a.isProcessInstanceRecord(r) &&
			r.Intent == intent &&
			a.valueBpmnElementType(r) == BpmnElementTypeProcess
	})
	return len(records) > 0
}

func (a *ProcessInstanceAssert) countCompletedElement(elementID string) int {
	records := a.stream.Filter(func(r Record) bool {
		return a.isProcessInstanceRecord(r) &&
			r.Intent == IntentElementCompleted &&
			a.valueElementID(r) == elementID
	})
	return len(records)
}

func (a *ProcessInstanceAssert) completedElementIDs() []string {
	records := a.stream.Filter(func(r Record) bool {
		return a.isProcessInstanceRecord(r) && r.Intent == IntentElementCompleted
	})
	ids := make([]string, 0, len(records))
	for _, r := range records {
		ids = append(ids, a.valueElementID(r))
	}
	return ids
}

func (a *ProcessInstanceAssert) elementsInWaitState() map[string]struct{} {
	records := a.stream.Filter(func(r Record) bool {
		return a.isProcessInstanceRecord(r)
	})

	doneKeys := make(map[int64]struct{})
	for _, r := range records {
		if r.Intent == IntentElementCompleted || r.Intent == IntentElementTerminated {
			doneKeys[r.Key] = struct{}{}
		}
	}

	waiting := make(map[string]struct{})
	for _, r := range records {
		if r.Intent == IntentElementActivated {
			if _, done := doneKeys[r.Key]; !done {
				waiting[a.valueElementID(r)] = struct{}{}
			}
		}
	}
	return waiting
}

func (a *ProcessInstanceAssert) processInstanceVariables() map[string]string {
	records := a.stream.Filter(func(r Record) bool {
		if r.ValueType != ValueTypeVariable || r.RecordType != RecordTypeEvent {
			return false
		}
		var v VariableValue
		if json.Unmarshal(r.Value, &v) != nil {
			return false
		}
		return v.ProcessInstanceKey == a.key
	})

	vars := make(map[string]string)
	for _, r := range records {
		var v VariableValue
		if json.Unmarshal(r.Value, &v) == nil {
			vars[v.Name] = v.Value
		}
	}
	return vars
}

func (a *ProcessInstanceAssert) openMessageSubscriptions() map[string]struct{} {
	type subKey struct {
		messageName        string
		elementInstanceKey int64
	}

	created := make(map[subKey]struct{})
	correlated := make(map[subKey]struct{})

	records := a.stream.Filter(func(r Record) bool {
		if r.ValueType != ValueTypeProcessMessageSubscription || r.RecordType != RecordTypeEvent {
			return false
		}
		var v ProcessMessageSubscriptionValue
		if json.Unmarshal(r.Value, &v) != nil {
			return false
		}
		return v.ProcessInstanceKey == a.key
	})

	for _, r := range records {
		var v ProcessMessageSubscriptionValue
		if json.Unmarshal(r.Value, &v) != nil {
			continue
		}
		sk := subKey{elementInstanceKey: v.ElementInstanceKey, messageName: v.MessageName}
		switch r.Intent {
		case IntentCreating, IntentCreated:
			created[sk] = struct{}{}
		case IntentCorrelated:
			correlated[sk] = struct{}{}
		}
	}

	open := make(map[string]struct{})
	for sk := range created {
		if _, ok := correlated[sk]; !ok {
			open[sk.messageName] = struct{}{}
		}
	}
	return open
}

func (a *ProcessInstanceAssert) incidentRecords() []Record {
	return a.stream.Filter(func(r Record) bool {
		if r.ValueType != ValueTypeIncident || r.Intent != IntentCreated {
			return false
		}
		var v IncidentValue
		if json.Unmarshal(r.Value, &v) != nil {
			return false
		}
		return v.ProcessInstanceKey == a.key
	})
}

func (a *ProcessInstanceAssert) countCorrelatedMessages(match func(ProcessMessageSubscriptionValue) bool) int {
	records := a.stream.Filter(func(r Record) bool {
		if r.ValueType != ValueTypeProcessMessageSubscription || r.Intent != IntentCorrelated || r.RecordType != RecordTypeEvent {
			return false
		}
		var v ProcessMessageSubscriptionValue
		if json.Unmarshal(r.Value, &v) != nil {
			return false
		}
		return v.ProcessInstanceKey == a.key && match(v)
	})
	return len(records)
}

func (a *ProcessInstanceAssert) calledProcessRecords() []Record {
	return a.stream.Filter(func(r Record) bool {
		if r.ValueType != ValueTypeProcessInstance || r.RecordType != RecordTypeEvent || r.Intent != IntentElementActivated {
			return false
		}
		var v ProcessInstanceValue
		if json.Unmarshal(r.Value, &v) != nil {
			return false
		}
		return v.ParentProcessInstanceKey == a.key && v.BpmnElementType == BpmnElementTypeProcess
	})
}

func setKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func variableNames(vars map[string]string) []string {
	names := make([]string, 0, len(vars))
	for k := range vars {
		names = append(names, k)
	}
	return names
}

func jsonEqual(a, b string) bool {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a == b {
		return true
	}

	var aVal, bVal any
	if json.Unmarshal([]byte(a), &aVal) != nil {
		return false
	}
	if json.Unmarshal([]byte(b), &bVal) != nil {
		return false
	}

	aBytes, _ := json.Marshal(aVal)
	bBytes, _ := json.Marshal(bVal)
	return bytes.Equal(aBytes, bBytes)
}

// ExtractVariables returns all current variables for this process instance as a map.
func (a *ProcessInstanceAssert) ExtractVariables() map[string]string {
	return a.processInstanceVariables()
}

// ExtractLatestIncident returns the IncidentValue of the most recent incident, or fails the test.
func (a *ProcessInstanceAssert) ExtractLatestIncident() IncidentValue {
	a.t.Helper()
	records := a.incidentRecords()
	if len(records) == 0 {
		a.t.Fatalf("expected process instance %d to have at least one incident, but found none", a.key)
	}
	latest := records[len(records)-1]
	var v IncidentValue
	if err := json.Unmarshal(latest.Value, &v); err != nil {
		a.t.Fatalf("failed to unmarshal incident value: %v", err)
	}
	return v
}

// ExtractingLatestIncident returns an IncidentAssert for the most recent incident
// on this process instance, allowing further fluent assertions on the incident.
func (a *ProcessInstanceAssert) ExtractingLatestIncident() *IncidentAssert {
	a.t.Helper()
	records := a.incidentRecords()
	if len(records) == 0 {
		a.t.Fatalf("expected process instance %d to have at least one incident, but found none", a.key)
	}
	return ForIncident(a.t, a.stream, records[len(records)-1].Key).WithContext(a.ctx)
}

// ExtractingVariablesAssert returns a VariablesMapAssert for fluent variable assertions.
func (a *ProcessInstanceAssert) ExtractingVariablesAssert() *VariablesMapAssert {
	a.t.Helper()
	return ForVariables(a.t, a.processInstanceVariables())
}

// String returns a debug description of the assert.
func (a *ProcessInstanceAssert) String() string {
	return fmt.Sprintf("ProcessInstanceAssert{key=%d}", a.key)
}
