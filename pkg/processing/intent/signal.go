package intent

// ThrowSignalIntent broadcasts a signal to all subscribers.
type ThrowSignalIntent struct {
	Header

	// SignalName is the BPMN signal name.
	SignalName string

	// Variables is optional payload (JSON bytes).
	Variables []byte
}

func (i *ThrowSignalIntent) IntentType() Type { return ThrowSignal }
