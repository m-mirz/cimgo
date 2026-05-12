package validation

import (
	"cimgo/cimgostructs"
	"fmt"
	"reflect"
	"strings"
)

// ValidateOPProfileSPARQL runs hand-written checks for 61970-301_Operation.
func ValidateOPProfileSPARQL(dataset *cimgostructs.CIMElementList) []Violation {
	return CheckMeasurementTerminalRequiredCases(dataset)
}

// CheckMeasurementTerminalRequiredCases implements opn452:Measurement.Terminal-requiredCases
// Profile: 61970-452_Operation-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: Measurement.Terminal must reference a Terminal of the Equipment referenced by
// Measurement.PowerSystemResource, unless measurementType is TapPosition or SwitchPosition.
func CheckMeasurementTerminalRequiredCases(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		var mType string
		var psrRef *struct {
			MRID string `xml:"resource,attr"`
		}
		var termRef *struct {
			MRID string `xml:"resource,attr"`
		}
		var class string

		val := reflect.ValueOf(obj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		if val.Kind() != reflect.Struct {
			continue
		}

		typeField := val.FieldByName("MeasurementType")
		psrField := val.FieldByName("PowerSystemResource")
		termField := val.FieldByName("Terminal")

		if !typeField.IsValid() || !psrField.IsValid() {
			continue
		}

		mType = typeField.String()
		if psrField.Kind() == reflect.Ptr && !psrField.IsNil() {
			psrRef = psrField.Interface().(*struct {
				MRID string `xml:"resource,attr"`
			})
		}
		if termField.IsValid() && termField.Kind() == reflect.Ptr && !termField.IsNil() {
			termRef = termField.Interface().(*struct {
				MRID string `xml:"resource,attr"`
			})
		}
		class = goTypeName(obj)

		if mType == "TapPosition" || mType == "SwitchPosition" {
			if termRef != nil {
				violations = append(violations, Violation{
					ObjectID: id, Class: class, Property: "Terminal",
					Message:  fmt.Sprintf("Measurement.Terminal should not be exchanged for measurementType '%s'.", mType),
					Severity: "sh:Violation",
				})
			}
			continue
		}

		if termRef == nil {
			violations = append(violations, Violation{
				ObjectID: id, Class: class, Property: "Terminal",
				Message:  fmt.Sprintf("Measurement.Terminal is required for measurementType '%s'.", mType),
				Severity: "sh:Violation",
			})
			continue
		}

		if psrRef == nil {
			continue
		}
		psrID := strings.TrimPrefix(psrRef.MRID, "#")
		termID := strings.TrimPrefix(termRef.MRID, "#")

		// Check if termID is a terminal of psrID
		found := false
		for _, t := range dataset.Terminals {
			if t.Id == termID && t.ConductingEquipment != nil && strings.TrimPrefix(t.ConductingEquipment.MRID, "#") == psrID {
				found = true
				break
			}
		}
		if !found {
			violations = append(violations, Violation{
				ObjectID: id, Class: class, Property: "Terminal",
				Message:  fmt.Sprintf("Terminal %s is not a terminal of PowerSystemResource %s.", termID, psrID),
				Severity: "sh:Violation",
			})
		}
	}

	return violations
}
