package validation

import (
	"cimgo/shaclmodel"
	"reflect"
)

// Violation is re-exported from shaclmodel so existing callers of
// validation.Violation keep compiling. The actual type lives in shaclmodel
// (a leaf package) so the generated shaclgen package can return it without
// pulling in validation, which would be a cycle.
type Violation = shaclmodel.Violation

func goTypeName(obj interface{}) string {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
