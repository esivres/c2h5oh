package bpmn_model

// --- Interfaces ---

// IntermediateCatchEvent catches a trigger during process execution.
type IntermediateCatchEvent interface {
	CatchEvent
	implIntermediateCatchEvent()
}

// IntermediateThrowEvent throws a trigger during process execution.
type IntermediateThrowEvent interface {
	ThrowEvent
	implIntermediateThrowEvent()
}

// BoundaryEvent is attached to an activity and catches triggers.
type BoundaryEvent interface {
	CatchEvent
	implBoundaryEvent()
	IsCancelActivity() bool
	SetCancelActivity(v bool)
	GetAttachedToRef() string
	SetAttachedToRef(ref string)
	GetAttachedTo() Activity
}

// EventDefinition is the abstract base for event definitions.
type EventDefinition interface {
	RootElement
	implEventDefinition()
}

// MessageEventDefinition defines a message-based event.
type MessageEventDefinition interface {
	EventDefinition
	implMessageEventDefinition()
	GetMessageRef() string
	SetMessageRef(ref string)
}

// TimerEventDefinition defines a timer-based event.
type TimerEventDefinition interface {
	EventDefinition
	implTimerEventDefinition()
}

// SignalEventDefinition defines a signal-based event.
type SignalEventDefinition interface {
	EventDefinition
	implSignalEventDefinition()
	GetSignalRef() string
	SetSignalRef(ref string)
}

// ErrorEventDefinition defines an error-based event.
type ErrorEventDefinition interface {
	EventDefinition
	implErrorEventDefinition()
	GetErrorRef() string
	SetErrorRef(ref string)
}

// EscalationEventDefinition defines an escalation-based event.
type EscalationEventDefinition interface {
	EventDefinition
	implEscalationEventDefinition()
	GetEscalationRef() string
	SetEscalationRef(ref string)
}

// ConditionalEventDefinition defines a condition-based event.
type ConditionalEventDefinition interface {
	EventDefinition
	implConditionalEventDefinition()
}

// LinkEventDefinition defines a link event (go-to).
type LinkEventDefinition interface {
	EventDefinition
	implLinkEventDefinition()
	GetLinkName() string
	SetLinkName(name string)
}

// CompensateEventDefinition triggers compensation.
type CompensateEventDefinition interface {
	EventDefinition
	implCompensateEventDefinition()
}

// CancelEventDefinition cancels a transaction.
type CancelEventDefinition interface {
	EventDefinition
	implCancelEventDefinition()
}

// TerminateEventDefinition terminates all activities.
type TerminateEventDefinition interface {
	EventDefinition
	implTerminateEventDefinition()
}

// --- Attribute descriptors ---

var (
	AttrBoundaryEventCancelActivity *Attribute[bool]
	AttrBoundaryEventAttachedToRef  *Attribute[string]
	AttrMessageEventDefMessageRef   *Attribute[string]
	AttrSignalEventDefSignalRef     *Attribute[string]
	AttrErrorEventDefErrorRef       *Attribute[string]
	AttrEscalationEventDefRef       *Attribute[string]
	AttrLinkEventDefName            *Attribute[string]
)

// --- Implementations ---

type IntermediateCatchEventImpl struct {
	CatchEventImpl
}

func (*IntermediateCatchEventImpl) implIntermediateCatchEvent() {}
func (i *IntermediateCatchEventImpl) BpmnElementType() string {
	return BPMN_ELEMENT_INTERMEDIATE_CATCH_EVENT
}

type IntermediateThrowEventImpl struct {
	ThrowEventImpl
}

func (*IntermediateThrowEventImpl) implIntermediateThrowEvent() {}
func (i *IntermediateThrowEventImpl) BpmnElementType() string {
	return BPMN_ELEMENT_INTERMEDIATE_THROW_EVENT
}

type BoundaryEventImpl struct {
	CatchEventImpl
}

func (*BoundaryEventImpl) implBoundaryEvent()            {}
func (b *BoundaryEventImpl) BpmnElementType() string     { return BPMN_ELEMENT_BOUNDARY_EVENT }
func (b *BoundaryEventImpl) IsCancelActivity() bool      { return AttrBoundaryEventCancelActivity.Get(b) }
func (b *BoundaryEventImpl) SetCancelActivity(v bool)    { AttrBoundaryEventCancelActivity.Set(b, v) }
func (b *BoundaryEventImpl) GetAttachedToRef() string    { return AttrBoundaryEventAttachedToRef.Get(b) }
func (b *BoundaryEventImpl) SetAttachedToRef(ref string) { AttrBoundaryEventAttachedToRef.Set(b, ref) }
func (b *BoundaryEventImpl) GetAttachedTo() Activity {
	ref := b.GetAttachedToRef()
	if ref == "" {
		return nil
	}
	if act, ok := GetTypedElementById[Activity](b.GetModelInstance(), ref); ok {
		return act
	}
	return nil
}

// --- Event Definition Implementations ---

type EventDefinitionImpl struct {
	BaseElementImpl
}

func (*EventDefinitionImpl) implEventDefinition()      {}
func (e *EventDefinitionImpl) BpmnElementType() string { return BPMN_ELEMENT_EVENT_DEFINITION }

type MessageEventDefinitionImpl struct {
	EventDefinitionImpl
}

func (*MessageEventDefinitionImpl) implMessageEventDefinition() {}
func (m *MessageEventDefinitionImpl) BpmnElementType() string {
	return BPMN_ELEMENT_MESSAGE_EVENT_DEFINITION
}
func (m *MessageEventDefinitionImpl) GetMessageRef() string {
	return AttrMessageEventDefMessageRef.Get(m)
}
func (m *MessageEventDefinitionImpl) SetMessageRef(ref string) {
	AttrMessageEventDefMessageRef.Set(m, ref)
}

type TimerEventDefinitionImpl struct {
	EventDefinitionImpl
}

func (*TimerEventDefinitionImpl) implTimerEventDefinition() {}
func (t *TimerEventDefinitionImpl) BpmnElementType() string {
	return BPMN_ELEMENT_TIMER_EVENT_DEFINITION
}

type SignalEventDefinitionImpl struct {
	EventDefinitionImpl
}

func (*SignalEventDefinitionImpl) implSignalEventDefinition() {}
func (s *SignalEventDefinitionImpl) BpmnElementType() string {
	return BPMN_ELEMENT_SIGNAL_EVENT_DEFINITION
}
func (s *SignalEventDefinitionImpl) GetSignalRef() string    { return AttrSignalEventDefSignalRef.Get(s) }
func (s *SignalEventDefinitionImpl) SetSignalRef(ref string) { AttrSignalEventDefSignalRef.Set(s, ref) }

type ErrorEventDefinitionImpl struct {
	EventDefinitionImpl
}

func (*ErrorEventDefinitionImpl) implErrorEventDefinition() {}
func (e *ErrorEventDefinitionImpl) BpmnElementType() string {
	return BPMN_ELEMENT_ERROR_EVENT_DEFINITION
}
func (e *ErrorEventDefinitionImpl) GetErrorRef() string    { return AttrErrorEventDefErrorRef.Get(e) }
func (e *ErrorEventDefinitionImpl) SetErrorRef(ref string) { AttrErrorEventDefErrorRef.Set(e, ref) }

type EscalationEventDefinitionImpl struct {
	EventDefinitionImpl
}

func (*EscalationEventDefinitionImpl) implEscalationEventDefinition() {}
func (e *EscalationEventDefinitionImpl) BpmnElementType() string {
	return BPMN_ELEMENT_ESCALATION_EVENT_DEFINITION
}
func (e *EscalationEventDefinitionImpl) GetEscalationRef() string {
	return AttrEscalationEventDefRef.Get(e)
}
func (e *EscalationEventDefinitionImpl) SetEscalationRef(ref string) {
	AttrEscalationEventDefRef.Set(e, ref)
}

type ConditionalEventDefinitionImpl struct {
	EventDefinitionImpl
}

func (*ConditionalEventDefinitionImpl) implConditionalEventDefinition() {}
func (c *ConditionalEventDefinitionImpl) BpmnElementType() string {
	return BPMN_ELEMENT_CONDITIONAL_EVENT_DEFINITION
}

type LinkEventDefinitionImpl struct {
	EventDefinitionImpl
}

func (*LinkEventDefinitionImpl) implLinkEventDefinition()  {}
func (l *LinkEventDefinitionImpl) BpmnElementType() string { return BPMN_ELEMENT_LINK_EVENT_DEFINITION }
func (l *LinkEventDefinitionImpl) GetLinkName() string     { return AttrLinkEventDefName.Get(l) }
func (l *LinkEventDefinitionImpl) SetLinkName(name string) { AttrLinkEventDefName.Set(l, name) }

type CompensateEventDefinitionImpl struct {
	EventDefinitionImpl
}

func (*CompensateEventDefinitionImpl) implCompensateEventDefinition() {}
func (c *CompensateEventDefinitionImpl) BpmnElementType() string {
	return BPMN_ELEMENT_COMPENSATE_EVENT_DEFINITION
}

type CancelEventDefinitionImpl struct {
	EventDefinitionImpl
}

func (*CancelEventDefinitionImpl) implCancelEventDefinition() {}
func (c *CancelEventDefinitionImpl) BpmnElementType() string {
	return BPMN_ELEMENT_CANCEL_EVENT_DEFINITION
}

type TerminateEventDefinitionImpl struct {
	EventDefinitionImpl
}

func (*TerminateEventDefinitionImpl) implTerminateEventDefinition() {}
func (t *TerminateEventDefinitionImpl) BpmnElementType() string {
	return BPMN_ELEMENT_TERMINATE_EVENT_DEFINITION
}

// --- Registration ---

func registerIntermediateCatchEventType(mb *ModelBuilder) {
	mb.DefineType((*IntermediateCatchEvent)(nil), BPMN_ELEMENT_INTERMEDIATE_CATCH_EVENT).
		Namespace(BPMN20_NS).
		ExtendsType((*CatchEvent)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &IntermediateCatchEventImpl{CatchEventImpl{EventImpl{FlowNodeImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}}}}
		})
}

func registerIntermediateThrowEventType(mb *ModelBuilder) {
	mb.DefineType((*IntermediateThrowEvent)(nil), BPMN_ELEMENT_INTERMEDIATE_THROW_EVENT).
		Namespace(BPMN20_NS).
		ExtendsType((*ThrowEvent)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &IntermediateThrowEventImpl{ThrowEventImpl{EventImpl{FlowNodeImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}}}}
		})
}

func registerBoundaryEventType(mb *ModelBuilder) {
	tb := mb.DefineType((*BoundaryEvent)(nil), BPMN_ELEMENT_BOUNDARY_EVENT).
		Namespace(BPMN20_NS).
		ExtendsType((*CatchEvent)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &BoundaryEventImpl{CatchEventImpl{EventImpl{FlowNodeImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}}}}
		})

	AttrBoundaryEventCancelActivity = tb.BoolAttribute(BPMN_ATTRIBUTE_CANCEL_ACTIVITY).DefaultValue(true).Build()
	AttrBoundaryEventAttachedToRef = tb.StringAttribute(BPMN_ATTRIBUTE_ATTACHED_TO_REF).Required().Build()
}

func registerEventDefinitionType(mb *ModelBuilder) {
	mb.DefineType((*EventDefinition)(nil), BPMN_ELEMENT_EVENT_DEFINITION).
		Namespace(BPMN20_NS).
		ExtendsType((*RootElement)(nil)).
		AbstractType()
}

func registerMessageEventDefinitionType(mb *ModelBuilder) {
	tb := mb.DefineType((*MessageEventDefinition)(nil), BPMN_ELEMENT_MESSAGE_EVENT_DEFINITION).
		Namespace(BPMN20_NS).
		ExtendsType((*EventDefinition)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &MessageEventDefinitionImpl{EventDefinitionImpl{BaseElementImpl{NewBaseInstance(ctx)}}}
		})

	AttrMessageEventDefMessageRef = tb.StringAttribute(BPMN_ATTRIBUTE_MESSAGE_REF).Build()
}

func registerTimerEventDefinitionType(mb *ModelBuilder) {
	mb.DefineType((*TimerEventDefinition)(nil), BPMN_ELEMENT_TIMER_EVENT_DEFINITION).
		Namespace(BPMN20_NS).
		ExtendsType((*EventDefinition)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &TimerEventDefinitionImpl{EventDefinitionImpl{BaseElementImpl{NewBaseInstance(ctx)}}}
		})
}

func registerSignalEventDefinitionType(mb *ModelBuilder) {
	tb := mb.DefineType((*SignalEventDefinition)(nil), BPMN_ELEMENT_SIGNAL_EVENT_DEFINITION).
		Namespace(BPMN20_NS).
		ExtendsType((*EventDefinition)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &SignalEventDefinitionImpl{EventDefinitionImpl{BaseElementImpl{NewBaseInstance(ctx)}}}
		})

	AttrSignalEventDefSignalRef = tb.StringAttribute(BPMN_ATTRIBUTE_SIGNAL_REF).Build()
}

func registerErrorEventDefinitionType(mb *ModelBuilder) {
	tb := mb.DefineType((*ErrorEventDefinition)(nil), BPMN_ELEMENT_ERROR_EVENT_DEFINITION).
		Namespace(BPMN20_NS).
		ExtendsType((*EventDefinition)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ErrorEventDefinitionImpl{EventDefinitionImpl{BaseElementImpl{NewBaseInstance(ctx)}}}
		})

	AttrErrorEventDefErrorRef = tb.StringAttribute(BPMN_ATTRIBUTE_ERROR_REF).Build()
}

func registerEscalationEventDefinitionType(mb *ModelBuilder) {
	tb := mb.DefineType((*EscalationEventDefinition)(nil), BPMN_ELEMENT_ESCALATION_EVENT_DEFINITION).
		Namespace(BPMN20_NS).
		ExtendsType((*EventDefinition)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &EscalationEventDefinitionImpl{EventDefinitionImpl{BaseElementImpl{NewBaseInstance(ctx)}}}
		})

	AttrEscalationEventDefRef = tb.StringAttribute(BPMN_ATTRIBUTE_ESCALATION_REF).Build()
}

func registerConditionalEventDefinitionType(mb *ModelBuilder) {
	mb.DefineType((*ConditionalEventDefinition)(nil), BPMN_ELEMENT_CONDITIONAL_EVENT_DEFINITION).
		Namespace(BPMN20_NS).
		ExtendsType((*EventDefinition)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ConditionalEventDefinitionImpl{EventDefinitionImpl{BaseElementImpl{NewBaseInstance(ctx)}}}
		})
}

func registerLinkEventDefinitionType(mb *ModelBuilder) {
	tb := mb.DefineType((*LinkEventDefinition)(nil), BPMN_ELEMENT_LINK_EVENT_DEFINITION).
		Namespace(BPMN20_NS).
		ExtendsType((*EventDefinition)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &LinkEventDefinitionImpl{EventDefinitionImpl{BaseElementImpl{NewBaseInstance(ctx)}}}
		})

	AttrLinkEventDefName = tb.StringAttribute(BPMN_ATTRIBUTE_LINK_NAME).Build()
}

func registerCompensateEventDefinitionType(mb *ModelBuilder) {
	mb.DefineType((*CompensateEventDefinition)(nil), BPMN_ELEMENT_COMPENSATE_EVENT_DEFINITION).
		Namespace(BPMN20_NS).
		ExtendsType((*EventDefinition)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &CompensateEventDefinitionImpl{EventDefinitionImpl{BaseElementImpl{NewBaseInstance(ctx)}}}
		})
}

func registerCancelEventDefinitionType(mb *ModelBuilder) {
	mb.DefineType((*CancelEventDefinition)(nil), BPMN_ELEMENT_CANCEL_EVENT_DEFINITION).
		Namespace(BPMN20_NS).
		ExtendsType((*EventDefinition)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &CancelEventDefinitionImpl{EventDefinitionImpl{BaseElementImpl{NewBaseInstance(ctx)}}}
		})
}

func registerTerminateEventDefinitionType(mb *ModelBuilder) {
	mb.DefineType((*TerminateEventDefinition)(nil), BPMN_ELEMENT_TERMINATE_EVENT_DEFINITION).
		Namespace(BPMN20_NS).
		ExtendsType((*EventDefinition)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &TerminateEventDefinitionImpl{EventDefinitionImpl{BaseElementImpl{NewBaseInstance(ctx)}}}
		})
}
