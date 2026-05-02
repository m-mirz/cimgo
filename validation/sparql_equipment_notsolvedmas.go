package validation

import (
	"cimgo/cimgostructs"
	"fmt"
	"strings"
)

// CheckACLineSegmentBaseVoltage implements eqcns.ACLineSegment-baseVoltage
// Description: The BaseVoltage at the two ends of ACLineSegments in a Line shall have the same
// BaseVoltage.nominalVoltage.
func CheckACLineSegmentBaseVoltage(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	terminalsByEquipment := make(map[string]map[int]*cimgostructs.Terminal)
	for _, term := range dataset.Terminals {
		if term.ConductingEquipment == nil {
			continue
		}
		eqID := strings.TrimPrefix(term.ConductingEquipment.MRID, "#")
		if _, ok := terminalsByEquipment[eqID]; !ok {
			terminalsByEquipment[eqID] = make(map[int]*cimgostructs.Terminal)
		}
		terminalsByEquipment[eqID][term.SequenceNumber] = term
	}

	nominalVoltage := func(term *cimgostructs.Terminal) (float64, bool) {
		if term == nil || term.TopologicalNode == nil {
			return 0, false
		}
		tnID := strings.TrimPrefix(term.TopologicalNode.MRID, "#")
		tnObj, ok := dataset.Elements[tnID]
		if !ok {
			return 0, false
		}
		tn, ok := tnObj.(*cimgostructs.TopologicalNode)
		if !ok || tn.BaseVoltage == nil {
			return 0, false
		}
		bvID := strings.TrimPrefix(tn.BaseVoltage.MRID, "#")
		bvObj, ok := dataset.Elements[bvID]
		if !ok {
			return 0, false
		}
		bv, ok := bvObj.(*cimgostructs.BaseVoltage)
		if !ok {
			return 0, false
		}
		return bv.NominalVoltage, true
	}

	for id, obj := range dataset.Elements {
		if _, ok := obj.(*cimgostructs.ACLineSegment); !ok {
			continue
		}
		terms := terminalsByEquipment[id]
		t1, t2 := terms[1], terms[2]
		v1, ok1 := nominalVoltage(t1)
		v2, ok2 := nominalVoltage(t2)
		if !ok1 || !ok2 {
			continue
		}
		if t1.TopologicalNode.MRID == t2.TopologicalNode.MRID {
			continue
		}
		if v1 != v2 {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "ACLineSegment",
				Property: "ACLineSegment.BaseVoltage",
				Message:  fmt.Sprintf("The ACLineSegment has different BaseVoltage.nominalVoltage at the two ends. Voltage at end 1 is: %v. Voltage at end 2 is: %v.", v1, v2),
				Severity: "sh.Warning",
			})
		}
	}

	return violations
}
