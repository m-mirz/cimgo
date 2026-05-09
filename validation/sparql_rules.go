package validation

import (
	"cimgo/cimgostructs"
	"cimgo/shaclgen"
	"cimgo/shaclmodel"
	"reflect"
)

// Violation is re-exported from shaclmodel so existing callers of
// validation.Violation keep compiling. The actual type lives in shaclmodel
// (a leaf package) so the generated shaclgen package can return it without
// pulling in validation, which would be a cycle.
type Violation = shaclmodel.Violation

func goTypeName(obj interface{}) string {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

// Hand-written profile checks. Each Validate<Profile>Profile bundles only the
// SPARQL-style cross-cutting rules that don't reduce to a single attribute
// constraint; per-attribute SHACL checks come from the generated profile
// orchestrators (see generated_index.go) and are run separately by
// ValidateAllProfiles.

// ValidateEquipmentProfile runs hand-written checks for 61970-301_Equipment-AP-Con-Complex-SHACL.
func ValidateEquipmentProfile(dataset *cimgostructs.CIMElementList) []Violation {
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

// ValidateDiagramLayoutProfile runs hand-written checks for 61970-301_DiagramLayout-AP-Con-Complex-SHACL.
func ValidateDiagramLayoutProfile(dataset *cimgostructs.CIMElementList) []Violation {
	return CheckDiagramObjectIdentifiedObjectType(dataset)
}

// ValidateTopologyNotSolvedMASProfile runs hand-written checks for
// 61970-301_Topology-AP-Con-Complex-NotSolvedMAS-SHACL and
// 61970-600_Topology-AP-Con-Complex-NotSolvedMAS-SHACL.
func ValidateTopologyNotSolvedMASProfile(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckTerminalPhasesConsistencyTopologicalNode(dataset)...)
	violations = append(violations, CheckSwitchSameTopologicalNode(dataset)...)
	violations = append(violations, CheckTerminalExch8TopologicalNode(dataset)...)
	return violations
}

// ValidateEquipmentNotSolvedMASProfile runs hand-written checks for
// 61970-301_Equipment-AP-Con-Complex-NotSolvedMAS-SHACL.
func ValidateEquipmentNotSolvedMASProfile(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckACLineSegmentBaseVoltage(dataset)...)
	violations = append(violations, CheckRegulatingControlTargetValueTapChanger(dataset)...)
	violations = append(violations, CheckACLineSegmentBaseVoltageDiff(dataset)...)
	violations = append(violations, CheckBoundaryPointBppl(dataset)...)
	violations = append(violations, CheckEquivalentInjectionRegulationCapabilityNotHVDC(dataset)...)
	return violations
}

// ValidateSSHProfile runs hand-written checks for
// 61970-301_SteadyStateHypothesis-AP-Con-Complex-SHACL and
// 61970-456_SteadyStateHypothesis-AP-Con-Complex-SHACL.
func ValidateSSHProfile(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckEnergySourceActivePowerConsumer(dataset)...)
	violations = append(violations, CheckRegulatingControlTargetDeadbandApplicability(dataset)...)
	violations = append(violations, CheckCsConverterValueRange(dataset)...)
	violations = append(violations, CheckCsConverterPPccControl(dataset)...)
	violations = append(violations, CheckVsConverterPPccControl(dataset)...)
	violations = append(violations, CheckVsConverterQPccControl(dataset)...)
	violations = append(violations, CheckEnergySourcePQ(dataset)...)
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

// ValidateSSHNotSolvedMASProfile runs hand-written checks for
// 61970-301_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS-SHACL and
// 61970-456_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS-SHACL.
func ValidateSSHNotSolvedMASProfile(dataset *cimgostructs.CIMElementList) []Violation {
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

// ValidateDynamicsProfile runs hand-written checks for
// 61970-457_Dynamics-AP-Con-Complex-SHACL and
// 61970-302_Dynamics-AP-Con-Complex-SHACL.
func ValidateDynamicsProfile(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckExcitationSystemDynamicsSynchronousMachineDynamics(dataset)...)
	violations = append(violations, CheckSynchronousMachineTimeConstantReactanceModelType(dataset)...)
	violations = append(violations, CheckTurbineGovernorMbaseEquation(dataset)...)
	violations = append(violations, CheckExcitationSystemGains(dataset)...)
	violations = append(violations, CheckPssInputSignals(dataset)...)
	violations = append(violations, CheckGovHydro4GainPoints(dataset)...)
	violations = append(violations, CheckLoadStaticModelAttributes(dataset)...)
	violations = append(violations, CheckRotatingMachineSaturation(dataset)...)
	violations = append(violations, CheckSynchronousMachineSimplifiedAttributes(dataset)...)
	violations = append(violations, CheckDynamicsAssociations(dataset)...)
	return violations
}
// ValidateShortCircuitProfile runs hand-written checks for 61970-301_ShortCircuit-AP-Con-Complex-SHACL.
func ValidateShortCircuitProfile(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckSeriesCompensatorVaristorUsage(dataset)...)
	violations = append(violations, CheckTransformerEndGrounding(dataset)...)
	violations = append(violations, CheckSynchronousMachineEarthing(dataset)...)
	violations = append(violations, CheckSeriesCompensatorVaristorRequired(dataset)...)
	return violations
}

// ValidateShortCircuitNotSolvedMASProfile runs hand-written checks for
// 61970-301_ShortCircuit-AP-Con-Complex-NotSolvedMAS-SHACL.
func ValidateShortCircuitNotSolvedMASProfile(dataset *cimgostructs.CIMElementList) []Violation {
	return CheckMutualCouplingTerminalsAssignment(dataset)
}

// ValidateStateVariablesProfile runs hand-written checks for
// 61970-301_StateVariables-AP-Con-Complex-SHACL and
// 61970-456_StateVariables-AP-Con-Complex-SHACL.
func ValidateStateVariablesProfile(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckCsConverterStateValueRange(dataset)...)
	violations = append(violations, CheckTopologicalIslandCount(dataset)...)
	return violations
}

// ValidateStateVariablesSolvedMASProfile runs hand-written checks for
// 61970-301_StateVariables-AP-Con-Complex-SolvedMAS-SHACL,
// 61970-456_StateVariables-AP-Con-Complex-SolvedMAS-SHACL and
// 61970-600-1_AllProfiles-AP-Con-Complex-SolvedMAS-SHACL.
func ValidateStateVariablesSolvedMASProfile(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckSvTapStepPositionRange(dataset)...)
	violations = append(violations, CheckSvTapStepPositionInteger(dataset)...)
	violations = append(violations, CheckSvTapStepPositionSync(dataset)...)
	violations = append(violations, CheckSvShuntCompensatorSectionsInteger(dataset)...)
	violations = append(violations, CheckSvShuntCompensatorSectionsSync(dataset)...)
	violations = append(violations, CheckAngleReference(dataset)...)
	violations = append(violations, CheckStateVariablesInstantiated(dataset)...)
	violations = append(violations, CheckSvStateVariablesInstance(dataset)...)
	violations = append(violations, CheckSvPowerFlowPLimits(dataset)...)
	violations = append(violations, CheckSvPowerFlowQLimits(dataset)...)
	violations = append(violations, CheckSvVoltageLimits(dataset)...)
	violations = append(violations, CheckRegulatingControlContradictory(dataset)...)
	violations = append(violations, CheckRegulatingControlSameIsland(dataset)...)
	return violations
}

// ValidateCommonRules runs hand-written checks for common rules (all600, io).
func ValidateCommonRules(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckFileHeaderExists(dataset)...)
	violations = append(violations, CheckMRIDUniqueness(dataset)...)
	violations = append(violations, CheckIDUUID(dataset)...)
	violations = append(violations, CheckIDDeprecated(dataset)...)
	violations = append(violations, CheckModelDateTimeUTC(dataset)...)
	violations = append(violations, CheckFloatSpecialValues(dataset)...)
	violations = append(violations, CheckModelingAuthoritySetNotEmpty(dataset)...)
	violations = append(violations, CheckIdentifiedObjectStringLengths(dataset)...)
	violations = append(violations, CheckDanglingReferences(dataset)...)
	return violations
}

// ValidateEquipmentBoundaryProfile runs hand-written checks for 61970-301_EquipmentBoundary.
func ValidateEquipmentBoundaryProfile(dataset *cimgostructs.CIMElementList) []Violation {
	return CheckBoundaryPointTieFlow(dataset)
}

// ValidateOperationProfile runs hand-written checks for 61970-301_Operation.
func ValidateOperationProfile(dataset *cimgostructs.CIMElementList) []Violation {
	return CheckMeasurementTerminalRequiredCases(dataset)
}

// ValidateAllProfiles runs every generated SHACL profile orchestrator plus
// every hand-written profile-level check. The two are independent: the
// generated set comes from shaclgen.ValidateAllGeneratedProfiles; the
// hand-written ones are the Validate<Profile>Profile functions above.
func ValidateAllProfiles(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, shaclgen.ValidateAllGeneratedProfiles(dataset)...)
	violations = append(violations, ValidateCommonRules(dataset)...)
	violations = append(violations, ValidateEquipmentProfile(dataset)...)
	violations = append(violations, ValidateEquipmentNotSolvedMASProfile(dataset)...)
	violations = append(violations, ValidateSSHProfile(dataset)...)
	violations = append(violations, ValidateSSHNotSolvedMASProfile(dataset)...)
	violations = append(violations, ValidateDynamicsProfile(dataset)...)
	violations = append(violations, ValidateShortCircuitProfile(dataset)...)
	violations = append(violations, ValidateShortCircuitNotSolvedMASProfile(dataset)...)
	violations = append(violations, ValidateStateVariablesProfile(dataset)...)
	violations = append(violations, ValidateStateVariablesSolvedMASProfile(dataset)...)
	violations = append(violations, ValidateDiagramLayoutProfile(dataset)...)
	violations = append(violations, ValidateTopologyNotSolvedMASProfile(dataset)...)
	violations = append(violations, ValidateEquipmentBoundaryProfile(dataset)...)
	violations = append(violations, ValidateOperationProfile(dataset)...)
	return violations
}
