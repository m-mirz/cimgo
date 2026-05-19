package validation

import (
	"cimgo/cimstructs"
	"fmt"
	"strings"
)

// ValidateTPNotSolvedMASProfileSPARQL runs hand-written checks for
// 61970-301_Topology-AP-Con-Complex-NotSolvedMAS-SHACL and
// 61970-600_Topology-AP-Con-Complex-NotSolvedMAS-SHACL.
func ValidateTPNotSolvedMASProfileSPARQL(dataset *cimstructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckTerminalPhasesConsistencyTopologicalNode(dataset)...)
	violations = append(violations, CheckSwitchSameTopologicalNode(dataset)...)
	violations = append(violations, CheckTerminalExch8TopologicalNode(dataset)...)
	return violations
}

// CheckTerminalPhasesConsistencyTopologicalNode implements topcns.Terminal.phases-consistencyTopologicalNode
// Profile: 61970-301_Topology-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: The phase code on terminals connecting the same TopologicalNode shall be consistent.
func CheckTerminalPhasesConsistencyTopologicalNode(dataset *cimstructs.CIMElementList) []Violation {
	var violations []Violation

	nodeTerminals := make(map[string][]*cimstructs.Terminal)
	for _, term := range dataset.Terminals {
		if term.TopologicalNode != nil {
			nodeID := strings.TrimPrefix(term.TopologicalNode.MRID, "#")
			nodeTerminals[nodeID] = append(nodeTerminals[nodeID], term)
		}
	}

	abcn := "http://iec.ch/TC57/CIM100#PhaseCode.ABCN"
	n := "http://iec.ch/TC57/CIM100#PhaseCode.N"
	abc := "http://iec.ch/TC57/CIM100#PhaseCode.ABC"

	for nodeID, terms := range nodeTerminals {
		if len(terms) < 2 {
			continue
		}
		for i := 0; i < len(terms); i++ {
			for j := i + 1; j < len(terms); j++ {
				val1 := ""
				if terms[i].Phases != nil {
					val1 = terms[i].Phases.URI
				}
				val2 := ""
				if terms[j].Phases != nil {
					val2 = terms[j].Phases.URI
				}

				failed := false
				if val1 != "" && val2 != "" {
					if (val1 == abcn || val1 == n) && (val2 != abcn && val2 != n) {
						failed = true
					} else if val1 == abc && val2 != abc {
						failed = true
					}
				} else if val1 != "" && val2 == "" {
					if val1 == abcn || val1 == n {
						failed = true
					}
				}

				if failed {
					violations = append(violations, Violation{
						ObjectID: nodeID,
						RuleID:   "topcns.Terminal.phases-consistencyTopologicalNode",
						Name:     "Terminal.phases-consistencyTopologicalNode",
						Class:    "TopologicalNode",
						Property: "Terminal.phases",
						Message:  fmt.Sprintf("The phase codes for the connected terminals are not consistent. Terminal %s code: %s, Terminal %s code: %s.", terms[i].MRID, val1, terms[j].MRID, val2),
						Severity: "sh:Violation",
					})
					goto NextNode
				}
			}
		}
	NextNode:
	}

	return violations
}

// CheckSwitchSameTopologicalNode implements topc456ns:Switch-sameTopologicalNode
// Profile: 61970-456_Topology-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: Terminals of a retained Switch shall not be connected to the same TopologicalNode.
func CheckSwitchSameTopologicalNode(dataset *cimstructs.CIMElementList) []Violation {
	var violations []Violation

	// Build index: ConductingEquipment ID -> []Terminal
	eqTerminals := make(map[string][]*cimstructs.Terminal)
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
		case *cimstructs.Switch:
			retained, class = v.Retained, "Switch"
		case *cimstructs.Breaker:
			retained, class = v.Retained, "Breaker"
		case *cimstructs.Disconnector:
			retained, class = v.Retained, "Disconnector"
		case *cimstructs.Fuse:
			retained, class = v.Retained, "Fuse"
		case *cimstructs.Jumper:
			retained, class = v.Retained, "Jumper"
		case *cimstructs.LoadBreakSwitch:
			retained, class = v.Retained, "LoadBreakSwitch"
		case *cimstructs.Cut:
			retained, class = v.Retained, "Cut"
		case *cimstructs.GroundDisconnector:
			retained, class = v.Retained, "GroundDisconnector"
		case *cimstructs.DisconnectingCircuitBreaker:
			retained, class = v.Retained, "DisconnectingCircuitBreaker"
		default:
			continue
		}

		if !retained {
			continue
		}

		terms := eqTerminals[id]
		if len(terms) < 2 {
			continue
		}

		// Find terminals with sequenceNumber 1 and 2
		var t1, t2 *cimstructs.Terminal
		for _, t := range terms {
			if t.SequenceNumber == 1 {
				t1 = t
			}
			if t.SequenceNumber == 2 {
				t2 = t
			}
		}

		if t1 == nil || t2 == nil {
			continue
		}

		getTN := func(t *cimstructs.Terminal) string {
			if t.TopologicalNode != nil {
				return strings.TrimPrefix(t.TopologicalNode.MRID, "#")
			}
			if t.ConnectivityNode != nil {
				cnID := strings.TrimPrefix(t.ConnectivityNode.MRID, "#")
				if cn, ok := dataset.Elements[cnID].(*cimstructs.ConnectivityNode); ok && cn.TopologicalNode != nil {
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
				RuleID:   "topc456ns:Switch-sameTopologicalNode",
				Name:     "Switch-sameTopologicalNode",
				Class:    class,
				Property: "retained",
				Message:  "Terminals of retained Switch connect to the same TopologicalNode.",
				Severity: "sh:Violation",
			})
		}
	}

	return violations
}

// CheckTerminalExch8TopologicalNode implements topc600ns:Terminal-EXCH8TopologicalNode
// Profile: 61970-600_Topology-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: Terminal.TopologicalNode is required if a RegulatingControl is associated.
func CheckTerminalExch8TopologicalNode(dataset *cimstructs.CIMElementList) []Violation {
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
						RuleID:   "topc600ns:Terminal-EXCH8TopologicalNode",
						Name:     "Terminal-EXCH8TopologicalNode",
						Class:    "Terminal",
						Property: "TopologicalNode",
						Message:  "The Terminal is referenced by a RegulatingControl but is not associated with a TopologicalNode.",
						Severity: "sh:Violation",
					})
				}
			}
		}
	}

	return violations
}
