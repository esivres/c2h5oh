// Package export defines the event notification system for the BPMN engine.
//
// Every state change in the engine produces an Event. Events are collected
// during a transaction and published after successful commit via an Exporter.
//
// Events represent facts (past tense): "process instance created",
// "element activated", "job completed". They are for external consumers:
// audit, analytics, monitoring, history.
package export

import "time"

// ValueType identifies the entity that changed.
type ValueType string

const (
	ValueProcessDefinition   ValueType = "PROCESS_DEFINITION"
	ValueProcessInstance     ValueType = "PROCESS_INSTANCE"
	ValueElementInstance     ValueType = "ELEMENT_INSTANCE"
	ValueVariable            ValueType = "VARIABLE"
	ValueJob                 ValueType = "JOB"
	ValueTimer               ValueType = "TIMER"
	ValueMessageSubscription ValueType = "MESSAGE_SUBSCRIPTION"
	ValueMessage             ValueType = "MESSAGE"
	ValueIncident            ValueType = "INCIDENT"
)

// RecordType describes what happened to the entity.
type RecordType string

const (
	RecordCreated    RecordType = "CREATED"
	RecordActivated  RecordType = "ACTIVATED"
	RecordCompleted  RecordType = "COMPLETED"
	RecordTerminated RecordType = "TERMINATED"
	RecordCanceled   RecordType = "CANCELED"
	RecordFailed     RecordType = "FAILED"
	RecordResolved   RecordType = "RESOLVED"
	RecordTriggered  RecordType = "TRIGGERED"
	RecordCorrelated RecordType = "CORRELATED"
	RecordUpdated    RecordType = "UPDATED"
	RecordDeleted    RecordType = "DELETED"
	RecordTimedOut   RecordType = "TIMED_OUT"
	RecordDeployed   RecordType = "DEPLOYED"
)

// Event represents a single state change in the engine.
type Event struct {
	// Position is the monotonically increasing sequence number within a partition.
	Position uint64

	// ValueType is the type of entity that changed.
	ValueType ValueType

	// RecordType is what happened.
	RecordType RecordType

	// Key is the entity key.
	Key uint64

	// ProcessInstanceKey is the associated process instance (0 if not applicable).
	ProcessInstanceKey uint64

	// ProcessDefinitionKey is the associated process definition.
	ProcessDefinitionKey uint64

	// ElementId is the BPMN element id (empty if not element-related).
	ElementId string

	// ElementType is the BPMN element type (empty if not element-related).
	ElementType string

	// Timestamp is when the event occurred.
	Timestamp time.Time

	// Value is the serialized entity state at the time of the event (JSON bytes).
	// Contains the full entity snapshot for consumers that need it.
	Value []byte
}
