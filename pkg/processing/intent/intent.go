// Package intent defines the intent model for the BPMN engine.
//
// An intent is a unit of work within a partition. Commands from the external world
// are converted into intents. Each intent is processed atomically and may produce
// follow-up intents that are added to the end of the partition queue.
//
// Intent carries data. Behavior (how to process the intent) is defined separately.
package intent

// Type identifies the kind of intent.
type Type string

const (
	// Process definition lifecycle
	DeployProcess Type = "DEPLOY_PROCESS"

	// Process instance lifecycle
	CreateProcessInstance   Type = "CREATE_PROCESS_INSTANCE"
	CancelProcessInstance   Type = "CANCEL_PROCESS_INSTANCE"
	CompleteProcessInstance Type = "COMPLETE_PROCESS_INSTANCE"

	// Element instance lifecycle
	ActivateElement  Type = "ACTIVATE_ELEMENT"
	ElementActivated Type = "ELEMENT_ACTIVATED"
	CompleteElement  Type = "COMPLETE_ELEMENT"
	ElementCompleted Type = "ELEMENT_COMPLETED"
	TerminateElement Type = "TERMINATE_ELEMENT"

	// Sequence flow
	TakeSequenceFlow Type = "TAKE_SEQUENCE_FLOW"

	// Job lifecycle
	CreateJob        Type = "CREATE_JOB"
	ActivateJob      Type = "ACTIVATE_JOB"
	CompleteJob      Type = "COMPLETE_JOB"
	FailJob          Type = "FAIL_JOB"
	ThrowJobError    Type = "THROW_JOB_ERROR"
	TimeOutJob       Type = "TIME_OUT_JOB"
	UpdateJobRetries Type = "UPDATE_JOB_RETRIES"
	UpdateJobTimeout Type = "UPDATE_JOB_TIMEOUT"

	// Variables
	SetVariables Type = "SET_VARIABLES"

	// Timer lifecycle
	CreateTimer  Type = "CREATE_TIMER"
	TriggerTimer Type = "TRIGGER_TIMER"
	CancelTimer  Type = "CANCEL_TIMER"

	// Message correlation
	PublishMessage    Type = "PUBLISH_MESSAGE"
	OpenSubscription  Type = "OPEN_SUBSCRIPTION"
	CorrelateMessage  Type = "CORRELATE_MESSAGE"
	CloseSubscription Type = "CLOSE_SUBSCRIPTION"

	// Signal
	ThrowSignal Type = "THROW_SIGNAL"

	// Form
	DeployForm Type = "DEPLOY_FORM"

	// Resource lifecycle
	DeleteResource Type = "DELETE_RESOURCE"

	// Incident lifecycle
	CreateIncident  Type = "CREATE_INCIDENT"
	ResolveIncident Type = "RESOLVE_INCIDENT"
)

// Origin indicates where the intent came from.
type Origin int

const (
	// External means the intent was created from a client command.
	// On error: retry with backoff, then create incident.
	External Origin = iota

	// Internal means the intent was spawned by another intent's behavior.
	// On error: create incident immediately.
	Internal
)

// Intent is the unit of work processed by the engine.
// Each concrete intent type embeds Header and adds type-specific fields.
type Intent interface {
	// IntentType returns the type of this intent.
	IntentType() Type

	// GetOrigin returns whether this intent is external or internal.
	GetOrigin() Origin

	// GetProcessInstanceKey returns the associated process instance key (0 if not applicable).
	GetProcessInstanceKey() uint64
}

// Header contains fields common to all intents.
type Header struct {
	// Key is the unique identifier assigned by the partition.
	Key uint64

	// Origin indicates external (from command) or internal (spawned by behavior).
	Origin Origin

	// ProcessInstanceKey is the associated process instance (0 if not applicable).
	ProcessInstanceKey uint64
}

func (h *Header) GetOrigin() Origin             { return h.Origin }
func (h *Header) GetProcessInstanceKey() uint64 { return h.ProcessInstanceKey }
func (h *Header) GetKey() uint64                { return h.Key }

// AssignKey sets the key if not already assigned. Used by the processor.
func (h *Header) AssignKey(k uint64) {
	if h.Key == 0 {
		h.Key = k
	}
}
