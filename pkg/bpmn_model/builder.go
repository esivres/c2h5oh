package bpmn_model

import (
	"reflect"
)

// ModelBuilder provides a fluent API for building a Model with its types.
type ModelBuilder struct {
	model *Model
	types []*TypeBuilder
}

// NewModelBuilder creates a new ModelBuilder with the given model name.
func NewModelBuilder(name string) *ModelBuilder {
	return &ModelBuilder{
		model: NewModel(name),
	}
}

// AlternativeNamespace registers an alternative namespace mapping.
func (mb *ModelBuilder) AlternativeNamespace(alternative, actual string) *ModelBuilder {
	mb.model.RegisterAlternativeNamespace(alternative, actual)
	return mb
}

// DefineType starts defining a new element type.
// goType should be a pointer to the interface or struct type, e.g. (*MyInterface)(nil).
func (mb *ModelBuilder) DefineType(goType interface{}, xmlName string) *TypeBuilder {
	tb := &TypeBuilder{
		modelBuilder: mb,
		typeName:     xmlName,
		goType:       reflect.TypeOf(goType),
		attributes:   make(map[string]*AttributeDescriptor),
	}
	mb.types = append(mb.types, tb)
	return tb
}

// Build finalizes the model, resolving type hierarchies and registering all types.
func (mb *ModelBuilder) Build() *Model {
	// Phase 1: Create ModelElementTypes and register them
	typeMap := make(map[reflect.Type]*ModelElementType)
	for _, tb := range mb.types {
		met := &ModelElementType{
			TypeName:      tb.typeName,
			TypeNamespace: tb.typeNamespace,
			GoType:        tb.goType,
			Attributes:    tb.attributes,
			ChildSpecs:    tb.childSpecs,
			Provider:      tb.provider,
			IsAbstract:    tb.isAbstract,
		}
		typeMap[tb.goType] = met
		mb.model.RegisterType(met)
	}

	// Phase 2: Resolve base types and build extending types lists
	for _, tb := range mb.types {
		if tb.baseGoType != nil {
			met := typeMap[tb.goType]
			baseMet := typeMap[tb.baseGoType]
			if baseMet == nil {
				// Try to find by Go type in model (might be registered externally)
				baseMet = mb.model.GetTypeByGoType(tb.baseGoType)
			}
			if baseMet != nil {
				met.BaseType = baseMet
				baseMet.ExtendingTypes = append(baseMet.ExtendingTypes, met)
			}
		}
	}

	// Set attribute owner types
	for _, tb := range mb.types {
		met := typeMap[tb.goType]
		for _, attr := range met.Attributes {
			attr.OwnerType = met
		}
		for _, cs := range met.ChildSpecs {
			cs.OwnerType = met
		}
	}

	return mb.model
}

// GetModel returns the model being built (for early access during registration).
func (mb *ModelBuilder) GetModel() *Model {
	return mb.model
}

// TypeBuilder provides a fluent API for defining a single element type.
type TypeBuilder struct {
	modelBuilder  *ModelBuilder
	typeName      string
	typeNamespace string
	goType        reflect.Type
	baseGoType    reflect.Type
	attributes    map[string]*AttributeDescriptor
	childSpecs    []*ChildSpec
	provider      InstanceProvider
	isAbstract    bool
}

// Namespace sets the XML namespace URI for the type.
func (tb *TypeBuilder) Namespace(ns string) *TypeBuilder {
	tb.typeNamespace = ns
	return tb
}

// ExtendsType sets the base type.
// baseType should be a pointer to the interface type, e.g. (*BaseInterface)(nil).
func (tb *TypeBuilder) ExtendsType(baseType interface{}) *TypeBuilder {
	tb.baseGoType = reflect.TypeOf(baseType)
	return tb
}

// AbstractType marks the type as abstract (cannot be instantiated directly).
func (tb *TypeBuilder) AbstractType() *TypeBuilder {
	tb.isAbstract = true
	return tb
}

// InstanceProvider sets the factory function for creating instances.
func (tb *TypeBuilder) InstanceProvider(provider InstanceProvider) *TypeBuilder {
	tb.provider = provider
	return tb
}

// StringAttribute starts building a string attribute.
func (tb *TypeBuilder) StringAttribute(name string) *AttributeBuilder[string] {
	return &AttributeBuilder[string]{
		typeBuilder: tb,
		descriptor: &AttributeDescriptor{
			Name: name,
		},
		converter: StringConverter{},
	}
}

// BoolAttribute starts building a boolean attribute.
func (tb *TypeBuilder) BoolAttribute(name string) *AttributeBuilder[bool] {
	return &AttributeBuilder[bool]{
		typeBuilder: tb,
		descriptor: &AttributeDescriptor{
			Name: name,
		},
		converter: BoolConverter{},
	}
}

// IntAttribute starts building an integer attribute.
func (tb *TypeBuilder) IntAttribute(name string) *AttributeBuilder[int] {
	return &AttributeBuilder[int]{
		typeBuilder: tb,
		descriptor: &AttributeDescriptor{
			Name: name,
		},
		converter: IntConverter{},
	}
}

// Float64Attribute starts building a float64 attribute.
func (tb *TypeBuilder) Float64Attribute(name string) *AttributeBuilder[float64] {
	return &AttributeBuilder[float64]{
		typeBuilder: tb,
		descriptor: &AttributeDescriptor{
			Name: name,
		},
		converter: Float64Converter{},
	}
}

// EnumAttribute starts building a string enum attribute with valid values.
func (tb *TypeBuilder) EnumAttribute(name string, validValues ...string) *AttributeBuilder[string] {
	return &AttributeBuilder[string]{
		typeBuilder: tb,
		descriptor: &AttributeDescriptor{
			Name: name,
		},
		converter: EnumConverter{ValidValues: validValues},
	}
}

// AddChildSpec adds a child element specification to the type.
func (tb *TypeBuilder) AddChildSpec(childTypeName, childTypeNS string, minOccurs, maxOccurs int) *ChildSpec {
	spec := &ChildSpec{
		ChildTypeName: childTypeName,
		ChildTypeNS:   childTypeNS,
		MinOccurs:     minOccurs,
		MaxOccurs:     maxOccurs,
	}
	tb.childSpecs = append(tb.childSpecs, spec)
	return spec
}

// Build finalizes the type definition and returns the ModelBuilder for chaining.
func (tb *TypeBuilder) Build() *ModelBuilder {
	return tb.modelBuilder
}

// AttributeBuilder provides a fluent API for defining attributes.
type AttributeBuilder[T any] struct {
	typeBuilder *TypeBuilder
	descriptor  *AttributeDescriptor
	converter   ValueConverter[T]
}

// Required marks the attribute as required.
func (ab *AttributeBuilder[T]) Required() *AttributeBuilder[T] {
	ab.descriptor.Required = true
	return ab
}

// IdAttribute marks the attribute as an ID attribute.
func (ab *AttributeBuilder[T]) IdAttribute() *AttributeBuilder[T] {
	ab.descriptor.IsId = true
	return ab
}

// DefaultValue sets the default value (as string representation).
func (ab *AttributeBuilder[T]) DefaultValue(v T) *AttributeBuilder[T] {
	ab.descriptor.DefaultValue = ab.converter.ToString(v)
	return ab
}

// AttributeNamespace sets the namespace URI for the attribute.
func (ab *AttributeBuilder[T]) AttributeNamespace(ns string) *AttributeBuilder[T] {
	ab.descriptor.Namespace = ns
	return ab
}

// Build finalizes the attribute and returns the typed Attribute accessor.
func (ab *AttributeBuilder[T]) Build() *Attribute[T] {
	ab.typeBuilder.attributes[ab.descriptor.Name] = ab.descriptor
	return &Attribute[T]{
		Descriptor: ab.descriptor,
		Converter:  ab.converter,
	}
}

// BuildTypeBuilder finalizes the attribute and returns the TypeBuilder for chaining.
func (ab *AttributeBuilder[T]) BuildTypeBuilder() *TypeBuilder {
	ab.Build()
	return ab.typeBuilder
}
