package cimproto

import (
	"cimgo/cimstructs"
	apiv1 "cimgo/proto/api/v1"
	"fmt"
	"reflect"
	"strings"
)

// ToProto converts a cimstructs.CIMElementList to its Protobuf equivalent apiv1.CIMElementList.
func ToProto(cimData *cimstructs.CIMElementList) (*apiv1.CIMElementList, error) {
	protoList := &apiv1.CIMElementList{}
	for id, elem := range cimData.Elements {
		err := AddElementToProto(protoList, elem)
		if err != nil {
			// We log but continue, as some elements might not be in the proto definition yet
			fmt.Printf("Warning: could not add element %s of type %T to proto list: %v\n", id, elem, err)
		}
	}
	return protoList, nil
}

// AddElementToProto dynamically maps a CIM element from the internal struct representation
// to the corresponding Protobuf message type and adds it to the CIMElementList.
func AddElementToProto(protoList *apiv1.CIMElementList, elem interface{}) error {
	elemVal := reflect.ValueOf(elem)
	if elemVal.Kind() == reflect.Ptr {
		elemVal = elemVal.Elem()
	}
	typeName := elemVal.Type().Name()

	// Find the field in CIMElementList that corresponds to this type
	listVal := reflect.ValueOf(protoList).Elem()
	field := listVal.FieldByName(typeName)
	if !field.IsValid() {
		return fmt.Errorf("no field for type %s in CIMElementList", typeName)
	}

	// Create the corresponding apiv1 struct
	sliceType := field.Type()
	ptrToStructType := sliceType.Elem() // e.g., *apiv1.ACLineSegment
	structType := ptrToStructType.Elem()

	newProtoElem := reflect.New(structType)
	MapFields(elem, newProtoElem.Interface())

	// Add to the slice
	field.Set(reflect.Append(field, newProtoElem))

	return nil
}

// MapFields copies fields from a cimgostructs struct to an apiv1 struct using reflection.
// It handles embedded fields by flattening them and maps IDs to MRID.
func MapFields(src, dst interface{}) {
	srcVal := reflect.ValueOf(src)
	if srcVal.Kind() == reflect.Ptr {
		srcVal = srcVal.Elem()
	}
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() == reflect.Ptr {
		dstVal = dstVal.Elem()
	}

	// Map primitive fields and handle "Super" for inheritance
	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		srcFieldType := srcVal.Type().Field(i)

		if srcFieldType.Anonymous {
			// Embedded field - map its fields to the same dst or to dst.Super
			superField := dstVal.FieldByName("Super")
			if superField.IsValid() && superField.Kind() == reflect.Ptr {
				if superField.IsNil() {
					superField.Set(reflect.New(superField.Type().Elem()))
				}
				MapFields(srcField.Interface(), superField.Interface())
			} else {
				// No Super field, try to map directly to dst
				MapFields(srcField.Interface(), dst)
			}
			continue
		}

		// Regular field
		fieldName := srcFieldType.Name
		// Protobuf generated fields might have different casing or suffixes
		dstField := dstVal.FieldByName(fieldName)
		if !dstField.IsValid() {
			// Try case-insensitive
			for j := 0; j < dstVal.NumField(); j++ {
				if strings.EqualFold(dstVal.Type().Field(j).Name, fieldName) {
					dstField = dstVal.Field(j)
					break
				}
			}
		}

		if dstField.IsValid() && dstField.CanSet() && dstField.Type() == srcField.Type() {
			dstField.Set(srcField)
		}
	}

	// Special case for ID -> mRID
	idField := srcVal.FieldByName("Id")
	if !idField.IsValid() {
		// Try through embedded Base
		if baseField := srcVal.FieldByName("Base"); baseField.IsValid() {
			idField = baseField.FieldByName("Id")
		}
	}

	if idField.IsValid() {
		mridField := dstVal.FieldByName("MRID")
		if !mridField.IsValid() {
			// Try to find it in Super chain
			curr := dstVal
			for {
				f := curr.FieldByName("MRID")
				if f.IsValid() {
					mridField = f
					break
				}
				super := curr.FieldByName("Super")
				if !super.IsValid() || super.IsNil() {
					break
				}
				curr = super.Elem()
			}
		}
		if mridField.IsValid() && mridField.CanSet() && mridField.Kind() == reflect.String {
			mridField.Set(idField)
		}
	}
}
