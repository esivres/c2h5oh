package bpmn_model

import (
	"fmt"
	"reflect"

	"github.com/beevik/etree"
	xmlm "github.com/esivres/c2h5oh/pkg/bpmn_model/xml"
)

// ModelInstance wraps an XML Document with a Model, providing typed element access.
type ModelInstance struct {
	model         *Model
	document      *xmlm.Document
	elementsById  map[string]ModelElementInstance
	elementsByDom map[*etree.Element]ModelElementInstance
}

// NewModelInstance creates a new ModelInstance with the given Model and Document.
func NewModelInstance(model *Model, doc *xmlm.Document) *ModelInstance {
	return &ModelInstance{
		model:         model,
		document:      doc,
		elementsById:  make(map[string]ModelElementInstance),
		elementsByDom: make(map[*etree.Element]ModelElementInstance),
	}
}

// GetModel returns the associated Model.
func (mi *ModelInstance) GetModel() *Model {
	return mi.model
}

// GetDocument returns the underlying XML Document.
func (mi *ModelInstance) GetDocument() *xmlm.Document {
	return mi.document
}

// RegisterElement registers an element instance by its ID and DOM element.
func (mi *ModelInstance) RegisterElement(instance ModelElementInstance) {
	domElem := instance.GetDomElement()
	mi.elementsByDom[domElem.Unwrap()] = instance

	id := domElem.GetAttribute("id")
	if id != "" {
		mi.elementsById[id] = instance
	}
}

// RegisterElementById registers (or updates) an element by ID.
func (mi *ModelInstance) RegisterElementById(id string, instance ModelElementInstance) {
	if id != "" {
		mi.elementsById[id] = instance
	}
}

// UnregisterElementById removes an element ID registration.
func (mi *ModelInstance) UnregisterElementById(id string) {
	delete(mi.elementsById, id)
}

// GetElementById returns the element instance with the given ID, or nil.
func (mi *ModelInstance) GetElementById(id string) ModelElementInstance {
	return mi.elementsById[id]
}

// GetElementByDom returns the element instance for a DOM element, or nil.
func (mi *ModelInstance) GetElementByDom(domElem *xmlm.Element) ModelElementInstance {
	return mi.elementsByDom[domElem.Unwrap()]
}

// GetElementsByType returns all registered elements that implement the given Go type.
func (mi *ModelInstance) GetElementsByType(goType reflect.Type) []ModelElementInstance {
	var result []ModelElementInstance
	for _, inst := range mi.elementsByDom {
		if reflect.TypeOf(inst).AssignableTo(goType) {
			result = append(result, inst)
		}
	}
	return result
}

// GetElementsByElementType returns all registered elements of the given ModelElementType.
func (mi *ModelInstance) GetElementsByElementType(t *ModelElementType) []ModelElementInstance {
	var result []ModelElementInstance
	for _, inst := range mi.elementsByDom {
		if t.IsTypeOf(inst) {
			result = append(result, inst)
		}
	}
	return result
}

// NewInstance creates a new element instance of the given type and registers it.
func (mi *ModelInstance) NewInstance(elemType *ModelElementType) (ModelElementInstance, error) {
	instance, err := elemType.NewInstance(mi)
	if err != nil {
		return nil, err
	}
	mi.RegisterElement(instance)
	return instance, nil
}

// GetDocumentElement returns the model instance for the root DOM element.
func (mi *ModelInstance) GetDocumentElement() ModelElementInstance {
	root := mi.document.Root()
	if root == nil {
		return nil
	}
	return mi.elementsByDom[root.Unwrap()]
}

// SetDocumentElement sets the root element of the document.
func (mi *ModelInstance) SetDocumentElement(instance ModelElementInstance) {
	mi.document.SetRoot(instance.GetDomElement())
	mi.RegisterElement(instance)
}

// ResolveElements walks the DOM tree and creates typed instances for all elements.
func (mi *ModelInstance) ResolveElements() error {
	root := mi.document.Root()
	if root == nil {
		return nil
	}
	return mi.resolveElement(root, nil)
}

func (mi *ModelInstance) resolveElement(domElem *xmlm.Element, _ ModelElementInstance) error {
	nsURI := domElem.NamespaceURI()
	localName := domElem.LocalName()

	elemType := mi.model.GetTypeByQName(nsURI, localName)
	if elemType == nil {
		// Unknown type — skip but still process children for known types
		for _, child := range domElem.GetChildElements() {
			if err := mi.resolveElement(child, nil); err != nil {
				return err
			}
		}
		return nil
	}

	if elemType.Provider == nil {
		return fmt.Errorf("no instance provider for type %s:%s", nsURI, localName)
	}

	ctx := &TypeInstanceContext{
		DomElement:    domElem,
		ModelInstance: mi,
		ElementType:   elemType,
	}
	instance := elemType.Provider(ctx)
	mi.RegisterElement(instance)

	// Process children
	for _, child := range domElem.GetChildElements() {
		if err := mi.resolveElement(child, instance); err != nil {
			return err
		}
	}

	return nil
}

// GetTypedElementById is a generic helper to get an element by ID with type assertion.
func GetTypedElementById[T ModelElementInstance](mi *ModelInstance, id string) (T, bool) {
	elem := mi.GetElementById(id)
	if elem == nil {
		var zero T
		return zero, false
	}
	typed, ok := elem.(T)
	return typed, ok
}

// GetTypedElements is a generic helper to get all elements of a specific Go type.
func GetTypedElements[T ModelElementInstance](mi *ModelInstance) []T {
	var result []T
	for _, inst := range mi.elementsByDom {
		if typed, ok := inst.(T); ok {
			result = append(result, typed)
		}
	}
	return result
}
