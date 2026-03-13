package intent

import "time"

// PublishMessageIntent requests publishing a message for correlation.
type PublishMessageIntent struct {
	Header

	// MessageName is the BPMN message name.
	MessageName string

	// CorrelationKey is the business key for correlation.
	CorrelationKey string

	// Variables is the message payload (JSON bytes).
	Variables []byte

	// TTL is how long the message lives if no subscription matches.
	TTL time.Duration
}

func (i *PublishMessageIntent) IntentType() Type { return PublishMessage }

// OpenSubscriptionIntent requests opening a message subscription.
type OpenSubscriptionIntent struct {
	Header

	// ElementInstanceKey is the element instance waiting for this message.
	ElementInstanceKey uint64

	// MessageName is the BPMN message name.
	MessageName string

	// CorrelationKey is the business key for correlation.
	CorrelationKey string
}

func (i *OpenSubscriptionIntent) IntentType() Type { return OpenSubscription }

// CorrelateMessageIntent signals that a message matched a subscription.
type CorrelateMessageIntent struct {
	Header

	// SubscriptionKey is the key of the matched subscription.
	SubscriptionKey uint64

	// MessageBufferKey is the key of the buffered message (0 if direct correlation).
	MessageBufferKey uint64

	// Variables is the message payload to merge into scope (JSON bytes).
	Variables []byte
}

func (i *CorrelateMessageIntent) IntentType() Type { return CorrelateMessage }

// CloseSubscriptionIntent requests closing a message subscription.
type CloseSubscriptionIntent struct {
	Header

	// SubscriptionKey is the key of the subscription to close.
	SubscriptionKey uint64
}

func (i *CloseSubscriptionIntent) IntentType() Type { return CloseSubscription }
