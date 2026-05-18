package validation

import "cimgo/cimstructs"

// ValidateSCProfileSPARQL runs hand-written checks for 61970-301_ShortCircuit-AP-Con-Complex-SHACL.
func ValidateSCProfileSPARQL(dataset *cimstructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckSeriesCompensatorVaristorUsage(dataset)...)
	violations = append(violations, CheckTransformerEndGrounding(dataset)...)
	violations = append(violations, CheckSynchronousMachineEarthing(dataset)...)
	violations = append(violations, CheckSeriesCompensatorVaristorRequired(dataset)...)
	return violations
}

// CheckSeriesCompensatorVaristorUsage implements scc.SeriesCompensator.varistorRatedCurrent-usage and scc.SeriesCompensator.varistorVoltageThreshold-usage
// Profile: 61970-301_ShortCircuit-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: varistorRatedCurrent and varistorVoltageThreshold are used for short circuit calculations and exchanged only if SeriesCompensator.varistorPresent is true.
func CheckSeriesCompensatorVaristorUsage(dataset *cimstructs.CIMElementList) []Violation {
	var violations []Violation

	for id, sc := range dataset.SeriesCompensators {
		if !sc.VaristorPresent {
			if sc.VaristorRatedCurrent != 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "SeriesCompensator",
					Property: "SeriesCompensator.varistorRatedCurrent",
					Message:  "The attribute is present and SeriesCompensator.varistorPresent is false.",
					Severity: "sh:Violation",
				})
			}
			if sc.VaristorVoltageThreshold != 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "SeriesCompensator",
					Property: "SeriesCompensator.varistorVoltageThreshold",
					Message:  "The attribute is present and SeriesCompensator.varistorPresent is false.",
					Severity: "sh:Violation",
				})
			}
		}
	}

	return violations
}

// CheckTransformerEndGrounding implements sc452:TransformerEnd-grounding
// Profile: 61970-452_ShortCircuit-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Missing required properties .rground or .xground when grounded=true.
func CheckTransformerEndGrounding(dataset *cimstructs.CIMElementList) []Violation {
	var violations []Violation
	for id, te := range dataset.PowerTransformerEnds {
		if te.Grounded {
			if te.Rground == 0 && te.Xground == 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "PowerTransformerEnd",
					Property: "grounded",
					Message:  "Missing required properties .rground or .xground when grounded=true.",
					Severity: "sh:Violation",
				})
			}
		}
	}
	return violations
}

// CheckSynchronousMachineEarthing implements sc452:SynchronousMachine-attributes
// Profile: 61970-452_ShortCircuit-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Missing required properties .earthingStarPointR or .earthingStarPointX when earthing=true.
func CheckSynchronousMachineEarthing(dataset *cimstructs.CIMElementList) []Violation {
	var violations []Violation
	for id, sm := range dataset.SynchronousMachines {
		if sm.Earthing {
			if sm.EarthingStarPointR == 0 && sm.EarthingStarPointX == 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "SynchronousMachine",
					Property: "earthing",
					Message:  "Missing required properties .earthingStarPointR or .earthingStarPointX when earthing=true.",
					Severity: "sh:Violation",
				})
			}
		}
	}
	return violations
}

// CheckSeriesCompensatorVaristorRequired implements scc600-2.SeriesCompensator.varistorRatedCurrent-required and scc600-2.SeriesCompensator.varistorVoltageThreshold-required
// Profile: 61970-600-2_ShortCircuit-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: The attributes varistorRatedCurrent and varistorVoltageThreshold are required if SeriesCompensator.varistorPresent is true.
func CheckSeriesCompensatorVaristorRequired(dataset *cimstructs.CIMElementList) []Violation {
	var violations []Violation
	for id, sc := range dataset.SeriesCompensators {
		if sc.VaristorPresent {
			if sc.VaristorRatedCurrent == 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "SeriesCompensator",
					Property: "SeriesCompensator.varistorRatedCurrent",
					Message:  "The attribute is missing when SeriesCompensator.varistorPresent is true.",
					Severity: "sh:Violation",
				})
			}
			if sc.VaristorVoltageThreshold == 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "SeriesCompensator",
					Property: "SeriesCompensator.varistorVoltageThreshold",
					Message:  "The attribute is missing when SeriesCompensator.varistorPresent is true.",
					Severity: "sh:Violation",
				})
			}
		}
	}
	return violations
}
