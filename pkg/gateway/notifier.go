package gateway

import "sync"

// Notifier is a generic pub/sub mechanism keyed by K.
// Subscribers receive a buffered(1) signal when Notify is called for their key.
type Notifier[K comparable] struct {
	subs map[K][]chan struct{}
	mu   sync.Mutex
}

// NewNotifier creates a new Notifier.
func NewNotifier[K comparable]() *Notifier[K] {
	return &Notifier[K]{
		subs: make(map[K][]chan struct{}),
	}
}

// Subscribe returns a channel that receives a signal when Notify is called for the given key.
func (n *Notifier[K]) Subscribe(key K) chan struct{} {
	ch := make(chan struct{}, 1)
	n.mu.Lock()
	n.subs[key] = append(n.subs[key], ch)
	n.mu.Unlock()
	return ch
}

// Unsubscribe removes a previously subscribed channel.
func (n *Notifier[K]) Unsubscribe(key K, ch chan struct{}) {
	n.mu.Lock()
	defer n.mu.Unlock()
	subs := n.subs[key]
	for i, s := range subs {
		if s == ch {
			n.subs[key] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	if len(n.subs[key]) == 0 {
		delete(n.subs, key)
	}
}

// Notify sends a non-blocking signal to all subscribers of the given key.
func (n *Notifier[K]) Notify(key K) {
	n.mu.Lock()
	subs := make([]chan struct{}, len(n.subs[key]))
	copy(subs, n.subs[key])
	n.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// CompletionNotifier notifies on process instance completion (keyed by uint64 instance key).
type CompletionNotifier = Notifier[uint64]

// DeployNotifier notifies on process deployment (keyed by string process ID).
type DeployNotifier = Notifier[string]

// JobNotifier notifies on job creation (keyed by string job type).
type JobNotifier = Notifier[string]

// NewCompletionNotifier creates a new CompletionNotifier.
func NewCompletionNotifier() *CompletionNotifier { return NewNotifier[uint64]() }

// NewDeployNotifier creates a new DeployNotifier.
func NewDeployNotifier() *DeployNotifier { return NewNotifier[string]() }

// NewJobNotifier creates a new JobNotifier.
func NewJobNotifier() *JobNotifier { return NewNotifier[string]() }
