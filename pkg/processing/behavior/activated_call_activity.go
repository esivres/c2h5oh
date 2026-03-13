package behavior

import (
	"context"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// activatedCallActivity handles ElementActivatedIntent for callActivity.
// Finds the called process definition and creates a child process instance.
func activatedCallActivity(ctx context.Context, s storage.Store, i *intent.ElementActivatedIntent) ([]intent.Intent, error) {
	if i.ProcessDefinitionKey == 0 {
		return nil, fmt.Errorf("call activity %q: no process definition key", i.ElementId)
	}

	bmi, err := loadBPMN(ctx, s, i.ProcessDefinitionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load BPMN for call activity: %w", err)
	}

	fn := findFlowNode(bmi, i.ElementId)
	if fn == nil {
		return nil, fmt.Errorf("call activity element %q not found in BPMN model", i.ElementId)
	}

	be, ok := fn.(bpmn_model.BaseElement)
	if !ok {
		return nil, fmt.Errorf("call activity %q: not a BaseElement", i.ElementId)
	}

	ce, found := bpmn_model.GetSingleExtensionElement[bpmn_model.ZeebeCalledElement](be)
	if !found {
		return nil, fmt.Errorf("call activity %q: no ZeebeCalledElement", i.ElementId)
	}

	calledProcessId := ce.GetProcessId()
	if calledProcessId == "" {
		return nil, fmt.Errorf("call activity %q: empty processId", i.ElementId)
	}

	calledDef, err := s.ProcessDefinitions().FindLatestByProcessId(ctx, calledProcessId)
	if err != nil {
		return nil, err
	}
	if calledDef == nil {
		return nil, fmt.Errorf("called process %q not found", calledProcessId)
	}

	return []intent.Intent{
		&intent.CreateProcessInstanceIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: i.ProcessInstanceKey,
			},
			ProcessDefinitionKey: calledDef.Key,
			BpmnProcessId:        calledProcessId,
			ParentKey:            i.ProcessInstanceKey,
			ParentElementKey:     i.ElementInstanceKey,
		},
	}, nil
}
