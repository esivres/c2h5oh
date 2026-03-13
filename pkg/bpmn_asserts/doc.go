// Package bpmn_assertions provides fluent test assertions and event waiting
// for Zeebe process instances, jobs, incidents, deployments, messages, forms,
// and process definitions.
//
// It is the Go equivalent of the Java zeebe-process-test assertions library.
// Records are collected from Zeebe's DebugLogExporter via the ContainerSuite's
// RecordStream, which parses JSON from container logs in real time.
//
// # Quick Start
//
// All assertions are created via For* constructors and support fluent chaining.
// Use WithContext to set a timeout for waiting assertions.
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	bpmn_assertions.ForProcessInstance(s.T(), s.RecordStream, processInstanceKey).
//	    WithContext(ctx).
//	    IsStarted().
//	    HasPassedElement("serviceTask1").
//	    HasPassedElementsInOrder("start", "serviceTask1", "end").
//	    IsCompleted().
//	    HasNoIncidents().
//	    HasVariableWithValue("result", "expected")
//
// # EventFilter — ожидание событий
//
// EventFilter — fluent-билдер для ожидания конкретных событий из RecordStream.
// Поддерживает таймауты и может возвращать предикат для WaitForExportedEvent.
//
// Шортхенды для типичных сценариев:
//
//	// Ждём завершения процесса
//	record, err := bpmn_assertions.Events().
//	    ProcessInstanceCompleted(processInstanceKey).
//	    WaitOn(s.RecordStream, 30*time.Second)
//
//	// Ждём прохождения элемента
//	record, err := bpmn_assertions.Events().
//	    ElementCompleted(processInstanceKey, "serviceTask1").
//	    WaitOn(s.RecordStream, 30*time.Second)
//
//	// Ждём создания джобы
//	record, err := bpmn_assertions.Events().
//	    JobCreated("kamundarf:adhoc:v1").
//	    WaitOn(s.RecordStream, 10*time.Second)
//
//	// Ждём переменную
//	record, err := bpmn_assertions.Events().
//	    VariableCreated(processInstanceKey, "result").
//	    WaitOn(s.RecordStream, 15*time.Second)
//
//	// Ждём инцидент
//	record, err := bpmn_assertions.Events().
//	    IncidentCreated(processInstanceKey).
//	    WaitOn(s.RecordStream, 10*time.Second)
//
//	// Ждём корреляцию сообщения
//	record, err := bpmn_assertions.Events().
//	    MessageCorrelated(processInstanceKey, "paymentReceived").
//	    WaitOn(s.RecordStream, 20*time.Second)
//
// Гибкие фильтры:
//
//	record, err := bpmn_assertions.Events().
//	    ValueType("JOB").
//	    Intent("FAILED").
//	    JobType("kamundarf:askai:v1").
//	    ProcessInstanceKey(key).
//	    WaitOn(s.RecordStream, 30*time.Second)
//
//	// Кастомный предикат
//	record, err := bpmn_assertions.Events().
//	    ValueType("PROCESS_INSTANCE").
//	    Where(func(r bpmn_assertions.Record) bool {
//	        return r.Key > 100
//	    }).
//	    WaitOnCtx(s.RecordStream, ctx)
//
// Предикат для WaitForExportedEvent:
//
//	pred := bpmn_assertions.Events().
//	    ProcessInstanceCompleted(processInstanceKey).
//	    RawPredicate()
//	raw, err := s.WaitForExportedEvent(ctx, pred)
//
// Point-in-time запросы:
//
//	records := bpmn_assertions.Events().
//	    ValueType("VARIABLE").ProcessInstanceKey(key).
//	    FindAll(s.RecordStream)
//
//	count := bpmn_assertions.Events().
//	    ElementCompleted(key, "loopTask").
//	    Count(s.RecordStream)
//
// Доступные шортхенды EventFilter:
//   - ProcessInstanceActivated, ProcessInstanceCompleted, ProcessInstanceTerminated
//   - ElementActivated, ElementCompleted
//   - JobCreated, JobCompleted
//   - IncidentCreated, IncidentResolved
//   - VariableCreated, VariableUpdated
//   - MessageCorrelated
//   - DeploymentCreated
//
// Доступные фильтры EventFilter:
//   - ValueType, Intent, RecordType, WithKey, Where
//   - ProcessInstanceKey, ElementID, ElementType, BpmnProcessID
//   - VariableName, JobType, MessageName, ErrorType
//
// Выходы EventFilter:
//   - RecordPredicate — func(Record) bool для RecordStream.WaitFor / Filter
//   - RawPredicate — func(json.RawMessage) bool для WaitForExportedEvent
//   - WaitOn(stream, timeout) — ждать с таймаутом
//   - WaitOnCtx(stream, ctx) — ждать с контекстом
//   - FindAll, FindFirst, Count — point-in-time запросы
//
// # Waiting vs Point-in-Time (assertions)
//
// Positive assertions (IsCompleted, HasPassedElement, HasVariable, etc.) WAIT
// until the condition is met or the context times out.
//
// Negative assertions (IsNotCompleted, HasNotPassedElement, HasNoIncidents, etc.)
// check the current state immediately (point-in-time). Call them AFTER a positive
// assertion that establishes a known state:
//
//	bpmn_assertions.ForProcessInstance(s.T(), s.RecordStream, key).
//	    WithContext(ctx).
//	    IsCompleted().              // waits until completed
//	    HasNotPassedElement("x").   // safe: process is done
//	    HasNoIncidents()            // safe: process is done
//
// # Available Asserts
//
// ProcessInstanceAssert (ForProcessInstance):
//   - Lifecycle: IsStarted, IsActive, IsCompleted, IsNotCompleted, IsTerminated, IsNotTerminated
//   - Elements: HasPassedElement, HasPassedElementTimes, HasNotPassedElement, HasPassedElementsInOrder
//   - Wait state: IsWaitingAtElements, IsNotWaitingAtElements, IsWaitingExactlyAtElements
//   - Variables: HasVariable, HasVariableWithValue, ExtractVariables, ExtractingVariablesAssert
//   - Incidents: HasAnyIncidents, HasNoIncidents, ExtractLatestIncident, ExtractingLatestIncident
//   - Messages: IsWaitingForMessages, IsNotWaitingForMessages
//   - Correlation: HasCorrelatedMessageByName, HasCorrelatedMessageByCorrelationKey
//   - Call activities: HasCalledProcess, HasNotCalledProcess, ExtractingLatestCalledProcess
//
// JobAssert (ForJob):
//   - HasElementID, HasBpmnProcessID, HasRetries, HasDeadline, HasType
//   - HasAnyIncidents, HasNoIncidents, ExtractingLatestIncident
//   - ExtractingVariables, ExtractingHeaders
//
// IncidentAssert (ForIncident):
//   - HasErrorType, HasErrorMessage, ExtractingErrorMessage
//   - WasRaisedInProcessInstance, OccurredOnElement, OccurredDuringJob
//   - IsResolved, IsUnresolved
//
// DeploymentAssert (ForDeployment):
//   - ContainsProcessesByBpmnProcessID, ContainsProcessesByResourceName
//   - ExtractingProcessByBpmnProcessID, ExtractingProcessByResourceName
//   - ExtractingFormByFormID, ExtractingFormByResourceName
//
// MessageAssert (ForMessage):
//   - HasBeenCorrelated, HasNotBeenCorrelated
//   - HasCreatedProcessInstance, HasNotCreatedProcessInstance
//   - HasExpired, HasNotExpired
//   - ExtractingProcessInstance
//
// ProcessDefinitionAssert (ForProcessDefinition):
//   - HasBpmnProcessID, HasVersion, HasResourceName
//   - HasAnyInstances, HasNoInstances, HasInstances
//
// FormAssert (ForForm):
//   - HasFormID, HasFormKey, HasVersion, HasResourceName
//
// VariablesMapAssert (ForVariables):
//   - ContainsVariable, DoesNotContainVariable, HasVariableWithValue
//   - HasSize, IsEmpty, Raw
//
// # Inspections
//
// For finding process instances started indirectly (timers, call activities):
//
//	// Find process started by a timer
//	instance := bpmn_assertions.FindProcessEvents(s.RecordStream).
//	    TriggeredByTimer("timerStartEvent").
//	    FindFirstProcessInstance()
//	instance.AssertThat(s.T(), s.RecordStream).
//	    WithContext(ctx).
//	    IsCompleted()
//
//	// Find child processes started via call activity
//	child := bpmn_assertions.FindProcessInstances(s.RecordStream).
//	    WithParentProcessInstanceKey(parentKey).
//	    WithBpmnProcessID("childProcess").
//	    FindLastProcessInstance()
//	child.AssertThat(s.T(), s.RecordStream).
//	    WithContext(ctx).
//	    IsCompleted()
//
// # Compact Log
//
// PrintCompact выводит компактный лог всех записей и сводку нерезолвленных инцидентов
// (аналог Java RecordStream.print(true)):
//
//	s.T().Log(s.RecordStream.PrintCompact())
//
// # Integration with ContainerSuite
//
// ContainerSuite automatically creates a RecordStream and feeds it from the
// container's DebugLogExporter logs. Access it via s.RecordStream:
//
//	type myTestSuite struct {
//	    *containersuite.ContainerSuite
//	    client zbc.Client
//	}
//
//	func (s *myTestSuite) TestProcess() {
//	    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	    defer cancel()
//
//	    deployment, _ := s.client.NewDeployResourceCommand().
//	        AddResource(bpmn, "test.bpmn").Send(ctx)
//	    result, _ := s.client.NewCreateInstanceCommand().
//	        ProcessDefinitionKey(deployment.GetDeployments()[0].GetProcess().GetProcessDefinitionKey()).
//	        Send(ctx)
//
//	    // EventFilter: wait for specific events
//	    bpmn_assertions.Events().
//	        ProcessInstanceCompleted(result.GetProcessInstanceKey()).
//	        WaitOn(s.RecordStream, 30*time.Second)
//
//	    // Assertions: fluent checks on process instance
//	    bpmn_assertions.ForProcessInstance(s.T(), s.RecordStream, result.GetProcessInstanceKey()).
//	        WithContext(ctx).
//	        IsCompleted().
//	        HasPassedElement("serviceTask1").
//	        HasNoIncidents()
//
//	    // Debug: print compact record log
//	    s.T().Log(s.RecordStream.PrintCompact())
//	}
package bpmn_asserts
