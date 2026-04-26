package validation

import (
	"cimgo/cimgostructs"
	"fmt"
	"reflect"
	"strings"
)

type Violation struct {
	ObjectID string
	Class    string
	Property string
	Message  string
	Severity string
}

func getCIMTypeNameSPARQL(obj interface{}) string {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

// CheckACDCTerminalSequenceNumbering implements eqc.ACDCTerminal.sequenceNumber-numbering
// Profile: 61970-301_Equipment-AP-Con-Complex
// Description: The sequence numbering starts with 1 and additional terminals should follow in increasing order.
// The first terminal is the starting point for a two terminal branch.
func CheckACDCTerminalSequenceNumbering(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	equipmentTerminals := make(map[string][]interface{})

	for _, term := range dataset.Terminals {
		if term.ConductingEquipment != nil {
			id := strings.TrimPrefix(term.ConductingEquipment.MRID, "#")
			equipmentTerminals[id] = append(equipmentTerminals[id], term)
		}
	}
	for _, term := range dataset.DCTerminals {
		if term.DCConductingEquipment != nil {
			id := strings.TrimPrefix(term.DCConductingEquipment.MRID, "#")
			equipmentTerminals[id] = append(equipmentTerminals[id], term)
		}
	}

	for eqID, terms := range equipmentTerminals {
		countsn := len(terms)
		seenSN := make(map[int]bool)
		minSN := 999999
		sumSN := 0

		for _, term := range terms {
			var sn int
			switch t := term.(type) {
			case *cimgostructs.Terminal:
				sn = t.SequenceNumber
			case *cimgostructs.DCTerminal:
				sn = t.SequenceNumber
			}

			seenSN[sn] = true
			if sn < minSN {
				minSN = sn
			}
			sumSN += sn
		}

		countdsn := len(seenSN)
		countterms := countsn

		failed := false
		if countsn != countdsn {
			failed = true
		} else if minSN != 1 {
			failed = true
		} else if countterms == 1 && sumSN != 1 {
			failed = true
		} else if countterms == 2 && sumSN != 3 {
			failed = true
		} else if countterms == 3 && sumSN != 6 {
			failed = true
		}

		if failed {
			violations = append(violations, Violation{
				ObjectID: eqID,
				Class:    "ConductingEquipment",
				Property: "ACDCTerminal.sequenceNumber",
				Message:  "There is no terminal with sequenceNumber=1 or the numbering is not unique.",
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckTerminalPhasesConsistencyEquipment implements eqc.Terminal.phases-consistencyEquipment
// Profile: 61970-301_Equipment-AP-Con-Complex
// Description: The phase code on terminals connecting same ConnectivityNode or same TopologicalNode
// as well as for equipment between two terminals shall be consistent.
func CheckTerminalPhasesConsistencyEquipment(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	equipmentTerminals := make(map[string]map[int]*cimgostructs.Terminal)

	for _, term := range dataset.Terminals {
		if term.ConductingEquipment != nil {
			eqID := strings.TrimPrefix(term.ConductingEquipment.MRID, "#")
			if _, ok := equipmentTerminals[eqID]; !ok {
				equipmentTerminals[eqID] = make(map[int]*cimgostructs.Terminal)
			}
			equipmentTerminals[eqID][term.SequenceNumber] = term
		}
	}

	for eqID, terms := range equipmentTerminals {
		term1, ok1 := terms[1]
		term2, ok2 := terms[2]

		if !ok1 || !ok2 {
			continue
		}

		val1 := ""
		if term1.Phases != nil {
			val1 = term1.Phases.URI
		}
		val2 := ""
		if term2.Phases != nil {
			val2 = term2.Phases.URI
		}

		abcn := "http://iec.ch/TC57/CIM100#PhaseCode.ABCN"
		n := "http://iec.ch/TC57/CIM100#PhaseCode.N"
		abc := "http://iec.ch/TC57/CIM100#PhaseCode.ABC"

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
				ObjectID: eqID,
				Class:    "ConductingEquipment",
				Property: "Terminal.phases",
				Message:  fmt.Sprintf("The phase codes for terminals of 2-terminal equipment are not consistent. Terminal 1 code:%s Terminal 2 code: %s.", val1, val2),
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckConductingEquipmentBaseVoltageUsage implements eqc.ConductingEquipment.BaseVoltage-usage
// Profile: 61970-301_Equipment-AP-Con-Complex
// Description: Use only when there is no voltage level container used and only one base voltage applies.
// For example, not used for transformers.
func CheckConductingEquipmentBaseVoltageUsage(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		typeName := getCIMTypeNameSPARQL(obj)
		if typeName == "ACLineSegment" || typeName == "EquivalentBranch" || typeName == "SeriesCompensator" || typeName == "Equipment" {
			continue
		}

		val := reflect.ValueOf(obj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		bvField := val.FieldByName("BaseVoltage")
		if !bvField.IsValid() || (bvField.Kind() == reflect.Ptr && bvField.IsNil()) {
			continue
		}

		ecField := val.FieldByName("EquipmentContainer")
		if ecField.IsValid() && ecField.Kind() == reflect.Ptr && !ecField.IsNil() {
			mridField := ecField.Elem().FieldByName("MRID")
			if !mridField.IsValid() {
				continue
			}
			ecMRID := mridField.String()
			ecID := strings.TrimPrefix(ecMRID, "#")

			if ecObj, ok := dataset.Elements[ecID]; ok {
				if getCIMTypeNameSPARQL(ecObj) == "VoltageLevel" {
					violations = append(violations, Violation{
						ObjectID: id,
						Class:    typeName,
						Property: "Equipment.EquipmentContainer",
						Message:  "The association ConductingEquipment.BaseVoltage is defined for a ConductingEquipment contained in a VoltageLevel.",
						Severity: "sh.Violation",
					})
				}
			}
		}
	}

	return violations
}
