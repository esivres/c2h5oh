package bpmn_model

// --- Interfaces ---

// Message represents a BPMN message.
type Message interface {
	RootElement
	implMessage()
	GetName() string
	SetName(name string)
}

// Signal represents a BPMN signal.
type Signal interface {
	RootElement
	implSignal()
	GetName() string
	SetName(name string)
}

// Error represents a BPMN error.
type Error interface {
	RootElement
	implError()
	GetName() string
	SetName(name string)
	GetErrorCode() string
	SetErrorCode(code string)
	GetStructureRef() string
	SetStructureRef(ref string)
}

// Escalation represents a BPMN escalation.
type Escalation interface {
	RootElement
	implEscalation()
	GetName() string
	SetName(name string)
	GetEscalationCode() string
	SetEscalationCode(code string)
}

// ItemDefinition defines data structures.
type ItemDefinition interface {
	RootElement
	implItemDefinition()
	GetStructureRef() string
	SetStructureRef(ref string)
}

// Collaboration represents a BPMN collaboration.
type Collaboration interface {
	RootElement
	implCollaboration()
	GetName() string
	SetName(name string)
	IsClosed() bool
	SetIsClosed(v bool)
}

// Participant represents a participant in a collaboration.
type Participant interface {
	BaseElement
	implParticipant()
	GetName() string
	SetName(name string)
	GetProcessRef() string
	SetProcessRef(ref string)
}

// MessageFlow connects participants in a collaboration.
type MessageFlow interface {
	BaseElement
	implMessageFlow()
	GetName() string
	SetName(name string)
	GetSourceRef() string
	SetSourceRef(ref string)
	GetTargetRef() string
	SetTargetRef(ref string)
}

// LaneSet is a set of lanes within a process.
type LaneSet interface {
	BaseElement
	implLaneSet()
}

// Lane represents a lane within a process.
type Lane interface {
	BaseElement
	implLane()
	GetName() string
	SetName(name string)
}

// DataObject represents a data object.
type DataObject interface {
	FlowElement
	implDataObject()
}

// DataObjectReference references a data object.
type DataObjectReference interface {
	FlowElement
	implDataObjectReference()
}

// Artifact is the abstract base for artifacts.
type Artifact interface {
	BaseElement
	implArtifact()
}

// TextAnnotation provides comments in the diagram.
type TextAnnotation interface {
	Artifact
	implTextAnnotation()
	GetText() string
	SetText(text string)
	GetTextFormat() string
	SetTextFormat(format string)
}

// Group visually groups elements.
type Group interface {
	Artifact
	implGroup()
	GetCategoryValueRef() string
	SetCategoryValueRef(ref string)
}

// Association connects an artifact to a flow element.
type Association interface {
	Artifact
	implAssociation()
	GetSourceRef() string
	SetSourceRef(ref string)
	GetTargetRef() string
	SetTargetRef(ref string)
	GetAssociationDirection() string
	SetAssociationDirection(dir string)
}

// MultiInstanceLoopCharacteristics defines multi-instance behavior.
type MultiInstanceLoopCharacteristics interface {
	BaseElement
	implMultiInstanceLoopCharacteristics()
	IsSequential() bool
	SetIsSequential(v bool)
}

// Documentation provides documentation for a BPMN element.
type Documentation interface {
	BaseElement
	implDocumentation()
	GetTextContent() string
	SetTextContent(text string)
	GetTextFormat() string
	SetTextFormat(format string)
}

// --- Attribute descriptors ---

var (
	AttrErrorCode                 *Attribute[string]
	AttrEscalationCode            *Attribute[string]
	AttrStructureRef              *Attribute[string]
	AttrCollaborationIsClosed     *Attribute[bool]
	AttrParticipantProcessRef     *Attribute[string]
	AttrMessageFlowSourceRef      *Attribute[string]
	AttrMessageFlowTargetRef      *Attribute[string]
	AttrTextAnnotationTextFormat  *Attribute[string]
	AttrGroupCategoryValueRef     *Attribute[string]
	AttrAssociationSourceRef      *Attribute[string]
	AttrAssociationTargetRef      *Attribute[string]
	AttrAssociationDirection      *Attribute[string]
	AttrMultiInstanceIsSequential *Attribute[bool]
	AttrDocumentationTextFormat   *Attribute[string]
)

// --- Implementations ---

type MessageImpl struct {
	BaseElementImpl
}

func (*MessageImpl) implMessage()              {}
func (m *MessageImpl) BpmnElementType() string { return BPMN_ELEMENT_MESSAGE }
func (m *MessageImpl) GetName() string         { return AttrName.Get(m) }
func (m *MessageImpl) SetName(name string)     { AttrName.Set(m, name) }

type SignalImpl struct {
	BaseElementImpl
}

func (*SignalImpl) implSignal()               {}
func (s *SignalImpl) BpmnElementType() string { return BPMN_ELEMENT_SIGNAL }
func (s *SignalImpl) GetName() string         { return AttrName.Get(s) }
func (s *SignalImpl) SetName(name string)     { AttrName.Set(s, name) }

type ErrorImpl struct {
	BaseElementImpl
}

func (*ErrorImpl) implError()                   {}
func (e *ErrorImpl) BpmnElementType() string    { return BPMN_ELEMENT_ERROR }
func (e *ErrorImpl) GetName() string            { return AttrName.Get(e) }
func (e *ErrorImpl) SetName(name string)        { AttrName.Set(e, name) }
func (e *ErrorImpl) GetErrorCode() string       { return AttrErrorCode.Get(e) }
func (e *ErrorImpl) SetErrorCode(code string)   { AttrErrorCode.Set(e, code) }
func (e *ErrorImpl) GetStructureRef() string    { return AttrStructureRef.Get(e) }
func (e *ErrorImpl) SetStructureRef(ref string) { AttrStructureRef.Set(e, ref) }

type EscalationImpl struct {
	BaseElementImpl
}

func (*EscalationImpl) implEscalation()                 {}
func (e *EscalationImpl) BpmnElementType() string       { return BPMN_ELEMENT_ESCALATION }
func (e *EscalationImpl) GetName() string               { return AttrName.Get(e) }
func (e *EscalationImpl) SetName(name string)           { AttrName.Set(e, name) }
func (e *EscalationImpl) GetEscalationCode() string     { return AttrEscalationCode.Get(e) }
func (e *EscalationImpl) SetEscalationCode(code string) { AttrEscalationCode.Set(e, code) }

type ItemDefinitionImpl struct {
	BaseElementImpl
}

func (*ItemDefinitionImpl) implItemDefinition()          {}
func (i *ItemDefinitionImpl) BpmnElementType() string    { return BPMN_ELEMENT_ITEM_DEFINITION }
func (i *ItemDefinitionImpl) GetStructureRef() string    { return AttrStructureRef.Get(i) }
func (i *ItemDefinitionImpl) SetStructureRef(ref string) { AttrStructureRef.Set(i, ref) }

type CollaborationImpl struct {
	BaseElementImpl
}

func (*CollaborationImpl) implCollaboration()        {}
func (c *CollaborationImpl) BpmnElementType() string { return BPMN_ELEMENT_COLLABORATION }
func (c *CollaborationImpl) GetName() string         { return AttrName.Get(c) }
func (c *CollaborationImpl) SetName(name string)     { AttrName.Set(c, name) }
func (c *CollaborationImpl) IsClosed() bool          { return AttrCollaborationIsClosed.Get(c) }
func (c *CollaborationImpl) SetIsClosed(v bool)      { AttrCollaborationIsClosed.Set(c, v) }

type ParticipantImpl struct {
	BaseElementImpl
}

func (*ParticipantImpl) implParticipant()           {}
func (p *ParticipantImpl) BpmnElementType() string  { return BPMN_ELEMENT_PARTICIPANT }
func (p *ParticipantImpl) GetName() string          { return AttrName.Get(p) }
func (p *ParticipantImpl) SetName(name string)      { AttrName.Set(p, name) }
func (p *ParticipantImpl) GetProcessRef() string    { return AttrParticipantProcessRef.Get(p) }
func (p *ParticipantImpl) SetProcessRef(ref string) { AttrParticipantProcessRef.Set(p, ref) }

type MessageFlowImpl struct {
	BaseElementImpl
}

func (*MessageFlowImpl) implMessageFlow()          {}
func (m *MessageFlowImpl) BpmnElementType() string { return BPMN_ELEMENT_MESSAGE_FLOW }
func (m *MessageFlowImpl) GetName() string         { return AttrName.Get(m) }
func (m *MessageFlowImpl) SetName(name string)     { AttrName.Set(m, name) }
func (m *MessageFlowImpl) GetSourceRef() string    { return AttrMessageFlowSourceRef.Get(m) }
func (m *MessageFlowImpl) SetSourceRef(ref string) { AttrMessageFlowSourceRef.Set(m, ref) }
func (m *MessageFlowImpl) GetTargetRef() string    { return AttrMessageFlowTargetRef.Get(m) }
func (m *MessageFlowImpl) SetTargetRef(ref string) { AttrMessageFlowTargetRef.Set(m, ref) }

type LaneSetImpl struct {
	BaseElementImpl
}

func (*LaneSetImpl) implLaneSet()              {}
func (l *LaneSetImpl) BpmnElementType() string { return BPMN_ELEMENT_LANE_SET }

type LaneImpl struct {
	BaseElementImpl
}

func (*LaneImpl) implLane()                 {}
func (l *LaneImpl) BpmnElementType() string { return BPMN_ELEMENT_LANE }
func (l *LaneImpl) GetName() string         { return AttrName.Get(l) }
func (l *LaneImpl) SetName(name string)     { AttrName.Set(l, name) }

type DataObjectImpl struct {
	FlowElementImpl
}

func (*DataObjectImpl) implDataObject()           {}
func (d *DataObjectImpl) BpmnElementType() string { return BPMN_ELEMENT_DATA_OBJECT }

type DataObjectReferenceImpl struct {
	FlowElementImpl
}

func (*DataObjectReferenceImpl) implDataObjectReference()  {}
func (d *DataObjectReferenceImpl) BpmnElementType() string { return BPMN_ELEMENT_DATA_OBJECT_REFERENCE }

type ArtifactImpl struct {
	BaseElementImpl
}

func (*ArtifactImpl) implArtifact()             {}
func (a *ArtifactImpl) BpmnElementType() string { return BPMN_ELEMENT_ARTIFACT }

type TextAnnotationImpl struct {
	ArtifactImpl
}

func (*TextAnnotationImpl) implTextAnnotation()       {}
func (t *TextAnnotationImpl) BpmnElementType() string { return BPMN_ELEMENT_TEXT_ANNOTATION }
func (t *TextAnnotationImpl) GetText() string {
	textElems := t.DomElement.GetChildElementsByNS(BPMN20_NS, "text")
	if len(textElems) > 0 {
		return textElems[0].GetTextContent()
	}
	return ""
}
func (t *TextAnnotationImpl) SetText(text string) {
	textElems := t.DomElement.GetChildElementsByNS(BPMN20_NS, "text")
	if len(textElems) > 0 {
		textElems[0].SetTextContent(text)
	}
	// If no text child exists, we'd need to create one — handled by builder
}
func (t *TextAnnotationImpl) GetTextFormat() string { return AttrTextAnnotationTextFormat.Get(t) }
func (t *TextAnnotationImpl) SetTextFormat(format string) {
	AttrTextAnnotationTextFormat.Set(t, format)
}

type GroupImpl struct {
	ArtifactImpl
}

func (*GroupImpl) implGroup()                       {}
func (g *GroupImpl) BpmnElementType() string        { return BPMN_ELEMENT_GROUP }
func (g *GroupImpl) GetCategoryValueRef() string    { return AttrGroupCategoryValueRef.Get(g) }
func (g *GroupImpl) SetCategoryValueRef(ref string) { AttrGroupCategoryValueRef.Set(g, ref) }

type AssociationImpl struct {
	ArtifactImpl
}

func (*AssociationImpl) implAssociation()                     {}
func (a *AssociationImpl) BpmnElementType() string            { return BPMN_ELEMENT_ASSOCIATION }
func (a *AssociationImpl) GetSourceRef() string               { return AttrAssociationSourceRef.Get(a) }
func (a *AssociationImpl) SetSourceRef(ref string)            { AttrAssociationSourceRef.Set(a, ref) }
func (a *AssociationImpl) GetTargetRef() string               { return AttrAssociationTargetRef.Get(a) }
func (a *AssociationImpl) SetTargetRef(ref string)            { AttrAssociationTargetRef.Set(a, ref) }
func (a *AssociationImpl) GetAssociationDirection() string    { return AttrAssociationDirection.Get(a) }
func (a *AssociationImpl) SetAssociationDirection(dir string) { AttrAssociationDirection.Set(a, dir) }

type MultiInstanceLoopCharacteristicsImpl struct {
	BaseElementImpl
}

func (*MultiInstanceLoopCharacteristicsImpl) implMultiInstanceLoopCharacteristics() {}
func (m *MultiInstanceLoopCharacteristicsImpl) BpmnElementType() string {
	return BPMN_ELEMENT_MULTI_INSTANCE_LOOP_CHARACTERISTICS
}
func (m *MultiInstanceLoopCharacteristicsImpl) IsSequential() bool {
	return AttrMultiInstanceIsSequential.Get(m)
}
func (m *MultiInstanceLoopCharacteristicsImpl) SetIsSequential(v bool) {
	AttrMultiInstanceIsSequential.Set(m, v)
}

type DocumentationImpl struct {
	BaseElementImpl
}

func (*DocumentationImpl) implDocumentation()            {}
func (d *DocumentationImpl) BpmnElementType() string     { return BPMN_ELEMENT_DOCUMENTATION }
func (d *DocumentationImpl) GetTextContent() string      { return d.DomElement.GetTextContent() }
func (d *DocumentationImpl) SetTextContent(text string)  { d.DomElement.SetTextContent(text) }
func (d *DocumentationImpl) GetTextFormat() string       { return AttrDocumentationTextFormat.Get(d) }
func (d *DocumentationImpl) SetTextFormat(format string) { AttrDocumentationTextFormat.Set(d, format) }

// --- Registration ---

func registerMessageType(mb *ModelBuilder) {
	mb.DefineType((*Message)(nil), BPMN_ELEMENT_MESSAGE).
		Namespace(BPMN20_NS).
		ExtendsType((*RootElement)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &MessageImpl{BaseElementImpl{NewBaseInstance(ctx)}}
		})
}

func registerSignalType(mb *ModelBuilder) {
	mb.DefineType((*Signal)(nil), BPMN_ELEMENT_SIGNAL).
		Namespace(BPMN20_NS).
		ExtendsType((*RootElement)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &SignalImpl{BaseElementImpl{NewBaseInstance(ctx)}}
		})
}

func registerErrorType(mb *ModelBuilder) {
	tb := mb.DefineType((*Error)(nil), BPMN_ELEMENT_ERROR).
		Namespace(BPMN20_NS).
		ExtendsType((*RootElement)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ErrorImpl{BaseElementImpl{NewBaseInstance(ctx)}}
		})

	AttrErrorCode = tb.StringAttribute(BPMN_ATTRIBUTE_ERROR_CODE).Build()
	AttrStructureRef = tb.StringAttribute(BPMN_ATTRIBUTE_STRUCTURE_REF).Build()
}

func registerEscalationType(mb *ModelBuilder) {
	tb := mb.DefineType((*Escalation)(nil), BPMN_ELEMENT_ESCALATION).
		Namespace(BPMN20_NS).
		ExtendsType((*RootElement)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &EscalationImpl{BaseElementImpl{NewBaseInstance(ctx)}}
		})

	AttrEscalationCode = tb.StringAttribute(BPMN_ATTRIBUTE_ESCALATION_CODE).Build()
}

func registerItemDefinitionType(mb *ModelBuilder) {
	mb.DefineType((*ItemDefinition)(nil), BPMN_ELEMENT_ITEM_DEFINITION).
		Namespace(BPMN20_NS).
		ExtendsType((*RootElement)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ItemDefinitionImpl{BaseElementImpl{NewBaseInstance(ctx)}}
		})
}

func registerCollaborationType(mb *ModelBuilder) {
	tb := mb.DefineType((*Collaboration)(nil), BPMN_ELEMENT_COLLABORATION).
		Namespace(BPMN20_NS).
		ExtendsType((*RootElement)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &CollaborationImpl{BaseElementImpl{NewBaseInstance(ctx)}}
		})

	AttrCollaborationIsClosed = tb.BoolAttribute(BPMN_ATTRIBUTE_IS_CLOSED).DefaultValue(false).Build()
}

func registerParticipantType(mb *ModelBuilder) {
	tb := mb.DefineType((*Participant)(nil), BPMN_ELEMENT_PARTICIPANT).
		Namespace(BPMN20_NS).
		ExtendsType((*BaseElement)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ParticipantImpl{BaseElementImpl{NewBaseInstance(ctx)}}
		})

	AttrParticipantProcessRef = tb.StringAttribute(BPMN_ATTRIBUTE_PROCESS_REF).Build()
}

func registerMessageFlowType(mb *ModelBuilder) {
	tb := mb.DefineType((*MessageFlow)(nil), BPMN_ELEMENT_MESSAGE_FLOW).
		Namespace(BPMN20_NS).
		ExtendsType((*BaseElement)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &MessageFlowImpl{BaseElementImpl{NewBaseInstance(ctx)}}
		})

	AttrMessageFlowSourceRef = tb.StringAttribute(BPMN_ATTRIBUTE_SOURCE_REF).Required().Build()
	AttrMessageFlowTargetRef = tb.StringAttribute(BPMN_ATTRIBUTE_TARGET_REF).Required().Build()
}

func registerLaneSetType(mb *ModelBuilder) {
	mb.DefineType((*LaneSet)(nil), BPMN_ELEMENT_LANE_SET).
		Namespace(BPMN20_NS).
		ExtendsType((*BaseElement)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &LaneSetImpl{BaseElementImpl{NewBaseInstance(ctx)}}
		})
}

func registerLaneType(mb *ModelBuilder) {
	mb.DefineType((*Lane)(nil), BPMN_ELEMENT_LANE).
		Namespace(BPMN20_NS).
		ExtendsType((*BaseElement)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &LaneImpl{BaseElementImpl{NewBaseInstance(ctx)}}
		})
}

func registerDataObjectType(mb *ModelBuilder) {
	mb.DefineType((*DataObject)(nil), BPMN_ELEMENT_DATA_OBJECT).
		Namespace(BPMN20_NS).
		ExtendsType((*FlowElement)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &DataObjectImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}
		})
}

func registerDataObjectReferenceType(mb *ModelBuilder) {
	mb.DefineType((*DataObjectReference)(nil), BPMN_ELEMENT_DATA_OBJECT_REFERENCE).
		Namespace(BPMN20_NS).
		ExtendsType((*FlowElement)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &DataObjectReferenceImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}
		})
}

func registerArtifactType(mb *ModelBuilder) {
	mb.DefineType((*Artifact)(nil), BPMN_ELEMENT_ARTIFACT).
		Namespace(BPMN20_NS).
		ExtendsType((*BaseElement)(nil)).
		AbstractType()
}

func registerTextAnnotationType(mb *ModelBuilder) {
	tb := mb.DefineType((*TextAnnotation)(nil), BPMN_ELEMENT_TEXT_ANNOTATION).
		Namespace(BPMN20_NS).
		ExtendsType((*Artifact)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &TextAnnotationImpl{ArtifactImpl{BaseElementImpl{NewBaseInstance(ctx)}}}
		})

	AttrTextAnnotationTextFormat = tb.StringAttribute(BPMN_ATTRIBUTE_TEXT_FORMAT).DefaultValue("text/plain").Build()
}

func registerGroupType(mb *ModelBuilder) {
	tb := mb.DefineType((*Group)(nil), BPMN_ELEMENT_GROUP).
		Namespace(BPMN20_NS).
		ExtendsType((*Artifact)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &GroupImpl{ArtifactImpl{BaseElementImpl{NewBaseInstance(ctx)}}}
		})

	AttrGroupCategoryValueRef = tb.StringAttribute(BPMN_ATTRIBUTE_CATEGORY_VALUE_REF).Build()
}

func registerAssociationType(mb *ModelBuilder) {
	tb := mb.DefineType((*Association)(nil), BPMN_ELEMENT_ASSOCIATION).
		Namespace(BPMN20_NS).
		ExtendsType((*Artifact)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &AssociationImpl{ArtifactImpl{BaseElementImpl{NewBaseInstance(ctx)}}}
		})

	AttrAssociationSourceRef = tb.StringAttribute(BPMN_ATTRIBUTE_SOURCE_REF).Required().Build()
	AttrAssociationTargetRef = tb.StringAttribute(BPMN_ATTRIBUTE_TARGET_REF).Required().Build()
	AttrAssociationDirection = tb.StringAttribute(BPMN_ATTRIBUTE_ASSOCIATION_DIRECTION).DefaultValue("None").Build()
}

func registerMultiInstanceLoopCharacteristicsType(mb *ModelBuilder) {
	tb := mb.DefineType((*MultiInstanceLoopCharacteristics)(nil), BPMN_ELEMENT_MULTI_INSTANCE_LOOP_CHARACTERISTICS).
		Namespace(BPMN20_NS).
		ExtendsType((*BaseElement)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &MultiInstanceLoopCharacteristicsImpl{BaseElementImpl{NewBaseInstance(ctx)}}
		})

	AttrMultiInstanceIsSequential = tb.BoolAttribute(BPMN_ATTRIBUTE_IS_SEQUENTIAL).DefaultValue(false).Build()
}

func registerDocumentationType(mb *ModelBuilder) {
	tb := mb.DefineType((*Documentation)(nil), BPMN_ELEMENT_DOCUMENTATION).
		Namespace(BPMN20_NS).
		ExtendsType((*BaseElement)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &DocumentationImpl{BaseElementImpl{NewBaseInstance(ctx)}}
		})

	AttrDocumentationTextFormat = tb.StringAttribute(BPMN_ATTRIBUTE_TEXT_FORMAT).DefaultValue("text/plain").Build()
}
