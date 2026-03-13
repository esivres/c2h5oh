package bpmn_asserts

import (
	"encoding/json"
	"testing"
)

// InspectedProcessInstance represents a process instance found through inspection.
type InspectedProcessInstance struct {
	ProcessInstanceKey int64
}

// --- ProcessEventInspections ---

// ProcessEventInspections allows finding processes started without explicit client commands
// (e.g., via timers, call activities). Works with PROCESS_EVENT records.
type ProcessEventInspections struct {
	stream               *RecordStream
	timerID              string
	processDefinitionKey int64
}

// FindProcessEvents creates a new ProcessEventInspections from the record stream.
func FindProcessEvents(stream *RecordStream) *ProcessEventInspections {
	return &ProcessEventInspections{stream: stream}
}

// TriggeredByTimer filters to process events triggered by the specified timer element.
func (i *ProcessEventInspections) TriggeredByTimer(timerID string) *ProcessEventInspections {
	return &ProcessEventInspections{
		stream:               i.stream,
		timerID:              timerID,
		processDefinitionKey: i.processDefinitionKey,
	}
}

// WithProcessDefinitionKey filters to process events with the specified process definition key.
func (i *ProcessEventInspections) WithProcessDefinitionKey(key int64) *ProcessEventInspections {
	return &ProcessEventInspections{
		stream:               i.stream,
		timerID:              i.timerID,
		processDefinitionKey: key,
	}
}

// FindFirstProcessInstance returns the first matching process instance, or nil if none found.
func (i *ProcessEventInspections) FindFirstProcessInstance() *InspectedProcessInstance {
	keys := i.processInstanceKeys()
	if len(keys) == 0 {
		return nil
	}
	return &InspectedProcessInstance{ProcessInstanceKey: keys[0]}
}

// FindLastProcessInstance returns the last matching process instance, or nil if none found.
func (i *ProcessEventInspections) FindLastProcessInstance() *InspectedProcessInstance {
	keys := i.processInstanceKeys()
	if len(keys) == 0 {
		return nil
	}
	return &InspectedProcessInstance{ProcessInstanceKey: keys[len(keys)-1]}
}

// FindProcessInstance returns the process instance at the specified index (0-based), or nil.
func (i *ProcessEventInspections) FindProcessInstance(index int) *InspectedProcessInstance {
	keys := i.processInstanceKeys()
	if index < 0 || index >= len(keys) {
		return nil
	}
	return &InspectedProcessInstance{ProcessInstanceKey: keys[index]}
}

func (i *ProcessEventInspections) processInstanceKeys() []int64 {
	var keys []int64
	seen := make(map[int64]struct{})

	if i.timerID != "" {
		timerRecords := i.stream.Filter(func(r Record) bool {
			if r.ValueType != ValueTypeTimer || r.Intent != IntentTriggered || r.RecordType != RecordTypeEvent {
				return false
			}
			var v TimerValue
			if json.Unmarshal(r.Value, &v) != nil {
				return false
			}
			match := v.TargetElementID == i.timerID
			if i.processDefinitionKey > 0 {
				match = match && v.ProcessDefinitionKey == i.processDefinitionKey
			}
			return match
		})
		for _, r := range timerRecords {
			var v TimerValue
			if json.Unmarshal(r.Value, &v) == nil && v.ProcessInstanceKey > 0 {
				if _, ok := seen[v.ProcessInstanceKey]; !ok {
					seen[v.ProcessInstanceKey] = struct{}{}
					keys = append(keys, v.ProcessInstanceKey)
				}
			}
		}
		return keys
	}

	records := i.stream.Filter(func(r Record) bool {
		if r.ValueType != ValueTypeProcessInstance || r.RecordType != RecordTypeEvent || r.Intent != IntentElementActivated {
			return false
		}
		var v ProcessInstanceValue
		if json.Unmarshal(r.Value, &v) != nil {
			return false
		}
		if v.BpmnElementType != string(ValueTypeProcess) {
			return false
		}
		if i.processDefinitionKey > 0 && v.ProcessDefinitionKey != i.processDefinitionKey {
			return false
		}
		return true
	})
	for _, r := range records {
		var v ProcessInstanceValue
		if json.Unmarshal(r.Value, &v) == nil {
			if _, ok := seen[v.ProcessInstanceKey]; !ok {
				seen[v.ProcessInstanceKey] = struct{}{}
				keys = append(keys, v.ProcessInstanceKey)
			}
		}
	}
	return keys
}

// --- ProcessInstanceInspections ---

// ProcessInstanceInspections allows finding child process instances
// started by other processes (via call activities).
type ProcessInstanceInspections struct {
	stream                   *RecordStream
	bpmnProcessID            string
	parentProcessInstanceKey int64
}

// FindProcessInstances creates a new ProcessInstanceInspections from the record stream.
func FindProcessInstances(stream *RecordStream) *ProcessInstanceInspections {
	return &ProcessInstanceInspections{stream: stream}
}

// WithParentProcessInstanceKey filters to process instances with the specified parent.
func (i *ProcessInstanceInspections) WithParentProcessInstanceKey(key int64) *ProcessInstanceInspections {
	return &ProcessInstanceInspections{
		stream:                   i.stream,
		parentProcessInstanceKey: key,
		bpmnProcessID:            i.bpmnProcessID,
	}
}

// WithBpmnProcessID filters to process instances with the specified BPMN process ID.
func (i *ProcessInstanceInspections) WithBpmnProcessID(bpmnProcessID string) *ProcessInstanceInspections {
	return &ProcessInstanceInspections{
		stream:                   i.stream,
		parentProcessInstanceKey: i.parentProcessInstanceKey,
		bpmnProcessID:            bpmnProcessID,
	}
}

// FindFirstProcessInstance returns the first matching process instance, or nil if none found.
func (i *ProcessInstanceInspections) FindFirstProcessInstance() *InspectedProcessInstance {
	keys := i.processInstanceKeys()
	if len(keys) == 0 {
		return nil
	}
	return &InspectedProcessInstance{ProcessInstanceKey: keys[0]}
}

// FindLastProcessInstance returns the last matching process instance, or nil if none found.
func (i *ProcessInstanceInspections) FindLastProcessInstance() *InspectedProcessInstance {
	keys := i.processInstanceKeys()
	if len(keys) == 0 {
		return nil
	}
	return &InspectedProcessInstance{ProcessInstanceKey: keys[len(keys)-1]}
}

// FindProcessInstance returns the process instance at the specified index (0-based), or nil.
func (i *ProcessInstanceInspections) FindProcessInstance(index int) *InspectedProcessInstance {
	keys := i.processInstanceKeys()
	if index < 0 || index >= len(keys) {
		return nil
	}
	return &InspectedProcessInstance{ProcessInstanceKey: keys[index]}
}

func (i *ProcessInstanceInspections) processInstanceKeys() []int64 {
	var keys []int64
	seen := make(map[int64]struct{})

	records := i.stream.Filter(func(r Record) bool {
		if r.ValueType != ValueTypeProcessInstance || r.RecordType != RecordTypeEvent || r.Intent != IntentElementActivated {
			return false
		}
		var v ProcessInstanceValue
		if json.Unmarshal(r.Value, &v) != nil {
			return false
		}
		if v.BpmnElementType != string(ValueTypeProcess) {
			return false
		}
		if i.parentProcessInstanceKey > 0 && v.ParentProcessInstanceKey != i.parentProcessInstanceKey {
			return false
		}
		if i.bpmnProcessID != "" && v.BpmnProcessID != i.bpmnProcessID {
			return false
		}
		return true
	})

	for _, r := range records {
		var v ProcessInstanceValue
		if json.Unmarshal(r.Value, &v) == nil {
			if _, ok := seen[v.ProcessInstanceKey]; !ok {
				seen[v.ProcessInstanceKey] = struct{}{}
				keys = append(keys, v.ProcessInstanceKey)
			}
		}
	}
	return keys
}

// --- Convenience: assert on InspectedProcessInstance ---

// AssertThat creates a ProcessInstanceAssert for the inspected process instance.
func (p *InspectedProcessInstance) AssertThat(t testing.TB, stream *RecordStream) *ProcessInstanceAssert {
	t.Helper()
	return ForProcessInstance(t, stream, p.ProcessInstanceKey)
}
