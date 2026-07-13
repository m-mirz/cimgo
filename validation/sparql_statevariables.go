package validation

import "cimgo/cimstructs"

// ValidateSVProfileSPARQL runs hand-written checks for
// 61970-301_StateVariables-AP-Con-Complex-SHACL and
// 61970-456_StateVariables-AP-Con-Complex-SHACL.
func ValidateSVProfileSPARQL(dataset *cimstructs.CIMDataset) []Violation {
	var violations []Violation
	violations = append(violations, CheckCsConverterStateValueRange(dataset)...)
	violations = append(violations, CheckTopologicalIslandCount(dataset)...)
	return violations
}

// CheckCsConverterStateValueRange implements svc.CsConverter.alpha/gamma-valueRangeTypical
// Profile: 61970-301_StateVariables-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: alpha and gamma values should be within typical ranges for rectifier and inverter modes respectively.
func CheckCsConverterStateValueRange(dataset *cimstructs.CIMDataset) []Violation {
	var violations []Violation

	for id, csc := range dataset.CsConverters {
		if csc.OperatingMode == nil {
			continue
		}

		mode := csc.OperatingMode.URI
		rectifier := "http://iec.ch/TC57/CIM100#CsOperatingModeKind.rectifier"
		inverter := "http://iec.ch/TC57/CIM100#CsOperatingModeKind.inverter"

		if mode == rectifier {
			if csc.Alpha < 10 || csc.Alpha > 18 {
				violations = append(violations, Violation{
					ObjectID: id,
					RuleID:   "svu:CsConverter.alpha-valueRangeTypical",
					Name:     "C:301:SV:CsConverter.alpha:valueRangeTypical",
					Class:    "CsConverter",
					Property: "CsConverter.alpha",
					Message:  "The alpha value is outside typical range (10-18 degrees) for a rectifier.",
					Severity: "sh:Warning",
				})
			}
		} else if mode == inverter {
			if csc.Gamma < 17 || csc.Gamma > 20 {
				violations = append(violations, Violation{
					ObjectID: id,
					RuleID:   "svu:CsConverter.gamma-valueRangeTypical",
					Name:     "C:301:SV:CsConverter.gamma:valueRangeTypical",
					Class:    "CsConverter",
					Property: "CsConverter.gamma",
					Message:  "The gamma value is outside typical range (17-20 degrees) for an inverter.",
					Severity: "sh:Warning",
				})
			}
		}
	}

	return violations
}

// CheckTopologicalIslandCount implements sv456:TopologicalIsland-instance
// Profile: 61970-456_StateVariables-AP-Con-Complex
// Origin: Derived from a complex SHACL constraint (minCount 1 with inversePath) that was too complex for automated code generation.
// Description: At least one TopologicalIsland instance shall be present per SV instance.
func CheckTopologicalIslandCount(dataset *cimstructs.CIMDataset) []Violation {
	var violations []Violation

	if len(dataset.TopologicalIslands) == 0 {
		violations = append(violations, Violation{
			ObjectID: "global",
			RuleID:   "sv456:TopologicalIsland-instance",
			Name:     "C:456:SV:TopologicalIsland:instance",
			Class:    "TopologicalIsland",
			Property: "rdf:type",
			Message:  "No TopologicalIsland instantiated.",
			Severity: "sh:Violation",
		})
	}

	return violations
}
