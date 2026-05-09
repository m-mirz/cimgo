package validation

import (
	"cimgo/cimgostructs"
	"fmt"
	"strings"
)

// CheckTerminalPhasesConsistencyTopologicalNode implements topcns.Terminal.phases-consistencyTopologicalNode
// Profile: 61970-301_Topology-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: The phase code on terminals connecting the same TopologicalNode shall be consistent.
func CheckTerminalPhasesConsistencyTopologicalNode(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	nodeTerminals := make(map[string][]*cimgostructs.Terminal)
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
						Class:    "TopologicalNode",
						Property: "Terminal.phases",
						Message:  fmt.Sprintf("The phase codes for the connected terminals are not consistent. Terminal %s code: %s, Terminal %s code: %s.", terms[i].MRID, val1, terms[j].MRID, val2),
						Severity: "sh.Violation",
					})
					goto NextNode
				}
			}
		}
	NextNode:
	}

	return violations
}
