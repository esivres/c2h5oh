package bpmn_model

import (
	"fmt"
	"os"

	xmlm "github.com/esivres/c2h5oh/pkg/bpmn_model/xml"
)

// bpmnModel is the singleton BPMN model with all registered types.
var bpmnModel *Model

func init() {
	bpmnModel = buildBpmnModel()
}

// GetBpmnModel returns the singleton BPMN model.
func GetBpmnModel() *Model {
	return bpmnModel
}

// buildBpmnModel creates and populates the BPMN model with all registered types.
func buildBpmnModel() *Model {
	mb := NewModelBuilder("BPMN")

	registerDefinitionsType(mb)
	registerBaseElementType(mb)
	registerRootElementType(mb)
	registerCallableElementType(mb)
	registerFlowElementType(mb)
	registerFlowNodeType(mb)
	registerEventType(mb)
	registerCatchEventType(mb)
	registerThrowEventType(mb)
	registerStartEventType(mb)
	registerEndEventType(mb)
	registerSequenceFlowType(mb)
	registerConditionExpressionType(mb)
	registerProcessType(mb)
	registerExtensionElementsType(mb)

	// Activities & Tasks (Stage 3)
	registerActivityType(mb)
	registerTaskType(mb)
	registerServiceTaskType(mb)
	registerUserTaskType(mb)
	registerScriptTaskType(mb)
	registerBusinessRuleTaskType(mb)
	registerSendTaskType(mb)
	registerReceiveTaskType(mb)
	registerManualTaskType(mb)
	registerSubProcessType(mb)
	registerAdHocSubProcessType(mb)
	registerCallActivityType(mb)

	// Gateways (Stage 3)
	registerGatewayType(mb)
	registerExclusiveGatewayType(mb)
	registerParallelGatewayType(mb)
	registerInclusiveGatewayType(mb)
	registerEventBasedGatewayType(mb)
	registerComplexGatewayType(mb)

	// Events (Stage 3)
	registerIntermediateCatchEventType(mb)
	registerIntermediateThrowEventType(mb)
	registerBoundaryEventType(mb)
	registerEventDefinitionType(mb)
	registerMessageEventDefinitionType(mb)
	registerTimerEventDefinitionType(mb)
	registerSignalEventDefinitionType(mb)
	registerErrorEventDefinitionType(mb)
	registerEscalationEventDefinitionType(mb)
	registerConditionalEventDefinitionType(mb)
	registerLinkEventDefinitionType(mb)
	registerCompensateEventDefinitionType(mb)
	registerCancelEventDefinitionType(mb)
	registerTerminateEventDefinitionType(mb)

	// Global elements (Stage 3)
	registerMessageType(mb)
	registerSignalType(mb)
	registerErrorType(mb)
	registerEscalationType(mb)
	registerItemDefinitionType(mb)

	// Collaboration (Stage 3)
	registerCollaborationType(mb)
	registerParticipantType(mb)
	registerMessageFlowType(mb)
	registerLaneSetType(mb)
	registerLaneType(mb)

	// Data (Stage 3)
	registerDataObjectType(mb)
	registerDataObjectReferenceType(mb)

	// Artifacts (Stage 3)
	registerArtifactType(mb)
	registerTextAnnotationType(mb)
	registerGroupType(mb)
	registerAssociationType(mb)

	// Multi-instance & Documentation (Stage 3)
	registerMultiInstanceLoopCharacteristicsType(mb)
	registerDocumentationType(mb)

	// BPMN DI (Stage 5)
	registerDiTypes(mb)

	// Zeebe extensions (Stage 6)
	registerZeebeTypes(mb)

	return mb.Build()
}

// --- Type registration functions ---

func registerDefinitionsType(mb *ModelBuilder) {
	tb := mb.DefineType((*Definitions)(nil), BPMN_ELEMENT_DEFINITIONS).
		Namespace(BPMN20_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &DefinitionsImpl{NewBaseInstance(ctx)}
		})

	AttrBaseElementId = tb.StringAttribute(BPMN_ATTRIBUTE_ID).IdAttribute().Build()
	AttrName = tb.StringAttribute(BPMN_ATTRIBUTE_NAME).Build()
	AttrDefinitionsTargetNamespace = tb.StringAttribute(BPMN_ATTRIBUTE_TARGET_NAMESPACE).Build()
	AttrDefinitionsExpressionLanguage = tb.StringAttribute(BPMN_ATTRIBUTE_EXPRESSION_LANGUAGE).Build()
	AttrDefinitionsTypeLanguage = tb.StringAttribute(BPMN_ATTRIBUTE_TYPE_LANGUAGE).Build()
	AttrDefinitionsExporter = tb.StringAttribute(BPMN_ATTRIBUTE_EXPORTER).Build()
	AttrDefinitionsExporterVersion = tb.StringAttribute(BPMN_ATTRIBUTE_EXPORTER_VERSION).Build()
}

func registerBaseElementType(mb *ModelBuilder) {
	mb.DefineType((*BaseElement)(nil), BPMN_ELEMENT_BASE_ELEMENT).
		Namespace(BPMN20_NS).
		AbstractType()
	// id attribute is already registered on Definitions; it's shared via the same descriptor
}

func registerRootElementType(mb *ModelBuilder) {
	mb.DefineType((*RootElement)(nil), BPMN_ELEMENT_ROOT_ELEMENT).
		Namespace(BPMN20_NS).
		ExtendsType((*BaseElement)(nil)).
		AbstractType()
}

func registerCallableElementType(mb *ModelBuilder) {
	mb.DefineType((*CallableElement)(nil), BPMN_ELEMENT_CALLABLE_ELEMENT).
		Namespace(BPMN20_NS).
		ExtendsType((*RootElement)(nil)).
		AbstractType()
}

func registerFlowElementType(mb *ModelBuilder) {
	mb.DefineType((*FlowElement)(nil), BPMN_ELEMENT_FLOW_ELEMENT).
		Namespace(BPMN20_NS).
		ExtendsType((*BaseElement)(nil)).
		AbstractType()
}

func registerFlowNodeType(mb *ModelBuilder) {
	mb.DefineType((*FlowNode)(nil), BPMN_ELEMENT_FLOW_NODE).
		Namespace(BPMN20_NS).
		ExtendsType((*FlowElement)(nil)).
		AbstractType()
}

func registerEventType(mb *ModelBuilder) {
	mb.DefineType((*Event)(nil), BPMN_ELEMENT_EVENT).
		Namespace(BPMN20_NS).
		ExtendsType((*FlowNode)(nil)).
		AbstractType()
}

func registerCatchEventType(mb *ModelBuilder) {
	mb.DefineType((*CatchEvent)(nil), BPMN_ELEMENT_CATCH_EVENT).
		Namespace(BPMN20_NS).
		ExtendsType((*Event)(nil)).
		AbstractType()
}

func registerThrowEventType(mb *ModelBuilder) {
	mb.DefineType((*ThrowEvent)(nil), BPMN_ELEMENT_THROW_EVENT).
		Namespace(BPMN20_NS).
		ExtendsType((*Event)(nil)).
		AbstractType()
}

func registerStartEventType(mb *ModelBuilder) {
	tb := mb.DefineType((*StartEvent)(nil), BPMN_ELEMENT_START_EVENT).
		Namespace(BPMN20_NS).
		ExtendsType((*CatchEvent)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &StartEventImpl{CatchEventImpl{EventImpl{FlowNodeImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}}}}
		})

	AttrStartEventIsInterrupting = tb.BoolAttribute(BPMN_ATTRIBUTE_IS_INTERRUPTING).DefaultValue(true).Build()
}

func registerEndEventType(mb *ModelBuilder) {
	mb.DefineType((*EndEvent)(nil), BPMN_ELEMENT_END_EVENT).
		Namespace(BPMN20_NS).
		ExtendsType((*ThrowEvent)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &EndEventImpl{ThrowEventImpl{EventImpl{FlowNodeImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}}}}
		})
}

func registerSequenceFlowType(mb *ModelBuilder) {
	tb := mb.DefineType((*SequenceFlow)(nil), BPMN_ELEMENT_SEQUENCE_FLOW).
		Namespace(BPMN20_NS).
		ExtendsType((*FlowElement)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &SequenceFlowImpl{FlowElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}
		})

	AttrSequenceFlowSourceRef = tb.StringAttribute(BPMN_ATTRIBUTE_SOURCE_REF).Required().Build()
	AttrSequenceFlowTargetRef = tb.StringAttribute(BPMN_ATTRIBUTE_TARGET_REF).Required().Build()
	AttrSequenceFlowIsImmediate = tb.BoolAttribute(BPMN_ATTRIBUTE_IS_IMMEDIATE).Build()
}

func registerConditionExpressionType(mb *ModelBuilder) {
	tb := mb.DefineType((*ConditionExpression)(nil), BPMN_ELEMENT_CONDITION_EXPRESSION).
		Namespace(BPMN20_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ConditionExpressionImpl{NewBaseInstance(ctx)}
		})

	AttrConditionExpressionType = tb.StringAttribute("type").
		AttributeNamespace(XSI_NS).
		Build()
}

func registerProcessType(mb *ModelBuilder) {
	tb := mb.DefineType((*Process)(nil), BPMN_ELEMENT_PROCESS).
		Namespace(BPMN20_NS).
		ExtendsType((*CallableElement)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ProcessImpl{CallableElementImpl{BaseElementImpl{NewBaseInstance(ctx)}}}
		})

	AttrProcessType = tb.StringAttribute(BPMN_ATTRIBUTE_PROCESS_TYPE).DefaultValue("None").Build()
	AttrProcessIsClosed = tb.BoolAttribute(BPMN_ATTRIBUTE_IS_CLOSED).DefaultValue(false).Build()
	AttrProcessIsExecutable = tb.BoolAttribute(BPMN_ATTRIBUTE_IS_EXECUTABLE).Build()
}

func registerExtensionElementsType(mb *ModelBuilder) {
	mb.DefineType((*ExtensionElements)(nil), BPMN_ELEMENT_EXTENSION_ELEMENTS).
		Namespace(BPMN20_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &ExtensionElementsImpl{NewBaseInstance(ctx)}
		})
}

// --- Public API ---

// BpmnModelInstance represents a parsed/created BPMN model document.
type BpmnModelInstance struct {
	*ModelInstance
}

// GetDefinitions returns the root Definitions element.
func (bmi *BpmnModelInstance) GetDefinitions() Definitions {
	root := bmi.GetDocumentElement()
	if root == nil {
		return nil
	}
	if def, ok := root.(Definitions); ok {
		return def
	}
	return nil
}

// ReadFromFile reads a BPMN model from a file.
func ReadFromFile(path string) (*BpmnModelInstance, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read bpmn file: %w", err)
	}
	return ReadFromBytes(data)
}

// ReadFromString reads a BPMN model from a string.
func ReadFromString(s string) (*BpmnModelInstance, error) {
	return ReadFromBytes([]byte(s))
}

// ReadFromBytes reads a BPMN model from bytes.
func ReadFromBytes(data []byte) (*BpmnModelInstance, error) {
	doc, err := xmlm.ParseDocument(data)
	if err != nil {
		return nil, fmt.Errorf("parse bpmn xml: %w", err)
	}

	mi := NewModelInstance(bpmnModel, doc)
	if err := mi.ResolveElements(); err != nil {
		return nil, fmt.Errorf("resolve bpmn elements: %w", err)
	}

	return &BpmnModelInstance{mi}, nil
}

// WriteToFile writes a BPMN model to a file.
func WriteToFile(bmi *BpmnModelInstance, path string) error {
	data, err := bmi.GetDocument().WriteToBytes()
	if err != nil {
		return fmt.Errorf("serialize bpmn: %w", err)
	}
	const filePermissions = 0o644
	return os.WriteFile(path, data, filePermissions)
}

// ConvertToString serializes a BPMN model to an XML string.
func ConvertToString(bmi *BpmnModelInstance) (string, error) {
	return bmi.GetDocument().WriteToString()
}

// newBpmnDocument creates a new document with standard BPMN namespace declarations.
func newBpmnDocument() *xmlm.Document {
	doc := xmlm.NewDocument()
	doc.RegisterNamespace("bpmn", BPMN20_NS)
	doc.RegisterNamespace("bpmndi", BPMNDI_NS)
	doc.RegisterNamespace("dc", DC_NS)
	doc.RegisterNamespace("di", DI_NS)
	doc.RegisterNamespace("zeebe", ZEEBE_NS)
	return doc
}

// CreateProcess creates a new ProcessBuilder with an empty process.
func CreateProcess() *ProcessBuilder {
	return newProcessBuilder("", false)
}

// CreateExecutableProcess creates a new ProcessBuilder with an executable process.
func CreateExecutableProcess(processId string) *ProcessBuilder {
	return newProcessBuilder(processId, true)
}

func newProcessBuilder(processId string, executable bool) *ProcessBuilder {
	bmi := createProcess(processId, executable)
	processes := GetTypedElements[Process](bmi.ModelInstance)
	var proc Process
	if len(processes) > 0 {
		proc = processes[0]
	}
	return &ProcessBuilder{
		ctx: &builderCtx{
			bmi:     bmi,
			process: proc,
		},
	}
}

func createProcess(processId string, executable bool) *BpmnModelInstance {
	doc := newBpmnDocument()
	mi := NewModelInstance(bpmnModel, doc)
	bmi := &BpmnModelInstance{mi}

	// Create definitions
	defType := bpmnModel.GetTypeByQName(BPMN20_NS, BPMN_ELEMENT_DEFINITIONS)
	defInst, _ := mi.NewInstance(defType)
	def := defInst.(Definitions)
	def.SetTargetNamespace("http://bpmn.io/schema/bpmn")

	// Add namespace declarations to root element
	domRoot := def.GetDomElement()
	domRoot.AddNamespaceDeclaration("bpmn", BPMN20_NS)
	domRoot.AddNamespaceDeclaration("bpmndi", BPMNDI_NS)
	domRoot.AddNamespaceDeclaration("dc", DC_NS)
	domRoot.AddNamespaceDeclaration("di", DI_NS)
	domRoot.AddNamespaceDeclaration("zeebe", ZEEBE_NS)

	mi.SetDocumentElement(def)

	// Create process
	procType := bpmnModel.GetTypeByQName(BPMN20_NS, BPMN_ELEMENT_PROCESS)
	procInst, _ := mi.NewInstance(procType)
	proc := procInst.(Process)
	if processId != "" {
		proc.SetId(processId)
	}
	proc.SetIsExecutable(executable)

	domRoot.AppendChild(proc.GetDomElement())
	mi.RegisterElement(proc)

	return bmi
}
