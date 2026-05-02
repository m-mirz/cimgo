package validation

import (
	"cimgo/cimgostructs"
	"fmt"
	"reflect"
	"strings"
)

// CheckSvTapStepPositionRange implements SvTapStep.position-valueRange (StateVariables SolvedMAS).
// Description: SvTapStep.position must be within [TapChanger.lowStep, TapChanger.highStep].
func CheckSvTapStepPositionRange(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	tapChangerStep := func(id string) (low, high int, ok bool) {
		obj, found := dataset.Elements[id]
		if !found {
			return 0, 0, false
		}
		val := reflect.ValueOf(obj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		lowField := val.FieldByName("LowStep")
		highField := val.FieldByName("HighStep")
		if !lowField.IsValid() || !highField.IsValid() {
			return 0, 0, false
		}
		return int(lowField.Int()), int(highField.Int()), true
	}

	for id, obj := range dataset.Elements {
		sv, ok := obj.(*cimgostructs.SvTapStep)
		if !ok || sv.TapChanger == nil {
			continue
		}
		tcID := strings.TrimPrefix(sv.TapChanger.MRID, "#")
		low, high, ok := tapChangerStep(tcID)
		if !ok {
			continue
		}
		if sv.Position < float64(low) || sv.Position > float64(high) {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "SvTapStep",
				Property: "SvTapStep.position",
				Message:  fmt.Sprintf("The value (%v) is out of range [%d,%d].", sv.Position, low, high),
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}
