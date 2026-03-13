package bpmn_model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDataDir = "/home/aleksandr/projects/krf-adhoc/internal/pkg/test/testdata"

// TestStress_ParseSimpleTest парсит simpleTest.bpmn: start -> end
func TestStress_ParseSimpleTest(t *testing.T) {
	bmi := readTestFile(t, "simpleTest.bpmn")

	procs := GetTypedElements[Process](bmi.ModelInstance)
	require.Len(t, procs, 1)
	assert.Equal(t, "Process_10reg3x", procs[0].GetId())
	assert.True(t, procs[0].IsExecutable())

	starts := GetTypedElements[StartEvent](bmi.ModelInstance)
	assert.Len(t, starts, 1)
	assert.Equal(t, "StartEvent_1", starts[0].GetId())

	ends := GetTypedElements[EndEvent](bmi.ModelInstance)
	assert.Len(t, ends, 1)
	assert.Equal(t, "simpleTest", ends[0].GetId())

	flows := GetTypedElements[SequenceFlow](bmi.ModelInstance)
	assert.Len(t, flows, 1)
	assert.Equal(t, "StartEvent_1", flows[0].GetSourceRef())
	assert.Equal(t, "simpleTest", flows[0].GetTargetRef())

	// DI
	diagrams := GetTypedElements[BpmnDiagram](bmi.ModelInstance)
	require.Len(t, diagrams, 1)
	plane := diagrams[0].GetPlane()
	assert.Equal(t, "Process_10reg3x", plane.GetBpmnElement())
	assert.Len(t, plane.GetShapes(), 2)
	assert.Len(t, plane.GetEdges(), 1)

	// Round-trip
	assertRoundTrip(t, bmi)
}

// TestStress_ParseSimpleUserTask парсит simple-user-task.bpmn: start -> userTask -> end
func TestStress_ParseSimpleUserTask(t *testing.T) {
	bmi := readTestFile(t, "simple-user-task.bpmn")

	procs := GetTypedElements[Process](bmi.ModelInstance)
	require.Len(t, procs, 1)

	userTasks := GetTypedElements[UserTask](bmi.ModelInstance)
	require.Len(t, userTasks, 1)
	ut := userTasks[0]
	assert.Equal(t, "validateDraft", ut.GetId())
	assert.Equal(t, "Проверка черновика", ut.GetName())

	// Zeebe ioMapping
	ioMapping, ok := GetSingleExtensionElement[ZeebeIoMapping](ut)
	require.True(t, ok, "UserTask should have zeebe:ioMapping")
	_ = ioMapping

	// Zeebe formDefinition
	formDef, ok := GetSingleExtensionElement[ZeebeFormDefinition](ut)
	require.True(t, ok, "UserTask should have zeebe:formDefinition")
	assert.Contains(t, formDef.GetFormKey(), "UserTaskForm_2qnsr41")

	flows := GetTypedElements[SequenceFlow](bmi.ModelInstance)
	assert.Len(t, flows, 2)

	assertRoundTrip(t, bmi)
}

// TestStress_ParseSimpleAskAi парсит simple_ask_ai_test.bpmn: start -> serviceTask -> end
func TestStress_ParseSimpleAskAi(t *testing.T) {
	bmi := readTestFile(t, "simple_ask_ai_test.bpmn")

	procs := GetTypedElements[Process](bmi.ModelInstance)
	require.Len(t, procs, 1)
	assert.Equal(t, "simpleAiTest", procs[0].GetId())

	svcTasks := GetTypedElements[ServiceTask](bmi.ModelInstance)
	require.Len(t, svcTasks, 1)
	svc := svcTasks[0]
	assert.Equal(t, "Activity_01ran10", svc.GetId())
	assert.Equal(t, "CallAi", svc.GetName())

	// Zeebe taskDefinition
	td, ok := GetSingleExtensionElement[ZeebeTaskDefinition](svc)
	require.True(t, ok)
	assert.Equal(t, "kamundarf:askai:v1", td.GetType())

	// Zeebe ioMapping
	ioMapping, ok := GetSingleExtensionElement[ZeebeIoMapping](svc)
	require.True(t, ok)
	_ = ioMapping

	// DI
	diagrams := GetTypedElements[BpmnDiagram](bmi.ModelInstance)
	require.Len(t, diagrams, 1)
	plane := diagrams[0].GetPlane()
	assert.Len(t, plane.GetShapes(), 3) // start, serviceTask, end
	assert.Len(t, plane.GetEdges(), 2)  // 2 flows

	assertRoundTrip(t, bmi)
}

// TestStress_ParseSimpleDadataQuery парсит simple_dadata_query.bpmn: start -> serviceTask -> end (с taskHeaders, properties)
func TestStress_ParseSimpleDadataQuery(t *testing.T) {
	bmi := readTestFile(t, "simple_dadata_query.bpmn")

	procs := GetTypedElements[Process](bmi.ModelInstance)
	require.Len(t, procs, 1)
	assert.Equal(t, "simpleQuery", procs[0].GetId())

	svcTasks := GetTypedElements[ServiceTask](bmi.ModelInstance)
	require.Len(t, svcTasks, 1)
	svc := svcTasks[0]
	assert.Equal(t, "Activity_0iaay60", svc.GetId())
	assert.Equal(t, "CallDaData", svc.GetName())

	// Zeebe taskDefinition with retries
	td, ok := GetSingleExtensionElement[ZeebeTaskDefinition](svc)
	require.True(t, ok)
	assert.Equal(t, "io.camunda:http-json:1", td.GetType())
	assert.Equal(t, "3", td.GetRetries())

	// Zeebe taskHeaders
	th, ok := GetSingleExtensionElement[ZeebeTaskHeaders](svc)
	require.True(t, ok)
	_ = th

	// Zeebe properties on startEvent
	starts := GetTypedElements[StartEvent](bmi.ModelInstance)
	require.Len(t, starts, 1)
	props, ok := GetSingleExtensionElement[ZeebeProperties](starts[0])
	require.True(t, ok)
	_ = props

	assertRoundTrip(t, bmi)
}

// TestStress_ParseAdHocAiTest парсит add_hoc_aitest.bpmn: start -> adHocSubProcess -> end
// adHocSubProcess содержит manualTask и scriptTask с zeebe:script
func TestStress_ParseAdHocAiTest(t *testing.T) {
	bmi := readTestFile(t, "add_hoc_aitest.bpmn")

	procs := GetTypedElements[Process](bmi.ModelInstance)
	require.Len(t, procs, 1)
	assert.Equal(t, "simpleAiTest", procs[0].GetId())

	// AdHocSubProcess is a registered type
	adHocs := GetTypedElements[AdHocSubProcess](bmi.ModelInstance)
	require.Len(t, adHocs, 1)
	assert.Equal(t, "Activity_01ran10", adHocs[0].GetId())
	assert.Equal(t, "CallAi", adHocs[0].GetName())

	// Zeebe taskDefinition on adHocSubProcess
	td, ok := GetSingleExtensionElement[ZeebeTaskDefinition](adHocs[0])
	require.True(t, ok)
	assert.Equal(t, "kamundarf:adhoc:v1", td.GetType())

	// Zeebe adHoc extension
	adHocExt, ok := GetSingleExtensionElement[ZeebeAdHoc](adHocs[0])
	require.True(t, ok)
	assert.Equal(t, "toolCallResults", adHocExt.GetOutputCollection())

	// Children inside adHocSubProcess: manualTask and scriptTask
	flowElements := adHocs[0].GetFlowElements()
	assert.GreaterOrEqual(t, len(flowElements), 2)

	manualTasks := GetTypedElements[ManualTask](bmi.ModelInstance)
	assert.Len(t, manualTasks, 1)
	assert.Equal(t, "Activity_0shu3rf", manualTasks[0].GetId())
	assert.Equal(t, "Психолог", manualTasks[0].GetName())

	scriptTasks := GetTypedElements[ScriptTask](bmi.ModelInstance)
	assert.Len(t, scriptTasks, 1)
	assert.Equal(t, "Activity_1c68ke4", scriptTasks[0].GetId())
	assert.Equal(t, "Найти адресс", scriptTasks[0].GetName())

	// ScriptTask has zeebe:script extension
	script, ok := GetSingleExtensionElement[ZeebeScript](scriptTasks[0])
	require.True(t, ok)
	assert.NotEmpty(t, script.GetExpression())

	// Start/End events still parsed
	starts := GetTypedElements[StartEvent](bmi.ModelInstance)
	assert.Len(t, starts, 1)
	ends := GetTypedElements[EndEvent](bmi.ModelInstance)
	assert.Len(t, ends, 1)

	// SequenceFlows still parsed (the ones at process level)
	flows := GetTypedElements[SequenceFlow](bmi.ModelInstance)
	assert.GreaterOrEqual(t, len(flows), 2) // at least start->adHoc, adHoc->end

	assertRoundTrip(t, bmi)
}

// TestStress_ParseChatTests парсит chat-tests.bpmn — сложная диаграмма с:
// adHocSubProcess, boundaryEvent, intermediateCatchEvent, sendTask, manualTask, messages, textAnnotation
func TestStress_ParseChatTests(t *testing.T) {
	bmi := readTestFile(t, "chat-tests.bpmn")

	procs := GetTypedElements[Process](bmi.ModelInstance)
	require.Len(t, procs, 1)
	assert.Equal(t, "chatProcessTest", procs[0].GetId())

	// Start event with zeebe:ioMapping
	starts := GetTypedElements[StartEvent](bmi.ModelInstance)
	assert.Len(t, starts, 1)
	assert.Equal(t, "Event_1dsq551", starts[0].GetId())

	// End event
	ends := GetTypedElements[EndEvent](bmi.ModelInstance)
	assert.Len(t, ends, 1)

	// AdHocSubProcess
	adHocs := GetTypedElements[AdHocSubProcess](bmi.ModelInstance)
	require.Len(t, adHocs, 1)
	assert.Equal(t, "Activity_09c93ky", adHocs[0].GetId())

	// Zeebe adHoc extension
	adHocExt, ok := GetSingleExtensionElement[ZeebeAdHoc](adHocs[0])
	require.True(t, ok)
	assert.Equal(t, "toolCallResults", adHocExt.GetOutputCollection())

	// BoundaryEvent
	boundaryEvents := GetTypedElements[BoundaryEvent](bmi.ModelInstance)
	assert.Len(t, boundaryEvents, 1)
	assert.Equal(t, "Event_14ik0xa", boundaryEvents[0].GetId())

	// SendTask (inside adHocSubProcess — should be parsed as child)
	sendTasks := GetTypedElements[SendTask](bmi.ModelInstance)
	assert.Len(t, sendTasks, 1)
	assert.Equal(t, "sendMessageToCustomer", sendTasks[0].GetId())

	// SendTask has zeebe:taskDefinition
	td, ok := GetSingleExtensionElement[ZeebeTaskDefinition](sendTasks[0])
	require.True(t, ok)
	assert.Equal(t, "krf:gateway:sendMessage", td.GetType())

	// IntermediateCatchEvent (inside adHocSubProcess)
	catchEvents := GetTypedElements[IntermediateCatchEvent](bmi.ModelInstance)
	assert.Len(t, catchEvents, 1)
	assert.Equal(t, "Event_15p6gay", catchEvents[0].GetId())

	// ManualTask (inside adHocSubProcess)
	manualTasks := GetTypedElements[ManualTask](bmi.ModelInstance)
	assert.Len(t, manualTasks, 1)

	// Messages (top-level)
	messages := GetTypedElements[Message](bmi.ModelInstance)
	assert.GreaterOrEqual(t, len(messages), 3) // several messages defined

	// Message with zeebe:subscription
	foundCorrelation := false
	for _, msg := range messages {
		sub, ok := GetSingleExtensionElement[ZeebeSubscription](msg)
		if ok && sub.GetCorrelationKey() != "" {
			foundCorrelation = true
			assert.Equal(t, "=sessionId", sub.GetCorrelationKey())
			break
		}
	}
	assert.True(t, foundCorrelation, "Expected at least one message with zeebe:subscription correlationKey")

	// TextAnnotation
	annotations := GetTypedElements[TextAnnotation](bmi.ModelInstance)
	assert.Len(t, annotations, 1)

	// DI
	diagrams := GetTypedElements[BpmnDiagram](bmi.ModelInstance)
	require.Len(t, diagrams, 1)
	plane := diagrams[0].GetPlane()
	assert.NotEmpty(t, plane.GetShapes())
	assert.NotEmpty(t, plane.GetEdges())

	assertRoundTrip(t, bmi)
}

// TestStress_BuildSimpleTest создаёт simpleTest через билдер: start -> end
func TestStress_BuildSimpleTest(t *testing.T) {
	bmi := CreateExecutableProcess("Process_10reg3x").
		StartEvent("StartEvent_1").
		EndEvent("simpleTest").
		Done()

	// Verify structure
	procs := GetTypedElements[Process](bmi.ModelInstance)
	require.Len(t, procs, 1)
	assert.Equal(t, "Process_10reg3x", procs[0].GetId())

	starts := GetTypedElements[StartEvent](bmi.ModelInstance)
	assert.Len(t, starts, 1)

	ends := GetTypedElements[EndEvent](bmi.ModelInstance)
	assert.Len(t, ends, 1)

	flows := GetTypedElements[SequenceFlow](bmi.ModelInstance)
	assert.Len(t, flows, 1)

	// Verify XML output
	xml, err := ConvertToString(bmi)
	require.NoError(t, err)
	assert.Contains(t, xml, "Process_10reg3x")
	assert.Contains(t, xml, "StartEvent_1")
	assert.Contains(t, xml, "simpleTest")
}

// TestStress_BuildSimpleAskAi создаёт simple_ask_ai_test через билдер: start -> serviceTask -> end
func TestStress_BuildSimpleAskAi(t *testing.T) {
	bmi := CreateExecutableProcess("simpleAiTest").
		StartEvent("StartEvent_1").
		ServiceTask("Activity_01ran10").Name("CallAi").
		ZeebeJobType("kamundarf:askai:v1").
		ZeebeInput("=\"generic\"", "provider").
		ZeebeInput("=userQuery", "userQuery").
		ZeebeOutput("=stats", "stats").
		ZeebeOutput("=result", "result").
		EndEvent("Event_0xxztxs").
		Done()

	// Verify structure
	svcTasks := GetTypedElements[ServiceTask](bmi.ModelInstance)
	require.Len(t, svcTasks, 1)
	svc := svcTasks[0]
	assert.Equal(t, "CallAi", svc.GetName())

	td, ok := GetSingleExtensionElement[ZeebeTaskDefinition](svc)
	require.True(t, ok)
	assert.Equal(t, "kamundarf:askai:v1", td.GetType())

	xml, err := ConvertToString(bmi)
	require.NoError(t, err)
	assert.Contains(t, xml, "kamundarf:askai:v1")
	assert.Contains(t, xml, "provider")
}

// TestStress_BuildSimpleDadataQuery создаёт simple_dadata_query через билдер
func TestStress_BuildSimpleDadataQuery(t *testing.T) {
	bmi := CreateExecutableProcess("simpleQuery").
		StartEvent("StartEvent_1").
		ServiceTask("Activity_0iaay60").Name("CallDaData").
		ZeebeJobType("io.camunda:http-json:1").
		ZeebeJobRetries("3").
		ZeebeInput("https://suggestions.dadata.ru/suggestions", "baseUrl").
		ZeebeInput("suggestFio", "operationId").
		ZeebeTaskHeader("resultExpression", "={\"suggestions\" : body.suggestions}").
		ZeebeTaskHeader("retryBackoff", "PT0S").
		EndEvent("Event_07t2nn1").
		Done()

	svcTasks := GetTypedElements[ServiceTask](bmi.ModelInstance)
	require.Len(t, svcTasks, 1)

	td, ok := GetSingleExtensionElement[ZeebeTaskDefinition](svcTasks[0])
	require.True(t, ok)
	assert.Equal(t, "io.camunda:http-json:1", td.GetType())
	assert.Equal(t, "3", td.GetRetries())

	th, ok := GetSingleExtensionElement[ZeebeTaskHeaders](svcTasks[0])
	require.True(t, ok)
	_ = th

	xml, err := ConvertToString(bmi)
	require.NoError(t, err)
	assert.Contains(t, xml, "io.camunda:http-json:1")
	assert.Contains(t, xml, "retryBackoff")
}

// TestStress_BuildAdHocSubProcess создаёт процесс с adHocSubProcess через билдер
func TestStress_BuildAdHocSubProcess(t *testing.T) {
	bmi := CreateExecutableProcess("adHocProcess").
		StartEvent("start").
		AdHocSubProcess("adhoc").Name("AI Agent").
		ZeebeJobType("kamundarf:adhoc:v1").
		ZeebeAdHocConfig("toolCallResults", "={id: _meta.id, name: _meta.name, content: toolCallResult}").
		ZeebeInput("=userQuery", "userQuery").
		ZeebeInput("=2000", "maxOutPutTokens").
		EndEvent("end").
		Done()

	adHocs := GetTypedElements[AdHocSubProcess](bmi.ModelInstance)
	require.Len(t, adHocs, 1)
	assert.Equal(t, "AI Agent", adHocs[0].GetName())

	td, ok := GetSingleExtensionElement[ZeebeTaskDefinition](adHocs[0])
	require.True(t, ok)
	assert.Equal(t, "kamundarf:adhoc:v1", td.GetType())

	adHocExt, ok := GetSingleExtensionElement[ZeebeAdHoc](adHocs[0])
	require.True(t, ok)
	assert.Equal(t, "toolCallResults", adHocExt.GetOutputCollection())

	xml, err := ConvertToString(bmi)
	require.NoError(t, err)
	assert.Contains(t, xml, "adHocSubProcess")
	assert.Contains(t, xml, "toolCallResults")
}

// TestStress_ParseAllFiles проверяет что все 6 файлов парсятся без ошибок
func TestStress_ParseAllFiles(t *testing.T) {
	files := []string{
		"simpleTest.bpmn",
		"simple-user-task.bpmn",
		"simple_ask_ai_test.bpmn",
		"simple_dadata_query.bpmn",
		"add_hoc_aitest.bpmn",
		"chat-tests.bpmn",
	}

	for _, f := range files {
		t.Run(f, func(t *testing.T) {
			bmi := readTestFile(t, f)

			// Basic sanity: at least one process
			procs := GetTypedElements[Process](bmi.ModelInstance)
			assert.NotEmpty(t, procs, "Expected at least one process")

			// Can convert back to string
			xml, err := ConvertToString(bmi)
			require.NoError(t, err)
			assert.NotEmpty(t, xml)

			// Re-parse the converted string
			bmi2, err := ReadFromString(xml)
			require.NoError(t, err, "Re-parse of converted XML failed")

			procs2 := GetTypedElements[Process](bmi2.ModelInstance)
			assert.Equal(t, len(procs), len(procs2), "Process count mismatch after round-trip")
		})
	}
}

// --- helpers ---

func readTestFile(t *testing.T, name string) *BpmnModelInstance {
	t.Helper()
	path := filepath.Join(testDataDir, name)
	data, err := os.ReadFile(path)
	require.NoError(t, err, "Failed to read test file: %s", name)

	bmi, err := ReadFromBytes(data)
	require.NoError(t, err, "Failed to parse test file: %s", name)
	return bmi
}

func assertRoundTrip(t *testing.T, bmi *BpmnModelInstance) {
	t.Helper()
	xml, err := ConvertToString(bmi)
	require.NoError(t, err, "ConvertToString failed")

	bmi2, err := ReadFromString(xml)
	require.NoError(t, err, "Re-parse after round-trip failed")

	procs1 := GetTypedElements[Process](bmi.ModelInstance)
	procs2 := GetTypedElements[Process](bmi2.ModelInstance)
	assert.Equal(t, len(procs1), len(procs2), "Process count changed after round-trip")
}
