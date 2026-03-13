package behavior

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// setVariables handles SetVariablesIntent.
//
// Merges provided variables into the target scope.
func setVariables(ctx context.Context, s storage.Store, i *intent.SetVariablesIntent) ([]intent.Intent, error) {
	varRepo := s.Variables()

	var vars map[string]json.RawMessage
	if err := json.Unmarshal(i.Variables, &vars); err != nil {
		return nil, fmt.Errorf("invalid variables JSON: %w", err)
	}

	for name, value := range vars {
		existing, err := varRepo.FindByName(ctx, i.ScopeKey, name)
		if err != nil {
			return nil, err
		}

		if existing != nil {
			if err := varRepo.Update(ctx, existing.Key, value); err != nil {
				return nil, err
			}
		} else {
			v := &storage.Variable{
				ProcessInstanceKey: i.ProcessInstanceKey,
				ScopeKey:           i.ScopeKey,
				Name:               name,
				Value:              value,
			}
			if err := varRepo.Create(ctx, v); err != nil {
				return nil, err
			}
		}
	}

	return nil, nil
}
