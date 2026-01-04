package cimgobase

import (
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
func DeepMerge(existing, new reflect.Value) {
	// We expect pointers to structs
	if existing.Kind() != reflect.Ptr || new.Kind() != reflect.Ptr {
		return
	}
	existingElem := existing.Elem()
	newElem := new.Elem()

	if existingElem.Kind() != reflect.Struct || newElem.Kind() != reflect.Struct {
		return
	}

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
				if existingField.CanAddr() {
					DeepMerge(existingField.Addr(), newField.Addr())
				} else {
					existingField.Set(newField)
				}
			} else {
				// for primitive types, slices, maps, and non-pointer structs, we just overwrite
				existingField.Set(newField)
			}
		}
	}
}
