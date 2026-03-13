package bpmn_asserts

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

// IncidentAssert provides fluent assertions on a Zeebe incident.
type IncidentAssert struct {
	t           testing.TB
	incidentKey int64
	stream      *RecordStream
	ctx         context.Context
}

// ForIncident creates a new IncidentAssert for the given incident key.
func ForIncident(t testing.TB, stream *RecordStream, incidentKey int64) *IncidentAssert {
	t.Helper()
	return &IncidentAssert{
		t:           t,
		incidentKey: incidentKey,
		stream:      stream,
		ctx:         context.Background(),
	}
}

// WithContext sets the timeout context for waiting assertions.
func (a *IncidentAssert) WithContext(ctx context.Context) *IncidentAssert {
	a.ctx = ctx
	return a
}

// HasErrorType asserts that the incident has the specified error type.
func (a *IncidentAssert) HasErrorType(expectedErrorType string) *IncidentAssert {
	a.t.Helper()
	v := a.getCreatedValue()
	if v.ErrorType != expectedErrorType {
		a.t.Fatalf("expected incident %d to have error type %q, but was %q",
			a.incidentKey, expectedErrorType, v.ErrorType)
	}
	return a
}

// HasErrorMessage asserts that the incident has the specified error message.
func (a *IncidentAssert) HasErrorMessage(expectedErrorMessage string) *IncidentAssert {
	a.t.Helper()
	v := a.getCreatedValue()
	if v.ErrorMessage != expectedErrorMessage {
		a.t.Fatalf("expected incident %d to have error message %q, but was %q",
			a.incidentKey, expectedErrorMessage, v.ErrorMessage)
	}
	return a
}

// ExtractingErrorMessage returns the error message for further custom assertions.
func (a *IncidentAssert) ExtractingErrorMessage() string {
	a.t.Helper()
	return a.getCreatedValue().ErrorMessage
}

// WasRaisedInProcessInstance asserts that the incident was raised in the specified process instance.
func (a *IncidentAssert) WasRaisedInProcessInstance(processInstanceKey int64) *IncidentAssert {
	a.t.Helper()
	v := a.getCreatedValue()
	if v.ProcessInstanceKey != processInstanceKey {
		a.t.Fatalf("expected incident %d to be raised in process instance %d, but was in %d",
			a.incidentKey, processInstanceKey, v.ProcessInstanceKey)
	}
	return a
}

// OccurredOnElement asserts that the incident occurred on the specified element.
func (a *IncidentAssert) OccurredOnElement(expectedElementID string) *IncidentAssert {
	a.t.Helper()
	v := a.getCreatedValue()
	if v.ElementID != expectedElementID {
		a.t.Fatalf("expected incident %d to have occurred on element %q, but was on %q",
			a.incidentKey, expectedElementID, v.ElementID)
	}
	return a
}

// OccurredDuringJob asserts that the incident occurred during the specified job.
func (a *IncidentAssert) OccurredDuringJob(expectedJobKey int64) *IncidentAssert {
	a.t.Helper()
	v := a.getCreatedValue()
	if v.JobKey != expectedJobKey {
		a.t.Fatalf("expected incident %d to have occurred during job %d, but was during job %d",
			a.incidentKey, expectedJobKey, v.JobKey)
	}
	return a
}

// IsResolved asserts that the incident has been resolved.
func (a *IncidentAssert) IsResolved() *IncidentAssert {
	a.t.Helper()
	_, err := a.stream.WaitFor(a.ctx, func(r Record) bool {
		return r.ValueType == ValueTypeIncident && r.Intent == IntentResolved && r.Key == a.incidentKey
	})
	if err != nil {
		a.t.Fatalf("expected incident %d to be resolved, but timed out waiting", a.incidentKey)
	}
	return a
}

// IsUnresolved asserts that the incident has NOT been resolved (point-in-time).
func (a *IncidentAssert) IsUnresolved() *IncidentAssert {
	a.t.Helper()
	resolved := a.stream.Filter(func(r Record) bool {
		return r.ValueType == ValueTypeIncident && r.Intent == IntentResolved && r.Key == a.incidentKey
	})
	if len(resolved) > 0 {
		a.t.Fatalf("expected incident %d to be unresolved, but it is resolved", a.incidentKey)
	}
	return a
}

// IncidentKey returns the incident key for use in further assertions.
func (a *IncidentAssert) IncidentKey() int64 {
	return a.incidentKey
}

func (a *IncidentAssert) getCreatedValue() IncidentValue {
	a.t.Helper()
	rec, err := a.stream.WaitFor(a.ctx, func(r Record) bool {
		return r.ValueType == ValueTypeIncident && r.Intent == IntentCreated && r.Key == a.incidentKey
	})
	if err != nil {
		a.t.Fatalf("expected to find CREATED record for incident %d, but timed out waiting", a.incidentKey)
	}
	var v IncidentValue
	if err := json.Unmarshal(rec.Value, &v); err != nil {
		a.t.Fatalf("failed to unmarshal incident value: %v", err)
	}
	return v
}

// String returns a debug description.
func (a *IncidentAssert) String() string {
	return fmt.Sprintf("IncidentAssert{key=%d}", a.incidentKey)
}
