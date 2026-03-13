package intent

// DeployProcessIntent requests deployment of a BPMN process definition.
type DeployProcessIntent struct {
	BpmnProcessId string
	Name          string
	Content       []byte
	ContentHash   []byte
	Header
}

func (i *DeployProcessIntent) IntentType() Type { return DeployProcess }

// CreateProcessInstanceIntent requests creation and start of a new process instance.
type CreateProcessInstanceIntent struct {
	BpmnProcessId string
	Variables     []byte
	Header
	ProcessDefinitionKey uint64
	Version              uint64
	ParentKey            uint64
	ParentElementKey     uint64
}

func (i *CreateProcessInstanceIntent) IntentType() Type { return CreateProcessInstance }

// CancelProcessInstanceIntent requests cancellation of a running process instance.
type CancelProcessInstanceIntent struct {
	Header
}

func (i *CancelProcessInstanceIntent) IntentType() Type { return CancelProcessInstance }

// CompleteProcessInstanceIntent signals that a process instance has completed normally.
type CompleteProcessInstanceIntent struct {
	Header
}

func (i *CompleteProcessInstanceIntent) IntentType() Type { return CompleteProcessInstance }
