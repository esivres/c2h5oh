package bpmn_asserts

import (
	"context"
	"encoding/json"
	"slices"
	"time"
)

// EventFilter is a fluent builder for constructing event predicates.
// Use it with RecordStream.WaitFor or ContainerSuite.WaitForExportedEvent.
//
// Usage:
//
//	// Wait for a process instance to complete
//	record, err := bpmn_assertions.Events().
//	    ProcessInstanceCompleted(processInstanceKey).
//	    WaitOn(s.RecordStream, 30*time.Second)
//
//	// Build a predicate for WaitForExportedEvent
//	pred := bpmn_assertions.Events().
//	    ElementCompleted(processInstanceKey, "serviceTask1").
//	    RawPredicate()
//	raw, err := s.WaitForExportedEvent(ctx, pred)
//
//	// Generic filters with typed constants
//	record, err := bpmn_assertions.Events().
//	    ValueType(bpmn_assertions.ValueTypeJob).
//	    Intent(bpmn_assertions.IntentCreated).
//	    Where(func(r Record) bool { return r.Key == jobKey }).
//	    WaitOnCtx(s.RecordStream, ctx)
type EventFilter struct {
	predicates []func(Record) bool
}

// Events creates a new EventFilter builder.
func Events() *EventFilter {
	return &EventFilter{}
}

// --- Generic filters ---

// ValueType filters records by value type (use ValueType* constants).
func (f *EventFilter) ValueType(vt ZeebeValueType) *EventFilter {
	f.predicates = append(f.predicates, func(r Record) bool {
		return r.ValueType == vt
	})
	return f
}

// Intent filters records by intent (use Intent* constants).
func (f *EventFilter) Intent(intent ZeebeIntent) *EventFilter {
	f.predicates = append(f.predicates, func(r Record) bool {
		return r.Intent == intent
	})
	return f
}

// RecordType filters records by record type (use RecordType* constants).
func (f *EventFilter) RecordType(rt ZeebeRecordType) *EventFilter {
	f.predicates = append(f.predicates, func(r Record) bool {
		return r.RecordType == rt
	})
	return f
}

// WithKey filters records by their key.
func (f *EventFilter) WithKey(key int64) *EventFilter {
	f.predicates = append(f.predicates, func(r Record) bool {
		return r.Key == key
	})
	return f
}

// Where adds a custom predicate.
func (f *EventFilter) Where(pred func(Record) bool) *EventFilter {
	f.predicates = append(f.predicates, pred)
	return f
}

// --- Value-specific filters (unmarshal Value internally) ---

// ProcessInstanceKey filters by processInstanceKey in the record value.
// Works for PROCESS_INSTANCE, VARIABLE, INCIDENT, JOB and other value types
// that contain a processInstanceKey field.
func (f *EventFilter) ProcessInstanceKey(key int64) *EventFilter {
	f.predicates = append(f.predicates, func(r Record) bool {
		var v struct {
			ProcessInstanceKey int64 `json:"processInstanceKey"`
		}
		return json.Unmarshal(r.Value, &v) == nil && v.ProcessInstanceKey == key
	})
	return f
}

// ElementID filters PROCESS_INSTANCE records by elementId.
func (f *EventFilter) ElementID(id string) *EventFilter {
	f.predicates = append(f.predicates, func(r Record) bool {
		var v struct {
			ElementID string `json:"elementId"`
		}
		return json.Unmarshal(r.Value, &v) == nil && v.ElementID == id
	})
	return f
}

// ElementType filters PROCESS_INSTANCE records by bpmnElementType.
func (f *EventFilter) ElementType(t string) *EventFilter {
	f.predicates = append(f.predicates, func(r Record) bool {
		var v struct {
			BpmnElementType string `json:"bpmnElementType"`
		}
		return json.Unmarshal(r.Value, &v) == nil && v.BpmnElementType == t
	})
	return f
}

// BpmnProcessID filters records by bpmnProcessId in the value.
func (f *EventFilter) BpmnProcessID(id string) *EventFilter {
	f.predicates = append(f.predicates, func(r Record) bool {
		var v DeploymentValue
		err := json.Unmarshal(r.Value, &v)
		return err == nil && slices.ContainsFunc(v.ProcessesMetadata, func(metadata DeploymentProcessMetadata) bool {
			return metadata.BpmnProcessID == id
		})
	})
	return f
}

// VariableName filters VARIABLE records by variable name.
func (f *EventFilter) VariableName(name string) *EventFilter {
	f.predicates = append(f.predicates, func(r Record) bool {
		var v struct {
			Name string `json:"name"`
		}
		return json.Unmarshal(r.Value, &v) == nil && v.Name == name
	})
	return f
}

// JobType filters JOB records by job type.
func (f *EventFilter) JobType(jobType string) *EventFilter {
	f.predicates = append(f.predicates, func(r Record) bool {
		var v JobValue
		return json.Unmarshal(r.Value, &v) == nil && v.Type == jobType
	})
	return f
}

// MessageName filters message-related records by messageName.
func (f *EventFilter) MessageName(name string) *EventFilter {
	f.predicates = append(f.predicates, func(r Record) bool {
		var v struct {
			MessageName string `json:"messageName"`
			Name        string `json:"name"`
		}
		if json.Unmarshal(r.Value, &v) != nil {
			return false
		}
		return v.MessageName == name || v.Name == name
	})
	return f
}

// ErrorType filters INCIDENT records by error type.
func (f *EventFilter) ErrorType(errorType string) *EventFilter {
	f.predicates = append(f.predicates, func(r Record) bool {
		var v struct {
			ErrorType string `json:"errorType"`
		}
		return json.Unmarshal(r.Value, &v) == nil && v.ErrorType == errorType
	})
	return f
}

// --- Shorthand builders for common patterns ---

// ProcessInstanceActivated filters for a process instance ELEMENT_ACTIVATED event.
func (f *EventFilter) ProcessInstanceActivated(processInstanceKey int64) *EventFilter {
	return f.ValueType(ValueTypeProcessInstance).
		RecordType(RecordTypeEvent).
		Intent(IntentElementActivated).
		ProcessInstanceKey(processInstanceKey).
		ElementType("PROCESS")
}

// ProcessInstanceCompleted filters for a process instance ELEMENT_COMPLETED event.
func (f *EventFilter) ProcessInstanceCompleted(processInstanceKey int64) *EventFilter {
	return f.ValueType(ValueTypeProcessInstance).
		RecordType(RecordTypeEvent).
		Intent(IntentElementCompleted).
		ProcessInstanceKey(processInstanceKey).
		ElementType("PROCESS")
}

// ProcessInstanceTerminated filters for a process instance ELEMENT_TERMINATED event.
func (f *EventFilter) ProcessInstanceTerminated(processInstanceKey int64) *EventFilter {
	return f.ValueType(ValueTypeProcessInstance).
		RecordType(RecordTypeEvent).
		Intent(IntentElementTerminated).
		ProcessInstanceKey(processInstanceKey).
		ElementType("PROCESS")
}

// ElementCompleted filters for a specific element's ELEMENT_COMPLETED event.
func (f *EventFilter) ElementCompleted(processInstanceKey int64, elementID string) *EventFilter {
	return f.ValueType(ValueTypeProcessInstance).
		RecordType(RecordTypeEvent).
		Intent(IntentElementCompleted).
		ProcessInstanceKey(processInstanceKey).
		ElementID(elementID)
}

// ElementActivated filters for a specific element's ELEMENT_ACTIVATED event.
func (f *EventFilter) ElementActivated(processInstanceKey int64, elementID string) *EventFilter {
	return f.ValueType(ValueTypeProcessInstance).
		RecordType(RecordTypeEvent).
		Intent(IntentElementActivated).
		ProcessInstanceKey(processInstanceKey).
		ElementID(elementID)
}

// JobCreated filters for a JOB CREATED event with the specified type.
func (f *EventFilter) JobCreated(jobType string) *EventFilter {
	return f.ValueType(ValueTypeJob).
		RecordType(RecordTypeEvent).
		Intent(IntentCreated).
		JobType(jobType)
}

// JobCompleted filters for a JOB COMPLETED event with the specified type.
func (f *EventFilter) JobCompleted(jobType string) *EventFilter {
	return f.ValueType(ValueTypeJob).
		RecordType(RecordTypeEvent).
		Intent(IntentCompleted).
		JobType(jobType)
}

// JobCanceled filters for a JOB CANCELED event with the specified type.
func (f *EventFilter) JobCanceled(jobType string) *EventFilter {
	return f.ValueType(ValueTypeJob).
		RecordType(RecordTypeEvent).
		Intent(IntentCanceled).
		JobType(jobType)
}

// IncidentCreated filters for an INCIDENT CREATED event for the given process instance.
func (f *EventFilter) IncidentCreated(processInstanceKey int64) *EventFilter {
	return f.ValueType(ValueTypeIncident).
		Intent(IntentCreated).
		ProcessInstanceKey(processInstanceKey)
}

// IncidentResolved filters for an INCIDENT RESOLVED event for the given process instance.
func (f *EventFilter) IncidentResolved(processInstanceKey int64) *EventFilter {
	return f.ValueType(ValueTypeIncident).
		Intent(IntentResolved).
		ProcessInstanceKey(processInstanceKey)
}

// VariableCreated filters for a VARIABLE CREATED event with the given name.
func (f *EventFilter) VariableCreated(processInstanceKey int64, name string) *EventFilter {
	return f.ValueType(ValueTypeVariable).
		RecordType(RecordTypeEvent).
		Intent(IntentCreated).
		ProcessInstanceKey(processInstanceKey).
		VariableName(name)
}

// VariableUpdated filters for a VARIABLE UPDATED event with the given name.
func (f *EventFilter) VariableUpdated(processInstanceKey int64, name string) *EventFilter {
	return f.ValueType(ValueTypeVariable).
		RecordType(RecordTypeEvent).
		Intent(IntentUpdated).
		ProcessInstanceKey(processInstanceKey).
		VariableName(name)
}

// MessageCorrelated filters for a PROCESS_MESSAGE_SUBSCRIPTION CORRELATED event.
func (f *EventFilter) MessageCorrelated(processInstanceKey int64, messageName string) *EventFilter {
	return f.ValueType(ValueTypeProcessMessageSubscription).
		RecordType(RecordTypeEvent).
		Intent(IntentCorrelated).
		ProcessInstanceKey(processInstanceKey).
		MessageName(messageName)
}

// DeploymentCreated filters for a DEPLOYMENT CREATED event.
func (f *EventFilter) DeploymentCreated() *EventFilter {
	return f.ValueType(ValueTypeDeployment).
		Intent(IntentCreated)
}

// --- Output: predicates ---

// RecordPredicate returns a predicate function for use with RecordStream.WaitFor or Filter.
func (f *EventFilter) RecordPredicate() func(Record) bool {
	preds := make([]func(Record) bool, len(f.predicates))
	copy(preds, f.predicates)
	return func(r Record) bool {
		for _, p := range preds {
			if !p(r) {
				return false
			}
		}
		return true
	}
}

// RawPredicate returns a predicate function for use with ContainerSuite.WaitForExportedEvent.
// It parses the raw JSON into a Record and applies all filters.
func (f *EventFilter) RawPredicate() func(json.RawMessage) bool {
	recordPred := f.RecordPredicate()
	return func(raw json.RawMessage) bool {
		var r Record
		if json.Unmarshal(raw, &r) != nil {
			return false
		}
		return recordPred(r)
	}
}

// --- Output: wait ---

// WaitOn waits for a matching record on the given stream with a timeout duration.
// Returns the matching Record or an error if the timeout expires.
func (f *EventFilter) WaitOn(stream *RecordStream, timeout time.Duration) (Record, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return stream.WaitFor(ctx, f.RecordPredicate())
}

// WaitOnCtx waits for a matching record on the given stream using the provided context.
func (f *EventFilter) WaitOnCtx(stream *RecordStream, ctx context.Context) (Record, error) {
	return stream.WaitFor(ctx, f.RecordPredicate())
}

// FindAll returns all matching records from the stream (point-in-time).
func (f *EventFilter) FindAll(stream *RecordStream) []Record {
	return stream.Filter(f.RecordPredicate())
}

// FindFirst returns the first matching record, or nil if none found (point-in-time).
func (f *EventFilter) FindFirst(stream *RecordStream) *Record {
	records := f.FindAll(stream)
	if len(records) == 0 {
		return nil
	}
	return &records[0]
}

// Count returns the number of matching records (point-in-time).
func (f *EventFilter) Count(stream *RecordStream) int {
	return len(f.FindAll(stream))
}
