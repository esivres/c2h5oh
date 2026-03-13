package bpmn_model

// Reference provides ID-based cross-references between elements.
// The source element has an attribute containing the target element's ID.
type Reference[T ModelElementInstance] struct {
	// AttributeName is the name of the attribute holding the reference ID.
	AttributeName string
	// AttributeNamespace is the namespace of the attribute (empty for non-namespaced).
	AttributeNamespace string
}

// GetId returns the referenced element's ID from the source element.
func (r *Reference[T]) GetId(source ModelElementInstance) string {
	domElem := source.GetDomElement()
	if r.AttributeNamespace != "" {
		return domElem.GetAttributeNS(r.AttributeNamespace, r.AttributeName)
	}
	return domElem.GetAttribute(r.AttributeName)
}

// Get resolves the reference, returning the target element.
func (r *Reference[T]) Get(source ModelElementInstance) (T, bool) {
	id := r.GetId(source)
	if id == "" {
		var zero T
		return zero, false
	}

	mi := source.GetModelInstance()
	if mi == nil {
		var zero T
		return zero, false
	}

	elem := mi.GetElementById(id)
	if elem == nil {
		var zero T
		return zero, false
	}

	typed, ok := elem.(T)
	return typed, ok
}

// Set sets the reference to the given target element (by its ID).
func (r *Reference[T]) Set(source ModelElementInstance, target T) {
	targetDom := target.GetDomElement()
	id := targetDom.GetAttribute("id")
	if id == "" {
		return
	}

	domElem := source.GetDomElement()
	if r.AttributeNamespace != "" {
		domElem.SetAttributeNS(r.AttributeNamespace, r.AttributeName, id)
	} else {
		domElem.SetAttribute(r.AttributeName, id)
	}
}

// SetById sets the reference by ID string directly.
func (r *Reference[T]) SetById(source ModelElementInstance, id string) {
	domElem := source.GetDomElement()
	if r.AttributeNamespace != "" {
		domElem.SetAttributeNS(r.AttributeNamespace, r.AttributeName, id)
	} else {
		domElem.SetAttribute(r.AttributeName, id)
	}
}

// Remove removes the reference attribute.
func (r *Reference[T]) Remove(source ModelElementInstance) {
	source.GetDomElement().RemoveAttribute(r.AttributeName)
}

// IdsReference handles space-separated ID references (e.g., "id1 id2 id3").
type IdsReference[T ModelElementInstance] struct {
	AttributeName      string
	AttributeNamespace string
}

// GetIds returns the list of referenced IDs.
func (r *IdsReference[T]) GetIds(source ModelElementInstance) []string {
	raw := source.GetDomElement().GetAttribute(r.AttributeName)
	if raw == "" {
		return nil
	}
	var ids []string
	for _, part := range splitWhitespace(raw) {
		if part != "" {
			ids = append(ids, part)
		}
	}
	return ids
}

// Get resolves all references, returning the target elements.
func (r *IdsReference[T]) Get(source ModelElementInstance) []T {
	ids := r.GetIds(source)
	if len(ids) == 0 {
		return nil
	}

	mi := source.GetModelInstance()
	if mi == nil {
		return nil
	}

	var result []T
	for _, id := range ids {
		elem := mi.GetElementById(id)
		if elem == nil {
			continue
		}
		if typed, ok := elem.(T); ok {
			result = append(result, typed)
		}
	}
	return result
}

func splitWhitespace(s string) []string {
	var result []string
	current := ""
	for _, c := range s {
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
