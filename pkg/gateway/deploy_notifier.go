package gateway

import "sync"

// DeployNotifier provides a pub/sub mechanism for process deployment events.
// Callers subscribe by processId and receive a notification when the deployment completes.
type DeployNotifier struct {
	mu   sync.Mutex
	subs map[string][]chan struct{}
}

// NewDeployNotifier creates a new deploy notifier.
func NewDeployNotifier() *DeployNotifier {
	return &DeployNotifier{
		subs: make(map[string][]chan struct{}),
	}
}

// Subscribe returns a channel that receives a signal when a process with the given id is deployed.
func (n *DeployNotifier) Subscribe(processId string) chan struct{} {
	ch := make(chan struct{}, 1)
	n.mu.Lock()
	n.subs[processId] = append(n.subs[processId], ch)
	n.mu.Unlock()
	return ch
}

// Unsubscribe removes a previously subscribed channel.
func (n *DeployNotifier) Unsubscribe(processId string, ch chan struct{}) {
	n.mu.Lock()
	defer n.mu.Unlock()
	subs := n.subs[processId]
	for i, s := range subs {
		if s == ch {
			n.subs[processId] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	if len(n.subs[processId]) == 0 {
		delete(n.subs, processId)
	}
}

// Notify sends a non-blocking signal to all subscribers of the given processId.
func (n *DeployNotifier) Notify(processId string) {
	n.mu.Lock()
	subs := make([]chan struct{}, len(n.subs[processId]))
	copy(subs, n.subs[processId])
	n.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
