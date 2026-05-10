package validation

import (
	"testing"
)

func TestValidatePSTType1EQ(t *testing.T) {
	// Smoke test: conformant EQ data should produce zero generated violations.
	// Hand-written cross-cutting checks (ValidateEquipmentProfile) and SPARQL/
	// model-metadata constraints in the CSV reference are out of scope here.
	dataset := loadDataset(t, "../CGMES-Test-Configurations/v3.0/PST/PST_PhaseTapChangerLinear_Type1/PST_Type1_EQ.xml")

	violations := ValidateEQProfile(dataset)
	t.Logf("Focus node,Path,Constraint Component,Message,Severity")
	for _, v := range violations {
		t.Logf("%s,%s,%s,%s,%s", v.ObjectID, v.Property, v.Class, v.Message, v.Severity)
	}
	t.Logf("Total violations: %d (expected 0 for conformant EQ data)", len(violations))
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for conformant EQ data, got %d", len(violations))
	}
}
