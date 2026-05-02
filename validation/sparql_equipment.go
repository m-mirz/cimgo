package validation

import (
	"cimgo/cimgostructs"
	"fmt"
	"reflect"
	"strings"
)

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

// CheckLoadResponseCharacteristicSum implements eqc.LoadResponseCharacteristic.exponentModel-exponentCoefficient
// Description: Sum of coefficients shall equal 1.
func CheckLoadResponseCharacteristicSum(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		lrc, ok := obj.(*cimgostructs.LoadResponseCharacteristic)
		if !ok || lrc.ExponentModel {
			continue
		}

		// Coefficient model
		pSum := lrc.PConstantCurrent + lrc.PConstantImpedance + lrc.PConstantPower
		qSum := lrc.QConstantCurrent + lrc.QConstantImpedance + lrc.QConstantPower

		// Use small epsilon for float comparison
		epsilon := 1e-6
		if (pSum != 0 && (pSum < 1-epsilon || pSum > 1+epsilon)) || (qSum != 0 && (qSum < 1-epsilon || qSum > 1+epsilon)) {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "LoadResponseCharacteristic",
				Property: "LoadResponseCharacteristic.exponentModel",
				Message:  fmt.Sprintf("The sum of coefficients does not equal 1 (P sum: %v, Q sum: %v).", pSum, qSum),
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckNonlinearShuntCompensatorPointCount implements eqc.ShuntCompensator.maximumSections-numberOfInstances
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
