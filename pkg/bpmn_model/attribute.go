package bpmn_model

import (
	"fmt"
	"strconv"
)

// AttributeDescriptor describes an XML attribute on a ModelElementType.
type AttributeDescriptor struct {
	Name         string
	Namespace    string // empty for non-namespaced attributes
	DefaultValue string
	Required     bool
	IsId         bool
	OwnerType    *ModelElementType
}

// ValueConverter converts between string representations and typed values.
type ValueConverter[T any] interface {
	FromString(s string) (T, error)
	ToString(v T) string
}

// Attribute provides typed read/write access to an XML attribute on element instances.
type Attribute[T any] struct {
	Descriptor *AttributeDescriptor
	Converter  ValueConverter[T]
}

// Get reads the attribute value from the instance's DOM element.
func (a *Attribute[T]) Get(instance ModelElementInstance) T {
	domElem := instance.GetDomElement()
	var raw string
	if a.Descriptor.Namespace != "" {
		raw = domElem.GetAttributeNS(a.Descriptor.Namespace, a.Descriptor.Name)
	} else {
		raw = domElem.GetAttribute(a.Descriptor.Name)
	}

	if raw == "" {
		if a.Descriptor.DefaultValue != "" {
			val, err := a.Converter.FromString(a.Descriptor.DefaultValue)
			if err == nil {
				return val
			}
		}
		var zero T
		return zero
	}

	val, err := a.Converter.FromString(raw)
	if err != nil {
		var zero T
		return zero
	}
	return val
}

// Set writes the attribute value to the instance's DOM element.
func (a *Attribute[T]) Set(instance ModelElementInstance, value T) {
	domElem := instance.GetDomElement()
	strVal := a.Converter.ToString(value)

	if a.Descriptor.Namespace != "" {
		domElem.SetAttributeNS(a.Descriptor.Namespace, a.Descriptor.Name, strVal)
	} else {
		domElem.SetAttribute(a.Descriptor.Name, strVal)
	}

	// Update ID registry
	if a.Descriptor.IsId {
		mi := instance.GetModelInstance()
		if mi != nil {
			mi.RegisterElementById(strVal, instance)
		}
	}
}

// Remove removes the attribute from the instance's DOM element.
func (a *Attribute[T]) Remove(instance ModelElementInstance) {
	instance.GetDomElement().RemoveAttribute(a.Descriptor.Name)
}

// GetRaw returns the raw string value of the attribute.
func (a *Attribute[T]) GetRaw(instance ModelElementInstance) string {
	if a.Descriptor.Namespace != "" {
		return instance.GetDomElement().GetAttributeNS(a.Descriptor.Namespace, a.Descriptor.Name)
	}
	return instance.GetDomElement().GetAttribute(a.Descriptor.Name)
}

// --- Built-in converters ---

// StringConverter converts strings (identity).
type StringConverter struct{}

func (StringConverter) FromString(s string) (string, error) { return s, nil }
func (StringConverter) ToString(v string) string            { return v }

// BoolConverter converts booleans.
type BoolConverter struct{}

func (BoolConverter) FromString(s string) (bool, error) {
	return strconv.ParseBool(s)
}
func (BoolConverter) ToString(v bool) string {
	return strconv.FormatBool(v)
}

// IntConverter converts integers.
type IntConverter struct{}

func (IntConverter) FromString(s string) (int, error) {
	v, err := strconv.Atoi(s)
	return v, err
}
func (IntConverter) ToString(v int) string {
	return strconv.Itoa(v)
}

// Float64Converter converts float64 values.
type Float64Converter struct{}

func (Float64Converter) FromString(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
func (Float64Converter) ToString(v float64) string {
	return fmt.Sprintf("%g", v)
}

// EnumConverter converts string enums (backed by string type).
type EnumConverter struct {
	ValidValues []string
}

func (c EnumConverter) FromString(s string) (string, error) {
	if len(c.ValidValues) > 0 {
		for _, v := range c.ValidValues {
			if v == s {
				return s, nil
			}
		}
		return "", fmt.Errorf("invalid enum value: %s", s)
	}
	return s, nil
}

func (c EnumConverter) ToString(v string) string {
	return v
}
