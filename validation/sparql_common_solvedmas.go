package validation

import (
	"cimgo/cimgostructs"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// ValidateCommonRulesSolvedMASSPARQL runs hand-written checks for common rules that require solved MAS, i.e. 61970-301-1_AllProfiles-AP-Con-Complex-SolvedMAS-SHACL.
func ValidateCommonRulesSolvedMASSPARQL(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	// Profile: 61970-456_AllProfiles-AP-Con-Complex-SolvedMAS-SHACL
	violations = append(violations, CheckAngleReference(dataset)...)
	// Profile: 61970-600-1_AllProfiles-AP-Con-Complex-SolvedMAS-SHACL
	violations = append(violations, CheckDanglingReferences(dataset)...)
	violations = append(violations, CheckSvTapStepPositionSync(dataset)...)
	violations = append(violations, CheckSvShuntCompensatorSectionsSync(dataset)...)
	violations = append(violations, CheckStateVariablesInstantiated(dataset)...)
	// Profile: 61970-600-1_AllProfiles-AP-Con-Complex-SolvedMAS-SHACL (mas600)
	violations = append(violations, CheckSvStatusInstance(dataset)...)
	violations = append(violations, CheckSvShuntCompensatorSectionsInstance(dataset)...)
	violations = append(violations, CheckSvTapStepInstance(dataset)...)
	// Profile: 61970-600-2_AllProfiles-AP-Con-Complex-SolvedMAS-SHACL
	violations = append(violations, CheckRegulatingControlContradictory(dataset)...)
	violations = append(violations, CheckRegulatingControlSameIsland(dataset)...)
	return violations
}

// CheckAngleReference implements sm456:Model-angleReference
// Profile: 61970-456_AllProfiles-AP-Con-Complex-SolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: The angle reference slack should be the SynchronousMachine connected to the
// TopologicalNode referenced by TopologicalIsland.AngleRefTopologicalNode.
func CheckAngleReference(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	// Find the AngleRefTopologicalNode from all islands
	angleRefTNs := make(map[string]string) // TN ID -> Island ID
	for id, island := range dataset.TopologicalIslands {
		if island.AngleRefTopologicalNode != nil {
			tnID := strings.TrimPrefix(island.AngleRefTopologicalNode.MRID, "#")
			angleRefTNs[tnID] = id
		}
	}

	// Find SynchronousMachines with highest referencePriority (usually 1)
	var highestPrioritySMs []*cimgostructs.SynchronousMachine
	minPriority := 9999
	for _, sm := range dataset.SynchronousMachines {
		if sm.ReferencePriority > 0 && sm.ReferencePriority < minPriority {
			minPriority = sm.ReferencePriority
			highestPrioritySMs = []*cimgostructs.SynchronousMachine{sm}
		} else if sm.ReferencePriority > 0 && sm.ReferencePriority == minPriority {
			highestPrioritySMs = append(highestPrioritySMs, sm)
		}
	}

	if len(highestPrioritySMs) == 0 {
		return violations // No priority defined
	}

	if len(highestPrioritySMs) > 1 {
		violations = append(violations, Violation{
			ObjectID: "global",
			Class:    "SynchronousMachine",
			Property: "referencePriority",
			Message:  "Multiple machines with highest SynchronousMachine.referencePriority found.",
			Severity: "sh.Violation",
		})
	}

	// Check if any of the highest priority SMs are at an angle reference node
	for _, sm := range highestPrioritySMs {
		// Find Terminal for this SM
		foundAtRefNode := false
		for _, term := range dataset.Terminals {
			if term.ConductingEquipment != nil && strings.TrimPrefix(term.ConductingEquipment.MRID, "#") == sm.Id {
				if term.TopologicalNode != nil {
					tnID := strings.TrimPrefix(term.TopologicalNode.MRID, "#")
					if _, isRef := angleRefTNs[tnID]; isRef {
						foundAtRefNode = true
						break
					}
				}
			}
		}

		if !foundAtRefNode {
			violations = append(violations, Violation{
				ObjectID: sm.Id,
				Class:    "SynchronousMachine",
				Property: "referencePriority",
				Message:  "The SynchronousMachine with highest priority is not connected to a TopologicalIsland.AngleRefTopologicalNode.",
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckDanglingReferences implements sm600:All-DanglingReferences
// Profile: 61970-600-1_AllProfiles-AP-Con-Complex-SolvedMAS-SHACL
// Origin: Derived from a SPARQL constraint.
// Description: All references in the instance files pointing to other instance files should be satisfied.
func CheckDanglingReferences(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	for id, obj := range dataset.Elements {
		val := reflect.ValueOf(obj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		if val.Kind() != reflect.Struct {
			continue
		}

		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			if field.Kind() == reflect.Ptr && !field.IsNil() {
				// Check for association struct { MRID string }
				mridField := field.Elem().FieldByName("MRID")
				if mridField.IsValid() && mridField.Kind() == reflect.String {
					targetID := strings.TrimPrefix(mridField.String(), "#")
					// Skip if it's an external URI or empty
					if targetID != "" && !strings.Contains(targetID, "://") && !strings.HasPrefix(targetID, "http") {
						if _, ok := dataset.Elements[targetID]; !ok {
							violations = append(violations, Violation{
								ObjectID: id,
								Class:    goTypeName(obj),
								Property: val.Type().Field(i).Name,
								Message:  fmt.Sprintf("Dangling reference to '%s'.", targetID),
								Severity: "sh.Violation",
							})
						}
					}
				}
			}
		}
	}
	return violations
}

// CheckStateVariablesInstantiated implements sm600:SvVoltage-SV__4, SvSwitch-SV__4, SvStatus-SV__4, etc.
// Profile: 61970-600-1_AllProfiles-AP-Con-Complex-SHACL
// Origin: Derived from a SPARQL constraint.
// Description: All state variables shall be instantiated for all energized elements part of a TopologicalIsland.
func CheckStateVariablesInstantiated(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	// Map of TN IDs that are part of an island
	tnToIsland := make(map[string]string)
	for id, island := range dataset.TopologicalIslands {
		if island.TopologicalNodes != nil {
			tnID := strings.TrimPrefix(island.TopologicalNodes.MRID, "#")
			tnToIsland[tnID] = id
		}
	}

	// 1. Check SvVoltage for all energized TNs
	for tnID, islandID := range tnToIsland {
		found := false
		for _, svv := range dataset.SvVoltages {
			if svv.TopologicalNode != nil && strings.TrimPrefix(svv.TopologicalNode.MRID, "#") == tnID {
				found = true
				break
			}
		}
		if !found {
			violations = append(violations, Violation{
				ObjectID: tnID, Class: "TopologicalNode", Property: "rdf:type",
				Message:  fmt.Sprintf("SvVoltage is not instantiated for energized TopologicalNode part of island %s.", islandID),
				Severity: "sh.Violation",
			})
		}
	}

	// 2. Check SvSwitch for all energized retained switches
	for id, sw := range dataset.Switchs {
		if !sw.Retained || !sw.InService {
			continue
		}
		// Check if it's energized (at least one terminal connected to an island)
		energized := false
		for _, term := range dataset.Terminals {
			if term.ConductingEquipment != nil && strings.TrimPrefix(term.ConductingEquipment.MRID, "#") == id {
				if term.TopologicalNode != nil {
					tnID := strings.TrimPrefix(term.TopologicalNode.MRID, "#")
					if _, ok := tnToIsland[tnID]; ok {
						energized = true
						break
					}
				}
			}
		}
		if !energized {
			continue
		}

		found := false
		for _, svsw := range dataset.SvSwitchs {
			if svsw.Switch != nil && strings.TrimPrefix(svsw.Switch.MRID, "#") == id {
				found = true
				break
			}
		}
		if !found {
			violations = append(violations, Violation{
				ObjectID: id, Class: "Switch", Property: "rdf:type",
				Message:  "SvSwitch not instantiated for energized retained Switch.",
				Severity: "sh.Violation",
			})
		}
	}

	// 3. Check SvStatus for all energized ConductingEquipment
	for id, obj := range dataset.Elements {
		energized := false
		for _, term := range dataset.Terminals {
			if term.ConductingEquipment != nil && strings.TrimPrefix(term.ConductingEquipment.MRID, "#") == id {
				if term.TopologicalNode != nil {
					tnID := strings.TrimPrefix(term.TopologicalNode.MRID, "#")
					if _, ok := tnToIsland[tnID]; ok {
						energized = true
						break
					}
				}
			}
		}
		if !energized {
			continue
		}

		found := false
		for _, svs := range dataset.SvStatuss {
			if svs.ConductingEquipment != nil && strings.TrimPrefix(svs.ConductingEquipment.MRID, "#") == id {
				found = true
				break
			}
		}
		if !found {
			violations = append(violations, Violation{
				ObjectID: id, Class: goTypeName(obj), Property: "rdf:type",
				Message:  "SvStatus is not instantiated for energized ConductingEquipment.",
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckRegulatingControlContradictory implements sm6002:RegulatingControl-samePoint
// Profile: 61970-600-2_AllProfiles-AP-Con-Complex-SolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: If multiple RegulatingControls control same point, targetValues must not be contradictory.
func CheckRegulatingControlContradictory(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	type regPoint struct {
		termID string
		mode   string
	}
	targets := make(map[regPoint][]string)

	for id, rc := range dataset.RegulatingControls {
		if !rc.Enabled || rc.Terminal == nil || rc.Mode == nil {
			continue
		}
		key := regPoint{
			termID: strings.TrimPrefix(rc.Terminal.MRID, "#"),
			mode:   rc.Mode.URI,
		}
		targets[key] = append(targets[key], id)
	}

	for _, ids := range targets {
		if len(ids) < 2 {
			continue
		}
		sort.Strings(ids)

		val0 := dataset.RegulatingControls[ids[0]].TargetValue
		for i := 1; i < len(ids); i++ {
			if dataset.RegulatingControls[ids[i]].TargetValue != val0 {
				violations = append(violations, Violation{
					ObjectID: ids[i],
					Class:    "RegulatingControl",
					Property: "RegulatingControl.targetValue",
					Message:  fmt.Sprintf("Enabled RegulatingControl-s of the same type associated with the same TopologicalNode have different target values. RegulatingControl ID: %s.", ids[i]),
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
		rcEnabled := true
		var sections float64

		switch v := scObj.(type) {
		case *cimgostructs.LinearShuntCompensator:
			controlEnabled = v.ControlEnabled
			sections = v.Sections
			if v.RegulatingControl != nil {
				rcID := strings.TrimPrefix(v.RegulatingControl.MRID, "#")
				if rc, ok := dataset.RegulatingControls[rcID]; ok {
					rcEnabled = rc.Enabled
				}
			}
		case *cimgostructs.NonlinearShuntCompensator:
			controlEnabled = v.ControlEnabled
			sections = v.Sections
			if v.RegulatingControl != nil {
				rcID := strings.TrimPrefix(v.RegulatingControl.MRID, "#")
				if rc, ok := dataset.RegulatingControls[rcID]; ok {
					rcEnabled = rc.Enabled
				}
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
		if !inService {
			continue
		}

		if !controlEnabled || !rcEnabled {
			if svsc.Sections != sections {
				violations = append(violations, Violation{
					ObjectID: scID, Class: goTypeName(scObj), Property: "ShuntCompensator.sections",
					Message:  fmt.Sprintf("SvShuntCompensatorSections.sections (%v) is not the same as ShuntCompensator.sections (%v) for non-regulating ShuntCompensator.", svsc.Sections, sections),
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
		rcEnabled := true
		var step float64

		val := reflect.ValueOf(tcObj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		if f := val.FieldByName("ControlEnabled"); f.IsValid() {
			controlEnabled = f.Bool()
		}
		if f := val.FieldByName("Step"); f.IsValid() {
			step = f.Float()
		}
		if f := val.FieldByName("TapChangerControl"); f.IsValid() && !f.IsNil() {
			tccID := strings.TrimPrefix(f.Elem().FieldByName("MRID").String(), "#")
			if tcc, ok := dataset.TapChangerControls[tccID]; ok {
				rcEnabled = tcc.Enabled
			}
		}

		if !controlEnabled || !rcEnabled {
			if svts.Position != step {
				violations = append(violations, Violation{
					ObjectID: tcID, Class: goTypeName(tcObj), Property: "TapChanger.step",
					Message:  fmt.Sprintf("SvTapStep.position (%v) is not the same as TapChanger.step (%v) for non-regulating TapChanger.", svts.Position, step),
					Severity: "sh.Violation",
				})
			}
		}
	}

	return violations
}

// CheckSvStatusInstance implements mas600:SvStatus-SV__4
// Profile: 61970-600-1_AllProfiles-AP-Con-Complex-SolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: SvStatus must be instantiated for all energized ConductingEquipment connected to a TopologicalIsland.
func CheckSvStatusInstance(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	tnInIsland := make(map[string]bool)
	for _, island := range dataset.TopologicalIslands {
		if island.TopologicalNodes != nil {
			tnInIsland[strings.TrimPrefix(island.TopologicalNodes.MRID, "#")] = true
		}
	}

	for id, obj := range dataset.Elements {
		typeName := goTypeName(obj)
		isCE := strings.HasPrefix(typeName, "Synchronous") || strings.HasPrefix(typeName, "Asynchronous") ||
			strings.HasPrefix(typeName, "Energy") || strings.HasPrefix(typeName, "Line") ||
			strings.HasPrefix(typeName, "Breaker") || strings.HasPrefix(typeName, "Disconnector")
		if !isCE || typeName == "Equipment" {
			continue
		}

		energized := false
		for _, term := range dataset.Terminals {
			if term.ConductingEquipment != nil && strings.TrimPrefix(term.ConductingEquipment.MRID, "#") == id {
				if term.TopologicalNode != nil && tnInIsland[strings.TrimPrefix(term.TopologicalNode.MRID, "#")] {
					energized = true
					break
				}
			}
		}
		if !energized {
			continue
		}

		found := false
		for _, svs := range dataset.SvStatuss {
			if svs.ConductingEquipment != nil && strings.TrimPrefix(svs.ConductingEquipment.MRID, "#") == id {
				found = true
				break
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

	return violations
}

// CheckSvShuntCompensatorSectionsInstance implements mas600:SvShuntCompensatorSections-SV__4
// Profile: 61970-600-1_AllProfiles-AP-Con-Complex-SolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: SvShuntCompensatorSections must be instantiated for all energized ShuntCompensators.
func CheckSvShuntCompensatorSectionsInstance(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	tnInIsland := make(map[string]bool)
	for _, island := range dataset.TopologicalIslands {
		if island.TopologicalNodes != nil {
			tnInIsland[strings.TrimPrefix(island.TopologicalNodes.MRID, "#")] = true
		}
	}

	for id, obj := range dataset.Elements {
		switch obj.(type) {
		case *cimgostructs.LinearShuntCompensator, *cimgostructs.NonlinearShuntCompensator:
		default:
			continue
		}

		energized := false
		for _, term := range dataset.Terminals {
			if term.ConductingEquipment != nil && strings.TrimPrefix(term.ConductingEquipment.MRID, "#") == id {
				if term.TopologicalNode != nil && tnInIsland[strings.TrimPrefix(term.TopologicalNode.MRID, "#")] {
					energized = true
					break
				}
			}
		}
		if !energized {
			continue
		}

		found := false
		for _, svsc := range dataset.SvShuntCompensatorSectionss {
			if svsc.ShuntCompensator != nil && strings.TrimPrefix(svsc.ShuntCompensator.MRID, "#") == id {
				found = true
				break
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

	return violations
}

// CheckSvTapStepInstance implements mas600:SvTapStep-SV__4
// Profile: 61970-600-1_AllProfiles-AP-Con-Complex-SolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: SvTapStep must be instantiated for all energized TapChangers.
func CheckSvTapStepInstance(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	tnInIsland := make(map[string]bool)
	for _, island := range dataset.TopologicalIslands {
		if island.TopologicalNodes != nil {
			tnInIsland[strings.TrimPrefix(island.TopologicalNodes.MRID, "#")] = true
		}
	}

	for id, obj := range dataset.Elements {
		switch obj.(type) {
		case *cimgostructs.RatioTapChanger, *cimgostructs.PhaseTapChangerLinear,
			*cimgostructs.PhaseTapChangerSymmetrical, *cimgostructs.PhaseTapChangerAsymmetrical,
			*cimgostructs.PhaseTapChangerTabular:
		default:
			continue
		}

		energized := false
		var teID string
		val := reflect.ValueOf(obj).Elem()
		if f := val.FieldByName("TransformerEnd"); f.IsValid() && !f.IsNil() {
			teID = strings.TrimPrefix(f.Elem().FieldByName("MRID").String(), "#")
		}
		if teObj, ok := dataset.Elements[teID]; ok {
			teVal := reflect.ValueOf(teObj).Elem()
			if f := teVal.FieldByName("Terminal"); f.IsValid() && !f.IsNil() {
				termID := strings.TrimPrefix(f.Elem().FieldByName("MRID").String(), "#")
				if term, ok := dataset.Terminals[termID]; ok {
					if term.TopologicalNode != nil && tnInIsland[strings.TrimPrefix(term.TopologicalNode.MRID, "#")] {
						energized = true
					}
				}
			}
		}
		if !energized {
			continue
		}

		found := false
		for _, svts := range dataset.SvTapSteps {
			if svts.TapChanger != nil && strings.TrimPrefix(svts.TapChanger.MRID, "#") == id {
				found = true
				break
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

	return violations
}

// CheckRegulatingControlSameIsland implements sm6002:RegulatingControl-point
// Profile: 61970-600-2_AllProfiles-AP-Con-Complex-SolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: The controlled point and the controlling equipment shall be located in the same TopologicalIsland.
func CheckRegulatingControlSameIsland(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	termToIsland := make(map[string]string)
	tnToIsland := make(map[string]string)
	for islandID, island := range dataset.TopologicalIslands {
		if island.TopologicalNodes != nil {
			tnID := strings.TrimPrefix(island.TopologicalNodes.MRID, "#")
			tnToIsland[tnID] = islandID
		}
	}
	for id, term := range dataset.Terminals {
		if term.TopologicalNode != nil {
			tnID := strings.TrimPrefix(term.TopologicalNode.MRID, "#")
			if islandID, ok := tnToIsland[tnID]; ok {
				termToIsland[id] = islandID
			}
		}
	}

	for id, rc := range dataset.RegulatingControls {
		if !rc.Enabled || rc.Terminal == nil {
			continue
		}
		rcTermID := strings.TrimPrefix(rc.Terminal.MRID, "#")
		rcIsland, ok := termToIsland[rcTermID]
		if !ok {
			continue
		}

		// 1. SynchronousMachines
		for _, sm := range dataset.SynchronousMachines {
			if sm.RegulatingControl != nil && strings.TrimPrefix(sm.RegulatingControl.MRID, "#") == id {
				for _, term := range dataset.Terminals {
					if term.ConductingEquipment != nil && strings.TrimPrefix(term.ConductingEquipment.MRID, "#") == sm.Id {
						if smIsland, ok := termToIsland[term.Id]; ok && smIsland != rcIsland {
							violations = append(violations, Violation{
								ObjectID: id, Class: "RegulatingControl", Property: "rdf:type",
								Message:  fmt.Sprintf("The controlled point and the controlling equipment (SynchronousMachine %s) are not located in the same TopologicalIsland.", sm.Id),
								Severity: "sh.Violation",
							})
						}
					}
				}
			}
		}

		// 2. TapChangers
		for tcID, obj := range dataset.Elements {
			var tcc *struct {
				MRID string "xml:\"resource,attr\""
			}
			var teID string
			class := ""

			switch v := obj.(type) {
			case *cimgostructs.RatioTapChanger:
				tcc, class = v.TapChangerControl, "RatioTapChanger"
				if v.TransformerEnd != nil {
					teID = strings.TrimPrefix(v.TransformerEnd.MRID, "#")
				}
			case *cimgostructs.PhaseTapChangerLinear:
				tcc, class = v.TapChangerControl, "PhaseTapChangerLinear"
				if v.TransformerEnd != nil {
					teID = strings.TrimPrefix(v.TransformerEnd.MRID, "#")
				}
			case *cimgostructs.PhaseTapChangerSymmetrical:
				tcc, class = v.TapChangerControl, "PhaseTapChangerSymmetrical"
				if v.TransformerEnd != nil {
					teID = strings.TrimPrefix(v.TransformerEnd.MRID, "#")
				}
			case *cimgostructs.PhaseTapChangerAsymmetrical:
				tcc, class = v.TapChangerControl, "PhaseTapChangerAsymmetrical"
				if v.TransformerEnd != nil {
					teID = strings.TrimPrefix(v.TransformerEnd.MRID, "#")
				}
			case *cimgostructs.PhaseTapChangerTabular:
				tcc, class = v.TapChangerControl, "PhaseTapChangerTabular"
				if v.TransformerEnd != nil {
					teID = strings.TrimPrefix(v.TransformerEnd.MRID, "#")
				}
			default:
				continue
			}

			if tcc != nil && strings.TrimPrefix(tcc.MRID, "#") == id {
				if teObj, ok := dataset.Elements[teID]; ok {
					teVal := reflect.ValueOf(teObj).Elem()
					termField := teVal.FieldByName("Terminal")
					if termField.IsValid() && !termField.IsNil() {
						termID := strings.TrimPrefix(termField.Elem().FieldByName("MRID").String(), "#")
						if tcIsland, ok := termToIsland[termID]; ok && tcIsland != rcIsland {
							violations = append(violations, Violation{
								ObjectID: id, Class: "RegulatingControl", Property: "rdf:type",
								Message:  fmt.Sprintf("The controlled point and the controlling equipment (%s %s) are not located in the same TopologicalIsland.", class, tcID),
								Severity: "sh.Violation",
							})
						}
					}
				}
			}
		}
	}

	return violations
}
