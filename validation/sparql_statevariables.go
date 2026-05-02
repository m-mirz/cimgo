package validation

import "cimgo/cimgostructs"

// CheckCsConverterStateValueRange implements svc.CsConverter.alpha/gamma-valueRangeTypical
func CheckCsConverterStateValueRange(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		csc, ok := obj.(*cimgostructs.CsConverter)
		if !ok || csc.OperatingMode == nil {
			continue
		}

		mode := csc.OperatingMode.URI
		rectifier := "http://iec.ch/TC57/CIM100#CsOperatingModeKind.rectifier"
		inverter := "http://iec.ch/TC57/CIM100#CsOperatingModeKind.inverter"

		if mode == rectifier {
			if csc.Alpha < 10 || csc.Alpha > 18 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "CsConverter",
					Property: "CsConverter.alpha",
					Message:  "The alpha value is outside typical range (10-18 degrees) for a rectifier.",
					Severity: "sh.Warning",
				})
			}
		} else if mode == inverter {
			if csc.Gamma < 17 || csc.Gamma > 20 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "CsConverter",
					Property: "CsConverter.gamma",
					Message:  "The gamma value is outside typical range (17-20 degrees) for an inverter.",
					Severity: "sh.Warning",
				})
			}
		}
	}

	return violations
}
