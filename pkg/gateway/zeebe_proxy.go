package gateway

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"time"

	zeebepb "github.com/camunda/zeebe/clients/go/v8/pkg/pb"

	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/engine"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ZeebeProxy implements Zeebe Gateway gRPC service and translates
// requests to the internal engine.
type ZeebeProxy struct {
	zeebepb.UnimplementedGatewayServer

	processor          *engine.Processor
	store              storage.Store
	jobNotifier        *JobNotifier
	deployNotifier     *DeployNotifier
	completionNotifier *CompletionNotifier
}

// NewZeebeProxy creates a new Zeebe-compatible gateway proxy.
func NewZeebeProxy(processor *engine.Processor, store storage.Store, jobNotifier *JobNotifier, deployNotifier *DeployNotifier, completionNotifier *CompletionNotifier) *ZeebeProxy {
	return &ZeebeProxy{
		processor:          processor,
		store:              store,
		jobNotifier:        jobNotifier,
		deployNotifier:     deployNotifier,
		completionNotifier: completionNotifier,
	}
}

// --- Topology ---

func (*ZeebeProxy) Topology(_ context.Context, _ *zeebepb.TopologyRequest) (*zeebepb.TopologyResponse, error) {
	return &zeebepb.TopologyResponse{
		Brokers: []*zeebepb.BrokerInfo{
			{
				NodeId: 0,
				Host:   "localhost",
				Port:   9090,
				Partitions: []*zeebepb.Partition{
					{PartitionId: 1, Role: zeebepb.Partition_LEADER, Health: zeebepb.Partition_HEALTHY},
				},
				Version: "rumunda-1.0.0",
			},
		},
		ClusterSize:       1,
		PartitionsCount:   1,
		ReplicationFactor: 1,
		GatewayVersion:    "rumunda-1.0.0",
	}, nil
}

// --- DeployProcess (deprecated in Zeebe, but still used by clients) ---

func (z *ZeebeProxy) DeployProcess(ctx context.Context, req *zeebepb.DeployProcessRequest) (*zeebepb.DeployProcessResponse, error) {
	if len(req.Processes) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no processes given")
	}

	var metadata []*zeebepb.ProcessMetadata
	var deployKey int64

	for _, proc := range req.Processes {
		def, err := z.deploySingleProcess(ctx, proc.Name, proc.Definition)
		if err != nil {
			return nil, err
		}
		if deployKey == 0 {
			deployKey = int64(def.Key)
		}
		metadata = append(metadata, &zeebepb.ProcessMetadata{
			BpmnProcessId:        def.BpmnProcessId,
			Version:              int32(def.Version),
			ProcessDefinitionKey: int64(def.Key),
			ResourceName:         proc.Name,
		})
	}

	return &zeebepb.DeployProcessResponse{
		Key:       deployKey,
		Processes: metadata,
	}, nil
}

// --- DeployResource ---

func (z *ZeebeProxy) DeployResource(ctx context.Context, req *zeebepb.DeployResourceRequest) (*zeebepb.DeployResourceResponse, error) {
	if len(req.Resources) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no resources given")
	}

	var deployments []*zeebepb.Deployment
	var deployKey int64

	for _, res := range req.Resources {
		def, err := z.deploySingleProcess(ctx, res.Name, res.Content)
		if err != nil {
			return nil, err
		}
		if deployKey == 0 {
			deployKey = int64(def.Key)
		}
		deployments = append(deployments, &zeebepb.Deployment{
			Metadata: &zeebepb.Deployment_Process{
				Process: &zeebepb.ProcessMetadata{
					BpmnProcessId:        def.BpmnProcessId,
					Version:              int32(def.Version),
					ProcessDefinitionKey: int64(def.Key),
					ResourceName:         res.Name,
				},
			},
		})
	}

	return &zeebepb.DeployResourceResponse{
		Key:         deployKey,
		Deployments: deployments,
	}, nil
}

// --- CreateProcessInstance ---

func (z *ZeebeProxy) CreateProcessInstance(ctx context.Context, req *zeebepb.CreateProcessInstanceRequest) (*zeebepb.CreateProcessInstanceResponse, error) {
	processId := req.BpmnProcessId
	if processId == "" && req.ProcessDefinitionKey == 0 {
		return nil, status.Error(codes.InvalidArgument, "either bpmn_process_id or process_definition_key is required")
	}

	var def *storage.ProcessDefinition
	var err error

	if req.ProcessDefinitionKey != 0 {
		def, err = z.store.ProcessDefinitions().FindByKey(ctx, uint64(req.ProcessDefinitionKey))
	} else {
		def, err = z.store.ProcessDefinitions().FindLatestByProcessId(ctx, processId)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find process: %v", err)
	}
	if def == nil {
		return nil, status.Errorf(codes.NotFound, "process not found")
	}

	var variables []byte
	if req.Variables != "" {
		variables = []byte(req.Variables)
	}

	createIntent := &intent.CreateProcessInstanceIntent{
		Header:               intent.Header{Origin: intent.External},
		ProcessDefinitionKey: def.Key,
		BpmnProcessId:        def.BpmnProcessId,
		Variables:            variables,
	}
	z.processor.Submit(createIntent)

	return &zeebepb.CreateProcessInstanceResponse{
		ProcessDefinitionKey: int64(def.Key),
		BpmnProcessId:        def.BpmnProcessId,
		Version:              int32(def.Version),
		ProcessInstanceKey:   int64(createIntent.Key),
	}, nil
}

// --- CreateProcessInstanceWithResult ---

func (z *ZeebeProxy) CreateProcessInstanceWithResult(ctx context.Context, req *zeebepb.CreateProcessInstanceWithResultRequest) (*zeebepb.CreateProcessInstanceWithResultResponse, error) {
	inner := req.Request
	if inner == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	processId := inner.BpmnProcessId
	if processId == "" && inner.ProcessDefinitionKey == 0 {
		return nil, status.Error(codes.InvalidArgument, "either bpmn_process_id or process_definition_key is required")
	}

	var def *storage.ProcessDefinition
	var err error
	if inner.ProcessDefinitionKey != 0 {
		def, err = z.store.ProcessDefinitions().FindByKey(ctx, uint64(inner.ProcessDefinitionKey))
	} else {
		def, err = z.store.ProcessDefinitions().FindLatestByProcessId(ctx, processId)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find process: %v", err)
	}
	if def == nil {
		return nil, status.Errorf(codes.NotFound, "process not found")
	}

	var variables []byte
	if inner.Variables != "" {
		variables = []byte(inner.Variables)
	}

	createIntent := &intent.CreateProcessInstanceIntent{
		Header:               intent.Header{Origin: intent.External},
		ProcessDefinitionKey: def.Key,
		BpmnProcessId:        def.BpmnProcessId,
		Variables:            variables,
	}
	z.processor.Submit(createIntent)
	piKey := createIntent.Key

	notify := z.completionNotifier.Subscribe(piKey)
	defer z.completionNotifier.Unsubscribe(piKey, notify)

	timeout := 10 * time.Second
	if req.RequestTimeout > 0 {
		timeout = time.Duration(req.RequestTimeout) * time.Millisecond
	}

	deadline := time.After(timeout)
	for {
		select {
		case <-ctx.Done():
			return nil, status.Error(codes.Canceled, "request canceled")
		case <-deadline:
			return nil, status.Errorf(codes.DeadlineExceeded, "process instance %d did not complete within timeout", piKey)
		case <-notify:
			resultVars, err := collectZeebeResultVariables(ctx, z.store, piKey, req.FetchVariables)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to collect variables: %v", err)
			}

			return &zeebepb.CreateProcessInstanceWithResultResponse{
				ProcessDefinitionKey: int64(def.Key),
				BpmnProcessId:        def.BpmnProcessId,
				Version:              int32(def.Version),
				ProcessInstanceKey:   int64(piKey),
				Variables:            resultVars,
			}, nil
		}
	}
}

// collectZeebeResultVariables reads variables from the process instance scope.
func collectZeebeResultVariables(ctx context.Context, store storage.Store, piKey uint64, fetchVariables []string) (string, error) {
	vars, err := store.Variables().FindByScope(ctx, piKey)
	if err != nil {
		return "", err
	}

	filter := make(map[string]bool, len(fetchVariables))
	for _, name := range fetchVariables {
		filter[name] = true
	}

	result := make(map[string]json.RawMessage, len(vars))
	for _, v := range vars {
		if len(filter) > 0 && !filter[v.Name] {
			continue
		}
		result[v.Name] = json.RawMessage(v.Value)
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// --- CompleteJob ---

func (z *ZeebeProxy) CompleteJob(ctx context.Context, req *zeebepb.CompleteJobRequest) (*zeebepb.CompleteJobResponse, error) {
	if req.JobKey == 0 {
		return nil, status.Error(codes.InvalidArgument, "job_key is required")
	}

	job, err := z.store.Jobs().GetByKey(ctx, uint64(req.JobKey))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find job: %v", err)
	}
	if job == nil {
		return nil, status.Errorf(codes.NotFound, "job %d not found", req.JobKey)
	}

	var variables []byte
	if req.Variables != "" {
		variables = []byte(req.Variables)
	}

	completeIntent := &intent.CompleteJobIntent{
		Header: intent.Header{
			Origin:             intent.External,
			ProcessInstanceKey: job.ProcessInstanceKey,
		},
		JobKey:    uint64(req.JobKey),
		Variables: variables,
	}
	z.processor.Submit(completeIntent)

	return &zeebepb.CompleteJobResponse{}, nil
}

// --- FailJob ---

func (z *ZeebeProxy) FailJob(ctx context.Context, req *zeebepb.FailJobRequest) (*zeebepb.FailJobResponse, error) {
	if req.JobKey == 0 {
		return nil, status.Error(codes.InvalidArgument, "job_key is required")
	}

	job, err := z.store.Jobs().GetByKey(ctx, uint64(req.JobKey))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find job: %v", err)
	}
	if job == nil {
		return nil, status.Errorf(codes.NotFound, "job %d not found", req.JobKey)
	}

	failIntent := &intent.FailJobIntent{
		Header: intent.Header{
			Origin:             intent.External,
			ProcessInstanceKey: job.ProcessInstanceKey,
		},
		JobKey:       uint64(req.JobKey),
		Retries:      int(req.Retries),
		ErrorMessage: req.ErrorMessage,
		RetryBackoff: time.Duration(req.RetryBackOff) * time.Millisecond,
	}
	z.processor.Submit(failIntent)

	return &zeebepb.FailJobResponse{}, nil
}

// --- ThrowError ---

func (z *ZeebeProxy) ThrowError(ctx context.Context, req *zeebepb.ThrowErrorRequest) (*zeebepb.ThrowErrorResponse, error) {
	if req.JobKey == 0 {
		return nil, status.Error(codes.InvalidArgument, "job_key is required")
	}

	job, err := z.store.Jobs().GetByKey(ctx, uint64(req.JobKey))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find job: %v", err)
	}
	if job == nil {
		return nil, status.Errorf(codes.NotFound, "job %d not found", req.JobKey)
	}

	throwIntent := &intent.ThrowJobErrorIntent{
		Header: intent.Header{
			Origin:             intent.External,
			ProcessInstanceKey: job.ProcessInstanceKey,
		},
		JobKey:       uint64(req.JobKey),
		ErrorCode:    req.ErrorCode,
		ErrorMessage: req.ErrorMessage,
	}
	z.processor.Submit(throwIntent)

	return &zeebepb.ThrowErrorResponse{}, nil
}

// --- PublishMessage ---

func (z *ZeebeProxy) PublishMessage(_ context.Context, req *zeebepb.PublishMessageRequest) (*zeebepb.PublishMessageResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	var variables []byte
	if req.Variables != "" {
		variables = []byte(req.Variables)
	}

	ttl := 5 * time.Minute
	if req.TimeToLive > 0 {
		ttl = time.Duration(req.TimeToLive) * time.Millisecond
	}

	msgIntent := &intent.PublishMessageIntent{
		Header:         intent.Header{Origin: intent.External},
		MessageName:    req.Name,
		CorrelationKey: req.CorrelationKey,
		Variables:      variables,
		TTL:            ttl,
	}
	z.processor.Submit(msgIntent)

	return &zeebepb.PublishMessageResponse{
		Key: int64(msgIntent.Key),
	}, nil
}

// --- BroadcastSignal ---

func (z *ZeebeProxy) BroadcastSignal(_ context.Context, req *zeebepb.BroadcastSignalRequest) (*zeebepb.BroadcastSignalResponse, error) {
	if req.SignalName == "" {
		return nil, status.Error(codes.InvalidArgument, "signal_name is required")
	}

	var variables []byte
	if req.Variables != "" {
		variables = []byte(req.Variables)
	}

	signalIntent := &intent.ThrowSignalIntent{
		Header:     intent.Header{Origin: intent.External},
		SignalName: req.SignalName,
		Variables:  variables,
	}
	z.processor.Submit(signalIntent)

	return &zeebepb.BroadcastSignalResponse{
		Key: int64(signalIntent.Key),
	}, nil
}

// --- ActivateJobs (server-side stream in Zeebe) ---

func (z *ZeebeProxy) ActivateJobs(req *zeebepb.ActivateJobsRequest, stream zeebepb.Gateway_ActivateJobsServer) error {
	if req.Type == "" {
		return status.Error(codes.InvalidArgument, "type is required")
	}
	if req.Worker == "" {
		return status.Error(codes.InvalidArgument, "worker is required")
	}

	maxJobs := int(req.MaxJobsToActivate)
	if maxJobs <= 0 {
		maxJobs = 1
	}

	timeout := 5 * time.Minute
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Millisecond
	}

	ctx := stream.Context()

	jobs, err := z.store.Jobs().FindActivatable(ctx, req.Type, maxJobs)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to find jobs: %v", err)
	}

	if len(jobs) == 0 {
		// If requestTimeout > 0, wait for jobs to appear
		if req.RequestTimeout > 0 {
			jobs, err = z.waitForJobs(ctx, req.Type, maxJobs, time.Duration(req.RequestTimeout)*time.Millisecond)
			if err != nil {
				return err
			}
		}
	}

	activated := z.activateAndConvert(ctx, jobs, req.Worker, timeout)
	if len(activated) > 0 {
		if err := stream.Send(&zeebepb.ActivateJobsResponse{Jobs: activated}); err != nil {
			return err
		}
	}

	return nil
}

// --- StreamActivatedJobs ---

func (z *ZeebeProxy) StreamActivatedJobs(req *zeebepb.StreamActivatedJobsRequest, stream zeebepb.Gateway_StreamActivatedJobsServer) error {
	if req.Type == "" {
		return status.Error(codes.InvalidArgument, "type is required")
	}
	if req.Worker == "" {
		return status.Error(codes.InvalidArgument, "worker is required")
	}

	timeout := 5 * time.Minute
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Millisecond
	}

	maxJobs := 10 // Zeebe proto does not have max_jobs for streaming; use sensible default
	ctx := stream.Context()

	notify := z.jobNotifier.Subscribe(req.Type)
	defer z.jobNotifier.Unsubscribe(req.Type, notify)

	// Initial poll
	if err := z.sendZeebeJobs(ctx, stream, req.Type, req.Worker, timeout, maxJobs); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-notify:
			if err := z.sendZeebeJobs(ctx, stream, req.Type, req.Worker, timeout, maxJobs); err != nil {
				return err
			}
		}
	}
}

// --- CancelProcessInstance ---

func (z *ZeebeProxy) CancelProcessInstance(_ context.Context, req *zeebepb.CancelProcessInstanceRequest) (*zeebepb.CancelProcessInstanceResponse, error) {
	if req.ProcessInstanceKey == 0 {
		return nil, status.Error(codes.InvalidArgument, "process_instance_key is required")
	}

	cancelIntent := &intent.CancelProcessInstanceIntent{
		Header: intent.Header{
			Origin:             intent.External,
			ProcessInstanceKey: uint64(req.ProcessInstanceKey),
		},
	}
	z.processor.Submit(cancelIntent)

	return &zeebepb.CancelProcessInstanceResponse{}, nil
}

// --- ResolveIncident ---

func (z *ZeebeProxy) ResolveIncident(_ context.Context, req *zeebepb.ResolveIncidentRequest) (*zeebepb.ResolveIncidentResponse, error) {
	if req.IncidentKey == 0 {
		return nil, status.Error(codes.InvalidArgument, "incident_key is required")
	}

	resolveIntent := &intent.ResolveIncidentIntent{
		Header:      intent.Header{Origin: intent.External},
		IncidentKey: uint64(req.IncidentKey),
	}
	z.processor.Submit(resolveIntent)

	return &zeebepb.ResolveIncidentResponse{}, nil
}

// --- UpdateJobRetries ---

func (z *ZeebeProxy) UpdateJobRetries(ctx context.Context, req *zeebepb.UpdateJobRetriesRequest) (*zeebepb.UpdateJobRetriesResponse, error) {
	if req.JobKey == 0 {
		return nil, status.Error(codes.InvalidArgument, "job_key is required")
	}
	if req.Retries < 1 {
		return nil, status.Error(codes.InvalidArgument, "retries must be > 0")
	}

	job, err := z.store.Jobs().GetByKey(ctx, uint64(req.JobKey))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find job: %v", err)
	}
	if job == nil {
		return nil, status.Errorf(codes.NotFound, "job %d not found", req.JobKey)
	}

	updateIntent := &intent.UpdateJobRetriesIntent{
		Header: intent.Header{
			Origin:             intent.External,
			ProcessInstanceKey: job.ProcessInstanceKey,
		},
		JobKey:  uint64(req.JobKey),
		Retries: int(req.Retries),
	}
	z.processor.Submit(updateIntent)

	return &zeebepb.UpdateJobRetriesResponse{}, nil
}

// --- SetVariables ---

func (z *ZeebeProxy) SetVariables(_ context.Context, req *zeebepb.SetVariablesRequest) (*zeebepb.SetVariablesResponse, error) {
	if req.ElementInstanceKey == 0 {
		return nil, status.Error(codes.InvalidArgument, "element_instance_key is required")
	}

	setVarsIntent := &intent.SetVariablesIntent{
		Header: intent.Header{
			Origin: intent.External,
		},
		ScopeKey:  uint64(req.ElementInstanceKey),
		Variables: []byte(req.Variables),
	}
	z.processor.Submit(setVarsIntent)

	return &zeebepb.SetVariablesResponse{
		Key: int64(setVarsIntent.Key),
	}, nil
}

// --- UpdateJobTimeout ---

func (z *ZeebeProxy) UpdateJobTimeout(ctx context.Context, req *zeebepb.UpdateJobTimeoutRequest) (*zeebepb.UpdateJobTimeoutResponse, error) {
	if req.JobKey == 0 {
		return nil, status.Error(codes.InvalidArgument, "job_key is required")
	}

	job, err := z.store.Jobs().GetByKey(ctx, uint64(req.JobKey))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find job: %v", err)
	}
	if job == nil {
		return nil, status.Errorf(codes.NotFound, "job %d not found", req.JobKey)
	}

	updateIntent := &intent.UpdateJobTimeoutIntent{
		Header: intent.Header{
			Origin:             intent.External,
			ProcessInstanceKey: job.ProcessInstanceKey,
		},
		JobKey:  uint64(req.JobKey),
		Timeout: time.Duration(req.Timeout) * time.Millisecond,
	}
	z.processor.Submit(updateIntent)

	return &zeebepb.UpdateJobTimeoutResponse{}, nil
}

// --- DeleteResource ---

func (z *ZeebeProxy) DeleteResource(_ context.Context, req *zeebepb.DeleteResourceRequest) (*zeebepb.DeleteResourceResponse, error) {
	if req.ResourceKey == 0 {
		return nil, status.Error(codes.InvalidArgument, "resource_key is required")
	}

	deleteIntent := &intent.DeleteResourceIntent{
		Header:      intent.Header{Origin: intent.External},
		ResourceKey: uint64(req.ResourceKey),
	}
	z.processor.Submit(deleteIntent)

	return &zeebepb.DeleteResourceResponse{}, nil
}

// ==================== helpers ====================

// deploySingleProcess deploys a single BPMN resource and returns the definition.
func (z *ZeebeProxy) deploySingleProcess(ctx context.Context, name string, definition []byte) (*storage.ProcessDefinition, error) {
	if len(definition) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty definition")
	}

	bmi, err := bpmn_model.ReadFromBytes(definition)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid BPMN XML: %v", err)
	}

	processes := bpmn_model.GetTypedElements[bpmn_model.Process](bmi.ModelInstance)
	if len(processes) == 0 {
		return nil, status.Error(codes.InvalidArgument, "BPMN XML contains no process")
	}

	processId := processes[0].GetId()
	if processId == "" {
		return nil, status.Error(codes.InvalidArgument, "process has no id")
	}

	contentHash := sha256.Sum256(definition)

	// Check for existing deployment (idempotent)
	existing, err := z.store.ProcessDefinitions().FindByContentHash(ctx, processId, contentHash[:])
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check deployment: %v", err)
	}
	if existing != nil {
		return existing, nil
	}

	if name == "" {
		name = processId
	}

	deployIntent := &intent.DeployProcessIntent{
		Header:        intent.Header{Origin: intent.External},
		BpmnProcessId: processId,
		Name:          name,
		Content:       definition,
		ContentHash:   contentHash[:],
	}
	notify := z.deployNotifier.Subscribe(processId)
	defer z.deployNotifier.Unsubscribe(processId, notify)

	z.processor.Submit(deployIntent)

	// Wait for the deploy notification or timeout
	const deployTimeout = 10 * time.Second
	deadline := time.After(deployTimeout)
	for {
		select {
		case <-ctx.Done():
			return nil, status.Error(codes.Canceled, "request canceled")
		case <-deadline:
			return nil, status.Error(codes.DeadlineExceeded, "deploy timed out")
		case <-notify:
			def, err := z.store.ProcessDefinitions().FindByContentHash(ctx, processId, contentHash[:])
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to check deployment: %v", err)
			}
			if def != nil {
				return def, nil
			}
		}
	}
}

// waitForJobs waits for jobs to become available using the notifier.
func (z *ZeebeProxy) waitForJobs(ctx context.Context, jobType string, maxJobs int, timeout time.Duration) ([]*storage.Job, error) {
	notify := z.jobNotifier.Subscribe(jobType)
	defer z.jobNotifier.Unsubscribe(jobType, notify)

	deadline := time.After(timeout)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-deadline:
			return nil, nil // timeout — return empty, no error
		case <-notify:
			jobs, err := z.store.Jobs().FindActivatable(ctx, jobType, maxJobs)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to find jobs: %v", err)
			}
			if len(jobs) > 0 {
				return jobs, nil
			}
		}
	}
}

// activateAndConvert activates jobs and converts them to Zeebe ActivatedJob format.
func (z *ZeebeProxy) activateAndConvert(ctx context.Context, jobs []*storage.Job, worker string, timeout time.Duration) []*zeebepb.ActivatedJob {
	var result []*zeebepb.ActivatedJob
	for _, job := range jobs {
		activateIntent := &intent.ActivateJobIntent{
			Header: intent.Header{
				Origin:             intent.External,
				ProcessInstanceKey: job.ProcessInstanceKey,
			},
			JobKey:  job.Key,
			Worker:  worker,
			Timeout: timeout,
		}
		z.processor.Submit(activateIntent)

		var elementId string
		ei, err := z.store.ProcessInstances().GetElementInstance(ctx, job.ElementInstanceKey)
		if err == nil && ei != nil {
			elementId = ei.ElementId
		}

		var variables string
		if len(job.Variables) > 0 {
			variables = string(job.Variables)
		}

		result = append(result, &zeebepb.ActivatedJob{
			Key:                  int64(job.Key),
			Type:                 job.Type,
			ProcessInstanceKey:   int64(job.ProcessInstanceKey),
			ProcessDefinitionKey: int64(job.ProcessDefinitionKey),
			ElementId:            elementId,
			ElementInstanceKey:   int64(job.ElementInstanceKey),
			Worker:               worker,
			Retries:              int32(job.Retries),
			Deadline:             job.Deadline.UnixMilli(),
			Variables:            variables,
		})
	}
	return result
}

// sendZeebeJobs finds, activates, and streams jobs to the Zeebe client.
func (z *ZeebeProxy) sendZeebeJobs(ctx context.Context, stream zeebepb.Gateway_StreamActivatedJobsServer, jobType, worker string, timeout time.Duration, maxJobs int) error {
	jobs, err := z.store.Jobs().FindActivatable(ctx, jobType, maxJobs)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to find jobs: %v", err)
	}

	activated := z.activateAndConvert(ctx, jobs, worker, timeout)
	for _, aj := range activated {
		if err := stream.Send(aj); err != nil {
			return err
		}
	}
	return nil
}
