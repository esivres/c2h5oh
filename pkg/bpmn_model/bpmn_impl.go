package bpmn_model

// Attribute descriptors — shared across all types that use them.
var (
	AttrBaseElementId *Attribute[string]
	AttrName          *Attribute[string]

	AttrDefinitionsTargetNamespace    *Attribute[string]
	AttrDefinitionsExporter           *Attribute[string]
	AttrDefinitionsExporterVersion    *Attribute[string]
	AttrDefinitionsExpressionLanguage *Attribute[string]
	AttrDefinitionsTypeLanguage       *Attribute[string]

	AttrProcessType         *Attribute[string]
	AttrProcessIsClosed     *Attribute[bool]
	AttrProcessIsExecutable *Attribute[bool]

	AttrSequenceFlowSourceRef   *Attribute[string]
	AttrSequenceFlowTargetRef   *Attribute[string]
	AttrSequenceFlowIsImmediate *Attribute[bool]

	AttrStartEventIsInterrupting *Attribute[bool]

	AttrConditionExpressionType *Attribute[string]
)

// --- Definitions ---

type DefinitionsImpl struct {
	BaseInstance
}

func (d *DefinitionsImpl) BpmnElementType() string      { return BPMN_ELEMENT_DEFINITIONS }
func (d *DefinitionsImpl) GetId() string                { return AttrBaseElementId.Get(d) }
func (d *DefinitionsImpl) SetId(id string)              { AttrBaseElementId.Set(d, id) }
func (d *DefinitionsImpl) GetName() string              { return AttrName.Get(d) }
func (d *DefinitionsImpl) SetName(name string)          { AttrName.Set(d, name) }
func (d *DefinitionsImpl) GetTargetNamespace() string   { return AttrDefinitionsTargetNamespace.Get(d) }
func (d *DefinitionsImpl) SetTargetNamespace(ns string) { AttrDefinitionsTargetNamespace.Set(d, ns) }
func (d *DefinitionsImpl) GetExporter() string          { return AttrDefinitionsExporter.Get(d) }
func (d *DefinitionsImpl) SetExporter(exporter string)  { AttrDefinitionsExporter.Set(d, exporter) }
func (d *DefinitionsImpl) GetExporterVersion() string   { return AttrDefinitionsExporterVersion.Get(d) }
func (d *DefinitionsImpl) SetExporterVersion(version string) {
	AttrDefinitionsExporterVersion.Set(d, version)
}
func (d *DefinitionsImpl) GetRootElements() []ModelElementInstance {
	mi := d.GetModelInstance()
	if mi == nil {
		return nil
	}
	var result []ModelElementInstance
	for _, child := range d.DomElement.GetChildElements() {
		if inst := mi.GetElementByDom(child); inst != nil {
			result = append(result, inst)
		}
	}
	return result
}

// --- BaseElement ---

type BaseElementImpl struct {
	BaseInstance
}

func (b *BaseElementImpl) BpmnElementType() string { return BPMN_ELEMENT_BASE_ELEMENT }
func (b *BaseElementImpl) GetId() string           { return AttrBaseElementId.Get(b) }
func (b *BaseElementImpl) SetId(id string)         { AttrBaseElementId.Set(b, id) }

func (b *BaseElementImpl) GetExtensionElements() ExtensionElements {
	mi := b.GetModelInstance()
	if mi == nil {
		return nil
	}
	extElems := b.DomElement.GetChildElementsByNS(BPMN20_NS, BPMN_ELEMENT_EXTENSION_ELEMENTS)
	if len(extElems) == 0 {
		return nil
	}
	if inst := mi.GetElementByDom(extElems[0]); inst != nil {
		if ee, ok := inst.(ExtensionElements); ok {
			return ee
		}
	}
	return nil
}

func (b *BaseElementImpl) GetOrCreateExtensionElements() ExtensionElements {
	if ee := b.GetExtensionElements(); ee != nil {
		return ee
	}
	mi := b.GetModelInstance()
	eeType := bpmnModel.GetTypeByQName(BPMN20_NS, BPMN_ELEMENT_EXTENSION_ELEMENTS)
	inst, _ := mi.NewInstance(eeType)
	ee := inst.(ExtensionElements)
	b.DomElement.AppendChild(ee.GetDomElement())
	return ee
}

// --- FlowElement ---

type FlowElementImpl struct {
	BaseElementImpl
}

func (f *FlowElementImpl) GetName() string     { return AttrName.Get(f) }
func (f *FlowElementImpl) SetName(name string) { AttrName.Set(f, name) }

// --- FlowNode ---

type FlowNodeImpl struct {
	FlowElementImpl
}

func (f *FlowNodeImpl) GetIncomingSequenceFlows() []SequenceFlow {
	mi := f.GetModelInstance()
	if mi == nil {
		return nil
	}

	var result []SequenceFlow
	incomingElems := f.DomElement.GetChildElementsByNS(BPMN20_NS, BPMN_ELEMENT_INCOMING)
	for _, elem := range incomingElems {
		refId := elem.GetTextContent()
		if refId != "" {
			if sf, ok := GetTypedElementById[SequenceFlow](mi, refId); ok {
				result = append(result, sf)
			}
		}
	}
	return result
}

func (f *FlowNodeImpl) GetOutgoingSequenceFlows() []SequenceFlow {
	mi := f.GetModelInstance()
	if mi == nil {
		return nil
	}

	var result []SequenceFlow
	outgoingElems := f.DomElement.GetChildElementsByNS(BPMN20_NS, BPMN_ELEMENT_OUTGOING)
	for _, elem := range outgoingElems {
		refId := elem.GetTextContent()
		if refId != "" {
			if sf, ok := GetTypedElementById[SequenceFlow](mi, refId); ok {
				result = append(result, sf)
			}
		}
	}
	return result
}

// --- Event ---

type EventImpl struct {
	FlowNodeImpl
}

// --- CatchEvent ---

type CatchEventImpl struct {
	EventImpl
}

func (*CatchEventImpl) implCatchEvent() {}

// --- ThrowEvent ---

type ThrowEventImpl struct {
	EventImpl
}

func (*ThrowEventImpl) implThrowEvent() {}

// --- StartEvent ---

type StartEventImpl struct {
	CatchEventImpl
}

func (*StartEventImpl) implStartEvent()            {}
func (s *StartEventImpl) BpmnElementType() string  { return BPMN_ELEMENT_START_EVENT }
func (s *StartEventImpl) IsInterrupting() bool     { return AttrStartEventIsInterrupting.Get(s) }
func (s *StartEventImpl) SetIsInterrupting(v bool) { AttrStartEventIsInterrupting.Set(s, v) }

// --- EndEvent ---

type EndEventImpl struct {
	ThrowEventImpl
}

func (*EndEventImpl) implEndEvent()             {}
func (e *EndEventImpl) BpmnElementType() string { return BPMN_ELEMENT_END_EVENT }

// --- SequenceFlow ---

type SequenceFlowImpl struct {
	FlowElementImpl
}

func (*SequenceFlowImpl) implSequenceFlow()         {}
func (s *SequenceFlowImpl) BpmnElementType() string { return BPMN_ELEMENT_SEQUENCE_FLOW }
func (s *SequenceFlowImpl) GetSourceRef() string    { return AttrSequenceFlowSourceRef.Get(s) }
func (s *SequenceFlowImpl) SetSourceRef(ref string) { AttrSequenceFlowSourceRef.Set(s, ref) }
func (s *SequenceFlowImpl) GetTargetRef() string    { return AttrSequenceFlowTargetRef.Get(s) }
func (s *SequenceFlowImpl) SetTargetRef(ref string) { AttrSequenceFlowTargetRef.Set(s, ref) }

func (s *SequenceFlowImpl) GetSource() FlowNode {
	ref := s.GetSourceRef()
	if ref == "" {
		return nil
	}
	mi := s.GetModelInstance()
	if mi == nil {
		return nil
	}
	if node, ok := GetTypedElementById[FlowNode](mi, ref); ok {
		return node
	}
	return nil
}

func (s *SequenceFlowImpl) GetTarget() FlowNode {
	ref := s.GetTargetRef()
	if ref == "" {
		return nil
	}
	mi := s.GetModelInstance()
	if mi == nil {
		return nil
	}
	if node, ok := GetTypedElementById[FlowNode](mi, ref); ok {
		return node
	}
	return nil
}

func (s *SequenceFlowImpl) GetConditionExpression() ConditionExpression {
	mi := s.GetModelInstance()
	if mi == nil {
		return nil
	}
	condElems := s.DomElement.GetChildElementsByNS(BPMN20_NS, BPMN_ELEMENT_CONDITION_EXPRESSION)
	if len(condElems) == 0 {
		return nil
	}
	if inst := mi.GetElementByDom(condElems[0]); inst != nil {
		if ce, ok := inst.(ConditionExpression); ok {
			return ce
		}
	}
	return nil
}

func (s *SequenceFlowImpl) SetConditionExpression(expr ConditionExpression) {
	// Remove existing
	condElems := s.DomElement.GetChildElementsByNS(BPMN20_NS, BPMN_ELEMENT_CONDITION_EXPRESSION)
	for _, elem := range condElems {
		s.DomElement.RemoveChild(elem)
	}
	if expr != nil {
		s.DomElement.AppendChild(expr.GetDomElement())
		if mi := s.GetModelInstance(); mi != nil {
			mi.RegisterElement(expr)
		}
	}
}

// --- ConditionExpression ---

type ConditionExpressionImpl struct {
	BaseInstance
}

func (c *ConditionExpressionImpl) BpmnElementType() string    { return BPMN_ELEMENT_CONDITION_EXPRESSION }
func (c *ConditionExpressionImpl) GetTextContent() string     { return c.DomElement.GetTextContent() }
func (c *ConditionExpressionImpl) SetTextContent(text string) { c.DomElement.SetTextContent(text) }
func (c *ConditionExpressionImpl) GetType() string            { return AttrConditionExpressionType.Get(c) }
func (c *ConditionExpressionImpl) SetType(t string)           { AttrConditionExpressionType.Set(c, t) }

// --- Process ---

type CallableElementImpl struct {
	BaseElementImpl
}

func (c *CallableElementImpl) GetName() string     { return AttrName.Get(c) }
func (c *CallableElementImpl) SetName(name string) { AttrName.Set(c, name) }

type ProcessImpl struct {
	CallableElementImpl
}

func (*ProcessImpl) implProcess()                {}
func (p *ProcessImpl) BpmnElementType() string   { return BPMN_ELEMENT_PROCESS }
func (p *ProcessImpl) GetProcessType() string    { return AttrProcessType.Get(p) }
func (p *ProcessImpl) SetProcessType(pt string)  { AttrProcessType.Set(p, pt) }
func (p *ProcessImpl) IsExecutable() bool        { return AttrProcessIsExecutable.Get(p) }
func (p *ProcessImpl) SetIsExecutable(exec bool) { AttrProcessIsExecutable.Set(p, exec) }
func (p *ProcessImpl) IsClosed() bool            { return AttrProcessIsClosed.Get(p) }
func (p *ProcessImpl) SetIsClosed(closed bool)   { AttrProcessIsClosed.Set(p, closed) }

func (p *ProcessImpl) GetFlowElements() []FlowElement {
	mi := p.GetModelInstance()
	if mi == nil {
		return nil
	}
	var result []FlowElement
	for _, child := range p.DomElement.GetChildElements() {
		if inst := mi.GetElementByDom(child); inst != nil {
			if fe, ok := inst.(FlowElement); ok {
				result = append(result, fe)
			}
		}
	}
	return result
}

func (p *ProcessImpl) AddFlowElement(elem FlowElement) {
	p.DomElement.AppendChild(elem.GetDomElement())
	if mi := p.GetModelInstance(); mi != nil {
		mi.RegisterElement(elem)
	}
}

// --- ExtensionElements ---

type ExtensionElementsImpl struct {
	BaseInstance
}

func (e *ExtensionElementsImpl) BpmnElementType() string { return BPMN_ELEMENT_EXTENSION_ELEMENTS }
func (e *ExtensionElementsImpl) GetElements() []ModelElementInstance {
	mi := e.GetModelInstance()
	if mi == nil {
		return nil
	}
	var result []ModelElementInstance
	for _, child := range e.DomElement.GetChildElements() {
		if inst := mi.GetElementByDom(child); inst != nil {
			result = append(result, inst)
		}
	}
	return result
}
