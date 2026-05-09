package validation

import (
	"bytes"
	"cimgo/cimgostructs"
	"cimgo/cimprofiles"
	"cimgo/shaclgen"
	"os"
	"strings"
	"testing"
)

// indexByID groups generated SHACL violations by their focus-node MRID so the
// per-object assertions below stay readable.
func indexByID(violations []Violation) map[string][]Violation {
	out := make(map[string][]Violation)
	for _, v := range violations {
		out[v.ObjectID] = append(out[v.ObjectID], v)
	}
	return out
}

func loadDataset(t *testing.T, path string) *cimgostructs.CIMElementList {
	t.Helper()
	dataset := cimgostructs.NewCIMElementList()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", path, err)
	}
	cimprofiles.DecodeProfile(bytes.NewReader(b), dataset)
	t.Logf("Loaded %d elements", len(dataset.Elements))
	return dataset
}

func logViolations(t *testing.T, byID map[string][]Violation) {
	for id, vs := range byID {
		for _, v := range vs {
			t.Logf("Object %s: [%s] %s: %s", id, v.Class, v.Property, v.Message)
		}
	}
}

func TestValidateCoordinateSystemCrsUrn(t *testing.T) {
	dataset := loadDataset(t, "../testdata/test_shacl_001_GL.xml")

	byID := indexByID(shaclgen.ValidateGeneratedGeographicallocation6196813ComplexProfile(dataset))

	if got := len(byID["CoordinateSystem.WGS84"]); got != 0 {
		t.Errorf("CoordinateSystem.WGS84 (default crsUrn): expected 0 violations, got %d: %v",
			got, byID["CoordinateSystem.WGS84"])
	}
	if got := len(byID["CoordinateSystem.ETRS89"]); got != 1 {
		t.Errorf("CoordinateSystem.ETRS89 (non-default crsUrn): expected 1 violation, got %d: %v",
			got, byID["CoordinateSystem.ETRS89"])
	}
	logViolations(t, byID)
}

func TestValidateDiagramObjectIdentifiedObject(t *testing.T) {
	// The rule says DiagramObject.IdentifiedObject must be an IRI and must NOT
	// point to a cim.GeneratingUnit (it should reference SynchronousMachine).
	dataset := loadDataset(t, "../testdata/test_shacl_002_DL.xml")

	byID := indexByID(shaclgen.ValidateGeneratedDiagramlayout61970301ComplexNotsolvedmasProfile(dataset))

	if got := len(byID["DiagramObject.OK"]); got != 0 {
		t.Errorf("DiagramObject.OK (points to SynchronousMachine): expected 0 violations, got %d: %v",
			got, byID["DiagramObject.OK"])
	}
	for _, badID := range []string{"DiagramObject.BAD", "TextDiagramObject.BAD"} {
		if got := len(byID[badID]); got != 1 {
			t.Errorf("%s (points to GeneratingUnit): expected 1 violation, got %d: %v",
				badID, got, byID[badID])
		}
	}
	logViolations(t, byID)
}

func TestValidateDiagramObjectPointSequenceNumber(t *testing.T) {
	// The rule says DiagramObjectPoint.sequenceNumber must be > 0 (sh:minExclusive 0.0).
	dataset := loadDataset(t, "../testdata/test_shacl_003_DL.xml")

	byID := indexByID(shaclgen.ValidateGeneratedDiagramlayout61970301ComplexProfile(dataset))

	if got := len(byID["DiagramObjectPoint.OK"]); got != 0 {
		t.Errorf("DiagramObjectPoint.OK (sequenceNumber=1): expected 0 violations, got %d: %v",
			got, byID["DiagramObjectPoint.OK"])
	}
	if got := len(byID["DiagramObjectPoint.NEG"]); got != 1 {
		t.Errorf("DiagramObjectPoint.NEG (sequenceNumber=-1): expected 1 violation, got %d: %v",
			got, byID["DiagramObjectPoint.NEG"])
	}
	logViolations(t, byID)
}

func TestEquipmentACLineSegmentLength(t *testing.T) {
	// ACLineSegment.length must be >= 0 (sh:minInclusive 0).
	dataset := loadDataset(t, "../testdata/test_shacl_004_EQ.xml")
	byID := indexByID(shaclgen.CheckEquipment61970301ComplexACLineSegmentLengthMinInclusive(dataset))
	if got := len(byID["ACLineSegment.OK"]); got != 0 {
		t.Errorf("ACLineSegment.OK (length=5): expected 0 violations, got %d: %v", got, byID["ACLineSegment.OK"])
	}
	if got := len(byID["ACLineSegment.BAD"]); got != 1 {
		t.Errorf("ACLineSegment.BAD (length=-1): expected 1 violation, got %d: %v", got, byID["ACLineSegment.BAD"])
	}
	logViolations(t, byID)
}

func TestEquipmentBaseVoltageNominalVoltage(t *testing.T) {
	// BaseVoltage.nominalVoltage must be > 0 (sh:minExclusive 0).
	dataset := loadDataset(t, "../testdata/test_shacl_004_EQ.xml")
	byID := indexByID(shaclgen.CheckEquipment61970301ComplexBaseVoltageNominalVoltageMinExclusive(dataset))
	if got := len(byID["BaseVoltage.OK"]); got != 0 {
		t.Errorf("BaseVoltage.OK (nominalVoltage=110): expected 0 violations, got %d: %v", got, byID["BaseVoltage.OK"])
	}
	if got := len(byID["BaseVoltage.BAD"]); got != 1 {
		t.Errorf("BaseVoltage.BAD (nominalVoltage=-1): expected 1 violation, got %d: %v", got, byID["BaseVoltage.BAD"])
	}
	logViolations(t, byID)
}

func TestSSHBatteryUnitStoredELessThanRatedE(t *testing.T) {
	// BatteryUnit.storedE must be < ratedE (sh:lessThan).
	dataset := loadDataset(t, "../testdata/test_shacl_005_SSH.xml")
	byID := indexByID(shaclgen.CheckSteadystatehypothesis61970301ComplexNotsolvedmasBatteryUnitStoredELessThan(dataset))
	if got := len(byID["BatteryUnit.OK"]); got != 0 {
		t.Errorf("BatteryUnit.OK (storedE=50 < ratedE=100): expected 0 violations, got %d: %v", got, byID["BatteryUnit.OK"])
	}
	if got := len(byID["BatteryUnit.BAD"]); got != 1 {
		t.Errorf("BatteryUnit.BAD (storedE=150 >= ratedE=100): expected 1 violation, got %d: %v", got, byID["BatteryUnit.BAD"])
	}
	logViolations(t, byID)
}

func TestSCPowerTransformerEndPhaseAngleClock(t *testing.T) {
	// PowerTransformerEnd.phaseAngleClock must be in [0, 11] (sh:maxInclusive 11).
	dataset := loadDataset(t, "../testdata/test_shacl_006_SC.xml")
	byID := indexByID(shaclgen.CheckShortcircuit61970301ComplexPowerTransformerEndPhaseAngleClockMaxInclusive(dataset))
	if got := len(byID["PowerTransformerEnd.OK"]); got != 0 {
		t.Errorf("PowerTransformerEnd.OK (phaseAngleClock=5): expected 0 violations, got %d: %v", got, byID["PowerTransformerEnd.OK"])
	}
	if got := len(byID["PowerTransformerEnd.BAD"]); got != 1 {
		t.Errorf("PowerTransformerEnd.BAD (phaseAngleClock=12): expected 1 violation, got %d: %v", got, byID["PowerTransformerEnd.BAD"])
	}
	logViolations(t, byID)
}

func TestSVSvVoltage(t *testing.T) {
	// SvVoltage.v must be > 0 (sh:minExclusive 0).
	dataset := loadDataset(t, "../testdata/test_shacl_007_SV.xml")
	byID := indexByID(shaclgen.CheckStatevariables61970301ComplexSvVoltageVMinExclusive(dataset))
	if got := len(byID["SvVoltage.OK"]); got != 0 {
		t.Errorf("SvVoltage.OK (v=110): expected 0 violations, got %d: %v", got, byID["SvVoltage.OK"])
	}
	if got := len(byID["SvVoltage.BAD"]); got != 1 {
		t.Errorf("SvVoltage.BAD (v=-1): expected 1 violation, got %d: %v", got, byID["SvVoltage.BAD"])
	}
	logViolations(t, byID)
}

func TestDYAsynchronousMachineTppoLessThanTpo(t *testing.T) {
	// AsynchronousMachineTimeConstantReactance.tppo must be < tpo (sh:lessThan).
	dataset := loadDataset(t, "../testdata/test_shacl_008_DY.xml")
	byID := indexByID(shaclgen.CheckDynamics61970302ComplexAsynchronousMachineTimeConstantReactanceTppoLessThan(dataset))
	if got := len(byID["AsynchronousMachineTimeConstantReactance.OK"]); got != 0 {
		t.Errorf("AMTCR.OK (tppo=0.01 < tpo=0.1): expected 0 violations, got %d: %v", got, byID["AsynchronousMachineTimeConstantReactance.OK"])
	}
	if got := len(byID["AsynchronousMachineTimeConstantReactance.BAD"]); got != 1 {
		t.Errorf("AMTCR.BAD (tppo=0.1 >= tpo=0.05): expected 1 violation, got %d: %v", got, byID["AsynchronousMachineTimeConstantReactance.BAD"])
	}
	logViolations(t, byID)
}

func TestOPAccumulatorLimitValue(t *testing.T) {
	// AccumulatorLimit.value must be > 0 (sh:minExclusive 0).
	dataset := loadDataset(t, "../testdata/test_shacl_009_OP.xml")
	byID := indexByID(shaclgen.CheckOperation61970301ComplexAccumulatorLimitValueMinExclusive(dataset))
	if got := len(byID["AccumulatorLimit.OK"]); got != 0 {
		t.Errorf("AccumulatorLimit.OK (value=5): expected 0 violations, got %d: %v", got, byID["AccumulatorLimit.OK"])
	}
	if got := len(byID["AccumulatorLimit.BAD"]); got != 1 {
		t.Errorf("AccumulatorLimit.BAD (value=-1): expected 1 violation, got %d: %v", got, byID["AccumulatorLimit.BAD"])
	}
	logViolations(t, byID)
}

func TestTPTopologicalNodeNameRequired(t *testing.T) {
	// TopologicalNode.name is required (sh:required).
	dataset := loadDataset(t, "../testdata/test_shacl_010_TP.xml")
	byID := indexByID(shaclgen.CheckTopology61970456ComplexTopologicalNodeNameRequired(dataset))
	if got := len(byID["TopologicalNode.OK"]); got != 0 {
		t.Errorf("TopologicalNode.OK (name present): expected 0 violations, got %d: %v", got, byID["TopologicalNode.OK"])
	}
	if got := len(byID["TopologicalNode.BAD"]); got != 1 {
		t.Errorf("TopologicalNode.BAD (name absent): expected 1 violation, got %d: %v", got, byID["TopologicalNode.BAD"])
	}
	logViolations(t, byID)
}

func TestValidateDiagramObjectIdentifiedObjectSPARQL(t *testing.T) {
	// The rule says DiagramObject.IdentifiedObject must NOT point to:
	// Diagram, DiagramObject, VisibilityLayer, DiagramStyle, DiagramObjectStyle, TextDiagramObject.
	dataset := loadDataset(t, "../testdata/test_shacl_012_DL_SPARQL.xml")

	byID := indexByID(ValidateDiagramLayoutProfile(dataset))

	if got := len(byID["DiagramObject.OK"]); got != 0 {
		t.Errorf("DiagramObject.OK: expected 0 violations, got %d: %v",
			got, byID["DiagramObject.OK"])
	}
	for _, badID := range []string{"DiagramObject.BAD", "TextDiagramObject.BAD"} {
		if got := len(byID[badID]); got != 1 {
			t.Errorf("%s: expected 1 violation, got %d: %v",
				badID, got, byID[badID])
		}
	}
	logViolations(t, byID)
}

func TestValidateBoundaryPointTieFlowSPARQL(t *testing.T) {
	// If isExcludedFromAreaInterchange is false (default), a TieFlow is required.
	// If true, no TieFlow should be modeled.
	dataset := loadDataset(t, "../testdata/test_shacl_013_EQBD_SPARQL.xml")

	byID := indexByID(ValidateEquipmentBoundaryProfile(dataset))

	if got := len(byID["BP.OK1"]); got != 0 {
		t.Errorf("BP.OK1: expected 0 violations, got %d: %v", got, byID["BP.OK1"])
	}
	if got := len(byID["BP.OK2"]); got != 0 {
		t.Errorf("BP.OK2: expected 0 violations, got %d: %v", got, byID["BP.OK2"])
	}
	if got := len(byID["BP.BAD1"]); got != 1 {
		t.Errorf("BP.BAD1: expected 1 violation, got %d: %v", got, byID["BP.BAD1"])
	}
	if got := len(byID["BP.BAD2"]); got != 1 {
		t.Errorf("BP.BAD2: expected 1 violation, got %d: %v", got, byID["BP.BAD2"])
	}
	logViolations(t, byID)
}

func TestValidateMutualCouplingSPARQL(t *testing.T) {
	// MutualCoupling.First_Terminal and Second_Terminal must point to different ACLineSegments.
	dataset := loadDataset(t, "../testdata/test_shacl_014_SC_SPARQL.xml")

	byID := indexByID(ValidateShortCircuitNotSolvedMASProfile(dataset))

	if got := len(byID["MC.OK"]); got != 0 {
		t.Errorf("MC.OK: expected 0 violations, got %d: %v", got, byID["MC.OK"])
	}
	if got := len(byID["MC.BAD.SAME"]); got != 1 {
		t.Errorf("MC.BAD.SAME: expected 1 violation, got %d: %v", got, byID["MC.BAD.SAME"])
	}
	if got := len(byID["MC.BAD.TYPE"]); got != 1 {
		t.Errorf("MC.BAD.TYPE: expected 1 violation, got %d: %v", got, byID["MC.BAD.TYPE"])
	}
	logViolations(t, byID)
}

func TestValidateSeriesCompensatorVaristorSPARQL(t *testing.T) {
	// varistorRatedCurrent/VoltageThreshold only exchanged if varistorPresent is true.
	dataset := loadDataset(t, "../testdata/test_shacl_015_SC_VARISTOR_SPARQL.xml")

	byID := indexByID(ValidateShortCircuitProfile(dataset))

	if got := len(byID["SC.OK.1"]); got != 0 {
		t.Errorf("SC.OK.1: expected 0 violations, got %d: %v", got, byID["SC.OK.1"])
	}
	if got := len(byID["SC.OK.2"]); got != 0 {
		t.Errorf("SC.OK.2: expected 0 violations, got %d: %v", got, byID["SC.OK.2"])
	}
	if got := len(byID["SC.BAD.1"]); got != 1 {
		t.Errorf("SC.BAD.1: expected 1 violation, got %d: %v", got, byID["SC.BAD.1"])
	}
	if got := len(byID["SC.BAD.2"]); got != 1 {
		t.Errorf("SC.BAD.2: expected 1 violation, got %d: %v", got, byID["SC.BAD.2"])
	}
	logViolations(t, byID)
}

func TestValidateCsConverterStateValueRangeSPARQL(t *testing.T) {
	// alpha [10, 18] for rectifier, gamma [17, 20] for inverter.
	dataset := loadDataset(t, "../testdata/test_shacl_016_SV_SPARQL.xml")

	byID := indexByID(ValidateStateVariablesProfile(dataset))

	if got := len(byID["CSC.RECT.OK"]); got != 0 {
		t.Errorf("CSC.RECT.OK: expected 0 violations, got %d: %v", got, byID["CSC.RECT.OK"])
	}
	if got := len(byID["CSC.INV.OK"]); got != 0 {
		t.Errorf("CSC.INV.OK: expected 0 violations, got %d: %v", got, byID["CSC.INV.OK"])
	}
	if got := len(byID["CSC.RECT.BAD"]); got != 1 {
		t.Errorf("CSC.RECT.BAD: expected 1 violation, got %d: %v", got, byID["CSC.RECT.BAD"])
	}
	if got := len(byID["CSC.INV.BAD"]); got != 1 {
		t.Errorf("CSC.INV.BAD: expected 1 violation, got %d: %v", got, byID["CSC.INV.BAD"])
	}
	logViolations(t, byID)
}

func TestValidateSvTapStepPositionRangeSPARQL(t *testing.T) {
	// position must be within [lowStep, highStep] of the associated TapChanger.
	dataset := loadDataset(t, "../testdata/test_shacl_017_SV_SOLVED_SPARQL.xml")

	byID := indexByID(ValidateStateVariablesSolvedMASProfile(dataset))

	if got := len(byID["SV.OK.1"]); got != 0 {
		t.Errorf("SV.OK.1: expected 0 violations, got %d: %v", got, byID["SV.OK.1"])
	}
	if got := len(byID["SV.BAD.LOW"]); got != 1 {
		t.Errorf("SV.BAD.LOW: expected 1 violation, got %d: %v", got, byID["SV.BAD.LOW"])
	}
	if got := len(byID["SV.BAD.HIGH"]); got != 1 {
		t.Errorf("SV.BAD.HIGH: expected 1 violation, got %d: %v", got, byID["SV.BAD.HIGH"])
	}
	logViolations(t, byID)
}

func TestValidateSSHSPARQL(t *testing.T) {
	// Various complex SSH NotSolvedMAS rules.
	dataset := loadDataset(t, "../testdata/test_shacl_018_SSH_SPARQL.xml")

	byID := indexByID(ValidateSSHNotSolvedMASProfile(dataset))

	badIDs := []string{
		"CA.INTERCHANGE.BAD",
		"CSC.INV.BAD.ALPHA",
		"CSC.RECT.BAD.GAMMA",
		"LSC.BAD.SECTIONS",
		"LSC.NONINT.SECTIONS",
		"NSC.BAD.SECTIONS",
		"RC.PF.BAD",
		"RTC.BAD.STEP",
	}
	for _, id := range badIDs {
		if got := len(byID[id]); got == 0 {
			t.Errorf("%s: expected violation, got none", id)
		}
	}
	logViolations(t, byID)
}

func TestValidateSSHComplexSPARQL(t *testing.T) {
	// Various complex SSH SPARQL rules.
	dataset := loadDataset(t, "../testdata/test_shacl_019_SSH_COMPLEX_SPARQL.xml")

	byID := indexByID(ValidateSSHProfile(dataset))

	badIDs := []string{
		"ES.CONSUMER",
		"RC.CONT.WITH.DEAD",
		"RC.DISC.WITHOUT.DEAD",
		"CSC.RECT.BAD.RANGE",
		"VSC.P.BAD.DROOP",
	}
	for _, id := range badIDs {
		if got := len(byID[id]); got == 0 {
			t.Errorf("%s: expected violation, got none", id)
		}
	}
	logViolations(t, byID)
}

func TestValidateTopologyNotSolvedMASSPARQL(t *testing.T) {
	// Terminals at the same TopologicalNode must have consistent phase codes.
	dataset := loadDataset(t, "../testdata/test_shacl_020_TP_SPARQL.xml")

	byID := indexByID(ValidateTopologyNotSolvedMASProfile(dataset))

	if got := len(byID["TN.OK"]); got != 0 {
		t.Errorf("TN.OK: expected 0 violations, got %d: %v", got, byID["TN.OK"])
	}
	if got := len(byID["TN.BAD"]); got != 1 {
		t.Errorf("TN.BAD: expected 1 violation, got %d: %v", got, byID["TN.BAD"])
	}
	logViolations(t, byID)
}

func TestValidateDynamicsMbaseEquationSPARQL(t *testing.T) {
	// mwbase must equal RotatingMachine.ratedPowerFactor * RotatingMachine.ratedS.
	dataset := loadDataset(t, "../testdata/test_shacl_021_DY_SPARQL.xml")

	byID := indexByID(ValidateDynamicsProfile(dataset))

	if got := len(byID["GOV.OK"]); got != 0 {
		t.Errorf("GOV.OK: expected 0 violations, got %d: %v", got, byID["GOV.OK"])
	}
	if got := len(byID["GOV.BAD"]); got != 1 {
		t.Errorf("GOV.BAD: expected 1 violation, got %d: %v", got, byID["GOV.BAD"])
	}
	logViolations(t, byID)
}

func TestValidateExcitationSystemDynamicsSPARQL(t *testing.T) {
	// ExcitationSystemDynamics.SynchronousMachineDynamics shall not point to a SynchronousMachineSimplified.
	dataset := loadDataset(t, "../testdata/test_shacl_022_DY_EXC_SPARQL.xml")

	byID := indexByID(ValidateDynamicsProfile(dataset))

	if got := len(byID["EXC.OK"]); got != 0 {
		t.Errorf("EXC.OK: expected 0 violations, got %d: %v", got, byID["EXC.OK"])
	}
	if got := len(byID["EXC.BAD"]); got != 1 {
		t.Errorf("EXC.BAD: expected 1 violation, got %d: %v", got, byID["EXC.BAD"])
	}
	logViolations(t, byID)
}

func TestValidateTopologyEXCH8TopologicalNodeSPARQL(t *testing.T) {
	// Terminal.TopologicalNode is required if a RegulatingControl is associated.
	dataset := loadDataset(t, "../testdata/test_shacl_023_TP_600_SPARQL.xml")

	byID := indexByID(ValidateTopologyNotSolvedMASProfile(dataset))

	if got := len(byID["Term.OK"]); got != 0 {
		t.Errorf("Term.OK: expected 0 violations, got %d: %v", got, byID["Term.OK"])
	}
	if got := len(byID["Term.BAD"]); got != 1 {
		t.Errorf("Term.BAD: expected 1 violation, got %d: %v", got, byID["Term.BAD"])
	}
	logViolations(t, byID)
}

func TestValidateCommonSPARQL(t *testing.T) {
	// Various common CGMES SPARQL rules.
	dataset := loadDataset(t, "../testdata/test_shacl_024_COMMON_SPARQL.xml")

	byID := indexByID(ValidateCommonRules(dataset))

	badIDs := []string{
		"urn:uuid:header-1",
		"Substation-Not-A-UUID",
		// Duplicate mRID: reported on both or either
		"_7336d36e-d917-4e54-9469-8730b200b3d5", // NaN
		"_6336d36e-d917-4e54-9469-8730b200b3d5", // Long Name
		"_5336d36e-d917-4e54-9469-8730b200b3d5", // Short Name
		"_4336d36e-d917-4e54-9469-8730b200b3d5", // EIC
	}
	for _, id := range badIDs {
		if got := len(byID[id]); got == 0 {
			t.Errorf("%s: expected violation, got none", id)
		}
	}
	// Check duplicate mRID specifically
	if len(byID["_8336d36e-d917-4e54-9469-8730b200b3d5"]) == 0 && len(byID["_9336d36e-d917-4e54-9469-8730b200b3d5"]) == 0 {
		t.Errorf("Duplicate mRID: expected violation on either _833... or _933..., got none")
	}
	logViolations(t, byID)
}

func TestValidateRegulatingControlTargetValueTapChangerSPARQL(t *testing.T) {
	// RegulatingControl.targetValue must be within TapChanger capability limits.
	dataset := loadDataset(t, "../testdata/test_shacl_025_EQ_NOTSOLVED_SPARQL.xml")

	byID := indexByID(ValidateEquipmentNotSolvedMASProfile(dataset))

	if got := len(byID["TCC.OK"]); got != 0 {
		t.Errorf("TCC.OK: expected 0 violations, got %d: %v", got, byID["TCC.OK"])
	}
	if got := len(byID["TCC.BAD"]); got != 1 {
		t.Errorf("TCC.BAD: expected 1 violation, got %d: %v", got, byID["TCC.BAD"])
	}
	logViolations(t, byID)
}

func TestValidateEQ452SPARQL(t *testing.T) {
	// Various complex EQ 452 SPARQL rules.
	dataset := loadDataset(t, "../testdata/test_shacl_026_EQ_452_SPARQL.xml")

	byID := indexByID(ValidateEquipmentProfile(dataset))

	if got := len(byID["SW.OK.SAME_VL"]); got != 0 {
		t.Errorf("SW.OK.SAME_VL: expected 0 violations, got %d: %v", got, byID["SW.OK.SAME_VL"])
	}
	if got := len(byID["SW.BAD.DIFF_BV"]); got != 1 {
		t.Errorf("SW.BAD.DIFF_BV: expected 1 violation, got %d: %v", got, byID["SW.BAD.DIFF_BV"])
	}
	if got := len(byID["Line.BAD.SAME_CN"]); got != 1 {
		t.Errorf("Line.BAD.SAME_CN: expected 1 violation, got %d: %v", got, byID["Line.BAD.SAME_CN"])
	}
	logViolations(t, byID)
}

func TestValidateOperationNotSolvedMASSPARQL(t *testing.T) {
	// Measurement.Terminal must reference a Terminal of the Equipment referenced by
	// Measurement.PowerSystemResource, unless measurementType is TapPosition or SwitchPosition.
	dataset := loadDataset(t, "../testdata/test_shacl_027_OP_NOTSOLVED_SPARQL.xml")

	byID := indexByID(ValidateOperationProfile(dataset))

	if got := len(byID["MEAS.OK"]); got != 0 {
		t.Errorf("MEAS.OK: expected 0 violations, got %d: %v", got, byID["MEAS.OK"])
	}
	if got := len(byID["MEAS.TAP.OK"]); got != 0 {
		t.Errorf("MEAS.TAP.OK: expected 0 violations, got %d: %v", got, byID["MEAS.TAP.OK"])
	}
	if got := len(byID["MEAS.BAD.TERMINAL"]); got != 1 {
		t.Errorf("MEAS.BAD.TERMINAL: expected 1 violation, got %d: %v", got, byID["MEAS.BAD.TERMINAL"])
	}
	if got := len(byID["MEAS.TAP.BAD"]); got != 1 {
		t.Errorf("MEAS.TAP.BAD: expected 1 violation, got %d: %v", got, byID["MEAS.TAP.BAD"])
	}
	if got := len(byID["MEAS.VOLT.BAD.ABSENT"]); got != 1 {
		t.Errorf("MEAS.VOLT.BAD.ABSENT: expected 1 violation, got %d: %v", got, byID["MEAS.VOLT.BAD.ABSENT"])
	}
	logViolations(t, byID)
}

func TestValidateSC452SPARQL(t *testing.T) {
	// Various complex SC 452 SPARQL rules.
	dataset := loadDataset(t, "../testdata/test_shacl_028_SC_452_SPARQL.xml")

	byID := indexByID(ValidateShortCircuitProfile(dataset))

	if got := len(byID["SM.OK"]); got != 0 {
		t.Errorf("SM.OK: expected 0 violations, got %d: %v", got, byID["SM.OK"])
	}
	if got := len(byID["SM.BAD"]); got != 1 {
		t.Errorf("SM.BAD: expected 1 violation, got %d: %v", got, byID["SM.BAD"])
	}
	if got := len(byID["PTE.OK"]); got != 0 {
		t.Errorf("PTE.OK: expected 0 violations, got %d: %v", got, byID["PTE.OK"])
	}
	if got := len(byID["PTE.BAD"]); got != 1 {
		t.Errorf("PTE.BAD: expected 1 violation, got %d: %v", got, byID["PTE.BAD"])
	}
	logViolations(t, byID)
}

func TestValidateAngleReferenceSPARQL(t *testing.T) {
	// Priority 1 SM must be at the AngleRefTopologicalNode.
	dataset := loadDataset(t, "../testdata/test_shacl_029_SOLVED_SPARQL.xml")

	byID := indexByID(ValidateStateVariablesSolvedMASProfile(dataset))

	if got := len(byID["SM.OK"]); got != 0 {
		t.Errorf("SM.OK: expected 0 violations, got %d: %v", got, byID["SM.OK"])
	}
	if got := len(byID["SM.BAD.NODE"]); got == 0 {
		t.Errorf("SM.BAD.NODE: expected violation, got none")
	}
	// Check for global violation due to duplicate priority 1 machines
	foundGlobal := false
	for _, v := range byID["global"] {
		if strings.Contains(v.Message, "Multiple machines") {
			foundGlobal = true; break
		}
	}
	if !foundGlobal {
		t.Errorf("global: expected violation for duplicate priority 1 machines, got %v", byID["global"])
	}
	// We expect TN.OTHER to have a violation for missing SvVoltage
	if got := len(byID["TN.OTHER"]); got == 0 {
		t.Errorf("TN.OTHER: expected violation for missing SvVoltage, got none")
	}
	logViolations(t, byID)
}

func TestValidateSvStateVariablesSolvedMASSPARQL(t *testing.T) {
	// Various complex SV SolvedMAS 456 SPARQL rules.
	dataset := loadDataset(t, "../testdata/test_shacl_030_SV_SOLVED_456_SPARQL.xml")

	byID := indexByID(ValidateStateVariablesSolvedMASProfile(dataset))

	badIDs := []string{
		"SM.ENERGIZED",  // Missing SvPowerFlow
		"SW.1",          // Missing SvSwitch
		"SVSC.BAD",      // Non-integer sections
		"SVTS.BAD",      // Non-integer position
		"SVV.BAD",       // < 0.4 pu
	}
	for _, id := range badIDs {
		if got := len(byID[id]); got == 0 {
			t.Errorf("%s: expected violation, got none", id)
		}
	}
	logViolations(t, byID)
}

func TestValidateTopologySameTopologicalNodeSPARQL(t *testing.T) {
	// Terminals of a retained Switch shall not be connected to the same TopologicalNode.
	dataset := loadDataset(t, "../testdata/test_shacl_031_TP_456_SPARQL.xml")

	byID := indexByID(ValidateTopologyNotSolvedMASProfile(dataset))

	if got := len(byID["SW.OK"]); got != 0 {
		t.Errorf("SW.OK: expected 0 violations, got %d: %v", got, byID["SW.OK"])
	}
	if got := len(byID["SW.BAD"]); got != 1 {
		t.Errorf("SW.BAD: expected 1 violation, got %d: %v", got, byID["SW.BAD"])
	}
	if got := len(byID["SW.NOT_RETAINED.OK"]); got != 0 {
		t.Errorf("SW.NOT_RETAINED.OK: expected 0 violations, got %d: %v", got, byID["SW.NOT_RETAINED.OK"])
	}
	logViolations(t, byID)
}

func TestValidateDynamics302SPARQL(t *testing.T) {
	// Various complex Dynamics 302 SPARQL rules.
	dataset := loadDataset(t, "../testdata/test_shacl_032_DY_302_SPARQL.xml")

	byID := indexByID(ValidateDynamicsProfile(dataset))

	badIDs := []string{
		"EXC.AC8B.BAD",
		"EXC.BBC.BAD",
		"EXC.DC4B.BAD",
		"PSS.2ST.BAD",
		"GOV.H4.SIMPLE.BAD",
		"GOV.H4.KAPLAN.BAD",
		"LOAD.STATIC.Z.BAD",
		"SM.SAT.BAD",
		"SMS.BAD",
		"MECH.BAD",
	}
	for _, id := range badIDs {
		if got := len(byID[id]); got == 0 {
			t.Errorf("%s: expected violation, got none", id)
		}
	}
	logViolations(t, byID)
}

func TestValidateSv600SolvedMASSPARQL(t *testing.T) {
	// Various complex SV SolvedMAS 600-1 SPARQL rules.
	dataset := loadDataset(t, "../testdata/test_shacl_033_SV_600_SOLVED_SPARQL.xml")

	byID := indexByID(ValidateAllProfiles(dataset))

	badIDs := []string{
		"S1",                      // Dangling reference
		"LSC.SYNC.BAD",           // Sync mismatch
		"RTC.SYNC.BAD",           // Sync mismatch
		"SM.ENERGIZED.NO_STATUS", // Missing SvStatus
		"LSC.ENERGIZED.NO_SVSC",  // Missing SvShuntCompensatorSections
	}
	for _, id := range badIDs {
		if got := len(byID[id]); got == 0 {
			t.Errorf("%s: expected violation, got none", id)
		}
	}
	logViolations(t, byID)
}

func TestValidateRegulatingControl6002SPARQL(t *testing.T) {
	// Various complex RC SolvedMAS 600-2 SPARQL rules.
	dataset := loadDataset(t, "../testdata/test_shacl_034_SOLVED_600_2_SPARQL.xml")

	byID := indexByID(ValidateStateVariablesSolvedMASProfile(dataset))

	if got := len(byID["RC.V.2"]); got != 1 {
		t.Errorf("RC.V.2: expected 1 violation for contradictory target, got %d: %v", got, byID["RC.V.2"])
	}
	if got := len(byID["RC.V.1"]); got == 0 {
		t.Errorf("RC.V.1: expected violation for machine/tap island mismatch, got none")
	}
	logViolations(t, byID)
}

func TestValidateEQ6002SPARQL(t *testing.T) {
	// Various complex EQ 600-2 SPARQL rules.
	dataset := loadDataset(t, "../testdata/test_shacl_035_EQ_600_2_SPARQL.xml")

	byID := indexByID(ValidateEquipmentProfile(dataset))

	if got := len(byID["global"]); got == 0 {
		t.Errorf("global: expected violation for substation count, got none")
	}
	if got := len(byID["RCC1"]); got != 1 {
		t.Errorf("RCC1: expected 1 violation for units, got %d: %v", got, byID["RCC1"])
	}
	if got := len(byID["RTC1"]); got != 1 {
		t.Errorf("RTC1: expected 1 violation for neutralU sync, got %d: %v", got, byID["RTC1"])
	}
	logViolations(t, byID)
}

func TestValidateSC6002SPARQL(t *testing.T) {
	// varistorRatedCurrent and varistorVoltageThreshold are required if SeriesCompensator.varistorPresent is true.
	dataset := loadDataset(t, "../testdata/test_shacl_036_SC_600_2_SPARQL.xml")

	byID := indexByID(ValidateShortCircuitProfile(dataset))

	if got := len(byID["SC.OK.1"]); got != 0 {
		t.Errorf("SC.OK.1: expected 0 violations, got %d: %v", got, byID["SC.OK.1"])
	}
	if got := len(byID["SC.OK.2"]); got != 0 {
		t.Errorf("SC.OK.2: expected 0 violations, got %d: %v", got, byID["SC.OK.2"])
	}
	if got := len(byID["SC.BAD.REQUIRED"]); got != 2 {
		t.Errorf("SC.BAD.REQUIRED: expected 2 violations (both missing), got %d: %v", got, byID["SC.BAD.REQUIRED"])
	}
	logViolations(t, byID)
}

func TestEQBDBoundaryPointFromEndIsoCode(t *testing.T) {
	// BoundaryPoint.fromEndIsoCode must be a valid European ISO-3166-1-alpha-2 code (sh:in).
	dataset := loadDataset(t, "../testdata/test_shacl_011_EQBD.xml")
	byID := indexByID(shaclgen.CheckEquipmentboundary61970301ComplexBoundaryPointFromEndIsoCodeIn(dataset))
	if got := len(byID["BoundaryPoint.OK"]); got != 0 {
		t.Errorf("BoundaryPoint.OK (fromEndIsoCode=DE): expected 0 violations, got %d: %v", got, byID["BoundaryPoint.OK"])
	}
	if got := len(byID["BoundaryPoint.BAD"]); got != 1 {
		t.Errorf("BoundaryPoint.BAD (fromEndIsoCode=XX): expected 1 violation, got %d: %v", got, byID["BoundaryPoint.BAD"])
	}
	logViolations(t, byID)
}

func TestValidatePSTType1EQ(t *testing.T) {
	// Smoke test: conformant EQ data should produce zero generated violations.
	// Hand-written cross-cutting checks (ValidateEquipmentProfile) and SPARQL/
	// model-metadata constraints in the CSV reference are out of scope here.
	dataset := loadDataset(t, "../CGMES-Test-Configurations/v3.0/PST/PST_PhaseTapChangerLinear_Type1/PST_Type1_EQ.xml")

	violations := shaclgen.ValidateGeneratedEquipment61970301ComplexProfile(dataset)
	t.Logf("Focus node,Path,Constraint Component,Message,Severity")
	for _, v := range violations {
		t.Logf("%s,%s,%s,%s,%s", v.ObjectID, v.Property, v.Class, v.Message, v.Severity)
	}
	t.Logf("Total violations: %d (expected 0 for conformant EQ data)", len(violations))
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for conformant EQ data, got %d", len(violations))
	}
}
