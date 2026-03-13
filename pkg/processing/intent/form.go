package intent

// DeployFormIntent requests deployment of a form definition.
type DeployFormIntent struct {
	Header

	// FormId is the form identifier.
	FormId string

	// Content is the raw form JSON.
	Content []byte

	// ContentHash is the hash for deduplication.
	ContentHash []byte
}

func (i *DeployFormIntent) IntentType() Type { return DeployForm }
