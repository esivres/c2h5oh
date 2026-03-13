package intent

// ActivateElementIntent requests activation of a BPMN element (creating a token).
type ActivateElementIntent struct {
	Header

	// ProcessDefinitionKey links to the process definition.
	ProcessDefinitionKey uint64

	// ElementId is the BPMN element id (e.g. "Activity_01ran10").
	ElementId string

	// ElementType is the BPMN element type (e.g. "serviceTask").
	ElementType string

	// FlowScopeKey is the element instance key of the enclosing scope.
	FlowScopeKey uint64

	// JobType is the job type for service/user tasks (empty for other elements).
	JobType string

	// Multi-instance fields (set when this is a child of a multi-instance body)
	MIInputVariable string // variable name for the current item
	MIInputValue    []byte // JSON-encoded value for the current item
	MIIndex         int    // index in the collection (0-based)
}

func (i *ActivateElementIntent) IntentType() Type { return ActivateElement }

// ElementActivatedIntent is emitted after an element instance has been created and
// input mappings applied. Type-specific behaviors are dispatched based on ElementType.
type ElementActivatedIntent struct {
	Header

	// ElementInstanceKey is the key of the activated element instance.
	ElementInstanceKey uint64

	// ProcessDefinitionKey links to the process definition.
	ProcessDefinitionKey uint64

	// ElementId is the BPMN element id.
	ElementId string

	// ElementType is the BPMN element type (used for secondary dispatch).
	ElementType string

	// FlowScopeKey is the element instance key of the enclosing scope.
	FlowScopeKey uint64

	// JobType is the pre-resolved job type (for service/user tasks).
	JobType string
}

func (i *ElementActivatedIntent) IntentType() Type       { return ElementActivated }
func (i *ElementActivatedIntent) GetElementType() string { return i.ElementType }

// CompleteElementIntent signals that an element instance has completed.
type CompleteElementIntent struct {
	Header

	// ElementInstanceKey is the key of the element instance to complete.
	ElementInstanceKey uint64

	// Variables is optional output variables (JSON bytes).
	Variables []byte
}

func (i *CompleteElementIntent) IntentType() Type { return CompleteElement }

// ElementCompletedIntent is emitted after an element instance state has been set to Completed
// and output mappings applied. Type-specific behaviors are dispatched based on ElementType.
type ElementCompletedIntent struct {
	Header

	// ElementInstanceKey is the key of the completed element instance.
	ElementInstanceKey uint64

	// ProcessDefinitionKey links to the process definition.
	ProcessDefinitionKey uint64

	// ElementId is the BPMN element id.
	ElementId string

	// ElementType is the BPMN element type (used for secondary dispatch).
	ElementType string

	// FlowScopeKey is the element instance key of the enclosing scope.
	FlowScopeKey uint64
}

func (i *ElementCompletedIntent) IntentType() Type       { return ElementCompleted }
func (i *ElementCompletedIntent) GetElementType() string { return i.ElementType }

// TerminateElementIntent requests termination of an element instance.
type TerminateElementIntent struct {
	Header

	// ElementInstanceKey is the key of the element instance to terminate.
	ElementInstanceKey uint64
}

func (i *TerminateElementIntent) IntentType() Type { return TerminateElement }

// TakeSequenceFlowIntent requests traversal of a sequence flow to its target element.
type TakeSequenceFlowIntent struct {
	Header

	// ProcessDefinitionKey links to the process definition.
	ProcessDefinitionKey uint64

	// SequenceFlowId is the BPMN sequence flow id.
	SequenceFlowId string

	// SourceElementId is the element that the flow originates from.
	SourceElementId string

	// TargetElementId is the element that the flow leads to.
	TargetElementId string

	// ConditionExpression is the FEEL expression to evaluate (empty if unconditional).
	ConditionExpression string

	// FlowScopeKey is the element instance key of the enclosing scope.
	FlowScopeKey uint64
}

func (i *TakeSequenceFlowIntent) IntentType() Type { return TakeSequenceFlow }
