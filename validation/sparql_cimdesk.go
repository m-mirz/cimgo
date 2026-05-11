package validation

import (
	"cimgo/cimgostructs"
	"fmt"
)

// CheckBaseVoltageInEQBD implements C:600:ALL:NA:EQBD2.
// Description: Every BaseVoltage in the merged model must also be defined in the Boundary EQ (EQBD).
// BaseVoltages that exist only in IGM EQ files — not in the shared boundary dataset — are flagged.
// Severity: sh:Warning (per CIMdesk classification).
func CheckBaseVoltageInEQBD(dataset *cimgostructs.CIMElementList, eqbdBaseVoltageIDs map[string]struct{}) []Violation {
	var violations []Violation
	for id, bv := range dataset.BaseVoltages {
		if _, inEQBD := eqbdBaseVoltageIDs[id]; inEQBD {
			continue
		}
		violations = append(violations, Violation{
			ObjectID:    id,
			Class:       "BaseVoltage",
			Property:    "rdf:type",
			Message:     fmt.Sprintf("BaseVoltage (%.4g kV) is not defined in Boundary EQ.", bv.NominalVoltage),
			Severity:    "sh:Warning",
			RuleID:      "eqbd2:EQBD2",
			Name:        "EQBD2",
			Description: "The BaseVoltage is not defined in Boundary EQ.",
		})
	}
	return violations
}
