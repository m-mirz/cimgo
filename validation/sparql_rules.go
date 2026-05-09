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
	violations = append(violations, CheckLoadResponseCharacteristicSum(dataset)...)
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
	return violations
}

// ValidateDiagramLayoutProfile runs hand-written checks for 61970-301_DiagramLayout-AP-Con-Complex-SHACL.
func ValidateDiagramLayoutProfile(dataset *cimgostructs.CIMElementList) []Violation {
	return CheckDiagramObjectIdentifiedObjectType(dataset)
}

// ValidateTopologyNotSolvedMASProfile runs hand-written checks for
// 61970-301_Topology-AP-Con-Complex-NotSolvedMAS-SHACL.
func ValidateTopologyNotSolvedMASProfile(dataset *cimgostructs.CIMElementList) []Violation {
	return CheckTerminalPhasesConsistencyTopologicalNode(dataset)
}

// ValidateEquipmentNotSolvedMASProfile runs hand-written checks for
// 61970-301_Equipment-AP-Con-Complex-NotSolvedMAS-SHACL.
func ValidateEquipmentNotSolvedMASProfile(dataset *cimgostructs.CIMElementList) []Violation {
	return CheckACLineSegmentBaseVoltage(dataset)
}

// ValidateSSHProfile runs hand-written checks for 61970-301_SteadyStateHypothesis-AP-Con-Complex-SHACL.
func ValidateSSHProfile(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckEnergySourceActivePowerConsumer(dataset)...)
	violations = append(violations, CheckRegulatingControlTargetDeadbandApplicability(dataset)...)
	violations = append(violations, CheckCsConverterValueRange(dataset)...)
	violations = append(violations, CheckCsConverterPPccControl(dataset)...)
	violations = append(violations, CheckVsConverterPPccControl(dataset)...)
	violations = append(violations, CheckVsConverterQPccControl(dataset)...)
	return violations
}

// ValidateSSHNotSolvedMASProfile runs hand-written checks for
// 61970-301_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS-SHACL.
func ValidateSSHNotSolvedMASProfile(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckLinearShuntCompensatorSectionsRange(dataset)...)
	violations = append(violations, CheckNonlinearShuntCompensatorSectionsValid(dataset)...)
	violations = append(violations, CheckRegulatingControlPowerFactorRequiredAttrs(dataset)...)
	violations = append(violations, CheckTapChangerStepInteger(dataset)...)
	return violations
}

// ValidateDynamicsProfile runs hand-written checks for 61970-457_Dynamics-AP-Con-Complex-SHACL.
func ValidateDynamicsProfile(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckExcitationSystemDynamicsSynchronousMachineDynamics(dataset)...)
	violations = append(violations, CheckSynchronousMachineTimeConstantReactanceModelType(dataset)...)
	return violations
}

// ValidateShortCircuitProfile runs hand-written checks for 61970-301_ShortCircuit-AP-Con-Complex-SHACL.
func ValidateShortCircuitProfile(dataset *cimgostructs.CIMElementList) []Violation {
	return CheckSeriesCompensatorVaristorUsage(dataset)
}

// ValidateShortCircuitNotSolvedMASProfile runs hand-written checks for
// 61970-301_ShortCircuit-AP-Con-Complex-NotSolvedMAS-SHACL.
func ValidateShortCircuitNotSolvedMASProfile(dataset *cimgostructs.CIMElementList) []Violation {
	return CheckMutualCouplingTerminalsAssignment(dataset)
}

// ValidateStateVariablesProfile runs hand-written checks for 61970-301_StateVariables-AP-Con-Complex-SHACL.
func ValidateStateVariablesProfile(dataset *cimgostructs.CIMElementList) []Violation {
	return CheckCsConverterStateValueRange(dataset)
}

// ValidateStateVariablesSolvedMASProfile runs hand-written checks for
// 61970-301_StateVariables-AP-Con-Complex-SolvedMAS-SHACL.
func ValidateStateVariablesSolvedMASProfile(dataset *cimgostructs.CIMElementList) []Violation {
	return CheckSvTapStepPositionRange(dataset)
}

// ValidateAllProfiles runs every generated SHACL profile orchestrator plus
// every hand-written profile-level check. The two are independent: the
// generated set comes from shaclgen.ValidateAllGeneratedProfiles; the
// hand-written ones are the Validate<Profile>Profile functions above.
func ValidateAllProfiles(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, shaclgen.ValidateAllGeneratedProfiles(dataset)...)
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
	return violations
}
