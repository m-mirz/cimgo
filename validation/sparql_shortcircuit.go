package validation

import "cimgo/cimgostructs"

// CheckSeriesCompensatorVaristorUsage implements scc.SeriesCompensator.varistorRatedCurrent-usage
func CheckSeriesCompensatorVaristorUsage(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		sc, ok := obj.(*cimgostructs.SeriesCompensator)
		if !ok {
			continue
		}

		if !sc.VaristorPresent {
			if sc.VaristorRatedCurrent != 0 || sc.VaristorVoltageThreshold != 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "SeriesCompensator",
					Property: "SeriesCompensator.varistorRatedCurrent",
					Message:  "The varistor attributes are present and SeriesCompensator.varistorPresent is false.",
					Severity: "sh.Violation",
				})
			}
		}
	}

	return violations
}

// CheckTransformerEndGrounding implements sc452:TransformerEnd-grounding
func CheckTransformerEndGrounding(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	for id, te := range dataset.PowerTransformerEnds {
		if te.Grounded {
			if te.Rground == 0 && te.Xground == 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "PowerTransformerEnd",
					Property: "grounded",
					Message:  "Missing required properties .rground or .xground when grounded=true.",
					Severity: "sh.Violation",
				})
			}
		}
	}
	return violations
}

// CheckSynchronousMachineEarthing implements sc452:SynchronousMachine-attributes
func CheckSynchronousMachineEarthing(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	for id, sm := range dataset.SynchronousMachines {
		if sm.Earthing {
			if sm.EarthingStarPointR == 0 && sm.EarthingStarPointX == 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "SynchronousMachine",
					Property: "earthing",
					Message:  "Missing required properties .earthingStarPointR or .earthingStarPointX when earthing=true.",
					Severity: "sh.Violation",
				})
			}
		}
	}
	return violations
}

// CheckSeriesCompensatorVaristorRequired implements sc600:SeriesCompensator.varistorRatedCurrent-required
func CheckSeriesCompensatorVaristorRequired(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	for id, sc := range dataset.SeriesCompensators {
		if sc.VaristorPresent {
			if sc.VaristorRatedCurrent == 0 || sc.VaristorVoltageThreshold == 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "SeriesCompensator",
					Property: "varistorPresent",
					Message:  "Missing required property .varistorRatedCurrent or .varistorVoltageThreshold when varistorPresent=true.",
					Severity: "sh.Violation",
				})
			}
		}
	}
	return violations
}
