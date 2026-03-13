package bpmn_asserts

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// Record represents a single exported event from Zeebe's DebugLogExporter.
// Field names match the JSON output (camelCase via Jackson serialization).
type Record struct {
	PartitionID          int             `json:"partitionId"`
	Position             int64           `json:"position"`
	SourceRecordPosition int64           `json:"sourceRecordPosition"`
	Key                  int64           `json:"key"`
	Timestamp            int64           `json:"timestamp"`
	RecordType           ZeebeRecordType `json:"recordType"`
	ValueType            ZeebeValueType  `json:"valueType"`
	Intent               ZeebeIntent     `json:"intent"`
	RejectionType        string          `json:"rejectionType"`
	RejectionReason      string          `json:"rejectionReason"`
	BrokerVersion        string          `json:"brokerVersion"`
	Value                json.RawMessage `json:"value"`
}

// ProcessInstanceValue is the value payload for PROCESS_INSTANCE records.
type ProcessInstanceValue struct {
	BpmnProcessID            string `json:"bpmnProcessId"`
	Version                  int32  `json:"version"`
	ProcessDefinitionKey     int64  `json:"processDefinitionKey"`
	ProcessInstanceKey       int64  `json:"processInstanceKey"`
	ElementID                string `json:"elementId"`
	FlowScopeKey             int64  `json:"flowScopeKey"`
	BpmnElementType          string `json:"bpmnElementType"`
	ParentProcessInstanceKey int64  `json:"parentProcessInstanceKey"`
	ParentElementInstanceKey int64  `json:"parentElementInstanceKey"`
	BpmnEventType            string `json:"bpmnEventType"`
	TenantID                 string `json:"tenantId"`
}

// VariableValue is the value payload for VARIABLE records.
type VariableValue struct {
	Name                 string `json:"name"`
	Value                string `json:"value"`
	ScopeKey             int64  `json:"scopeKey"`
	ProcessInstanceKey   int64  `json:"processInstanceKey"`
	ProcessDefinitionKey int64  `json:"processDefinitionKey"`
	BpmnProcessID        string `json:"bpmnProcessId"`
	TenantID             string `json:"tenantId"`
}

// IncidentValue is the value payload for INCIDENT records.
type IncidentValue struct {
	ErrorType            string `json:"errorType"`
	ErrorMessage         string `json:"errorMessage"`
	BpmnProcessID        string `json:"bpmnProcessId"`
	ProcessInstanceKey   int64  `json:"processInstanceKey"`
	ElementID            string `json:"elementId"`
	ElementInstanceKey   int64  `json:"elementInstanceKey"`
	JobKey               int64  `json:"jobKey"`
	ProcessDefinitionKey int64  `json:"processDefinitionKey"`
	VariableScopeKey     int64  `json:"variableScopeKey"`
	TenantID             string `json:"tenantId"`
}

// JobValue is the value payload for JOB records.
type JobValue struct {
	Type                     string         `json:"type"`
	Worker                   string         `json:"worker"`
	Retries                  int32          `json:"retries"`
	Deadline                 int64          `json:"deadline"`
	ErrorMessage             string         `json:"errorMessage"`
	ErrorCode                string         `json:"errorCode"`
	ElementID                string         `json:"elementId"`
	ElementInstanceKey       int64          `json:"elementInstanceKey"`
	BpmnProcessID            string         `json:"bpmnProcessId"`
	ProcessDefinitionVersion int32          `json:"processDefinitionVersion"`
	ProcessInstanceKey       int64          `json:"processInstanceKey"`
	ProcessDefinitionKey     int64          `json:"processDefinitionKey"`
	CustomHeaders            map[string]any `json:"customHeaders"`
	Variables                map[string]any `json:"variables"`
	TenantID                 string         `json:"tenantId"`
}

// ProcessMessageSubscriptionValue is the value payload for PROCESS_MESSAGE_SUBSCRIPTION records.
type ProcessMessageSubscriptionValue struct {
	ProcessInstanceKey int64          `json:"processInstanceKey"`
	ElementInstanceKey int64          `json:"elementInstanceKey"`
	MessageName        string         `json:"messageName"`
	CorrelationKey     string         `json:"correlationKey"`
	BpmnProcessID      string         `json:"bpmnProcessId"`
	MessageKey         int64          `json:"messageKey"`
	ElementID          string         `json:"elementId"`
	Variables          map[string]any `json:"variables"`
	IsInterrupting     bool           `json:"isInterrupting"`
	TenantID           string         `json:"tenantId"`
}

// DeploymentValue is the value payload for DEPLOYMENT records.
type DeploymentValue struct {
	Resources                    []DeploymentResource             `json:"resources"`
	ProcessesMetadata            []DeploymentProcessMetadata      `json:"processesMetadata"`
	DecisionRequirementsMetadata []DeploymentDecisionReqsMetadata `json:"decisionRequirementsMetadata"`
	DecisionsMetadata            []DeploymentDecisionMetadata     `json:"decisionsMetadata"`
	FormMetadata                 []DeploymentFormMetadata         `json:"formMetadata"`
	TenantID                     string                           `json:"tenantId"`
	DeploymentKey                int64                            `json:"deploymentKey"`
}

// DeploymentResource is a single resource within a deployment.
type DeploymentResource struct {
	ResourceName string `json:"resourceName"`
}

// DeploymentProcessMetadata describes a process within a deployment.
type DeploymentProcessMetadata struct {
	BpmnProcessID        string `json:"bpmnProcessId"`
	Version              int32  `json:"version"`
	ProcessDefinitionKey int64  `json:"processDefinitionKey"`
	ResourceName         string `json:"resourceName"`
	IsDuplicate          bool   `json:"isDuplicate"`
}

// DeploymentFormMetadata describes a form within a deployment.
type DeploymentFormMetadata struct {
	FormID       string `json:"formId"`
	Version      int32  `json:"version"`
	FormKey      int64  `json:"formKey"`
	ResourceName string `json:"resourceName"`
	IsDuplicate  bool   `json:"isDuplicate"`
}

// DeploymentDecisionMetadata describes a decision within a deployment.
type DeploymentDecisionMetadata struct {
	DecisionID             string `json:"decisionId"`
	Version                int32  `json:"version"`
	DecisionKey            int64  `json:"decisionKey"`
	DecisionName           string `json:"decisionName"`
	DecisionRequirementsID string `json:"decisionRequirementsId"`
	IsDuplicate            bool   `json:"isDuplicate"`
}

// DeploymentDecisionReqsMetadata describes decision requirements within a deployment.
type DeploymentDecisionReqsMetadata struct {
	DecisionRequirementsID   string `json:"decisionRequirementsId"`
	DecisionRequirementsName string `json:"decisionRequirementsName"`
	ResourceName             string `json:"resourceName"`
	IsDuplicate              bool   `json:"isDuplicate"`
}

// MessageValue is the value payload for MESSAGE records.
type MessageValue struct {
	Name           string         `json:"name"`
	CorrelationKey string         `json:"correlationKey"`
	MessageID      string         `json:"messageId"`
	TimeToLive     int64          `json:"timeToLive"`
	Variables      map[string]any `json:"variables"`
	TenantID       string         `json:"tenantId"`
	Deadline       int64          `json:"deadline"`
}

// MessageStartEventSubscriptionValue is the value payload for MESSAGE_START_EVENT_SUBSCRIPTION records.
type MessageStartEventSubscriptionValue struct {
	ProcessDefinitionKey int64          `json:"processDefinitionKey"`
	StartEventID         string         `json:"startEventId"`
	MessageName          string         `json:"messageName"`
	BpmnProcessID        string         `json:"bpmnProcessId"`
	CorrelationKey       string         `json:"correlationKey"`
	MessageKey           int64          `json:"messageKey"`
	ProcessInstanceKey   int64          `json:"processInstanceKey"`
	Variables            map[string]any `json:"variables"`
	TenantID             string         `json:"tenantId"`
}

// TimerValue is the value payload for TIMER records.
type TimerValue struct {
	ElementInstanceKey   int64  `json:"elementInstanceKey"`
	DueDate              int64  `json:"dueDate"`
	Repetitions          int32  `json:"repetitions"`
	TargetElementID      string `json:"targetElementId"`
	ProcessInstanceKey   int64  `json:"processInstanceKey"`
	ProcessDefinitionKey int64  `json:"processDefinitionKey"`
	TenantID             string `json:"tenantId"`
}

// FormValue is the value payload for FORM records.
type FormValue struct {
	FormID       string `json:"formId"`
	Version      int32  `json:"version"`
	FormKey      int64  `json:"formKey"`
	ResourceName string `json:"resourceName"`
	IsDuplicate  bool   `json:"isDuplicate"`
	TenantID     string `json:"tenantId"`
}

// ProcessCreationValue is the value payload for PROCESS_INSTANCE_CREATION records.
type ProcessCreationValue struct {
	BpmnProcessID        string         `json:"bpmnProcessId"`
	Version              int32          `json:"version"`
	ProcessDefinitionKey int64          `json:"processDefinitionKey"`
	ProcessInstanceKey   int64          `json:"processInstanceKey"`
	Variables            map[string]any `json:"variables"`
	TenantID             string         `json:"tenantId"`
}

// ProcessDefinitionValue is the value payload for PROCESS records (definition, not instance).
type ProcessDefinitionValue struct {
	BpmnProcessID        string `json:"bpmnProcessId"`
	Version              int32  `json:"version"`
	ProcessDefinitionKey int64  `json:"processDefinitionKey"`
	ResourceName         string `json:"resourceName"`
	TenantID             string `json:"tenantId"`
}

// ProcessEventRecordValue is the value payload for PROCESS_EVENT records.
type ProcessEventRecordValue struct {
	ScopeKey             int64          `json:"scopeKey"`
	ProcessDefinitionKey int64          `json:"processDefinitionKey"`
	TargetElementID      string         `json:"targetElementId"`
	Variables            map[string]any `json:"variables"`
	TenantID             string         `json:"tenantId"`
}

// ParseValue unmarshals the record's Value into the specified type.
//
//	v, err := bpmn_assertions.ParseValue[bpmn_assertions.ProcessInstanceValue](record)
//
//	// In predicates:
//	Events().Where(func(r Record) bool {
//	    v, ok := bpmn_assertions.TryParseValue[JobValue](r)
//	    return ok && v.Type == "kamundarf:adhoc:v1"
//	})
func ParseValue[T any](r Record) (T, error) {
	var v T
	err := json.Unmarshal(r.Value, &v)
	return v, err
}

// MustParseValue unmarshals the record's Value into the specified type.
// Panics if unmarshalling fails.
func MustParseValue[T any](r Record) T {
	v, err := ParseValue[T](r)
	if err != nil {
		panic(fmt.Sprintf("bpmn_assertions: failed to parse record value (valueType=%s, intent=%s): %v",
			r.ValueType, r.Intent, err))
	}
	return v
}

// TryParseValue unmarshals the record's Value into the specified type.
// Returns the parsed value and true on success, or zero value and false on error.
// Convenient for use in predicates and filters.
func TryParseValue[T any](r Record) (T, bool) {
	v, err := ParseValue[T](r)
	return v, err == nil
}

type streamWaiter struct {
	pred func(Record) bool
	ch   chan Record
}

// RecordStream accumulates exported records from Zeebe container logs.
// It supports both point-in-time filtering and blocking waits for new records.
type RecordStream struct {
	mu      sync.Mutex
	records []Record
	waiters []streamWaiter
}

// NewRecordStream creates an empty RecordStream.
func NewRecordStream() *RecordStream {
	return &RecordStream{}
}

// Add parses a raw JSON record and appends it to the stream.
// Called by the log consumer when a DebugLogExporter line is detected.
func (rs *RecordStream) Add(raw json.RawMessage) {
	var rec Record
	if err := json.Unmarshal(raw, &rec); err != nil {
		return
	}

	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.records = append(rs.records, rec)

	remaining := rs.waiters[:0]
	for _, w := range rs.waiters {
		if w.pred(rec) {
			w.ch <- rec
		} else {
			remaining = append(remaining, w)
		}
	}
	rs.waiters = remaining
}

// Filter returns all records matching the predicate (point-in-time snapshot).
func (rs *RecordStream) Filter(pred func(Record) bool) []Record {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	var result []Record
	for _, r := range rs.records {
		if pred(r) {
			result = append(result, r)
		}
	}
	return result
}

// WaitFor blocks until a record matching the predicate appears or ctx is cancelled.
// It first checks existing records, then waits for new ones.
func (rs *RecordStream) WaitFor(ctx context.Context, pred func(Record) bool) (Record, error) {
	rs.mu.Lock()

	for _, r := range rs.records {
		if pred(r) {
			rs.mu.Unlock()
			return r, nil
		}
	}

	ch := make(chan Record, 1)
	rs.waiters = append(rs.waiters, streamWaiter{pred: pred, ch: ch})
	rs.mu.Unlock()

	select {
	case r := <-ch:
		return r, nil
	case <-ctx.Done():
		rs.mu.Lock()
		for i, w := range rs.waiters {
			if w.ch == ch {
				rs.waiters = append(rs.waiters[:i], rs.waiters[i+1:]...)
				break
			}
		}
		rs.mu.Unlock()
		return Record{}, ctx.Err()
	}
}

// Reset clears all accumulated records and pending waiters.
func (rs *RecordStream) Reset() {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.records = nil
	rs.waiters = nil
}
