package validation

import (
	"cimgo/cimstructs"
	"fmt"
	"math"
	"reflect"
	"strings"
)

// ValidateDYProfileSPARQL runs hand-written checks for
// 61970-457_Dynamics-AP-Con-Complex-SHACL and
// 61970-302_Dynamics-AP-Con-Complex-SHACL.
func ValidateDYProfileSPARQL(dataset *cimstructs.CIMDataset) []Violation {
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
	violations = append(violations, CheckGovSteamFV3T5(dataset)...)
	return violations
}

// CheckExcitationSystemDynamicsSynchronousMachineDynamics implements dy457:ExcitationSystemDynamics.SynchronousMachineDynamicsSynchronousMachineSimplified-valueType
// Profile: 61970-457_Dynamics-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: ExcitationSystemDynamics.SynchronousMachineDynamics shall not point to a SynchronousMachineSimplified.
func CheckExcitationSystemDynamicsSynchronousMachineDynamics(dataset *cimstructs.CIMDataset) []Violation {
	var violations []Violation

	for id, obj := range dataset.ByID { // Using reflect to check if it's a subtype of ExcitationSystemDynamics
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

		if targetObj, ok := dataset.ByID[targetID]; ok {
			if _, ok := targetObj.(*cimstructs.SynchronousMachineSimplified); ok {
				violations = append(violations, Violation{
					ObjectID: id,
					RuleID:   "dy457:ExcitationSystemDynamics.SynchronousMachineDynamicsSynchronousMachineSimplified-valueType",
					Name:     "C:457:DY:ExcitationSystemDynamics.SynchronousMachineDynamics:reference",
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
func CheckSynchronousMachineTimeConstantReactanceModelType(dataset *cimstructs.CIMDataset) []Violation {
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
					RuleID:   "dy457:SynchronousMachineTimeConstantReactance-modelType-SubtransientRoundRotorSimplified",
					Name:     "C:457:DY:RotatingMachineDynamics:modelType-SubtransientRoundRotorSimplified",
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
					RuleID:   "dy457:SynchronousMachineTimeConstantReactance-modelType-SubtransientRoundRotor",
					Name:     "C:457:DY:RotatingMachineDynamics:modelType-SubtransientRoundRotor",
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
					RuleID:   "dy457:SynchronousMachineTimeConstantReactance-modelType-SubtransientSalientPole",
					Name:     "C:457:DY:RotatingMachineDynamics:modelType-SubtransientSalientPole",
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
func CheckTurbineGovernorMbaseEquation(dataset *cimstructs.CIMDataset) []Violation {
	var violations []Violation

	for id, obj := range dataset.ByID {
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
		smdObj, ok := dataset.ByID[smdID]
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
				RuleID:   "dyn457:TurbineGovernorDynamics-mbaseEquation",
				Name:     "C:457:DY:mwbase:equation",
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
func CheckExcitationSystemGains(dataset *cimstructs.CIMDataset) []Violation {
	var violations []Violation

	for id, v := range dataset.ExcAC8Bs {
		if v.Kir == 0 && v.Kpr <= 0 {
			violations = append(violations, Violation{
				ObjectID: id, RuleID: "dyu:ExcAC8B.kpr-valueRange", Name: "C:302:DY:ExcAC8B.kpr:valueRange",
				Class: "ExcAC8B", Property: "ExcAC8B.kpr",
				Message: "The value negative or zero when ExcAC8B.kir = 0.", Severity: "sh:Violation",
			})
		}
	}
	for id, v := range dataset.ExcIEEEAC8Bs {
		if v.Kir == 0 && v.Kpr <= 0 {
			violations = append(violations, Violation{
				ObjectID: id, RuleID: "dyu:ExcIEEEAC8B.kpr-valueRange", Name: "C:302:DY:ExcIEEEAC8B.kpr:valueRange",
				Class: "ExcIEEEAC8B", Property: "ExcIEEEAC8B.kpr",
				Message: "The value negative or zero when ExcIEEEAC8B.kir = 0.", Severity: "sh:Violation",
			})
		}
	}
	for id, v := range dataset.ExcIEEEAC7Bs {
		if v.Kia == 0 && v.Kpa <= 0 {
			violations = append(violations, Violation{
				ObjectID: id, RuleID: "dyu:ExcIEEEAC7B.kpa-valueRange", Name: "C:302:DY:ExcIEEEAC7B.kpa:valueRange",
				Class: "ExcIEEEAC7B", Property: "ExcIEEEAC7B.kpa",
				Message: "The value negative or zero when ExcIEEEAC7B.kia = 0.", Severity: "sh:Violation",
			})
		}
		if v.Kir == 0 && v.Kpr <= 0 {
			violations = append(violations, Violation{
				ObjectID: id, RuleID: "dyu:ExcIEEEAC7B.kpr-valueRange", Name: "C:302:DY:ExcIEEEAC7B.kpr:valueRange",
				Class: "ExcIEEEAC7B", Property: "ExcIEEEAC7B.kpr",
				Message: "The value negative or zero when ExcIEEEAC7B.kir = 0.", Severity: "sh:Violation",
			})
		}
	}
	for id, v := range dataset.ExcBBCs {
		if v.K == 0 {
			violations = append(violations, Violation{
				ObjectID: id, RuleID: "dyu:ExcBBC.k-valueRange", Name: "C:302:DY:ExcBBC.k:valueRange",
				Class: "ExcBBC", Property: "ExcBBC.k",
				Message: "The value is 0.", Severity: "sh:Violation",
			})
		}
	}
	for id, v := range dataset.ExcIEEEDC4Bs {
		if v.Kd > 0 && v.Td <= 0 {
			violations = append(violations, Violation{
				ObjectID: id, RuleID: "dyu:ExcIEEEDC4B.td-valueRange", Name: "C:302:DY:ExcIEEEDC4B.td:valueRange",
				Class: "ExcIEEEDC4B", Property: "ExcIEEEDC4B.td",
				Message: "The value negative or zero when ExcIEEEDC4B.kd > 0.", Severity: "sh:Violation",
			})
		}
	}
	for id, v := range dataset.ExcSEXSs {
		if v.Tc > 0 && v.Kc <= 0 {
			violations = append(violations, Violation{
				ObjectID: id, RuleID: "dyu:ExcSEXS.kc-valueRange", Name: "C:302:DY:ExcSEXS.kc:valueRange",
				Class: "ExcSEXS", Property: "ExcSEXS.kc",
				Message: "The value negative or zero when ExcSEXS.tc > 0.", Severity: "sh:Violation",
			})
		}
	}
	return violations
}

// CheckGovSteamFV3T5 implements C:302:DY:GovSteamFV3.t5:valueRange
// Profile: 61970-302_Dynamics-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: t5 (time constant of second boiler pass/reheater) must be >= 0.
func CheckGovSteamFV3T5(dataset *cimstructs.CIMDataset) []Violation {
	var violations []Violation
	for id, v := range dataset.GovSteamFV3s {
		if v.T5 < 0 {
			violations = append(violations, Violation{
				ObjectID: id, RuleID: "dyu:GovSteamFV3.t5-valueRange", Name: "C:302:DY:GovSteamFV3.t5:valueRange",
				Class: "GovSteamFV3", Property: "GovSteamFV3.t5",
				Message: "The value is negative.", Severity: "sh:Violation",
			})
		}
	}
	return violations
}

// CheckPssInputSignals implements signal uniqueness for PSS
// Profile: 61970-302_Dynamics-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: inputSignal1Type shall be different than inputSignal2Type.
func CheckPssInputSignals(dataset *cimstructs.CIMDataset) []Violation {
	var violations []Violation

	for id, v := range dataset.Pss2STs {
		if v.InputSignal1Type != nil && v.InputSignal2Type != nil && v.InputSignal1Type.URI == v.InputSignal2Type.URI {
			violations = append(violations, Violation{
				ObjectID: id, RuleID: "dyu:Pss2ST-inputSignals", Name: "C:302:DY:Pss2ST:inputSignals",
				Class: "Pss2ST", Property: "Pss2ST.inputSignal1Type",
				Message: "Input signal #1 and input signal #2 are not different.", Severity: "sh:Violation",
			})
		}
	}
	for id, v := range dataset.PssWECCs {
		if v.InputSignal1Type != nil && v.InputSignal2Type != nil && v.InputSignal1Type.URI == v.InputSignal2Type.URI {
			violations = append(violations, Violation{
				ObjectID: id, RuleID: "dyu:PssWECC-inputSignals", Name: "C:302:DY:PssWECC:inputSignals",
				Class: "PssWECC", Property: "PssWECC.inputSignal1Type",
				Message: "Input signal #1 and input signal #2 are not different.", Severity: "sh:Violation",
			})
		}
	}
	return violations
}

// CheckGovHydro4GainPoints implements various point sequence rules for GovHydro4
// Profile: 61970-302_Dynamics-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Validates sequential gain points (gv0-gv5, pgv0-pgv5) and the Kaplan blade
// servo points (bgv0-bgv5, which must be 0 for simple and francisPelton models) based on
// model type.
func CheckGovHydro4GainPoints(dataset *cimstructs.CIMDataset) []Violation {
	var violations []Violation

	for id, v := range dataset.GovHydro4s {
		if v.Model == nil {
			continue
		}

		m := v.Model.URI
		simple := "http://iec.ch/TC57/CIM100#GovHydro4ModelKind.simple"
		francisPelton := "http://iec.ch/TC57/CIM100#GovHydro4ModelKind.francisPelton"
		kaplan := "http://iec.ch/TC57/CIM100#GovHydro4ModelKind.kaplan"

		checkZero := func(val float64, prop, ruleID, name string) {
			if val != 0 {
				violations = append(violations, Violation{
					ObjectID: id, RuleID: ruleID, Name: name,
					Class: "GovHydro4", Property: "GovHydro4." + prop,
					Message: "The value is not 0 when GovHydro4.model is simple.", Severity: "sh:Violation",
				})
			}
		}
		if m == simple {
			checkZero(v.Bmax, "bmax", "dyu:GovHydro4.bmax-valueRange", "C:302:DY:GovHydro4.bmax:valueRange")
			checkZero(v.Bgv0, "bgv0", "dyu:GovHydro4.bgv0-valueRange", "C:302:DY:GovHydro4.bgv0:valueRange")
			checkZero(v.Bgv1, "bgv1", "dyu:GovHydro4.bgv1-valueRange", "C:302:DY:GovHydro4.bgv1:valueRange")
			checkZero(v.Bgv2, "bgv2", "dyu:GovHydro4.bgv2-valueRange", "C:302:DY:GovHydro4.bgv2:valueRange")
			checkZero(v.Bgv3, "bgv3", "dyu:GovHydro4.bgv3-valueRange", "C:302:DY:GovHydro4.bgv3:valueRange")
			checkZero(v.Bgv4, "bgv4", "dyu:GovHydro4.bgv4-valueRange", "C:302:DY:GovHydro4.bgv4:valueRange")
			checkZero(v.Bgv5, "bgv5", "dyu:GovHydro4.bgv5-valueRange", "C:302:DY:GovHydro4.bgv5:valueRange")
			checkZero(v.Gv0, "gv0", "dyu:GovHydro4.gv0-valueRange", "C:302:DY:GovHydro4.gv0:valueRange")
			checkZero(v.Gv1, "gv1", "dyu:GovHydro4.gv1-valueRange", "C:302:DY:GovHydro4.gv1:valueRange")
			checkZero(v.Gv2, "gv2", "dyu:GovHydro4.gv2-valueRange", "C:302:DY:GovHydro4.gv2:valueRange")
			checkZero(v.Gv3, "gv3", "dyu:GovHydro4.gv3-valueRange", "C:302:DY:GovHydro4.gv3:valueRange")
			checkZero(v.Gv4, "gv4", "dyu:GovHydro4.gv4-valueRange", "C:302:DY:GovHydro4.gv4:valueRange")
			checkZero(v.Gv5, "gv5", "dyu:GovHydro4.gv5-valueRange", "C:302:DY:GovHydro4.gv5:valueRange")
			checkZero(v.Pgv0, "pgv0", "dyu:GovHydro4.pgv0-valueRange", "C:302:DY:GovHydro4.pgv0:valueRange")
			checkZero(v.Pgv1, "pgv1", "dyu:GovHydro4.pgv1-valueRange", "C:302:DY:GovHydro4.pgv1:valueRange")
			checkZero(v.Pgv2, "pgv2", "dyu:GovHydro4.pgv2-valueRange", "C:302:DY:GovHydro4.pgv2:valueRange")
			checkZero(v.Pgv3, "pgv3", "dyu:GovHydro4.pgv3-valueRange", "C:302:DY:GovHydro4.pgv3:valueRange")
			checkZero(v.Pgv4, "pgv4", "dyu:GovHydro4.pgv4-valueRange", "C:302:DY:GovHydro4.pgv4:valueRange")
			checkZero(v.Pgv5, "pgv5", "dyu:GovHydro4.pgv5-valueRange", "C:302:DY:GovHydro4.pgv5:valueRange")
		} else if m == francisPelton || m == kaplan {
			if m == francisPelton && v.Bmax != 0 {
				violations = append(violations, Violation{
					ObjectID: id, RuleID: "dyu:GovHydro4.bmax-valueRange", Name: "C:302:DY:GovHydro4.bmax:valueRange",
					Class: "GovHydro4", Property: "GovHydro4.bmax",
					Message: "The value is not 0 when GovHydro4.model is francisPelton.", Severity: "sh:Violation",
				})
			}
			if m == francisPelton {
				checkZeroFP := func(val float64, prop, ruleID, name string) {
					if val != 0 {
						violations = append(violations, Violation{
							ObjectID: id, RuleID: ruleID, Name: name,
							Class: "GovHydro4", Property: "GovHydro4." + prop,
							Message: "The value is not 0 when GovHydro4.model is francisPelton.", Severity: "sh:Violation",
						})
					}
				}
				checkZeroFP(v.Bgv0, "bgv0", "dyu:GovHydro4.bgv0-valueRange", "C:302:DY:GovHydro4.bgv0:valueRange")
				checkZeroFP(v.Bgv1, "bgv1", "dyu:GovHydro4.bgv1-valueRange", "C:302:DY:GovHydro4.bgv1:valueRange")
				checkZeroFP(v.Bgv2, "bgv2", "dyu:GovHydro4.bgv2-valueRange", "C:302:DY:GovHydro4.bgv2:valueRange")
				checkZeroFP(v.Bgv3, "bgv3", "dyu:GovHydro4.bgv3-valueRange", "C:302:DY:GovHydro4.bgv3:valueRange")
				checkZeroFP(v.Bgv4, "bgv4", "dyu:GovHydro4.bgv4-valueRange", "C:302:DY:GovHydro4.bgv4:valueRange")
				checkZeroFP(v.Bgv5, "bgv5", "dyu:GovHydro4.bgv5-valueRange", "C:302:DY:GovHydro4.bgv5:valueRange")
			}
			// Sequence checks
			if v.Gv1 <= v.Gv0 {
				violations = append(violations, Violation{
					ObjectID: id, RuleID: "dyu:GovHydro4.gv1-valueRange", Name: "C:302:DY:GovHydro4.gv1:valueRange",
					Class: "GovHydro4", Property: "GovHydro4.gv1",
					Message: "The value is not greater than GovHydro4.gv0 when GovHydro4.model is francisPelton or kaplan.", Severity: "sh:Violation",
				})
			}
			if v.Gv2 <= v.Gv1 {
				violations = append(violations, Violation{
					ObjectID: id, RuleID: "dyu:GovHydro4.gv2-valueRange", Name: "C:302:DY:GovHydro4.gv2:valueRange",
					Class: "GovHydro4", Property: "GovHydro4.gv2",
					Message: "The value is not greater than GovHydro4.gv1 when GovHydro4.model is francisPelton or kaplan.", Severity: "sh:Violation",
				})
			}
			if v.Gv3 <= v.Gv2 {
				violations = append(violations, Violation{
					ObjectID: id, RuleID: "dyu:GovHydro4.gv3-valueRange", Name: "C:302:DY:GovHydro4.gv3:valueRange",
					Class: "GovHydro4", Property: "GovHydro4.gv3",
					Message: "The value is not greater than GovHydro4.gv2 when GovHydro4.model is francisPelton or kaplan.", Severity: "sh:Violation",
				})
			}
			if v.Gv4 <= v.Gv3 {
				violations = append(violations, Violation{
					ObjectID: id, RuleID: "dyu:GovHydro4.gv4-valueRange", Name: "C:302:DY:GovHydro4.gv4:valueRange",
					Class: "GovHydro4", Property: "GovHydro4.gv4",
					Message: "The value is not greater than GovHydro4.gv3 when GovHydro4.model is francisPelton or kaplan.", Severity: "sh:Violation",
				})
			}
			if v.Gv5 <= v.Gv4 || v.Gv5 >= 1.0 {
				violations = append(violations, Violation{
					ObjectID: id, RuleID: "dyu:GovHydro4.gv5-valueRange", Name: "C:302:DY:GovHydro4.gv5:valueRange",
					Class: "GovHydro4", Property: "GovHydro4.gv5",
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
func CheckLoadStaticModelAttributes(dataset *cimstructs.CIMDataset) []Violation {
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
					ObjectID: id, RuleID: "dyu:LoadStatic.staticLoadModelType-constantZ", Name: "C:302:DY:StaticLoadModelKind.constantZ:requiredAttributes",
					Class: "LoadStatic", Property: "LoadStatic.staticLoadModelType",
					Message: "The load is represented as a constant impedance but other properties (attributes) are defined.", Severity: "sh:Violation",
				})
			}
		} else if m == exponential {
			// Required: kp1, kp2, kp3, kpf, ep1, ep2, ep3, kq1, kq2, kq3, kqf, eq1, eq2, eq3.
			// Prohibited: kp4, kq4.
			// Note: Check for non-zero as proxy for presence in Go structs.
			if v.Kp4 != 0 || v.Kq4 != 0 {
				violations = append(violations, Violation{
					ObjectID: id, RuleID: "dyu:LoadStatic.staticLoadModelType-exponental", Name: "C:302:DY:StaticLoadModelKind.exponential:requiredAttributes",
					Class: "LoadStatic", Property: "LoadStatic.staticLoadModelType",
					Message: "Unnecessary properties defined for exponential model type (kp4/kq4).", Severity: "sh:Violation",
				})
			}
		} else if m == zIP1 {
			// Required: kp1, kp2, kp3, kpf, kq1, kq2, kq3, kqf.
			// Prohibited: ep1, ep2, ep3, eq1, eq2, eq3, kp4, kq4.
			if v.Ep1 != 0 || v.Ep2 != 0 || v.Ep3 != 0 || v.Eq1 != 0 || v.Eq2 != 0 || v.Eq3 != 0 || v.Kp4 != 0 || v.Kq4 != 0 {
				violations = append(violations, Violation{
					ObjectID: id, RuleID: "dyu:LoadStatic.staticLoadModelType-zIP1", Name: "C:302:DY:StaticLoadModelKind.zIP1:requiredAttributes",
					Class: "LoadStatic", Property: "LoadStatic.staticLoadModelType",
					Message: "Unnecessary properties defined for zIP1 model type.", Severity: "sh:Violation",
				})
			}
		} else if m == zIP2 {
			// Required: kp1, kp2, kp3, kp4, kpf, kq1, kq2, kq3, kq4, kqf.
			// Prohibited: ep1, ep2, ep3, eq1, eq2, eq3.
			if v.Ep1 != 0 || v.Ep2 != 0 || v.Ep3 != 0 || v.Eq1 != 0 || v.Eq2 != 0 || v.Eq3 != 0 {
				violations = append(violations, Violation{
					ObjectID: id, RuleID: "dyu:LoadStatic.staticLoadModelType-zIP2", Name: "C:302:DY:StaticLoadModelKind.zIP2:requiredAttributes",
					Class: "LoadStatic", Property: "LoadStatic.staticLoadModelType",
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
func CheckRotatingMachineSaturation(dataset *cimstructs.CIMDataset) []Violation {
	var violations []Violation

	for id, obj := range dataset.ByID {
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
					ObjectID: id, RuleID: "dyu:RotatingMachineDynamics.saturationFactor120-valueRange", Name: "C:302:DY:RotatingMachineDynamics.saturationFactor120:valueRange",
					Class: goTypeName(obj), Property: "RotatingMachineDynamics.saturationFactor120",
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
func CheckSynchronousMachineSimplifiedAttributes(dataset *cimstructs.CIMDataset) []Violation {
	var violations []Violation

	for id, sms := range dataset.SynchronousMachineSimplifieds {
		if sms.SaturationFactor != 0 || sms.SaturationFactor120 != 0 {
			violations = append(violations, Violation{
				ObjectID: id, RuleID: "dyu:SynchronousMachineSimplified-requiredAttributes", Name: "C:302:DY:SynchronousMachineSimplified:requiredAttributes",
				Class: "SynchronousMachineSimplified", Property: "rdf:type",
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
func CheckDynamicsAssociations(dataset *cimstructs.CIMDataset) []Violation {
	var violations []Violation

	for id, obj := range dataset.ByID {
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
				typeName := goTypeName(obj)
				if strings.HasPrefix(typeName, "Gov") {
					violations = append(violations, Violation{
						ObjectID: id, RuleID: "dyu:TurbineGovernorDynamics", Name: "C:302:DY:TurbineGovernorDynamics:associationsCondition",
						Class: typeName, Property: "rdf:type",
						Message: "Required association to either SynchronousMachineDynamics or to AsynchronousMachineDynamics is missing.", Severity: "sh:Violation",
					})
				} else if strings.HasPrefix(typeName, "Mech") {
					violations = append(violations, Violation{
						ObjectID: id, RuleID: "dyu:MechanicalLoadDynamics", Name: "C:302:DY:MechanicalLoadDynamics:associationsCondition",
						Class: typeName, Property: "rdf:type",
						Message: "Required association to either SynchronousMachineDynamics or to AsynchronousMachineDynamics is missing.", Severity: "sh:Violation",
					})
				}
			}
		}
	}
	return violations
}
