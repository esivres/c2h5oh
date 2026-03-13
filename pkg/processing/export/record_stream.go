package export

import (
	"context"
	"encoding/json"
	"time"

	"github.com/esivres/c2h5oh/pkg/bpmn_asserts"
)

// RecordStreamExporter converts export.Events into bpmn_asserts.Record JSON
// and feeds them into a RecordStream for use in assertions.
type RecordStreamExporter struct {
	stream *bpmn_asserts.RecordStream
}

// NewRecordStreamExporter creates an exporter that feeds into a RecordStream.
func NewRecordStreamExporter(stream *bpmn_asserts.RecordStream) *RecordStreamExporter {
	return &RecordStreamExporter{stream: stream}
}

func (e *RecordStreamExporter) Export(_ context.Context, events []Event) error {
	for _, ev := range events {
		rec := e.toRecord(ev)
		raw, err := json.Marshal(rec)
		if err != nil {
			continue
		}
		e.stream.Add(raw)
	}
	return nil
}

func (*RecordStreamExporter) toRecord(ev Event) bpmn_asserts.Record {
	return bpmn_asserts.Record{
		PartitionID: 1,
		Position:    int64(ev.Position),
		Key:         int64(ev.Key),
		Timestamp:   ev.Timestamp.UnixMilli(),
		RecordType:  bpmn_asserts.RecordTypeEvent,
		ValueType:   mapValueType(ev.ValueType),
		Intent:      mapIntent(ev.ValueType, ev.RecordType),
		Value:       buildValue(ev),
	}
}

func mapValueType(vt ValueType) bpmn_asserts.ZeebeValueType {
	switch vt {
	case ValueProcessDefinition:
		return bpmn_asserts.ValueTypeProcess
	case ValueProcessInstance:
		return bpmn_asserts.ValueTypeProcessInstance
	case ValueElementInstance:
		return bpmn_asserts.ValueTypeProcessInstance // Elements map to PROCESS_INSTANCE in Zeebe
	case ValueVariable:
		return bpmn_asserts.ValueTypeVariable
	case ValueJob:
		return bpmn_asserts.ValueTypeJob
	case ValueTimer:
		return bpmn_asserts.ValueTypeTimer
	case ValueMessageSubscription:
		return bpmn_asserts.ValueTypeProcessMessageSubscription
	case ValueMessage:
		return bpmn_asserts.ValueTypeMessage
	case ValueIncident:
		return bpmn_asserts.ValueTypeIncident
	default:
		return bpmn_asserts.ZeebeValueType(string(vt))
	}
}

func mapIntent(vt ValueType, rt RecordType) bpmn_asserts.ZeebeIntent {
	// Element instances use ELEMENT_* intents in Zeebe
	if vt == ValueElementInstance {
		switch rt {
		case RecordActivated:
			return bpmn_asserts.IntentElementActivated
		case RecordCompleted:
			return bpmn_asserts.IntentElementCompleted
		case RecordTerminated:
			return bpmn_asserts.IntentElementTerminated
		case RecordCreated:
			return bpmn_asserts.IntentElementActivating
		}
	}

	switch rt {
	case RecordCreated:
		return bpmn_asserts.IntentCreated
	case RecordActivated:
		return bpmn_asserts.IntentActivated
	case RecordCompleted:
		return bpmn_asserts.IntentCompleted
	case RecordTerminated:
		return bpmn_asserts.IntentCanceled
	case RecordCanceled:
		return bpmn_asserts.IntentCanceled
	case RecordFailed:
		return bpmn_asserts.IntentFailed
	case RecordResolved:
		return bpmn_asserts.IntentResolved
	case RecordTriggered:
		return bpmn_asserts.IntentTriggered
	case RecordCorrelated:
		return bpmn_asserts.IntentCorrelated
	case RecordUpdated:
		return bpmn_asserts.IntentUpdated
	case RecordTimedOut:
		return bpmn_asserts.IntentTimedOut
	case RecordDeployed:
		return bpmn_asserts.IntentCreated
	default:
		return bpmn_asserts.ZeebeIntent(string(rt))
	}
}

func buildValue(ev Event) json.RawMessage {
	var v any

	switch ev.ValueType {
	case ValueProcessInstance, ValueElementInstance:
		v = bpmn_asserts.ProcessInstanceValue{
			ProcessInstanceKey:   int64(ev.ProcessInstanceKey),
			ProcessDefinitionKey: int64(ev.ProcessDefinitionKey),
			ElementID:            ev.ElementId,
			BpmnElementType:      ev.ElementType,
		}
	case ValueVariable:
		// Try to extract name/value from ev.Value
		var varData struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		}
		if ev.Value != nil {
			_ = json.Unmarshal(ev.Value, &varData)
		}
		v = bpmn_asserts.VariableValue{
			Name:               varData.Name,
			Value:              varData.Value,
			ProcessInstanceKey: int64(ev.ProcessInstanceKey),
		}
	case ValueJob:
		var jobData struct {
			Type    string `json:"type"`
			Worker  string `json:"worker"`
			Retries int32  `json:"retries"`
		}
		if ev.Value != nil {
			_ = json.Unmarshal(ev.Value, &jobData)
		}
		v = bpmn_asserts.JobValue{
			Type:               jobData.Type,
			Worker:             jobData.Worker,
			Retries:            jobData.Retries,
			ElementID:          ev.ElementId,
			ProcessInstanceKey: int64(ev.ProcessInstanceKey),
		}
	case ValueIncident:
		var incData struct {
			ErrorType    string `json:"errorType"`
			ErrorMessage string `json:"errorMessage"`
		}
		if ev.Value != nil {
			_ = json.Unmarshal(ev.Value, &incData)
		}
		v = bpmn_asserts.IncidentValue{
			ErrorType:          incData.ErrorType,
			ErrorMessage:       incData.ErrorMessage,
			ElementID:          ev.ElementId,
			ProcessInstanceKey: int64(ev.ProcessInstanceKey),
		}
	default:
		if ev.Value != nil {
			return ev.Value
		}
		v = map[string]any{}
	}

	raw, err := json.Marshal(v)
	if err != nil {
		return []byte(`{}`)
	}
	return raw
}

// EmitEvent is a helper to create and emit an export event.
// Used by the processor after successful intent processing.
func EmitEvent(c *Collector, vt ValueType, rt RecordType, key, piKey, pdKey uint64, elemId, elemType string, value []byte) {
	c.Record(vt, rt, key, piKey, pdKey, elemId, elemType, value)
}

// EmitElementEvent is a shorthand for element instance events.
func EmitElementEvent(c *Collector, rt RecordType, key, piKey, pdKey uint64, elemId, elemType string) {
	c.Record(ValueElementInstance, rt, key, piKey, pdKey, elemId, elemType, nil)
}

// EmitProcessInstanceEvent is a shorthand for process instance events.
func EmitProcessInstanceEvent(c *Collector, rt RecordType, key, piKey, pdKey uint64) {
	c.Record(ValueProcessInstance, rt, key, piKey, pdKey, "", "process", nil)
}

// EmitJobEvent is a shorthand for job events.
func EmitJobEvent(c *Collector, rt RecordType, key, piKey, pdKey uint64, elemId string, jobType string) {
	value, _ := json.Marshal(map[string]any{"type": jobType})
	c.Record(ValueJob, rt, key, piKey, pdKey, elemId, "", value)
}

// Clock is used to allow deterministic timestamps in tests.
var Clock func() time.Time
