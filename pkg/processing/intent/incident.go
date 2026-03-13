package intent

// CreateIncidentIntent requests creation of an incident.
type CreateIncidentIntent struct {
	ErrorType    string
	ErrorMessage string
	Header
	ElementInstanceKey uint64
	JobKey             uint64
}

func (i *CreateIncidentIntent) IntentType() Type { return CreateIncident }

// ResolveIncidentIntent requests resolution of an incident.
type ResolveIncidentIntent struct {
	Header

	// IncidentKey is the key of the incident to resolve.
	IncidentKey uint64
}

func (i *ResolveIncidentIntent) IntentType() Type { return ResolveIncident }
