package behavior

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	xmlm "github.com/esivres/c2h5oh/pkg/bpmn_model/xml"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

const bpmnNS = "http://www.omg.org/spec/BPMN/20100524/MODEL"

// loadBPMN loads and parses the BPMN model from a process definition key.
func loadBPMN(ctx context.Context, s storage.Store, pdKey uint64) (*bpmn_model.BpmnModelInstance, error) {
	def, err := s.ProcessDefinitions().FindByKey(ctx, pdKey)
	if err != nil {
		return nil, err
	}
	if def == nil {
		return nil, fmt.Errorf("process definition not found: %d", pdKey)
	}
	return bpmn_model.ReadFromBytes(def.Content)
}

// timerSpec holds the parsed timer configuration.
type timerSpec struct {
	DueDate       time.Time
	Repetitions   int           // 0 = one-shot, -1 = infinite, N = count
	CycleDuration time.Duration // interval for cycle timers
}

// parseTimerSpec extracts timer configuration from a timerEventDefinition DOM element.
// Supports timeDuration, timeDate, and timeCycle (R[count]/duration).
func parseTimerSpec(timerDef *xmlm.Element, ns string) (*timerSpec, error) {
	// Check timeCycle child (e.g. "R3/PT10S" or "R/PT1M")
	cycles := timerDef.GetChildElementsByNS(ns, "timeCycle")
	if len(cycles) > 0 {
		expr := cycles[0].GetTextContent()
		reps, dur, err := parseTimerCycle(expr)
		if err != nil {
			return nil, err
		}
		return &timerSpec{
			DueDate:       time.Now().Add(dur),
			Repetitions:   reps,
			CycleDuration: dur,
		}, nil
	}

	// Check timeDuration child
	durations := timerDef.GetChildElementsByNS(ns, "timeDuration")
	if len(durations) > 0 {
		expr := durations[0].GetTextContent()
		d, err := parseISO8601Duration(expr)
		if err != nil {
			return nil, err
		}
		return &timerSpec{DueDate: time.Now().Add(d)}, nil
	}

	// Check timeDate child
	dates := timerDef.GetChildElementsByNS(ns, "timeDate")
	if len(dates) > 0 {
		expr := dates[0].GetTextContent()
		t, err := time.Parse(time.RFC3339, expr)
		if err != nil {
			return nil, fmt.Errorf("invalid timeDate %q: %w", expr, err)
		}
		return &timerSpec{DueDate: t}, nil
	}

	// Default: trigger immediately
	return &timerSpec{DueDate: time.Now()}, nil
}

// parseTimerCycle parses a timer cycle expression like "R3/PT10S" or "R/PT1M".
func parseTimerCycle(expr string) (int, time.Duration, error) {
	if len(expr) < 2 || expr[0] != 'R' {
		return 0, 0, fmt.Errorf("invalid timer cycle: %q", expr)
	}

	slashIdx := findChar(expr, '/')
	if slashIdx < 0 {
		return 0, 0, fmt.Errorf("invalid timer cycle (no slash): %q", expr)
	}

	repsStr := expr[1:slashIdx]
	var reps int
	if repsStr == "" {
		reps = -1
	} else {
		n, err := parseInt(repsStr)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid repetitions in timer cycle %q: %w", expr, err)
		}
		reps = int(n)
	}

	durStr := expr[slashIdx+1:]
	dur, err := parseISO8601Duration(durStr)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid duration in timer cycle %q: %w", expr, err)
	}

	return reps, dur, nil
}

// parseISO8601Duration parses a simple ISO 8601 duration (e.g. PT10S, PT1M, PT1H, P1D).
func parseISO8601Duration(s string) (time.Duration, error) {
	if len(s) < 3 || s[0] != 'P' {
		return 0, fmt.Errorf("invalid ISO 8601 duration: %q", s)
	}

	s = s[1:]
	var total time.Duration

	if idx := findChar(s, 'D'); idx >= 0 {
		days, err := parseInt(s[:idx])
		if err != nil {
			return 0, err
		}
		total += time.Duration(days) * 24 * time.Hour
		s = s[idx+1:]
	}

	if len(s) > 0 && s[0] == 'T' {
		s = s[1:]
		if idx := findChar(s, 'H'); idx >= 0 {
			hours, err := parseInt(s[:idx])
			if err != nil {
				return 0, err
			}
			total += time.Duration(hours) * time.Hour
			s = s[idx+1:]
		}
		if idx := findChar(s, 'M'); idx >= 0 {
			mins, err := parseInt(s[:idx])
			if err != nil {
				return 0, err
			}
			total += time.Duration(mins) * time.Minute
			s = s[idx+1:]
		}
		if idx := findChar(s, 'S'); idx >= 0 {
			secs, err := parseInt(s[:idx])
			if err != nil {
				return 0, err
			}
			total += time.Duration(secs) * time.Second
			s = s[idx+1:]
		}
	}

	return total, nil
}

func findChar(s string, c byte) int {
	for i := range s {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func parseInt(s string) (int64, error) {
	var n int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid number: %q", s)
		}
		n = n*10 + int64(c-'0')
	}
	return n, nil
}

// resolveMessageName finds the message name from the model by message ref id.
func resolveMessageName(bmi *bpmn_model.BpmnModelInstance, messageRef string) string {
	if messageRef == "" {
		return ""
	}
	messages := bpmn_model.GetTypedElements[bpmn_model.Message](bmi.ModelInstance)
	for _, m := range messages {
		if m.GetId() == messageRef {
			return m.GetName()
		}
	}
	return messageRef
}

// resolveSignalName finds the signal name from the model by signal ref id.
func resolveSignalName(bmi *bpmn_model.BpmnModelInstance, signalRef string) string {
	if signalRef == "" {
		return ""
	}
	signals := bpmn_model.GetTypedElements[bpmn_model.Signal](bmi.ModelInstance)
	for _, sig := range signals {
		if sig.GetId() == signalRef {
			return sig.GetName()
		}
	}
	return signalRef
}

// resolveCorrelationKey evaluates the correlation key expression from ZeebeSubscription.
func resolveCorrelationKey(ctx context.Context, s storage.Store, bmi *bpmn_model.BpmnModelInstance, messageRef string, scopeKey uint64) string {
	return ""
}

// resolveElementType maps a BPMN model FlowNode to its element type string.
func resolveElementType(node bpmn_model.FlowNode) string {
	switch node.(type) {
	case bpmn_model.StartEvent:
		return "startEvent"
	case bpmn_model.EndEvent:
		return "endEvent"
	case bpmn_model.ServiceTask:
		return "serviceTask"
	case bpmn_model.UserTask:
		return "userTask"
	case bpmn_model.ScriptTask:
		return "scriptTask"
	case bpmn_model.BusinessRuleTask:
		return "businessRuleTask"
	case bpmn_model.SendTask:
		return "sendTask"
	case bpmn_model.ReceiveTask:
		return "receiveTask"
	case bpmn_model.ManualTask:
		return "manualTask"
	case bpmn_model.ExclusiveGateway:
		return "exclusiveGateway"
	case bpmn_model.ParallelGateway:
		return "parallelGateway"
	case bpmn_model.InclusiveGateway:
		return "inclusiveGateway"
	case bpmn_model.EventBasedGateway:
		return "eventBasedGateway"
	case bpmn_model.AdHocSubProcess:
		return "adHocSubProcess"
	case bpmn_model.SubProcess:
		return "subProcess"
	case bpmn_model.CallActivity:
		return "callActivity"
	case bpmn_model.IntermediateCatchEvent:
		return "intermediateCatchEvent"
	case bpmn_model.IntermediateThrowEvent:
		return "intermediateThrowEvent"
	case bpmn_model.BoundaryEvent:
		return "boundaryEvent"
	default:
		return "unknown"
	}
}

// resolveJobType extracts the job type from ZeebeTaskDefinition extension.
func resolveJobType(bmi *bpmn_model.BpmnModelInstance, node bpmn_model.FlowNode) string {
	if be, ok := node.(bpmn_model.BaseElement); ok {
		td, found := bpmn_model.GetSingleExtensionElement[bpmn_model.ZeebeTaskDefinition](be)
		if found {
			return td.GetType()
		}
	}
	return ""
}

// activateTarget creates an ActivateElementIntent for a target flow node.
func activateTarget(piKey, pdKey, flowScopeKey uint64, target bpmn_model.FlowNode, bmi *bpmn_model.BpmnModelInstance) *intent.ActivateElementIntent {
	return &intent.ActivateElementIntent{
		Header: intent.Header{
			Origin:             intent.Internal,
			ProcessInstanceKey: piKey,
		},
		ProcessDefinitionKey: pdKey,
		ElementId:            target.GetId(),
		ElementType:          resolveElementType(target),
		FlowScopeKey:         flowScopeKey,
		JobType:              resolveJobType(bmi, target),
	}
}

// autoCompleteFromActivated creates a CompleteElementIntent from an ElementActivatedIntent.
func autoCompleteFromActivated(i *intent.ElementActivatedIntent) []intent.Intent {
	return []intent.Intent{
		&intent.CompleteElementIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: i.ProcessInstanceKey,
			},
			ElementInstanceKey: i.ElementInstanceKey,
		},
	}
}

// loadScope loads variables for a given scope key into a map.
func loadScope(ctx context.Context, s storage.Store, scopeKey uint64) (map[string]any, error) {
	vars, err := s.Variables().FindByScope(ctx, scopeKey)
	if err != nil {
		return nil, err
	}

	scope := make(map[string]any, len(vars))
	for _, v := range vars {
		var val any
		if err := json.Unmarshal(v.Value, &val); err != nil {
			scope[v.Name] = string(v.Value)
		} else {
			scope[v.Name] = val
		}
	}
	return scope, nil
}

// findFlowNode finds a FlowNode by element id in the BPMN model.
func findFlowNode(bmi *bpmn_model.BpmnModelInstance, elementId string) bpmn_model.FlowNode {
	flowNodes := bpmn_model.GetTypedElements[bpmn_model.FlowNode](bmi.ModelInstance)
	for _, fn := range flowNodes {
		if fn.GetId() == elementId {
			return fn
		}
	}
	return nil
}
