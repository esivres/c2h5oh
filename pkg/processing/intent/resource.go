package intent

// DeleteResourceIntent requests deletion of a deployed resource (process definition or form).
type DeleteResourceIntent struct {
	Header

	// ResourceKey is the key of the resource to delete.
	ResourceKey uint64
}

func (i *DeleteResourceIntent) IntentType() Type { return DeleteResource }
