package validation

import (
	"strings"
	"testing"
)

func TestValidatePSTType1EQ(t *testing.T) {
	// Conformant EQ data should produce zero sh.Violation findings.
	// The ValiMate reference CSV reports 0 errors and 1 sh:Warning (Substation
	// count design note) and 1 sh:Info (PROF10-EQ: missing EQBD reference).
	dataset := loadDataset(t, "../CGMES-Test-Configurations/v3.0/PST/PST_PhaseTapChangerLinear_Type1/PST_Type1_EQ.xml")

	violations := append(ValidateEQProfile(dataset), ValidateCommonRulesSPARQL(dataset)...)
	t.Logf("Focus node,Path,Constraint Component,Message,Severity")
	var errCount, infoEQBDCount int
	for _, v := range violations {
		t.Logf("%s,%s,%s,%s,%s", v.ObjectID, v.Property, v.Class, v.Message, v.Severity)
		if v.Severity == "sh.Violation" {
			errCount++
		}
		if v.Severity == "sh.Info" && strings.Contains(v.Message, "EQBD") {
			infoEQBDCount++
		}
	}
	t.Logf("Total: %d violations, %d non-violations (expected 0 violations, 1 PROF10-EQ sh.Info)", errCount, len(violations)-errCount)
	if errCount != 0 {
		t.Errorf("expected 0 sh.Violation findings for conformant EQ data, got %d", errCount)
	}
	if infoEQBDCount != 1 {
		t.Errorf("expected 1 PROF10-EQ sh.Info (missing EQBD ref), got %d", infoEQBDCount)
	}
}
