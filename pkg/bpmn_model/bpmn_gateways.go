package bpmn_model

// --- Interfaces ---

// Gateway is the abstract base for all gateways.
type Gateway interface {
	FlowNode
	implGateway()
	GetGatewayDirection() string
	SetGatewayDirection(dir string)
	GetDefaultFlow() SequenceFlow
	SetDefaultFlowRef(ref string)
}

// ExclusiveGateway routes flow to exactly one outgoing path.
type ExclusiveGateway interface {
	Gateway
	implExclusiveGateway()
}

// ParallelGateway splits/joins flow into/from parallel paths.
type ParallelGateway interface {
	Gateway
	implParallelGateway()
}

// InclusiveGateway routes flow to one or more outgoing paths.
type InclusiveGateway interface {
	Gateway
	implInclusiveGateway()
}

// EventBasedGateway routes flow based on events.
type EventBasedGateway interface {
	Gateway
	implEventBasedGateway()
	IsInstantiate() bool
	SetInstantiate(v bool)
	GetEventGatewayType() string
	SetEventGatewayType(t string)
}

// ComplexGateway uses complex conditions to route flow.
type ComplexGateway interface {
	Gateway
	implComplexGateway()
}

// --- Attribute descriptors ---

var (
	AttrGatewayDirection        *Attribute[string]
	AttrGatewayDefaultRef       *Attribute[string]
	AttrEventGatewayInstantiate *Attribute[bool]
	AttrEventGatewayType        *Attribute[string]
)

// --- Implementations ---

type GatewayImpl struct {
	FlowNodeImpl
}

func (*GatewayImpl) implGateway()                     {}
func (g *GatewayImpl) BpmnElementType() string        { return BPMN_ELEMENT_GATEWAY }
func (g *GatewayImpl) GetGatewayDirection() string    { return AttrGatewayDirection.Get(g) }
func (g *GatewayImpl) SetGatewayDirection(dir string) { AttrGatewayDirection.Set(g, dir) }
func (g *GatewayImpl) GetDefaultFlow() SequenceFlow {
	ref := AttrGatewayDefaultRef.Get(g)
	if ref == "" {
		return nil
	}
	if sf, ok := GetTypedElementById[SequenceFlow](g.GetModelInstance(), ref); ok {
		return sf
	}
	return nil
}
func (g *GatewayImpl) SetDefaultFlowRef(ref string) { AttrGatewayDefaultRef.Set(g, ref) }

type ExclusiveGatewayImpl struct {
	GatewayImpl
}

func (*ExclusiveGatewayImpl) implExclusiveGateway()     {}
func (e *ExclusiveGatewayImpl) BpmnElementType() string { return BPMN_ELEMENT_EXCLUSIVE_GATEWAY }

type ParallelGatewayImpl struct {
	GatewayImpl
}

func (*ParallelGatewayImpl) implParallelGateway()      {}
func (p *ParallelGatewayImpl) BpmnElementType() string { return BPMN_ELEMENT_PARALLEL_GATEWAY }

type InclusiveGatewayImpl struct {
	GatewayImpl
}

func (*InclusiveGatewayImpl) implInclusiveGateway()     {}
func (i *InclusiveGatewayImpl) BpmnElementType() string { return BPMN_ELEMENT_INCLUSIVE_GATEWAY }

type EventBasedGatewayImpl struct {
	GatewayImpl
}

func (*EventBasedGatewayImpl) implEventBasedGateway()         {}
func (e *EventBasedGatewayImpl) BpmnElementType() string      { return BPMN_ELEMENT_EVENT_BASED_GATEWAY }
func (e *EventBasedGatewayImpl) IsInstantiate() bool          { return AttrEventGatewayInstantiate.Get(e) }
func (e *EventBasedGatewayImpl) SetInstantiate(v bool)        { AttrEventGatewayInstantiate.Set(e, v) }
func (e *EventBasedGatewayImpl) GetEventGatewayType() string  { return AttrEventGatewayType.Get(e) }
func (e *EventBasedGatewayImpl) SetEventGatewayType(t string) { AttrEventGatewayType.Set(e, t) }

type ComplexGatewayImpl struct {
	GatewayImpl
}

func (*ComplexGatewayImpl) implComplexGateway()       {}
func (c *ComplexGatewayImpl) BpmnElementType() string { return BPMN_ELEMENT_COMPLEX_GATEWAY }

// --- Registration ---

func registerGatewayType(mb *ModelBuilder) {
	tb := mb.DefineType((*Gateway)(nil), BPMN_ELEMENT_GATEWAY).
		Namespace(BPMN20_NS).
		ExtendsType((*FlowNode)(nil)).
		AbstractType()

	AttrGatewayDirection = tb.StringAttribute(BPMN_ATTRIBUTE_GATEWAY_DIRECTION).DefaultValue("Unspecified").Build()
	AttrGatewayDefaultRef = tb.StringAttribute(BPMN_ATTRIBUTE_DEFAULT).Build()
}

func registerExclusiveGatewayType(mb *ModelBuilder) {
	mb.DefineType((*ExclusiveGateway)(nil), BPMN_ELEMENT_EXCLUSIVE_GATEWAY).
		Namespace(BPMN20_NS).
		ExtendsType((*Gateway)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ExclusiveGatewayImpl{GatewayImpl{FlowNodeImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}}}
		})
}

func registerParallelGatewayType(mb *ModelBuilder) {
	mb.DefineType((*ParallelGateway)(nil), BPMN_ELEMENT_PARALLEL_GATEWAY).
		Namespace(BPMN20_NS).
		ExtendsType((*Gateway)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ParallelGatewayImpl{GatewayImpl{FlowNodeImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}}}
		})
}

func registerInclusiveGatewayType(mb *ModelBuilder) {
	mb.DefineType((*InclusiveGateway)(nil), BPMN_ELEMENT_INCLUSIVE_GATEWAY).
		Namespace(BPMN20_NS).
		ExtendsType((*Gateway)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &InclusiveGatewayImpl{GatewayImpl{FlowNodeImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}}}
		})
}

func registerEventBasedGatewayType(mb *ModelBuilder) {
	tb := mb.DefineType((*EventBasedGateway)(nil), BPMN_ELEMENT_EVENT_BASED_GATEWAY).
		Namespace(BPMN20_NS).
		ExtendsType((*Gateway)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &EventBasedGatewayImpl{GatewayImpl{FlowNodeImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}}}
		})

	AttrEventGatewayInstantiate = tb.BoolAttribute(BPMN_ATTRIBUTE_INSTANTIATE).DefaultValue(false).Build()
	AttrEventGatewayType = tb.StringAttribute(BPMN_ATTRIBUTE_EVENT_GATEWAY_TYPE).DefaultValue("Exclusive").Build()
}

func registerComplexGatewayType(mb *ModelBuilder) {
	mb.DefineType((*ComplexGateway)(nil), BPMN_ELEMENT_COMPLEX_GATEWAY).
		Namespace(BPMN20_NS).
		ExtendsType((*Gateway)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ComplexGatewayImpl{GatewayImpl{FlowNodeImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}}}
		})
}
