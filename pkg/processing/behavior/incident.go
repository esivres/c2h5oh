package behavior

import (
	"context"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// createIncident handles CreateIncidentIntent.
func createIncident(ctx context.Context, s storage.Store, i *intent.CreateIncidentIntent) ([]intent.Intent, error) {
	inc := &storage.Incident{
		Key:                i.Key,
		ProcessInstanceKey: i.ProcessInstanceKey,
		ElementInstanceKey: i.ElementInstanceKey,
		JobKey:             i.JobKey,
		Type:               storage.IncidentType(i.ErrorType),
		State:              storage.IncidentCreated,
		ErrorMessage:       i.ErrorMessage,
		CreatedAt:          time.Now(),
	}
	return nil, s.Incidents().Create(ctx, inc)
}

// resolveIncident handles ResolveIncidentIntent.
func resolveIncident(ctx context.Context, s storage.Store, i *intent.ResolveIncidentIntent) ([]intent.Intent, error) {
	return nil, s.Incidents().Resolve(ctx, i.IncidentKey)
}
