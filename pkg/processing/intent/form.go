package intent

// DeployFormIntent requests deployment of a form definition.
type DeployFormIntent struct {
	FormId      string
	Content     []byte
	ContentHash []byte
	Header
}

func (i *DeployFormIntent) IntentType() Type { return DeployForm }
