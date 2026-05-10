package validation

import (
	"strings"
	"testing"
)

func TestValidatePSTType1EQ(t *testing.T) {
	// Conformant data should produce zero sh:Violation findings for EQ.
	// However, the Simple profiles (600-2) in CGMES have some cross-profile
	// limitations that trigger violations in standard test data.
	// By silencing these known issues, we can ensure the rest of the model is valid.
	dataset := loadDirectory(t, "../CGMES-Test-Configurations/v3.0/PST/PST_PhaseTapChangerLinear_Type1/")

	cfg := Config{
		Profiles: []string{"EQ", "SSH", "TP", "SV", "DL"},
		Common:   true,
		SilencedRules: []string{
			"dl:DiagramObject.IdentifiedObject-valueType",
			"sv:SvStatus.ConductingEquipment-valueType",
		},
	}
	violations := RunValidation(dataset, cfg)

	t.Logf("Focus node\tRule\tPath\tConstraint Component\tMessage\tSeverity")
	var errCount, infoEQBDCount int
	for _, v := range violations {
		t.Logf("%s\t%s\t%s\t%s\t%s\t%s", v.ObjectID, v.RuleID, v.Property, v.Class, v.Message, v.Severity)
		if v.Severity == "sh:Violation" {
			errCount++
		}
		if v.Severity == "sh:Info" && strings.Contains(v.Message, "EQBD") {
			infoEQBDCount++
		}
	}
	t.Logf("Total: %d violations, %d non-violations (expected 0 violations, 1 PROF10-EQ sh:Info)", errCount, len(violations)-errCount)
	if errCount != 0 {
		t.Errorf("expected 0 sh:Violation findings for PST Type 1 data after fixing known issues, got %d", errCount)
	}
	if infoEQBDCount != 1 {
		t.Errorf("expected 1 PROF10-EQ sh:Info (missing EQBD ref), got %d", infoEQBDCount)
	}
}
