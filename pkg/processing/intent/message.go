package intent

import "time"

// PublishMessageIntent requests publishing a message for correlation.
type PublishMessageIntent struct {
	MessageName    string
	CorrelationKey string
	Variables      []byte
	Header
	TTL time.Duration
}

func (i *PublishMessageIntent) IntentType() Type { return PublishMessage }

// OpenSubscriptionIntent requests opening a message subscription.
type OpenSubscriptionIntent struct {
	MessageName    string
	CorrelationKey string
	Header
	ElementInstanceKey uint64
}

func (i *OpenSubscriptionIntent) IntentType() Type { return OpenSubscription }

// CorrelateMessageIntent signals that a message matched a subscription.
type CorrelateMessageIntent struct {
	Variables []byte
	Header
	SubscriptionKey  uint64
	MessageBufferKey uint64
}

func (i *CorrelateMessageIntent) IntentType() Type { return CorrelateMessage }

// CloseSubscriptionIntent requests closing a message subscription.
type CloseSubscriptionIntent struct {
	Header

	// SubscriptionKey is the key of the subscription to close.
	SubscriptionKey uint64
}

func (i *CloseSubscriptionIntent) IntentType() Type { return CloseSubscription }
