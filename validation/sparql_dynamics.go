package validation

import (
	"cimgo/cimgostructs"
	"fmt"
	"math"
	"reflect"
	"strings"
)

// ValidateDYProfileSPARQL runs hand-written checks for
// 61970-457_Dynamics-AP-Con-Complex-SHACL and
// 61970-302_Dynamics-AP-Con-Complex-SHACL.
func ValidateDYProfileSPARQL(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckExcitationSystemDynamicsSynchronousMachineDynamics(dataset)...)
	violations = append(violations, CheckSynchronousMachineTimeConstantReactanceModelType(dataset)...)
	violations = append(violations, CheckTurbineGovernorMbaseEquation(dataset)...)
	violations = append(violations, CheckExcitationSystemGains(dataset)...)
	violations = append(violations, CheckPssInputSignals(dataset)...)
	violations = append(violations, CheckGovHydro4GainPoints(dataset)...)
	violations = append(violations, CheckLoadStaticModelAttributes(dataset)...)
	violations = append(violations, CheckRotatingMachineSaturation(dataset)...)
	violations = append(violations, CheckSynchronousMachineSimplifiedAttributes(dataset)...)
	violations = append(violations, CheckDynamicsAssociations(dataset)...)
	return violations
}

// CheckExcitationSystemDynamicsSynchronousMachineDynamics implements dy457:ExcitationSystemDynamics.SynchronousMachineDynamicsSynchronousMachineSimplified-valueType
// Profile: 61970-457_Dynamics-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: ExcitationSystemDynamics.SynchronousMachineDynamics shall not point to a SynchronousMachineSimplified.
func CheckExcitationSystemDynamicsSynchronousMachineDynamics(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements { // Using reflect to check if it's a subtype of ExcitationSystemDynamics
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
					Severity: "sh:Violation",
				})
			}
		}
	}

	return violations
}

// CheckSynchronousMachineTimeConstantReactanceModelType implements dy457:SynchronousMachineTimeConstantReactance-modelType rules
// Profile: 61970-457_Dynamics-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Validates synchronous machine reactance parameters based on the modelKind and rotorKind (Annex A).
func CheckSynchronousMachineTimeConstantReactanceModelType(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, sm := range dataset.SynchronousMachineTimeConstantReactances {
		if sm.ModelType == nil || sm.RotorType == nil {
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
					Severity: "sh:Violation",
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
					Severity: "sh:Violation",
				})
			}
		} else if mt == subtransient && rt == salientPole {
			if sm.SaturationFactorQAxis != 0 || sm.SaturationFactor120QAxis != 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "SynchronousMachineTimeConstantReactance",
					Property: "SynchronousMachineTimeConstantReactance.modelType",
					Message:  "Missing attributes or default values not provided according to 61970-457 Annex A (subtransient/salientPole).",
					Severity: "sh:Violation",
				})
			}
		}
	}

	return violations
}

// CheckTurbineGovernorMbaseEquation implements dyn457:TurbineGovernorDynamics-mbaseEquation
// Profile: 61970-457_Dynamics-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
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
				Severity: "sh:Violation",
			})
		}
	}

	return violations
}

// CheckExcitationSystemGains implements various gain rules for excitation systems
// Profile: 61970-302_Dynamics-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Validates gain and time constant invariants for various excitation systems.
func CheckExcitationSystemGains(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, v := range dataset.ExcAC8Bs {
		if v.Kir == 0 && v.Kpr <= 0 {
			violations = append(violations, Violation{
				ObjectID: id, Class: "ExcAC8B", Property: "ExcAC8B.kpr",
				Message: "The value negative or zero when ExcAC8B.kir = 0.", Severity: "sh:Violation",
			})
		}
	}
	for id, v := range dataset.ExcIEEEAC8Bs {
		if v.Kir == 0 && v.Kpr <= 0 {
			violations = append(violations, Violation{
				ObjectID: id, Class: "ExcIEEEAC8B", Property: "ExcIEEEAC8B.kpr",
				Message: "The value negative or zero when ExcIEEEAC8B.kir = 0.", Severity: "sh:Violation",
			})
		}
	}
	for id, v := range dataset.ExcIEEEAC7Bs {
		if v.Kia == 0 && v.Kpa <= 0 {
			violations = append(violations, Violation{
				ObjectID: id, Class: "ExcIEEEAC7B", Property: "ExcIEEEAC7B.kpa",
				Message: "The value negative or zero when ExcIEEEAC7B.kia = 0.", Severity: "sh:Violation",
			})
		}
		if v.Kir == 0 && v.Kpr <= 0 {
			violations = append(violations, Violation{
				ObjectID: id, Class: "ExcIEEEAC7B", Property: "ExcIEEEAC7B.kpr",
				Message: "The value negative or zero when ExcIEEEAC7B.kir = 0.", Severity: "sh:Violation",
			})
		}
	}
	for id, v := range dataset.ExcBBCs {
		if v.K == 0 {
			violations = append(violations, Violation{
				ObjectID: id, Class: "ExcBBC", Property: "ExcBBC.k",
				Message: "The value is 0.", Severity: "sh:Violation",
			})
		}
	}
	for id, v := range dataset.ExcIEEEDC4Bs {
		if v.Kd > 0 && v.Td <= 0 {
			violations = append(violations, Violation{
				ObjectID: id, Class: "ExcIEEEDC4B", Property: "ExcIEEEDC4B.td",
				Message: "The value negative or zero when ExcIEEEDC4B.kd > 0.", Severity: "sh:Violation",
			})
		}
	}
	return violations
}

// CheckPssInputSignals implements signal uniqueness for PSS
// Profile: 61970-302_Dynamics-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: inputSignal1Type shall be different than inputSignal2Type.
func CheckPssInputSignals(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, v := range dataset.Pss2STs {
		if v.InputSignal1Type != nil && v.InputSignal2Type != nil && v.InputSignal1Type.URI == v.InputSignal2Type.URI {
			violations = append(violations, Violation{
				ObjectID: id, Class: "Pss2ST", Property: "Pss2ST.inputSignal1Type",
				Message: "Input signal #1 and input signal #2 are not different.", Severity: "sh:Violation",
			})
		}
	}
	for id, v := range dataset.PssWECCs {
		if v.InputSignal1Type != nil && v.InputSignal2Type != nil && v.InputSignal1Type.URI == v.InputSignal2Type.URI {
			violations = append(violations, Violation{
				ObjectID: id, Class: "PssWECC", Property: "PssWECC.inputSignal1Type",
				Message: "Input signal #1 and input signal #2 are not different.", Severity: "sh:Violation",
			})
		}
	}
	return violations
}

// CheckGovHydro4GainPoints implements various point sequence rules for GovHydro4
// Profile: 61970-302_Dynamics-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Validates sequential gain points (gv0-gv5, pgv0-pgv5) based on model type.
func CheckGovHydro4GainPoints(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, v := range dataset.GovHydro4s {
		if v.Model == nil {
			continue
		}

		m := v.Model.URI
		simple := "http://iec.ch/TC57/CIM100#GovHydro4ModelKind.simple"
		francisPelton := "http://iec.ch/TC57/CIM100#GovHydro4ModelKind.francisPelton"
		kaplan := "http://iec.ch/TC57/CIM100#GovHydro4ModelKind.kaplan"

		if m == simple {
			checkZero := func(val float64, prop string) {
				if val != 0 {
					violations = append(violations, Violation{
						ObjectID: id, Class: "GovHydro4", Property: "GovHydro4." + prop,
						Message: "The value is not 0 when GovHydro4.model is simple.", Severity: "sh:Violation",
					})
				}
			}
			checkZero(v.Bmax, "bmax")
			checkZero(v.Gv0, "gv0")
			checkZero(v.Gv1, "gv1")
			checkZero(v.Gv2, "gv2")
			checkZero(v.Gv3, "gv3")
			checkZero(v.Gv4, "gv4")
			checkZero(v.Gv5, "gv5")
			checkZero(v.Pgv0, "pgv0")
			checkZero(v.Pgv1, "pgv1")
			checkZero(v.Pgv2, "pgv2")
			checkZero(v.Pgv3, "pgv3")
			checkZero(v.Pgv4, "pgv4")
			checkZero(v.Pgv5, "pgv5")
		} else if m == francisPelton || m == kaplan {
			if m == francisPelton && v.Bmax != 0 {
				violations = append(violations, Violation{
					ObjectID: id, Class: "GovHydro4", Property: "GovHydro4.bmax",
					Message: "The value is not 0 when GovHydro4.model is francisPelton.", Severity: "sh:Violation",
				})
			}
			// Sequence checks
			if v.Gv1 <= v.Gv0 {
				violations = append(violations, Violation{
					ObjectID: id, Class: "GovHydro4", Property: "GovHydro4.gv1",
					Message: "The value is not greater than GovHydro4.gv0 when GovHydro4.model is francisPelton or kaplan.", Severity: "sh:Violation",
				})
			}
			if v.Gv2 <= v.Gv1 {
				violations = append(violations, Violation{
					ObjectID: id, Class: "GovHydro4", Property: "GovHydro4.gv2",
					Message: "The value is not greater than GovHydro4.gv1 when GovHydro4.model is francisPelton or kaplan.", Severity: "sh:Violation",
				})
			}
			if v.Gv3 <= v.Gv2 {
				violations = append(violations, Violation{
					ObjectID: id, Class: "GovHydro4", Property: "GovHydro4.gv3",
					Message: "The value is not greater than GovHydro4.gv2 when GovHydro4.model is francisPelton or kaplan.", Severity: "sh:Violation",
				})
			}
			if v.Gv4 <= v.Gv3 {
				violations = append(violations, Violation{
					ObjectID: id, Class: "GovHydro4", Property: "GovHydro4.gv4",
					Message: "The value is not greater than GovHydro4.gv3 when GovHydro4.model is francisPelton or kaplan.", Severity: "sh:Violation",
				})
			}
			if v.Gv5 <= v.Gv4 || v.Gv5 >= 1.0 {
				violations = append(violations, Violation{
					ObjectID: id, Class: "GovHydro4", Property: "GovHydro4.gv5",
					Message: "The value is either not greater than GovHydro4.gv4 or it is not less than 1 when GovHydro4.model is francisPelton or kaplan.", Severity: "sh:Violation",
				})
			}
		}
	}
	return violations
}

// CheckLoadStaticModelAttributes implements required/prohibited rules for LoadStatic models
// Profile: 61970-302_Dynamics-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Validates required and prohibited attributes for various LoadStatic model types.
func CheckLoadStaticModelAttributes(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, v := range dataset.LoadStatics {
		if v.StaticLoadModelType == nil {
			continue
		}

		m := v.StaticLoadModelType.URI
		constantZ := "http://iec.ch/TC57/CIM100#StaticLoadModelKind.constantZ"
		exponential := "http://iec.ch/TC57/CIM100#StaticLoadModelKind.exponential"
		zIP1 := "http://iec.ch/TC57/CIM100#StaticLoadModelKind.zIP1"
		zIP2 := "http://iec.ch/TC57/CIM100#StaticLoadModelKind.zIP2"

		if m == constantZ {
			if v.Kp1 != 0 || v.Kp2 != 0 || v.Kp3 != 0 || v.Kp4 != 0 || v.Kpf != 0 ||
				v.Kq1 != 0 || v.Kq2 != 0 || v.Kq3 != 0 || v.Kq4 != 0 || v.Kqf != 0 ||
				v.Ep1 != 0 || v.Ep2 != 0 || v.Ep3 != 0 ||
				v.Eq1 != 0 || v.Eq2 != 0 || v.Eq3 != 0 {
				violations = append(violations, Violation{
					ObjectID: id, Class: "LoadStatic", Property: "LoadStatic.staticLoadModelType",
					Message: "The load is represented as a constant impedance but other properties (attributes) are defined.", Severity: "sh:Violation",
				})
			}
		} else if m == exponential {
			// Required: kp1, kp2, kp3, kpf, ep1, ep2, ep3, kq1, kq2, kq3, kqf, eq1, eq2, eq3.
			// Prohibited: kp4, kq4.
			// Note: Check for non-zero as proxy for presence in Go structs.
			if v.Kp4 != 0 || v.Kq4 != 0 {
				violations = append(violations, Violation{
					ObjectID: id, Class: "LoadStatic", Property: "LoadStatic.staticLoadModelType",
					Message: "Unnecessary properties defined for exponential model type (kp4/kq4).", Severity: "sh:Violation",
				})
			}
		} else if m == zIP1 {
			// Required: kp1, kp2, kp3, kpf, kq1, kq2, kq3, kqf.
			// Prohibited: ep1, ep2, ep3, eq1, eq2, eq3, kp4, kq4.
			if v.Ep1 != 0 || v.Ep2 != 0 || v.Ep3 != 0 || v.Eq1 != 0 || v.Eq2 != 0 || v.Eq3 != 0 || v.Kp4 != 0 || v.Kq4 != 0 {
				violations = append(violations, Violation{
					ObjectID: id, Class: "LoadStatic", Property: "LoadStatic.staticLoadModelType",
					Message: "Unnecessary properties defined for zIP1 model type.", Severity: "sh:Violation",
				})
			}
		} else if m == zIP2 {
			// Required: kp1, kp2, kp3, kp4, kpf, kq1, kq2, kq3, kq4, kqf.
			// Prohibited: ep1, ep2, ep3, eq1, eq2, eq3.
			if v.Ep1 != 0 || v.Ep2 != 0 || v.Ep3 != 0 || v.Eq1 != 0 || v.Eq2 != 0 || v.Eq3 != 0 {
				violations = append(violations, Violation{
					ObjectID: id, Class: "LoadStatic", Property: "LoadStatic.staticLoadModelType",
					Message: "Unnecessary properties defined for zIP2 model type.", Severity: "sh:Violation",
				})
			}
		}
	}
	return violations
}

// CheckRotatingMachineSaturation implements saturation constraints for rotating machines
// Profile: 61970-302_Dynamics-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: saturationFactor120 must be >= saturationFactor.
func CheckRotatingMachineSaturation(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		val := reflect.ValueOf(obj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		if val.Kind() != reflect.Struct {
			continue
		}

		s1Field := val.FieldByName("SaturationFactor")
		s2Field := val.FieldByName("SaturationFactor120")
		if s1Field.IsValid() && s2Field.IsValid() {
			s1 := s1Field.Float()
			s2 := s2Field.Float()
			if s2 < s1 {
				violations = append(violations, Violation{
					ObjectID: id, Class: goTypeName(obj), Property: "RotatingMachineDynamics.saturationFactor120",
					Message: "The value is less than RotatingMachineDynamics.saturationFactor.", Severity: "sh:Violation",
				})
			}
		}
	}
	return violations
}

// CheckSynchronousMachineSimplifiedAttributes prohibits saturation for simplified machines
// Profile: 61970-302_Dynamics-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Saturation related attributes are not needed for SynchronousMachineSimplified.
func CheckSynchronousMachineSimplifiedAttributes(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, sms := range dataset.SynchronousMachineSimplifieds {
		if sms.SaturationFactor != 0 || sms.SaturationFactor120 != 0 {
			violations = append(violations, Violation{
				ObjectID: id, Class: "SynchronousMachineSimplified", Property: "rdf:type",
				Message: "Saturation related attributes are not needed for SynchronousMachineSimplified.", Severity: "sh:Violation",
			})
		}
	}
	return violations
}

// CheckDynamicsAssociations ensures governors and loads point to a machine dynamics model
// Profile: 61970-302_Dynamics-AP-Con-Complex
// Origin: Derived from a manual complex constraint (typically SPARQL or class association rule).
// Description: TurbineGovernorDynamics and MechanicalLoadDynamics shall point to either a Synchronous or Asynchronous Machine Dynamics model.
func CheckDynamicsAssociations(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		val := reflect.ValueOf(obj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		if val.Kind() != reflect.Struct {
			continue
		}

		smdField := val.FieldByName("SynchronousMachineDynamics")
		amdField := val.FieldByName("AsynchronousMachineDynamics")

		// If both fields exist but both are nil, then it is a violation
		if smdField.IsValid() && amdField.IsValid() {
			if smdField.IsNil() && amdField.IsNil() {
				// Only for specific target classes
				typeName := goTypeName(obj)
				if strings.HasPrefix(typeName, "Gov") || strings.HasPrefix(typeName, "Mech") || strings.HasSuffix(typeName, "UserDefined") {
					violations = append(violations, Violation{
						ObjectID: id, Class: typeName, Property: "rdf:type",
						Message: "Required association to either SynchronousMachineDynamics or to AsynchronousMachineDynamics is missing.", Severity: "sh:Violation",
					})
				}
			}
		}
	}
	return violations
}
