package behavior

import (
	"context"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// activatedJobTask handles ElementActivatedIntent for job-based tasks:
// serviceTask, userTask, scriptTask, businessRuleTask, sendTask.
// Creates a job for external workers.
func activatedJobTask(_ context.Context, _ storage.Store, i *intent.ElementActivatedIntent) ([]intent.Intent, error) {
	jobType := i.JobType
	if jobType == "" && i.ElementType == "userTask" {
		jobType = "io.camunda.zeebe:userTask"
	}
	return []intent.Intent{
		&intent.CreateJobIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: i.ProcessInstanceKey,
			},
			ElementInstanceKey:   i.ElementInstanceKey,
			ProcessDefinitionKey: i.ProcessDefinitionKey,
			Type:                 jobType,
			Retries:              3,
		},
	}, nil
}
