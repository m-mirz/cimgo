package validation

import (
	"cimgo/cimgostructs"
	"strings"
)

// CheckSwitchSameTopologicalNode implements tpn456:Switch-sameTopologicalNode
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

	for id, sw := range dataset.Switchs {
		if !sw.Retained {
			continue
		}
		terms := eqTerminals[id]
		if len(terms) < 2 {
			continue
		}
		
		t1 := terms[0]
		t2 := terms[1]
		
		var tn1, tn2 string
		if t1.TopologicalNode != nil {
			tn1 = strings.TrimPrefix(t1.TopologicalNode.MRID, "#")
		} else if t1.ConnectivityNode != nil {
			cnID := strings.TrimPrefix(t1.ConnectivityNode.MRID, "#")
			if cn, ok := dataset.ConnectivityNodes[cnID]; ok && cn.TopologicalNode != nil {
				tn1 = strings.TrimPrefix(cn.TopologicalNode.MRID, "#")
			}
		}

		if t2.TopologicalNode != nil {
			tn2 = strings.TrimPrefix(t2.TopologicalNode.MRID, "#")
		} else if t2.ConnectivityNode != nil {
			cnID := strings.TrimPrefix(t2.ConnectivityNode.MRID, "#")
			if cn, ok := dataset.ConnectivityNodes[cnID]; ok && cn.TopologicalNode != nil {
				tn2 = strings.TrimPrefix(cn.TopologicalNode.MRID, "#")
			}
		}

		if tn1 != "" && tn1 == tn2 {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "Switch",
				Property: "retained",
				Message:  "Terminals of retained Switch connect to the same TopologicalNode.",
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckTerminalExch8TopologicalNode implements tpn600:Terminal-EXCH8TopologicalNode
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
