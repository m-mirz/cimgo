package validation

import (
	"cimgo/cimgostructs"
	"fmt"
	"reflect"
	"strings"
)

// CheckSvTapStepPositionRange implements SvTapStep.position-valueRange (StateVariables SolvedMAS).
// Profile: 61970-301_StateVariables-AP-Con-Complex-SolvedMAS
// Origin: Derived from a SPARQL constraint.
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

// CheckSvShuntCompensatorSectionsInteger implements svs456:SvShuntCompensatorSections.sections-value
// Profile: 61970-456_StateVariables-AP-Con-Complex-SolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: In cases where RegulatingControl.discrete is true and RegulatingControl.enabled is true, SvShuntCompensatorSections.sections shall be integer.
func CheckSvShuntCompensatorSectionsInteger(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, svsc := range dataset.SvShuntCompensatorSectionss {
		if svsc.ShuntCompensator == nil {
			continue
		}
		scID := strings.TrimPrefix(svsc.ShuntCompensator.MRID, "#")
		scObj, ok := dataset.Elements[scID]
		if !ok {
			continue
		}

		// Find RegulatingControl
		var rc *cimgostructs.RegulatingControl
		val := reflect.ValueOf(scObj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		rcField := val.FieldByName("RegulatingControl")
		if rcField.IsValid() && rcField.Kind() == reflect.Ptr && !rcField.IsNil() {
			rcID := strings.TrimPrefix(rcField.Elem().FieldByName("MRID").String(), "#")
			if rcObj, ok := dataset.Elements[rcID]; ok {
				rc, _ = rcObj.(*cimgostructs.RegulatingControl)
			}
		}

		if rc != nil && rc.Enabled && rc.Discrete {
			if svsc.Sections != float64(int(svsc.Sections)) {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "SvShuntCompensatorSections",
					Property: "SvShuntCompensatorSections.sections",
					Message:  fmt.Sprintf("The value (%v) is not integer for an active discrete regulating control.", svsc.Sections),
					Severity: "sh.Violation",
				})
			}
		}
	}

	return violations
}

// CheckSvTapStepPositionInteger implements svs456:SvTapStep.position-value
// Profile: 61970-456_StateVariables-AP-Con-Complex-SolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: In cases where RegulatingControl.discrete is true and RegulatingControl.enabled is true, SvTapStep.position shall be integer.
func CheckSvTapStepPositionInteger(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, svts := range dataset.SvTapSteps {
		if svts.TapChanger == nil {
			continue
		}
		tcID := strings.TrimPrefix(svts.TapChanger.MRID, "#")
		tcObj, ok := dataset.Elements[tcID]
		if !ok {
			continue
		}

		// Find TapChangerControl
		var tcc *cimgostructs.TapChangerControl
		val := reflect.ValueOf(tcObj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		tccField := val.FieldByName("TapChangerControl")
		if tccField.IsValid() && tccField.Kind() == reflect.Ptr && !tccField.IsNil() {
			tccID := strings.TrimPrefix(tccField.Elem().FieldByName("MRID").String(), "#")
			if tccObj, ok := dataset.Elements[tccID]; ok {
				tcc, _ = tccObj.(*cimgostructs.TapChangerControl)
			}
		}

		if tcc != nil && tcc.Enabled && tcc.Discrete {
			if svts.Position != float64(int(svts.Position)) {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "SvTapStep",
					Property: "SvTapStep.position",
					Message:  fmt.Sprintf("The value (%v) is not integer for an active discrete regulating control.", svts.Position),
					Severity: "sh.Violation",
				})
			}
		}
	}

	return violations
}

// CheckSvStateVariablesInstance implements svs456:SvPowerFlow-instance, svs456:SvSwitch-instance, mas600:SvStatus-SV__4, mas600:SvShuntCompensatorSections-SV__4 and mas600:SvTapStep-SV__4
// Profile: 61970-456_StateVariables-AP-Con-Complex-SolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: minimum instantiation requirements for SvPowerFlow and SvSwitch.
func CheckSvStateVariablesInstance(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	// Pre-build indexes
	inServiceMap := make(map[string]bool)
	for _, svs := range dataset.SvStatuss {
		if svs.ConductingEquipment != nil && svs.InService {
			inServiceMap[strings.TrimPrefix(svs.ConductingEquipment.MRID, "#")] = true
		}
	}

	tnInIsland := make(map[string]bool)
	for _, island := range dataset.TopologicalIslands {
		if island.TopologicalNodes != nil {
			tnInIsland[strings.TrimPrefix(island.TopologicalNodes.MRID, "#")] = true
		}
	}

	// 1. Check SvSwitch for ALL switching devices
	for id, obj := range dataset.Elements {
		isSwitch := false
		switch obj.(type) {
		case *cimgostructs.Switch, *cimgostructs.Breaker, *cimgostructs.LoadBreakSwitch,
			*cimgostructs.Disconnector, *cimgostructs.Fuse, *cimgostructs.Jumper,
			*cimgostructs.GroundDisconnector, *cimgostructs.DisconnectingCircuitBreaker,
			*cimgostructs.Cut:
			isSwitch = true
		}
		if !isSwitch { continue }

		found := false
		for _, svsw := range dataset.SvSwitchs {
			if svsw.Switch != nil && strings.TrimPrefix(svsw.Switch.MRID, "#") == id {
				found = true; break
			}
		}
		if !found {
			violations = append(violations, Violation{
				ObjectID: id, Class: goTypeName(obj), Property: "rdf:type",
				Message:  "SvSwitch not instantiated.",
				Severity: "sh.Violation",
			})
		}
	}

	// 2. Check SvPowerFlow for energized equipment
	for id, obj := range dataset.Elements {
		isTarget := false
		switch obj.(type) {
		case *cimgostructs.NonConformLoad, *cimgostructs.EquivalentInjection, *cimgostructs.EnergySource,
			*cimgostructs.ExternalNetworkInjection, *cimgostructs.PowerElectronicsConnection,
			*cimgostructs.AsynchronousMachine, *cimgostructs.EnergyConsumer, *cimgostructs.LinearShuntCompensator,
			*cimgostructs.NonlinearShuntCompensator, *cimgostructs.StaticVarCompensator,
			*cimgostructs.SynchronousMachine, *cimgostructs.StationSupply, *cimgostructs.ConformLoad:
			isTarget = true
		}
		if !isTarget { continue }

		if !inServiceMap[id] { continue }

		energized := false
		for _, term := range dataset.Terminals {
			if term.ConductingEquipment != nil && strings.TrimPrefix(term.ConductingEquipment.MRID, "#") == id {
				if term.TopologicalNode != nil && tnInIsland[strings.TrimPrefix(term.TopologicalNode.MRID, "#")] {
					energized = true; break
				}
			}
		}
		if !energized { continue }

		found := false
		for _, svpf := range dataset.SvPowerFlows {
			if svpf.Terminal != nil {
				tID := strings.TrimPrefix(svpf.Terminal.MRID, "#")
				if t, ok := dataset.Terminals[tID]; ok && t.ConductingEquipment != nil && strings.TrimPrefix(t.ConductingEquipment.MRID, "#") == id {
					found = true; break
				}
			}
		}

		if !found {
			violations = append(violations, Violation{
				ObjectID: id, Class: goTypeName(obj), Property: "rdf:type",
				Message:  "SvPowerFlow is not instantiated for energized equipment.",
				Severity: "sh.Violation",
			})
		}
	}

	// 3. Check SvStatus for ALL energized ConductingEquipment
	for id, obj := range dataset.Elements {
		typeName := goTypeName(obj)
		isCE := strings.HasPrefix(typeName, "Synchronous") || strings.HasPrefix(typeName, "Asynchronous") ||
			strings.HasPrefix(typeName, "Energy") || strings.HasPrefix(typeName, "Line") ||
			strings.HasPrefix(typeName, "Breaker") || strings.HasPrefix(typeName, "Disconnector")
		if !isCE || typeName == "Equipment" { continue }

		energized := false
		for _, term := range dataset.Terminals {
			if term.ConductingEquipment != nil && strings.TrimPrefix(term.ConductingEquipment.MRID, "#") == id {
				if term.TopologicalNode != nil && tnInIsland[strings.TrimPrefix(term.TopologicalNode.MRID, "#")] {
					energized = true; break
				}
			}
		}
		if !energized { continue }

		found := false
		for _, svs := range dataset.SvStatuss {
			if svs.ConductingEquipment != nil && strings.TrimPrefix(svs.ConductingEquipment.MRID, "#") == id {
				found = true; break
			}
		}
		if !found {
			violations = append(violations, Violation{
				ObjectID: id, Class: goTypeName(obj), Property: "rdf:type",
				Message:  "SvStatus is not instantiated for a ConductingEquipment connected to a TopologicalNode which is referenced by a TopologicalIsland.",
				Severity: "sh.Violation",
			})
		}
	}

	// 4. Check SvShuntCompensatorSections for energized shunt compensators
	for id, obj := range dataset.Elements {
		switch obj.(type) {
		case *cimgostructs.LinearShuntCompensator, *cimgostructs.NonlinearShuntCompensator:
			energized := false
			for _, term := range dataset.Terminals {
				if term.ConductingEquipment != nil && strings.TrimPrefix(term.ConductingEquipment.MRID, "#") == id {
					if term.TopologicalNode != nil && tnInIsland[strings.TrimPrefix(term.TopologicalNode.MRID, "#")] {
						energized = true; break
					}
				}
			}
			if !energized { continue }

			found := false
			for _, svsc := range dataset.SvShuntCompensatorSectionss {
				if svsc.ShuntCompensator != nil && strings.TrimPrefix(svsc.ShuntCompensator.MRID, "#") == id {
					found = true; break
				}
			}
			if !found {
				violations = append(violations, Violation{
					ObjectID: id, Class: goTypeName(obj), Property: "rdf:type",
					Message:  "SvShuntCompensatorSections is not instantiated for an energized ShuntCompensator.",
					Severity: "sh.Violation",
				})
			}
		}
	}

	// 5. Check SvTapStep for energized tap changers
	for id, obj := range dataset.Elements {
		switch obj.(type) {
		case *cimgostructs.RatioTapChanger, *cimgostructs.PhaseTapChangerLinear, 
			*cimgostructs.PhaseTapChangerSymmetrical, *cimgostructs.PhaseTapChangerAsymmetrical,
			*cimgostructs.PhaseTapChangerTabular:
			
			energized := false
			var teID string
			val := reflect.ValueOf(obj).Elem()
			teField := val.FieldByName("TransformerEnd")
			if teField.IsValid() && !teField.IsNil() {
				teID = strings.TrimPrefix(teField.Elem().FieldByName("MRID").String(), "#")
			}
			
			if teObj, ok := dataset.Elements[teID]; ok {
				teVal := reflect.ValueOf(teObj).Elem()
				termField := teVal.FieldByName("Terminal")
				if termField.IsValid() && !termField.IsNil() {
					termID := strings.TrimPrefix(termField.Elem().FieldByName("MRID").String(), "#")
					if term, ok := dataset.Terminals[termID]; ok {
						if term.TopologicalNode != nil && tnInIsland[strings.TrimPrefix(term.TopologicalNode.MRID, "#")] {
							energized = true
						}
					}
				}
			}
			
			if !energized { continue }

			found := false
			for _, svts := range dataset.SvTapSteps {
				if svts.TapChanger != nil && strings.TrimPrefix(svts.TapChanger.MRID, "#") == id {
					found = true; break
				}
			}
			if !found {
				violations = append(violations, Violation{
					ObjectID: id, Class: goTypeName(obj), Property: "rdf:type",
					Message:  "SvTapStep is not instantiated for an energized TapChanger.",
					Severity: "sh.Violation",
				})
			}
		}
	}

	return violations
}

// CheckSvShuntCompensatorSectionsSync implements mas600:SvShuntCompensatorSections.sections-SV__4
// Profile: 61970-600-1_AllProfiles-AP-Con-Complex-SolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: SvShuntCompensatorSections.sections shall be the same as ShuntCompensator.sections for non-regulating shunt compensators.
func CheckSvShuntCompensatorSectionsSync(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for _, svsc := range dataset.SvShuntCompensatorSectionss {
		if svsc.ShuntCompensator == nil {
			continue
		}
		scID := strings.TrimPrefix(svsc.ShuntCompensator.MRID, "#")
		scObj, ok := dataset.Elements[scID]
		if !ok {
			continue
		}

		var controlEnabled bool
		var rcEnabled bool = true
		var sections float64

		switch v := scObj.(type) {
		case *cimgostructs.LinearShuntCompensator:
			controlEnabled = v.ControlEnabled
			sections = v.Sections
			if v.RegulatingControl != nil {
				rcID := strings.TrimPrefix(v.RegulatingControl.MRID, "#")
				if rc, ok := dataset.RegulatingControls[rcID]; ok { rcEnabled = rc.Enabled }
			}
		case *cimgostructs.NonlinearShuntCompensator:
			controlEnabled = v.ControlEnabled
			sections = v.Sections
			if v.RegulatingControl != nil {
				rcID := strings.TrimPrefix(v.RegulatingControl.MRID, "#")
				if rc, ok := dataset.RegulatingControls[rcID]; ok { rcEnabled = rc.Enabled }
			}
		default:
			continue
		}

		inService := false
		for _, svs := range dataset.SvStatuss {
			if svs.ConductingEquipment != nil && strings.TrimPrefix(svs.ConductingEquipment.MRID, "#") == scID {
				inService = svs.InService
				break
			}
		}
		if !inService { continue }

		if !controlEnabled || !rcEnabled {
			if svsc.Sections != sections {
				violations = append(violations, Violation{
					ObjectID: scID, Class: goTypeName(scObj), Property: "ShuntCompensator.sections",
					Message: fmt.Sprintf("SvShuntCompensatorSections.sections (%v) is not the same as ShuntCompensator.sections (%v) for non-regulating ShuntCompensator.", svsc.Sections, sections),
					Severity: "sh.Violation",
				})
			}
		}
	}

	return violations
}

// CheckSvTapStepPositionSync implements mas600:SvTapStep.position-SV__4
// Profile: 61970-600-1_AllProfiles-AP-Con-Complex-SolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: SvTapStep.position shall be the same as TapChanger.step for non-regulating tap changers.
func CheckSvTapStepPositionSync(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for _, svts := range dataset.SvTapSteps {
		if svts.TapChanger == nil {
			continue
		}
		tcID := strings.TrimPrefix(svts.TapChanger.MRID, "#")
		tcObj, ok := dataset.Elements[tcID]
		if !ok {
			continue
		}

		var controlEnabled bool
		var rcEnabled bool = true
		var step float64

		val := reflect.ValueOf(tcObj)
		if val.Kind() == reflect.Ptr { val = val.Elem() }
		
		ceField := val.FieldByName("ControlEnabled")
		if ceField.IsValid() { controlEnabled = ceField.Bool() }
		
		stepField := val.FieldByName("Step")
		if stepField.IsValid() { step = stepField.Float() }

		tccField := val.FieldByName("TapChangerControl")
		if tccField.IsValid() && !tccField.IsNil() {
			tccID := strings.TrimPrefix(tccField.Elem().FieldByName("MRID").String(), "#")
			if tcc, ok := dataset.TapChangerControls[tccID]; ok { rcEnabled = tcc.Enabled }
		}

		if !controlEnabled || !rcEnabled {
			if svts.Position != step {
				violations = append(violations, Violation{
					ObjectID: tcID, Class: goTypeName(tcObj), Property: "TapChanger.step",
					Message: fmt.Sprintf("SvTapStep.position (%v) is not the same as TapChanger.step (%v) for non-regulating TapChanger.", svts.Position, step),
					Severity: "sh.Violation",
				})
			}
		}
	}

	return violations
}
