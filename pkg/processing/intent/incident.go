package intent

// CreateIncidentIntent requests creation of an incident.
type CreateIncidentIntent struct {
	Header

	// ElementInstanceKey is the affected element instance.
	ElementInstanceKey uint64

	// JobKey is the affected job (0 if not job-related).
	JobKey uint64

	// ErrorType categorizes the incident (maps to storage.IncidentType).
	ErrorType string

	// ErrorMessage describes what went wrong.
	ErrorMessage string
}

func (i *CreateIncidentIntent) IntentType() Type { return CreateIncident }

// ResolveIncidentIntent requests resolution of an incident.
type ResolveIncidentIntent struct {
	Header

	// IncidentKey is the key of the incident to resolve.
	IncidentKey uint64
}

func (i *ResolveIncidentIntent) IntentType() Type { return ResolveIncident }
