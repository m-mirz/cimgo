package validation

import (
	"cimgo/cimgostructs"
	"fmt"
	"math"
	"reflect"
	"strings"
)

// CheckExcitationSystemDynamicsSynchronousMachineDynamics implements dy457.ExcitationSystemDynamics.SynchronousMachineDynamicsSynchronousMachineSimplified-valueType
func CheckExcitationSystemDynamicsSynchronousMachineDynamics(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		// Using reflect to check if it's a subtype of ExcitationSystemDynamics
		typeName := goTypeName(obj)
		if !strings.HasPrefix(typeName, "Exc") {
			continue
		}

		val := reflect.ValueOf(obj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		smdField := val.FieldByName("SynchronousMachineDynamics")
		if !smdField.IsValid() || smdField.IsNil() {
			continue
		}

		// It's a struct with MRID
		mridField := smdField.Elem().FieldByName("MRID")
		if !mridField.IsValid() {
			continue
		}
		targetID := strings.TrimPrefix(mridField.String(), "#")

		if targetObj, ok := dataset.Elements[targetID]; ok {
			if _, ok := targetObj.(*cimgostructs.SynchronousMachineSimplified); ok {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    typeName,
					Property: "ExcitationSystemDynamics.SynchronousMachineDynamics",
					Message:  "The association ExcitationSystemDynamics.SynchronousMachineDynamics points to an object of type SynchronousMachineSimplified.",
					Severity: "sh.Violation",
				})
			}
		}
	}

	return violations
}

// CheckSynchronousMachineTimeConstantReactanceModelType implements dy457.SynchronousMachineTimeConstantReactance-modelType rules
func CheckSynchronousMachineTimeConstantReactanceModelType(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		sm, ok := obj.(*cimgostructs.SynchronousMachineTimeConstantReactance)
		if !ok || sm.ModelType == nil || sm.RotorType == nil {
			continue
		}

		mt := sm.ModelType.URI
		rt := sm.RotorType.URI
		subtransientSimplified := "http://iec.ch/TC57/CIM100#SynchronousMachineModelKind.subtransientSimplified"
		subtransient := "http://iec.ch/TC57/CIM100#SynchronousMachineModelKind.subtransient"
		roundRotor := "http://iec.ch/TC57/CIM100#RotorKind.roundRotor"
		salientPole := "http://iec.ch/TC57/CIM100#RotorKind.salientPole"

		if mt == subtransientSimplified && rt == roundRotor {
			if sm.StatorResistance != 0 || sm.SaturationFactorQAxis != 0 || sm.SaturationFactor120QAxis != 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "SynchronousMachineTimeConstantReactance",
					Property: "SynchronousMachineTimeConstantReactance.modelType",
					Message:  "Missing attributes or default values not provided according to 61970-457 Annex A (subtransientSimplified/roundRotor).",
					Severity: "sh.Violation",
				})
			}
		} else if mt == subtransient && rt == roundRotor {
			// Check if required fields are present (non-zero)
			if sm.SaturationFactorQAxis == 0 || sm.SaturationFactor120QAxis == 0 || sm.SaturationFactor == 0 || sm.SaturationFactor120 == 0 || sm.XQuadTrans == 0 || sm.Tpqo == 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "SynchronousMachineTimeConstantReactance",
					Property: "SynchronousMachineTimeConstantReactance.modelType",
					Message:  "Missing attributes or default values not provided according to 61970-457 Annex A (subtransient/roundRotor).",
					Severity: "sh.Violation",
				})
			}
		} else if mt == subtransient && rt == salientPole {
			if sm.SaturationFactorQAxis != 0 || sm.SaturationFactor120QAxis != 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "SynchronousMachineTimeConstantReactance",
					Property: "SynchronousMachineTimeConstantReactance.modelType",
					Message:  "Missing attributes or default values not provided according to 61970-457 Annex A (subtransient/salientPole).",
					Severity: "sh.Violation",
				})
			}
		}
	}

	return violations
}

// CheckTurbineGovernorMbaseEquation implements dyn457:TurbineGovernorDynamics-mbaseEquation
// Description: mwbase parameter shall correspond to RotatingMachine.ratedPowerFactor * RotatingMachine.ratedS.
func CheckTurbineGovernorMbaseEquation(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		val := reflect.ValueOf(obj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		if val.Kind() != reflect.Struct {
			continue
		}

		mwbaseField := val.FieldByName("Mwbase")
		if !mwbaseField.IsValid() {
			continue
		}
		mwbase := mwbaseField.Float()
		if mwbase == 0 {
			continue
		}

		smdField := val.FieldByName("SynchronousMachineDynamics")
		if !smdField.IsValid() || smdField.IsNil() {
			continue
		}

		smdID := strings.TrimPrefix(smdField.Elem().FieldByName("MRID").String(), "#")
		smdObj, ok := dataset.Elements[smdID]
		if !ok {
			continue
		}

		// smdObj is likely a SynchronousMachineUserDefined or similar
		// We need to find the actual SynchronousMachine
		smdVal := reflect.ValueOf(smdObj)
		if smdVal.Kind() == reflect.Ptr {
			smdVal = smdVal.Elem()
		}
		smField := smdVal.FieldByName("SynchronousMachine")
		if !smField.IsValid() || smField.IsNil() {
			continue
		}
		smID := strings.TrimPrefix(smField.Elem().FieldByName("MRID").String(), "#")
		sm, ok := dataset.SynchronousMachines[smID]
		if !ok {
			continue
		}

		expected := sm.RatedPowerFactor * sm.RatedS
		epsilon := 0.001
		if math.Abs(mwbase-expected) > epsilon {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    goTypeName(obj),
				Property: "mwbase",
				Message:  fmt.Sprintf("The value %v does not equal RotatingMachine.ratedPowerFactor * RotatingMachine.ratedS (%v).", mwbase, expected),
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}
