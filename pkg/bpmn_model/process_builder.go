package bpmn_model

import (
	"fmt"
	"sync/atomic"
)

// --- ID Generation ---

var idSeq atomic.Int64

func nextId(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, idSeq.Add(1))
}

// --- Builder Context (shared state across builder chain) ---

type builderCtx struct {
	bmi          *BpmnModelInstance
	process      Process
	currentNode  FlowNode
	gatewayStack []Gateway
	condExpr     string
	defaultFlow  bool
	diGenerated  bool
}

func (ctx *builderCtx) createElement(bpmnElementName string, id string) ModelElementInstance {
	elemType := bpmnModel.GetTypeByQName(BPMN20_NS, bpmnElementName)
	inst, _ := ctx.bmi.NewInstance(elemType)
	if be, ok := inst.(BaseElement); ok {
		if id == "" {
			id = nextId(bpmnElementName)
		}
		be.SetId(id)
	}
	return inst
}

func (ctx *builderCtx) createFlowNodeAndConnect(bpmnElementName string, id string) FlowNode {
	inst := ctx.createElement(bpmnElementName, id)
	node := inst.(FlowNode)
	ctx.process.AddFlowElement(node)

	if ctx.currentNode != nil {
		ctx.createSequenceFlow(ctx.currentNode, node)
	}

	ctx.currentNode = node
	return node
}

func (ctx *builderCtx) createSequenceFlow(source, target FlowNode) SequenceFlow {
	mi := ctx.bmi.ModelInstance
	flowType := bpmnModel.GetTypeByQName(BPMN20_NS, BPMN_ELEMENT_SEQUENCE_FLOW)
	flowInst, _ := mi.NewInstance(flowType)
	flow := flowInst.(SequenceFlow)

	flowId := nextId("Flow")
	flow.SetId(flowId)
	flow.SetSourceRef(source.GetId())
	flow.SetTargetRef(target.GetId())

	// Add <bpmn:outgoing> to source
	addFlowRef(mi, source, BPMN_ELEMENT_OUTGOING, flowId)
	// Add <bpmn:incoming> to target
	addFlowRef(mi, target, BPMN_ELEMENT_INCOMING, flowId)

	// Apply pending condition expression
	if ctx.condExpr != "" {
		condType := bpmnModel.GetTypeByQName(BPMN20_NS, BPMN_ELEMENT_CONDITION_EXPRESSION)
		condInst, _ := mi.NewInstance(condType)
		cond := condInst.(ConditionExpression)
		cond.SetType("bpmn:tFormalExpression")
		cond.SetTextContent(ctx.condExpr)
		flow.SetConditionExpression(cond)
		ctx.condExpr = ""
	}

	// Apply pending default flow
	if ctx.defaultFlow {
		if gw, ok := source.(Gateway); ok {
			gw.SetDefaultFlowRef(flowId)
		} else if act, ok := source.(Activity); ok {
			act.SetDefaultFlowRef(flowId)
		}
		ctx.defaultFlow = false
	}

	ctx.process.AddFlowElement(flow)
	return flow
}

func addFlowRef(mi *ModelInstance, node FlowNode, elementName string, flowId string) {
	doc := mi.GetDocument()
	refElem := doc.CreateElement(BPMN20_NS, elementName)
	refElem.SetTextContent(flowId)
	node.GetDomElement().AppendChild(refElem)
}

// --- ProcessBuilder ---

// ProcessBuilder is the entry point for building a BPMN process via fluent API.
type ProcessBuilder struct {
	ctx *builderCtx
}

func (b *ProcessBuilder) StartEvent(id string) *StartEventBuilder {
	node := b.ctx.createFlowNodeAndConnect(BPMN_ELEMENT_START_EVENT, id)
	return &StartEventBuilder{AbstractBuilder: AbstractBuilder{ctx: b.ctx}, element: node.(StartEvent)}
}

func (b *ProcessBuilder) Done() *BpmnModelInstance {
	if !b.ctx.diGenerated {
		b.ctx.diGenerated = true
		generateDI(b.ctx.bmi, b.ctx.process)
	}
	return b.ctx.bmi
}

// --- AbstractBuilder (shared flow-creation and navigation methods) ---

// AbstractBuilder provides methods available on all flow node builders.
type AbstractBuilder struct {
	ctx *builderCtx
}

// ConditionExpression sets the condition for the next sequence flow.
func (b *AbstractBuilder) ConditionExpression(expr string) *AbstractBuilder {
	b.ctx.condExpr = expr
	return b
}

// DefaultFlow marks the next sequence flow as the default flow of the current gateway/activity.
func (b *AbstractBuilder) DefaultFlow() *AbstractBuilder {
	b.ctx.defaultFlow = true
	return b
}

// MoveToNode sets the current context to the node with the given ID.
func (b *AbstractBuilder) MoveToNode(id string) *AbstractBuilder {
	if node, ok := GetTypedElementById[FlowNode](b.ctx.bmi.ModelInstance, id); ok {
		b.ctx.currentNode = node
	}
	return b
}

// MoveToLastGateway sets the current context to the last created gateway.
func (b *AbstractBuilder) MoveToLastGateway() *AbstractBuilder {
	if len(b.ctx.gatewayStack) > 0 {
		last := b.ctx.gatewayStack[len(b.ctx.gatewayStack)-1]
		b.ctx.currentNode = last
	}
	return b
}

// ConnectTo creates a sequence flow from the current node to the node with the given ID.
func (b *AbstractBuilder) ConnectTo(id string) *AbstractBuilder {
	if target, ok := GetTypedElementById[FlowNode](b.ctx.bmi.ModelInstance, id); ok {
		if b.ctx.currentNode != nil {
			b.ctx.createSequenceFlow(b.ctx.currentNode, target)
		}
		b.ctx.currentNode = target
	}
	return b
}

// Done finalizes the builder, generates DI, and returns the model instance.
func (b *AbstractBuilder) Done() *BpmnModelInstance {
	if !b.ctx.diGenerated {
		b.ctx.diGenerated = true
		generateDI(b.ctx.bmi, b.ctx.process)
	}
	return b.ctx.bmi
}

// --- Flow node creation methods ---

func (b *AbstractBuilder) StartEvent(id string) *StartEventBuilder {
	node := b.ctx.createFlowNodeAndConnect(BPMN_ELEMENT_START_EVENT, id)
	return &StartEventBuilder{AbstractBuilder: AbstractBuilder{ctx: b.ctx}, element: node.(StartEvent)}
}

func (b *AbstractBuilder) EndEvent(id string) *EndEventBuilder {
	node := b.ctx.createFlowNodeAndConnect(BPMN_ELEMENT_END_EVENT, id)
	return &EndEventBuilder{AbstractBuilder: AbstractBuilder{ctx: b.ctx}, element: node.(EndEvent)}
}

func (b *AbstractBuilder) ServiceTask(id string) *ServiceTaskBuilder {
	node := b.ctx.createFlowNodeAndConnect(BPMN_ELEMENT_SERVICE_TASK, id)
	return &ServiceTaskBuilder{AbstractBuilder: AbstractBuilder{ctx: b.ctx}, element: node.(ServiceTask)}
}

func (b *AbstractBuilder) UserTask(id string) *UserTaskBuilder {
	node := b.ctx.createFlowNodeAndConnect(BPMN_ELEMENT_USER_TASK, id)
	return &UserTaskBuilder{AbstractBuilder: AbstractBuilder{ctx: b.ctx}, element: node.(UserTask)}
}

func (b *AbstractBuilder) ScriptTask(id string) *ScriptTaskBuilder {
	node := b.ctx.createFlowNodeAndConnect(BPMN_ELEMENT_SCRIPT_TASK, id)
	return &ScriptTaskBuilder{AbstractBuilder: AbstractBuilder{ctx: b.ctx}, element: node.(ScriptTask)}
}

func (b *AbstractBuilder) BusinessRuleTask(id string) *BusinessRuleTaskBuilder {
	node := b.ctx.createFlowNodeAndConnect(BPMN_ELEMENT_BUSINESS_RULE_TASK, id)
	return &BusinessRuleTaskBuilder{AbstractBuilder: AbstractBuilder{ctx: b.ctx}, element: node.(BusinessRuleTask)}
}

func (b *AbstractBuilder) SendTask(id string) *SendTaskBuilder {
	node := b.ctx.createFlowNodeAndConnect(BPMN_ELEMENT_SEND_TASK, id)
	return &SendTaskBuilder{AbstractBuilder: AbstractBuilder{ctx: b.ctx}, element: node.(SendTask)}
}

func (b *AbstractBuilder) ReceiveTask(id string) *ReceiveTaskBuilder {
	node := b.ctx.createFlowNodeAndConnect(BPMN_ELEMENT_RECEIVE_TASK, id)
	return &ReceiveTaskBuilder{AbstractBuilder: AbstractBuilder{ctx: b.ctx}, element: node.(ReceiveTask)}
}

func (b *AbstractBuilder) ManualTask(id string) *ManualTaskBuilder {
	node := b.ctx.createFlowNodeAndConnect(BPMN_ELEMENT_MANUAL_TASK, id)
	return &ManualTaskBuilder{AbstractBuilder: AbstractBuilder{ctx: b.ctx}, element: node.(ManualTask)}
}

func (b *AbstractBuilder) ExclusiveGateway(id string) *ExclusiveGatewayBuilder {
	node := b.ctx.createFlowNodeAndConnect(BPMN_ELEMENT_EXCLUSIVE_GATEWAY, id)
	gw := node.(ExclusiveGateway)
	b.ctx.gatewayStack = append(b.ctx.gatewayStack, gw)
	return &ExclusiveGatewayBuilder{AbstractBuilder: AbstractBuilder{ctx: b.ctx}, element: gw}
}

func (b *AbstractBuilder) ParallelGateway(id string) *ParallelGatewayBuilder {
	node := b.ctx.createFlowNodeAndConnect(BPMN_ELEMENT_PARALLEL_GATEWAY, id)
	gw := node.(ParallelGateway)
	b.ctx.gatewayStack = append(b.ctx.gatewayStack, gw)
	return &ParallelGatewayBuilder{AbstractBuilder: AbstractBuilder{ctx: b.ctx}, element: gw}
}

func (b *AbstractBuilder) InclusiveGateway(id string) *InclusiveGatewayBuilder {
	node := b.ctx.createFlowNodeAndConnect(BPMN_ELEMENT_INCLUSIVE_GATEWAY, id)
	gw := node.(InclusiveGateway)
	b.ctx.gatewayStack = append(b.ctx.gatewayStack, gw)
	return &InclusiveGatewayBuilder{AbstractBuilder: AbstractBuilder{ctx: b.ctx}, element: gw}
}

func (b *AbstractBuilder) EventBasedGateway(id string) *EventBasedGatewayBuilder {
	node := b.ctx.createFlowNodeAndConnect(BPMN_ELEMENT_EVENT_BASED_GATEWAY, id)
	gw := node.(EventBasedGateway)
	b.ctx.gatewayStack = append(b.ctx.gatewayStack, gw)
	return &EventBasedGatewayBuilder{AbstractBuilder: AbstractBuilder{ctx: b.ctx}, element: gw}
}

func (b *AbstractBuilder) SubProcess(id string) *SubProcessBuilder {
	node := b.ctx.createFlowNodeAndConnect(BPMN_ELEMENT_SUB_PROCESS, id)
	return &SubProcessBuilder{AbstractBuilder: AbstractBuilder{ctx: b.ctx}, element: node.(SubProcess)}
}

func (b *AbstractBuilder) AdHocSubProcess(id string) *AdHocSubProcessBuilder {
	node := b.ctx.createFlowNodeAndConnect(BPMN_ELEMENT_AD_HOC_SUB_PROCESS, id)
	return &AdHocSubProcessBuilder{AbstractBuilder: AbstractBuilder{ctx: b.ctx}, element: node.(AdHocSubProcess)}
}

func (b *AbstractBuilder) CallActivity(id string) *CallActivityBuilder {
	node := b.ctx.createFlowNodeAndConnect(BPMN_ELEMENT_CALL_ACTIVITY, id)
	return &CallActivityBuilder{AbstractBuilder: AbstractBuilder{ctx: b.ctx}, element: node.(CallActivity)}
}

func (b *AbstractBuilder) IntermediateCatchEvent(id string) *IntermediateCatchEventBuilder {
	node := b.ctx.createFlowNodeAndConnect(BPMN_ELEMENT_INTERMEDIATE_CATCH_EVENT, id)
	return &IntermediateCatchEventBuilder{AbstractBuilder: AbstractBuilder{ctx: b.ctx}, element: node.(IntermediateCatchEvent)}
}

func (b *AbstractBuilder) IntermediateThrowEvent(id string) *IntermediateThrowEventBuilder {
	node := b.ctx.createFlowNodeAndConnect(BPMN_ELEMENT_INTERMEDIATE_THROW_EVENT, id)
	return &IntermediateThrowEventBuilder{AbstractBuilder: AbstractBuilder{ctx: b.ctx}, element: node.(IntermediateThrowEvent)}
}

// --- Specific Builders ---

// StartEventBuilder builds a StartEvent.
type StartEventBuilder struct {
	AbstractBuilder
	element StartEvent
}

func (b *StartEventBuilder) Name(name string) *StartEventBuilder {
	b.element.SetName(name)
	return b
}

// EndEventBuilder builds an EndEvent.
type EndEventBuilder struct {
	AbstractBuilder
	element EndEvent
}

func (b *EndEventBuilder) Name(name string) *EndEventBuilder {
	b.element.SetName(name)
	return b
}

// ServiceTaskBuilder builds a ServiceTask.
type ServiceTaskBuilder struct {
	AbstractBuilder
	element ServiceTask
}

func (b *ServiceTaskBuilder) Name(name string) *ServiceTaskBuilder {
	b.element.SetName(name)
	return b
}

func (b *ServiceTaskBuilder) Implementation(impl string) *ServiceTaskBuilder {
	b.element.SetImplementation(impl)
	return b
}

func (b *ServiceTaskBuilder) ZeebeJobType(jobType string) *ServiceTaskBuilder {
	td := getOrCreateExtElement[ZeebeTaskDefinition](b.ctx.bmi.ModelInstance, b.element, ZEEBE_NS, ZEEBE_ELEMENT_TASK_DEFINITION)
	td.SetType(jobType)
	return b
}

func (b *ServiceTaskBuilder) ZeebeJobRetries(retries string) *ServiceTaskBuilder {
	td := getOrCreateExtElement[ZeebeTaskDefinition](b.ctx.bmi.ModelInstance, b.element, ZEEBE_NS, ZEEBE_ELEMENT_TASK_DEFINITION)
	td.SetRetries(retries)
	return b
}

func (b *ServiceTaskBuilder) ZeebeTaskHeader(key, value string) *ServiceTaskBuilder {
	mi := b.ctx.bmi.ModelInstance
	headers := getOrCreateExtElement[ZeebeTaskHeaders](mi, b.element, ZEEBE_NS, ZEEBE_ELEMENT_TASK_HEADERS)
	headerType := bpmnModel.GetTypeByQName(ZEEBE_NS, ZEEBE_ELEMENT_HEADER)
	inst, _ := mi.NewInstance(headerType)
	header := inst.(ZeebeTaskHeader)
	header.SetKey(key)
	header.SetValue(value)
	headers.GetDomElement().AppendChild(header.GetDomElement())
	return b
}

func (b *ServiceTaskBuilder) ZeebeInput(source, target string) *ServiceTaskBuilder {
	mi := b.ctx.bmi.ModelInstance
	mapping := getOrCreateExtElement[ZeebeIoMapping](mi, b.element, ZEEBE_NS, ZEEBE_ELEMENT_IO_MAPPING)
	inputType := bpmnModel.GetTypeByQName(ZEEBE_NS, ZEEBE_ELEMENT_INPUT)
	inst, _ := mi.NewInstance(inputType)
	input := inst.(ZeebeInput)
	input.SetSource(source)
	input.SetTarget(target)
	mapping.GetDomElement().AppendChild(input.GetDomElement())
	return b
}

func (b *ServiceTaskBuilder) ZeebeOutput(source, target string) *ServiceTaskBuilder {
	mi := b.ctx.bmi.ModelInstance
	mapping := getOrCreateExtElement[ZeebeIoMapping](mi, b.element, ZEEBE_NS, ZEEBE_ELEMENT_IO_MAPPING)
	outputType := bpmnModel.GetTypeByQName(ZEEBE_NS, ZEEBE_ELEMENT_OUTPUT)
	inst, _ := mi.NewInstance(outputType)
	output := inst.(ZeebeOutput)
	output.SetSource(source)
	output.SetTarget(target)
	mapping.GetDomElement().AppendChild(output.GetDomElement())
	return b
}

// UserTaskBuilder builds a UserTask.
type UserTaskBuilder struct {
	AbstractBuilder
	element UserTask
}

func (b *UserTaskBuilder) Name(name string) *UserTaskBuilder {
	b.element.SetName(name)
	return b
}

func (b *UserTaskBuilder) Implementation(impl string) *UserTaskBuilder {
	b.element.SetImplementation(impl)
	return b
}

func (b *UserTaskBuilder) ZeebeAssignee(assignee string) *UserTaskBuilder {
	ad := getOrCreateExtElement[ZeebeAssignmentDefinition](b.ctx.bmi.ModelInstance, b.element, ZEEBE_NS, ZEEBE_ELEMENT_ASSIGNMENT_DEFINITION)
	ad.SetAssignee(assignee)
	return b
}

func (b *UserTaskBuilder) ZeebeCandidateGroups(groups string) *UserTaskBuilder {
	ad := getOrCreateExtElement[ZeebeAssignmentDefinition](b.ctx.bmi.ModelInstance, b.element, ZEEBE_NS, ZEEBE_ELEMENT_ASSIGNMENT_DEFINITION)
	ad.SetCandidateGroups(groups)
	return b
}

func (b *UserTaskBuilder) ZeebeCandidateUsers(users string) *UserTaskBuilder {
	ad := getOrCreateExtElement[ZeebeAssignmentDefinition](b.ctx.bmi.ModelInstance, b.element, ZEEBE_NS, ZEEBE_ELEMENT_ASSIGNMENT_DEFINITION)
	ad.SetCandidateUsers(users)
	return b
}

func (b *UserTaskBuilder) ZeebeFormId(formId string) *UserTaskBuilder {
	fd := getOrCreateExtElement[ZeebeFormDefinition](b.ctx.bmi.ModelInstance, b.element, ZEEBE_NS, ZEEBE_ELEMENT_FORM_DEFINITION)
	fd.SetFormId(formId)
	return b
}

func (b *UserTaskBuilder) ZeebeFormKey(formKey string) *UserTaskBuilder {
	fd := getOrCreateExtElement[ZeebeFormDefinition](b.ctx.bmi.ModelInstance, b.element, ZEEBE_NS, ZEEBE_ELEMENT_FORM_DEFINITION)
	fd.SetFormKey(formKey)
	return b
}

// ScriptTaskBuilder builds a ScriptTask.
type ScriptTaskBuilder struct {
	AbstractBuilder
	element ScriptTask
}

func (b *ScriptTaskBuilder) Name(name string) *ScriptTaskBuilder {
	b.element.SetName(name)
	return b
}

func (b *ScriptTaskBuilder) ScriptFormat(format string) *ScriptTaskBuilder {
	b.element.SetScriptFormat(format)
	return b
}

func (b *ScriptTaskBuilder) ZeebeScript(expression, resultVariable string) *ScriptTaskBuilder {
	s := getOrCreateExtElement[ZeebeScript](b.ctx.bmi.ModelInstance, b.element, ZEEBE_NS, ZEEBE_ELEMENT_SCRIPT)
	s.SetExpression(expression)
	s.SetResultVariable(resultVariable)
	return b
}

func (b *ScriptTaskBuilder) ZeebeJobType(jobType string) *ScriptTaskBuilder {
	td := getOrCreateExtElement[ZeebeTaskDefinition](b.ctx.bmi.ModelInstance, b.element, ZEEBE_NS, ZEEBE_ELEMENT_TASK_DEFINITION)
	td.SetType(jobType)
	return b
}

// BusinessRuleTaskBuilder builds a BusinessRuleTask.
type BusinessRuleTaskBuilder struct {
	AbstractBuilder
	element BusinessRuleTask
}

func (b *BusinessRuleTaskBuilder) Name(name string) *BusinessRuleTaskBuilder {
	b.element.SetName(name)
	return b
}

func (b *BusinessRuleTaskBuilder) Implementation(impl string) *BusinessRuleTaskBuilder {
	b.element.SetImplementation(impl)
	return b
}

func (b *BusinessRuleTaskBuilder) ZeebeCalledDecision(decisionId, resultVariable string) *BusinessRuleTaskBuilder {
	cd := getOrCreateExtElement[ZeebeCalledDecision](b.ctx.bmi.ModelInstance, b.element, ZEEBE_NS, ZEEBE_ELEMENT_CALLED_DECISION)
	cd.SetDecisionId(decisionId)
	cd.SetResultVariable(resultVariable)
	return b
}

// SendTaskBuilder builds a SendTask.
type SendTaskBuilder struct {
	AbstractBuilder
	element SendTask
}

func (b *SendTaskBuilder) Name(name string) *SendTaskBuilder {
	b.element.SetName(name)
	return b
}

func (b *SendTaskBuilder) Implementation(impl string) *SendTaskBuilder {
	b.element.SetImplementation(impl)
	return b
}

// ReceiveTaskBuilder builds a ReceiveTask.
type ReceiveTaskBuilder struct {
	AbstractBuilder
	element ReceiveTask
}

func (b *ReceiveTaskBuilder) Name(name string) *ReceiveTaskBuilder {
	b.element.SetName(name)
	return b
}

func (b *ReceiveTaskBuilder) Implementation(impl string) *ReceiveTaskBuilder {
	b.element.SetImplementation(impl)
	return b
}

func (b *ReceiveTaskBuilder) ZeebeCorrelationKey(key string) *ReceiveTaskBuilder {
	sub := getOrCreateExtElement[ZeebeSubscription](b.ctx.bmi.ModelInstance, b.element, ZEEBE_NS, ZEEBE_ELEMENT_SUBSCRIPTION)
	sub.SetCorrelationKey(key)
	return b
}

// ManualTaskBuilder builds a ManualTask.
type ManualTaskBuilder struct {
	AbstractBuilder
	element ManualTask
}

func (b *ManualTaskBuilder) Name(name string) *ManualTaskBuilder {
	b.element.SetName(name)
	return b
}

// ExclusiveGatewayBuilder builds an ExclusiveGateway.
type ExclusiveGatewayBuilder struct {
	AbstractBuilder
	element ExclusiveGateway
}

func (b *ExclusiveGatewayBuilder) Name(name string) *ExclusiveGatewayBuilder {
	b.element.SetName(name)
	return b
}

// ParallelGatewayBuilder builds a ParallelGateway.
type ParallelGatewayBuilder struct {
	AbstractBuilder
	element ParallelGateway
}

func (b *ParallelGatewayBuilder) Name(name string) *ParallelGatewayBuilder {
	b.element.SetName(name)
	return b
}

// InclusiveGatewayBuilder builds an InclusiveGateway.
type InclusiveGatewayBuilder struct {
	AbstractBuilder
	element InclusiveGateway
}

func (b *InclusiveGatewayBuilder) Name(name string) *InclusiveGatewayBuilder {
	b.element.SetName(name)
	return b
}

// EventBasedGatewayBuilder builds an EventBasedGateway.
type EventBasedGatewayBuilder struct {
	AbstractBuilder
	element EventBasedGateway
}

func (b *EventBasedGatewayBuilder) Name(name string) *EventBasedGatewayBuilder {
	b.element.SetName(name)
	return b
}

// SubProcessBuilder builds a SubProcess.
type SubProcessBuilder struct {
	AbstractBuilder
	element SubProcess
}

func (b *SubProcessBuilder) Name(name string) *SubProcessBuilder {
	b.element.SetName(name)
	return b
}

// AdHocSubProcessBuilder builds an AdHocSubProcess.
type AdHocSubProcessBuilder struct {
	AbstractBuilder
	element AdHocSubProcess
}

func (b *AdHocSubProcessBuilder) Name(name string) *AdHocSubProcessBuilder {
	b.element.SetName(name)
	return b
}

func (b *AdHocSubProcessBuilder) ZeebeAdHocConfig(outputCollection, outputElement string) *AdHocSubProcessBuilder {
	ah := getOrCreateExtElement[ZeebeAdHoc](b.ctx.bmi.ModelInstance, b.element, ZEEBE_NS, ZEEBE_ELEMENT_AD_HOC)
	ah.SetOutputCollection(outputCollection)
	ah.SetOutputElement(outputElement)
	return b
}

func (b *AdHocSubProcessBuilder) ZeebeJobType(jobType string) *AdHocSubProcessBuilder {
	td := getOrCreateExtElement[ZeebeTaskDefinition](b.ctx.bmi.ModelInstance, b.element, ZEEBE_NS, ZEEBE_ELEMENT_TASK_DEFINITION)
	td.SetType(jobType)
	return b
}

func (b *AdHocSubProcessBuilder) ZeebeInput(source, target string) *AdHocSubProcessBuilder {
	ioMapping := getOrCreateExtElement[ZeebeIoMapping](b.ctx.bmi.ModelInstance, b.element, ZEEBE_NS, ZEEBE_ELEMENT_IO_MAPPING)
	inputType := bpmnModel.GetTypeByQName(ZEEBE_NS, ZEEBE_ELEMENT_INPUT)
	inst, _ := b.ctx.bmi.ModelInstance.NewInstance(inputType)
	input := inst.(ZeebeInput)
	input.SetSource(source)
	input.SetTarget(target)
	ioMapping.GetDomElement().AppendChild(input.GetDomElement())
	return b
}

func (b *AdHocSubProcessBuilder) ZeebeOutput(source, target string) *AdHocSubProcessBuilder {
	ioMapping := getOrCreateExtElement[ZeebeIoMapping](b.ctx.bmi.ModelInstance, b.element, ZEEBE_NS, ZEEBE_ELEMENT_IO_MAPPING)
	outputType := bpmnModel.GetTypeByQName(ZEEBE_NS, ZEEBE_ELEMENT_OUTPUT)
	inst, _ := b.ctx.bmi.ModelInstance.NewInstance(outputType)
	output := inst.(ZeebeOutput)
	output.SetSource(source)
	output.SetTarget(target)
	ioMapping.GetDomElement().AppendChild(output.GetDomElement())
	return b
}

// CallActivityBuilder builds a CallActivity.
type CallActivityBuilder struct {
	AbstractBuilder
	element CallActivity
}

func (b *CallActivityBuilder) Name(name string) *CallActivityBuilder {
	b.element.SetName(name)
	return b
}

func (b *CallActivityBuilder) CalledElement(elem string) *CallActivityBuilder {
	b.element.SetCalledElement(elem)
	return b
}

func (b *CallActivityBuilder) ZeebeProcessId(processId string) *CallActivityBuilder {
	ce := getOrCreateExtElement[ZeebeCalledElement](b.ctx.bmi.ModelInstance, b.element, ZEEBE_NS, ZEEBE_ELEMENT_CALLED_ELEMENT)
	ce.SetProcessId(processId)
	return b
}

func (b *CallActivityBuilder) ZeebePropagateAllChildVariables(v bool) *CallActivityBuilder {
	ce := getOrCreateExtElement[ZeebeCalledElement](b.ctx.bmi.ModelInstance, b.element, ZEEBE_NS, ZEEBE_ELEMENT_CALLED_ELEMENT)
	ce.SetPropagateAllChildVariables(v)
	return b
}

// IntermediateCatchEventBuilder builds an IntermediateCatchEvent.
type IntermediateCatchEventBuilder struct {
	AbstractBuilder
	element IntermediateCatchEvent
}

func (b *IntermediateCatchEventBuilder) Name(name string) *IntermediateCatchEventBuilder {
	b.element.SetName(name)
	return b
}

// IntermediateThrowEventBuilder builds an IntermediateThrowEvent.
type IntermediateThrowEventBuilder struct {
	AbstractBuilder
	element IntermediateThrowEvent
}

func (b *IntermediateThrowEventBuilder) Name(name string) *IntermediateThrowEventBuilder {
	b.element.SetName(name)
	return b
}
