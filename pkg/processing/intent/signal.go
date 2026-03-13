package intent

// ThrowSignalIntent broadcasts a signal to all subscribers.
type ThrowSignalIntent struct {
	SignalName string
	Variables  []byte
	Header
}

func (i *ThrowSignalIntent) IntentType() Type { return ThrowSignal }
