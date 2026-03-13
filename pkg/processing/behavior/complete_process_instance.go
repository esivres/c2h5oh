package behavior

import (
	"context"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// completeProcessInstance handles CompleteProcessInstanceIntent.
func completeProcessInstance(ctx context.Context, s storage.Store, i *intent.CompleteProcessInstanceIntent) ([]intent.Intent, error) {
	// Get the process instance to check for parent (call activity)
	pi, err := s.ProcessInstances().GetInstance(ctx, i.ProcessInstanceKey)
	if err != nil {
		return nil, err
	}

	if err := s.ProcessInstances().UpdateInstanceState(ctx, i.ProcessInstanceKey, storage.ProcessInstanceCompleted); err != nil {
		return nil, err
	}

	// If this is a child process (called from a call activity), complete the parent element
	if pi != nil && pi.ParentElementKey != 0 {
		return []intent.Intent{
			&intent.CompleteElementIntent{
				Header: intent.Header{
					Origin:             intent.Internal,
					ProcessInstanceKey: pi.ParentKey,
				},
				ElementInstanceKey: pi.ParentElementKey,
			},
		}, nil
	}

	return nil, nil
}
