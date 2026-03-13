package bpmn_model

import (
	xmlm "github.com/esivres/c2h5oh/pkg/bpmn_model/xml"
)

// ChildSpec describes a child element relationship on a ModelElementType.
type ChildSpec struct {
	ChildTypeName string
	ChildTypeNS   string
	MinOccurs     int
	MaxOccurs     int // -1 for unbounded
	OwnerType     *ModelElementType
}

// ChildElement provides typed access to a single child element (maxOccurs=1).
type ChildElement[T ModelElementInstance] struct {
	Spec *ChildSpec
}

// Get returns the single child element, or the zero value if not found.
func (c *ChildElement[T]) Get(parent ModelElementInstance) (T, bool) {
	mi := parent.GetModelInstance()
	domElem := parent.GetDomElement()

	prefix := ""
	if c.Spec.ChildTypeNS != "" && domElem.GetDocument() != nil {
		prefix = domElem.GetDocument().LookupPrefix(c.Spec.ChildTypeNS)
	}

	for _, child := range domElem.GetChildElements() {
		if child.LocalName() == c.Spec.ChildTypeName && child.Prefix() == prefix {
			if inst := mi.GetElementByDom(child); inst != nil {
				if typed, ok := inst.(T); ok {
					return typed, true
				}
			}
		}
	}
	var zero T
	return zero, false
}

// Set sets the single child element, replacing any existing one.
func (c *ChildElement[T]) Set(parent ModelElementInstance, child T) {
	domParent := parent.GetDomElement()
	mi := parent.GetModelInstance()

	prefix := ""
	if c.Spec.ChildTypeNS != "" && domParent.GetDocument() != nil {
		prefix = domParent.GetDocument().LookupPrefix(c.Spec.ChildTypeNS)
	}

	// Remove existing child of this type
	for _, existingChild := range domParent.GetChildElements() {
		if existingChild.LocalName() == c.Spec.ChildTypeName && existingChild.Prefix() == prefix {
			domParent.RemoveChild(existingChild)
			break
		}
	}

	domParent.AppendChild(child.GetDomElement())
	mi.RegisterElement(child)
}

// Remove removes the child element.
func (c *ChildElement[T]) Remove(parent ModelElementInstance) {
	domParent := parent.GetDomElement()

	prefix := ""
	if c.Spec.ChildTypeNS != "" && domParent.GetDocument() != nil {
		prefix = domParent.GetDocument().LookupPrefix(c.Spec.ChildTypeNS)
	}

	for _, child := range domParent.GetChildElements() {
		if child.LocalName() == c.Spec.ChildTypeName && child.Prefix() == prefix {
			domParent.RemoveChild(child)
			return
		}
	}
}

// ChildElementCollection provides typed access to multiple child elements.
type ChildElementCollection[T ModelElementInstance] struct {
	Spec *ChildSpec
}

// Get returns all child elements of the specified type.
func (c *ChildElementCollection[T]) Get(parent ModelElementInstance) []T {
	mi := parent.GetModelInstance()
	if mi == nil {
		return nil
	}

	var result []T
	domElem := parent.GetDomElement()

	for _, child := range domElem.GetChildElements() {
		inst := mi.GetElementByDom(child)
		if inst == nil {
			continue
		}
		if typed, ok := inst.(T); ok {
			result = append(result, typed)
		}
	}
	return result
}

// GetByNS returns child elements matching the spec's namespace and name.
func (c *ChildElementCollection[T]) GetByNS(parent ModelElementInstance) []T {
	mi := parent.GetModelInstance()
	if mi == nil {
		return nil
	}

	var result []T
	domElem := parent.GetDomElement()

	prefix := ""
	if c.Spec.ChildTypeNS != "" && domElem.GetDocument() != nil {
		prefix = domElem.GetDocument().LookupPrefix(c.Spec.ChildTypeNS)
	}

	for _, child := range domElem.GetChildElements() {
		if child.LocalName() != c.Spec.ChildTypeName || child.Prefix() != prefix {
			continue
		}
		inst := mi.GetElementByDom(child)
		if inst == nil {
			continue
		}
		if typed, ok := inst.(T); ok {
			result = append(result, typed)
		}
	}
	return result
}

// Add adds a child element.
func (c *ChildElementCollection[T]) Add(parent ModelElementInstance, child T) {
	parent.GetDomElement().AppendChild(child.GetDomElement())
	if mi := parent.GetModelInstance(); mi != nil {
		mi.RegisterElement(child)
	}
}

// Remove removes a child element.
func (c *ChildElementCollection[T]) Remove(parent ModelElementInstance, child T) {
	parent.GetDomElement().RemoveChild(child.GetDomElement())
}

// ChildElementCollectionUntyped provides access to child elements matching
// a type hierarchy (any subtype of the declared child type).
type ChildElementCollectionUntyped struct {
	Spec      *ChildSpec
	ChildType *ModelElementType
}

// Get returns all child instances whose element type is a subtype of ChildType.
func (c *ChildElementCollectionUntyped) Get(parent ModelElementInstance) []ModelElementInstance {
	mi := parent.GetModelInstance()
	if mi == nil {
		return nil
	}

	var result []ModelElementInstance
	for _, child := range parent.GetDomElement().GetChildElements() {
		inst := mi.GetElementByDom(child)
		if inst == nil {
			continue
		}
		if c.ChildType.IsTypeOf(inst) {
			result = append(result, inst)
		}
	}
	return result
}

// Add adds a child element.
func (c *ChildElementCollectionUntyped) Add(parent ModelElementInstance, child ModelElementInstance) {
	parent.GetDomElement().AppendChild(child.GetDomElement())
	if mi := parent.GetModelInstance(); mi != nil {
		mi.RegisterElement(child)
	}
}

// findChildDomElement is a helper to find a child DOM element by namespace and name.
func findChildDomElement(parent *xmlm.Element, nsURI, localName string) *xmlm.Element {
	children := parent.GetChildElementsByNS(nsURI, localName)
	if len(children) > 0 {
		return children[0]
	}
	return nil
}
