package validation

import (
	"cimgo/cimgostructs"
	"fmt"
	"strings"
)

// CheckLinearShuntCompensatorSectionsRange implements sshcns.ShuntCompensator.sections-valueLinear
// Description: For LinearShuntCompensator the value shall be between zero and ShuntCompensator.maximumSections.
func CheckLinearShuntCompensatorSectionsRange(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		lsc, ok := obj.(*cimgostructs.LinearShuntCompensator)
		if !ok {
			continue
		}
		if lsc.Sections < 0 || lsc.Sections > float64(lsc.MaximumSections) {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "LinearShuntCompensator",
				Property: "ShuntCompensator.sections",
				Message:  fmt.Sprintf("The value (%v) is not between zero and ShuntCompensator.maximumSections (%d).", lsc.Sections, lsc.MaximumSections),
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckNonlinearShuntCompensatorSectionsValid implements sshcns.ShuntCompensator.sections-valueNonLinear
// Description: For NonlinearShuntCompensator-s, sections shall only be set to one of the
// NonlinearShuntCompenstorPoint.sectionNumber.
func CheckNonlinearShuntCompensatorSectionsValid(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	pointSections := make(map[string]map[int]bool)
	for _, obj := range dataset.Elements {
		point, ok := obj.(*cimgostructs.NonlinearShuntCompensatorPoint)
		if !ok || point.NonlinearShuntCompensator == nil {
			continue
		}
		nscID := strings.TrimPrefix(point.NonlinearShuntCompensator.MRID, "#")
		if _, ok := pointSections[nscID]; !ok {
			pointSections[nscID] = make(map[int]bool)
		}
		pointSections[nscID][point.SectionNumber] = true
	}

	for id, obj := range dataset.Elements {
		nsc, ok := obj.(*cimgostructs.NonlinearShuntCompensator)
		if !ok {
			continue
		}
		section := nsc.Sections
		if section != float64(int(section)) || !pointSections[id][int(section)] {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "NonlinearShuntCompensator",
				Property: "ShuntCompensator.sections",
				Message:  fmt.Sprintf("The value (%v) does not equal one of the NonlinearShuntCompenstorPoint.sectionNumber.", section),
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckRegulatingControlPowerFactorRequiredAttrs implements sshcns.RegulatingControl-requiredAttributes
// Description: When mode=powerFactor, both minAllowedTargetValue and maxAllowedTargetValue must be present.
func CheckRegulatingControlPowerFactorRequiredAttrs(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	check := func(id, class string, mode *struct {
		URI string `xml:"resource,attr"`
	}, minVal, maxVal float64) {
		if mode == nil || mode.URI != cimgostructs.RegulatingControlModeKindpowerFactor {
			return
		}
		if minVal == 0 || maxVal == 0 {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    class,
				Property: "RegulatingControl.mode",
				Message:  "Both minAllowedTargetValue and maxAllowedTargetValue are not provided for RegulatingControl in mode powerFactor.",
				Severity: "sh.Violation",
			})
		}
	}

	for id, obj := range dataset.Elements {
		switch v := obj.(type) {
		case *cimgostructs.RegulatingControl:
			check(id, "RegulatingControl", v.Mode, v.MinAllowedTargetValue, v.MaxAllowedTargetValue)
		case *cimgostructs.TapChangerControl:
			check(id, "TapChangerControl", v.Mode, v.MinAllowedTargetValue, v.MaxAllowedTargetValue)
		}
	}

	return violations
}

// CheckTapChangerStepInteger implements sshcns.TapChanger.step-valueType
// Description: For a discrete TapChangerControl the step value shall be integer.
func CheckTapChangerStepInteger(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	tccDiscrete := make(map[string]bool)
	for id, obj := range dataset.Elements {
		if tcc, ok := obj.(*cimgostructs.TapChangerControl); ok {
			tccDiscrete[id] = tcc.Discrete
		}
	}

	report := func(id, class string, step float64, tcc *struct {
		MRID string `xml:"resource,attr"`
	}) {
		if tcc == nil {
			return
		}
		tccID := strings.TrimPrefix(tcc.MRID, "#")
		if !tccDiscrete[tccID] {
			return
		}
		if step != float64(int(step)) {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    class,
				Property: "TapChanger.step",
				Message:  fmt.Sprintf("Non-integer value (%v) for a discrete TapChangerControl.", step),
				Severity: "sh.Violation",
			})
		}
	}

	for id, obj := range dataset.Elements {
		switch v := obj.(type) {
		case *cimgostructs.RatioTapChanger:
			report(id, "RatioTapChanger", v.Step, v.TapChangerControl)
		case *cimgostructs.PhaseTapChangerLinear:
			report(id, "PhaseTapChangerLinear", v.Step, v.TapChangerControl)
		case *cimgostructs.PhaseTapChangerSymmetrical:
			report(id, "PhaseTapChangerSymmetrical", v.Step, v.TapChangerControl)
		case *cimgostructs.PhaseTapChangerAsymmetrical:
			report(id, "PhaseTapChangerAsymmetrical", v.Step, v.TapChangerControl)
		case *cimgostructs.PhaseTapChangerTabular:
			report(id, "PhaseTapChangerTabular", v.Step, v.TapChangerControl)
		}
	}

	return violations
}
