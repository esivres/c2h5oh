package storage

import (
	"context"
	"time"
)

// MessageSubscriptionState represents the lifecycle of a message subscription.
type MessageSubscriptionState int

const (
	MessageSubscriptionOpened MessageSubscriptionState = iota
	MessageSubscriptionCorrelated
	MessageSubscriptionClosed
)

// MessageSubscription is a mutable entity representing a pending message catch event.
type MessageSubscription struct {
	// Key is the unique identifier.
	Key uint64

	// ProcessInstanceKey links to the owning process instance.
	ProcessInstanceKey uint64

	// ElementInstanceKey links to the element instance waiting for this message.
	ElementInstanceKey uint64

	// MessageName is the BPMN message name to correlate on.
	MessageName string

	// CorrelationKey is the business key for correlation.
	CorrelationKey string

	// State is the current lifecycle state.
	State MessageSubscriptionState

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time
}

// MessageBuffer is a buffered message waiting for a matching subscription.
type MessageBuffer struct {
	// Key is the unique identifier.
	Key uint64

	// MessageName is the BPMN message name.
	MessageName string

	// CorrelationKey is the business key for correlation.
	CorrelationKey string

	// Variables is the serialized message payload (JSON bytes).
	Variables []byte

	// ExpiresAt is the TTL deadline after which the message is dropped.
	ExpiresAt time.Time

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time
}

// MessageSubscriptionRepository manages message subscriptions and buffered messages.
type MessageSubscriptionRepository interface {
	// CreateSubscription stores a new message subscription.
	CreateSubscription(ctx context.Context, sub *MessageSubscription) error

	// Correlate marks a subscription as correlated.
	Correlate(ctx context.Context, key uint64) error

	// CloseSubscription marks a subscription as closed.
	CloseSubscription(ctx context.Context, key uint64) error

	// FindOpenSubscriptions finds open subscriptions matching name and correlation key.
	FindOpenSubscriptions(ctx context.Context, messageName string, correlationKey string) ([]*MessageSubscription, error)

	// FindByProcessInstance returns all subscriptions for a process instance.
	FindByProcessInstance(ctx context.Context, piKey uint64) ([]*MessageSubscription, error)

	// BufferMessage stores a message for later correlation.
	BufferMessage(ctx context.Context, msg *MessageBuffer) error

	// FindBufferedMessages finds buffered messages matching name and correlation key.
	// Only returns messages where ExpiresAt > now.
	FindBufferedMessages(ctx context.Context, messageName string, correlationKey string) ([]*MessageBuffer, error)

	// CleanExpired removes buffered messages where ExpiresAt <= now.
	CleanExpired(ctx context.Context, now time.Time) (int64, error)
}
