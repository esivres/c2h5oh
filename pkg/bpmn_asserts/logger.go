package bpmn_asserts

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// PrintCompact returns a compact human-readable summary of all records in the stream.
// Includes a record log table and an unresolved incidents summary.
// This is the Go equivalent of RecordStream.print(true) from the Java library.
func (rs *RecordStream) PrintCompact() string {
	var sb strings.Builder

	sb.WriteString("\n")
	logRecords(rs, &sb)
	logUnresolvedIncidents(rs, &sb)

	return sb.String()
}

func logRecords(rs *RecordStream, sb *strings.Builder) {
	records := rs.Filter(func(Record) bool { return true })

	sb.WriteString("The following records have been recorded during this test:")
	for _, r := range records {
		sb.WriteString("\n")
		fmt.Fprintf(sb, "| %-20s%-35s%-30s| ", r.RecordType, r.ValueType, r.Intent)
		sb.WriteString(logRecordDetails(r))
	}
	sb.WriteString("\n")
}

func logRecordDetails(r Record) string {
	switch r.ValueType {
	case ValueTypeProcessInstance:
		return logProcessInstanceRecord(r)
	case ValueTypeJob:
		return logJobRecord(r)
	case ValueTypeDeployment:
		return logDeploymentRecord(r)
	case ValueTypeIncident:
		return logIncidentRecord(r)
	case ValueTypeMessage:
		return logMessageRecord(r)
	case ValueTypeMessageSubscription:
		return logMessageSubscriptionRecord(r)
	case ValueTypeProcessMessageSubscription:
		return logProcessMessageSubscriptionRecord(r)
	case ValueTypeJobBatch:
		return logJobBatchRecord(r)
	case ValueTypeTimer:
		return logTimerRecord(r)
	case ValueTypeMessageStartEventSub:
		return logMessageStartEventRecord(r)
	case ValueTypeVariable:
		return logVariableRecord(r)
	case ValueTypeProcessInstanceCreation:
		return logProcessCreationRecord(r)
	case ValueTypeError:
		return logErrorRecord(r)
	case ValueTypeProcess:
		return logProcessDefinitionRecord(r)
	case ValueTypeProcessEvent:
		return logProcessEventRecord(r)
	case ValueTypeForm:
		return logFormRecord(r)
	case ValueTypeUserTask:
		return logUserTaskRecord(r)
	case ValueTypeSignal:
		return logSignalRecord(r)
	case ValueTypeSignalSubscription:
		return logSignalSubscriptionRecord(r)
	default:
		return ""
	}
}

func logProcessInstanceRecord(r Record) string {
	var v ProcessInstanceValue
	if json.Unmarshal(r.Value, &v) != nil {
		return ""
	}
	return fmt.Sprintf("(Element id: %s), (Element type: %s), (Event type: %s), (Process id: %s)",
		v.ElementID, v.BpmnElementType, v.BpmnEventType, v.BpmnProcessID)
}

func logJobRecord(r Record) string {
	if r.RecordType != RecordTypeEvent {
		return ""
	}
	var v JobValue
	if json.Unmarshal(r.Value, &v) != nil {
		return ""
	}
	s := fmt.Sprintf("(Element id: %s), (Job type: %s)", v.ElementID, v.Type)
	if len(v.Variables) > 0 {
		s += ", " + formatVariablesMap(v.Variables)
	}
	return s
}

func logDeploymentRecord(r Record) string {
	var v DeploymentValue
	if json.Unmarshal(r.Value, &v) != nil {
		return ""
	}
	if len(v.Resources) == 0 {
		return ""
	}
	names := make([]string, 0, len(v.Resources))
	for _, res := range v.Resources {
		names = append(names, res.ResourceName)
	}
	return fmt.Sprintf("(Processes: [%s])", strings.Join(names, ", "))
}

func logIncidentRecord(r Record) string {
	if r.RecordType != RecordTypeEvent {
		return ""
	}
	var v IncidentValue
	if json.Unmarshal(r.Value, &v) != nil {
		return ""
	}
	return fmt.Sprintf("(Element id: %s), (Process id: %s)", v.ElementID, v.BpmnProcessID)
}

func logMessageRecord(r Record) string {
	var v MessageValue
	if json.Unmarshal(r.Value, &v) != nil {
		return ""
	}
	s := fmt.Sprintf("(Message name: %s), (Correlation key: %s)", v.Name, v.CorrelationKey)
	if len(v.Variables) > 0 {
		s += ", " + formatVariablesMap(v.Variables)
	}
	return s
}

func logMessageSubscriptionRecord(r Record) string {
	var v struct {
		Variables      map[string]any `json:"variables"`
		MessageName    string         `json:"messageName"`
		CorrelationKey string         `json:"correlationKey"`
	}
	if json.Unmarshal(r.Value, &v) != nil {
		return ""
	}
	s := fmt.Sprintf("(Message name: %s), (Correlation key: %s)", v.MessageName, v.CorrelationKey)
	if len(v.Variables) > 0 {
		s += ", " + formatVariablesMap(v.Variables)
	}
	return s
}

func logProcessMessageSubscriptionRecord(r Record) string {
	var v ProcessMessageSubscriptionValue
	if json.Unmarshal(r.Value, &v) != nil {
		return ""
	}
	s := fmt.Sprintf("(Message name: %s)", v.MessageName)
	if r.RecordType == RecordTypeEvent {
		s += fmt.Sprintf(", (Correlation key: %s), (Element id: %s)", v.CorrelationKey, v.ElementID)
	}
	if len(v.Variables) > 0 {
		s += ", " + formatVariablesMap(v.Variables)
	}
	return s
}

func logJobBatchRecord(r Record) string {
	var v struct {
		Worker string `json:"worker"`
		Type   string `json:"type"`
	}
	if json.Unmarshal(r.Value, &v) != nil {
		return ""
	}
	return fmt.Sprintf("(Worker: %s), (Job type: %s)", v.Worker, v.Type)
}

func logTimerRecord(r Record) string {
	var v TimerValue
	if json.Unmarshal(r.Value, &v) != nil {
		return ""
	}
	dueDate := time.UnixMilli(v.DueDate).Format(time.RFC3339)
	return fmt.Sprintf("(Element id: %s), (Due date: %s)", v.TargetElementID, dueDate)
}

func logMessageStartEventRecord(r Record) string {
	var v MessageStartEventSubscriptionValue
	if json.Unmarshal(r.Value, &v) != nil {
		return ""
	}
	return fmt.Sprintf("(Process id: %s), (Start event id: %s), (Message name: %s), (Correlation key: %s)",
		v.BpmnProcessID, v.StartEventID, v.MessageName, v.CorrelationKey)
}

func logVariableRecord(r Record) string {
	var v VariableValue
	if json.Unmarshal(r.Value, &v) != nil {
		return ""
	}
	return fmt.Sprintf("(Name: %s), (Value: %s)", v.Name, v.Value)
}

func logProcessCreationRecord(r Record) string {
	var v ProcessCreationValue
	if json.Unmarshal(r.Value, &v) != nil {
		return ""
	}
	s := fmt.Sprintf("(Process id: %s)", v.BpmnProcessID)
	if len(v.Variables) > 0 {
		s += ", " + formatVariablesMap(v.Variables)
	}
	return s
}

func logErrorRecord(r Record) string {
	var v struct {
		ExceptionMessage string `json:"exceptionMessage"`
	}
	if json.Unmarshal(r.Value, &v) != nil {
		return ""
	}
	return fmt.Sprintf("(Exception message: %s)", v.ExceptionMessage)
}

func logProcessDefinitionRecord(r Record) string {
	var v ProcessDefinitionValue
	if json.Unmarshal(r.Value, &v) != nil {
		return ""
	}
	return fmt.Sprintf("(Process: %s)", v.ResourceName)
}

func logProcessEventRecord(r Record) string {
	var v ProcessEventRecordValue
	if json.Unmarshal(r.Value, &v) != nil {
		return ""
	}
	s := fmt.Sprintf("(Target element id: %s)", v.TargetElementID)
	if len(v.Variables) > 0 {
		s += ", " + formatVariablesMap(v.Variables)
	}
	return s
}

func logFormRecord(r Record) string {
	var v FormValue
	if json.Unmarshal(r.Value, &v) != nil {
		return ""
	}
	return fmt.Sprintf("(Form: %s)", v.ResourceName)
}

func logUserTaskRecord(r Record) string {
	if r.RecordType != RecordTypeEvent {
		return ""
	}
	var v struct {
		ElementID string `json:"elementId"`
	}
	if json.Unmarshal(r.Value, &v) != nil {
		return ""
	}
	return fmt.Sprintf("(Element id: %s)", v.ElementID)
}

func logSignalRecord(r Record) string {
	var v struct {
		Variables  map[string]any `json:"variables"`
		SignalName string         `json:"signalName"`
	}
	if json.Unmarshal(r.Value, &v) != nil {
		return ""
	}
	s := fmt.Sprintf("(Signal name: %s)", v.SignalName)
	if len(v.Variables) > 0 {
		s += ", " + formatVariablesMap(v.Variables)
	}
	return s
}

func logSignalSubscriptionRecord(r Record) string {
	var v struct {
		BpmnProcessID         string `json:"bpmnProcessId"`
		CatchEventID          string `json:"catchEventId"`
		SignalName            string `json:"signalName"`
		CatchEventInstanceKey int64  `json:"catchEventInstanceKey"`
	}
	if json.Unmarshal(r.Value, &v) != nil {
		return ""
	}
	return fmt.Sprintf("(Process id: %s), (Catch event id: %s), (Signal name: %s), (Catch event instance key: %d)",
		v.BpmnProcessID, v.CatchEventID, v.SignalName, v.CatchEventInstanceKey)
}

func logUnresolvedIncidents(rs *RecordStream, sb *strings.Builder) {
	resolvedKeys := make(map[int64]struct{})
	resolved := rs.Filter(func(r Record) bool {
		return r.ValueType == ValueTypeIncident && r.Intent == IntentResolved
	})
	for _, r := range resolved {
		resolvedKeys[r.Key] = struct{}{}
	}

	created := rs.Filter(func(r Record) bool {
		if r.ValueType != ValueTypeIncident || r.Intent != IntentCreated {
			return false
		}
		_, isResolved := resolvedKeys[r.Key]
		return !isResolved
	})

	if len(created) == 0 {
		return
	}

	sb.WriteString("\nUnresolved incident(s) exist at the end of this test\n")
	for _, r := range created {
		var v IncidentValue
		if json.Unmarshal(r.Value, &v) != nil {
			continue
		}
		fmt.Fprintf(sb, "On element %s in process %s\n", v.ElementID, v.BpmnProcessID)
		fmt.Fprintf(sb, "\t- Error type: %s\n", v.ErrorType)
		fmt.Fprintf(sb, "\t- Error message: %s\n", v.ErrorMessage)
	}
	sb.WriteString("\nIf you did not expect any incidents to occur, then we recommend investigating ")
	sb.WriteString("these. These incidents may indicate what went wrong in your test case\n")
}

func formatVariablesMap(vars map[string]any) string {
	if len(vars) == 0 {
		return ""
	}
	parts := make([]string, 0, len(vars))
	for k, v := range vars {
		parts = append(parts, fmt.Sprintf("%s -> %v", k, v))
	}
	return fmt.Sprintf("(Variables: [%s])", strings.Join(parts, ", "))
}
