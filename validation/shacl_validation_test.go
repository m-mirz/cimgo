package validation

import (
	"bytes"
	"cimgo/cimgostructs"
	"cimgo/cimprofiles"
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
	dataset := loadDataset(t, "../testdata/test_shacl_001_GL.xml")

	byID := indexByID(ValidateGeneratedGeographicallocation6196813ComplexProfile(dataset))

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

	byID := indexByID(ValidateGeneratedDiagramlayout61970301ComplexNotsolvedmasProfile(dataset))

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

	byID := indexByID(ValidateGeneratedDiagramlayout61970301ComplexProfile(dataset))

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

func TestValidatePSTType1EQ(t *testing.T) {
	// Smoke test: conformant EQ data should produce zero generated violations.
	// Hand-written cross-cutting checks (ValidateEquipmentProfile) and SPARQL/
	// model-metadata constraints in the CSV reference are out of scope here.
	dataset := loadDataset(t, "../CGMES-Test-Configurations/v3.0/PST/PST_PhaseTapChangerLinear_Type1/PST_Type1_EQ.xml")

	violations := ValidateGeneratedEquipment61970301ComplexProfile(dataset)
	t.Logf("Focus node,Path,Constraint Component,Message,Severity")
	for _, v := range violations {
		t.Logf("%s,%s,%s,%s,%s", v.ObjectID, v.Property, v.Class, v.Message, v.Severity)
	}
	t.Logf("Total violations: %d (expected 0 for conformant EQ data)", len(violations))
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for conformant EQ data, got %d", len(violations))
	}
}
