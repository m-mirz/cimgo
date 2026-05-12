package validation

import "cimgo/cimgostructs"

// ValidateSSHProfileSPARQL runs hand-written checks for
// 61970-301_SteadyStateHypothesis-AP-Con-Complex-SHACL and
// 61970-456_SteadyStateHypothesis-AP-Con-Complex-SHACL.
func ValidateSSHProfileSPARQL(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckEnergySourceActivePowerConsumer(dataset)...)
	violations = append(violations, CheckRegulatingControlTargetDeadbandApplicability(dataset)...)
	violations = append(violations, CheckCsConverterValueRange(dataset)...)
	violations = append(violations, CheckCsConverterPPccControl(dataset)...)
	violations = append(violations, CheckVsConverterPPccControl(dataset)...)
	violations = append(violations, CheckVsConverterQPccControl(dataset)...)
	violations = append(violations, CheckEnergySourcePQ(dataset)...)
	violations = append(violations, CheckSynchronousMachineOperatingModeMatch(dataset)...)
	violations = append(violations, CheckGeneratingUnitSingleActivePowerSlack(dataset)...)
	violations = append(violations, CheckExternalNetworkInjectionLimits(dataset)...)
	violations = append(violations, CheckEquivalentInjectionLimits(dataset)...)
	violations = append(violations, CheckRotatingMachineCurveLimits(dataset)...)
	violations = append(violations, CheckRegulatingControlTargetValuePositive(dataset)...)
	return violations
}

// CheckEnergySourceActivePowerConsumer implements sshc.EnergySource.activePower-consumer
// Profile: 61970-301_SteadyStateHypothesis-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Load sign convention is used, i.e. positive sign means flow out from a node.
// Warning if EnergySource is a consumer (activePower > 0).
func CheckEnergySourceActivePowerConsumer(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		es, ok := obj.(*cimgostructs.EnergySource)
		if !ok {
			continue
		}

		if es.ActivePower > 0 {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "EnergySource",
				Property: "EnergySource.activePower",
				Message:  "EnergySource that is a consumer (activePower > 0).",
				Severity: "sh:Warning",
			})
		}
	}

	return violations
}

// CheckRegulatingControlTargetDeadbandApplicability implements sshc.RegulatingControl.targetDeadband-applicability
// Profile: 61970-301_SteadyStateHypothesis-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Either RegulatingControl.targetDeadband is provided for a continuous control or it is not provided for a discrete control.
func CheckRegulatingControlTargetDeadbandApplicability(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		rc, ok := obj.(*cimgostructs.RegulatingControl)
		if !ok {
			// Also check TapChangerControl if it's separate or if it inherits
			tcc, ok := obj.(*cimgostructs.TapChangerControl)
			if !ok {
				continue
			}
			if (tcc.TargetDeadband != 0 && !tcc.Discrete) || (tcc.TargetDeadband == 0 && tcc.Discrete) {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "TapChangerControl",
					Property: "RegulatingControl.discrete",
					Message:  "Either RegulatingControl.targetDeadband is provided for a continuous control or it is not provided for a discrete control.",
					Severity: "sh:Violation",
				})
			}
			continue
		}

		if (rc.TargetDeadband != 0 && !rc.Discrete) || (rc.TargetDeadband == 0 && rc.Discrete) {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "RegulatingControl",
				Property: "RegulatingControl.discrete",
				Message:  "Either RegulatingControl.targetDeadband is provided for a continuous control or it is not provided for a discrete control.",
				Severity: "sh:Violation",
			})
		}
	}

	return violations
}

// CheckCsConverterValueRange implements sshc.CsConverter.maxAlpha/maxGamma/minAlpha/minGamma-valueRangeTypical
// Profile: 61970-301_SteadyStateHypothesis-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Validates that CsConverter firing and extinction angles are within typical ranges.
func CheckCsConverterValueRange(dataset *cimgostructs.CIMElementList) []Violation {
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
			if csc.MaxAlpha > 18 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "CsConverter",
					Property: "CsConverter.maxAlpha",
					Message:  "The maxAlpha value is greater than 18 for a rectifier.",
					Severity: "sh:Warning",
				})
			}
			if csc.MinAlpha < 10 || csc.MinAlpha > csc.MaxAlpha {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "CsConverter",
					Property: "CsConverter.minAlpha",
					Message:  "The minAlpha value is less than 10 or greater than CsConverter.maxAlpha for a rectifier.",
					Severity: "sh:Warning",
				})
			}
		} else if mode == inverter {
			if csc.MaxGamma > 20 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "CsConverter",
					Property: "CsConverter.maxGamma",
					Message:  "The maxGamma value is greater than 20 for an inverter.",
					Severity: "sh:Warning",
				})
			}
			if csc.MinGamma < 17 || csc.MinGamma > csc.MaxGamma {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "CsConverter",
					Property: "CsConverter.minGamma",
					Message:  "The minGamma value is less than 17 or greater than CsConverter.maxGamma for an inverter.",
					Severity: "sh:Warning",
				})
			}
		}
	}

	return violations
}

// CheckCsConverterPPccControl implements sshc.CsConverter.pPccControl-targetValueIdc/Udc/Ppcc
// Profile: 61970-301_SteadyStateHypothesis-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Validates required target values for CsConverter.pPccControl based on the selected control mode.
func CheckCsConverterPPccControl(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		csc, ok := obj.(*cimgostructs.CsConverter)
		if !ok || csc.PPccControl == nil {
			continue
		}

		control := csc.PPccControl.URI
		dcCurrent := "http://iec.ch/TC57/CIM100#CsPpccControlKind.dcCurrent"
		dcVoltage := "http://iec.ch/TC57/CIM100#CsPpccControlKind.dcVoltage"
		activePower := "http://iec.ch/TC57/CIM100#CsPpccControlKind.activePower"

		if control == dcCurrent && csc.TargetIdc == 0 {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "CsConverter",
				Property: "CsConverter.pPccControl",
				Message:  "CsConverter.targetIdc is not provided for a converter with CsPpccControlKind.dcCurrent.",
				Severity: "sh:Violation",
			})
		} else if control == dcVoltage && csc.TargetUdc == 0 {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "CsConverter",
				Property: "CsConverter.pPccControl",
				Message:  "ACDCConverter.targetUdc is not provided for a converter with CsPpccControlKind.dcVoltage.",
				Severity: "sh:Violation",
			})
		} else if control == activePower && csc.TargetPpcc == 0 {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "CsConverter",
				Property: "CsConverter.pPccControl",
				Message:  "ACDCConverter.targetPpcc is not provided for a converter with CsPpccControlKind.activePower.",
				Severity: "sh:Violation",
			})
		}
	}

	return violations
}

// CheckVsConverterPPccControl implements sshc.VsConverter.pPccControl rules
// Profile: 61970-301_SteadyStateHypothesis-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Validates required target values for VsConverter.pPccControl based on the selected control mode.
func CheckVsConverterPPccControl(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		vsc, ok := obj.(*cimgostructs.VsConverter)
		if !ok || vsc.PPccControl == nil {
			continue
		}

		control := vsc.PPccControl.URI
		prefix := "http://iec.ch/TC57/CIM100#VsPpccControlKind."

		switch control {
		case prefix + "pPccAndUdcDroop":
			if vsc.TargetPpcc == 0 || vsc.TargetUdc == 0 || vsc.Droop == 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "VsConverter",
					Property: "VsConverter.pPccControl",
					Message:  "One or all among ACDCConverter.targetPpcc, ACDCConverter.targetUdc and VsConverter.droop are not provided for VsPpccControlKind.pPccAndUdcDroop.",
					Severity: "sh:Violation",
				})
			}
		case prefix + "pPccAndUdcDroopWithCompensation":
			if vsc.TargetPpcc == 0 || vsc.TargetUdc == 0 || vsc.Droop == 0 || vsc.DroopCompensation == 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "VsConverter",
					Property: "VsConverter.pPccControl",
					Message:  "One or all among ACDCConverter.targetPpcc, ACDCConverter.targetUdc, VsConverter.droop and VsConverter.droopCompensation are not provided for VsPpccControlKind.pPccAndUdcDroopWithCompensation.",
					Severity: "sh:Violation",
				})
			}
		case prefix + "pPccAndUdcDroopPilot":
			if vsc.TargetPpcc == 0 || vsc.TargetUdc == 0 || vsc.Droop == 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "VsConverter",
					Property: "VsConverter.pPccControl",
					Message:  "One or all among ACDCConverter.targetPpcc, ACDCConverter.targetUdc and VsConverter.droop are not provided for VsPpccControlKind.pPccAndUdcDroopPilot.",
					Severity: "sh:Violation",
				})
			}
		case prefix + "udc":
			if vsc.TargetUdc == 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "VsConverter",
					Property: "VsConverter.pPccControl",
					Message:  "ACDCConverter.targetUdc is not provided for VsPpccControlKind.udc.",
					Severity: "sh:Violation",
				})
			}
		case prefix + "pPcc":
			if vsc.TargetPpcc == 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "VsConverter",
					Property: "VsConverter.pPccControl",
					Message:  "ACDCConverter.targetPpcc is not provided for VsPpccControlKind.pPcc.",
					Severity: "sh:Violation",
				})
			}
		case prefix + "phasePcc":
			if vsc.TargetPhasePcc == 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "VsConverter",
					Property: "VsConverter.pPccControl",
					Message:  "VsConverter.targetPhasePcc is not provided for VsPpccControlKind.phasePcc.",
					Severity: "sh:Violation",
				})
			}
		}
	}

	return violations
}

// CheckVsConverterQPccControl implements sshc.VsConverter.qPccControl rules
// Profile: 61970-301_SteadyStateHypothesis-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Validates required target values for VsConverter.qPccControl based on the selected control mode.
func CheckVsConverterQPccControl(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		vsc, ok := obj.(*cimgostructs.VsConverter)
		if !ok || vsc.QPccControl == nil {
			continue
		}

		control := vsc.QPccControl.URI
		prefix := "http://iec.ch/TC57/CIM100#VsQpccControlKind."

		switch control {
		case prefix + "powerFactorPcc":
			if vsc.TargetPowerFactorPcc == 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "VsConverter",
					Property: "VsConverter.qPccControl",
					Message:  "VsConverter.targetPowerFactorPcc is not provided for VsQpccControlKind.powerFactorPcc.",
					Severity: "sh:Violation",
				})
			}
		case prefix + "pulseWidthModulation":
			if vsc.TargetPWMfactor == 0 || vsc.TargetPhasePcc == 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "VsConverter",
					Property: "VsConverter.qPccControl",
					Message:  "VsConverter.targetPWMfactor and/or VsConverter.targetPhasePcc are not provided for VsQpccControlKind.pulseWidthModulation.",
					Severity: "sh:Violation",
				})
			}
		case prefix + "reactivePcc":
			if vsc.TargetQpcc == 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "VsConverter",
					Property: "VsConverter.qPccControl",
					Message:  "VsConverter.targetQpcc is not provided for VsQpccControlKind.reactivePcc.",
					Severity: "sh:Violation",
				})
			}
		case prefix + "voltagePcc":
			if vsc.TargetUpcc == 0 {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    "VsConverter",
					Property: "VsConverter.qPccControl",
					Message:  "VsConverter.targetUpcc is not provided for VsQpccControlKind.voltagePcc.",
					Severity: "sh:Violation",
				})
			}
		}
	}

	return violations
}

// CheckEnergySourcePQ implements sshc456:EnergySource-EnergySourcePQ
// Profile: 61970-456_SteadyStateHypothesis-AP-Con-Complex
// Origin: Derived from a manual complex constraint (described as textual condition in SHACL).
// Description: voltageAngle and voltageMagnitude shall only be used when modeling a voltage source.
func CheckEnergySourcePQ(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	for id, obj := range dataset.Elements {
		es, ok := obj.(*cimgostructs.EnergySource)
		if !ok {
			continue
		}

		if es.VoltageAngle != 0 || es.VoltageMagnitude != 0 {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    "EnergySource",
				Property: "EnergySource.voltageAngle",
				Message:  "EnergySource modelled as voltage source (attributes voltageAngle and voltageMagnitude are used). Please assess depending on the use case.",
				Severity: "sh:Warning",
			})
		}
	}

	return violations
}
