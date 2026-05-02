package validation

import (
	"cimgo/cimgostructs"
	"reflect"
	"strings"
)

// CheckExcitationSystemDynamicsSynchronousMachineDynamics implements dyn457.ExcitationSystemDynamics.SynchronousMachineDynamicsSynchronousMachineSimplified-valueType
func CheckExcitationSystemDynamicsSynchronousMachineDynamics(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		// Using reflect to check if it's a subtype of ExcitationSystemDynamics
		// but since we only have some structs, let's check by type name prefix or manual list
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

// CheckSynchronousMachineTimeConstantReactanceModelType implements dyn457.SynchronousMachineTimeConstantReactance-modelType rules
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
