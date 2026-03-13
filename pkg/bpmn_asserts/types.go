package bpmn_asserts

// ZeebeValueType identifies the type of value carried by a record.
type ZeebeValueType string

const (
	ValueTypeProcessInstance            ZeebeValueType = "PROCESS_INSTANCE"
	ValueTypeJob                        ZeebeValueType = "JOB"
	ValueTypeJobBatch                   ZeebeValueType = "JOB_BATCH"
	ValueTypeDeployment                 ZeebeValueType = "DEPLOYMENT"
	ValueTypeIncident                   ZeebeValueType = "INCIDENT"
	ValueTypeMessage                    ZeebeValueType = "MESSAGE"
	ValueTypeMessageSubscription        ZeebeValueType = "MESSAGE_SUBSCRIPTION"
	ValueTypeProcessMessageSubscription ZeebeValueType = "PROCESS_MESSAGE_SUBSCRIPTION"
	ValueTypeMessageStartEventSub       ZeebeValueType = "MESSAGE_START_EVENT_SUBSCRIPTION"
	ValueTypeTimer                      ZeebeValueType = "TIMER"
	ValueTypeVariable                   ZeebeValueType = "VARIABLE"
	ValueTypeProcessInstanceCreation    ZeebeValueType = "PROCESS_INSTANCE_CREATION"
	ValueTypeError                      ZeebeValueType = "ERROR"
	ValueTypeProcess                    ZeebeValueType = "PROCESS"
	ValueTypeProcessEvent               ZeebeValueType = "PROCESS_EVENT"
	ValueTypeForm                       ZeebeValueType = "FORM"
	ValueTypeUserTask                   ZeebeValueType = "USER_TASK"
	ValueTypeSignal                     ZeebeValueType = "SIGNAL"
	ValueTypeSignalSubscription         ZeebeValueType = "SIGNAL_SUBSCRIPTION"
)

// ZeebeRecordType identifies the type of a record (event, command, or rejection).
type ZeebeRecordType string

const (
	RecordTypeEvent            ZeebeRecordType = "EVENT"
	RecordTypeCommand          ZeebeRecordType = "COMMAND"
	RecordTypeCommandRejection ZeebeRecordType = "COMMAND_REJECTION"
)

// ZeebeIntent describes the intent of a record (what happened or was requested).
type ZeebeIntent string

const (
	IntentCreated              ZeebeIntent = "CREATED"
	IntentCreating             ZeebeIntent = "CREATING"
	IntentCompleted            ZeebeIntent = "COMPLETED"
	IntentActivated            ZeebeIntent = "ACTIVATED"
	IntentResolved             ZeebeIntent = "RESOLVED"
	IntentExpired              ZeebeIntent = "EXPIRED"
	IntentCorrelated           ZeebeIntent = "CORRELATED"
	IntentTriggered            ZeebeIntent = "TRIGGERED"
	IntentUpdated              ZeebeIntent = "UPDATED"
	IntentFailed               ZeebeIntent = "FAILED"
	IntentDistribute           ZeebeIntent = "DISTRIBUTE"
	IntentPublished            ZeebeIntent = "PUBLISHED"
	IntentElementActivated     ZeebeIntent = "ELEMENT_ACTIVATED"
	IntentElementCompleted     ZeebeIntent = "ELEMENT_COMPLETED"
	IntentElementTerminated    ZeebeIntent = "ELEMENT_TERMINATED"
	IntentElementCompleting    ZeebeIntent = "ELEMENT_COMPLETING"
	IntentElementActivating    ZeebeIntent = "ELEMENT_ACTIVATING"
	IntentElementTerminating   ZeebeIntent = "ELEMENT_TERMINATING"
	IntentCanceled             ZeebeIntent = "CANCELED"
	IntentTimedOut             ZeebeIntent = "TIMED_OUT"
	IntentErrorThrown          ZeebeIntent = "ERROR_THROWN"
	IntentRecurredAfterBackoff ZeebeIntent = "RECURRED_AFTER_BACKOFF"
)

// BpmnElementType constants for element type discrimination in record values.
const BpmnElementTypeProcess = "PROCESS"
