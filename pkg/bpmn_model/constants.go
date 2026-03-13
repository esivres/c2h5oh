package bpmn_model

// XML Namespace URIs for BPMN 2.0 and related specifications.
const (
	BPMN20_NS  = "http://www.omg.org/spec/BPMN/20100524/MODEL"
	BPMNDI_NS  = "http://www.omg.org/spec/BPMN/20100524/DI"
	DC_NS      = "http://www.omg.org/spec/DD/20100524/DC"
	DI_NS      = "http://www.omg.org/spec/DD/20100524/DI"
	ZEEBE_NS   = "http://camunda.org/schema/zeebe/1.0"
	MODELER_NS = "http://camunda.org/schema/modeler/1.0"
	XSI_NS     = "http://www.w3.org/2001/XMLSchema-instance"
)

// BPMN element names.
const (
	BPMN_ELEMENT_DEFINITIONS          = "definitions"
	BPMN_ELEMENT_BASE_ELEMENT         = "baseElement"
	BPMN_ELEMENT_ROOT_ELEMENT         = "rootElement"
	BPMN_ELEMENT_CALLABLE_ELEMENT     = "callableElement"
	BPMN_ELEMENT_FLOW_ELEMENT         = "flowElement"
	BPMN_ELEMENT_FLOW_NODE            = "flowNode"
	BPMN_ELEMENT_EVENT                = "event"
	BPMN_ELEMENT_CATCH_EVENT          = "catchEvent"
	BPMN_ELEMENT_THROW_EVENT          = "throwEvent"
	BPMN_ELEMENT_START_EVENT          = "startEvent"
	BPMN_ELEMENT_END_EVENT            = "endEvent"
	BPMN_ELEMENT_SEQUENCE_FLOW        = "sequenceFlow"
	BPMN_ELEMENT_CONDITION_EXPRESSION = "conditionExpression"
	BPMN_ELEMENT_PROCESS              = "process"
	BPMN_ELEMENT_INCOMING             = "incoming"
	BPMN_ELEMENT_OUTGOING             = "outgoing"
	BPMN_ELEMENT_EXTENSION_ELEMENTS   = "extensionElements"

	// Activities
	BPMN_ELEMENT_ACTIVITY           = "activity"
	BPMN_ELEMENT_TASK               = "task"
	BPMN_ELEMENT_SERVICE_TASK       = "serviceTask"
	BPMN_ELEMENT_USER_TASK          = "userTask"
	BPMN_ELEMENT_SCRIPT_TASK        = "scriptTask"
	BPMN_ELEMENT_BUSINESS_RULE_TASK = "businessRuleTask"
	BPMN_ELEMENT_SEND_TASK          = "sendTask"
	BPMN_ELEMENT_RECEIVE_TASK       = "receiveTask"
	BPMN_ELEMENT_MANUAL_TASK        = "manualTask"
	BPMN_ELEMENT_SUB_PROCESS        = "subProcess"
	BPMN_ELEMENT_AD_HOC_SUB_PROCESS = "adHocSubProcess"
	BPMN_ELEMENT_CALL_ACTIVITY      = "callActivity"

	// Gateways
	BPMN_ELEMENT_GATEWAY             = "gateway"
	BPMN_ELEMENT_EXCLUSIVE_GATEWAY   = "exclusiveGateway"
	BPMN_ELEMENT_PARALLEL_GATEWAY    = "parallelGateway"
	BPMN_ELEMENT_INCLUSIVE_GATEWAY   = "inclusiveGateway"
	BPMN_ELEMENT_EVENT_BASED_GATEWAY = "eventBasedGateway"
	BPMN_ELEMENT_COMPLEX_GATEWAY     = "complexGateway"

	// Events
	BPMN_ELEMENT_INTERMEDIATE_CATCH_EVENT = "intermediateCatchEvent"
	BPMN_ELEMENT_INTERMEDIATE_THROW_EVENT = "intermediateThrowEvent"
	BPMN_ELEMENT_BOUNDARY_EVENT           = "boundaryEvent"

	// Event Definitions
	BPMN_ELEMENT_EVENT_DEFINITION             = "eventDefinition"
	BPMN_ELEMENT_MESSAGE_EVENT_DEFINITION     = "messageEventDefinition"
	BPMN_ELEMENT_TIMER_EVENT_DEFINITION       = "timerEventDefinition"
	BPMN_ELEMENT_SIGNAL_EVENT_DEFINITION      = "signalEventDefinition"
	BPMN_ELEMENT_ERROR_EVENT_DEFINITION       = "errorEventDefinition"
	BPMN_ELEMENT_ESCALATION_EVENT_DEFINITION  = "escalationEventDefinition"
	BPMN_ELEMENT_CONDITIONAL_EVENT_DEFINITION = "conditionalEventDefinition"
	BPMN_ELEMENT_LINK_EVENT_DEFINITION        = "linkEventDefinition"
	BPMN_ELEMENT_COMPENSATE_EVENT_DEFINITION  = "compensateEventDefinition"
	BPMN_ELEMENT_CANCEL_EVENT_DEFINITION      = "cancelEventDefinition"
	BPMN_ELEMENT_TERMINATE_EVENT_DEFINITION   = "terminateEventDefinition"

	// Global elements
	BPMN_ELEMENT_MESSAGE         = "message"
	BPMN_ELEMENT_SIGNAL          = "signal"
	BPMN_ELEMENT_ERROR           = "error"
	BPMN_ELEMENT_ESCALATION      = "escalation"
	BPMN_ELEMENT_ITEM_DEFINITION = "itemDefinition"

	// Collaboration
	BPMN_ELEMENT_COLLABORATION = "collaboration"
	BPMN_ELEMENT_PARTICIPANT   = "participant"
	BPMN_ELEMENT_MESSAGE_FLOW  = "messageFlow"
	BPMN_ELEMENT_LANE_SET      = "laneSet"
	BPMN_ELEMENT_LANE          = "lane"

	// Data
	BPMN_ELEMENT_DATA_OBJECT           = "dataObject"
	BPMN_ELEMENT_DATA_OBJECT_REFERENCE = "dataObjectReference"

	// Artifacts
	BPMN_ELEMENT_ARTIFACT        = "artifact"
	BPMN_ELEMENT_TEXT_ANNOTATION = "textAnnotation"
	BPMN_ELEMENT_GROUP           = "group"
	BPMN_ELEMENT_ASSOCIATION     = "association"

	// Multi-instance
	BPMN_ELEMENT_MULTI_INSTANCE_LOOP_CHARACTERISTICS = "multiInstanceLoopCharacteristics"

	// Completion condition
	BPMN_ELEMENT_COMPLETION_CONDITION = "completionCondition"

	// Documentation
	BPMN_ELEMENT_DOCUMENTATION = "documentation"
)

// BPMN attribute names.
const (
	BPMN_ATTRIBUTE_ID                  = "id"
	BPMN_ATTRIBUTE_NAME                = "name"
	BPMN_ATTRIBUTE_TARGET_NAMESPACE    = "targetNamespace"
	BPMN_ATTRIBUTE_EXPRESSION_LANGUAGE = "expressionLanguage"
	BPMN_ATTRIBUTE_TYPE_LANGUAGE       = "typeLanguage"
	BPMN_ATTRIBUTE_EXPORTER            = "exporter"
	BPMN_ATTRIBUTE_EXPORTER_VERSION    = "exporterVersion"
	BPMN_ATTRIBUTE_IS_EXECUTABLE       = "isExecutable"
	BPMN_ATTRIBUTE_IS_CLOSED           = "isClosed"
	BPMN_ATTRIBUTE_PROCESS_TYPE        = "processType"
	BPMN_ATTRIBUTE_SOURCE_REF          = "sourceRef"
	BPMN_ATTRIBUTE_TARGET_REF          = "targetRef"
	BPMN_ATTRIBUTE_IS_IMMEDIATE        = "isImmediate"
	BPMN_ATTRIBUTE_IS_INTERRUPTING     = "isInterrupting"

	// Activity
	BPMN_ATTRIBUTE_IS_FOR_COMPENSATION = "isForCompensation"
	BPMN_ATTRIBUTE_DEFAULT             = "default"
	BPMN_ATTRIBUTE_IMPLEMENTATION      = "implementation"
	BPMN_ATTRIBUTE_SCRIPT_FORMAT       = "scriptFormat"
	BPMN_ATTRIBUTE_CALLED_ELEMENT      = "calledElement"
	BPMN_ATTRIBUTE_TRIGGERED_BY_EVENT  = "triggeredByEvent"

	// Gateway
	BPMN_ATTRIBUTE_GATEWAY_DIRECTION  = "gatewayDirection"
	BPMN_ATTRIBUTE_INSTANTIATE        = "instantiate"
	BPMN_ATTRIBUTE_EVENT_GATEWAY_TYPE = "eventGatewayType"

	// BoundaryEvent
	BPMN_ATTRIBUTE_CANCEL_ACTIVITY = "cancelActivity"
	BPMN_ATTRIBUTE_ATTACHED_TO_REF = "attachedToRef"

	// Event definitions
	BPMN_ATTRIBUTE_MESSAGE_REF    = "messageRef"
	BPMN_ATTRIBUTE_SIGNAL_REF     = "signalRef"
	BPMN_ATTRIBUTE_ERROR_REF      = "errorRef"
	BPMN_ATTRIBUTE_ESCALATION_REF = "escalationRef"

	// Global elements
	BPMN_ATTRIBUTE_ERROR_CODE      = "errorCode"
	BPMN_ATTRIBUTE_ESCALATION_CODE = "escalationCode"
	BPMN_ATTRIBUTE_STRUCTURE_REF   = "structureRef"

	// Collaboration
	BPMN_ATTRIBUTE_PROCESS_REF = "processRef"

	// Artifacts
	BPMN_ATTRIBUTE_TEXT_FORMAT           = "textFormat"
	BPMN_ATTRIBUTE_ASSOCIATION_DIRECTION = "associationDirection"
	BPMN_ATTRIBUTE_CATEGORY_VALUE_REF    = "categoryValueRef"

	// Multi-instance
	BPMN_ATTRIBUTE_IS_SEQUENTIAL = "isSequential"

	// Link
	BPMN_ATTRIBUTE_LINK_NAME = "name"
)

// Zeebe extension element names.
const (
	ZEEBE_ELEMENT_TASK_DEFINITION       = "taskDefinition"
	ZEEBE_ELEMENT_IO_MAPPING            = "ioMapping"
	ZEEBE_ELEMENT_INPUT                 = "input"
	ZEEBE_ELEMENT_OUTPUT                = "output"
	ZEEBE_ELEMENT_TASK_HEADERS          = "taskHeaders"
	ZEEBE_ELEMENT_HEADER                = "header"
	ZEEBE_ELEMENT_SUBSCRIPTION          = "subscription"
	ZEEBE_ELEMENT_CALLED_ELEMENT        = "calledElement"
	ZEEBE_ELEMENT_CALLED_DECISION       = "calledDecision"
	ZEEBE_ELEMENT_FORM_DEFINITION       = "formDefinition"
	ZEEBE_ELEMENT_USER_TASK             = "userTask"
	ZEEBE_ELEMENT_ASSIGNMENT_DEFINITION = "assignmentDefinition"
	ZEEBE_ELEMENT_EXECUTION_LISTENERS   = "executionListeners"
	ZEEBE_ELEMENT_EXECUTION_LISTENER    = "executionListener"
	ZEEBE_ELEMENT_PROPERTIES            = "properties"
	ZEEBE_ELEMENT_PROPERTY              = "property"
	ZEEBE_ELEMENT_LOOP_CHARACTERISTICS  = "loopCharacteristics"
	ZEEBE_ELEMENT_SCRIPT                = "script"
	ZEEBE_ELEMENT_PUBLISH_MESSAGE       = "publishMessage"
	ZEEBE_ELEMENT_TASK_SCHEDULE         = "taskSchedule"
	ZEEBE_ELEMENT_AD_HOC                = "adHoc"
)

// Zeebe attribute names.
const (
	ZEEBE_ATTRIBUTE_TYPE                           = "type"
	ZEEBE_ATTRIBUTE_RETRIES                        = "retries"
	ZEEBE_ATTRIBUTE_SOURCE                         = "source"
	ZEEBE_ATTRIBUTE_TARGET                         = "target"
	ZEEBE_ATTRIBUTE_KEY                            = "key"
	ZEEBE_ATTRIBUTE_VALUE                          = "value"
	ZEEBE_ATTRIBUTE_CORRELATION_KEY                = "correlationKey"
	ZEEBE_ATTRIBUTE_PROCESS_ID                     = "processId"
	ZEEBE_ATTRIBUTE_PROPAGATE_ALL_CHILD_VARIABLES  = "propagateAllChildVariables"
	ZEEBE_ATTRIBUTE_PROPAGATE_ALL_PARENT_VARIABLES = "propagateAllParentVariables"
	ZEEBE_ATTRIBUTE_DECISION_ID                    = "decisionId"
	ZEEBE_ATTRIBUTE_RESULT_VARIABLE                = "resultVariable"
	ZEEBE_ATTRIBUTE_FORM_ID                        = "formId"
	ZEEBE_ATTRIBUTE_FORM_KEY                       = "formKey"
	ZEEBE_ATTRIBUTE_ASSIGNEE                       = "assignee"
	ZEEBE_ATTRIBUTE_CANDIDATE_GROUPS               = "candidateGroups"
	ZEEBE_ATTRIBUTE_CANDIDATE_USERS                = "candidateUsers"
	ZEEBE_ATTRIBUTE_EVENT_TYPE                     = "eventType"
	ZEEBE_ATTRIBUTE_NAME                           = "name"
	ZEEBE_ATTRIBUTE_INPUT_COLLECTION               = "inputCollection"
	ZEEBE_ATTRIBUTE_INPUT_ELEMENT                  = "inputElement"
	ZEEBE_ATTRIBUTE_OUTPUT_COLLECTION              = "outputCollection"
	ZEEBE_ATTRIBUTE_OUTPUT_ELEMENT                 = "outputElement"
	ZEEBE_ATTRIBUTE_EXPRESSION                     = "expression"
	ZEEBE_ATTRIBUTE_MESSAGE_ID                     = "messageId"
	ZEEBE_ATTRIBUTE_DUE_DATE                       = "dueDate"
	ZEEBE_ATTRIBUTE_FOLLOW_UP_DATE                 = "followUpDate"
	ZEEBE_ATTRIBUTE_ACTIVE_ELEMENTS_COLLECTION     = "activeElementsCollection"
)
