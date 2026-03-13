package bpmn_asserts

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

// FormAssert provides fluent assertions on a deployed Zeebe form.
type FormAssert struct {
	t       testing.TB
	formKey int64
	stream  *RecordStream
	ctx     context.Context
}

// ForForm creates a new FormAssert for the given form key.
func ForForm(t testing.TB, stream *RecordStream, formKey int64) *FormAssert {
	t.Helper()
	return &FormAssert{
		t:       t,
		formKey: formKey,
		stream:  stream,
		ctx:     context.Background(),
	}
}

// WithContext sets the timeout context for waiting assertions.
func (a *FormAssert) WithContext(ctx context.Context) *FormAssert {
	a.ctx = ctx
	return a
}

// HasFormID asserts that the form has the specified form ID.
func (a *FormAssert) HasFormID(expectedFormID string) *FormAssert {
	a.t.Helper()
	v := a.getFormValue()
	if v.FormID != expectedFormID {
		a.t.Fatalf("expected form %d to have form ID %q, but was %q",
			a.formKey, expectedFormID, v.FormID)
	}
	return a
}

// HasFormKey asserts that the form has the specified form key.
func (a *FormAssert) HasFormKey(expectedFormKey int64) *FormAssert {
	a.t.Helper()
	if a.formKey != expectedFormKey {
		a.t.Fatalf("expected form key to be %d, but was %d", expectedFormKey, a.formKey)
	}
	return a
}

// HasVersion asserts that the form has the specified version.
func (a *FormAssert) HasVersion(expectedVersion int32) *FormAssert {
	a.t.Helper()
	v := a.getFormValue()
	if v.Version != expectedVersion {
		a.t.Fatalf("expected form %d to have version %d, but was %d",
			a.formKey, expectedVersion, v.Version)
	}
	return a
}

// HasResourceName asserts that the form has the specified resource name.
func (a *FormAssert) HasResourceName(expectedResourceName string) *FormAssert {
	a.t.Helper()
	v := a.getFormValue()
	if v.ResourceName != expectedResourceName {
		a.t.Fatalf("expected form %d to have resource name %q, but was %q",
			a.formKey, expectedResourceName, v.ResourceName)
	}
	return a
}

func (a *FormAssert) getFormValue() FormValue {
	a.t.Helper()
	rec, err := a.stream.WaitFor(a.ctx, func(r Record) bool {
		return r.ValueType == ValueTypeForm && r.Intent == IntentCreated && r.Key == a.formKey
	})
	if err != nil {
		a.t.Fatalf("expected to find CREATED record for form %d, but timed out waiting", a.formKey)
	}
	var v FormValue
	if err := json.Unmarshal(rec.Value, &v); err != nil {
		a.t.Fatalf("failed to unmarshal form value: %v", err)
	}
	return v
}

// String returns a debug description.
func (a *FormAssert) String() string {
	return fmt.Sprintf("FormAssert{key=%d}", a.formKey)
}
