package validation

import (
	"cimgo/cimgostructs"
	"fmt"
	"strings"
)

// ValidateSSHNotSolvedMASProfileSPARQL runs hand-written checks for
// 61970-301_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS-SHACL and
// 61970-456_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS-SHACL.
func ValidateSSHNotSolvedMASProfileSPARQL(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckLinearShuntCompensatorSectionsRange(dataset)...)
	violations = append(violations, CheckNonlinearShuntCompensatorSectionsValid(dataset)...)
	violations = append(violations, CheckShuntCompensatorSectionsInteger(dataset)...)
	violations = append(violations, CheckRegulatingControlPowerFactorRequiredAttrs(dataset)...)
	violations = append(violations, CheckTapChangerStepInteger(dataset)...)
	violations = append(violations, CheckCsConverterTargetAlphaApplicability(dataset)...)
	violations = append(violations, CheckCsConverterTargetGammaApplicability(dataset)...)
	violations = append(violations, CheckControlAreaNetInterchangeCalculation(dataset)...)
	violations = append(violations, CheckEquivalentInjectionRegulation(dataset)...)
	violations = append(violations, CheckRotatingMachinePLimits(dataset)...)
	violations = append(violations, CheckRotatingMachineQLimits(dataset)...)
	violations = append(violations, CheckSynchronousMachineOperatingModeMatch(dataset)...)
	violations = append(violations, CheckGeneratingUnitSingleActivePowerSlack(dataset)...)
	violations = append(violations, CheckExternalNetworkInjectionLimits(dataset)...)
	violations = append(violations, CheckEquivalentInjectionLimits(dataset)...)
	violations = append(violations, CheckRotatingMachineCurveLimits(dataset)...)
	violations = append(violations, CheckRegulatingControlTargetValuePositive(dataset)...)
	return violations
}

// CheckLinearShuntCompensatorSectionsRange implements sshcns.ShuntCompensator.sections-valueLinear
// Profile: 61970-301_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
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
				Severity: "sh:Violation",
			})
		}
	}

	return violations
}

// CheckNonlinearShuntCompensatorSectionsValid implements sshcns.ShuntCompensator.sections-valueNonLinear
// Profile: 61970-301_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
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
				Severity: "sh:Violation",
			})
		}
	}

	return violations
}

// CheckRegulatingControlPowerFactorRequiredAttrs implements sshcns.RegulatingControl-requiredAttributes
// Profile: 61970-301_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
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
				Severity: "sh:Violation",
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
// Profile: 61970-301_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
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
				Severity: "sh:Violation",
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

// CheckCsConverterTargetAlphaApplicability implements sshn301.CsConverter.targetAlpha-applicability
// Profile: 61970-301_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: targetAlpha must not be set for inverters; it is only valid for rectifiers
// with continuous (non-discrete) tap changer control at the PCC terminal transformer.
func CheckCsConverterTargetAlphaApplicability(dataset *cimgostructs.CIMElementList) []Violation {
	return checkCsConverterTargetAngleApplicability(dataset, true)
}

// CheckCsConverterTargetGammaApplicability implements sshn301.CsConverter.targetGamma-applicability
// Profile: 61970-301_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: targetGamma must not be set for rectifiers; it is only valid for inverters
// with continuous (non-discrete) tap changer control at the PCC terminal transformer.
func CheckCsConverterTargetGammaApplicability(dataset *cimgostructs.CIMElementList) []Violation {
	return checkCsConverterTargetAngleApplicability(dataset, false)
}

// checkCsConverterTargetAngleApplicability is the shared implementation for the alpha/gamma
// applicability checks. forAlpha=true checks targetAlpha (only valid for rectifiers),
// forAlpha=false checks targetGamma (only valid for inverters).
func checkCsConverterTargetAngleApplicability(dataset *cimgostructs.CIMElementList, forAlpha bool) []Violation {
	// Build index: terminalID → RegulatingControl.discrete
	rcDiscrete := make(map[string]bool)
	for _, obj := range dataset.Elements {
		switch v := obj.(type) {
		case *cimgostructs.RegulatingControl:
			if v.Terminal != nil {
				id := strings.TrimPrefix(v.Terminal.MRID, "#")
				rcDiscrete[id] = v.Discrete
			}
		case *cimgostructs.TapChangerControl:
			if v.Terminal != nil {
				id := strings.TrimPrefix(v.Terminal.MRID, "#")
				rcDiscrete[id] = v.Discrete
			}
		}
	}

	var violations []Violation
	for id, obj := range dataset.Elements {
		cs, ok := obj.(*cimgostructs.CsConverter)
		if !ok {
			continue
		}

		var value float64
		var property, msg string
		if forAlpha {
			value = cs.TargetAlpha
			property = "CsConverter.targetAlpha"
			msg = "CsConverter.targetAlpha is provided for an inverter or discrete tap changer control is used or RegulatingControl is not provided."
		} else {
			value = cs.TargetGamma
			property = "CsConverter.targetGamma"
			msg = "CsConverter.targetGamma is provided for a rectifier or discrete tap changer control is used or RegulatingControl is not provided."
		}
		if value == 0 || cs.OperatingMode == nil {
			continue
		}

		mode := cs.OperatingMode.URI
		invalidMode := cimgostructs.CsOperatingModeKindinverter
		if !forAlpha {
			invalidMode = cimgostructs.CsOperatingModeKindrectifier
		}
		if mode == invalidMode {
			violations = append(violations, Violation{ObjectID: id, Class: "CsConverter", Property: property, Message: msg, Severity: "sh:Violation"})
			continue
		}

		// Check OPTIONAL: PccTerminal → PowerTransformer → RegulatingControl
		if cs.PccTerminal == nil {
			violations = append(violations, Violation{ObjectID: id, Class: "CsConverter", Property: property, Message: msg, Severity: "sh:Violation"})
			continue
		}
		pccTermID := strings.TrimPrefix(cs.PccTerminal.MRID, "#")
		pccTerm, hasTerm := dataset.Terminals[pccTermID]
		if !hasTerm || pccTerm.ConductingEquipment == nil {
			violations = append(violations, Violation{ObjectID: id, Class: "CsConverter", Property: property, Message: msg, Severity: "sh:Violation"})
			continue
		}
		eqID := strings.TrimPrefix(pccTerm.ConductingEquipment.MRID, "#")
		if _, isPT := dataset.Elements[eqID].(*cimgostructs.PowerTransformer); !isPT {
			violations = append(violations, Violation{ObjectID: id, Class: "CsConverter", Property: property, Message: msg, Severity: "sh:Violation"})
			continue
		}
		discrete, hasRC := rcDiscrete[pccTermID]
		if !hasRC || discrete {
			violations = append(violations, Violation{ObjectID: id, Class: "CsConverter", Property: property, Message: msg, Severity: "sh:Violation"})
		}
	}
	return violations
}

// CheckControlAreaNetInterchangeCalculation implements sshn301.ControlArea-netInterchangeCalculation
// Profile: 61970-301_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: For ControlArea of type Interchange, the netInterchange value must equal the
// sum of EquivalentInjection.p values for EquivalentInjections connected to BoundaryPoint terminals.
func CheckControlAreaNetInterchangeCalculation(dataset *cimgostructs.CIMElementList) []Violation {
	// Build index: connectivityNodeID → true if a BoundaryPoint references it
	cnHasBoundaryPoint := make(map[string]bool)
	for _, obj := range dataset.Elements {
		bp, ok := obj.(*cimgostructs.BoundaryPoint)
		if !ok || bp.ConnectivityNode == nil {
			continue
		}
		cnHasBoundaryPoint[strings.TrimPrefix(bp.ConnectivityNode.MRID, "#")] = true
	}

	// Build index: controlAreaID → []terminalIDs from TieFlows
	caTerminals := make(map[string][]string)
	for _, obj := range dataset.Elements {
		tf, ok := obj.(*cimgostructs.TieFlow)
		if !ok || tf.ControlArea == nil || tf.Terminal == nil {
			continue
		}
		caID := strings.TrimPrefix(tf.ControlArea.MRID, "#")
		termID := strings.TrimPrefix(tf.Terminal.MRID, "#")
		caTerminals[caID] = append(caTerminals[caID], termID)
	}

	var violations []Violation
	for id, obj := range dataset.Elements {
		ca, ok := obj.(*cimgostructs.ControlArea)
		if !ok || ca.Type == nil || ca.Type.URI != cimgostructs.ControlAreaTypeKindInterchange || ca.NetInterchange == 0 {
			continue
		}

		var sum float64
		for _, termID := range caTerminals[id] {
			term, ok := dataset.Terminals[termID]
			if !ok || term.ConnectivityNode == nil || term.ConductingEquipment == nil {
				continue
			}
			cnID := strings.TrimPrefix(term.ConnectivityNode.MRID, "#")
			if !cnHasBoundaryPoint[cnID] {
				continue
			}
			eqID := strings.TrimPrefix(term.ConductingEquipment.MRID, "#")
			ei, ok := dataset.Elements[eqID].(*cimgostructs.EquivalentInjection)
			if !ok {
				continue
			}
			sum += ei.P
		}

		if ca.NetInterchange != sum {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "ControlArea",
				Property: "ControlArea.netInterchange",
				Message:  fmt.Sprintf("The sum of the EquivalentInjections which are connected to the BoundaryPoint-s differs from the ControlArea.netInterchange. ControlArea.netInterchange= %v. Sum of the EquivalentInjections= %v.", ca.NetInterchange, sum),
				Severity: "sh:Violation",
			})
		}
	}
	return violations
}

// CheckEquivalentInjectionRegulation implements sshn456:EquivalentInjection-regulation
// Profile: 61970-456_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
func CheckEquivalentInjectionRegulation(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	for id, ei := range dataset.EquivalentInjections {
		if ei.RegulationCapability {
			if !ei.RegulationStatus || ei.RegulationTarget == 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "EquivalentInjection",
					Property: "regulationStatus",
					Message:  "EquivalentInjection.regulationStatus and regulationTarget are required when regulationCapability is true.",
					Severity: "sh:Violation",
				})
			}
		} else {
			if ei.RegulationStatus || ei.RegulationTarget != 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "EquivalentInjection",
					Property: "regulationStatus",
					Message:  "EquivalentInjection.regulationStatus and regulationTarget should not be exchanged when regulationCapability is false.",
					Severity: "sh:Violation",
				})
			}
		}
	}
	return violations
}

// CheckRotatingMachinePLimits implements sshn456:RotatingMachine.p-limits
// Profile: 61970-456_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
func CheckRotatingMachinePLimits(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	for id, obj := range dataset.Elements {
		var p float64
		var guRef *struct {
			MRID string `xml:"resource,attr"`
		}

		if sm, ok := obj.(*cimgostructs.SynchronousMachine); ok {
			p, guRef = sm.P, sm.GeneratingUnit
		} else if am, ok := obj.(*cimgostructs.AsynchronousMachine); ok {
			p, guRef = am.P, am.GeneratingUnit
		} else {
			continue
		}

		if guRef == nil {
			continue
		}
		guID := strings.TrimPrefix(guRef.MRID, "#")
		gu, ok := dataset.GeneratingUnits[guID]
		if !ok {
			continue
		}

		negP := -p
		if p == 0 {
			negP = 0
		}

		if negP < gu.MinOperatingP || negP > gu.MaxOperatingP {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    goTypeName(obj),
				Property: "RotatingMachine.p",
				Message:  fmt.Sprintf("Negated active power (%v) is outside of the range [Min:%v, Max:%v] of associated GeneratingUnit.", negP, gu.MinOperatingP, gu.MaxOperatingP),
				Severity: "sh:Violation",
			})
		}
	}
	return violations
}

// CheckRotatingMachineQLimits implements sshn456:RotatingMachine.q-limits
// Profile: 61970-456_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
func CheckRotatingMachineQLimits(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	for id, sm := range dataset.SynchronousMachines {
		if !sm.InService || sm.InitialReactiveCapabilityCurve != nil {
			continue
		}

		negQ := -sm.Q
		if sm.Q == 0 {
			negQ = 0
		}

		if negQ < sm.MinQ || negQ > sm.MaxQ {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "SynchronousMachine",
				Property: "RotatingMachine.q",
				Message:  fmt.Sprintf("Negated reactive power (%v) is outside of the range [Min:%v, Max:%v] (no ReactiveCapabilityCurve).", negQ, sm.MinQ, sm.MaxQ),
				Severity: "sh:Violation",
			})
		}
	}
	return violations
}

// CheckSynchronousMachineOperatingModeMatch implements sshn456:SynchronousMachine.operatingMode-matchType
// Profile: 61970-456_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
func CheckSynchronousMachineOperatingModeMatch(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	for id, sm := range dataset.SynchronousMachines {
		if sm.OperatingMode == nil || sm.Type == nil {
			continue
		}
		mode := sm.OperatingMode.URI
		kind := sm.Type.URI

		valid := false
		switch {
		case strings.HasSuffix(mode, "motor"):
			valid = strings.HasSuffix(kind, "motor") || strings.HasSuffix(kind, "generatorOrMotor") || strings.HasSuffix(kind, "motorOrCondenser") || strings.HasSuffix(kind, "generatorOrCondenserOrMotor")
		case strings.HasSuffix(mode, "condenser"):
			valid = strings.HasSuffix(kind, "condenser") || strings.HasSuffix(kind, "generatorOrCondenser") || strings.HasSuffix(kind, "motorOrCondenser") || strings.HasSuffix(kind, "generatorOrCondenserOrMotor")
		case strings.HasSuffix(mode, "generator"):
			valid = strings.HasSuffix(kind, "generator") || strings.HasSuffix(kind, "generatorOrMotor") || strings.HasSuffix(kind, "generatorOrCondenser") || strings.HasSuffix(kind, "generatorOrCondenserOrMotor")
		}

		if !valid {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "SynchronousMachine",
				Property: "SynchronousMachine.operatingMode",
				Message:  fmt.Sprintf("SynchronousMachine.operatingMode (%v) is not consistent with SynchronousMachine.type (%v).", mode, kind),
				Severity: "sh:Violation",
			})
		}
	}
	return violations
}

// CheckGeneratingUnitSingleActivePowerSlack implements sshn456:GeneratingUnit-singleActivePowerSlack
// Profile: 61970-456_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
func CheckGeneratingUnitSingleActivePowerSlack(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	// Rule: one generator has GeneratingUnit.normalPF set to a highest value (non-zero) and all other generating units have a zero GeneratingUnit.normalPF.
	// Actually, this is per ControlArea.

	caSlacks := make(map[string][]string) // ControlArea ID -> []GeneratingUnit ID
	for _, cagu := range dataset.ControlAreaGeneratingUnits {
		if cagu.ControlArea == nil || cagu.GeneratingUnit == nil {
			continue
		}
		caID := strings.TrimPrefix(cagu.ControlArea.MRID, "#")
		guID := strings.TrimPrefix(cagu.GeneratingUnit.MRID, "#")

		gu, ok := dataset.GeneratingUnits[guID]
		if ok && gu.NormalPF > 0 {
			caSlacks[caID] = append(caSlacks[caID], guID)
		}
	}

	for caID, slacks := range caSlacks {
		if len(slacks) > 1 {
			violations = append(violations, Violation{
				ObjectID: caID,
				Class:    "ControlArea",
				Property: "rdf:type",
				Message:  fmt.Sprintf("Multiple generating units (%v) in ControlArea %s have non-zero normalPF.", slacks, caID),
				Severity: "sh:Violation",
			})
		}
	}

	return violations
}

// CheckExternalNetworkInjectionLimits implements sshn456:ExternalNetworkInjection.p-limits and q-limits
// Profile: 61970-456_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
func CheckExternalNetworkInjectionLimits(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	for id, eni := range dataset.ExternalNetworkInjections {
		if !eni.InService {
			continue
		}
		negP := -eni.P
		if eni.P == 0 {
			negP = 0
		}
		if negP < eni.MinP || negP > eni.MaxP {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "ExternalNetworkInjection",
				Property: "p",
				Message:  fmt.Sprintf("Negated active power (%v) is outside of the range [Min:%v, Max:%v].", negP, eni.MinP, eni.MaxP),
				Severity: "sh:Violation",
			})
		}
		negQ := -eni.Q
		if eni.Q == 0 {
			negQ = 0
		}
		if negQ < eni.MinQ || negQ > eni.MaxQ {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "ExternalNetworkInjection",
				Property: "q",
				Message:  fmt.Sprintf("Negated reactive power (%v) is outside of the range [Min:%v, Max:%v].", negQ, eni.MinQ, eni.MaxQ),
				Severity: "sh:Violation",
			})
		}
	}
	return violations
}

// CheckEquivalentInjectionLimits implements sshn456:EquivalentInjection.p-limits and q-limits
// Profile: 61970-456_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
func CheckEquivalentInjectionLimits(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	for id, ei := range dataset.EquivalentInjections {
		if !ei.InService {
			continue
		}
		negP := -ei.P
		if ei.P == 0 {
			negP = 0
		}
		if negP < ei.MinP || negP > ei.MaxP {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "EquivalentInjection",
				Property: "p",
				Message:  fmt.Sprintf("Negated active power (%v) is outside of the range [Min:%v, Max:%v].", negP, ei.MinP, ei.MaxP),
				Severity: "sh:Violation",
			})
		}
		negQ := -ei.Q
		if ei.Q == 0 {
			negQ = 0
		}
		if negQ < ei.MinQ || negQ > ei.MaxQ {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "EquivalentInjection",
				Property: "q",
				Message:  fmt.Sprintf("Negated reactive power (%v) is outside of the range [Min:%v, Max:%v].", negQ, ei.MinQ, ei.MaxQ),
				Severity: "sh:Violation",
			})
		}
	}
	return violations
}

// CheckRotatingMachineCurveLimits implements sshn456:RotatingMachine-pAndQcapabilityCurveP/Q
// Profile: 61970-456_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
func CheckRotatingMachineCurveLimits(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, sm := range dataset.SynchronousMachines {
		if !sm.InService || sm.InitialReactiveCapabilityCurve == nil {
			continue
		}

		rccID := strings.TrimPrefix(sm.InitialReactiveCapabilityCurve.MRID, "#")
		var xvals []float64
		var y1vals []float64
		var y2vals []float64

		for _, cdObj := range dataset.Elements {
			if cd, ok := cdObj.(*cimgostructs.CurveData); ok && cd.Curve != nil {
				if strings.TrimPrefix(cd.Curve.MRID, "#") == rccID {
					xvals = append(xvals, cd.Xvalue)
					y1vals = append(y1vals, cd.Y1value)
					y2vals = append(y2vals, cd.Y2value)
				}
			}
		}

		if len(xvals) == 0 {
			continue
		}

		minX, maxX := xvals[0], xvals[0]
		minY1, maxY2 := y1vals[0], y2vals[0]
		for i, x := range xvals {
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y1vals[i] < minY1 {
				minY1 = y1vals[i]
			}
			if y2vals[i] > maxY2 {
				maxY2 = y2vals[i]
			}
		}

		negP := -sm.P
		if sm.P == 0 {
			negP = 0
		}
		negQ := -sm.Q
		if sm.Q == 0 {
			negQ = 0
		}

		if negP < minX || negP > maxX {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "SynchronousMachine",
				Property: "RotatingMachine.p",
				Message:  fmt.Sprintf("Negated active power (%v) is outside of curve x-range [%v, %v].", negP, minX, maxX),
				Severity: "sh:Violation",
			})
		}
		if negQ < minY1 || negQ > maxY2 {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "SynchronousMachine",
				Property: "RotatingMachine.q",
				Message:  fmt.Sprintf("Negated reactive power (%v) is outside of curve y-range [%v, %v].", negQ, minY1, maxY2),
				Severity: "sh:Violation",
			})
		}
	}

	return violations
}

// CheckRegulatingControlTargetValuePositive implements sshn456:RegulatingControl.targetValue-value
// Profile: 61970-456_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
func CheckRegulatingControlTargetValuePositive(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	for id, rc := range dataset.RegulatingControls {
		if rc.Mode != nil && strings.HasSuffix(rc.Mode.URI, "voltage") {
			if rc.TargetValue <= 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "RegulatingControl",
					Property: "targetValue",
					Message:  "RegulatingControl.targetValue shall be positive value in cases where the RegulatingControl.mode is set to voltage.",
					Severity: "sh:Violation",
				})
			}
		}
	}
	return violations
}

// CheckShuntCompensatorSectionsInteger implements sshc456ns:ShuntCompensator.sections-value
// Profile: 61970-456_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: In cases where RegulatingControl.discrete is true and RegulatingControl.enabled is true, ShuntCompensator.sections shall be integer.
func CheckShuntCompensatorSectionsInteger(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		var sections float64
		var rcID string
		var class string

		if lsc, ok := obj.(*cimgostructs.LinearShuntCompensator); ok {
			sections = lsc.Sections
			class = "LinearShuntCompensator"
			if lsc.RegulatingControl != nil {
				rcID = strings.TrimPrefix(lsc.RegulatingControl.MRID, "#")
			}
		} else if nsc, ok := obj.(*cimgostructs.NonlinearShuntCompensator); ok {
			sections = nsc.Sections
			class = "NonlinearShuntCompensator"
			if nsc.RegulatingControl != nil {
				rcID = strings.TrimPrefix(nsc.RegulatingControl.MRID, "#")
			}
		} else {
			continue
		}

		if rcID != "" {
			if rc, ok := dataset.RegulatingControls[rcID]; ok && rc.Enabled && rc.Discrete {
				if sections != float64(int(sections)) {
					violations = append(violations, Violation{
						ObjectID: id,
						Class:    class,
						Property: "ShuntCompensator.sections",
						Message:  fmt.Sprintf("The value (%v) is not integer for an active discrete regulating control.", sections),
						Severity: "sh:Violation",
					})
				}
			}
		}
	}

	return violations
}
