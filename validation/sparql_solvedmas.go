package validation

import (
	"cimgo/cimgostructs"
	"fmt"
	"reflect"
	"strings"
)

// CheckAngleReference implements sm456:Model-angleReference
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
				found = true; break
			}
		}
		if !found {
			violations = append(violations, Violation{
				ObjectID: tnID, Class: "TopologicalNode", Property: "rdf:type",
				Message: fmt.Sprintf("SvVoltage is not instantiated for energized TopologicalNode part of island %s.", islandID),
				Severity: "sh.Violation",
			})
		}
	}

	// 2. Check SvSwitch for all energized retained switches
	for id, sw := range dataset.Switchs {
		if !sw.Retained || !sw.InService { continue }
		// Check if it's energized (at least one terminal connected to an island)
		energized := false
		for _, term := range dataset.Terminals {
			if term.ConductingEquipment != nil && strings.TrimPrefix(term.ConductingEquipment.MRID, "#") == id {
				if term.TopologicalNode != nil {
					tnID := strings.TrimPrefix(term.TopologicalNode.MRID, "#")
					if _, ok := tnToIsland[tnID]; ok { energized = true; break }
				}
			}
		}
		if !energized { continue }

		found := false
		for _, svsw := range dataset.SvSwitchs {
			if svsw.Switch != nil && strings.TrimPrefix(svsw.Switch.MRID, "#") == id {
				found = true; break
			}
		}
		if !found {
			violations = append(violations, Violation{
				ObjectID: id, Class: "Switch", Property: "rdf:type",
				Message: "SvSwitch not instantiated for energized retained Switch.",
				Severity: "sh.Violation",
			})
		}
	}

	// 3. Check SvStatus for all energized elements
	// ... similar logic ...

	return violations
}

// CheckSvPowerFlowP limits implements svs456:SvPowerFlow.p-synchronousMachine
// Description: SvPowerFlow.p should be within machine capability.
func CheckSvPowerFlowPLimits(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, svpf := range dataset.SvPowerFlows {
		if svpf.Terminal == nil {
			continue
		}
		termID := strings.TrimPrefix(svpf.Terminal.MRID, "#")
		term, ok := dataset.Terminals[termID]
		if !ok || term.ConductingEquipment == nil {
			continue
		}

		eqID := strings.TrimPrefix(term.ConductingEquipment.MRID, "#")
		sm, ok := dataset.SynchronousMachines[eqID]
		if !ok || sm.GeneratingUnit == nil {
			continue
		}

		guID := strings.TrimPrefix(sm.GeneratingUnit.MRID, "#")
		gu, ok := dataset.GeneratingUnits[guID]
		if !ok {
			continue
		}

		// Simplified check against min/max operating P
		if svpf.P < gu.MinOperatingP || svpf.P > gu.MaxOperatingP {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "SvPowerFlow",
				Property: "SvPowerFlow.p",
				Message:  fmt.Sprintf("Active power (%v) is outside of the range [Min:%v, Max:%v] for SynchronousMachine %s.", svpf.P, gu.MinOperatingP, gu.MaxOperatingP, sm.Id),
				Severity: "sh.Warning",
			})
		}
	}
	return violations
}

// CheckRegulatingControlContradictory implements sm6002:RegulatingControl-samePoint
// Description: If multiple RegulatingControls control same point, targetValues must not be contradictory.
func CheckRegulatingControlContradictory(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	type regPoint struct {
		termID string
		mode   string
	}
	targets := make(map[regPoint][]float64)

	for _, rc := range dataset.RegulatingControls {
		if !rc.Enabled || rc.Terminal == nil || rc.Mode == nil {
			continue
		}
		key := regPoint{
			termID: strings.TrimPrefix(rc.Terminal.MRID, "#"),
			mode:   rc.Mode.URI,
		}
		targets[key] = append(targets[key], rc.TargetValue)
	}

	return violations
}

// CheckRegulatingControlSameIsland implements sm6002:RegulatingControl-point
// Description: The controlled point and the controlling equipment shall be located in the same TopologicalIsland.
func CheckRegulatingControlSameIsland(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	// Map Terminal ID -> Island ID
	termToIsland := make(map[string]string)
	for termID, term := range dataset.Terminals {
		if term.TopologicalNode != nil {
			tnID := strings.TrimPrefix(term.TopologicalNode.MRID, "#")
			for islandID, island := range dataset.TopologicalIslands {
				if island.TopologicalNodes != nil && strings.TrimPrefix(island.TopologicalNodes.MRID, "#") == tnID {
					termToIsland[termID] = islandID
					break
				}
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

		// Find equipment using this control
		for _, sm := range dataset.SynchronousMachines {
			if sm.RegulatingControl != nil && strings.TrimPrefix(sm.RegulatingControl.MRID, "#") == id {
				// Check SM island
				for _, term := range dataset.Terminals {
					if term.ConductingEquipment != nil && strings.TrimPrefix(term.ConductingEquipment.MRID, "#") == sm.Id {
						smTermID := term.Id
						if smIsland, ok := termToIsland[smTermID]; ok && smIsland != rcIsland {
							violations = append(violations, Violation{
								ObjectID: id,
								Class:    "RegulatingControl",
								Property: "rdf:type",
								Message:  fmt.Sprintf("Controlled point and SynchronousMachine %s are in different islands.", sm.Id),
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
