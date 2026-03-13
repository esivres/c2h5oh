package gateway

import "sync"

// CompletionNotifier provides a pub/sub mechanism for process instance completion events.
// Callers subscribe by process instance key and receive a notification when the process completes.
type CompletionNotifier struct {
	mu   sync.Mutex
	subs map[uint64][]chan struct{}
}

// NewCompletionNotifier creates a new completion notifier.
func NewCompletionNotifier() *CompletionNotifier {
	return &CompletionNotifier{
		subs: make(map[uint64][]chan struct{}),
	}
}

// Subscribe returns a channel that receives a signal when the given process instance completes.
func (n *CompletionNotifier) Subscribe(piKey uint64) chan struct{} {
	ch := make(chan struct{}, 1)
	n.mu.Lock()
	n.subs[piKey] = append(n.subs[piKey], ch)
	n.mu.Unlock()
	return ch
}

// Unsubscribe removes a previously subscribed channel.
func (n *CompletionNotifier) Unsubscribe(piKey uint64, ch chan struct{}) {
	n.mu.Lock()
	defer n.mu.Unlock()
	subs := n.subs[piKey]
	for i, s := range subs {
		if s == ch {
			n.subs[piKey] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	if len(n.subs[piKey]) == 0 {
		delete(n.subs, piKey)
	}
}

// Notify sends a non-blocking signal to all subscribers of the given process instance key.
func (n *CompletionNotifier) Notify(piKey uint64) {
	n.mu.Lock()
	subs := make([]chan struct{}, len(n.subs[piKey]))
	copy(subs, n.subs[piKey])
	n.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
