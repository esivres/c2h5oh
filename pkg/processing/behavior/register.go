package behavior

import "github.com/esivres/c2h5oh/pkg/processing/intent"

// DefaultRegistry creates a registry with all standard BPMN engine behaviors.
func DefaultRegistry() *Registry {
	r := NewRegistry()

	// Process definition
	r.Register(intent.DeployProcess, Typed(deployProcess))

	// Process instance
	r.Register(intent.CreateProcessInstance, Typed(createProcessInstance))
	r.Register(intent.CompleteProcessInstance, Typed(completeProcessInstance))
	r.Register(intent.CancelProcessInstance, Typed(cancelProcessInstance))

	// Element instance — phase 1 (generic)
	r.Register(intent.ActivateElement, Typed(activateElement))
	r.Register(intent.CompleteElement, Typed(completeElement))
	r.Register(intent.TerminateElement, Typed(terminateElement))

	// Element completed — phase 2 (type-specific)
	r.Register(intent.ElementCompleted, Typed(completedDefault))
	r.RegisterWithElementType(intent.ElementCompleted, "exclusiveGateway", Typed(completedExclusiveGateway))
	r.RegisterWithElementType(intent.ElementCompleted, "inclusiveGateway", Typed(completedInclusiveGateway))
	r.RegisterWithElementType(intent.ElementCompleted, "endEvent", Typed(completedEndEvent))
	r.RegisterWithElementType(intent.ElementCompleted, "boundaryEvent", Typed(completedBoundaryEvent))
	r.RegisterWithElementType(intent.ElementCompleted, "intermediateCatchEvent", Typed(completedCatchEvent))
	r.RegisterWithElementType(intent.ElementCompleted, "intermediateThrowEvent", Typed(completedThrowEvent))

	// Element activated — phase 2 (type-specific)
	r.Register(intent.ElementActivated, Typed(activatedAutoComplete)) // default fallback
	r.RegisterWithElementType(intent.ElementActivated, "serviceTask", Typed(activatedJobTask))
	r.RegisterWithElementType(intent.ElementActivated, "userTask", Typed(activatedJobTask))
	r.RegisterWithElementType(intent.ElementActivated, "scriptTask", Typed(activatedJobTask))
	r.RegisterWithElementType(intent.ElementActivated, "businessRuleTask", Typed(activatedJobTask))
	r.RegisterWithElementType(intent.ElementActivated, "sendTask", Typed(activatedJobTask))
	r.RegisterWithElementType(intent.ElementActivated, "receiveTask", Typed(activatedReceiveTask))
	r.RegisterWithElementType(intent.ElementActivated, "intermediateCatchEvent", Typed(activatedCatchEvent))
	r.RegisterWithElementType(intent.ElementActivated, "boundaryEvent", Typed(activatedCatchEvent))
	r.RegisterWithElementType(intent.ElementActivated, "intermediateThrowEvent", Typed(activatedThrowEvent))
	r.RegisterWithElementType(intent.ElementActivated, "parallelGateway", Typed(activatedParallelGateway))
	r.RegisterWithElementType(intent.ElementActivated, "inclusiveGateway", Typed(activatedParallelGateway))
	r.RegisterWithElementType(intent.ElementActivated, "subProcess", Typed(activatedSubProcess))
	r.RegisterWithElementType(intent.ElementActivated, "callActivity", Typed(activatedCallActivity))
	r.RegisterWithElementType(intent.ElementActivated, "adHocSubProcess", Typed(activatedAdHocSubProcess))

	// Sequence flow
	r.Register(intent.TakeSequenceFlow, Typed(takeSequenceFlow))

	// Job
	r.Register(intent.CreateJob, Typed(createJob))
	r.Register(intent.ActivateJob, Typed(activateJob))
	r.Register(intent.CompleteJob, Typed(completeJob))
	r.Register(intent.FailJob, Typed(failJob))
	r.Register(intent.ThrowJobError, Typed(throwJobError))
	r.Register(intent.TimeOutJob, Typed(timeoutJob))
	r.Register(intent.UpdateJobRetries, Typed(updateJobRetries))
	r.Register(intent.UpdateJobTimeout, Typed(updateJobTimeout))

	// Variables
	r.Register(intent.SetVariables, Typed(setVariables))

	// Incident
	r.Register(intent.CreateIncident, Typed(createIncident))
	r.Register(intent.ResolveIncident, Typed(resolveIncident))

	// Timer
	r.Register(intent.CreateTimer, Typed(createTimer))
	r.Register(intent.TriggerTimer, Typed(triggerTimer))
	r.Register(intent.CancelTimer, Typed(cancelTimer))

	// Signal
	r.Register(intent.ThrowSignal, Typed(throwSignal))

	// Resource
	r.Register(intent.DeleteResource, Typed(deleteResource))

	// Form
	r.Register(intent.DeployForm, Typed(deployForm))

	// Message correlation
	r.Register(intent.PublishMessage, Typed(publishMessage))
	r.Register(intent.OpenSubscription, Typed(openSubscription))
	r.Register(intent.CorrelateMessage, Typed(correlateMessage))
	r.Register(intent.CloseSubscription, Typed(closeSubscription))

	return r
}
