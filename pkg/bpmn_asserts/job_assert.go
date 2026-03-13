package bpmn_asserts

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

// JobAssert provides fluent assertions on a Zeebe job.
type JobAssert struct {
	t      testing.TB
	ctx    context.Context
	stream *RecordStream
	jobKey int64
}

// ForJob creates a new JobAssert for the given job key.
func ForJob(t testing.TB, stream *RecordStream, jobKey int64) *JobAssert {
	t.Helper()
	return &JobAssert{
		t:      t,
		jobKey: jobKey,
		stream: stream,
		ctx:    context.Background(),
	}
}

// WithContext sets the timeout context for waiting assertions.
func (a *JobAssert) WithContext(ctx context.Context) *JobAssert {
	a.ctx = ctx
	return a
}

// HasElementID asserts that the job belongs to the specified element.
func (a *JobAssert) HasElementID(expectedElementID string) *JobAssert {
	a.t.Helper()
	v := a.getJobValue()
	if v.ElementID != expectedElementID {
		a.t.Fatalf("expected job %d to have element ID %q, but was %q",
			a.jobKey, expectedElementID, v.ElementID)
	}
	return a
}

// HasBpmnProcessID asserts that the job belongs to the specified BPMN process.
func (a *JobAssert) HasBpmnProcessID(expectedBpmnProcessID string) *JobAssert {
	a.t.Helper()
	v := a.getJobValue()
	if v.BpmnProcessID != expectedBpmnProcessID {
		a.t.Fatalf("expected job %d to have BPMN process ID %q, but was %q",
			a.jobKey, expectedBpmnProcessID, v.BpmnProcessID)
	}
	return a
}

// HasRetries asserts that the job has the specified number of retries.
func (a *JobAssert) HasRetries(expectedRetries int32) *JobAssert {
	a.t.Helper()
	v := a.getJobValue()
	if v.Retries != expectedRetries {
		a.t.Fatalf("expected job %d to have %d retries, but was %d",
			a.jobKey, expectedRetries, v.Retries)
	}
	return a
}

// HasDeadline asserts that the job deadline is within the specified offset of the expected value.
func (a *JobAssert) HasDeadline(expectedDeadline int64, offset int64) *JobAssert {
	a.t.Helper()
	v := a.getJobValue()
	diff := v.Deadline - expectedDeadline
	if diff < 0 {
		diff = -diff
	}
	if diff > offset {
		a.t.Fatalf("expected job %d deadline to be %d (±%d), but was %d",
			a.jobKey, expectedDeadline, offset, v.Deadline)
	}
	return a
}

// HasType asserts that the job has the specified type.
func (a *JobAssert) HasType(expectedType string) *JobAssert {
	a.t.Helper()
	v := a.getJobValue()
	if v.Type != expectedType {
		a.t.Fatalf("expected job %d to have type %q, but was %q",
			a.jobKey, expectedType, v.Type)
	}
	return a
}

// IsCanceled asserts that the job has been canceled (waits).
func (a *JobAssert) IsCanceled() *JobAssert {
	a.t.Helper()
	_, err := a.stream.WaitFor(a.ctx, func(r Record) bool {
		return r.ValueType == ValueTypeJob && r.Intent == IntentCanceled && r.Key == a.jobKey
	})
	if err != nil {
		a.t.Fatalf("expected job %d to be canceled, but timed out waiting", a.jobKey)
	}
	return a
}

// IsNotCanceled asserts that the job has NOT been canceled (point-in-time).
func (a *JobAssert) IsNotCanceled() *JobAssert {
	a.t.Helper()
	canceled := a.stream.Filter(func(r Record) bool {
		return r.ValueType == ValueTypeJob && r.Intent == IntentCanceled && r.Key == a.jobKey
	})
	if len(canceled) > 0 {
		a.t.Fatalf("expected job %d NOT to be canceled, but it is", a.jobKey)
	}
	return a
}

// HasAnyIncidents asserts that at least one incident has been created for this job.
func (a *JobAssert) HasAnyIncidents() *JobAssert {
	a.t.Helper()
	_, err := a.stream.WaitFor(a.ctx, func(r Record) bool {
		if r.ValueType != ValueTypeIncident || r.Intent != IntentCreated {
			return false
		}
		var v IncidentValue
		if json.Unmarshal(r.Value, &v) != nil {
			return false
		}
		return v.JobKey == a.jobKey
	})
	if err != nil {
		a.t.Fatalf("expected job %d to have incidents, but timed out waiting", a.jobKey)
	}
	return a
}

// HasNoIncidents asserts that no incidents have been created for this job (point-in-time).
func (a *JobAssert) HasNoIncidents() *JobAssert {
	a.t.Helper()
	incidents := a.stream.Filter(func(r Record) bool {
		if r.ValueType != ValueTypeIncident || r.Intent != IntentCreated {
			return false
		}
		var v IncidentValue
		if json.Unmarshal(r.Value, &v) != nil {
			return false
		}
		return v.JobKey == a.jobKey
	})
	if len(incidents) > 0 {
		a.t.Fatalf("expected job %d to have no incidents, but found %d", a.jobKey, len(incidents))
	}
	return a
}

// ExtractingLatestIncident returns an IncidentAssert for the most recent incident on this job.
func (a *JobAssert) ExtractingLatestIncident() *IncidentAssert {
	a.t.Helper()
	incidents := a.stream.Filter(func(r Record) bool {
		if r.ValueType != ValueTypeIncident || r.Intent != IntentCreated {
			return false
		}
		var v IncidentValue
		if json.Unmarshal(r.Value, &v) != nil {
			return false
		}
		return v.JobKey == a.jobKey
	})
	if len(incidents) == 0 {
		a.t.Fatalf("expected job %d to have at least one incident, but found none", a.jobKey)
	}
	return ForIncident(a.t, a.stream, incidents[len(incidents)-1].Key).WithContext(a.ctx)
}

// ExtractingVariables returns the job's variables for custom assertions.
func (a *JobAssert) ExtractingVariables() map[string]any {
	a.t.Helper()
	return a.getJobValue().Variables
}

// ExtractingHeaders returns the job's custom headers for custom assertions.
func (a *JobAssert) ExtractingHeaders() map[string]any {
	a.t.Helper()
	return a.getJobValue().CustomHeaders
}

func (a *JobAssert) getJobValue() JobValue {
	a.t.Helper()
	rec, err := a.stream.WaitFor(a.ctx, func(r Record) bool {
		return r.ValueType == ValueTypeJob && r.Intent == IntentCreated && r.Key == a.jobKey
	})
	if err != nil {
		a.t.Fatalf("expected to find CREATED record for job %d, but timed out waiting", a.jobKey)
	}
	var v JobValue
	if err := json.Unmarshal(rec.Value, &v); err != nil {
		a.t.Fatalf("failed to unmarshal job value: %v", err)
	}
	return v
}

// String returns a debug description.
func (a *JobAssert) String() string {
	return fmt.Sprintf("JobAssert{key=%d}", a.jobKey)
}
