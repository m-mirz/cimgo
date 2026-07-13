package validation

import (
	"cimgo/cimstructs"
	"fmt"
	"reflect"
	"strings"
)

// ValidateSVSolvedMASProfileSPARQL runs hand-written checks for
// 61970-301_StateVariables-AP-Con-Complex-SolvedMAS-SHACL and
// 61970-456_StateVariables-AP-Con-Complex-SolvedMAS-SHACL
func ValidateSVSolvedMASProfileSPARQL(dataset *cimstructs.CIMDataset) []Violation {
	var violations []Violation
	violations = append(violations, CheckSvTapStepPositionRange(dataset)...)
	violations = append(violations, CheckSvTapStepPositionInteger(dataset)...)
	violations = append(violations, CheckSvShuntCompensatorSectionsInteger(dataset)...)
	violations = append(violations, CheckSvSwitchInstance(dataset)...)
	violations = append(violations, CheckSvPowerFlowInstance(dataset)...)
	violations = append(violations, CheckSvPowerFlowPLimits(dataset)...)
	violations = append(violations, CheckSvPowerFlowQLimits(dataset)...)
	violations = append(violations, CheckSvVoltageLimits(dataset)...)
	return violations
}

// CheckSvTapStepPositionRange implements SvTapStep.position-valueRange (StateVariables SolvedMAS).
// Profile: 61970-301_StateVariables-AP-Con-Complex-SolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: SvTapStep.position must be within [TapChanger.lowStep, TapChanger.highStep].
func CheckSvTapStepPositionRange(dataset *cimstructs.CIMDataset) []Violation {
	var violations []Violation

	tapChangerStep := func(id string) (low, high int, ok bool) {
		obj, found := dataset.ByID[id]
		if !found {
			return 0, 0, false
		}
		val := reflect.ValueOf(obj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		lowField := val.FieldByName("LowStep")
		highField := val.FieldByName("HighStep")
		if !lowField.IsValid() || !highField.IsValid() {
			return 0, 0, false
		}
		return int(lowField.Int()), int(highField.Int()), true
	}

	for id, obj := range dataset.ByID {
		sv, ok := obj.(*cimstructs.SvTapStep)
		if !ok || sv.TapChanger == nil {
			continue
		}
		tcID := strings.TrimPrefix(sv.TapChanger.MRID, "#")
		low, high, ok := tapChangerStep(tcID)
		if !ok {
			continue
		}
		if sv.Position < float64(low) || sv.Position > float64(high) {
			violations = append(violations, Violation{
				ObjectID: id,
				RuleID:   "svs301:SvTapStep.position-valueRange",
				Name:     "C:301:SV:SvTapStep.position:valueRange",
				Class:    "SvTapStep",
				Property: "SvTapStep.position",
				Message:  fmt.Sprintf("The value (%v) is out of range [%d,%d].", sv.Position, low, high),
				Severity: "sh:Violation",
			})
		}
	}

	return violations
}

// CheckSvShuntCompensatorSectionsInteger implements svs456:SvShuntCompensatorSections.sections-value
// Profile: 61970-456_StateVariables-AP-Con-Complex-SolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: In cases where RegulatingControl.discrete is true and RegulatingControl.enabled is true, SvShuntCompensatorSections.sections shall be integer.
func CheckSvShuntCompensatorSectionsInteger(dataset *cimstructs.CIMDataset) []Violation {
	var violations []Violation

	for id, svsc := range dataset.SvShuntCompensatorSectionss {
		if svsc.ShuntCompensator == nil {
			continue
		}
		scID := strings.TrimPrefix(svsc.ShuntCompensator.MRID, "#")
		scObj, ok := dataset.ByID[scID]
		if !ok {
			continue
		}

		// Find RegulatingControl
		var rc *cimstructs.RegulatingControl
		val := reflect.ValueOf(scObj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		rcField := val.FieldByName("RegulatingControl")
		if rcField.IsValid() && rcField.Kind() == reflect.Ptr && !rcField.IsNil() {
			rcID := strings.TrimPrefix(rcField.Elem().FieldByName("MRID").String(), "#")
			if rcObj, ok := dataset.ByID[rcID]; ok {
				rc, _ = rcObj.(*cimstructs.RegulatingControl)
			}
		}

		if rc != nil && rc.Enabled && rc.Discrete {
			if svsc.Sections != float64(int(svsc.Sections)) {
				violations = append(violations, Violation{
					ObjectID: id,
					RuleID:   "svs456:SvShuntCompensatorSections.sections-value",
					Name:     "C:456:SV:SvShuntCompensatorSections.sections:value",
					Class:    "SvShuntCompensatorSections",
					Property: "SvShuntCompensatorSections.sections",
					Message:  fmt.Sprintf("The value (%v) is not integer for an active discrete regulating control.", svsc.Sections),
					Severity: "sh:Violation",
				})
			}
		}
	}

	return violations
}

// CheckSvTapStepPositionInteger implements svs456:SvTapStep.position-value
// Profile: 61970-456_StateVariables-AP-Con-Complex-SolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: In cases where RegulatingControl.discrete is true and RegulatingControl.enabled is true, SvTapStep.position shall be integer.
func CheckSvTapStepPositionInteger(dataset *cimstructs.CIMDataset) []Violation {
	var violations []Violation

	for id, svts := range dataset.SvTapSteps {
		if svts.TapChanger == nil {
			continue
		}
		tcID := strings.TrimPrefix(svts.TapChanger.MRID, "#")
		tcObj, ok := dataset.ByID[tcID]
		if !ok {
			continue
		}

		// Find TapChangerControl
		var tcc *cimstructs.TapChangerControl
		val := reflect.ValueOf(tcObj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		tccField := val.FieldByName("TapChangerControl")
		if tccField.IsValid() && tccField.Kind() == reflect.Ptr && !tccField.IsNil() {
			tccID := strings.TrimPrefix(tccField.Elem().FieldByName("MRID").String(), "#")
			if tccObj, ok := dataset.ByID[tccID]; ok {
				tcc, _ = tccObj.(*cimstructs.TapChangerControl)
			}
		}

		if tcc != nil && tcc.Enabled && tcc.Discrete {
			if svts.Position != float64(int(svts.Position)) {
				violations = append(violations, Violation{
					ObjectID: id,
					RuleID:   "svs456:SvTapStep.position-value",
					Name:     "C:456:SV:SvTapStep.position:value",
					Class:    "SvTapStep",
					Property: "SvTapStep.position",
					Message:  fmt.Sprintf("The value (%v) is not integer for an active discrete regulating control.", svts.Position),
					Severity: "sh:Violation",
				})
			}
		}
	}

	return violations
}

// CheckSvSwitchInstance implements svs456:SvSwitch-instance
// Profile: 61970-456_StateVariables-AP-Con-Complex-SolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: SvSwitch must be instantiated for all switching devices.
func CheckSvSwitchInstance(dataset *cimstructs.CIMDataset) []Violation {
	var violations []Violation

	for id, obj := range dataset.ByID {
		switch obj.(type) {
		case *cimstructs.Switch, *cimstructs.Breaker, *cimstructs.LoadBreakSwitch,
			*cimstructs.Disconnector, *cimstructs.Fuse, *cimstructs.Jumper,
			*cimstructs.GroundDisconnector, *cimstructs.DisconnectingCircuitBreaker,
			*cimstructs.Cut:
		default:
			continue
		}

		found := false
		for _, svsw := range dataset.SvSwitchs {
			if svsw.Switch != nil && strings.TrimPrefix(svsw.Switch.MRID, "#") == id {
				found = true
				break
			}
		}
		if !found {
			violations = append(violations, Violation{
				ObjectID: id,
				RuleID:   "svs456:SvSwitch-instance",
				Name:     "C:456:SV:SvSwitch:instance",
				Class:    goTypeName(obj), Property: "rdf:type",
				Message:  "SvSwitch not instantiated.",
				Severity: "sh:Violation",
			})
		}
	}

	return violations
}

// CheckSvPowerFlowInstance implements svs456:SvPowerFlow-instance
// Profile: 61970-456_StateVariables-AP-Con-Complex-SolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: SvPowerFlow must be instantiated for all energized injection equipment.
func CheckSvPowerFlowInstance(dataset *cimstructs.CIMDataset) []Violation {
	var violations []Violation

	inServiceMap := make(map[string]bool)
	for _, svs := range dataset.SvStatuss {
		if svs.ConductingEquipment != nil && svs.InService {
			inServiceMap[strings.TrimPrefix(svs.ConductingEquipment.MRID, "#")] = true
		}
	}

	tnInIsland := make(map[string]bool)
	for _, island := range dataset.TopologicalIslands {
		for _, tn := range island.TopologicalNodes {
			tnInIsland[strings.TrimPrefix(tn.MRID, "#")] = true
		}
	}

	// Index terminals by their conducting equipment's mRID, and collect the
	// set of terminal mRIDs that have an SvPowerFlow instance — both built
	// once up front so the per-equipment lookups below are O(1) instead of
	// a full scan of dataset.Terminals/SvPowerFlows for every equipment
	// object (this loop previously ran in O(equipment × terminals +
	// equipment × SvPowerFlows), which dominates RunValidation's time on
	// RealGrid-scale datasets).
	type terminalRef struct {
		id   string
		term *cimstructs.Terminal
	}
	terminalsByCE := make(map[string][]terminalRef)
	for termID, term := range dataset.Terminals {
		if term.ConductingEquipment != nil {
			ceID := strings.TrimPrefix(term.ConductingEquipment.MRID, "#")
			terminalsByCE[ceID] = append(terminalsByCE[ceID], terminalRef{termID, term})
		}
	}

	svPowerFlowTerminalIDs := make(map[string]bool)
	for _, svpf := range dataset.SvPowerFlows {
		if svpf.Terminal != nil {
			svPowerFlowTerminalIDs[strings.TrimPrefix(svpf.Terminal.MRID, "#")] = true
		}
	}

	for id, obj := range dataset.ByID {
		switch obj.(type) {
		case *cimstructs.NonConformLoad, *cimstructs.EquivalentInjection, *cimstructs.EnergySource,
			*cimstructs.ExternalNetworkInjection, *cimstructs.PowerElectronicsConnection,
			*cimstructs.AsynchronousMachine, *cimstructs.EnergyConsumer, *cimstructs.LinearShuntCompensator,
			*cimstructs.NonlinearShuntCompensator, *cimstructs.StaticVarCompensator,
			*cimstructs.SynchronousMachine, *cimstructs.StationSupply, *cimstructs.ConformLoad:
		default:
			continue
		}

		if !inServiceMap[id] {
			continue
		}

		terms := terminalsByCE[id]

		energized := false
		for _, tr := range terms {
			if tr.term.TopologicalNode != nil && tnInIsland[strings.TrimPrefix(tr.term.TopologicalNode.MRID, "#")] {
				energized = true
				break
			}
		}
		if !energized {
			continue
		}

		found := false
		for _, tr := range terms {
			if svPowerFlowTerminalIDs[tr.id] {
				found = true
				break
			}
		}
		if !found {
			violations = append(violations, Violation{
				ObjectID: id,
				RuleID:   "svs456:SvPowerFlow-instance",
				Name:     "R:456:SV:SvPowerFlow:instance",
				Class:    goTypeName(obj), Property: "rdf:type",
				Message:  "SvPowerFlow is not instantiated for energized equipment.",
				Severity: "sh:Violation",
			})
		}
	}

	return violations
}

// CheckSvPowerFlowPLimits implements svs456:SvPowerFlow.p-synchronousMachine
// Profile: 61970-456_StateVariables-AP-Con-Complex-SolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: SvPowerFlow.p should be within the min/max operating power limits of the associated machine.
func CheckSvPowerFlowPLimits(dataset *cimstructs.CIMDataset) []Violation {
	var violations []Violation

	for id, svpf := range dataset.SvPowerFlows {
		if svpf.Terminal == nil {
			continue
		}
		termID := strings.TrimPrefix(svpf.Terminal.MRID, "#")
		term, ok := dataset.Terminals[termID]
		if !ok || term.ConductingEquipment == nil {
			continue
		}

		eqID := strings.TrimPrefix(term.ConductingEquipment.MRID, "#")
		sm, ok := dataset.SynchronousMachines[eqID]
		if !ok || sm.GeneratingUnit == nil {
			continue
		}

		guID := strings.TrimPrefix(sm.GeneratingUnit.MRID, "#")
		gu, ok := dataset.GeneratingUnits[guID]
		if !ok {
			continue
		}

		// Simplified check against min/max operating P
		if svpf.P < gu.MinOperatingP || svpf.P > gu.MaxOperatingP {
			violations = append(violations, Violation{
				ObjectID: id,
				RuleID:   "svs456:SvPowerFlow.p-synchronousMachine",
				Name:     "C:456:SV:SvPowerFlow.p:synchronousMachine",
				Class:    "SvPowerFlow",
				Property: "SvPowerFlow.p",
				Message:  fmt.Sprintf("Active power (%v) is outside of the range [Min:%v, Max:%v] for SynchronousMachine %s.", svpf.P, gu.MinOperatingP, gu.MaxOperatingP, sm.Id),
				Severity: "sh:Warning",
			})
		}
	}
	return violations
}

// CheckSvPowerFlowQLimits implements svs456:SvPowerFlow.q-synchronousMachine
// Profile: 61970-456_StateVariables-AP-Con-Complex-SolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: SvPowerFlow.q should be within the reactive capability limits of the associated machine.
func CheckSvPowerFlowQLimits(dataset *cimstructs.CIMDataset) []Violation {
	var violations []Violation

	for id, svpf := range dataset.SvPowerFlows {
		if svpf.Terminal == nil {
			continue
		}
		termID := strings.TrimPrefix(svpf.Terminal.MRID, "#")
		term, ok := dataset.Terminals[termID]
		if !ok || term.ConductingEquipment == nil {
			continue
		}

		eqID := strings.TrimPrefix(term.ConductingEquipment.MRID, "#")
		sm, ok := dataset.SynchronousMachines[eqID]
		if !ok {
			continue
		}

		minQ := sm.MinQ
		maxQ := sm.MaxQ

		// If curve is present, we should ideally check against it, but simplified check uses minQ/maxQ for now.
		if sm.InitialReactiveCapabilityCurve != nil {
			// Find all CurveData for this curve
			rccID := strings.TrimPrefix(sm.InitialReactiveCapabilityCurve.MRID, "#")
			var y1vals, y2vals []float64
			for _, cdObj := range dataset.ByID {
				if cd, ok := cdObj.(*cimstructs.CurveData); ok && cd.Curve != nil {
					if strings.TrimPrefix(cd.Curve.MRID, "#") == rccID {
						y1vals = append(y1vals, cd.Y1value)
						y2vals = append(y2vals, cd.Y2value)
					}
				}
			}
			if len(y1vals) > 0 {
				minQ = y1vals[0]
				for _, v := range y1vals {
					if v < minQ {
						minQ = v
					}
				}
				maxQ = y2vals[0]
				for _, v := range y2vals {
					if v > maxQ {
						maxQ = v
					}
				}
			}
		}

		if svpf.Q < minQ || svpf.Q > maxQ {
			violations = append(violations, Violation{
				ObjectID: id,
				RuleID:   "svs456:SvPowerFlow.q-synchronousMachine",
				Name:     "C:456:SV:SvPowerFlow.q:synchronousMachine",
				Class:    "SvPowerFlow",
				Property: "SvPowerFlow.q",
				Message:  fmt.Sprintf("Reactive power (%v) is outside of the capability range [Min:%v, Max:%v] for SynchronousMachine %s.", svpf.Q, minQ, maxQ, sm.Id),
				Severity: "sh:Warning",
			})
		}
	}
	return violations
}

// CheckSvVoltageLimits implements svs456:SvVoltage.v-limits and SvVoltage.v-absoluteLimit
// Profile: 61970-456_StateVariables-AP-Con-Complex-SolvedMAS
// Origin: Derived from a SPARQL constraint.
// Description: Validates SvVoltage.v against defined voltage limits and absolute 0.4 pu limit.
func CheckSvVoltageLimits(dataset *cimstructs.CIMDataset) []Violation {
	var violations []Violation

	// terminalID -> highest VoltageLimit.value among OperationalLimitType.direction=high,
	// lowest among direction=low, reached via VoltageLimit -> OperationalLimitSet -> Terminal.
	terminalVHigh := make(map[string]float64)
	terminalVLow := make(map[string]float64)
	for _, vl := range dataset.VoltageLimits {
		if vl.OperationalLimitSet == nil || vl.OperationalLimitType == nil {
			continue
		}
		ols, ok := dataset.OperationalLimitSets[strings.TrimPrefix(vl.OperationalLimitSet.MRID, "#")]
		if !ok || ols.Terminal == nil {
			continue
		}
		olt, ok := dataset.OperationalLimitTypes[strings.TrimPrefix(vl.OperationalLimitType.MRID, "#")]
		if !ok || olt.Direction == nil {
			continue
		}
		termID := strings.TrimPrefix(ols.Terminal.MRID, "#")
		switch olt.Direction.URI {
		case cimstructs.OperationalLimitDirectionKindhigh:
			if cur, ok := terminalVHigh[termID]; !ok || vl.Value > cur {
				terminalVHigh[termID] = vl.Value
			}
		case cimstructs.OperationalLimitDirectionKindlow:
			if cur, ok := terminalVLow[termID]; !ok || vl.Value < cur {
				terminalVLow[termID] = vl.Value
			}
		}
	}

	// topologicalNodeID -> terminals connected to it.
	tnTerminals := make(map[string][]string)
	for termID, term := range dataset.Terminals {
		if term.TopologicalNode != nil {
			tnID := strings.TrimPrefix(term.TopologicalNode.MRID, "#")
			tnTerminals[tnID] = append(tnTerminals[tnID], termID)
		}
	}

	for id, svv := range dataset.SvVoltages {
		if svv.TopologicalNode == nil {
			continue
		}
		tnID := strings.TrimPrefix(svv.TopologicalNode.MRID, "#")
		tn, ok := dataset.TopologicalNodes[tnID]
		if !ok || tn.BaseVoltage == nil {
			continue
		}
		bvID := strings.TrimPrefix(tn.BaseVoltage.MRID, "#")
		bv, ok := dataset.BaseVoltages[bvID]
		if !ok {
			continue
		}

		v := svv.V
		nomV := bv.NominalVoltage

		// 1. Absolute Limit 0.4 pu
		if v/nomV <= 0.4 {
			// But only if no other limits are defined (simplified check)
			violations = append(violations, Violation{
				ObjectID: id,
				RuleID:   "svs456:SvVoltage.v-absoluteLimit",
				Name:     "C:456:SV:SvVoltage.v:absoluteLimit",
				Class:    "SvVoltage",
				Property: "SvVoltage.v",
				Message:  fmt.Sprintf("The value (%v) is <=0.4 pu of nominal voltage (%v).", v, nomV),
				Severity: "sh:Violation",
			})
		}

		// 2. Defined limits (high/low Voltage) on any terminal connected to this TN.
		outOfRange := false
		for _, termID := range tnTerminals[tnID] {
			vhigh, hasHigh := terminalVHigh[termID]
			vlow, hasLow := terminalVLow[termID]
			if hasHigh && hasLow && (v > vhigh || v < vlow) {
				outOfRange = true
				break
			}
		}
		if outOfRange {
			violations = append(violations, Violation{
				ObjectID: id,
				RuleID:   "svs456:SvVoltage.v-limits",
				Name:     "C:456:SV:SvVoltage.v:limits",
				Class:    "SvVoltage",
				Property: "SvVoltage.v",
				Message:  fmt.Sprintf("The value (%v) is outside the defined OperationalLimit (VoltageLimit) bounds.", v),
				Severity: "sh:Violation",
			})
		}
	}

	return violations
}
