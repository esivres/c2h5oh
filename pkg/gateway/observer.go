package gateway

import (
	"github.com/esivres/c2h5oh/pkg/processing/intent"
)

// JobCreatedObserver returns an IntentObserver that notifies the JobNotifier
// whenever a CreateJobIntent is successfully processed.
// Pass the returned function to engine.Config.Observers.
func JobCreatedObserver(n *JobNotifier) func(intent.Intent) {
	return func(i intent.Intent) {
		if cj, ok := i.(*intent.CreateJobIntent); ok {
			n.Notify(cj.Type)
		}
	}
}

// DeployedObserver returns an IntentObserver that notifies the DeployNotifier
// whenever a DeployProcessIntent is successfully processed.
func DeployedObserver(n *DeployNotifier) func(intent.Intent) {
	return func(i intent.Intent) {
		if dp, ok := i.(*intent.DeployProcessIntent); ok {
			n.Notify(dp.BpmnProcessId)
		}
	}
}

// ProcessCompletedObserver returns an IntentObserver that notifies the CompletionNotifier
// whenever a CompleteProcessInstanceIntent is successfully processed.
func ProcessCompletedObserver(n *CompletionNotifier) func(intent.Intent) {
	return func(i intent.Intent) {
		if cp, ok := i.(*intent.CompleteProcessInstanceIntent); ok {
			n.Notify(cp.ProcessInstanceKey)
		}
	}
}
