package validation

import (
	"cimgo/cimgostructs"
	"strings"
)

// CheckBoundaryPointTieFlow implements eqbdn301:BoundaryPoint.isExcludedFromAreaInterchange-requiredTieFlow
// Description: If isExcludedFromAreaInterchange is false (default), a TieFlow is required.
func CheckBoundaryPointTieFlow(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	// Build index: Terminal ID -> []TieFlow ID
	terminalTieFlows := make(map[string][]string)
	for id, tf := range dataset.TieFlows {
		if tf.Terminal != nil {
			termID := strings.TrimPrefix(tf.Terminal.MRID, "#")
			terminalTieFlows[termID] = append(terminalTieFlows[termID], id)
		}
	}

	for id, bp := range dataset.BoundaryPoints {
		if bp.ConnectivityNode == nil {
			continue
		}
		cnID := strings.TrimPrefix(bp.ConnectivityNode.MRID, "#")

		// Find terminals at this CN
		var bpTerminals []string
		for termID, term := range dataset.Terminals {
			if term.ConnectivityNode != nil && strings.TrimPrefix(term.ConnectivityNode.MRID, "#") == cnID {
				bpTerminals = append(bpTerminals, termID)
			}
		}

		hasTieFlow := false
		for _, termID := range bpTerminals {
			if len(terminalTieFlows[termID]) > 0 {
				hasTieFlow = true
				break
			}
		}

		if bp.IsExcludedFromAreaInterchange {
			if hasTieFlow {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "BoundaryPoint",
					Property: "isExcludedFromAreaInterchange",
					Message:  "TieFlow is modelled but isExcludedFromAreaInterchange is true.",
					Severity: "sh.Violation",
				})
			}
		} else {
			if !hasTieFlow {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "BoundaryPoint",
					Property: "isExcludedFromAreaInterchange",
					Message:  "TieFlow is required but not modelled for this BoundaryPoint.",
					Severity: "sh.Violation",
				})
			}
		}
	}

	return violations
}
