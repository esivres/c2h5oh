package bpmn_model

// --- Interfaces ---

// Activity is the abstract base for tasks and sub-processes.
type Activity interface {
	FlowNode
	IsForCompensation() bool
	SetIsForCompensation(v bool)
	GetDefaultFlow() SequenceFlow
	SetDefaultFlowRef(ref string)
}

// Task is a generic task (abstract base for specific task types).
type Task interface {
	Activity
	implTask()
}

// ServiceTask is an automated task executed by a service.
type ServiceTask interface {
	Task
	implServiceTask()
	GetImplementation() string
	SetImplementation(impl string)
}

// UserTask is a task performed by a human.
type UserTask interface {
	Task
	implUserTask()
	GetImplementation() string
	SetImplementation(impl string)
}

// ScriptTask is a task that executes a script.
type ScriptTask interface {
	Task
	implScriptTask()
	GetScriptFormat() string
	SetScriptFormat(format string)
}

// BusinessRuleTask invokes a business rule engine (e.g., DMN).
type BusinessRuleTask interface {
	Task
	implBusinessRuleTask()
	GetImplementation() string
	SetImplementation(impl string)
}

// SendTask sends a message.
type SendTask interface {
	Task
	implSendTask()
	GetImplementation() string
	SetImplementation(impl string)
}

// ReceiveTask waits for a message.
type ReceiveTask interface {
	Task
	implReceiveTask()
	GetImplementation() string
	SetImplementation(impl string)
}

// ManualTask is performed without automation.
type ManualTask interface {
	Task
	implManualTask()
}

// SubProcess is an embedded sub-process containing flow elements.
type SubProcess interface {
	Activity
	implSubProcess()
	IsTriggeredByEvent() bool
	SetTriggeredByEvent(v bool)
	GetFlowElements() []FlowElement
	AddFlowElement(elem FlowElement)
}

// AdHocSubProcess is an ad-hoc sub-process whose inner activities can be executed in any order.
type AdHocSubProcess interface {
	SubProcess
	implAdHocSubProcess()
	GetCompletionCondition() string
	SetCompletionCondition(expr string)
	GetCancelRemainingInstances() bool
	SetCancelRemainingInstances(v bool)
}

// CallActivity calls an external process or global task.
type CallActivity interface {
	Activity
	implCallActivity()
	GetCalledElement() string
	SetCalledElement(elem string)
}

// --- Attribute descriptors ---

var (
	AttrActivityIsForCompensation  *Attribute[bool]
	AttrActivityDefaultRef         *Attribute[string]
	AttrServiceTaskImplementation  *Attribute[string]
	AttrUserTaskImplementation     *Attribute[string]
	AttrScriptTaskScriptFormat     *Attribute[string]
	AttrBusinessRuleTaskImpl       *Attribute[string]
	AttrSendTaskImplementation     *Attribute[string]
	AttrReceiveTaskImplementation  *Attribute[string]
	AttrSubProcessTriggeredByEvent *Attribute[bool]
	AttrCallActivityCalledElement  *Attribute[string]
	AttrCancelRemainingInstances   *Attribute[bool]
)

// --- Implementations ---

type ActivityImpl struct {
	FlowNodeImpl
}

func (a *ActivityImpl) BpmnElementType() string     { return BPMN_ELEMENT_ACTIVITY }
func (a *ActivityImpl) IsForCompensation() bool     { return AttrActivityIsForCompensation.Get(a) }
func (a *ActivityImpl) SetIsForCompensation(v bool) { AttrActivityIsForCompensation.Set(a, v) }
func (a *ActivityImpl) GetDefaultFlow() SequenceFlow {
	ref := AttrActivityDefaultRef.Get(a)
	if ref == "" {
		return nil
	}
	if sf, ok := GetTypedElementById[SequenceFlow](a.GetModelInstance(), ref); ok {
		return sf
	}
	return nil
}
func (a *ActivityImpl) SetDefaultFlowRef(ref string) { AttrActivityDefaultRef.Set(a, ref) }

type TaskImpl struct {
	ActivityImpl
}

func (t *TaskImpl) BpmnElementType() string { return BPMN_ELEMENT_TASK }
func (*TaskImpl) implTask()                 {}

type ServiceTaskImpl struct {
	TaskImpl
}

func (*ServiceTaskImpl) implServiceTask()                {}
func (s *ServiceTaskImpl) BpmnElementType() string       { return BPMN_ELEMENT_SERVICE_TASK }
func (s *ServiceTaskImpl) GetImplementation() string     { return AttrServiceTaskImplementation.Get(s) }
func (s *ServiceTaskImpl) SetImplementation(impl string) { AttrServiceTaskImplementation.Set(s, impl) }

type UserTaskImpl struct {
	TaskImpl
}

func (*UserTaskImpl) implUserTask()                   {}
func (u *UserTaskImpl) BpmnElementType() string       { return BPMN_ELEMENT_USER_TASK }
func (u *UserTaskImpl) GetImplementation() string     { return AttrUserTaskImplementation.Get(u) }
func (u *UserTaskImpl) SetImplementation(impl string) { AttrUserTaskImplementation.Set(u, impl) }

type ScriptTaskImpl struct {
	TaskImpl
}

func (*ScriptTaskImpl) implScriptTask()                 {}
func (s *ScriptTaskImpl) BpmnElementType() string       { return BPMN_ELEMENT_SCRIPT_TASK }
func (s *ScriptTaskImpl) GetScriptFormat() string       { return AttrScriptTaskScriptFormat.Get(s) }
func (s *ScriptTaskImpl) SetScriptFormat(format string) { AttrScriptTaskScriptFormat.Set(s, format) }

type BusinessRuleTaskImpl struct {
	TaskImpl
}

func (*BusinessRuleTaskImpl) implBusinessRuleTask()           {}
func (b *BusinessRuleTaskImpl) BpmnElementType() string       { return BPMN_ELEMENT_BUSINESS_RULE_TASK }
func (b *BusinessRuleTaskImpl) GetImplementation() string     { return AttrBusinessRuleTaskImpl.Get(b) }
func (b *BusinessRuleTaskImpl) SetImplementation(impl string) { AttrBusinessRuleTaskImpl.Set(b, impl) }

type SendTaskImpl struct {
	TaskImpl
}

func (*SendTaskImpl) implSendTask()                   {}
func (s *SendTaskImpl) BpmnElementType() string       { return BPMN_ELEMENT_SEND_TASK }
func (s *SendTaskImpl) GetImplementation() string     { return AttrSendTaskImplementation.Get(s) }
func (s *SendTaskImpl) SetImplementation(impl string) { AttrSendTaskImplementation.Set(s, impl) }

type ReceiveTaskImpl struct {
	TaskImpl
}

func (*ReceiveTaskImpl) implReceiveTask()                {}
func (r *ReceiveTaskImpl) BpmnElementType() string       { return BPMN_ELEMENT_RECEIVE_TASK }
func (r *ReceiveTaskImpl) GetImplementation() string     { return AttrReceiveTaskImplementation.Get(r) }
func (r *ReceiveTaskImpl) SetImplementation(impl string) { AttrReceiveTaskImplementation.Set(r, impl) }

type ManualTaskImpl struct {
	TaskImpl
}

func (*ManualTaskImpl) implManualTask()           {}
func (m *ManualTaskImpl) BpmnElementType() string { return BPMN_ELEMENT_MANUAL_TASK }

type SubProcessImpl struct {
	ActivityImpl
}

func (*SubProcessImpl) implSubProcess()              {}
func (s *SubProcessImpl) BpmnElementType() string    { return BPMN_ELEMENT_SUB_PROCESS }
func (s *SubProcessImpl) IsTriggeredByEvent() bool   { return AttrSubProcessTriggeredByEvent.Get(s) }
func (s *SubProcessImpl) SetTriggeredByEvent(v bool) { AttrSubProcessTriggeredByEvent.Set(s, v) }

func (s *SubProcessImpl) GetFlowElements() []FlowElement {
	mi := s.GetModelInstance()
	if mi == nil {
		return nil
	}
	var result []FlowElement
	for _, child := range s.DomElement.GetChildElements() {
		if inst := mi.GetElementByDom(child); inst != nil {
			if fe, ok := inst.(FlowElement); ok {
				result = append(result, fe)
			}
		}
	}
	return result
}

func (s *SubProcessImpl) AddFlowElement(elem FlowElement) {
	s.DomElement.AppendChild(elem.GetDomElement())
	if mi := s.GetModelInstance(); mi != nil {
		mi.RegisterElement(elem)
	}
}

type AdHocSubProcessImpl struct {
	SubProcessImpl
}

func (*AdHocSubProcessImpl) implAdHocSubProcess()      {}
func (a *AdHocSubProcessImpl) BpmnElementType() string { return BPMN_ELEMENT_AD_HOC_SUB_PROCESS }

func (a *AdHocSubProcessImpl) GetCompletionCondition() string {
	children := a.DomElement.GetChildElementsByNS(BPMN20_NS, BPMN_ELEMENT_COMPLETION_CONDITION)
	if len(children) == 0 {
		return ""
	}
	return children[0].GetTextContent()
}

func (a *AdHocSubProcessImpl) SetCompletionCondition(expr string) {
	// Remove existing completionCondition children
	children := a.DomElement.GetChildElementsByNS(BPMN20_NS, BPMN_ELEMENT_COMPLETION_CONDITION)
	for _, child := range children {
		a.DomElement.RemoveChild(child)
	}
	if expr != "" {
		mi := a.GetModelInstance()
		doc := mi.GetDocument()
		condElem := doc.CreateElement(BPMN20_NS, BPMN_ELEMENT_COMPLETION_CONDITION)
		condElem.SetTextContent(expr)
		a.DomElement.AppendChild(condElem)
	}
}

func (a *AdHocSubProcessImpl) GetCancelRemainingInstances() bool {
	return AttrCancelRemainingInstances.Get(a)
}

func (a *AdHocSubProcessImpl) SetCancelRemainingInstances(v bool) {
	AttrCancelRemainingInstances.Set(a, v)
}

type CallActivityImpl struct {
	ActivityImpl
}

func (*CallActivityImpl) implCallActivity()              {}
func (c *CallActivityImpl) BpmnElementType() string      { return BPMN_ELEMENT_CALL_ACTIVITY }
func (c *CallActivityImpl) GetCalledElement() string     { return AttrCallActivityCalledElement.Get(c) }
func (c *CallActivityImpl) SetCalledElement(elem string) { AttrCallActivityCalledElement.Set(c, elem) }

// --- Registration ---

func registerActivityType(mb *ModelBuilder) {
	tb := mb.DefineType((*Activity)(nil), BPMN_ELEMENT_ACTIVITY).
		Namespace(BPMN20_NS).
		ExtendsType((*FlowNode)(nil)).
		AbstractType()

	AttrActivityIsForCompensation = tb.BoolAttribute(BPMN_ATTRIBUTE_IS_FOR_COMPENSATION).DefaultValue(false).Build()
	AttrActivityDefaultRef = tb.StringAttribute(BPMN_ATTRIBUTE_DEFAULT).Build()
}

func registerTaskType(mb *ModelBuilder) {
	mb.DefineType((*Task)(nil), BPMN_ELEMENT_TASK).
		Namespace(BPMN20_NS).
		ExtendsType((*Activity)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &TaskImpl{ActivityImpl{FlowNodeImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}}}
		})
}

func registerServiceTaskType(mb *ModelBuilder) {
	tb := mb.DefineType((*ServiceTask)(nil), BPMN_ELEMENT_SERVICE_TASK).
		Namespace(BPMN20_NS).
		ExtendsType((*Task)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ServiceTaskImpl{TaskImpl{ActivityImpl{FlowNodeImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}}}}
		})

	AttrServiceTaskImplementation = tb.StringAttribute(BPMN_ATTRIBUTE_IMPLEMENTATION).Build()
}

func registerUserTaskType(mb *ModelBuilder) {
	tb := mb.DefineType((*UserTask)(nil), BPMN_ELEMENT_USER_TASK).
		Namespace(BPMN20_NS).
		ExtendsType((*Task)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &UserTaskImpl{TaskImpl{ActivityImpl{FlowNodeImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}}}}
		})

	AttrUserTaskImplementation = tb.StringAttribute(BPMN_ATTRIBUTE_IMPLEMENTATION).Build()
}

func registerScriptTaskType(mb *ModelBuilder) {
	tb := mb.DefineType((*ScriptTask)(nil), BPMN_ELEMENT_SCRIPT_TASK).
		Namespace(BPMN20_NS).
		ExtendsType((*Task)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ScriptTaskImpl{TaskImpl{ActivityImpl{FlowNodeImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}}}}
		})

	AttrScriptTaskScriptFormat = tb.StringAttribute(BPMN_ATTRIBUTE_SCRIPT_FORMAT).Build()
}

func registerBusinessRuleTaskType(mb *ModelBuilder) {
	tb := mb.DefineType((*BusinessRuleTask)(nil), BPMN_ELEMENT_BUSINESS_RULE_TASK).
		Namespace(BPMN20_NS).
		ExtendsType((*Task)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &BusinessRuleTaskImpl{TaskImpl{ActivityImpl{FlowNodeImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}}}}
		})

	AttrBusinessRuleTaskImpl = tb.StringAttribute(BPMN_ATTRIBUTE_IMPLEMENTATION).Build()
}

func registerSendTaskType(mb *ModelBuilder) {
	tb := mb.DefineType((*SendTask)(nil), BPMN_ELEMENT_SEND_TASK).
		Namespace(BPMN20_NS).
		ExtendsType((*Task)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &SendTaskImpl{TaskImpl{ActivityImpl{FlowNodeImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}}}}
		})

	AttrSendTaskImplementation = tb.StringAttribute(BPMN_ATTRIBUTE_IMPLEMENTATION).Build()
}

func registerReceiveTaskType(mb *ModelBuilder) {
	tb := mb.DefineType((*ReceiveTask)(nil), BPMN_ELEMENT_RECEIVE_TASK).
		Namespace(BPMN20_NS).
		ExtendsType((*Task)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ReceiveTaskImpl{TaskImpl{ActivityImpl{FlowNodeImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}}}}
		})

	AttrReceiveTaskImplementation = tb.StringAttribute(BPMN_ATTRIBUTE_IMPLEMENTATION).Build()
}

func registerManualTaskType(mb *ModelBuilder) {
	mb.DefineType((*ManualTask)(nil), BPMN_ELEMENT_MANUAL_TASK).
		Namespace(BPMN20_NS).
		ExtendsType((*Task)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ManualTaskImpl{TaskImpl{ActivityImpl{FlowNodeImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}}}}
		})
}

func registerSubProcessType(mb *ModelBuilder) {
	tb := mb.DefineType((*SubProcess)(nil), BPMN_ELEMENT_SUB_PROCESS).
		Namespace(BPMN20_NS).
		ExtendsType((*Activity)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &SubProcessImpl{ActivityImpl{FlowNodeImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}}}
		})

	AttrSubProcessTriggeredByEvent = tb.BoolAttribute(BPMN_ATTRIBUTE_TRIGGERED_BY_EVENT).DefaultValue(false).Build()
}

func registerAdHocSubProcessType(mb *ModelBuilder) {
	tb := mb.DefineType((*AdHocSubProcess)(nil), BPMN_ELEMENT_AD_HOC_SUB_PROCESS).
		Namespace(BPMN20_NS).
		ExtendsType((*SubProcess)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &AdHocSubProcessImpl{SubProcessImpl{ActivityImpl{FlowNodeImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}}}}
		})

	AttrCancelRemainingInstances = tb.BoolAttribute("cancelRemainingInstances").DefaultValue(true).Build()
}

func registerCallActivityType(mb *ModelBuilder) {
	tb := mb.DefineType((*CallActivity)(nil), BPMN_ELEMENT_CALL_ACTIVITY).
		Namespace(BPMN20_NS).
		ExtendsType((*Activity)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &CallActivityImpl{ActivityImpl{FlowNodeImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}}}
		})

	AttrCallActivityCalledElement = tb.StringAttribute(BPMN_ATTRIBUTE_CALLED_ELEMENT).Build()
}
