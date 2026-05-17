package validation

import (
	"cimgo/cimstructs"
	"fmt"
	"strings"
)

// ValidateSCNotSolvedMASProfileSPARQL runs hand-written checks for
// 61970-301_ShortCircuit-AP-Con-Complex-NotSolvedMAS-SHACL.
func ValidateSCNotSolvedMASProfileSPARQL(dataset *cimstructs.CIMElementList) []Violation {
	return CheckMutualCouplingTerminalsAssignment(dataset)
}

// CheckMutualCouplingTerminalsAssignment implements sccns.MutualCoupling-terminalsAssignment
// Profile: 61970-301_ShortCircuit-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: The first and second terminals of a mutual coupling should point to different
// ACLineSegments (or generic Equipment).
func CheckMutualCouplingTerminalsAssignment(dataset *cimstructs.CIMElementList) []Violation {
	var violations []Violation

	conductingEquipmentOf := func(termRef *struct {
		MRID string `xml:"resource,attr"`
	}) (string, interface{}, bool) {
		if termRef == nil {
			return "", nil, false
		}
		termID := strings.TrimPrefix(termRef.MRID, "#")
		termObj, ok := dataset.Elements[termID]
		if !ok {
			return "", nil, false
		}
		term, ok := termObj.(*cimstructs.Terminal)
		if !ok || term.ConductingEquipment == nil {
			return "", nil, false
		}
		eqID := strings.TrimPrefix(term.ConductingEquipment.MRID, "#")
		eqObj, ok := dataset.Elements[eqID]
		if !ok {
			return eqID, nil, true
		}
		return eqID, eqObj, true
	}

	for id, obj := range dataset.Elements {
		mc, ok := obj.(*cimstructs.MutualCoupling)
		if !ok {
			continue
		}
		eq1ID, eq1Obj, ok1 := conductingEquipmentOf(mc.First_Terminal)
		eq2ID, eq2Obj, ok2 := conductingEquipmentOf(mc.Second_Terminal)
		if !ok1 || !ok2 {
			continue
		}

		isLineLike := func(o interface{}) bool {
			if o == nil {
				return false
			}
			switch o.(type) {
			case *cimstructs.ACLineSegment, *cimstructs.Equipment:
				return true
			}
			return false
		}

		if !isLineLike(eq1Obj) || !isLineLike(eq2Obj) || eq1ID == eq2ID {
			t1 := goTypeName(eq1Obj)
			t2 := goTypeName(eq2Obj)
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "MutualCoupling",
				Property: "MutualCoupling.First_Terminal",
				Message:  fmt.Sprintf("The terminals are either not related to ACLineSegment or the first and the second terminal associations are not pointing to different ACLineSegments. Type line 1: %s. Type line 2: %s.", t1, t2),
				Severity: "sh:Violation",
			})
		}
	}

	return violations
}
