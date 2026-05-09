package validation

import (
	"cimgo/cimgostructs"
	"strings"
)

// CheckSwitchSameTopologicalNode implements topc456ns:Switch-sameTopologicalNode
// Profile: 61970-456_Topology-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: Terminals of a retained Switch shall not be connected to the same TopologicalNode.
func CheckSwitchSameTopologicalNode(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	// Build index: ConductingEquipment ID -> []Terminal
	eqTerminals := make(map[string][]*cimgostructs.Terminal)
	for _, term := range dataset.Terminals {
		if term.ConductingEquipment != nil {
			eqID := strings.TrimPrefix(term.ConductingEquipment.MRID, "#")
			eqTerminals[eqID] = append(eqTerminals[eqID], term)
		}
	}

	for id, obj := range dataset.Elements {
		// Extract Retained flag
		retained := false
		class := ""
		
		switch v := obj.(type) {
		case *cimgostructs.Switch: retained, class = v.Retained, "Switch"
		case *cimgostructs.Breaker: retained, class = v.Retained, "Breaker"
		case *cimgostructs.Disconnector: retained, class = v.Retained, "Disconnector"
		case *cimgostructs.Fuse: retained, class = v.Retained, "Fuse"
		case *cimgostructs.Jumper: retained, class = v.Retained, "Jumper"
		case *cimgostructs.LoadBreakSwitch: retained, class = v.Retained, "LoadBreakSwitch"
		case *cimgostructs.Cut: retained, class = v.Retained, "Cut"
		case *cimgostructs.GroundDisconnector: retained, class = v.Retained, "GroundDisconnector"
		case *cimgostructs.DisconnectingCircuitBreaker: retained, class = v.Retained, "DisconnectingCircuitBreaker"
		default: continue
		}

		if !retained {
			continue
		}
		
		terms := eqTerminals[id]
		if len(terms) < 2 {
			continue
		}
		
		// Find terminals with sequenceNumber 1 and 2
		var t1, t2 *cimgostructs.Terminal
		for _, t := range terms {
			if t.SequenceNumber == 1 { t1 = t }
			if t.SequenceNumber == 2 { t2 = t }
		}
		
		if t1 == nil || t2 == nil {
			continue
		}
		
		getTN := func(t *cimgostructs.Terminal) string {
			if t.TopologicalNode != nil {
				return strings.TrimPrefix(t.TopologicalNode.MRID, "#")
			}
			if t.ConnectivityNode != nil {
				cnID := strings.TrimPrefix(t.ConnectivityNode.MRID, "#")
				if cn, ok := dataset.Elements[cnID].(*cimgostructs.ConnectivityNode); ok && cn.TopologicalNode != nil {
					return strings.TrimPrefix(cn.TopologicalNode.MRID, "#")
				}
			}
			return ""
		}

		tn1 := getTN(t1)
		tn2 := getTN(t2)

		if tn1 != "" && tn1 == tn2 {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    class,
				Property: "retained",
				Message:  "Terminals of retained Switch connect to the same TopologicalNode.",
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckTerminalExch8TopologicalNode implements topc600ns:Terminal-EXCH8TopologicalNode
// Profile: 61970-600_Topology-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: Terminal.TopologicalNode is required if a RegulatingControl is associated.
func CheckTerminalExch8TopologicalNode(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	// Find terminals used in RegulatingControl
	rcTerminals := make(map[string]bool)
	for _, rc := range dataset.RegulatingControls {
		if rc.Terminal != nil {
			rcTerminals[strings.TrimPrefix(rc.Terminal.MRID, "#")] = true
		}
	}

	for id, term := range dataset.Terminals {
		if rcTerminals[id] {
			if term.TopologicalNode == nil {
				// Also check if ConnectivityNode has a TN
				hasTN := false
				if term.ConnectivityNode != nil {
					cnID := strings.TrimPrefix(term.ConnectivityNode.MRID, "#")
					if cn, ok := dataset.ConnectivityNodes[cnID]; ok && cn.TopologicalNode != nil {
						hasTN = true
					}
				}
				if !hasTN {
					violations = append(violations, Violation{
						ObjectID: id,
						Class:    "Terminal",
						Property: "TopologicalNode",
						Message:  "The Terminal is referenced by a RegulatingControl but is not associated with a TopologicalNode.",
						Severity: "sh.Violation",
					})
				}
			}
		}
	}

	return violations
}
