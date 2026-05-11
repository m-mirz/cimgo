package validation

import (
	"cimgo/cimgostructs"
	"fmt"
	"math"
	"strings"
)

// xrRatioThreshold is the threshold above which an ACLineSegment.x/r ratio is flagged.
const xrRatioThreshold = 50.0

// regulatingControlVoltageDevWarning is the lower bound (inclusive) for a target-voltage deviation warning.
const regulatingControlVoltageDevWarning = 0.10

// ref strips the leading "#" from an rdf:resource MRID before use as a map key.
func ref(mrid string) string { return strings.TrimPrefix(mrid, "#") }

// ValidateCIMdeskQualityChecks runs modeling quality checks derived from CIMdesk behaviour.
// These checks are not encoded in the CGMES SHACL TTL files.
func ValidateCIMdeskQualityChecks(dataset *cimgostructs.CIMElementList) []Violation {
	var v []Violation
	v = append(v, CheckNoTapChangerControls(dataset)...)
	v = append(v, CheckNoRegulatingControls(dataset)...)
	v = append(v, CheckNoShuntCompensators(dataset)...)
	v = append(v, CheckSubstationHasNoVoltageLevels(dataset)...)
	v = append(v, CheckControlAreaHasNoChildren(dataset)...)
	v = append(v, CheckNoLocationsForConductors(dataset)...)
	v = append(v, CheckACLineSegmentXRRatio(dataset)...)
	v = append(v, CheckBaseVoltageDuplicateNominalVoltage(dataset)...)
	v = append(v, CheckPowerTransformerEndsSameNominalVoltage(dataset)...)
	v = append(v, CheckConnectivityNodeOpenEnded(dataset)...)
	v = append(v, CheckDisconnectorCrossVoltageLevel(dataset)...)
	v = append(v, CheckConformLoadCrossContainer(dataset)...)
	v = append(v, CheckRegulatingControlTargetVoltageMismatch(dataset)...)
	return v
}

// CheckNoTapChangerControls fires when PowerTransformers are present but no TapChangerControls exist.
func CheckNoTapChangerControls(dataset *cimgostructs.CIMElementList) []Violation {
	if len(dataset.TapChangerControls) > 0 || len(dataset.PowerTransformers) == 0 {
		return nil
	}
	return []Violation{{
		ObjectID: "global",
		Class:    "PowerTransformer",
		Property: "RegulatingControl",
		Message:  "No TapChangerControls are found. None of the PowerTransformers are used for voltage regulation.",
		Severity: "sh:Warning",
	}}
}

// CheckNoRegulatingControls fires when voltage-regulating equipment is present but no
// RegulatingControls exist. The check covers SynchronousMachine, LinearShuntCompensator,
// NonlinearShuntCompensator, and StaticVarCompensator.
func CheckNoRegulatingControls(dataset *cimgostructs.CIMElementList) []Violation {
	hasRC := len(dataset.RegulatingControls)+len(dataset.TapChangerControls) > 0
	hasEquip := len(dataset.SynchronousMachines)+
		len(dataset.LinearShuntCompensators)+
		len(dataset.NonlinearShuntCompensators)+
		len(dataset.StaticVarCompensators) > 0
	if hasRC || !hasEquip {
		return nil
	}
	return []Violation{{
		ObjectID: "global",
		Class:    "RegulatingControl",
		Property: "rdf:type",
		Message:  "No RegulatingControls are found. None of the RegulatingCondEqs (SynchronousMachine, ShuntCompensator, StaticVarCompensator) are used for voltage regulation.",
		Severity: "sh:Warning",
	}}
}

// CheckNoShuntCompensators fires when no LinearShuntCompensator or NonlinearShuntCompensator
// objects are present in the dataset.
func CheckNoShuntCompensators(dataset *cimgostructs.CIMElementList) []Violation {
	if len(dataset.LinearShuntCompensators)+len(dataset.NonlinearShuntCompensators) > 0 {
		return nil
	}
	// Only fire when there is actual EQ content.
	if len(dataset.PowerTransformers) == 0 && len(dataset.ACLineSegments) == 0 {
		return nil
	}
	return []Violation{{
		ObjectID: "global",
		Class:    "ShuntCompensator",
		Property: "rdf:type",
		Message:  "No ShuntCompensator objects (LinearShuntCompensator, NonlinearShuntCompensator) are found; at least one is expected.",
		Severity: "sh:Warning",
	}}
}

// CheckSubstationHasNoVoltageLevels fires for each Substation that has neither a child
// VoltageLevel nor a child ConnectivityNode. Boundary substations (EQBD) have ConnectivityNodes
// but no VoltageLevels and are intentionally excluded.
func CheckSubstationHasNoVoltageLevels(dataset *cimgostructs.CIMElementList) []Violation {
	hasVL := make(map[string]struct{})
	for _, vl := range dataset.VoltageLevels {
		if vl.Substation != nil {
			hasVL[ref(vl.Substation.MRID)] = struct{}{}
		}
	}
	hasCN := make(map[string]struct{})
	for _, cn := range dataset.ConnectivityNodes {
		if cn.ConnectivityNodeContainer != nil {
			hasCN[ref(cn.ConnectivityNodeContainer.MRID)] = struct{}{}
		}
	}
	var violations []Violation
	for id, sub := range dataset.Substations {
		if _, ok := hasVL[id]; ok {
			continue
		}
		if _, ok := hasCN[id]; ok {
			continue // boundary substation — it has CNs but no VoltageLevels
		}
		violations = append(violations, Violation{
			ObjectID: id,
			Class:    "Substation",
			Name:     sub.Name,
			Property: "VoltageLevel",
			Message:  "The Substation has no child VoltageLevels and is not referenced by any instance.",
			Severity: "sh:Warning",
		})
	}
	return violations
}

// CheckControlAreaHasNoChildren fires for each ControlArea that has neither
// ControlAreaGeneratingUnits nor TieFlows referencing it.
func CheckControlAreaHasNoChildren(dataset *cimgostructs.CIMElementList) []Violation {
	hasCAGU := make(map[string]struct{})
	for _, cagu := range dataset.ControlAreaGeneratingUnits {
		if cagu.ControlArea != nil {
			hasCAGU[ref(cagu.ControlArea.MRID)] = struct{}{}
		}
	}
	hasTF := make(map[string]struct{})
	for _, tf := range dataset.TieFlows {
		if tf.ControlArea != nil {
			hasTF[ref(tf.ControlArea.MRID)] = struct{}{}
		}
	}
	var violations []Violation
	for id, ca := range dataset.ControlAreas {
		if _, ok := hasCAGU[id]; ok {
			continue
		}
		if _, ok := hasTF[id]; ok {
			continue
		}
		violations = append(violations, Violation{
			ObjectID: id,
			Class:    "ControlArea",
			Name:     ca.Name,
			Property: "ControlAreaGeneratingUnit",
			Message:  "The ControlArea has no child instances (no ControlAreaGeneratingUnits and no TieFlows reference it).",
			Severity: "sh:Warning",
		})
	}
	return violations
}

// CheckNoLocationsForConductors fires for each ACLineSegment or DCLineSegment that has no
// Location pointing to it. Locations live in the GL profile; if none are loaded the check is skipped.
func CheckNoLocationsForConductors(dataset *cimgostructs.CIMElementList) []Violation {
	if len(dataset.Locations) == 0 {
		return nil
	}
	covered := make(map[string]struct{})
	for _, loc := range dataset.Locations {
		if loc.PowerSystemResources != nil {
			covered[ref(loc.PowerSystemResources.MRID)] = struct{}{}
		}
	}
	var violations []Violation
	for id, seg := range dataset.ACLineSegments {
		if _, ok := covered[id]; ok {
			continue
		}
		violations = append(violations, Violation{
			ObjectID: id,
			Class:    "ACLineSegment",
			Name:     seg.Name,
			Property: "Location",
			Message:  "No Location is associated with this ACLineSegment.",
			Severity: "sh:Warning",
		})
	}
	for id, seg := range dataset.DCLineSegments {
		if _, ok := covered[id]; ok {
			continue
		}
		violations = append(violations, Violation{
			ObjectID: id,
			Class:    "DCLineSegment",
			Name:     seg.Name,
			Property: "Location",
			Message:  "No Location is associated with this DCLineSegment.",
			Severity: "sh:Warning",
		})
	}
	return violations
}

// CheckACLineSegmentXRRatio fires when ACLineSegment.x / ACLineSegment.r exceeds xrRatioThreshold.
// Zero resistance (r == 0) is skipped; those are handled by the SHACL r-value range check.
func CheckACLineSegmentXRRatio(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	for id, seg := range dataset.ACLineSegments {
		if seg.R == 0 || seg.X == 0 {
			continue
		}
		ratio := seg.X / seg.R
		if ratio > xrRatioThreshold {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "ACLineSegment",
				Name:     seg.Name,
				Property: "ACLineSegment.x",
				Message:  fmt.Sprintf("ACLineSegment.x/ACLineSegment.r ratio (%.4g) exceeds the threshold of %g.", ratio, xrRatioThreshold),
				Severity: "sh:Warning",
			})
		}
	}
	return violations
}

// CheckBaseVoltageDuplicateNominalVoltage fires when two or more BaseVoltage objects share
// the same nominalVoltage value.
func CheckBaseVoltageDuplicateNominalVoltage(dataset *cimgostructs.CIMElementList) []Violation {
	byVoltage := make(map[float64][]string)
	for id, bv := range dataset.BaseVoltages {
		byVoltage[bv.NominalVoltage] = append(byVoltage[bv.NominalVoltage], id)
	}
	var violations []Violation
	for voltage, ids := range byVoltage {
		if len(ids) < 2 {
			continue
		}
		for _, id := range ids {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "BaseVoltage",
				Property: "BaseVoltage.nominalVoltage",
				Message:  fmt.Sprintf("BaseVoltage.nominalVoltage (%.4g kV) is shared by %d BaseVoltage instances.", voltage, len(ids)),
				Severity: "sh:Warning",
			})
		}
	}
	return violations
}

// CheckPowerTransformerEndsSameNominalVoltage fires when all PowerTransformerEnds of a
// PowerTransformer have the same ratedU value (no voltage transformation).
func CheckPowerTransformerEndsSameNominalVoltage(dataset *cimgostructs.CIMElementList) []Violation {
	endsByPT := make(map[string][]*cimgostructs.PowerTransformerEnd)
	for _, end := range dataset.PowerTransformerEnds {
		if end.PowerTransformer == nil {
			continue
		}
		ptID := ref(end.PowerTransformer.MRID)
		endsByPT[ptID] = append(endsByPT[ptID], end)
	}
	var violations []Violation
	for ptID, ends := range endsByPT {
		if len(ends) < 2 {
			continue
		}
		ref0 := ends[0].RatedU
		if ref0 == 0 {
			continue
		}
		allSame := true
		for _, end := range ends[1:] {
			if end.RatedU != ref0 {
				allSame = false
				break
			}
		}
		if allSame {
			pt := dataset.PowerTransformers[ptID]
			name := ""
			if pt != nil {
				name = pt.Name
			}
			violations = append(violations, Violation{
				ObjectID: ptID,
				Class:    "PowerTransformer",
				Name:     name,
				Property: "PowerTransformerEnd.ratedU",
				Message:  fmt.Sprintf("All PowerTransformerEnds have the same ratedU (%.4g kV); no voltage transformation occurs.", ref0),
				Severity: "sh:Warning",
			})
		}
	}
	return violations
}

// CheckConnectivityNodeOpenEnded fires for each ConnectivityNode that has exactly one Terminal.
func CheckConnectivityNodeOpenEnded(dataset *cimgostructs.CIMElementList) []Violation {
	count := make(map[string]int)
	for _, t := range dataset.Terminals {
		if t.ConnectivityNode != nil {
			count[ref(t.ConnectivityNode.MRID)]++
		}
	}
	var violations []Violation
	for id, cn := range dataset.ConnectivityNodes {
		if count[id] == 1 {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "ConnectivityNode",
				Name:     cn.Name,
				Property: "Terminal",
				Message:  "The ConnectivityNode is open-ended: only one Terminal is connected to it.",
				Severity: "sh:Warning",
			})
		}
	}
	return violations
}

// CheckDisconnectorCrossVoltageLevel fires for each Disconnector whose Terminals connect
// ConnectivityNodes that belong to two different VoltageLevels (both containers must be
// VoltageLevels; connections to Bays, boundary Substations, etc. are excluded).
func CheckDisconnectorCrossVoltageLevel(dataset *cimgostructs.CIMElementList) []Violation {
	// Only include ConnectivityNodes whose container is a VoltageLevel.
	cnVoltageLevel := make(map[string]string) // CN MRID → VoltageLevel MRID
	for id, cn := range dataset.ConnectivityNodes {
		if cn.ConnectivityNodeContainer == nil {
			continue
		}
		contID := ref(cn.ConnectivityNodeContainer.MRID)
		if _, isVL := dataset.VoltageLevels[contID]; isVL {
			cnVoltageLevel[id] = contID
		}
	}
	termsByEquip := make(map[string][]*cimgostructs.Terminal)
	for _, t := range dataset.Terminals {
		if t.ConductingEquipment != nil {
			termsByEquip[ref(t.ConductingEquipment.MRID)] = append(termsByEquip[ref(t.ConductingEquipment.MRID)], t)
		}
	}
	var violations []Violation
	for id, disc := range dataset.Disconnectors {
		terms := termsByEquip[id]
		if len(terms) < 2 {
			continue
		}
		vlIDs := make(map[string]struct{})
		for _, t := range terms {
			if t.ConnectivityNode != nil {
				if vl, ok := cnVoltageLevel[ref(t.ConnectivityNode.MRID)]; ok {
					vlIDs[vl] = struct{}{}
				}
			}
		}
		// Both terminals must be in VoltageLevels and they must be different.
		if len(vlIDs) > 1 {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "Disconnector",
				Name:     disc.Name,
				Property: "Terminal.ConnectivityNode",
				Message:  "The two ConnectivityNodes the Disconnector connects are in different VoltageLevels.",
				Severity: "sh:Warning",
			})
		}
	}
	return violations
}

// CheckConformLoadCrossContainer fires for each ConformLoad whose EquipmentContainer differs
// from the ConnectivityNodeContainer of its connected Terminal's ConnectivityNode.
func CheckConformLoadCrossContainer(dataset *cimgostructs.CIMElementList) []Violation {
	cnContainer := make(map[string]string)
	for id, cn := range dataset.ConnectivityNodes {
		if cn.ConnectivityNodeContainer != nil {
			cnContainer[id] = ref(cn.ConnectivityNodeContainer.MRID)
		}
	}
	termsByEquip := make(map[string][]*cimgostructs.Terminal)
	for _, t := range dataset.Terminals {
		if t.ConductingEquipment != nil {
			termsByEquip[ref(t.ConductingEquipment.MRID)] = append(termsByEquip[ref(t.ConductingEquipment.MRID)], t)
		}
	}
	var violations []Violation
	for id, load := range dataset.ConformLoads {
		if load.EquipmentContainer == nil {
			continue
		}
		equipContainer := ref(load.EquipmentContainer.MRID)
		for _, t := range termsByEquip[id] {
			if t.ConnectivityNode == nil {
				continue
			}
			cnCont, ok := cnContainer[ref(t.ConnectivityNode.MRID)]
			if !ok {
				continue
			}
			if cnCont != equipContainer {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "ConformLoad",
					Name:     load.Name,
					Property: "EquipmentContainer",
					Message:  "The ConformLoad and its connected TopologicalNodes are not contained by the same EquipmentContainer.",
					Severity: "sh:Warning",
				})
				break
			}
		}
	}
	return violations
}

// CheckRegulatingControlTargetVoltageMismatch fires for each voltage-mode RegulatingControl
// whose targetValue deviates 10 % or more from the nominalVoltage of the regulated node.
func CheckRegulatingControlTargetVoltageMismatch(dataset *cimgostructs.CIMElementList) []Violation {
	// CN → nominalVoltage (kV): CN → VoltageLevel → BaseVoltage.
	cnNominalKV := make(map[string]float64)
	for cnID, cn := range dataset.ConnectivityNodes {
		if cn.ConnectivityNodeContainer == nil {
			continue
		}
		vl, ok := dataset.VoltageLevels[ref(cn.ConnectivityNodeContainer.MRID)]
		if !ok || vl.BaseVoltage == nil {
			continue
		}
		bv, ok := dataset.BaseVoltages[ref(vl.BaseVoltage.MRID)]
		if !ok {
			continue
		}
		cnNominalKV[cnID] = bv.NominalVoltage
	}

	termCN := make(map[string]string) // Terminal MRID → CN MRID
	for id, t := range dataset.Terminals {
		if t.ConnectivityNode != nil {
			termCN[id] = ref(t.ConnectivityNode.MRID)
		}
	}

	var violations []Violation
	for id, obj := range dataset.Elements {
		rc, ok := obj.(*cimgostructs.RegulatingControl)
		if !ok {
			continue
		}
		if rc.Mode == nil || rc.Mode.URI != cimgostructs.RegulatingControlModeKindvoltage {
			continue
		}
		if rc.Terminal == nil {
			continue
		}
		cnID, ok := termCN[ref(rc.Terminal.MRID)]
		if !ok {
			continue
		}
		nominalKV, ok := cnNominalKV[cnID]
		if !ok || nominalKV == 0 {
			continue
		}
		targetKV := applyUnitMultiplier(rc.TargetValue, rc.TargetValueUnitMultiplier)
		if targetKV == 0 {
			continue
		}
		deviation := math.Abs(targetKV-nominalKV) / nominalKV
		if deviation < regulatingControlVoltageDevWarning {
			continue
		}
		violations = append(violations, Violation{
			ObjectID: id,
			Class:    "RegulatingControl",
			Name:     rc.Name,
			Property: "RegulatingControl.targetValue",
			Message: fmt.Sprintf(
				"RegulatingControl target voltage (%.4g kV) deviates %.1f%% from the nominal voltage (%.4g kV) of the regulated node.",
				targetKV, deviation*100, nominalKV,
			),
			Severity: "sh:Warning",
		})
	}
	return violations
}

// applyUnitMultiplier converts targetValue to kV using the UnitMultiplier URI suffix.
// Voltage-control targetValues default to kV when no multiplier is set.
func applyUnitMultiplier(value float64, mult *struct{ URI string `xml:"resource,attr"` }) float64 {
	if mult == nil {
		return value
	}
	suffix := mult.URI
	if idx := strings.LastIndexAny(suffix, "#."); idx >= 0 {
		suffix = suffix[idx+1:]
	}
	switch suffix {
	case "M":
		return value * 1000
	case "G":
		return value * 1_000_000
	case "m":
		return value / 1000
	default: // "k" or unrecognised — already kV
		return value
	}
}
