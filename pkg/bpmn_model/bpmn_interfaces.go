package bpmn_model

// BpmnModelElementInstance is the base interface for all BPMN model elements.
type BpmnModelElementInstance interface {
	ModelElementInstance
	// BpmnElementType returns the BPMN XML element name (e.g. "startEvent", "process").
	BpmnElementType() string
}

// Definitions is the root element of a BPMN model.
type Definitions interface {
	BpmnModelElementInstance
	GetId() string
	SetId(id string)
	GetName() string
	SetName(name string)
	GetTargetNamespace() string
	SetTargetNamespace(ns string)
	GetExporter() string
	SetExporter(exporter string)
	GetExporterVersion() string
	SetExporterVersion(version string)
	GetRootElements() []ModelElementInstance
}

// BaseElement is the abstract base for most BPMN elements.
type BaseElement interface {
	BpmnModelElementInstance
	GetId() string
	SetId(id string)
	GetExtensionElements() ExtensionElements
	GetOrCreateExtensionElements() ExtensionElements
}

// RootElement is a top-level element within Definitions.
type RootElement interface {
	BaseElement
}

// CallableElement is an element that can be called (Process, GlobalTask, etc.).
type CallableElement interface {
	RootElement
	GetName() string
	SetName(name string)
}

// FlowElement is an element within a Process flow.
type FlowElement interface {
	BaseElement
	GetName() string
	SetName(name string)
}

// FlowNode is an element that can have incoming and outgoing sequence flows.
type FlowNode interface {
	FlowElement
	GetIncomingSequenceFlows() []SequenceFlow
	GetOutgoingSequenceFlows() []SequenceFlow
}

// Event is the abstract base for all events.
type Event interface {
	FlowNode
}

// CatchEvent is an event that catches a trigger.
type CatchEvent interface {
	Event
	implCatchEvent()
}

// ThrowEvent is an event that throws a trigger.
type ThrowEvent interface {
	Event
	implThrowEvent()
}

// StartEvent is the starting point of a process.
type StartEvent interface {
	CatchEvent
	IsInterrupting() bool
	SetIsInterrupting(v bool)
	implStartEvent()
}

// EndEvent is the end point of a process.
type EndEvent interface {
	ThrowEvent
	implEndEvent()
}

// SequenceFlow connects two FlowNodes.
type SequenceFlow interface {
	FlowElement
	implSequenceFlow()
	GetSourceRef() string
	SetSourceRef(ref string)
	GetTargetRef() string
	SetTargetRef(ref string)
	GetSource() FlowNode
	GetTarget() FlowNode
	GetConditionExpression() ConditionExpression
	SetConditionExpression(expr ConditionExpression)
}

// ConditionExpression defines a condition on a SequenceFlow.
type ConditionExpression interface {
	BpmnModelElementInstance
	GetTextContent() string
	SetTextContent(text string)
	GetType() string
	SetType(t string)
}

// Process is a BPMN process definition.
type Process interface {
	CallableElement
	implProcess()
	GetProcessType() string
	SetProcessType(pt string)
	IsExecutable() bool
	SetIsExecutable(exec bool)
	IsClosed() bool
	SetIsClosed(closed bool)
	GetFlowElements() []FlowElement
	AddFlowElement(elem FlowElement)
}

// ExtensionElements is a container for extension elements within a BaseElement.
type ExtensionElements interface {
	BpmnModelElementInstance
	GetElements() []ModelElementInstance
}
