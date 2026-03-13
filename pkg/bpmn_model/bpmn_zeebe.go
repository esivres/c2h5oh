package bpmn_model

// --- Interfaces ---

// ZeebeTaskDefinition defines a Zeebe job worker task (type, retries).
type ZeebeTaskDefinition interface {
	ModelElementInstance
	implZeebeTaskDefinition()
	GetType() string
	SetType(t string)
	GetRetries() string
	SetRetries(r string)
}

// ZeebeIoMapping is a container for Zeebe input/output mappings.
type ZeebeIoMapping interface {
	ModelElementInstance
	implZeebeIoMapping()
	GetInputs() []ZeebeInput
	GetOutputs() []ZeebeOutput
}

// ZeebeInput defines an input mapping expression.
type ZeebeInput interface {
	ModelElementInstance
	implZeebeInput()
	GetSource() string
	SetSource(s string)
	GetTarget() string
	SetTarget(t string)
}

// ZeebeOutput defines an output mapping expression.
type ZeebeOutput interface {
	ModelElementInstance
	implZeebeOutput()
	GetSource() string
	SetSource(s string)
	GetTarget() string
	SetTarget(t string)
}

// ZeebeTaskHeaders is a container for task headers.
type ZeebeTaskHeaders interface {
	ModelElementInstance
	implZeebeTaskHeaders()
	GetHeaders() []ZeebeTaskHeader
}

// ZeebeTaskHeader defines a single task header key-value pair.
type ZeebeTaskHeader interface {
	ModelElementInstance
	implZeebeTaskHeader()
	GetKey() string
	SetKey(k string)
	GetValue() string
	SetValue(v string)
}

// ZeebeSubscription defines a message subscription correlation key.
type ZeebeSubscription interface {
	ModelElementInstance
	implZeebeSubscription()
	GetCorrelationKey() string
	SetCorrelationKey(k string)
}

// ZeebeCalledElement defines a called process for CallActivity.
type ZeebeCalledElement interface {
	ModelElementInstance
	implZeebeCalledElement()
	GetProcessId() string
	SetProcessId(id string)
	GetPropagateAllChildVariables() bool
	SetPropagateAllChildVariables(v bool)
	GetPropagateAllParentVariables() bool
	SetPropagateAllParentVariables(v bool)
}

// ZeebeCalledDecision defines a called DMN decision.
type ZeebeCalledDecision interface {
	ModelElementInstance
	implZeebeCalledDecision()
	GetDecisionId() string
	SetDecisionId(id string)
	GetResultVariable() string
	SetResultVariable(v string)
}

// ZeebeFormDefinition defines a form reference for user tasks.
type ZeebeFormDefinition interface {
	ModelElementInstance
	implZeebeFormDefinition()
	GetFormId() string
	SetFormId(id string)
	GetFormKey() string
	SetFormKey(key string)
}

// ZeebeUserTask marks a user task for Zeebe user task behavior.
type ZeebeUserTask interface {
	ModelElementInstance
	implZeebeUserTask()
}

// ZeebeAssignmentDefinition defines assignee and candidates for user tasks.
type ZeebeAssignmentDefinition interface {
	ModelElementInstance
	implZeebeAssignmentDefinition()
	GetAssignee() string
	SetAssignee(a string)
	GetCandidateGroups() string
	SetCandidateGroups(g string)
	GetCandidateUsers() string
	SetCandidateUsers(u string)
}

// ZeebeExecutionListeners is a container for execution listeners.
type ZeebeExecutionListeners interface {
	ModelElementInstance
	implZeebeExecutionListeners()
	GetListeners() []ZeebeExecutionListener
}

// ZeebeExecutionListener defines a single execution listener.
type ZeebeExecutionListener interface {
	ModelElementInstance
	implZeebeExecutionListener()
	GetEventType() string
	SetEventType(t string)
	GetType() string
	SetType(t string)
	GetRetries() string
	SetRetries(r string)
}

// ZeebeProperties is a container for custom properties.
type ZeebeProperties interface {
	ModelElementInstance
	implZeebeProperties()
	GetProperties() []ZeebeProperty
}

// ZeebeProperty defines a single custom property.
type ZeebeProperty interface {
	ModelElementInstance
	implZeebeProperty()
	GetName() string
	SetName(n string)
	GetValue() string
	SetValue(v string)
}

// ZeebeLoopCharacteristics defines multi-instance loop settings.
type ZeebeLoopCharacteristics interface {
	ModelElementInstance
	implZeebeLoopCharacteristics()
	GetInputCollection() string
	SetInputCollection(c string)
	GetInputElement() string
	SetInputElement(e string)
	GetOutputCollection() string
	SetOutputCollection(c string)
	GetOutputElement() string
	SetOutputElement(e string)
}

// ZeebeScript defines a Zeebe script task expression.
type ZeebeScript interface {
	ModelElementInstance
	implZeebeScript()
	GetExpression() string
	SetExpression(e string)
	GetResultVariable() string
	SetResultVariable(v string)
}

// ZeebePublishMessage defines message publishing parameters.
type ZeebePublishMessage interface {
	ModelElementInstance
	implZeebePublishMessage()
	GetCorrelationKey() string
	SetCorrelationKey(k string)
	GetMessageId() string
	SetMessageId(id string)
}

// ZeebeTaskSchedule defines due date and follow-up date for user tasks.
type ZeebeTaskSchedule interface {
	ModelElementInstance
	implZeebeTaskSchedule()
	GetDueDate() string
	SetDueDate(d string)
	GetFollowUpDate() string
	SetFollowUpDate(d string)
}

// ZeebeAdHoc defines ad-hoc sub-process configuration for Zeebe.
type ZeebeAdHoc interface {
	ModelElementInstance
	implZeebeAdHoc()
	GetActiveElementsCollection() string
	SetActiveElementsCollection(e string)
	GetOutputCollection() string
	SetOutputCollection(c string)
	GetOutputElement() string
	SetOutputElement(e string)
}

// --- Attribute descriptors ---

var (
	AttrZeebeTaskDefType    *Attribute[string]
	AttrZeebeTaskDefRetries *Attribute[string]
	AttrZeebeSource         *Attribute[string]
	AttrZeebeTarget         *Attribute[string]
	AttrZeebeHeaderKey      *Attribute[string]
	AttrZeebeHeaderValue    *Attribute[string]
	AttrZeebeCorrelationKey *Attribute[string]

	AttrZeebeProcessId                   *Attribute[string]
	AttrZeebePropagateAllChildVariables  *Attribute[bool]
	AttrZeebePropagateAllParentVariables *Attribute[bool]
	AttrZeebeDecisionId                  *Attribute[string]
	AttrZeebeResultVariable              *Attribute[string]
	AttrZeebeFormId                      *Attribute[string]
	AttrZeebeFormKey                     *Attribute[string]
	AttrZeebeAssignee                    *Attribute[string]
	AttrZeebeCandidateGroups             *Attribute[string]
	AttrZeebeCandidateUsers              *Attribute[string]

	AttrZeebeListenerEventType *Attribute[string]
	AttrZeebeListenerType      *Attribute[string]
	AttrZeebeListenerRetries   *Attribute[string]

	AttrZeebePropertyName  *Attribute[string]
	AttrZeebePropertyValue *Attribute[string]

	AttrZeebeInputCollection  *Attribute[string]
	AttrZeebeInputElement     *Attribute[string]
	AttrZeebeOutputCollection *Attribute[string]
	AttrZeebeOutputElement    *Attribute[string]

	AttrZeebeExpression               *Attribute[string]
	AttrZeebeScriptResultVar          *Attribute[string]
	AttrZeebePubMsgCorrelationKey     *Attribute[string]
	AttrZeebePubMsgMessageId          *Attribute[string]
	AttrZeebeDueDate                  *Attribute[string]
	AttrZeebeFollowUpDate             *Attribute[string]
	AttrZeebeActiveElementsCollection *Attribute[string]
)

// --- Implementations ---

type ZeebeTaskDefinitionImpl struct{ BaseInstance }

func (*ZeebeTaskDefinitionImpl) implZeebeTaskDefinition() {}
func (z *ZeebeTaskDefinitionImpl) GetType() string        { return AttrZeebeTaskDefType.Get(z) }
func (z *ZeebeTaskDefinitionImpl) SetType(t string)       { AttrZeebeTaskDefType.Set(z, t) }
func (z *ZeebeTaskDefinitionImpl) GetRetries() string     { return AttrZeebeTaskDefRetries.Get(z) }
func (z *ZeebeTaskDefinitionImpl) SetRetries(r string)    { AttrZeebeTaskDefRetries.Set(z, r) }

type ZeebeIoMappingImpl struct{ BaseInstance }

func (*ZeebeIoMappingImpl) implZeebeIoMapping() {}
func (z *ZeebeIoMappingImpl) GetInputs() []ZeebeInput {
	return getZeebeChildren[ZeebeInput](z, ZEEBE_ELEMENT_INPUT)
}
func (z *ZeebeIoMappingImpl) GetOutputs() []ZeebeOutput {
	return getZeebeChildren[ZeebeOutput](z, ZEEBE_ELEMENT_OUTPUT)
}

type ZeebeInputImpl struct{ BaseInstance }

func (*ZeebeInputImpl) implZeebeInput()      {}
func (z *ZeebeInputImpl) GetSource() string  { return AttrZeebeSource.Get(z) }
func (z *ZeebeInputImpl) SetSource(s string) { AttrZeebeSource.Set(z, s) }
func (z *ZeebeInputImpl) GetTarget() string  { return AttrZeebeTarget.Get(z) }
func (z *ZeebeInputImpl) SetTarget(t string) { AttrZeebeTarget.Set(z, t) }

type ZeebeOutputImpl struct{ BaseInstance }

func (*ZeebeOutputImpl) implZeebeOutput()     {}
func (z *ZeebeOutputImpl) GetSource() string  { return AttrZeebeSource.Get(z) }
func (z *ZeebeOutputImpl) SetSource(s string) { AttrZeebeSource.Set(z, s) }
func (z *ZeebeOutputImpl) GetTarget() string  { return AttrZeebeTarget.Get(z) }
func (z *ZeebeOutputImpl) SetTarget(t string) { AttrZeebeTarget.Set(z, t) }

type ZeebeTaskHeadersImpl struct{ BaseInstance }

func (*ZeebeTaskHeadersImpl) implZeebeTaskHeaders() {}
func (z *ZeebeTaskHeadersImpl) GetHeaders() []ZeebeTaskHeader {
	return getZeebeChildren[ZeebeTaskHeader](z, ZEEBE_ELEMENT_HEADER)
}

type ZeebeTaskHeaderImpl struct{ BaseInstance }

func (*ZeebeTaskHeaderImpl) implZeebeTaskHeader() {}
func (z *ZeebeTaskHeaderImpl) GetKey() string     { return AttrZeebeHeaderKey.Get(z) }
func (z *ZeebeTaskHeaderImpl) SetKey(k string)    { AttrZeebeHeaderKey.Set(z, k) }
func (z *ZeebeTaskHeaderImpl) GetValue() string   { return AttrZeebeHeaderValue.Get(z) }
func (z *ZeebeTaskHeaderImpl) SetValue(v string)  { AttrZeebeHeaderValue.Set(z, v) }

type ZeebeSubscriptionImpl struct{ BaseInstance }

func (*ZeebeSubscriptionImpl) implZeebeSubscription()       {}
func (z *ZeebeSubscriptionImpl) GetCorrelationKey() string  { return AttrZeebeCorrelationKey.Get(z) }
func (z *ZeebeSubscriptionImpl) SetCorrelationKey(k string) { AttrZeebeCorrelationKey.Set(z, k) }

type ZeebeCalledElementImpl struct{ BaseInstance }

func (*ZeebeCalledElementImpl) implZeebeCalledElement()  {}
func (z *ZeebeCalledElementImpl) GetProcessId() string   { return AttrZeebeProcessId.Get(z) }
func (z *ZeebeCalledElementImpl) SetProcessId(id string) { AttrZeebeProcessId.Set(z, id) }
func (z *ZeebeCalledElementImpl) GetPropagateAllChildVariables() bool {
	return AttrZeebePropagateAllChildVariables.Get(z)
}
func (z *ZeebeCalledElementImpl) SetPropagateAllChildVariables(v bool) {
	AttrZeebePropagateAllChildVariables.Set(z, v)
}
func (z *ZeebeCalledElementImpl) GetPropagateAllParentVariables() bool {
	return AttrZeebePropagateAllParentVariables.Get(z)
}
func (z *ZeebeCalledElementImpl) SetPropagateAllParentVariables(v bool) {
	AttrZeebePropagateAllParentVariables.Set(z, v)
}

type ZeebeCalledDecisionImpl struct{ BaseInstance }

func (*ZeebeCalledDecisionImpl) implZeebeCalledDecision()     {}
func (z *ZeebeCalledDecisionImpl) GetDecisionId() string      { return AttrZeebeDecisionId.Get(z) }
func (z *ZeebeCalledDecisionImpl) SetDecisionId(id string)    { AttrZeebeDecisionId.Set(z, id) }
func (z *ZeebeCalledDecisionImpl) GetResultVariable() string  { return AttrZeebeResultVariable.Get(z) }
func (z *ZeebeCalledDecisionImpl) SetResultVariable(v string) { AttrZeebeResultVariable.Set(z, v) }

type ZeebeFormDefinitionImpl struct{ BaseInstance }

func (*ZeebeFormDefinitionImpl) implZeebeFormDefinition() {}
func (z *ZeebeFormDefinitionImpl) GetFormId() string      { return AttrZeebeFormId.Get(z) }
func (z *ZeebeFormDefinitionImpl) SetFormId(id string)    { AttrZeebeFormId.Set(z, id) }
func (z *ZeebeFormDefinitionImpl) GetFormKey() string     { return AttrZeebeFormKey.Get(z) }
func (z *ZeebeFormDefinitionImpl) SetFormKey(key string)  { AttrZeebeFormKey.Set(z, key) }

type ZeebeUserTaskImpl struct{ BaseInstance }

func (*ZeebeUserTaskImpl) implZeebeUserTask() {}

type ZeebeAssignmentDefinitionImpl struct{ BaseInstance }

func (*ZeebeAssignmentDefinitionImpl) implZeebeAssignmentDefinition() {}
func (z *ZeebeAssignmentDefinitionImpl) GetAssignee() string          { return AttrZeebeAssignee.Get(z) }
func (z *ZeebeAssignmentDefinitionImpl) SetAssignee(a string)         { AttrZeebeAssignee.Set(z, a) }
func (z *ZeebeAssignmentDefinitionImpl) GetCandidateGroups() string {
	return AttrZeebeCandidateGroups.Get(z)
}
func (z *ZeebeAssignmentDefinitionImpl) SetCandidateGroups(g string) {
	AttrZeebeCandidateGroups.Set(z, g)
}
func (z *ZeebeAssignmentDefinitionImpl) GetCandidateUsers() string {
	return AttrZeebeCandidateUsers.Get(z)
}
func (z *ZeebeAssignmentDefinitionImpl) SetCandidateUsers(u string) {
	AttrZeebeCandidateUsers.Set(z, u)
}

type ZeebeExecutionListenersImpl struct{ BaseInstance }

func (*ZeebeExecutionListenersImpl) implZeebeExecutionListeners() {}
func (z *ZeebeExecutionListenersImpl) GetListeners() []ZeebeExecutionListener {
	return getZeebeChildren[ZeebeExecutionListener](z, ZEEBE_ELEMENT_EXECUTION_LISTENER)
}

type ZeebeExecutionListenerImpl struct{ BaseInstance }

func (*ZeebeExecutionListenerImpl) implZeebeExecutionListener() {}
func (z *ZeebeExecutionListenerImpl) GetEventType() string      { return AttrZeebeListenerEventType.Get(z) }
func (z *ZeebeExecutionListenerImpl) SetEventType(t string)     { AttrZeebeListenerEventType.Set(z, t) }
func (z *ZeebeExecutionListenerImpl) GetType() string           { return AttrZeebeListenerType.Get(z) }
func (z *ZeebeExecutionListenerImpl) SetType(t string)          { AttrZeebeListenerType.Set(z, t) }
func (z *ZeebeExecutionListenerImpl) GetRetries() string        { return AttrZeebeListenerRetries.Get(z) }
func (z *ZeebeExecutionListenerImpl) SetRetries(r string)       { AttrZeebeListenerRetries.Set(z, r) }

type ZeebePropertiesImpl struct{ BaseInstance }

func (*ZeebePropertiesImpl) implZeebeProperties() {}
func (z *ZeebePropertiesImpl) GetProperties() []ZeebeProperty {
	return getZeebeChildren[ZeebeProperty](z, ZEEBE_ELEMENT_PROPERTY)
}

type ZeebePropertyImpl struct{ BaseInstance }

func (*ZeebePropertyImpl) implZeebeProperty()  {}
func (z *ZeebePropertyImpl) GetName() string   { return AttrZeebePropertyName.Get(z) }
func (z *ZeebePropertyImpl) SetName(n string)  { AttrZeebePropertyName.Set(z, n) }
func (z *ZeebePropertyImpl) GetValue() string  { return AttrZeebePropertyValue.Get(z) }
func (z *ZeebePropertyImpl) SetValue(v string) { AttrZeebePropertyValue.Set(z, v) }

type ZeebeLoopCharacteristicsImpl struct{ BaseInstance }

func (*ZeebeLoopCharacteristicsImpl) implZeebeLoopCharacteristics() {}
func (z *ZeebeLoopCharacteristicsImpl) GetInputCollection() string {
	return AttrZeebeInputCollection.Get(z)
}
func (z *ZeebeLoopCharacteristicsImpl) SetInputCollection(c string) {
	AttrZeebeInputCollection.Set(z, c)
}
func (z *ZeebeLoopCharacteristicsImpl) GetInputElement() string  { return AttrZeebeInputElement.Get(z) }
func (z *ZeebeLoopCharacteristicsImpl) SetInputElement(e string) { AttrZeebeInputElement.Set(z, e) }
func (z *ZeebeLoopCharacteristicsImpl) GetOutputCollection() string {
	return AttrZeebeOutputCollection.Get(z)
}
func (z *ZeebeLoopCharacteristicsImpl) SetOutputCollection(c string) {
	AttrZeebeOutputCollection.Set(z, c)
}
func (z *ZeebeLoopCharacteristicsImpl) GetOutputElement() string {
	return AttrZeebeOutputElement.Get(z)
}
func (z *ZeebeLoopCharacteristicsImpl) SetOutputElement(e string) { AttrZeebeOutputElement.Set(z, e) }

type ZeebeScriptImpl struct{ BaseInstance }

func (*ZeebeScriptImpl) implZeebeScript()             {}
func (z *ZeebeScriptImpl) GetExpression() string      { return AttrZeebeExpression.Get(z) }
func (z *ZeebeScriptImpl) SetExpression(e string)     { AttrZeebeExpression.Set(z, e) }
func (z *ZeebeScriptImpl) GetResultVariable() string  { return AttrZeebeScriptResultVar.Get(z) }
func (z *ZeebeScriptImpl) SetResultVariable(v string) { AttrZeebeScriptResultVar.Set(z, v) }

type ZeebePublishMessageImpl struct{ BaseInstance }

func (*ZeebePublishMessageImpl) implZeebePublishMessage() {}
func (z *ZeebePublishMessageImpl) GetCorrelationKey() string {
	return AttrZeebePubMsgCorrelationKey.Get(z)
}
func (z *ZeebePublishMessageImpl) SetCorrelationKey(k string) {
	AttrZeebePubMsgCorrelationKey.Set(z, k)
}
func (z *ZeebePublishMessageImpl) GetMessageId() string   { return AttrZeebePubMsgMessageId.Get(z) }
func (z *ZeebePublishMessageImpl) SetMessageId(id string) { AttrZeebePubMsgMessageId.Set(z, id) }

type ZeebeTaskScheduleImpl struct{ BaseInstance }

func (*ZeebeTaskScheduleImpl) implZeebeTaskSchedule()     {}
func (z *ZeebeTaskScheduleImpl) GetDueDate() string       { return AttrZeebeDueDate.Get(z) }
func (z *ZeebeTaskScheduleImpl) SetDueDate(d string)      { AttrZeebeDueDate.Set(z, d) }
func (z *ZeebeTaskScheduleImpl) GetFollowUpDate() string  { return AttrZeebeFollowUpDate.Get(z) }
func (z *ZeebeTaskScheduleImpl) SetFollowUpDate(d string) { AttrZeebeFollowUpDate.Set(z, d) }

type ZeebeAdHocImpl struct{ BaseInstance }

func (*ZeebeAdHocImpl) implZeebeAdHoc() {}
func (z *ZeebeAdHocImpl) GetActiveElementsCollection() string {
	return AttrZeebeActiveElementsCollection.Get(z)
}
func (z *ZeebeAdHocImpl) SetActiveElementsCollection(e string) {
	AttrZeebeActiveElementsCollection.Set(z, e)
}
func (z *ZeebeAdHocImpl) GetOutputCollection() string  { return AttrZeebeOutputCollection.Get(z) }
func (z *ZeebeAdHocImpl) SetOutputCollection(c string) { AttrZeebeOutputCollection.Set(z, c) }
func (z *ZeebeAdHocImpl) GetOutputElement() string     { return AttrZeebeOutputElement.Get(z) }
func (z *ZeebeAdHocImpl) SetOutputElement(e string)    { AttrZeebeOutputElement.Set(z, e) }

// --- Helper: get typed children of a Zeebe container ---

func getZeebeChildren[T ModelElementInstance](parent ModelElementInstance, localName string) []T {
	mi := parent.GetModelInstance()
	if mi == nil {
		return nil
	}
	var result []T
	for _, child := range parent.GetDomElement().GetChildElementsByNS(ZEEBE_NS, localName) {
		if inst := mi.GetElementByDom(child); inst != nil {
			if typed, ok := inst.(T); ok {
				result = append(result, typed)
			}
		}
	}
	return result
}

// GetSingleExtensionElement returns the first extension element of the given type.
func GetSingleExtensionElement[T ModelElementInstance](element BaseElement) (T, bool) {
	ee := element.GetExtensionElements()
	if ee == nil {
		var zero T
		return zero, false
	}
	for _, child := range ee.GetElements() {
		if typed, ok := child.(T); ok {
			return typed, true
		}
	}
	var zero T
	return zero, false
}

// getOrCreateExtElement gets or creates a single extension element of the given type.
func getOrCreateExtElement[T ModelElementInstance](mi *ModelInstance, element BaseElement, ns string, localName string) T {
	ee := element.GetOrCreateExtensionElements()
	// Check existing children
	for _, child := range ee.GetDomElement().GetChildElementsByNS(ns, localName) {
		if inst := mi.GetElementByDom(child); inst != nil {
			if typed, ok := inst.(T); ok {
				return typed
			}
		}
	}
	// Create new
	elemType := bpmnModel.GetTypeByQName(ns, localName)
	inst, _ := mi.NewInstance(elemType)
	ee.GetDomElement().AppendChild(inst.GetDomElement())
	return inst.(T)
}

// --- Registration ---

func registerZeebeTypes(mb *ModelBuilder) {
	registerZeebeTaskDefinitionType(mb)
	registerZeebeIoMappingType(mb)
	registerZeebeInputType(mb)
	registerZeebeOutputType(mb)
	registerZeebeTaskHeadersType(mb)
	registerZeebeTaskHeaderType(mb)
	registerZeebeSubscriptionType(mb)
	registerZeebeCalledElementType(mb)
	registerZeebeCalledDecisionType(mb)
	registerZeebeFormDefinitionType(mb)
	registerZeebeUserTaskType(mb)
	registerZeebeAssignmentDefinitionType(mb)
	registerZeebeExecutionListenersType(mb)
	registerZeebeExecutionListenerType(mb)
	registerZeebePropertiesType(mb)
	registerZeebePropertyType(mb)
	registerZeebeLoopCharacteristicsType(mb)
	registerZeebeScriptType(mb)
	registerZeebePublishMessageType(mb)
	registerZeebeTaskScheduleType(mb)
	registerZeebeAdHocType(mb)
}

func registerZeebeTaskDefinitionType(mb *ModelBuilder) {
	tb := mb.DefineType((*ZeebeTaskDefinition)(nil), ZEEBE_ELEMENT_TASK_DEFINITION).
		Namespace(ZEEBE_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ZeebeTaskDefinitionImpl{NewBaseInstance(ctx)}
		})
	AttrZeebeTaskDefType = tb.StringAttribute(ZEEBE_ATTRIBUTE_TYPE).Build()
	AttrZeebeTaskDefRetries = tb.StringAttribute(ZEEBE_ATTRIBUTE_RETRIES).DefaultValue("3").Build()
}

func registerZeebeIoMappingType(mb *ModelBuilder) {
	mb.DefineType((*ZeebeIoMapping)(nil), ZEEBE_ELEMENT_IO_MAPPING).
		Namespace(ZEEBE_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ZeebeIoMappingImpl{NewBaseInstance(ctx)}
		})
}

func registerZeebeInputType(mb *ModelBuilder) {
	tb := mb.DefineType((*ZeebeInput)(nil), ZEEBE_ELEMENT_INPUT).
		Namespace(ZEEBE_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ZeebeInputImpl{NewBaseInstance(ctx)}
		})
	AttrZeebeSource = tb.StringAttribute(ZEEBE_ATTRIBUTE_SOURCE).Build()
	AttrZeebeTarget = tb.StringAttribute(ZEEBE_ATTRIBUTE_TARGET).Build()
}

func registerZeebeOutputType(mb *ModelBuilder) {
	mb.DefineType((*ZeebeOutput)(nil), ZEEBE_ELEMENT_OUTPUT).
		Namespace(ZEEBE_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ZeebeOutputImpl{NewBaseInstance(ctx)}
		})
	// Reuses AttrZeebeSource/AttrZeebeTarget from ZeebeInput
}

func registerZeebeTaskHeadersType(mb *ModelBuilder) {
	mb.DefineType((*ZeebeTaskHeaders)(nil), ZEEBE_ELEMENT_TASK_HEADERS).
		Namespace(ZEEBE_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ZeebeTaskHeadersImpl{NewBaseInstance(ctx)}
		})
}

func registerZeebeTaskHeaderType(mb *ModelBuilder) {
	tb := mb.DefineType((*ZeebeTaskHeader)(nil), ZEEBE_ELEMENT_HEADER).
		Namespace(ZEEBE_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ZeebeTaskHeaderImpl{NewBaseInstance(ctx)}
		})
	AttrZeebeHeaderKey = tb.StringAttribute(ZEEBE_ATTRIBUTE_KEY).Build()
	AttrZeebeHeaderValue = tb.StringAttribute(ZEEBE_ATTRIBUTE_VALUE).Build()
}

func registerZeebeSubscriptionType(mb *ModelBuilder) {
	tb := mb.DefineType((*ZeebeSubscription)(nil), ZEEBE_ELEMENT_SUBSCRIPTION).
		Namespace(ZEEBE_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ZeebeSubscriptionImpl{NewBaseInstance(ctx)}
		})
	AttrZeebeCorrelationKey = tb.StringAttribute(ZEEBE_ATTRIBUTE_CORRELATION_KEY).Build()
}

func registerZeebeCalledElementType(mb *ModelBuilder) {
	tb := mb.DefineType((*ZeebeCalledElement)(nil), ZEEBE_ELEMENT_CALLED_ELEMENT).
		Namespace(ZEEBE_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ZeebeCalledElementImpl{NewBaseInstance(ctx)}
		})
	AttrZeebeProcessId = tb.StringAttribute(ZEEBE_ATTRIBUTE_PROCESS_ID).Build()
	AttrZeebePropagateAllChildVariables = tb.BoolAttribute(ZEEBE_ATTRIBUTE_PROPAGATE_ALL_CHILD_VARIABLES).DefaultValue(true).Build()
	AttrZeebePropagateAllParentVariables = tb.BoolAttribute(ZEEBE_ATTRIBUTE_PROPAGATE_ALL_PARENT_VARIABLES).DefaultValue(false).Build()
}

func registerZeebeCalledDecisionType(mb *ModelBuilder) {
	tb := mb.DefineType((*ZeebeCalledDecision)(nil), ZEEBE_ELEMENT_CALLED_DECISION).
		Namespace(ZEEBE_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ZeebeCalledDecisionImpl{NewBaseInstance(ctx)}
		})
	AttrZeebeDecisionId = tb.StringAttribute(ZEEBE_ATTRIBUTE_DECISION_ID).Build()
	AttrZeebeResultVariable = tb.StringAttribute(ZEEBE_ATTRIBUTE_RESULT_VARIABLE).Build()
}

func registerZeebeFormDefinitionType(mb *ModelBuilder) {
	tb := mb.DefineType((*ZeebeFormDefinition)(nil), ZEEBE_ELEMENT_FORM_DEFINITION).
		Namespace(ZEEBE_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ZeebeFormDefinitionImpl{NewBaseInstance(ctx)}
		})
	AttrZeebeFormId = tb.StringAttribute(ZEEBE_ATTRIBUTE_FORM_ID).Build()
	AttrZeebeFormKey = tb.StringAttribute(ZEEBE_ATTRIBUTE_FORM_KEY).Build()
}

func registerZeebeUserTaskType(mb *ModelBuilder) {
	mb.DefineType((*ZeebeUserTask)(nil), ZEEBE_ELEMENT_USER_TASK).
		Namespace(ZEEBE_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ZeebeUserTaskImpl{NewBaseInstance(ctx)}
		})
}

func registerZeebeAssignmentDefinitionType(mb *ModelBuilder) {
	tb := mb.DefineType((*ZeebeAssignmentDefinition)(nil), ZEEBE_ELEMENT_ASSIGNMENT_DEFINITION).
		Namespace(ZEEBE_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ZeebeAssignmentDefinitionImpl{NewBaseInstance(ctx)}
		})
	AttrZeebeAssignee = tb.StringAttribute(ZEEBE_ATTRIBUTE_ASSIGNEE).Build()
	AttrZeebeCandidateGroups = tb.StringAttribute(ZEEBE_ATTRIBUTE_CANDIDATE_GROUPS).Build()
	AttrZeebeCandidateUsers = tb.StringAttribute(ZEEBE_ATTRIBUTE_CANDIDATE_USERS).Build()
}

func registerZeebeExecutionListenersType(mb *ModelBuilder) {
	mb.DefineType((*ZeebeExecutionListeners)(nil), ZEEBE_ELEMENT_EXECUTION_LISTENERS).
		Namespace(ZEEBE_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ZeebeExecutionListenersImpl{NewBaseInstance(ctx)}
		})
}

func registerZeebeExecutionListenerType(mb *ModelBuilder) {
	tb := mb.DefineType((*ZeebeExecutionListener)(nil), ZEEBE_ELEMENT_EXECUTION_LISTENER).
		Namespace(ZEEBE_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ZeebeExecutionListenerImpl{NewBaseInstance(ctx)}
		})
	AttrZeebeListenerEventType = tb.StringAttribute(ZEEBE_ATTRIBUTE_EVENT_TYPE).Build()
	AttrZeebeListenerType = tb.StringAttribute(ZEEBE_ATTRIBUTE_TYPE).Build()
	AttrZeebeListenerRetries = tb.StringAttribute(ZEEBE_ATTRIBUTE_RETRIES).Build()
}

func registerZeebePropertiesType(mb *ModelBuilder) {
	mb.DefineType((*ZeebeProperties)(nil), ZEEBE_ELEMENT_PROPERTIES).
		Namespace(ZEEBE_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ZeebePropertiesImpl{NewBaseInstance(ctx)}
		})
}

func registerZeebePropertyType(mb *ModelBuilder) {
	tb := mb.DefineType((*ZeebeProperty)(nil), ZEEBE_ELEMENT_PROPERTY).
		Namespace(ZEEBE_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ZeebePropertyImpl{NewBaseInstance(ctx)}
		})
	AttrZeebePropertyName = tb.StringAttribute(ZEEBE_ATTRIBUTE_NAME).Build()
	AttrZeebePropertyValue = tb.StringAttribute(ZEEBE_ATTRIBUTE_VALUE).Build()
}

func registerZeebeLoopCharacteristicsType(mb *ModelBuilder) {
	tb := mb.DefineType((*ZeebeLoopCharacteristics)(nil), ZEEBE_ELEMENT_LOOP_CHARACTERISTICS).
		Namespace(ZEEBE_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ZeebeLoopCharacteristicsImpl{NewBaseInstance(ctx)}
		})
	AttrZeebeInputCollection = tb.StringAttribute(ZEEBE_ATTRIBUTE_INPUT_COLLECTION).Build()
	AttrZeebeInputElement = tb.StringAttribute(ZEEBE_ATTRIBUTE_INPUT_ELEMENT).Build()
	AttrZeebeOutputCollection = tb.StringAttribute(ZEEBE_ATTRIBUTE_OUTPUT_COLLECTION).Build()
	AttrZeebeOutputElement = tb.StringAttribute(ZEEBE_ATTRIBUTE_OUTPUT_ELEMENT).Build()
}

func registerZeebeScriptType(mb *ModelBuilder) {
	tb := mb.DefineType((*ZeebeScript)(nil), ZEEBE_ELEMENT_SCRIPT).
		Namespace(ZEEBE_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ZeebeScriptImpl{NewBaseInstance(ctx)}
		})
	AttrZeebeExpression = tb.StringAttribute(ZEEBE_ATTRIBUTE_EXPRESSION).Build()
	AttrZeebeScriptResultVar = tb.StringAttribute(ZEEBE_ATTRIBUTE_RESULT_VARIABLE).Build()
}

func registerZeebePublishMessageType(mb *ModelBuilder) {
	tb := mb.DefineType((*ZeebePublishMessage)(nil), ZEEBE_ELEMENT_PUBLISH_MESSAGE).
		Namespace(ZEEBE_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ZeebePublishMessageImpl{NewBaseInstance(ctx)}
		})
	AttrZeebePubMsgCorrelationKey = tb.StringAttribute(ZEEBE_ATTRIBUTE_CORRELATION_KEY).Build()
	AttrZeebePubMsgMessageId = tb.StringAttribute(ZEEBE_ATTRIBUTE_MESSAGE_ID).Build()
}

func registerZeebeTaskScheduleType(mb *ModelBuilder) {
	tb := mb.DefineType((*ZeebeTaskSchedule)(nil), ZEEBE_ELEMENT_TASK_SCHEDULE).
		Namespace(ZEEBE_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ZeebeTaskScheduleImpl{NewBaseInstance(ctx)}
		})
	AttrZeebeDueDate = tb.StringAttribute(ZEEBE_ATTRIBUTE_DUE_DATE).Build()
	AttrZeebeFollowUpDate = tb.StringAttribute(ZEEBE_ATTRIBUTE_FOLLOW_UP_DATE).Build()
}

func registerZeebeAdHocType(mb *ModelBuilder) {
	tb := mb.DefineType((*ZeebeAdHoc)(nil), ZEEBE_ELEMENT_AD_HOC).
		Namespace(ZEEBE_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ZeebeAdHocImpl{NewBaseInstance(ctx)}
		})
	AttrZeebeActiveElementsCollection = tb.StringAttribute(ZEEBE_ATTRIBUTE_ACTIVE_ELEMENTS_COLLECTION).Build()
	// Reuses AttrZeebeOutputCollection/AttrZeebeOutputElement from ZeebeLoopCharacteristics
}
