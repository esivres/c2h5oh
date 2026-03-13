package bpmn_asserts

import (
	"encoding/json"
	"fmt"
	"testing"
)

// --- Test record builders ---

func makeRecord(valueType ZeebeValueType, recordType ZeebeRecordType, intent ZeebeIntent, key int64, value any) Record {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal test value: %v", err))
	}
	return Record{
		PartitionID: 1,
		Position:    1,
		Key:         key,
		RecordType:  recordType,
		ValueType:   valueType,
		Intent:      intent,
		Value:       raw,
	}
}

func makeProcessInstanceRecord(intent ZeebeIntent, key, processInstanceKey int64, elementID, elementType string) Record {
	return makeRecord(ValueTypeProcessInstance, RecordTypeEvent, intent, key, ProcessInstanceValue{
		ProcessInstanceKey: processInstanceKey,
		ElementID:          elementID,
		BpmnElementType:    elementType,
		BpmnProcessID:      "testProcess",
	})
}

func makeProcessRecord(intent ZeebeIntent, key, processInstanceKey int64, elementID, elementType, processID string) Record {
	return makeRecord(ValueTypeProcessInstance, RecordTypeEvent, intent, key, ProcessInstanceValue{
		ProcessInstanceKey:       processInstanceKey,
		ElementID:                elementID,
		BpmnElementType:          elementType,
		BpmnProcessID:            processID,
		ParentProcessInstanceKey: 0,
	})
}

func makeChildProcessRecord(intent ZeebeIntent, key, processInstanceKey, parentKey int64, processID string) Record {
	return makeRecord(ValueTypeProcessInstance, RecordTypeEvent, intent, key, ProcessInstanceValue{
		ProcessInstanceKey:       processInstanceKey,
		ElementID:                processID,
		BpmnElementType:          "PROCESS",
		BpmnProcessID:            processID,
		ParentProcessInstanceKey: parentKey,
	})
}

func makeVariableRecord(intent ZeebeIntent, key, processInstanceKey int64, name, value string) Record {
	return makeRecord(ValueTypeVariable, RecordTypeEvent, intent, key, VariableValue{
		Name:               name,
		Value:              value,
		ProcessInstanceKey: processInstanceKey,
	})
}

func makeIncidentRecord(intent ZeebeIntent, key, processInstanceKey int64, errorType, errorMsg, elementID string, jobKey int64) Record {
	return makeRecord(ValueTypeIncident, RecordTypeEvent, intent, key, IncidentValue{
		ErrorType:          errorType,
		ErrorMessage:       errorMsg,
		ProcessInstanceKey: processInstanceKey,
		ElementID:          elementID,
		JobKey:             jobKey,
		BpmnProcessID:      "testProcess",
	})
}

func makeJobRecord(intent ZeebeIntent, key int64, jobType, elementID string, processInstanceKey int64, retries int32) Record {
	return makeRecord(ValueTypeJob, RecordTypeEvent, intent, key, JobValue{
		Type:               jobType,
		ElementID:          elementID,
		ProcessInstanceKey: processInstanceKey,
		Retries:            retries,
		BpmnProcessID:      "testProcess",
		Variables:          map[string]any{"input": "value"},
		CustomHeaders:      map[string]any{"header1": "val1"},
	})
}

func makeMessageSubscriptionRecord(intent ZeebeIntent, key, processInstanceKey, elementInstanceKey int64, messageName string) Record {
	return makeRecord(ValueTypeProcessMessageSubscription, RecordTypeEvent, intent, key, ProcessMessageSubscriptionValue{
		ProcessInstanceKey: processInstanceKey,
		ElementInstanceKey: elementInstanceKey,
		MessageName:        messageName,
		CorrelationKey:     "corr-123",
		ElementID:          "messageCatch",
		MessageKey:         key + 1000,
	})
}

func makeDeploymentRecord(intent ZeebeIntent, key int64, processIDs []string, resourceNames []string) Record {
	metadata := make([]DeploymentProcessMetadata, len(processIDs))
	resources := make([]DeploymentResource, len(resourceNames))
	for i, id := range processIDs {
		rn := ""
		if i < len(resourceNames) {
			rn = resourceNames[i]
		}
		metadata[i] = DeploymentProcessMetadata{
			BpmnProcessID:        id,
			Version:              1,
			ProcessDefinitionKey: int64(100 + i),
			ResourceName:         rn,
		}
	}
	for i, n := range resourceNames {
		resources[i] = DeploymentResource{ResourceName: n}
	}
	return makeRecord(ValueTypeDeployment, RecordTypeEvent, intent, key, DeploymentValue{
		ProcessesMetadata: metadata,
		Resources:         resources,
		DeploymentKey:     key,
	})
}

func makeMessageRecord(intent ZeebeIntent, key int64, name, correlationKey string) Record {
	return makeRecord(ValueTypeMessage, RecordTypeEvent, intent, key, MessageValue{
		Name:           name,
		CorrelationKey: correlationKey,
	})
}

func makeProcessDefinitionRecord(intent ZeebeIntent, key int64, bpmnProcessID string, version int32, resourceName string) Record {
	return makeRecord(ValueTypeProcess, RecordTypeEvent, intent, key, ProcessDefinitionValue{
		BpmnProcessID:        bpmnProcessID,
		Version:              version,
		ProcessDefinitionKey: key,
		ResourceName:         resourceName,
	})
}

func makeFormRecord(intent ZeebeIntent, key int64, formID string, version int32, resourceName string) Record {
	return makeRecord(ValueTypeForm, RecordTypeEvent, intent, key, FormValue{
		FormID:       formID,
		Version:      version,
		FormKey:      key,
		ResourceName: resourceName,
	})
}

func makeMessageCorrelatedSubscription(key, processInstanceKey, messageKey int64, messageName string) Record {
	return makeRecord(ValueTypeProcessMessageSubscription, RecordTypeEvent, IntentCorrelated, key, ProcessMessageSubscriptionValue{
		ProcessInstanceKey: processInstanceKey,
		MessageName:        messageName,
		MessageKey:         messageKey,
		CorrelationKey:     "corr-123",
	})
}

func makeMessageStartEventCorrelated(key, processInstanceKey, messageKey int64, messageName string) Record {
	return makeRecord(ValueTypeMessageStartEventSub, RecordTypeEvent, IntentCorrelated, key, MessageStartEventSubscriptionValue{
		ProcessInstanceKey: processInstanceKey,
		MessageName:        messageName,
		MessageKey:         messageKey,
	})
}

func makeTimerRecord(intent ZeebeIntent, key int64, targetElementID string, processInstanceKey, processDefinitionKey int64) Record {
	return makeRecord(ValueTypeTimer, RecordTypeEvent, intent, key, TimerValue{
		TargetElementID:      targetElementID,
		ProcessInstanceKey:   processInstanceKey,
		ProcessDefinitionKey: processDefinitionKey,
	})
}

// feedRecords adds multiple records to a stream.
func feedRecords(stream *RecordStream, records ...Record) {
	for _, r := range records {
		raw, _ := json.Marshal(r)
		stream.Add(raw)
	}
}

// fatalCatcher replaces testing.TB to capture Fatal calls instead of terminating.
type fatalCatcher struct {
	testing.TB
	fatalMsg string
	fataled  bool
}

func (f *fatalCatcher) Fatalf(format string, args ...any) {
	f.fatalMsg = fmt.Sprintf(format, args...)
	f.fataled = true
	panic("fatalf-caught") // use panic to stop execution flow
}

func (f *fatalCatcher) Fatal(args ...any) {
	f.fatalMsg = fmt.Sprint(args...)
	f.fataled = true
	panic("fatalf-caught")
}

func (f *fatalCatcher) Helper() {}

// catchFatal runs fn and catches any Fatal from fatalCatcher.
func catchFatal(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			if r != "fatalf-caught" {
				panic(r)
			}
		}
	}()
	fn()
}
