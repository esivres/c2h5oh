package bpmn_asserts

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

// DeploymentAssert provides fluent assertions on a Zeebe deployment.
type DeploymentAssert struct {
	t             testing.TB
	deploymentKey int64
	stream        *RecordStream
	ctx           context.Context
}

// ForDeployment creates a new DeploymentAssert for the given deployment key.
func ForDeployment(t testing.TB, stream *RecordStream, deploymentKey int64) *DeploymentAssert {
	t.Helper()
	return &DeploymentAssert{
		t:             t,
		deploymentKey: deploymentKey,
		stream:        stream,
		ctx:           context.Background(),
	}
}

// WithContext sets the timeout context for waiting assertions.
func (a *DeploymentAssert) WithContext(ctx context.Context) *DeploymentAssert {
	a.ctx = ctx
	return a
}

// ContainsProcessesByBpmnProcessID asserts that the deployment contains processes
// with all the specified BPMN process IDs.
func (a *DeploymentAssert) ContainsProcessesByBpmnProcessID(expectedIDs ...string) *DeploymentAssert {
	a.t.Helper()
	v := a.getDeploymentValue()
	found := make(map[string]bool)
	for _, pm := range v.ProcessesMetadata {
		found[pm.BpmnProcessID] = true
	}
	for _, id := range expectedIDs {
		if !found[id] {
			a.t.Fatalf("expected deployment %d to contain process %q, but deployed processes are %v",
				a.deploymentKey, id, deployedProcessIDs(v))
		}
	}
	return a
}

// ContainsProcessesByResourceName asserts that the deployment contains processes
// with all the specified resource names.
func (a *DeploymentAssert) ContainsProcessesByResourceName(expectedNames ...string) *DeploymentAssert {
	a.t.Helper()
	v := a.getDeploymentValue()
	found := make(map[string]bool)
	for _, pm := range v.ProcessesMetadata {
		found[pm.ResourceName] = true
	}
	for _, name := range expectedNames {
		if !found[name] {
			a.t.Fatalf("expected deployment %d to contain process with resource %q, but it does not",
				a.deploymentKey, name)
		}
	}
	return a
}

// ExtractingProcessByBpmnProcessID returns a ProcessDefinitionAssert for the process
// with the specified BPMN process ID within this deployment.
func (a *DeploymentAssert) ExtractingProcessByBpmnProcessID(bpmnProcessID string) *ProcessDefinitionAssert {
	a.t.Helper()
	v := a.getDeploymentValue()
	for _, pm := range v.ProcessesMetadata {
		if pm.BpmnProcessID == bpmnProcessID {
			return ForProcessDefinition(a.t, a.stream, pm.ProcessDefinitionKey).WithContext(a.ctx)
		}
	}
	a.t.Fatalf("expected deployment %d to contain process %q, but deployed processes are %v",
		a.deploymentKey, bpmnProcessID, deployedProcessIDs(v))
	return nil
}

// ExtractingProcessByResourceName returns a ProcessDefinitionAssert for the process
// with the specified resource name within this deployment.
func (a *DeploymentAssert) ExtractingProcessByResourceName(resourceName string) *ProcessDefinitionAssert {
	a.t.Helper()
	v := a.getDeploymentValue()
	for _, pm := range v.ProcessesMetadata {
		if pm.ResourceName == resourceName {
			return ForProcessDefinition(a.t, a.stream, pm.ProcessDefinitionKey).WithContext(a.ctx)
		}
	}
	a.t.Fatalf("expected deployment %d to contain process with resource %q, but it does not",
		a.deploymentKey, resourceName)
	return nil
}

// ExtractingFormByFormID returns a FormAssert for the form with the specified form ID.
func (a *DeploymentAssert) ExtractingFormByFormID(formID string) *FormAssert {
	a.t.Helper()
	v := a.getDeploymentValue()
	for _, fm := range v.FormMetadata {
		if fm.FormID == formID {
			return ForForm(a.t, a.stream, fm.FormKey).WithContext(a.ctx)
		}
	}
	a.t.Fatalf("expected deployment %d to contain form %q, but it does not", a.deploymentKey, formID)
	return nil
}

// ExtractingFormByResourceName returns a FormAssert for the form with the specified resource name.
func (a *DeploymentAssert) ExtractingFormByResourceName(resourceName string) *FormAssert {
	a.t.Helper()
	v := a.getDeploymentValue()
	for _, fm := range v.FormMetadata {
		if fm.ResourceName == resourceName {
			return ForForm(a.t, a.stream, fm.FormKey).WithContext(a.ctx)
		}
	}
	a.t.Fatalf("expected deployment %d to contain form with resource %q, but it does not",
		a.deploymentKey, resourceName)
	return nil
}

func (a *DeploymentAssert) getDeploymentValue() DeploymentValue {
	a.t.Helper()
	rec, err := a.stream.WaitFor(a.ctx, func(r Record) bool {
		return r.ValueType == ValueTypeDeployment && r.Intent == IntentCreated && r.Key == a.deploymentKey
	})
	if err != nil {
		a.t.Fatalf("expected to find CREATED record for deployment %d, but timed out waiting", a.deploymentKey)
	}
	var v DeploymentValue
	if err := json.Unmarshal(rec.Value, &v); err != nil {
		a.t.Fatalf("failed to unmarshal deployment value: %v", err)
	}
	return v
}

func deployedProcessIDs(v DeploymentValue) []string {
	ids := make([]string, 0, len(v.ProcessesMetadata))
	for _, pm := range v.ProcessesMetadata {
		ids = append(ids, pm.BpmnProcessID)
	}
	return ids
}

// String returns a debug description.
func (a *DeploymentAssert) String() string {
	return fmt.Sprintf("DeploymentAssert{key=%d}", a.deploymentKey)
}
