package validation

import (
	"bytes"
	"cimgo/cimgostructs"
	"cimgo/cimprofiles"
	"cimgo/shaclgen"
	"os"
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
	dataset := loadDataset(t, "../testdata/test_shacl_GL_001.xml")

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
	dataset := loadDataset(t, "../testdata/test_shacl_DL_001.xml")

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
	dataset := loadDataset(t, "../testdata/test_shacl_DL_002.xml")

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
	dataset := loadDataset(t, "../testdata/test_shacl_EQ_001.xml")
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
	dataset := loadDataset(t, "../testdata/test_shacl_EQ_001.xml")
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
	dataset := loadDataset(t, "../testdata/test_shacl_SSH_001.xml")
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
	dataset := loadDataset(t, "../testdata/test_shacl_SC_001.xml")
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
	dataset := loadDataset(t, "../testdata/test_shacl_SV_001.xml")
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
	dataset := loadDataset(t, "../testdata/test_shacl_DY_001.xml")
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
	dataset := loadDataset(t, "../testdata/test_shacl_OP_001.xml")
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
	dataset := loadDataset(t, "../testdata/test_shacl_TP_001.xml")
	byID := indexByID(shaclgen.CheckTopology61970456ComplexTopologicalNodeNameRequired(dataset))
	if got := len(byID["TopologicalNode.OK"]); got != 0 {
		t.Errorf("TopologicalNode.OK (name present): expected 0 violations, got %d: %v", got, byID["TopologicalNode.OK"])
	}
	if got := len(byID["TopologicalNode.BAD"]); got != 1 {
		t.Errorf("TopologicalNode.BAD (name absent): expected 1 violation, got %d: %v", got, byID["TopologicalNode.BAD"])
	}
	logViolations(t, byID)
}

func TestEQBDBoundaryPointFromEndIsoCode(t *testing.T) {
	// BoundaryPoint.fromEndIsoCode must be a valid European ISO-3166-1-alpha-2 code (sh:in).
	dataset := loadDataset(t, "../testdata/test_shacl_EQBD_001.xml")
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
