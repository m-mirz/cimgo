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
