package validation

import (
	"cimgo/cimgostructs"
	"reflect"
)

type Violation struct {
	ObjectID string
	Class    string
	Property string
	Message  string
	Severity string
}

func goTypeName(obj interface{}) string {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

// ValidateEquipmentProfile runs checks from 61970-301_Equipment-AP-Con-Complex-SHACL (eqc.*).
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

// ValidateDiagramLayoutProfile runs checks from 61970-301_DiagramLayout-AP-Con-Complex-SHACL (dlc.*).
func ValidateDiagramLayoutProfile(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckDiagramObjectIdentifiedObjectType(dataset)...)
	return violations
}

// ValidateTopologyNotSolvedMASProfile runs checks from
// 61970-301_Topology-AP-Con-Complex-NotSolvedMAS-SHACL (topcns.*).
func ValidateTopologyNotSolvedMASProfile(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckTerminalPhasesConsistencyTopologicalNode(dataset)...)
	return violations
}

// ValidateEquipmentNotSolvedMASProfile runs checks from
// 61970-301_Equipment-AP-Con-Complex-NotSolvedMAS-SHACL (eqcns.*).
func ValidateEquipmentNotSolvedMASProfile(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckACLineSegmentBaseVoltage(dataset)...)
	return violations
}

// ValidateSSHProfile runs checks from 61970-301_SteadyStateHypothesis-AP-Con-Complex-SHACL (sshc.*).
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

// ValidateSSHNotSolvedMASProfile runs checks from
// 61970-301_SteadyStateHypothesis-AP-Con-Complex-NotSolvedMAS-SHACL (sshcns.*).
func ValidateSSHNotSolvedMASProfile(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckLinearShuntCompensatorSectionsRange(dataset)...)
	violations = append(violations, CheckNonlinearShuntCompensatorSectionsValid(dataset)...)
	violations = append(violations, CheckRegulatingControlPowerFactorRequiredAttrs(dataset)...)
	violations = append(violations, CheckTapChangerStepInteger(dataset)...)
	return violations
}

// ValidateDynamicsProfile runs checks from 61970-457_Dynamics-AP-Con-Complex-SHACL (dyn457.*).
func ValidateDynamicsProfile(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckExcitationSystemDynamicsSynchronousMachineDynamics(dataset)...)
	violations = append(violations, CheckSynchronousMachineTimeConstantReactanceModelType(dataset)...)
	return violations
}

// ValidateDynamicsNotSolvedMASProfile runs checks from
// 61970-457_Dynamics-AP-Con-Complex-NotSolvedMAS-SHACL (dyn457ns.*).
func ValidateDynamicsNotSolvedMASProfile(dataset *cimgostructs.CIMElementList) []Violation {
	return nil
}

// ValidateShortCircuitProfile runs checks from 61970-301_ShortCircuit-AP-Con-Complex-SHACL (scc.*).
func ValidateShortCircuitProfile(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckSeriesCompensatorVaristorUsage(dataset)...)
	return violations
}

// ValidateShortCircuitNotSolvedMASProfile runs checks from
// 61970-301_ShortCircuit-AP-Con-Complex-NotSolvedMAS-SHACL (sccns.*).
func ValidateShortCircuitNotSolvedMASProfile(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckMutualCouplingTerminalsAssignment(dataset)...)
	return violations
}

// ValidateStateVariablesProfile runs checks from 61970-301_StateVariables-AP-Con-Complex-SHACL (svc.*).
func ValidateStateVariablesProfile(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckCsConverterStateValueRange(dataset)...)
	return violations
}

// ValidateStateVariablesSolvedMASProfile runs checks from
// 61970-301_StateVariables-AP-Con-Complex-SolvedMAS-SHACL.
func ValidateStateVariablesSolvedMASProfile(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, CheckSvTapStepPositionRange(dataset)...)
	return violations
}

func ValidateAllProfiles(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, ValidateEquipmentProfile(dataset)...)
	violations = append(violations, ValidateEquipmentNotSolvedMASProfile(dataset)...)
	violations = append(violations, ValidateSSHProfile(dataset)...)
	violations = append(violations, ValidateSSHNotSolvedMASProfile(dataset)...)
	violations = append(violations, ValidateDynamicsProfile(dataset)...)
	violations = append(violations, ValidateDynamicsNotSolvedMASProfile(dataset)...)
	violations = append(violations, ValidateShortCircuitProfile(dataset)...)
	violations = append(violations, ValidateShortCircuitNotSolvedMASProfile(dataset)...)
	violations = append(violations, ValidateStateVariablesProfile(dataset)...)
	violations = append(violations, ValidateStateVariablesSolvedMASProfile(dataset)...)
	violations = append(violations, ValidateDiagramLayoutProfile(dataset)...)
	violations = append(violations, ValidateTopologyNotSolvedMASProfile(dataset)...)
	return violations
}
