package bpmn_model

import (
	"fmt"
	"reflect"
	"sync"

	xmlm "github.com/esivres/c2h5oh/pkg/bpmn_model/xml"
)

// Model is a registry of element types. It maps XML qualified names and Go types
// to ModelElementType descriptors.
type Model struct {
	typesByQName  map[string]*ModelElementType
	typesByGoType map[reflect.Type]*ModelElementType
	alternativeNS map[string]string
	name          string
	mu            sync.RWMutex
}

// NewModel creates a new Model with the given name.
func NewModel(name string) *Model {
	return &Model{
		name:          name,
		typesByQName:  make(map[string]*ModelElementType),
		typesByGoType: make(map[reflect.Type]*ModelElementType),
		alternativeNS: make(map[string]string),
	}
}

func qnameKey(nsURI, localName string) string {
	return nsURI + "#" + localName
}

// RegisterType registers a ModelElementType in the model.
func (m *Model) RegisterType(t *ModelElementType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := qnameKey(t.TypeNamespace, t.TypeName)
	m.typesByQName[key] = t
	if t.GoType != nil {
		m.typesByGoType[t.GoType] = t
	}
	t.Model = m
}

// GetTypeByQName looks up a type by namespace URI and local name.
// It also checks alternative namespace mappings.
func (m *Model) GetTypeByQName(nsURI, localName string) *ModelElementType {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := qnameKey(nsURI, localName)
	if t, ok := m.typesByQName[key]; ok {
		return t
	}

	// Try resolving alternative namespace
	if actual, ok := m.alternativeNS[nsURI]; ok {
		key = qnameKey(actual, localName)
		if t, ok := m.typesByQName[key]; ok {
			return t
		}
	}

	return nil
}

// GetTypeByGoType looks up a type by its Go reflect.Type.
func (m *Model) GetTypeByGoType(t reflect.Type) *ModelElementType {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.typesByGoType[t]
}

// RegisterAlternativeNamespace maps an alternative namespace URI to an actual one.
func (m *Model) RegisterAlternativeNamespace(alternative, actual string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alternativeNS[alternative] = actual
}

// ResolveNamespace returns the actual namespace URI, resolving alternatives.
func (m *Model) ResolveNamespace(nsURI string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if actual, ok := m.alternativeNS[nsURI]; ok {
		return actual
	}
	return nsURI
}

// Name returns the model name.
func (m *Model) Name() string {
	return m.name
}

// Types returns all registered types.
func (m *Model) Types() []*ModelElementType {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*ModelElementType, 0, len(m.typesByGoType))
	for _, t := range m.typesByGoType {
		result = append(result, t)
	}
	return result
}

// ModelElementType describes an XML element type in the model.
type ModelElementType struct {
	GoType         reflect.Type
	Model          *Model
	BaseType       *ModelElementType
	Attributes     map[string]*AttributeDescriptor
	Provider       InstanceProvider
	TypeName       string
	TypeNamespace  string
	ExtendingTypes []*ModelElementType
	ChildSpecs     []*ChildSpec
	IsAbstract     bool
}

// InstanceProvider creates a new ModelElementInstance for a given context.
type InstanceProvider func(ctx *TypeInstanceContext) ModelElementInstance

// TypeInstanceContext provides context for creating new element instances.
type TypeInstanceContext struct {
	DomElement    *xmlm.Element
	ModelInstance *ModelInstance
	ElementType   *ModelElementType
}

// IsTypeOf returns true if the given instance is of this type or a subtype.
func (t *ModelElementType) IsTypeOf(instance ModelElementInstance) bool {
	instanceType := instance.GetElementType()
	for current := instanceType; current != nil; current = current.BaseType {
		if current == t {
			return true
		}
	}
	return false
}

// GetAttribute returns the attribute descriptor with the given name,
// searching this type and base types.
func (t *ModelElementType) GetAttribute(name string) *AttributeDescriptor {
	for current := t; current != nil; current = current.BaseType {
		if attr, ok := current.Attributes[name]; ok {
			return attr
		}
	}
	return nil
}

// NewInstance creates a new instance of this type.
func (t *ModelElementType) NewInstance(mi *ModelInstance) (ModelElementInstance, error) {
	if t.IsAbstract {
		return nil, fmt.Errorf("cannot instantiate abstract type %s", t.TypeName)
	}
	if t.Provider == nil {
		return nil, fmt.Errorf("no instance provider for type %s", t.TypeName)
	}

	doc := mi.GetDocument()
	domElem := doc.CreateElement(t.TypeNamespace, t.TypeName)

	ctx := &TypeInstanceContext{
		DomElement:    domElem,
		ModelInstance: mi,
		ElementType:   t,
	}

	instance := t.Provider(ctx)
	return instance, nil
}

// ModelElementInstance is the base interface for all model element instances.
type ModelElementInstance interface {
	GetDomElement() *xmlm.Element
	SetDomElement(*xmlm.Element)
	GetModelInstance() *ModelInstance
	SetModelInstance(*ModelInstance)
	GetElementType() *ModelElementType
	SetElementType(*ModelElementType)
}

// BaseInstance provides a default implementation of ModelElementInstance.
// Concrete element types embed this struct.
type BaseInstance struct {
	DomElement  *xmlm.Element
	ModelInst   *ModelInstance
	ElementType *ModelElementType
}

// NewBaseInstance creates a BaseInstance from a TypeInstanceContext.
func NewBaseInstance(ctx *TypeInstanceContext) BaseInstance {
	return BaseInstance{
		DomElement:  ctx.DomElement,
		ModelInst:   ctx.ModelInstance,
		ElementType: ctx.ElementType,
	}
}

func (b *BaseInstance) GetDomElement() *xmlm.Element       { return b.DomElement }
func (b *BaseInstance) SetDomElement(e *xmlm.Element)      { b.DomElement = e }
func (b *BaseInstance) GetModelInstance() *ModelInstance   { return b.ModelInst }
func (b *BaseInstance) SetModelInstance(mi *ModelInstance) { b.ModelInst = mi }
func (b *BaseInstance) GetElementType() *ModelElementType  { return b.ElementType }
func (b *BaseInstance) SetElementType(t *ModelElementType) { b.ElementType = t }
