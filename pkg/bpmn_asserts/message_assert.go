package bpmn_asserts

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

// MessageAssert provides fluent assertions on a published Zeebe message.
type MessageAssert struct {
	t          testing.TB
	ctx        context.Context
	stream     *RecordStream
	messageKey int64
}

// ForMessage creates a new MessageAssert for the given message key.
func ForMessage(t testing.TB, stream *RecordStream, messageKey int64) *MessageAssert {
	t.Helper()
	return &MessageAssert{
		t:          t,
		messageKey: messageKey,
		stream:     stream,
		ctx:        context.Background(),
	}
}

// WithContext sets the timeout context for waiting assertions.
func (a *MessageAssert) WithContext(ctx context.Context) *MessageAssert {
	a.ctx = ctx
	return a
}

// HasBeenCorrelated asserts that the message has been correlated to a process instance
// via process message subscription.
func (a *MessageAssert) HasBeenCorrelated() *MessageAssert {
	a.t.Helper()
	keys := a.correlatedProcessInstanceKeys()
	if len(keys) == 0 {
		a.t.Fatalf("expected message %d to have been correlated, but it was not", a.messageKey)
	}
	return a
}

// HasNotBeenCorrelated asserts that the message has NOT been correlated (point-in-time).
func (a *MessageAssert) HasNotBeenCorrelated() *MessageAssert {
	a.t.Helper()
	keys := a.correlatedProcessInstanceKeys()
	if len(keys) > 0 {
		a.t.Fatalf("expected message %d NOT to have been correlated, but it was correlated to %v",
			a.messageKey, keys)
	}
	return a
}

// HasCreatedProcessInstance asserts that the message has triggered a message start event,
// creating a new process instance.
func (a *MessageAssert) HasCreatedProcessInstance() *MessageAssert {
	a.t.Helper()
	keys := a.startEventProcessInstanceKeys()
	if len(keys) == 0 {
		a.t.Fatalf("expected message %d to have created a process instance via start event, but it did not",
			a.messageKey)
	}
	return a
}

// HasNotCreatedProcessInstance asserts that the message has NOT created a process instance (point-in-time).
func (a *MessageAssert) HasNotCreatedProcessInstance() *MessageAssert {
	a.t.Helper()
	keys := a.startEventProcessInstanceKeys()
	if len(keys) > 0 {
		a.t.Fatalf("expected message %d NOT to have created a process instance, but it created %v",
			a.messageKey, keys)
	}
	return a
}

// HasExpired asserts that the message has expired.
func (a *MessageAssert) HasExpired() *MessageAssert {
	a.t.Helper()
	_, err := a.stream.WaitFor(a.ctx, func(r Record) bool {
		return r.ValueType == ValueTypeMessage && r.Intent == IntentExpired && r.Key == a.messageKey
	})
	if err != nil {
		a.t.Fatalf("expected message %d to have expired, but timed out waiting", a.messageKey)
	}
	return a
}

// HasNotExpired asserts that the message has NOT expired (point-in-time).
func (a *MessageAssert) HasNotExpired() *MessageAssert {
	a.t.Helper()
	expired := a.stream.Filter(func(r Record) bool {
		return r.ValueType == ValueTypeMessage && r.Intent == IntentExpired && r.Key == a.messageKey
	})
	if len(expired) > 0 {
		a.t.Fatalf("expected message %d NOT to have expired, but it has", a.messageKey)
	}
	return a
}

// ExtractingProcessInstance returns a ProcessInstanceAssert for the process instance
// that the message was correlated to. Fails if the message was not correlated or
// was correlated to multiple process instances.
func (a *MessageAssert) ExtractingProcessInstance() *ProcessInstanceAssert {
	a.t.Helper()
	keys := a.correlatedProcessInstanceKeys()
	startKeys := a.startEventProcessInstanceKeys()
	keys = append(keys, startKeys...)
	if len(keys) == 0 {
		a.t.Fatalf("expected message %d to have been correlated or created a process instance, but it did not",
			a.messageKey)
	}
	return ForProcessInstance(a.t, a.stream, keys[len(keys)-1]).WithContext(a.ctx)
}

func (a *MessageAssert) correlatedProcessInstanceKeys() []int64 {
	records := a.stream.Filter(func(r Record) bool {
		if r.ValueType != ValueTypeProcessMessageSubscription || r.Intent != IntentCorrelated || r.RecordType != RecordTypeEvent {
			return false
		}
		var v ProcessMessageSubscriptionValue
		if json.Unmarshal(r.Value, &v) != nil {
			return false
		}
		return v.MessageKey == a.messageKey
	})
	keys := make([]int64, 0, len(records))
	for _, r := range records {
		var v ProcessMessageSubscriptionValue
		if json.Unmarshal(r.Value, &v) == nil {
			keys = append(keys, v.ProcessInstanceKey)
		}
	}
	return keys
}

func (a *MessageAssert) startEventProcessInstanceKeys() []int64 {
	records := a.stream.Filter(func(r Record) bool {
		if r.ValueType != ValueTypeMessageStartEventSub || r.Intent != IntentCorrelated || r.RecordType != RecordTypeEvent {
			return false
		}
		var v MessageStartEventSubscriptionValue
		if json.Unmarshal(r.Value, &v) != nil {
			return false
		}
		return v.MessageKey == a.messageKey
	})
	keys := make([]int64, 0, len(records))
	for _, r := range records {
		var v MessageStartEventSubscriptionValue
		if json.Unmarshal(r.Value, &v) == nil && v.ProcessInstanceKey > 0 {
			keys = append(keys, v.ProcessInstanceKey)
		}
	}
	return keys
}

// String returns a debug description.
func (a *MessageAssert) String() string {
	return fmt.Sprintf("MessageAssert{key=%d}", a.messageKey)
}
