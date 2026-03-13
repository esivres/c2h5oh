package intent

// ActivateElementIntent requests activation of a BPMN element (creating a token).
type ActivateElementIntent struct {
	ElementId       string
	ElementType     string
	JobType         string
	MIInputVariable string
	MIInputValue    []byte
	Header
	ProcessDefinitionKey uint64
	FlowScopeKey         uint64
	MIIndex              int
}

func (i *ActivateElementIntent) IntentType() Type { return ActivateElement }

// ElementActivatedIntent is emitted after an element instance has been created and
// input mappings applied. Type-specific behaviors are dispatched based on ElementType.
type ElementActivatedIntent struct {
	ElementId   string
	ElementType string
	JobType     string
	Header
	ElementInstanceKey   uint64
	ProcessDefinitionKey uint64
	FlowScopeKey         uint64
}

func (i *ElementActivatedIntent) IntentType() Type       { return ElementActivated }
func (i *ElementActivatedIntent) GetElementType() string { return i.ElementType }

// CompleteElementIntent signals that an element instance has completed.
type CompleteElementIntent struct {
	Variables []byte
	Header
	ElementInstanceKey uint64
}

func (i *CompleteElementIntent) IntentType() Type { return CompleteElement }

// ElementCompletedIntent is emitted after an element instance state has been set to Completed
// and output mappings applied. Type-specific behaviors are dispatched based on ElementType.
type ElementCompletedIntent struct {
	ElementId   string
	ElementType string
	Header
	ElementInstanceKey   uint64
	ProcessDefinitionKey uint64
	FlowScopeKey         uint64
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
	SequenceFlowId      string
	SourceElementId     string
	TargetElementId     string
	ConditionExpression string
	Header
	ProcessDefinitionKey uint64
	FlowScopeKey         uint64
}

func (i *TakeSequenceFlowIntent) IntentType() Type { return TakeSequenceFlow }
