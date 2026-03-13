package behavior

import (
	"context"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// activatedAdHocSubProcess handles ElementActivatedIntent for adHocSubProcess.
// Evaluates activeElementsCollection and activates requested elements.
func activatedAdHocSubProcess(ctx context.Context, s storage.Store, i *intent.ElementActivatedIntent) ([]intent.Intent, error) {
	if i.ProcessDefinitionKey == 0 {
		return nil, nil // wait
	}

	bmi, err := loadBPMN(ctx, s, i.ProcessDefinitionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load BPMN for ad-hoc subprocess: %w", err)
	}

	ah := findAdHocSubProcess(bmi, i.ElementId)
	if ah == nil {
		return nil, nil
	}

	return activateAdHocElements(ctx, s, bmi, ah, i.FlowScopeKey, i.ElementInstanceKey, i.ProcessInstanceKey, i.ProcessDefinitionKey)
}
