package validation

import (
	"cimgo/cimgostructs"
	"fmt"
	"strings"
)

// ValidateEQNotSolvedMASProfileSPARQL runs hand-written checks for
// 61970-301_Equipment-AP-Con-Complex-NotSolvedMAS-SHACL.
func ValidateEQNotSolvedMASProfileSPARQL(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckACLineSegmentBaseVoltage(dataset)...)
	violations = append(violations, CheckRegulatingControlTargetValueTapChanger(dataset)...)
	violations = append(violations, CheckACLineSegmentBaseVoltageDiff(dataset)...)
	violations = append(violations, CheckBoundaryPointBppl(dataset)...)
	violations = append(violations, CheckEquivalentInjectionRegulationCapabilityNotHVDC(dataset)...)
	return violations
}

// CheckACLineSegmentBaseVoltage implements eqcns.ACLineSegment-baseVoltage
// Profile: 61970-301_Equipment-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: The BaseVoltage at the two ends of ACLineSegments in a Line shall have the same
// BaseVoltage.nominalVoltage.
func CheckACLineSegmentBaseVoltage(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	terminalsByEquipment := make(map[string]map[int]*cimgostructs.Terminal)
	for _, term := range dataset.Terminals {
		if term.ConductingEquipment == nil {
			continue
		}
		eqID := strings.TrimPrefix(term.ConductingEquipment.MRID, "#")
		if _, ok := terminalsByEquipment[eqID]; !ok {
			terminalsByEquipment[eqID] = make(map[int]*cimgostructs.Terminal)
		}
		terminalsByEquipment[eqID][term.SequenceNumber] = term
	}

	nominalVoltage := func(term *cimgostructs.Terminal) (float64, bool) {
		if term == nil || term.TopologicalNode == nil {
			return 0, false
		}
		tnID := strings.TrimPrefix(term.TopologicalNode.MRID, "#")
		tnObj, ok := dataset.Elements[tnID]
		if !ok {
			return 0, false
		}
		tn, ok := tnObj.(*cimgostructs.TopologicalNode)
		if !ok || tn.BaseVoltage == nil {
			return 0, false
		}
		bvID := strings.TrimPrefix(tn.BaseVoltage.MRID, "#")
		bvObj, ok := dataset.Elements[bvID]
		if !ok {
			return 0, false
		}
		bv, ok := bvObj.(*cimgostructs.BaseVoltage)
		if !ok {
			return 0, false
		}
		return bv.NominalVoltage, true
	}

	for id, obj := range dataset.Elements {
		if _, ok := obj.(*cimgostructs.ACLineSegment); !ok {
			continue
		}
		terms := terminalsByEquipment[id]
		t1, t2 := terms[1], terms[2]
		v1, ok1 := nominalVoltage(t1)
		v2, ok2 := nominalVoltage(t2)
		if !ok1 || !ok2 {
			continue
		}
		if t1.TopologicalNode.MRID == t2.TopologicalNode.MRID {
			continue
		}
		if v1 != v2 {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "ACLineSegment",
				Property: "ACLineSegment.BaseVoltage",
				Message:  fmt.Sprintf("The ACLineSegment has different BaseVoltage.nominalVoltage at the two ends. Voltage at end 1 is: %v. Voltage at end 2 is: %v.", v1, v2),
				Severity: "sh.Warning",
			})
		}
	}

	return violations
}

// CheckRegulatingControlTargetValueTapChanger implements eqn452:RegulatingControl.targetValue-tapChanger
// Profile: 61970-452_Equipment-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: RegulatingControl.targetValue shall be within TapChanger capability limits.
func CheckRegulatingControlTargetValueTapChanger(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		var rc *cimgostructs.RegulatingControl
		switch v := obj.(type) {
		case *cimgostructs.RegulatingControl:
			rc = v
		case *cimgostructs.TapChangerControl:
			rc = &v.RegulatingControl
		default:
			continue
		}

		if rc.Mode == nil || !strings.HasSuffix(rc.Mode.URI, "voltage") || !rc.Enabled {
			continue
		}

		// Find associated RatioTapChanger
		for _, tcObj := range dataset.Elements {
			rtc, ok := tcObj.(*cimgostructs.RatioTapChanger)
			if !ok || rtc.TapChangerControl == nil || strings.TrimPrefix(rtc.TapChangerControl.MRID, "#") != id || !rtc.ControlEnabled {
				continue
			}

			// Get BaseVoltage nominal voltage
			var nominalU float64
			if rc.Terminal != nil {
				tID := strings.TrimPrefix(rc.Terminal.MRID, "#")
				t, ok := dataset.Elements[tID].(*cimgostructs.Terminal)
				if ok && t.ConnectivityNode != nil {
					cnID := strings.TrimPrefix(t.ConnectivityNode.MRID, "#")
					cn, ok := dataset.Elements[cnID].(*cimgostructs.ConnectivityNode)
					if ok && cn.ConnectivityNodeContainer != nil {
						cncID := strings.TrimPrefix(cn.ConnectivityNodeContainer.MRID, "#")
						if vl, ok := dataset.Elements[cncID].(*cimgostructs.VoltageLevel); ok && vl.BaseVoltage != nil {
							bvID := strings.TrimPrefix(vl.BaseVoltage.MRID, "#")
							if bvObj, ok := dataset.Elements[bvID]; ok {
								if bv, ok := bvObj.(*cimgostructs.BaseVoltage); ok {
									nominalU = bv.NominalVoltage
								}
							}
						}
					}
				}
			}

			if nominalU == 0 {
				continue
			}

			targetPU := rc.TargetValue / nominalU
			upperLimit := 1 + (rtc.StepVoltageIncrement/100)*float64(rtc.HighStep-rtc.NeutralStep)
			lowerLimit := 1 - (rtc.StepVoltageIncrement/100)*float64(rtc.NeutralStep-rtc.LowStep)

			if targetPU < lowerLimit || targetPU > upperLimit {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "RegulatingControl",
					Property: "RegulatingControl.targetValue",
					Message:  fmt.Sprintf("Target value PU (%v) is outside TapChanger capability limits [%v, %v].", targetPU, lowerLimit, upperLimit),
					Severity: "sh.Violation",
				})
			}
		}
	}

	return violations
}

// CheckACLineSegmentBaseVoltageDiff implements eqn600:ACLineSegment-BaseVoltageDiff
// Profile: 61970-600_Equipment-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: 10% difference of BaseVoltage.nominalVoltage allowed at two ends of ACLineSegment.
func CheckACLineSegmentBaseVoltageDiff(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	terminalsByEquipment := make(map[string]map[int]*cimgostructs.Terminal)
	for _, term := range dataset.Terminals {
		if term.ConductingEquipment == nil {
			continue
		}
		eqID := strings.TrimPrefix(term.ConductingEquipment.MRID, "#")
		if _, ok := terminalsByEquipment[eqID]; !ok {
			terminalsByEquipment[eqID] = make(map[int]*cimgostructs.Terminal)
		}
		terminalsByEquipment[eqID][term.SequenceNumber] = term
	}

	nominalVoltage := func(term *cimgostructs.Terminal) (float64, bool) {
		if term == nil || term.TopologicalNode == nil {
			return 0, false
		}
		tnID := strings.TrimPrefix(term.TopologicalNode.MRID, "#")
		tnObj, ok := dataset.Elements[tnID]
		if !ok {
			return 0, false
		}
		tn, ok := tnObj.(*cimgostructs.TopologicalNode)
		if !ok || tn.BaseVoltage == nil {
			return 0, false
		}
		bvID := strings.TrimPrefix(tn.BaseVoltage.MRID, "#")
		bvObj, ok := dataset.Elements[bvID]
		if !ok {
			return 0, false
		}
		bv, ok := bvObj.(*cimgostructs.BaseVoltage)
		if !ok {
			return 0, false
		}
		return bv.NominalVoltage, true
	}

	for id, obj := range dataset.Elements {
		if _, ok := obj.(*cimgostructs.ACLineSegment); !ok {
			continue
		}
		terms := terminalsByEquipment[id]
		t1, ok1 := terms[1]
		t2, ok2 := terms[2]
		if !ok1 || !ok2 {
			continue
		}
		v1, ok1v := nominalVoltage(t1)
		v2, ok2v := nominalVoltage(t2)
		if !ok1v || !ok2v {
			continue
		}

		diff := 0.0
		if v1 < v2 {
			diff = (v2 - v1) / v1
		} else {
			diff = (v1 - v2) / v2
		}

		if diff > 0.1 {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "ACLineSegment",
				Property: "rdf:type",
				Message:  fmt.Sprintf("More than 10%% difference of BaseVoltage.nominalVoltage at the two ends (V1: %v, V2: %v).", v1, v2),
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}

// CheckBoundaryPointBppl implements eqn600:BoundaryPoint-bppl1Bppl2/bppl3
// Profile: 61970-600_Equipment-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: Boundary points (ConnectivityNodes) must have connected EquivalentInjections and at least one two-terminal ConductingEquipment.
func CheckBoundaryPointBppl(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	// Identify Boundary Points (ConnectivityNodes associated with eu:BoundaryPoint)
	bpToCN := make(map[string]string)
	for id, obj := range dataset.Elements {
		if bp, ok := obj.(*cimgostructs.BoundaryPoint); ok && bp.ConnectivityNode != nil {
			bpToCN[id] = strings.TrimPrefix(bp.ConnectivityNode.MRID, "#")
		}
	}

	for _, cnID := range bpToCN {
		// Check connected equipment
		hasEqInjection := false
		hasTwoTerminalEq := false

		for _, tObj := range dataset.Elements {
			if t, ok := tObj.(*cimgostructs.Terminal); ok && t.ConnectivityNode != nil && strings.TrimPrefix(t.ConnectivityNode.MRID, "#") == cnID {
				if t.ConductingEquipment == nil {
					continue
				}
				eqID := strings.TrimPrefix(t.ConductingEquipment.MRID, "#")
				eq, ok := dataset.Elements[eqID]
				if !ok {
					continue
				}
				if _, ok := eq.(*cimgostructs.EquivalentInjection); ok {
					hasEqInjection = true
				}
				// Check if it's two-terminal
				if _, ok := eq.(*cimgostructs.ACLineSegment); ok {
					hasTwoTerminalEq = true
				}
				if _, ok := eq.(*cimgostructs.PowerTransformer); ok {
					hasTwoTerminalEq = true
				}
				if _, ok := eq.(*cimgostructs.Breaker); ok {
					hasTwoTerminalEq = true
				}
				if _, ok := eq.(*cimgostructs.Disconnector); ok {
					hasTwoTerminalEq = true
				}
			}
		}

		if !hasEqInjection {
			violations = append(violations, Violation{
				ObjectID: cnID,
				Class:    "ConnectivityNode",
				Property: "rdf:type",
				Message:  "Boundary Point ConnectivityNode does not have an EquivalentInjection connected.",
				Severity: "sh.Violation",
			})
		}
		if !hasTwoTerminalEq {
			violations = append(violations, Violation{
				ObjectID: cnID,
				Class:    "ConnectivityNode",
				Property: "rdf:type",
				Message:  "Boundary Point ConnectivityNode does not have a two-terminal ConductingEquipment connected.",
				Severity: "sh.Info",
			})
		}
	}

	return violations
}

// CheckEquivalentInjectionRegulationCapabilityNotHVDC implements eqn600:EquivalentInjection.regulationCapability-notHVDC
// Profile: 61970-600_Equipment-AP-Con-Complex-NotSolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: EquivalentInjection at non-HVDC BoundaryPoint shall have regulationCapability=false and no ReactiveCapabilityCurve.
func CheckEquivalentInjectionRegulationCapabilityNotHVDC(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		ei, ok := obj.(*cimgostructs.EquivalentInjection)
		if !ok {
			continue
		}

		// Find if it's connected to a non-HVDC BoundaryPoint
		isNonHVDCBP := false
		for _, tObj := range dataset.Elements {
			if t, ok := tObj.(*cimgostructs.Terminal); ok && t.ConductingEquipment != nil && strings.TrimPrefix(t.ConductingEquipment.MRID, "#") == id {
				if t.ConnectivityNode != nil {
					cnID := strings.TrimPrefix(t.ConnectivityNode.MRID, "#")
					for _, bpObj := range dataset.Elements {
						if bp, ok := bpObj.(*cimgostructs.BoundaryPoint); ok && bp.ConnectivityNode != nil && strings.TrimPrefix(bp.ConnectivityNode.MRID, "#") == cnID {
							if !bp.IsDirectCurrent {
								isNonHVDCBP = true
								break
							}
						}
					}
				}
			}
		}

		if isNonHVDCBP {
			if ei.RegulationCapability || ei.ReactiveCapabilityCurve != nil {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "EquivalentInjection",
					Property: "EquivalentInjection.regulationCapability",
					Message:  "EquivalentInjection at non-HVDC BoundaryPoint has regulationCapability=true or a ReactiveCapabilityCurve.",
					Severity: "sh.Violation",
				})
			}
		}
	}

	return violations
}
