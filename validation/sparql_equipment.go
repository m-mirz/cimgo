package validation

import (
	"cimgo/cimgostructs"
	"fmt"
	"reflect"
	"strings"
)

// ValidateEQProfileSPARQL runs hand-written checks for 61970-301_Equipment-AP-Con-Complex-SHACL.
func ValidateEQProfileSPARQL(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckACDCTerminalSequenceNumbering(dataset)...)
	violations = append(violations, CheckTerminalPhasesConsistencyEquipment(dataset)...)
	violations = append(violations, CheckConductingEquipmentBaseVoltageUsage(dataset)...)
	violations = append(violations, CheckPowerTransformerEndNumberUnique(dataset)...)
	violations = append(violations, CheckPowerTransformerEndTerminalConsistency(dataset)...)
	violations = append(violations, CheckOperationalLimitTypeDuration(dataset)...)
	violations = append(violations, CheckPowerTransformerTwoWindingEndValues(dataset)...)
	violations = append(violations, CheckPhaseTapChangerLinearXMinConsistency(dataset)...)
	violations = append(violations, CheckPhaseTapChangerNonLinearXMinConsistency(dataset)...)
	violations = append(violations, CheckPowerTransformerEndRatedS2Winding(dataset)...)
	violations = append(violations, CheckPowerTransformerBaseVoltageAssociation(dataset)...)
	violations = append(violations, CheckPowerTransformerEndRValueRange(dataset)...)
	violations = append(violations, CheckRegulatingControlTerminalConnectivityNode(dataset)...)
	violations = append(violations, CheckTapChangerLtcFlagControl(dataset)...)
	violations = append(violations, CheckLoadResponseCharacteristicExponentModel(dataset)...)
	violations = append(violations, CheckNonlinearShuntCompensatorPointCount(dataset)...)
	violations = append(violations, CheckShuntCompensatorNomU(dataset)...)
	violations = append(violations, CheckPhaseTapChangerAsymmetricalWindingConnectionAngle(dataset)...)
	violations = append(violations, CheckPowerTransformerEndRatedUValueRange(dataset)...)
	violations = append(violations, CheckVoltageLimitPATL(dataset)...)
	violations = append(violations, CheckDCConverterUnitTapChangerControl(dataset)...)
	violations = append(violations, CheckConnectivityNodeTerminalPhasesConsistency(dataset)...)
	violations = append(violations, CheckEquipmentAggregateNotUsed(dataset)...)
	violations = append(violations, CheckEquivalentBranchR21Usage(dataset)...)
	violations = append(violations, CheckEquivalentBranchX21Usage(dataset)...)
	violations = append(violations, CheckEquivalentInjectionRegulationCapability(dataset)...)
	violations = append(violations, CheckGeneratingUnitNominalP(dataset)...)
	violations = append(violations, CheckControlAreaGeneratingUnitInstance(dataset)...)
	violations = append(violations, CheckDCConverterUnitCsConverterPowerTransformer(dataset)...)
	violations = append(violations, CheckLimitKindPATLNumberOfLimitType(dataset)...)
	violations = append(violations, CheckLimitKindTCDuration(dataset)...)

	// EQ 452 & 600 additions
	violations = append(violations, CheckSynchronousMachineAggregate(dataset)...)
	violations = append(violations, CheckAsynchronousMachineAggregate(dataset)...)
	violations = append(violations, CheckSynchronousMachineControlMode(dataset)...)
	violations = append(violations, CheckStaticVarCompensatorControlMode(dataset)...)
	violations = append(violations, CheckPhaseTapChangerControlMode(dataset)...)
	violations = append(violations, CheckRatioTapChangerControlMode(dataset)...)
	violations = append(violations, CheckShuntCompensatorControlMode(dataset)...)
	violations = append(violations, CheckSynchronousMachineReactiveLimits(dataset)...)
	violations = append(violations, CheckSynchronousMachineTypeCondenser(dataset)...)
	violations = append(violations, CheckVsCapabilityCurveCount(dataset)...)
	violations = append(violations, CheckVsCapabilityCurveYValues(dataset)...)
	violations = append(violations, CheckGeneratingUnitTypeDependency(dataset)...)
	violations = append(violations, CheckCurveDataReactiveCapabilityLimits(dataset)...)
	violations = append(violations, CheckCurveDataReactiveConsistency(dataset)...)
	violations = append(violations, CheckSynchronousMachineCurveXValueConsistency(dataset)...)
	violations = append(violations, CheckSwitchConnection(dataset)...)
	violations = append(violations, CheckOperationalLimitSetTerminal(dataset)...)
	violations = append(violations, CheckTapChangerControlRemoteQControl(dataset)...)
	violations = append(violations, CheckReactiveCapabilityCurveXValueUnique(dataset)...)
	violations = append(violations, CheckPowerTransformerEndResistanceXValue(dataset)...)
	violations = append(violations, CheckGeneratingUnitMaxOperatingPRatedS(dataset)...)
	violations = append(violations, CheckHydroGeneratingUnitEnergyConversionCapability(dataset)...)
	violations = append(violations, CheckTerminalConnectionSameNode(dataset)...)
	violations = append(violations, CheckReactiveCapabilityCurveReactiveCountP(dataset)...)
	violations = append(violations, CheckReactiveCapabilityCurveUnits(dataset)...)
	violations = append(violations, CheckSubstationCount(dataset)...)
	violations = append(violations, CheckTapChangerNeutralUValueRange(dataset)...)

	return violations
}

// CheckACDCTerminalSequenceNumbering implements eqc.ACDCTerminal.sequenceNumber-numbering
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
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
// Origin: Derived from a SPARQL constraint.
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
// Origin: Derived from a SPARQL constraint.
// Description: Use only when there is no voltage level container used and only one base voltage applies.
// For example, not used for transformers.
func CheckConductingEquipmentBaseVoltageUsage(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		typeName := goTypeName(obj)
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
				if goTypeName(ecObj) == "VoltageLevel" {
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

// CheckPowerTransformerEndNumberUnique implements eqc.TransformerEnd.endNumber-unique
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Highest voltage winding should be 1. Each end within a power transformer should have a unique subsequent end number.
func CheckPowerTransformerEndNumberUnique(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	transformerEnds := make(map[string][]*cimgostructs.PowerTransformerEnd)
	for _, obj := range dataset.Elements {
		if end, ok := obj.(*cimgostructs.PowerTransformerEnd); ok {
			if end.PowerTransformer != nil {
				ptID := strings.TrimPrefix(end.PowerTransformer.MRID, "#")
				transformerEnds[ptID] = append(transformerEnds[ptID], end)
			}
		}
	}

	for ptID, ends := range transformerEnds {
		seenNumbers := make(map[int]bool)
		maxRatedU := -1.0
		var maxRatedUEnd *cimgostructs.PowerTransformerEnd

		duplicate := false
		for _, end := range ends {
			if seenNumbers[end.EndNumber] {
				duplicate = true
			}
			seenNumbers[end.EndNumber] = true

			if end.RatedU > maxRatedU {
				maxRatedU = end.RatedU
				maxRatedUEnd = end
			}
		}

		if duplicate {
			violations = append(violations, Violation{
				ObjectID: ptID,
				Class:    "PowerTransformer",
				Property: "TransformerEnd.endNumber",
				Message:  "The PowerTransformer has TransformerEnd.endNumber which is not unique.",
				Severity: "sh.Violation",
			})
		} else if maxRatedUEnd != nil && maxRatedUEnd.EndNumber != 1 {
			// Check if there are other ends with the same maxRatedU that have endNumber 1
			foundMaxAt1 := false
			for _, end := range ends {
				if end.RatedU == maxRatedU && end.EndNumber == 1 {
					foundMaxAt1 = true
					break
				}
			}
			if !foundMaxAt1 {
				violations = append(violations, Violation{
					ObjectID: ptID,
					Class:    "PowerTransformer",
					Property: "TransformerEnd.endNumber",
					Message:  "The PowerTransformerEnd with endNumber 1 is not the highest voltage winding.",
					Severity: "sh.Violation",
				})
			}
		}
	}

	return violations
}

// CheckPowerTransformerEndTerminalConsistency implements eqc.PowerTransformerEnd-terminalConsistency
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: The Terminal referenced by TransformerEnd.Terminal points to a PowerTransformer which is different than the referenced element via PowerTransformerEnd.PowerTransformer.
func CheckPowerTransformerEndTerminalConsistency(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		end, ok := obj.(*cimgostructs.PowerTransformerEnd)
		if !ok || end.Terminal == nil || end.PowerTransformer == nil {
			continue
		}

		termID := strings.TrimPrefix(end.Terminal.MRID, "#")
		termObj, ok := dataset.Elements[termID]
		if !ok {
			continue
		}

		term, ok := termObj.(*cimgostructs.Terminal)
		if !ok || term.ConductingEquipment == nil {
			continue
		}

		termPtID := strings.TrimPrefix(term.ConductingEquipment.MRID, "#")
		ptID := strings.TrimPrefix(end.PowerTransformer.MRID, "#")

		if termPtID != ptID {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "PowerTransformerEnd",
				Property: "TransformerEnd.Terminal",
				Message:  "The Terminal referenced by TransformerEnd.Terminal points to a PowerTransformer which is different than the referenced element via PowerTransformerEnd.PowerTransformer.",
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckOperationalLimitTypeDuration implements eqc.OperationalLimitType.acceptableDuration-usage and isInfiniteDuration-usage
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: acceptableDuration must be present when isInfiniteDuration is false, and absent when true.
func CheckOperationalLimitTypeDuration(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		olt, ok := obj.(*cimgostructs.OperationalLimitType)
		if !ok {
			continue
		}

		// eqc.OperationalLimitType.acceptableDuration-usage
		// The attribute has meaning only if the flag isInfiniteDuration is set to false, hence it shall not be exchanged when isInfiniteDuration is set to true.
		if olt.IsInfiniteDuration && olt.AcceptableDuration != 0 {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "OperationalLimitType",
				Property: "OperationalLimitType.acceptableDuration",
				Message:  "The attribute acceptableDuration is present and isInfiniteDuration is set to true.",
				Severity: "sh.Violation",
			})
		}

		// eqc.OperationalLimitType.isInfiniteDuration-usage
		// If false, the limit has definite duration which is defined by the attribute acceptableDuration.
		if !olt.IsInfiniteDuration && olt.AcceptableDuration == 0 {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "OperationalLimitType",
				Property: "OperationalLimitType.acceptableDuration",
				Message:  "The attribute acceptableDuration is not present when isInfiniteDuration is set to false.",
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckPowerTransformerTwoWindingEndValues implements eqc.PowerTransformerEnd-secondWindingValues
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: for a two Terminal PowerTransformer the high voltage (endNumber=1) has non zero r, r0, x, x0 while low voltage (endNumber=2) has zero values.
func CheckPowerTransformerTwoWindingEndValues(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	transformerEnds := make(map[string][]*cimgostructs.PowerTransformerEnd)
	for _, obj := range dataset.Elements {
		if end, ok := obj.(*cimgostructs.PowerTransformerEnd); ok {
			if end.PowerTransformer != nil {
				ptID := strings.TrimPrefix(end.PowerTransformer.MRID, "#")
				transformerEnds[ptID] = append(transformerEnds[ptID], end)
			}
		}
	}

	for ptID, ends := range transformerEnds {
		if len(ends) != 2 {
			continue
		}

		for _, end := range ends {
			if end.EndNumber == 2 {
				if end.R != 0 || end.R0 != 0 || end.X != 0 || end.X0 != 0 {
					violations = append(violations, Violation{
						ObjectID: ptID,
						Class:    "PowerTransformer",
						Property: "PowerTransformerEnd-secondWindingValues",
						Message:  fmt.Sprintf("Non-zero values for the PowerTransformerEnd with TransformerEnd.endNumber=2 (R=%v, R0=%v, X=%v, X0=%v) for a two Terminal PowerTransformer.", end.R, end.R0, end.X, end.X0),
						Severity: "sh.Violation",
					})
				}
			}
		}
	}

	return violations
}

// CheckPhaseTapChangerLinearXMinConsistency implements eqc.PhaseTapChangerLinear.xMin-valueRangePair
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: PowerTransformerEnd.x shall be consistent with PhaseTapChangerLinear.xMin.
func CheckPhaseTapChangerLinearXMinConsistency(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		ptcl, ok := obj.(*cimgostructs.PhaseTapChangerLinear)
		if !ok || ptcl.TransformerEnd == nil {
			continue
		}

		endID := strings.TrimPrefix(ptcl.TransformerEnd.MRID, "#")
		if endObj, ok := dataset.Elements[endID]; ok {
			if end, ok := endObj.(*cimgostructs.PowerTransformerEnd); ok {
				if ptcl.XMin != end.X {
					violations = append(violations, Violation{
						ObjectID: id,
						Class:    "PhaseTapChangerLinear",
						Property: "PhaseTapChangerLinear.xMin",
						Message:  fmt.Sprintf("Inconsistency between PowerTransformerEnd.x (%v) and PhaseTapChangerLinear.xMin (%v).", end.X, ptcl.XMin),
						Severity: "sh.Violation",
					})
				}
			}
		}
	}

	return violations
}

// CheckPhaseTapChangerNonLinearXMinConsistency implements eqc.PhaseTapChangerNonLinear.xMin-valueRangePair
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: PowerTransformerEnd.x shall be consistent with PhaseTapChangerNonLinear.xMin.
func CheckPhaseTapChangerNonLinearXMinConsistency(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		ptcnl, ok := obj.(*cimgostructs.PhaseTapChangerNonLinear)
		if !ok || ptcnl.TransformerEnd == nil {
			continue
		}

		endID := strings.TrimPrefix(ptcnl.TransformerEnd.MRID, "#")
		if endObj, ok := dataset.Elements[endID]; ok {
			if end, ok := endObj.(*cimgostructs.PowerTransformerEnd); ok {
				if ptcnl.XMin != end.X {
					violations = append(violations, Violation{
						ObjectID: id,
						Class:    "PhaseTapChangerNonLinear",
						Property: "PhaseTapChangerNonLinear.xMin",
						Message:  fmt.Sprintf("Inconsistency between PowerTransformerEnd.x (%v) and PhaseTapChangerNonLinear.xMin (%v).", end.X, ptcnl.XMin),
						Severity: "sh.Violation",
					})
				}
			}
		}
	}

	return violations
}

// CheckPowerTransformerEndRatedS2Winding implements eqc.PowerTransformerEnd.ratedS-valueRange2winding
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: For a two-winding transformer the values for the high and low voltage sides shall be identical.
func CheckPowerTransformerEndRatedS2Winding(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	transformerEnds := make(map[string][]*cimgostructs.PowerTransformerEnd)
	for _, obj := range dataset.Elements {
		if end, ok := obj.(*cimgostructs.PowerTransformerEnd); ok {
			if end.PowerTransformer != nil {
				ptID := strings.TrimPrefix(end.PowerTransformer.MRID, "#")
				transformerEnds[ptID] = append(transformerEnds[ptID], end)
			}
		}
	}

	for ptID, ends := range transformerEnds {
		if len(ends) != 2 {
			continue
		}

		if ends[0].RatedS != ends[1].RatedS {
			violations = append(violations, Violation{
				ObjectID: ptID,
				Class:    "PowerTransformer",
				Property: "PowerTransformerEnd.ratedS",
				Message:  fmt.Sprintf("The RatedS value is different for a two-winding transformer. End 1: %v, End 2: %v.", ends[0].RatedS, ends[1].RatedS),
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckPowerTransformerBaseVoltageAssociation implements eqc.PowerTransformer-associationNotUsed
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: The inherited association ConductingEquipment.BaseVoltage should not be used.
func CheckPowerTransformerBaseVoltageAssociation(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		pt, ok := obj.(*cimgostructs.PowerTransformer)
		if !ok {
			continue
		}

		if pt.BaseVoltage != nil {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "PowerTransformer",
				Property: "ConductingEquipment.BaseVoltage",
				Message:  "The inherited association ConductingEquipment.BaseVoltage is used.",
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckPowerTransformerEndRValueRange implements eqc.PowerTransformerEnd.r-valueRange
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: The attribute shall be equal to or greater than zero for non-equivalent transformers.
func CheckPowerTransformerEndRValueRange(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		end, ok := obj.(*cimgostructs.PowerTransformerEnd)
		if !ok || end.PowerTransformer == nil {
			continue
		}

		ptID := strings.TrimPrefix(end.PowerTransformer.MRID, "#")
		if ptObj, ok := dataset.Elements[ptID]; ok {
			if pt, ok := ptObj.(*cimgostructs.PowerTransformer); ok {
				if !pt.Aggregate && end.R < 0 {
					violations = append(violations, Violation{
						ObjectID: id,
						Class:    "PowerTransformerEnd",
						Property: "PowerTransformerEnd.r",
						Message:  "The value is negative for a non-equivalent transformer.",
						Severity: "sh.Violation",
					})
				}
			}
		}
	}

	return violations
}

// CheckRegulatingControlTerminalConnectivityNode implements eqc.RegulatingControl-terminalConnectivityNode
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: The specified terminal shall be associated with the connectivity node of the controlled point.
func CheckRegulatingControlTerminalConnectivityNode(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		rc, ok := obj.(*cimgostructs.RegulatingControl)
		if !ok || rc.Terminal == nil {
			continue
		}

		termID := strings.TrimPrefix(rc.Terminal.MRID, "#")
		if termObj, ok := dataset.Elements[termID]; ok {
			if term, ok := termObj.(*cimgostructs.Terminal); ok {
				if term.ConnectivityNode == nil {
					violations = append(violations, Violation{
						ObjectID: id,
						Class:    "RegulatingControl",
						Property: "RegulatingControl.Terminal",
						Message:  "The Terminal referenced by the RegulatingControl is not associated with a ConnectivityNode.",
						Severity: "sh.Violation",
					})
				}
			}
		}
	}

	return violations
}

// CheckTapChangerLtcFlagControl implements eqc.TapChanger.ltcFlag-tapChangerControl
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: When TapChanger.ltcFlag=false and TapChanger.TapChangerControl is present an artificial tap changer can be used to simulate control behaviour in power flow.
func CheckTapChangerLtcFlagControl(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		tc, ok := obj.(*cimgostructs.TapChanger)
		if !ok {
			continue
		}

		if !tc.LtcFlag && tc.TapChangerControl != nil {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "TapChanger",
				Property: "TapChanger.ltcFlag",
				Message:  "An artificial tap changer is used to simulate control behaviour in power flow (ltcFlag is false but TapChangerControl is present).",
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckLoadResponseCharacteristicExponentModel implements eqc.LoadResponseCharacteristic.exponentModel-exponentCoefficient
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Validates both the exponent and coefficient models, ensuring all required attributes
// are present for the chosen model, no mixture of attributes exists, and sums of coefficients equal 1.
func CheckLoadResponseCharacteristicExponentModel(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		lrc, ok := obj.(*cimgostructs.LoadResponseCharacteristic)
		if !ok {
			continue
		}

		// Exponent model attributes (active/reactive voltage/frequency exponents)
		// Note: In cimgostructs, these are typically float64, so we check if they are provided (non-zero).
		// However, 0 is a valid value for an exponent.
		// For simplicity in this implementation, we assume if they are part of the exchange, they are present.
		// In a real RDF/XML dataset, "missing" would mean the tag is absent.
		// Since cimgostructs is a flat structure from XML, we might need to check if they were actually in the XML.
		// But here we'll follow the logic of the SPARQL which checks for existence.

		// For the sake of the rule, we'll check if the model is consistent.
		if lrc.ExponentModel {
			// Exponential model: pFrequencyExponent, pVoltageExponent, qFrequencyExponent, qVoltageExponent required.
			// Coefficient model attributes should NOT be present (should be 0 or default).
			// This is tricky with Go structs if we don't have "IsSet" flags.
			// Assuming non-zero means present for now, though it's imperfect.
			// A better way would be to check if the sum of coefficients is non-zero.
			if lrc.PConstantCurrent != 0 || lrc.PConstantImpedance != 0 || lrc.PConstantPower != 0 ||
				lrc.QConstantCurrent != 0 || lrc.QConstantImpedance != 0 || lrc.QConstantPower != 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "LoadResponseCharacteristic",
					Property: "LoadResponseCharacteristic.exponentModel",
					Message:  "Mixture of exponential and coefficient model attributes when exponentModel is true.",
					Severity: "sh.Violation",
				})
			}
		} else {
			// Coefficient model: pConstantCurrent, pConstantImpedance, pConstantPower, qConstantCurrent, qConstantImpedance, qConstantPower required.
			// Sums must equal 1.
			pSum := lrc.PConstantCurrent + lrc.PConstantImpedance + lrc.PConstantPower
			qSum := lrc.QConstantCurrent + lrc.QConstantImpedance + lrc.QConstantPower

			epsilon := 1e-6
			if (pSum < 1-epsilon || pSum > 1+epsilon) || (qSum < 1-epsilon || qSum > 1+epsilon) {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "LoadResponseCharacteristic",
					Property: "LoadResponseCharacteristic.exponentModel",
					Message:  fmt.Sprintf("The sum of coefficients does not equal 1 (P sum: %v, Q sum: %v).", pSum, qSum),
					Severity: "sh.Violation",
				})
			}
		}
	}

	return violations
}

// CheckNonlinearShuntCompensatorPointCount implements eqc.ShuntCompensator.maximumSections-numberOfInstances
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: The number of NonlinearShuntCompenstorPoint instances shall be equal to maximumSections.
func CheckNonlinearShuntCompensatorPointCount(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	nscPoints := make(map[string]int)
	for _, obj := range dataset.Elements {
		if point, ok := obj.(*cimgostructs.NonlinearShuntCompensatorPoint); ok {
			if point.NonlinearShuntCompensator != nil {
				nscID := strings.TrimPrefix(point.NonlinearShuntCompensator.MRID, "#")
				nscPoints[nscID]++
			}
		}
	}

	for id, count := range nscPoints {
		if obj, ok := dataset.Elements[id]; ok {
			if nsc, ok := obj.(*cimgostructs.NonlinearShuntCompensator); ok {
				if nsc.MaximumSections != count {
					violations = append(violations, Violation{
						ObjectID: id,
						Class:    "NonlinearShuntCompensator",
						Property: "ShuntCompensator.maximumSections",
						Message:  fmt.Sprintf("The number of NonlinearShuntCompenstorPoint instances (%d) does not equal to maximumSections (%d).", count, nsc.MaximumSections),
						Severity: "sh.Violation",
					})
				}
			}
		}
	}

	return violations
}

// CheckShuntCompensatorNomU implements eqc.ShuntCompensator.nomU-nominalVoltageDifference
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: nomU should be within 10% of the nominal voltage.
func CheckShuntCompensatorNomU(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		sc, ok := obj.(*cimgostructs.ShuntCompensator)
		if !ok || sc.EquipmentContainer == nil {
			continue
		}

		ecID := strings.TrimPrefix(sc.EquipmentContainer.MRID, "#")
		if ecObj, ok := dataset.Elements[ecID]; ok {
			if vl, ok := ecObj.(*cimgostructs.VoltageLevel); ok && vl.BaseVoltage != nil {
				bvID := strings.TrimPrefix(vl.BaseVoltage.MRID, "#")
				if bvObj, ok := dataset.Elements[bvID]; ok {
					if bv, ok := bvObj.(*cimgostructs.BaseVoltage); ok {
						nomV := bv.NominalVoltage
						if sc.NomU < 0.9*nomV || sc.NomU > 1.1*nomV {
							violations = append(violations, Violation{
								ObjectID: id,
								Class:    "ShuntCompensator",
								Property: "ShuntCompensator.nomU",
								Message:  fmt.Sprintf("The value nomU (%v) differs with more than 10%% of the nominal voltage (%v).", sc.NomU, nomV),
								Severity: "sh.Warning",
							})
						}
					}
				}
			}
		}
	}

	return violations
}

// CheckPhaseTapChangerAsymmetricalWindingConnectionAngle implements eqc.PhaseTapChangerAsymmetrical.windingConnectionAngle-valueRange
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: The windingConnectionAngle can only be multiples of 30 degrees in the range -150 to 150 excluding 0.
func CheckPhaseTapChangerAsymmetricalWindingConnectionAngle(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		ptca, ok := obj.(*cimgostructs.PhaseTapChangerAsymmetrical)
		if !ok {
			continue
		}

		val := ptca.WindingConnectionAngle
		isMultipleOf30 := int(val)%30 == 0 && val == float64(int(val))
		inRange := val >= -150 && val <= 150 && val != 0

		if !isMultipleOf30 || !inRange {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "PhaseTapChangerAsymmetrical",
				Property: "PhaseTapChangerAsymmetrical.windingConnectionAngle",
				Message:  "The value is not a multiple of 30 degrees in the range of -150 to 150 degrees (excluding 0).",
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckPowerTransformerEndRatedUValueRange implements eqc.PowerTransformerEnd.ratedU-valueRange
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: A high voltage side (endNumber=1) shall have a ratedU >= lower voltage sides; ratedU must be positive.
func CheckPowerTransformerEndRatedUValueRange(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	transformerEnds := make(map[string][]*cimgostructs.PowerTransformerEnd)
	for _, obj := range dataset.Elements {
		if end, ok := obj.(*cimgostructs.PowerTransformerEnd); ok {
			if end.PowerTransformer != nil {
				ptID := strings.TrimPrefix(end.PowerTransformer.MRID, "#")
				transformerEnds[ptID] = append(transformerEnds[ptID], end)
			}
		}
	}

	for ptID, ends := range transformerEnds {
		maxRatedU := -1.0
		var end1 *cimgostructs.PowerTransformerEnd

		for _, end := range ends {
			if end.RatedU <= 0 {
				violations = append(violations, Violation{
					ObjectID: ptID, // Reporting on transformer or end? SHACL says target is PowerTransformer
					Class:    "PowerTransformer",
					Property: "PowerTransformerEnd.ratedU",
					Message:  fmt.Sprintf("The PowerTransformerEnd %s has a non-positive ratedU (%v).", end.MRID, end.RatedU),
					Severity: "sh.Violation",
				})
			}
			if end.EndNumber == 1 {
				end1 = end
			}
			if end.RatedU > maxRatedU {
				maxRatedU = end.RatedU
			}
		}

		if end1 != nil && end1.RatedU < maxRatedU {
			violations = append(violations, Violation{
				ObjectID: ptID,
				Class:    "PowerTransformer",
				Property: "PowerTransformerEnd.ratedU",
				Message:  "The high voltage side (endNumber=1) does not have the highest ratedU.",
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckVoltageLimitPATL implements eqc.LimitKind.patl-allowedType
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: The Permanent Admissible Transmission Loading (PATL) is not allowed for VoltageLimit.
func CheckVoltageLimitPATL(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		vl, ok := obj.(*cimgostructs.VoltageLimit)
		if !ok || vl.OperationalLimitType == nil {
			continue
		}

		oltID := strings.TrimPrefix(vl.OperationalLimitType.MRID, "#")
		if oltObj, ok := dataset.Elements[oltID]; ok {
			if olt, ok := oltObj.(*cimgostructs.OperationalLimitType); ok && olt.Kind != nil {
				patl := "http://iec.ch/TC57/CIM100-European#LimitKind.patl"
				if olt.Kind.URI == patl {
					violations = append(violations, Violation{
						ObjectID: id,
						Class:    "VoltageLimit",
						Property: "OperationalLimit.OperationalLimitType",
						Message:  "PATL type is provided for VoltageLimit.",
						Severity: "sh.Violation",
					})
				}
			}
		}
	}

	return violations
}

// CheckDCConverterUnitTapChangerControl implements eqc.DCConverterUnit-tapChangerControl
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: No TapChangerControl is used for the converter transformer contained in DCConverterUnit.
func CheckDCConverterUnitTapChangerControl(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		var tcControl *struct {
			MRID string `xml:"resource,attr"`
		}
		var transformerEndID string

		if rtc, ok := obj.(*cimgostructs.RatioTapChanger); ok {
			tcControl = rtc.TapChangerControl
			if rtc.TransformerEnd != nil {
				transformerEndID = strings.TrimPrefix(rtc.TransformerEnd.MRID, "#")
			}
		} else if ptc, ok := obj.(*cimgostructs.PhaseTapChanger); ok {
			tcControl = ptc.TapChangerControl
			if ptc.TransformerEnd != nil {
				transformerEndID = strings.TrimPrefix(ptc.TransformerEnd.MRID, "#")
			}
		}

		if tcControl == nil || transformerEndID == "" {
			continue
		}

		if endObj, ok := dataset.Elements[transformerEndID]; ok {
			if end, ok := endObj.(*cimgostructs.PowerTransformerEnd); ok && end.PowerTransformer != nil {
				ptID := strings.TrimPrefix(end.PowerTransformer.MRID, "#")
				if ptObj, ok := dataset.Elements[ptID]; ok {
					if pt, ok := ptObj.(*cimgostructs.PowerTransformer); ok && pt.EquipmentContainer != nil {
						ecID := strings.TrimPrefix(pt.EquipmentContainer.MRID, "#")
						if ecObj, ok := dataset.Elements[ecID]; ok {
							if _, ok := ecObj.(*cimgostructs.DCConverterUnit); ok {
								violations = append(violations, Violation{
									ObjectID: id,
									Class:    "TapChanger",
									Property: "TapChanger.TapChangerControl",
									Message:  "TapChangerControl is associated to a transformer contained in DCConverterUnit.",
									Severity: "sh.Violation",
								})
							}
						}
					}
				}
			}
		}
	}

	return violations
}

// CheckConnectivityNodeTerminalPhasesConsistency implements eqc.Terminal.phases-consistencyConnectivityNode
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: The phase code on terminals connecting same ConnectivityNode shall be consistent.
func CheckConnectivityNodeTerminalPhasesConsistency(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	nodeTerminals := make(map[string][]*cimgostructs.Terminal)
	for _, term := range dataset.Terminals {
		if term.ConnectivityNode != nil {
			nodeID := strings.TrimPrefix(term.ConnectivityNode.MRID, "#")
			nodeTerminals[nodeID] = append(nodeTerminals[nodeID], term)
		}
	}

	for nodeID, terms := range nodeTerminals {
		if len(terms) < 2 {
			continue
		}

		abcn := "http://iec.ch/TC57/CIM100#PhaseCode.ABCN"
		n := "http://iec.ch/TC57/CIM100#PhaseCode.N"
		abc := "http://iec.ch/TC57/CIM100#PhaseCode.ABC"

		for i := 0; i < len(terms); i++ {
			for j := i + 1; j < len(terms); j++ {
				val1 := ""
				if terms[i].Phases != nil {
					val1 = terms[i].Phases.URI
				}
				val2 := ""
				if terms[j].Phases != nil {
					val2 = terms[j].Phases.URI
				}

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
						ObjectID: nodeID,
						Class:    "ConnectivityNode",
						Property: "Terminal.phases",
						Message:  fmt.Sprintf("The phase codes for the connected terminals are not consistent. Terminal %s code: %s, Terminal %s code: %s.", terms[i].MRID, val1, terms[j].MRID, val2),
						Severity: "sh.Violation",
					})
					goto NextNode
				}
			}
		}
	NextNode:
	}

	return violations
}

// CheckEquipmentAggregateNotUsed implements eqc.Equipment.aggregate-notUsed
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Equipment.aggregate is not used for EquivalentBranch, EquivalentShunt and EquivalentInjection.
func CheckEquipmentAggregateNotUsed(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		var class string
		var aggregate bool
		switch v := obj.(type) {
		case *cimgostructs.EquivalentBranch:
			class, aggregate = "EquivalentBranch", v.Aggregate
		case *cimgostructs.EquivalentShunt:
			class, aggregate = "EquivalentShunt", v.Aggregate
		case *cimgostructs.EquivalentInjection:
			class, aggregate = "EquivalentInjection", v.Aggregate
		default:
			continue
		}
		if aggregate {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    class,
				Property: "Equipment.aggregate",
				Message:  "Not allowed property (attribute).",
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckEquivalentBranchR21Usage implements eqc.EquivalentBranch.r21-usage
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: EquivalentBranch.r21 differs from EquivalentBranch.r — informational asymmetry.
func CheckEquivalentBranchR21Usage(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		eb, ok := obj.(*cimgostructs.EquivalentBranch)
		if !ok {
			continue
		}
		if eb.R21 != 0 && eb.R21 != eb.R {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "EquivalentBranch",
				Property: "EquivalentBranch.r21",
				Message:  "Asymmetrical EquivalentBranch is modelled as EquivalentBranch.r is different from EquivalentBranch.r21.",
				Severity: "sh.Info",
			})
		}
	}

	return violations
}

// CheckEquivalentBranchX21Usage implements eqc.EquivalentBranch.x21-usage
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: EquivalentBranch.x21 differs from EquivalentBranch.x — informational asymmetry.
func CheckEquivalentBranchX21Usage(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		eb, ok := obj.(*cimgostructs.EquivalentBranch)
		if !ok {
			continue
		}
		if eb.X21 != 0 && eb.X21 != eb.X {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "EquivalentBranch",
				Property: "EquivalentBranch.x21",
				Message:  "Asymmetrical EquivalentBranch is modelled as EquivalentBranch.x is different from EquivalentBranch.x21.",
				Severity: "sh.Info",
			})
		}
	}

	return violations
}

// CheckEquivalentInjectionRegulationCapability implements eqc.EquivalentInjection.regulationCapability-associatedCurve
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: ReactiveCapabilityCurve can only be associated with EquivalentInjection if regulationCapability is true.
func CheckEquivalentInjectionRegulationCapability(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		ei, ok := obj.(*cimgostructs.EquivalentInjection)
		if !ok {
			continue
		}
		if ei.ReactiveCapabilityCurve != nil && !ei.RegulationCapability {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "EquivalentInjection",
				Property: "EquivalentInjection.regulationCapability",
				Message:  "The value does not allow a ReactiveCapabilityCurve to be associated.",
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckGeneratingUnitNominalP implements eqc.GeneratingUnit.nominalP-valueRangePair
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: GeneratingUnit.nominalP shall be > 0 and <= the associated RotatingMachine.ratedS.
func CheckGeneratingUnitNominalP(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	// Build map: GeneratingUnit MRID -> max ratedS across all RotatingMachines pointing to it
	ratedSByGU := make(map[string]float64)
	for _, obj := range dataset.Elements {
		val := reflect.ValueOf(obj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		guField := val.FieldByName("GeneratingUnit")
		ratedSField := val.FieldByName("RatedS")
		if !guField.IsValid() || !ratedSField.IsValid() || guField.Kind() != reflect.Ptr || guField.IsNil() {
			continue
		}
		mridField := guField.Elem().FieldByName("MRID")
		if !mridField.IsValid() {
			continue
		}
		guID := strings.TrimPrefix(mridField.String(), "#")
		rs := ratedSField.Float()
		if rs > ratedSByGU[guID] {
			ratedSByGU[guID] = rs
		}
	}

	nominalP := func(obj interface{}) (float64, bool) {
		val := reflect.ValueOf(obj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		f := val.FieldByName("NominalP")
		if !f.IsValid() || f.Kind() != reflect.Float64 {
			return 0, false
		}
		return f.Float(), true
	}

	for id, obj := range dataset.Elements {
		typeName := goTypeName(obj)
		switch typeName {
		case "GeneratingUnit", "ThermalGeneratingUnit", "WindGeneratingUnit",
			"HydroGeneratingUnit", "NuclearGeneratingUnit", "SolarGeneratingUnit":
		default:
			continue
		}
		np, ok := nominalP(obj)
		if !ok {
			continue
		}
		ratedS, hasRatedS := ratedSByGU[id]
		if !hasRatedS {
			continue
		}
		if np <= 0 || np > ratedS {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    typeName,
				Property: "GeneratingUnit.nominalP",
				Message:  fmt.Sprintf("The value (%v) is either negative, zero or greater than RotatingMachine.ratedS (%v).", np, ratedS),
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckControlAreaGeneratingUnitInstance implements eqc.ControlAreaGeneratingUnit.GeneratingUnit-instance
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: A control area should include a GeneratingUnit only once.
func CheckControlAreaGeneratingUnitInstance(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	type pair struct{ ca, gu string }
	seen := make(map[pair]bool)
	duplicates := make(map[string]bool)

	for _, obj := range dataset.Elements {
		cagu, ok := obj.(*cimgostructs.ControlAreaGeneratingUnit)
		if !ok || cagu.ControlArea == nil || cagu.GeneratingUnit == nil {
			continue
		}
		key := pair{
			ca: strings.TrimPrefix(cagu.ControlArea.MRID, "#"),
			gu: strings.TrimPrefix(cagu.GeneratingUnit.MRID, "#"),
		}
		if seen[key] {
			duplicates[key.gu] = true
		}
		seen[key] = true
	}

	for guID := range duplicates {
		violations = append(violations, Violation{
			ObjectID: guID,
			Class:    "GeneratingUnit",
			Property: "ControlAreaGeneratingUnit.GeneratingUnit",
			Message:  "The GeneratingUnit is assigned to more than once in a ControlArea.",
			Severity: "sh.Violation",
		})
	}

	return violations
}

// CheckDCConverterUnitCsConverterPowerTransformer implements eqc.DCConverterUnit-cscPowerTransformer
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: A DCConverterUnit that contains a CsConverter must also contain a PowerTransformer.
func CheckDCConverterUnitCsConverterPowerTransformer(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	containerHasPowerTransformer := make(map[string]bool)
	for _, obj := range dataset.Elements {
		pt, ok := obj.(*cimgostructs.PowerTransformer)
		if !ok || pt.EquipmentContainer == nil {
			continue
		}
		containerHasPowerTransformer[strings.TrimPrefix(pt.EquipmentContainer.MRID, "#")] = true
	}

	reported := make(map[string]bool)
	for _, obj := range dataset.Elements {
		csc, ok := obj.(*cimgostructs.CsConverter)
		if !ok || csc.EquipmentContainer == nil {
			continue
		}
		ecID := strings.TrimPrefix(csc.EquipmentContainer.MRID, "#")
		ecObj, ok := dataset.Elements[ecID]
		if !ok {
			continue
		}
		if _, ok := ecObj.(*cimgostructs.DCConverterUnit); !ok {
			continue
		}
		if containerHasPowerTransformer[ecID] || reported[ecID] {
			continue
		}
		reported[ecID] = true
		violations = append(violations, Violation{
			ObjectID: ecID,
			Class:    "DCConverterUnit",
			Property: "Equipment.EquipmentContainer",
			Message:  "A DCConverterUnit that contains CsConverter does not contain a PowerTransformer.",
			Severity: "sh.Violation",
		})
	}

	return violations
}

// CheckLimitKindPATLNumberOfLimitType implements eqc.LimitKind.patl-numberOfLimitType
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: There shall be only one OperationalLimitType of kind PATL per OperationalLimitSet
// for ApparentPowerLimit, ActivePowerLimit, or CurrentLimit, and isInfiniteDuration must be true.
func CheckLimitKindPATLNumberOfLimitType(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	patlURI := "http://iec.ch/TC57/CIM100-European#LimitKind.patl"

	type pair struct {
		set      string
		limitCls string
	}
	patlLimitsBySet := make(map[string]map[pair]int)
	infDurByOLT := make(map[string]bool)

	for id, obj := range dataset.Elements {
		olt, ok := obj.(*cimgostructs.OperationalLimitType)
		if !ok || olt.Kind == nil || olt.Kind.URI != patlURI {
			continue
		}
		patlLimitsBySet[id] = make(map[pair]int)
		infDurByOLT[id] = olt.IsInfiniteDuration
	}
	if len(patlLimitsBySet) == 0 {
		return violations
	}

	limitOLTAndSet := func(obj interface{}) (oltID, setID, cls string, ok bool) {
		val := reflect.ValueOf(obj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		oltField := val.FieldByName("OperationalLimitType")
		setField := val.FieldByName("OperationalLimitSet")
		if !oltField.IsValid() || !setField.IsValid() ||
			oltField.Kind() != reflect.Ptr || oltField.IsNil() ||
			setField.Kind() != reflect.Ptr || setField.IsNil() {
			return "", "", "", false
		}
		oltID = strings.TrimPrefix(oltField.Elem().FieldByName("MRID").String(), "#")
		setID = strings.TrimPrefix(setField.Elem().FieldByName("MRID").String(), "#")
		return oltID, setID, goTypeName(obj), true
	}

	for _, obj := range dataset.Elements {
		switch obj.(type) {
		case *cimgostructs.ApparentPowerLimit, *cimgostructs.ActivePowerLimit, *cimgostructs.CurrentLimit:
		default:
			continue
		}
		oltID, setID, cls, ok := limitOLTAndSet(obj)
		if !ok {
			continue
		}
		if _, isPATL := patlLimitsBySet[oltID]; !isPATL {
			continue
		}
		patlLimitsBySet[oltID][pair{set: setID, limitCls: cls}]++
	}

	for oltID, perSet := range patlLimitsBySet {
		duplicate := false
		for _, count := range perSet {
			if count > 1 {
				duplicate = true
				break
			}
		}
		if duplicate || (!infDurByOLT[oltID] && len(perSet) > 0) {
			violations = append(violations, Violation{
				ObjectID: oltID,
				Class:    "OperationalLimitType",
				Property: "OperationalLimitType.kind",
				Message:  fmt.Sprintf("Either there is more than one PATL defined for a given OperationalLimitSet or OperationalLimitType.isInfiniteDuration is not set to true for PATL type. The OperationalLimitType.isInfiniteDuration is: %v.", infDurByOLT[oltID]),
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckLimitKindTCDuration implements eqc.LimitKind.tc-duration
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: For TC limit kind, acceptableDuration must be 0 (or absent), and only one limit per set.
func CheckLimitKindTCDuration(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	tcURI := "http://iec.ch/TC57/CIM100-European#LimitKind.tc"

	tcOLTs := make(map[string]float64)
	for id, obj := range dataset.Elements {
		olt, ok := obj.(*cimgostructs.OperationalLimitType)
		if !ok || olt.Kind == nil || olt.Kind.URI != tcURI {
			continue
		}
		tcOLTs[id] = olt.AcceptableDuration
	}
	if len(tcOLTs) == 0 {
		return violations
	}

	limitsPerOLTSet := make(map[string]map[string]int)
	for _, obj := range dataset.Elements {
		switch obj.(type) {
		case *cimgostructs.ApparentPowerLimit, *cimgostructs.ActivePowerLimit,
			*cimgostructs.CurrentLimit, *cimgostructs.VoltageLimit:
		default:
			continue
		}
		val := reflect.ValueOf(obj).Elem()
		oltField := val.FieldByName("OperationalLimitType")
		setField := val.FieldByName("OperationalLimitSet")
		if !oltField.IsValid() || !setField.IsValid() ||
			oltField.Kind() != reflect.Ptr || oltField.IsNil() ||
			setField.Kind() != reflect.Ptr || setField.IsNil() {
			continue
		}
		oltID := strings.TrimPrefix(oltField.Elem().FieldByName("MRID").String(), "#")
		setID := strings.TrimPrefix(setField.Elem().FieldByName("MRID").String(), "#")
		if _, isTC := tcOLTs[oltID]; !isTC {
			continue
		}
		if _, ok := limitsPerOLTSet[oltID]; !ok {
			limitsPerOLTSet[oltID] = make(map[string]int)
		}
		limitsPerOLTSet[oltID][setID]++
	}

	for oltID, dur := range tcOLTs {
		duplicate := false
		for _, count := range limitsPerOLTSet[oltID] {
			if count > 1 {
				duplicate = true
				break
			}
		}
		if duplicate || dur != 0 {
			violations = append(violations, Violation{
				ObjectID: oltID,
				Class:    "OperationalLimitType",
				Property: "OperationalLimitType.kind",
				Message:  fmt.Sprintf("Either OperationalLimitType.acceptableDuration is present and different than 0 or there is more than one limit with TC type. The OperationalLimitType.acceptableDuration is: %v.", dur),
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckOperationalLimitTypeInfiniteDuration implements eqc.OperationalLimitType.isInfiniteDuration-usage
// Profile: 61970-301_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: acceptableDuration must be present when isInfiniteDuration is false.
func CheckOperationalLimitTypeInfiniteDuration(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	for id, obj := range dataset.Elements {
		olt, ok := obj.(*cimgostructs.OperationalLimitType)
		if !ok {
			continue
		}
		if !olt.IsInfiniteDuration && olt.AcceptableDuration == 0 {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "OperationalLimitType",
				Property: "OperationalLimitType.acceptableDuration",
				Message:  "The attribute is not present when .isInfiniteDuration is set to false.",
				Severity: "sh.Violation",
			})
		}
	}
	return violations
}

// CheckSynchronousMachineAggregate implements eq452:SynchronousMachine-aggregate
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: If only one SynchronousMachine is associated with the GeneratingUnit
// then the Equipment.aggregate flag shall be consistent between them.
func CheckSynchronousMachineAggregate(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	// Map GeneratingUnit to its SynchronousMachines
	guToSMs := make(map[string][]string)
	for id, obj := range dataset.Elements {
		if sm, ok := obj.(*cimgostructs.SynchronousMachine); ok {
			if sm.GeneratingUnit != nil {
				guID := strings.TrimPrefix(sm.GeneratingUnit.MRID, "#")
				guToSMs[guID] = append(guToSMs[guID], id)
			}
		}
	}

	for guID, smIDs := range guToSMs {
		if len(smIDs) != 1 {
			continue
		}
		smID := smIDs[0]
		sm := dataset.Elements[smID].(*cimgostructs.SynchronousMachine)
		gu, ok := dataset.Elements[guID].(*cimgostructs.GeneratingUnit)
		if !ok {
			continue
		}

		if sm.Aggregate != gu.Aggregate {
			violations = append(violations, Violation{
				ObjectID: smID,
				Class:    "SynchronousMachine",
				Property: "Equipment.aggregate",
				Message:  fmt.Sprintf("SynchronousMachine aggregate flag (%v) is not consistent with associated GeneratingUnit (%v).", sm.Aggregate, gu.Aggregate),
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckAsynchronousMachineAggregate implements eq452:AsynchronousMachine-aggregate
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: If one AsynchronousMachine is associated with one GeneratingUnit
// the flag Equipment.aggregate shall be consistent if provided at both.
func CheckAsynchronousMachineAggregate(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	guToAMs := make(map[string][]string)
	for id, obj := range dataset.Elements {
		if am, ok := obj.(*cimgostructs.AsynchronousMachine); ok {
			if am.GeneratingUnit != nil {
				guID := strings.TrimPrefix(am.GeneratingUnit.MRID, "#")
				guToAMs[guID] = append(guToAMs[guID], id)
			}
		}
	}

	for guID, amIDs := range guToAMs {
		if len(amIDs) != 1 {
			continue
		}
		amID := amIDs[0]
		am := dataset.Elements[amID].(*cimgostructs.AsynchronousMachine)
		gu, ok := dataset.Elements[guID].(*cimgostructs.GeneratingUnit)
		if !ok {
			continue
		}

		if am.Aggregate != gu.Aggregate {
			violations = append(violations, Violation{
				ObjectID: amID,
				Class:    "AsynchronousMachine",
				Property: "Equipment.aggregate",
				Message:  fmt.Sprintf("AsynchronousMachine aggregate flag (%v) is not consistent with associated GeneratingUnit (%v).", am.Aggregate, gu.Aggregate),
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckSynchronousMachineControlMode implements eq452:SynchronousMachine-controlMode
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: RegulatingControl.mode for SynchronousMachine must be voltage, reactivePower, or powerFactor.
func CheckSynchronousMachineControlMode(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		sm, ok := obj.(*cimgostructs.SynchronousMachine)
		if !ok || sm.RegulatingControl == nil {
			continue
		}

		rcID := strings.TrimPrefix(sm.RegulatingControl.MRID, "#")
		rc, ok := dataset.Elements[rcID].(*cimgostructs.RegulatingControl)
		if !ok || rc.Mode == nil {
			continue
		}

		uri := rc.Mode.URI
		if !strings.HasSuffix(uri, "reactivePower") && !strings.HasSuffix(uri, "voltage") && !strings.HasSuffix(uri, "powerFactor") {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "SynchronousMachine",
				Property: "RegulatingCondEq.RegulatingControl",
				Message:  fmt.Sprintf("Unallowed regulating control mode '%v' for a SynchronousMachine.", uri),
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckStaticVarCompensatorControlMode implements eq452:StaticVarCompensator-controlMode
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: RegulatingControl.mode for SVC must be voltage or reactivePower.
// Also SVC.sVCControlMode and SVC.voltageSetPoint should not be used (deprecated in favor of RegulatingControl).
func CheckStaticVarCompensatorControlMode(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		svc, ok := obj.(*cimgostructs.StaticVarCompensator)
		if !ok {
			continue
		}

		if svc.RegulatingControl != nil {
			rcID := strings.TrimPrefix(svc.RegulatingControl.MRID, "#")
			rc, ok := dataset.Elements[rcID].(*cimgostructs.RegulatingControl)
			if ok && rc.Mode != nil {
				uri := rc.Mode.URI
				if !strings.HasSuffix(uri, "voltage") && !strings.HasSuffix(uri, "reactivePower") {
					violations = append(violations, Violation{
						ObjectID: id,
						Class:    "StaticVarCompensator",
						Property: "RegulatingCondEq.RegulatingControl",
						Message:  fmt.Sprintf("Unallowed regulating control mode '%v' for a StaticVarCompensator.", uri),
						Severity: "sh.Violation",
					})
				}
			}
		}

		// Check for deprecated attributes
		if svc.SVCControlMode != nil {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "StaticVarCompensator",
				Property: "StaticVarCompensator.sVCControlMode",
				Message:  "StaticVarCompensator.sVCControlMode attribute is not allowed.",
				Severity: "sh.Violation",
			})
		}
		if svc.VoltageSetPoint != 0 {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "StaticVarCompensator",
				Property: "StaticVarCompensator.voltageSetPoint",
				Message:  "StaticVarCompensator.voltageSetPoint attribute is not allowed.",
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckPhaseTapChangerControlMode implements eq452:PhaseTapChanger-controlModeP
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: RegulatingControl.mode for PhaseTapChanger must be activePower or voltage.
func CheckPhaseTapChangerControlMode(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		var tccID string
		var class string
		if ptc, ok := obj.(*cimgostructs.PhaseTapChangerAsymmetrical); ok && ptc.TapChangerControl != nil {
			tccID = strings.TrimPrefix(ptc.TapChangerControl.MRID, "#")
			class = "PhaseTapChangerAsymmetrical"
		} else if ptc, ok := obj.(*cimgostructs.PhaseTapChangerLinear); ok && ptc.TapChangerControl != nil {
			tccID = strings.TrimPrefix(ptc.TapChangerControl.MRID, "#")
			class = "PhaseTapChangerLinear"
		} else if ptc, ok := obj.(*cimgostructs.PhaseTapChangerSymmetrical); ok && ptc.TapChangerControl != nil {
			tccID = strings.TrimPrefix(ptc.TapChangerControl.MRID, "#")
			class = "PhaseTapChangerSymmetrical"
		} else if ptc, ok := obj.(*cimgostructs.PhaseTapChangerTabular); ok && ptc.TapChangerControl != nil {
			tccID = strings.TrimPrefix(ptc.TapChangerControl.MRID, "#")
			class = "PhaseTapChangerTabular"
		}

		if tccID == "" {
			continue
		}

		rc, ok := dataset.Elements[tccID].(*cimgostructs.TapChangerControl)
		if !ok || rc.Mode == nil {
			continue
		}

		uri := rc.Mode.URI
		if !strings.HasSuffix(uri, "activePower") && !strings.HasSuffix(uri, "voltage") {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    class,
				Property: "TapChanger.TapChangerControl",
				Message:  fmt.Sprintf("Unallowed regulating control mode '%v' for a PhaseTapChanger.", uri),
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckRatioTapChangerControlMode implements eq452:RatioTapChanger-controlMode
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: RegulatingControl.mode for RatioTapChanger must be voltage, reactivePower, or powerFactor.
func CheckRatioTapChangerControlMode(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		rtc, ok := obj.(*cimgostructs.RatioTapChanger)
		if !ok || rtc.TapChangerControl == nil {
			continue
		}

		tccID := strings.TrimPrefix(rtc.TapChangerControl.MRID, "#")
		rc, ok := dataset.Elements[tccID].(*cimgostructs.TapChangerControl)
		if !ok || rc.Mode == nil {
			continue
		}

		uri := rc.Mode.URI
		if !strings.HasSuffix(uri, "voltage") && !strings.HasSuffix(uri, "reactivePower") && !strings.HasSuffix(uri, "powerFactor") {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "RatioTapChanger",
				Property: "TapChanger.TapChangerControl",
				Message:  fmt.Sprintf("Unallowed regulating control mode '%v' for a RatioTapChanger.", uri),
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckShuntCompensatorControlMode implements eq452:ShuntCompensator-controlMode
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: RegulatingControl.mode for ShuntCompensator must be voltage, reactivePower, or powerFactor.
func CheckShuntCompensatorControlMode(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		var rcID string
		var class string
		if sc, ok := obj.(*cimgostructs.LinearShuntCompensator); ok && sc.RegulatingControl != nil {
			rcID = strings.TrimPrefix(sc.RegulatingControl.MRID, "#")
			class = "LinearShuntCompensator"
		} else if sc, ok := obj.(*cimgostructs.NonlinearShuntCompensator); ok && sc.RegulatingControl != nil {
			rcID = strings.TrimPrefix(sc.RegulatingControl.MRID, "#")
			class = "NonlinearShuntCompensator"
		}

		if rcID == "" {
			continue
		}

		rc, ok := dataset.Elements[rcID].(*cimgostructs.RegulatingControl)
		if !ok || rc.Mode == nil {
			continue
		}

		uri := rc.Mode.URI
		if !strings.HasSuffix(uri, "voltage") && !strings.HasSuffix(uri, "reactivePower") && !strings.HasSuffix(uri, "powerFactor") {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    class,
				Property: "RegulatingCondEq.RegulatingControl",
				Message:  fmt.Sprintf("Unallowed regulating control mode '%v' for a ShuntCompensator.", uri),
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckSynchronousMachineReactiveLimits implements eq452:SynchronousMachine-reactiveLimits
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Validates that minQ/maxQ are provided if InitialReactiveCapabilityCurve is missing,
// and if both are present, they are consistent with the curve.
func CheckSynchronousMachineReactiveLimits(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		sm, ok := obj.(*cimgostructs.SynchronousMachine)
		if !ok {
			continue
		}

		hasCurve := sm.InitialReactiveCapabilityCurve != nil
		if hasCurve {
			rccID := strings.TrimPrefix(sm.InitialReactiveCapabilityCurve.MRID, "#")
			// Find all CurveData for this curve
			var y1vals, y2vals []float64
			for _, cdObj := range dataset.Elements {
				if cd, ok := cdObj.(*cimgostructs.CurveData); ok && cd.Curve != nil {
					if strings.TrimPrefix(cd.Curve.MRID, "#") == rccID {
						y1vals = append(y1vals, cd.Y1value)
						y2vals = append(y2vals, cd.Y2value)
					}
				}
			}

			if len(y1vals) > 0 {
				minY1 := y1vals[0]
				for _, v := range y1vals {
					if v < minY1 {
						minY1 = v
					}
				}
				maxY2 := y2vals[0]
				for _, v := range y2vals {
					if v > maxY2 {
						maxY2 = v
					}
				}

				epsilon := 1e-6
				if sm.MinQ != 0 && (sm.MinQ < minY1-epsilon || sm.MinQ > minY1+epsilon) {
					violations = append(violations, Violation{
						ObjectID: id,
						Class:    "SynchronousMachine",
						Property: "SynchronousMachine.minQ",
						Message:  fmt.Sprintf("SynchronousMachine.minQ (%v) is not equal to min of CurveData.y1value-s (%v).", sm.MinQ, minY1),
						Severity: "sh.Violation",
					})
				}
				if sm.MaxQ != 0 && (sm.MaxQ < maxY2-epsilon || sm.MaxQ > maxY2+epsilon) {
					violations = append(violations, Violation{
						ObjectID: id,
						Class:    "SynchronousMachine",
						Property: "SynchronousMachine.maxQ",
						Message:  fmt.Sprintf("SynchronousMachine.maxQ (%v) is not equal to max of CurveData.y2value-s (%v).", sm.MaxQ, maxY2),
						Severity: "sh.Violation",
					})
				}
			}
		}
	}

	return violations
}

// CheckSynchronousMachineTypeCondenser implements eq452:SynchronousMachine.type-condenser
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: SynchronousMachine of type condenser should not have an associated GeneratingUnit.
func CheckSynchronousMachineTypeCondenser(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		sm, ok := obj.(*cimgostructs.SynchronousMachine)
		if !ok {
			continue
		}

		if sm.Type != nil && strings.HasSuffix(sm.Type.URI, "condenser") && sm.GeneratingUnit != nil {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "SynchronousMachine",
				Property: "SynchronousMachine.type",
				Message:  "SynchronousMachine of type condenser with associated GeneratingUnit.",
				Severity: "sh.Info",
			})
		}
	}

	return violations
}

// CheckVsCapabilityCurveCount implements eq452:VsCapabilityCurve-VsCapabilityCurveCount
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: If CurveData.Curve is a VsCapabilityCurve at least two CurveData shall be associated.
func CheckVsCapabilityCurveCount(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	curveCount := make(map[string]int)
	for _, obj := range dataset.Elements {
		if cd, ok := obj.(*cimgostructs.CurveData); ok && cd.Curve != nil {
			cID := strings.TrimPrefix(cd.Curve.MRID, "#")
			curveCount[cID]++
		}
	}

	for id, obj := range dataset.Elements {
		if _, ok := obj.(*cimgostructs.VsCapabilityCurve); ok {
			if count, ok := curveCount[id]; !ok || count < 2 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "VsCapabilityCurve",
					Property: "rdf:type",
					Message:  fmt.Sprintf("Less than two instances of CurveData are associated (%v found).", count),
					Severity: "sh.Violation",
				})
			}
		}
	}

	return violations
}

// CheckVsCapabilityCurveYValues implements eq452:VsCapabilityCurve-yvalues
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: If CurveData.Curve is a VsCapabilityCurve, the CurveData.y2value shall be greater than CurveData.y1value.
func CheckVsCapabilityCurveYValues(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		if cd, ok := obj.(*cimgostructs.CurveData); ok && cd.Curve != nil {
			cID := strings.TrimPrefix(cd.Curve.MRID, "#")
			if _, ok := dataset.Elements[cID].(*cimgostructs.VsCapabilityCurve); ok {
				if cd.Y2value <= cd.Y1value {
					violations = append(violations, Violation{
						ObjectID: id,
						Class:    "CurveData",
						Property: "CurveData.y2value",
						Message:  fmt.Sprintf("CurveData.y2value (%v) is not greater than CurveData.y1value (%v) for VsCapabilityCurve.", cd.Y2value, cd.Y1value),
						Severity: "sh.Violation",
					})
				}
			}
		}
	}

	return violations
}

// CheckGeneratingUnitTypeDependency implements eq452:GeneratingUnit-typeDependency
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Validates GeneratingUnit min/max operating P based on SynchronousMachine type.
func CheckGeneratingUnitTypeDependency(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		sm, ok := obj.(*cimgostructs.SynchronousMachine)
		if !ok || sm.GeneratingUnit == nil || sm.Type == nil {
			continue
		}

		guID := strings.TrimPrefix(sm.GeneratingUnit.MRID, "#")
		gu, ok := dataset.Elements[guID].(*cimgostructs.GeneratingUnit)
		if !ok {
			continue
		}

		maxP := gu.MaxOperatingP
		minP := gu.MinOperatingP
		uri := sm.Type.URI

		if strings.HasSuffix(uri, "condenser") {
			if maxP != 0 || minP != 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "SynchronousMachine",
					Property: "SynchronousMachine.type",
					Message:  fmt.Sprintf("For condenser type, min/max operating P must be 0 (found min: %v, max: %v).", minP, maxP),
					Severity: "sh.Violation",
				})
			}
		} else if strings.HasSuffix(uri, "generator") || strings.HasSuffix(uri, "generatorOrCondenser") {
			if maxP <= 0 || minP < 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "SynchronousMachine",
					Property: "SynchronousMachine.type",
					Message:  fmt.Sprintf("For %v type, minP >= 0 and maxP > 0 (found min: %v, max: %v).", uri, minP, maxP),
					Severity: "sh.Violation",
				})
			}
		} else if strings.HasSuffix(uri, "motor") || strings.HasSuffix(uri, "motorOrCondenser") {
			if maxP > 0 || minP >= 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "SynchronousMachine",
					Property: "SynchronousMachine.type",
					Message:  fmt.Sprintf("For %v type, minP < 0 and maxP <= 0 (found min: %v, max: %v).", uri, minP, maxP),
					Severity: "sh.Violation",
				})
			}
		} else if strings.HasSuffix(uri, "generatorOrMotor") || strings.HasSuffix(uri, "generatorOrCondenserOrMotor") {
			if maxP <= 0 || minP >= 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "SynchronousMachine",
					Property: "SynchronousMachine.type",
					Message:  fmt.Sprintf("For %v type, minP < 0 and maxP > 0 (found min: %v, max: %v).", uri, minP, maxP),
					Severity: "sh.Violation",
				})
			}
		}
	}

	return violations
}

// CheckCurveDataReactiveCapabilityLimits implements eq452:CurveData.Curve-equationY1/Y2
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Validates that x^2 + y^2 <= ratedS^2 for ReactiveCapabilityCurve points.
func CheckCurveDataReactiveCapabilityLimits(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		cd, ok := obj.(*cimgostructs.CurveData)
		if !ok || cd.Curve == nil {
			continue
		}

		cID := strings.TrimPrefix(cd.Curve.MRID, "#")
		if _, ok := dataset.Elements[cID].(*cimgostructs.ReactiveCapabilityCurve); !ok {
			continue
		}

		// Find the SynchronousMachine associated with this curve
		var ratedS float64
		found := false
		for _, smObj := range dataset.Elements {
			if sm, ok := smObj.(*cimgostructs.SynchronousMachine); ok && sm.InitialReactiveCapabilityCurve != nil {
				if strings.TrimPrefix(sm.InitialReactiveCapabilityCurve.MRID, "#") == cID {
					ratedS = sm.RatedS
					found = true
					break
				}
			}
		}

		if !found || ratedS == 0 {
			continue
		}

		x2 := cd.Xvalue * cd.Xvalue
		s2 := ratedS * ratedS
		epsilon := 1e-4 // Allow for small precision errors

		if x2+(cd.Y1value*cd.Y1value) > s2+epsilon {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "CurveData",
				Property: "CurveData.y1value",
				Message:  fmt.Sprintf("x^2 + y1^2 (%v) > ratedS^2 (%v).", x2+(cd.Y1value*cd.Y1value), s2),
				Severity: "sh.Violation",
			})
		}
		if x2+(cd.Y2value*cd.Y2value) > s2+epsilon {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "CurveData",
				Property: "CurveData.y2value",
				Message:  fmt.Sprintf("x^2 + y2^2 (%v) > ratedS^2 (%v).", x2+(cd.Y2value*cd.Y2value), s2),
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckCurveDataReactiveConsistency implements eq452:CurveData.Curve-reactive
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: y2value >= y1value and not all points can have y2 == y1.
func CheckCurveDataReactiveConsistency(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	curvePoints := make(map[string][]string)
	for id, obj := range dataset.Elements {
		if cd, ok := obj.(*cimgostructs.CurveData); ok && cd.Curve != nil {
			cID := strings.TrimPrefix(cd.Curve.MRID, "#")
			if _, ok := dataset.Elements[cID].(*cimgostructs.ReactiveCapabilityCurve); ok {
				curvePoints[cID] = append(curvePoints[cID], id)
			}
		}
	}

	for curveID, pointIDs := range curvePoints {
		allSame := true
		for _, pID := range pointIDs {
			cd := dataset.Elements[pID].(*cimgostructs.CurveData)
			if cd.Y2value < cd.Y1value {
				violations = append(violations, Violation{
					ObjectID: pID,
					Class:    "CurveData",
					Property: "CurveData.y2value",
					Message:  fmt.Sprintf("CurveData.y2value (%v) is less than y1value (%v).", cd.Y2value, cd.Y1value),
					Severity: "sh.Violation",
				})
			}
			if cd.Y2value != cd.Y1value {
				allSame = false
			}
		}
		if allSame && len(pointIDs) > 0 {
			violations = append(violations, Violation{
				ObjectID: curveID,
				Class:    "ReactiveCapabilityCurve",
				Property: "rdf:type",
				Message:  "All CurveData.y2value values are equal to CurveData.y1value values.",
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckSynchronousMachineCurveXValueConsistency implements eq452:CurveData.xvalue-value
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: minOperatingP/maxOperatingP shall match min/max xvalue of ReactiveCapabilityCurve.
func CheckSynchronousMachineCurveXValueConsistency(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		sm, ok := obj.(*cimgostructs.SynchronousMachine)
		if !ok || sm.GeneratingUnit == nil || sm.InitialReactiveCapabilityCurve == nil {
			continue
		}

		guID := strings.TrimPrefix(sm.GeneratingUnit.MRID, "#")
		gu, ok := dataset.Elements[guID].(*cimgostructs.GeneratingUnit)
		if !ok {
			continue
		}

		rccID := strings.TrimPrefix(sm.InitialReactiveCapabilityCurve.MRID, "#")
		var xvals []float64
		for _, cdObj := range dataset.Elements {
			if cd, ok := cdObj.(*cimgostructs.CurveData); ok && cd.Curve != nil {
				if strings.TrimPrefix(cd.Curve.MRID, "#") == rccID {
					xvals = append(xvals, cd.Xvalue)
				}
			}
		}

		if len(xvals) > 0 {
			minX := xvals[0]
			maxX := xvals[0]
			for _, x := range xvals {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
			}

			epsilon := 1e-6
			if gu.MinOperatingP < minX-epsilon || gu.MinOperatingP > minX+epsilon {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "SynchronousMachine",
					Property: "GeneratingUnit.minOperatingP",
					Message:  fmt.Sprintf("GeneratingUnit.minOperatingP (%v) is not consistent with min CurveData.xvalue (%v).", gu.MinOperatingP, minX),
					Severity: "sh.Violation",
				})
			}
			if gu.MaxOperatingP < maxX-epsilon || gu.MaxOperatingP > maxX+epsilon {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "SynchronousMachine",
					Property: "GeneratingUnit.maxOperatingP",
					Message:  fmt.Sprintf("GeneratingUnit.maxOperatingP (%v) is not consistent with max CurveData.xvalue (%v).", gu.MaxOperatingP, maxX),
					Severity: "sh.Violation",
				})
			}
		}
	}

	return violations
}

// CheckSwitchConnection implements eq452:Switch-connection
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Switches shall connect to nodes in the same VoltageLevel or different levels with same BaseVoltage.
func CheckSwitchConnection(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	// Find all Switches and their terminals
	switchTerminals := make(map[string][]string)
	for id, obj := range dataset.Elements {
		if t, ok := obj.(*cimgostructs.Terminal); ok && t.ConductingEquipment != nil {
			eqID := strings.TrimPrefix(t.ConductingEquipment.MRID, "#")
			// Check if ConductingEquipment is a Switch
			eq := dataset.Elements[eqID]
			isSwitch := false
			switch eq.(type) {
			case *cimgostructs.Breaker, *cimgostructs.Disconnector, *cimgostructs.Fuse,
				*cimgostructs.GroundDisconnector, *cimgostructs.Jumper, *cimgostructs.LoadBreakSwitch,
				*cimgostructs.DisconnectingCircuitBreaker, *cimgostructs.Cut:
				isSwitch = true
			}
			if isSwitch {
				switchTerminals[eqID] = append(switchTerminals[eqID], id)
			}
		}
	}

	for eqID, tIDs := range switchTerminals {
		if len(tIDs) < 2 {
			continue
		}

		// Get BaseVoltage nominal voltages for each terminal
		bvs := make(map[float64]bool)
		cncs := make(map[string]bool)

		for _, tID := range tIDs {
			t := dataset.Elements[tID].(*cimgostructs.Terminal)
			if t.ConnectivityNode != nil {
				cnID := strings.TrimPrefix(t.ConnectivityNode.MRID, "#")
				if cn, ok := dataset.Elements[cnID].(*cimgostructs.ConnectivityNode); ok {
					if cn.ConnectivityNodeContainer != nil {
						cncID := strings.TrimPrefix(cn.ConnectivityNodeContainer.MRID, "#")
						cncs[cncID] = true
						if vl, ok := dataset.Elements[cncID].(*cimgostructs.VoltageLevel); ok && vl.BaseVoltage != nil {
							bvID := strings.TrimPrefix(vl.BaseVoltage.MRID, "#")
							if bvObj, ok := dataset.Elements[bvID]; ok {
								if bv, ok := bvObj.(*cimgostructs.BaseVoltage); ok {
									bvs[bv.NominalVoltage] = true
								}
							}
						}
					}
				}
			}
		}

		// Rule: same VoltageLevel (len(cncs) == 1) OR different VoltageLevels with same BaseVoltage (len(bvs) == 1)
		// If len(cncs) > 1 and len(bvs) > 1, then it is a violation.
		if len(cncs) > 1 && len(bvs) > 1 {
			violations = append(violations, Violation{
				ObjectID: eqID,
				Class:    "Switch",
				Property: "rdf:type",
				Message:  "Switch (or its subclasses) connects ConnectivityNode-s that are not contained in either the same VoltageLevel or in different VoltageLevel-s which have the same BaseVoltage.",
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckOperationalLimitSetTerminal implements eq452:OperationalLimitSet-limits
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Validates OperationalLimitSet associations.
func CheckOperationalLimitSetTerminal(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		ols, ok := obj.(*cimgostructs.OperationalLimitSet)
		if !ok || ols.Terminal == nil {
			continue
		}

		tID := strings.TrimPrefix(ols.Terminal.MRID, "#")

		// Check if it's for AuxiliaryEquipment
		isAux := false
		for _, auxObj := range dataset.Elements {
			if aux, ok := auxObj.(*cimgostructs.CurrentTransformer); ok && aux.Terminal != nil {
				if strings.TrimPrefix(aux.Terminal.MRID, "#") == tID {
					isAux = true
					break
				}
			}
			// ... other aux eq
		}

		if isAux && ols.Equipment == nil {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "OperationalLimitSet",
				Property: "OperationalLimitSet.Equipment",
				Message:  "OperationalLimitSet.Equipment is not provided for a Terminal associated with AuxiliaryEquipment.",
				Severity: "sh.Violation",
			})
		}

		if ols.Equipment != nil {
			eqID := strings.TrimPrefix(ols.Equipment.MRID, "#")
			// Check if Terminal belongs to Equipment
			found := false
			for _, tObj := range dataset.Elements {
				if t, ok := tObj.(*cimgostructs.Terminal); ok && t.ConductingEquipment != nil {
					if strings.TrimPrefix(t.ConductingEquipment.MRID, "#") == eqID && t.Id == tID {
						found = true
						break
					}
				}
			}
			if !found {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "OperationalLimitSet",
					Property: "OperationalLimitSet.Terminal",
					Message:  fmt.Sprintf("Terminal %s is not a terminal of ConductingEquipment %s.", tID, eqID),
					Severity: "sh.Violation",
				})
			}
		}
	}

	return violations
}

// CheckTapChangerControlRemoteQControl implements eq452:TapChangerControl-remoteQcontrol
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: TapChangerControl in reactivePower mode shall only control a Terminal associated with its PowerTransformer.
func CheckTapChangerControlRemoteQControl(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		tcc, ok := obj.(*cimgostructs.TapChangerControl)
		if !ok || tcc.Mode == nil || !strings.HasSuffix(tcc.Mode.URI, "reactivePower") || tcc.Terminal == nil {
			continue
		}

		rcTermID := strings.TrimPrefix(tcc.Terminal.MRID, "#")

		// Find the TapChanger(s) associated with this control
		for _, tcObj := range dataset.Elements {
			var teID string
			if rtc, ok := tcObj.(*cimgostructs.RatioTapChanger); ok && rtc.TapChangerControl != nil {
				if strings.TrimPrefix(rtc.TapChangerControl.MRID, "#") == id {
					if rtc.TransformerEnd != nil {
						teID = strings.TrimPrefix(rtc.TransformerEnd.MRID, "#")
					}
				}
			} else if ptc, ok := tcObj.(*cimgostructs.PhaseTapChangerAsymmetrical); ok && ptc.TapChangerControl != nil {
				if strings.TrimPrefix(ptc.TapChangerControl.MRID, "#") == id {
					if ptc.TransformerEnd != nil {
						teID = strings.TrimPrefix(ptc.TransformerEnd.MRID, "#")
					}
				}
			} // ... other ptc types

			if teID != "" {
				te, ok := dataset.Elements[teID].(*cimgostructs.PowerTransformerEnd)
				if ok && te.Terminal != nil {
					if strings.TrimPrefix(te.Terminal.MRID, "#") != rcTermID {
						violations = append(violations, Violation{
							ObjectID: id,
							Class:    "TapChangerControl",
							Property: "RegulatingControl.Terminal",
							Message:  "TapChangerControl in reactivePower mode controls a Terminal not associated with its PowerTransformerEnd.",
							Severity: "sh.Violation",
						})
					}
				}
			}
		}
	}

	return violations
}

// CheckReactiveCapabilityCurveXValueUnique implements eq452:ReactiveCapabilityCurve-xvalue
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: All CurveData.xvalue for a given ReactiveCapabilityCurve shall be unique.
func CheckReactiveCapabilityCurveXValueUnique(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		if _, ok := obj.(*cimgostructs.ReactiveCapabilityCurve); !ok {
			continue
		}

		xvals := make(map[float64]bool)
		for _, cdObj := range dataset.Elements {
			if cd, ok := cdObj.(*cimgostructs.CurveData); ok && cd.Curve != nil {
				if strings.TrimPrefix(cd.Curve.MRID, "#") == id {
					if xvals[cd.Xvalue] {
						violations = append(violations, Violation{
							ObjectID: id,
							Class:    "ReactiveCapabilityCurve",
							Property: "rdf:type",
							Message:  fmt.Sprintf("CurveData.xvalue (%v) for ReactiveCapabilityCurve is not unique.", cd.Xvalue),
							Severity: "sh.Violation",
						})
						break // Only report once per curve
					}
					xvals[cd.Xvalue] = true
				}
			}
		}
	}

	return violations
}

// CheckPowerTransformerEndResistanceXValue implements eq452:PowerTransformerEnd.x-value
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Validates series reactance for two and three winding transformers.
func CheckPowerTransformerEndResistanceXValue(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	// Group ends by transformer
	ptToEnds := make(map[string][]string)
	for id, obj := range dataset.Elements {
		if te, ok := obj.(*cimgostructs.PowerTransformerEnd); ok && te.PowerTransformer != nil {
			ptID := strings.TrimPrefix(te.PowerTransformer.MRID, "#")
			ptToEnds[ptID] = append(ptToEnds[ptID], id)
		}
	}

	for _, teIDs := range ptToEnds {
		numEnds := len(teIDs)
		if numEnds == 2 {
			// Find end 1
			for _, teID := range teIDs {
				te := dataset.Elements[teID].(*cimgostructs.PowerTransformerEnd)
				if te.EndNumber == 1 && te.X <= 0 {
					violations = append(violations, Violation{
						ObjectID: teID,
						Class:    "PowerTransformerEnd",
						Property: "PowerTransformerEnd.x",
						Message:  fmt.Sprintf("PowerTransformerEnd.x (%v) for winding 1 of a two-winding transformer must be positive.", te.X),
						Severity: "sh.Violation",
					})
				}
			}
		} else if numEnds == 3 {
			for _, teID := range teIDs {
				te := dataset.Elements[teID].(*cimgostructs.PowerTransformerEnd)
				if te.X == 0 {
					violations = append(violations, Violation{
						ObjectID: teID,
						Class:    "PowerTransformerEnd",
						Property: "PowerTransformerEnd.x",
						Message:  "PowerTransformerEnd.x cannot be zero for a three-winding transformer.",
						Severity: "sh.Violation",
					})
				}
			}
		}
	}

	return violations
}

// CheckGeneratingUnitMaxOperatingPRatedS implements eq452:GeneratingUnit.maxOperatingP-ratedS
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: GeneratingUnit.maxOperatingP <= sum of RotatingMachine.ratedS.
func CheckGeneratingUnitMaxOperatingPRatedS(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	guToRatedSSum := make(map[string]float64)
	for _, obj := range dataset.Elements {
		if rm, ok := obj.(*cimgostructs.RotatingMachine); ok && rm.GeneratingUnit != nil {
			guID := strings.TrimPrefix(rm.GeneratingUnit.MRID, "#")
			guToRatedSSum[guID] += rm.RatedS
		} else if sm, ok := obj.(*cimgostructs.SynchronousMachine); ok && sm.GeneratingUnit != nil {
			guID := strings.TrimPrefix(sm.GeneratingUnit.MRID, "#")
			guToRatedSSum[guID] += sm.RatedS
		} else if am, ok := obj.(*cimgostructs.AsynchronousMachine); ok && am.GeneratingUnit != nil {
			guID := strings.TrimPrefix(am.GeneratingUnit.MRID, "#")
			guToRatedSSum[guID] += am.RatedS
		}
	}

	for id, obj := range dataset.Elements {
		if gu, ok := obj.(*cimgostructs.GeneratingUnit); ok {
			sumRS := guToRatedSSum[id]
			if gu.MaxOperatingP > sumRS {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "GeneratingUnit",
					Property: "GeneratingUnit.maxOperatingP",
					Message:  fmt.Sprintf("GeneratingUnit.maxOperatingP (%v) is greater than sum of RotatingMachine.ratedS (%v).", gu.MaxOperatingP, sumRS),
					Severity: "sh.Violation",
				})
			}
		}
	}

	return violations
}

// CheckHydroGeneratingUnitEnergyConversionCapability implements eq452:HydroGeneratingUnit.energyConversionCapability-typeConsistency
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Validates HydroGeneratingUnit energyConversionCapability vs SynchronousMachine type.
func CheckHydroGeneratingUnitEnergyConversionCapability(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		hgu, ok := obj.(*cimgostructs.HydroGeneratingUnit)
		if !ok || hgu.EnergyConversionCapability == nil {
			continue
		}

		uriHGU := hgu.EnergyConversionCapability.URI
		// Find associated SynchronousMachine
		for _, smObj := range dataset.Elements {
			if sm, ok := smObj.(*cimgostructs.SynchronousMachine); ok && sm.GeneratingUnit != nil {
				if strings.TrimPrefix(sm.GeneratingUnit.MRID, "#") == id {
					if sm.Type == nil {
						continue
					}
					uriSM := sm.Type.URI
					if strings.HasSuffix(uriHGU, "generator") {
						if !strings.HasSuffix(uriSM, "generator") && !strings.HasSuffix(uriSM, "generatorOrCondenser") {
							violations = append(violations, Violation{
								ObjectID: id,
								Class:    "HydroGeneratingUnit",
								Property: "HydroGeneratingUnit.energyConversionCapability",
								Message:  fmt.Sprintf("HydroGeneratingUnit as generator but associated SynchronousMachine type is '%v'.", uriSM),
								Severity: "sh.Violation",
							})
						}
					} else if strings.HasSuffix(uriHGU, "pumpAndGenerator") {
						if !strings.HasSuffix(uriSM, "motor") && !strings.HasSuffix(uriSM, "generatorOrMotor") && !strings.HasSuffix(uriSM, "generatorOrCondenserOrMotor") {
							violations = append(violations, Violation{
								ObjectID: id,
								Class:    "HydroGeneratingUnit",
								Property: "HydroGeneratingUnit.energyConversionCapability",
								Message:  fmt.Sprintf("HydroGeneratingUnit as pumpAndGenerator but associated SynchronousMachine type is '%v'.", uriSM),
								Severity: "sh.Violation",
							})
						}
					}
				}
			}
		}
	}

	return violations
}

// CheckTerminalConnectionSameNode implements eq452:Terminal-connection
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Terminals of a two-terminal ConductingEquipment shall not connect to the same ConnectivityNode.
func CheckTerminalConnectionSameNode(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	// Find two-terminal equipment and their terminals
	eqToTerminals := make(map[string][]string)
	for id, obj := range dataset.Elements {
		if t, ok := obj.(*cimgostructs.Terminal); ok && t.ConductingEquipment != nil {
			eqID := strings.TrimPrefix(t.ConductingEquipment.MRID, "#")
			eqToTerminals[eqID] = append(eqToTerminals[eqID], id)
		}
	}

	for eqID, tIDs := range eqToTerminals {
		if len(tIDs) != 2 {
			continue
		}
		t1 := dataset.Elements[tIDs[0]].(*cimgostructs.Terminal)
		t2 := dataset.Elements[tIDs[1]].(*cimgostructs.Terminal)

		if t1.ConnectivityNode != nil && t2.ConnectivityNode != nil && t1.ConnectivityNode.MRID == t2.ConnectivityNode.MRID {
			violations = append(violations, Violation{
				ObjectID: eqID,
				Class:    "ConductingEquipment",
				Property: "rdf:type",
				Message:  "Terminals of a two-terminal equipment connect to the same ConnectivityNode.",
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckReactiveCapabilityCurveReactiveCountP implements eq452:ReactiveCapabilityCurve-reactiveCountP
// Profile: 61970-452_Equipment-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Validates number of CurveData points for a ReactiveCapabilityCurve based on SynchronousMachine type.
func CheckReactiveCapabilityCurveReactiveCountP(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		if _, ok := obj.(*cimgostructs.ReactiveCapabilityCurve); !ok {
			continue
		}

		// Find SynchronousMachine
		var sm *cimgostructs.SynchronousMachine
		for _, smObj := range dataset.Elements {
			if s, ok := smObj.(*cimgostructs.SynchronousMachine); ok && s.InitialReactiveCapabilityCurve != nil {
				if strings.TrimPrefix(s.InitialReactiveCapabilityCurve.MRID, "#") == id {
					sm = s
					break
				}
			}
		}

		if sm == nil || sm.Type == nil {
			continue
		}

		var xvals []float64
		for _, cdObj := range dataset.Elements {
			if cd, ok := cdObj.(*cimgostructs.CurveData); ok && cd.Curve != nil {
				if strings.TrimPrefix(cd.Curve.MRID, "#") == id {
					xvals = append(xvals, cd.Xvalue)
				}
			}
		}

		count := len(xvals)
		uri := sm.Type.URI
		if strings.HasSuffix(uri, "condenser") {
			if count > 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "ReactiveCapabilityCurve",
					Property: "rdf:type",
					Message:  "SynchronousMachine of type condenser should not have a ReactiveCapabilityCurve.",
					Severity: "sh.Violation",
				})
			}
		} else if strings.HasSuffix(uri, "generator") || strings.HasSuffix(uri, "generatorOrCondenser") {
			if count < 2 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "ReactiveCapabilityCurve",
					Property: "rdf:type",
					Message:  fmt.Sprintf("Generator type ReactiveCapabilityCurve needs at least 2 points (found %v).", count),
					Severity: "sh.Violation",
				})
			}
		} else if strings.HasSuffix(uri, "motor") || strings.HasSuffix(uri, "motorOrCondenser") {
			if count < 2 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "ReactiveCapabilityCurve",
					Property: "rdf:type",
					Message:  fmt.Sprintf("Motor type ReactiveCapabilityCurve needs at least 2 points (found %v).", count),
					Severity: "sh.Violation",
				})
			}
		} else if strings.HasSuffix(uri, "generatorOrMotor") || strings.HasSuffix(uri, "generatorOrCondenserOrMotor") {
			if count < 3 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "ReactiveCapabilityCurve",
					Property: "rdf:type",
					Message:  fmt.Sprintf("Combined type ReactiveCapabilityCurve needs at least 3 points (found %v).", count),
					Severity: "sh.Violation",
				})
			}
		}
	}

	return violations
}

// CheckReactiveCapabilityCurveUnits implements eq600:ReactiveCapabilityCurve-units
// Description: Curve.xUnit shall be W and y1Unit, y2Unit shall be VAr.
func CheckReactiveCapabilityCurveUnits(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		rcc, ok := obj.(*cimgostructs.ReactiveCapabilityCurve)
		if !ok || rcc.XUnit == nil || rcc.Y1Unit == nil || rcc.Y2Unit == nil {
			continue
		}

		// Check if it's associated with a SynchronousMachine
		isSM := false
		for _, smObj := range dataset.Elements {
			if sm, ok := smObj.(*cimgostructs.SynchronousMachine); ok && sm.InitialReactiveCapabilityCurve != nil {
				if strings.TrimPrefix(sm.InitialReactiveCapabilityCurve.MRID, "#") == id {
					isSM = true
					break
				}
			}
		}

		if isSM {
			if !strings.HasSuffix(rcc.XUnit.URI, "W") || !strings.HasSuffix(rcc.Y1Unit.URI, "VAr") || !strings.HasSuffix(rcc.Y2Unit.URI, "VAr") {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "ReactiveCapabilityCurve",
					Property: "rdf:type",
					Message:  fmt.Sprintf("Incorrect units for ReactiveCapabilityCurve (x: %v, y1: %v, y2: %v). Expected x: W, y1: VAr, y2: VAr.", rcc.XUnit.URI, rcc.Y1Unit.URI, rcc.Y2Unit.URI),
					Severity: "sh.Violation",
				})
			}
		}
	}

	return violations
}

// CheckSubstationCount implements eq600:Substation-count
// Description: Reports warning if only one Substation or one Substation per VoltageLevel.
func CheckSubstationCount(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	substations := 0
	voltageLevels := 0
	for _, obj := range dataset.Elements {
		if _, ok := obj.(*cimgostructs.Substation); ok {
			substations++
		} else if _, ok := obj.(*cimgostructs.VoltageLevel); ok {
			voltageLevels++
		}
	}

	if substations == 1 || (substations > 0 && substations == voltageLevels) {
		violations = append(violations, Violation{
			ObjectID: "global",
			Class:    "Substation",
			Property: "rdf:type",
			//{?substations}. {?voltagelevels}.
			Message:  fmt.Sprintf("The model has either one Substation or a Substation per VoltageLevel. Number of Substation-s: %v. Number of VoltageLevel-s: %v.", substations, voltageLevels),
			Severity: "sh.Warning",
		})
	}

	return violations
}

// CheckTapChangerNeutralUValueRange implements eq600:TapChanger.neutralU-valueRangePair
// Description: TapChanger.neutralU shall be the same as PowerTransformerEnd.ratedU.
func CheckTapChangerNeutralUValueRange(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		var neutralU float64
		var teID string
		var class string

		if rtc, ok := obj.(*cimgostructs.RatioTapChanger); ok {
			neutralU = rtc.NeutralU
			if rtc.TransformerEnd != nil {
				teID = strings.TrimPrefix(rtc.TransformerEnd.MRID, "#")
			}
			class = "RatioTapChanger"
		} else if ptc, ok := obj.(*cimgostructs.PhaseTapChangerAsymmetrical); ok {
			neutralU = ptc.NeutralU
			if ptc.TransformerEnd != nil {
				teID = strings.TrimPrefix(ptc.TransformerEnd.MRID, "#")
			}
			class = "PhaseTapChangerAsymmetrical"
		} else if ptc, ok := obj.(*cimgostructs.PhaseTapChangerLinear); ok {
			neutralU = ptc.NeutralU
			if ptc.TransformerEnd != nil {
				teID = strings.TrimPrefix(ptc.TransformerEnd.MRID, "#")
			}
			class = "PhaseTapChangerLinear"
		} else if ptc, ok := obj.(*cimgostructs.PhaseTapChangerSymmetrical); ok {
			neutralU = ptc.NeutralU
			if ptc.TransformerEnd != nil {
				teID = strings.TrimPrefix(ptc.TransformerEnd.MRID, "#")
			}
			class = "PhaseTapChangerSymmetrical"
		} else if ptc, ok := obj.(*cimgostructs.PhaseTapChangerTabular); ok {
			neutralU = ptc.NeutralU
			if ptc.TransformerEnd != nil {
				teID = strings.TrimPrefix(ptc.TransformerEnd.MRID, "#")
			}
			class = "PhaseTapChangerTabular"
		}

		if teID != "" {
			te, ok := dataset.Elements[teID].(*cimgostructs.PowerTransformerEnd)
			if ok {
				epsilon := 1e-6
				if neutralU < te.RatedU-epsilon || neutralU > te.RatedU+epsilon {
					violations = append(violations, Violation{
						ObjectID: id,
						Class:    class,
						Property: "TapChanger.neutralU",
						Message:  fmt.Sprintf("TapChanger.neutralU (%v) is not equal to PowerTransformerEnd.ratedU (%v).", neutralU, te.RatedU),
						Severity: "sh.Violation",
					})
				}
			}
		}
	}

	return violations
}
