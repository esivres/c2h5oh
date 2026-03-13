package gateway

import "sync"

// JobNotifier provides a pub/sub mechanism for job creation events.
// Workers subscribe by job type and receive notifications when new jobs are created.
type JobNotifier struct {
	mu   sync.Mutex
	subs map[string][]chan struct{}
}

// NewJobNotifier creates a new job notifier.
func NewJobNotifier() *JobNotifier {
	return &JobNotifier{
		subs: make(map[string][]chan struct{}),
	}
}

// Subscribe returns a channel that receives a signal whenever a job of the given type is created.
// The channel is buffered(1) so a single notification is never lost, but multiple rapid
// notifications may coalesce into one (which is fine — the subscriber polls storage anyway).
func (n *JobNotifier) Subscribe(jobType string) chan struct{} {
	ch := make(chan struct{}, 1)
	n.mu.Lock()
	n.subs[jobType] = append(n.subs[jobType], ch)
	n.mu.Unlock()
	return ch
}

// Unsubscribe removes a previously subscribed channel.
func (n *JobNotifier) Unsubscribe(jobType string, ch chan struct{}) {
	n.mu.Lock()
	defer n.mu.Unlock()
	subs := n.subs[jobType]
	for i, s := range subs {
		if s == ch {
			n.subs[jobType] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	if len(n.subs[jobType]) == 0 {
		delete(n.subs, jobType)
	}
}

// Notify sends a non-blocking signal to all subscribers of the given job type.
func (n *JobNotifier) Notify(jobType string) {
	n.mu.Lock()
	subs := make([]chan struct{}, len(n.subs[jobType]))
	copy(subs, n.subs[jobType])
	n.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
