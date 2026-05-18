package cimbase

import (
	"fmt"
	"reflect"
)

// CIMElement is an interface implemented by all CIM structs.
type CIMElement interface {
	GetId() string
}

// Base is a base struct for all CIM structs.
type Base struct {
	Id string `xml:"ID,attr"`
}

// GetId returns the ID of the CIM element.
func (b *Base) GetId() string {
	return b.Id
}

// DeepMerge recursively merges the fields of two structs.
// It merges the fields from 'new' into 'existing'.
// Both 'existing' and 'new' must be pointers to structs.
func DeepMerge(existing, new reflect.Value) error {

	// We expect pointers to structs
	if existing.Kind() != reflect.Ptr || new.Kind() != reflect.Ptr {
		return fmt.Errorf("both existing and new must be pointers to structs")
	}
	existingElem := existing.Elem()
	newElem := new.Elem()

	if existingElem.Kind() != reflect.Struct || newElem.Kind() != reflect.Struct {
		return fmt.Errorf("both existing and new must be structs")
	}

	// Types match exactly
	if existingElem.Type() == newElem.Type() {
		performLoopMerge(existingElem, newElem)
		return nil
	}

	// Types differ - check for embedding
	indices := FindEmbeddedField(existingElem.Type(), newElem.Type())
	if indices != nil {
		// Find the specific nested field that matches the 'new' type
		targetField := existingElem.FieldByIndex(indices)

		// Now that we have the matching sub-struct, perform the loop merge
		if targetField.Kind() == reflect.Struct {
			performLoopMerge(targetField, newElem)
		}
		return nil
	}

	return fmt.Errorf("could not merge new element into existing: types do not match (%T vs %T)", existing, new)
}

func performLoopMerge(existingElem, newElem reflect.Value) {
	for i := 0; i < newElem.NumField(); i++ {
		newField := newElem.Field(i)
		existingField := existingElem.Field(i)

		if !existingField.CanSet() {
			continue
		}

		// if the new field is not its zero value
		if !reflect.DeepEqual(newField.Interface(), reflect.Zero(newField.Type()).Interface()) {
			// if the field is a pointer to a struct, recurse
			if newField.Kind() == reflect.Ptr && newField.Elem().Kind() == reflect.Struct {
				if existingField.IsNil() {
					// if existing is nil, just set it to the new value
					existingField.Set(newField)
				} else {
					// both are pointers to structs, so we can recurse
					DeepMerge(existingField, newField)
				}
			} else if newField.Kind() == reflect.Struct {
				DeepMerge(existingField.Addr(), newField.Addr())
			} else {
				// for primitive types, slices, maps, and non-pointer structs, we just overwrite
				existingField.Set(newField)
			}
		}
	}
}

func FindEmbeddedField(outerType reflect.Type, targetType reflect.Type) []int {
	// Standardize to non-pointer types for comparison
	if outerType.Kind() == reflect.Ptr {
		outerType = outerType.Elem()
	}
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}

	for i := 0; i < outerType.NumField(); i++ {
		field := outerType.Field(i)

		// Check if this field is the one we want
		if field.Type == targetType {
			return field.Index
		}

		// If it's an embedded struct, recurse into it
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			deepIndex := FindEmbeddedField(field.Type, targetType)
			if deepIndex != nil {
				// Return the full path to the field (field index sequence)
				return append(field.Index, deepIndex...)
			}
		}
	}
	return nil
}
