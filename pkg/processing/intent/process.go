package intent

// DeployProcessIntent requests deployment of a BPMN process definition.
type DeployProcessIntent struct {
	Header

	// BpmnProcessId is the BPMN process id attribute.
	BpmnProcessId string

	// Name is the human-readable process name.
	Name string

	// Content is the raw BPMN XML.
	Content []byte

	// ContentHash is the SHA-256 hash for deduplication.
	ContentHash []byte
}

func (i *DeployProcessIntent) IntentType() Type { return DeployProcess }

// CreateProcessInstanceIntent requests creation and start of a new process instance.
type CreateProcessInstanceIntent struct {
	Header

	// ProcessDefinitionKey is the key of the deployed process definition.
	// If 0, lookup by BpmnProcessId + Version.
	ProcessDefinitionKey uint64

	// BpmnProcessId is the process id for lookup (used if ProcessDefinitionKey == 0).
	BpmnProcessId string

	// Version is the specific version to use (0 = latest).
	Version uint64

	// Variables is the initial variables (JSON bytes).
	Variables []byte

	// ParentKey is the parent process instance key (0 if top-level).
	ParentKey uint64

	// ParentElementKey is the call activity element instance key in the parent.
	ParentElementKey uint64
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
