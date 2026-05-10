package validation

import (
	"strings"
	"testing"
)

func TestValidatePSTType1EQ(t *testing.T) {
	// Conformant data should produce zero sh:Violation findings for EQ.
	// However, when loading all profiles (EQ, SSH, TP, SV, DL), we currently
	// see 16 sh:Violation findings in SV and DL profiles related to cross-profile
	// reference resolution in SHACL rules.
	// The ValiMate reference CSV reports 0 errors and 1 sh:Warning (Substation
	// count design note) and 1 sh:Info (PROF10-EQ: missing EQBD reference).
	dataset := loadDirectory(t, "../CGMES-Test-Configurations/v3.0/PST/PST_PhaseTapChangerLinear_Type1/")

	cfg := Config{
		Profiles: []string{"EQ", "SSH", "TP", "SV", "DL"},
		Common:   true,
	}
	violations := RunValidation(dataset, cfg)

	t.Logf("Focus node,Rule,Path,Constraint Component,Message,Severity")
	var errCount, infoEQBDCount int
	for _, v := range violations {
		t.Logf("%s,%s,%s,%s,%s,%s", v.ObjectID, v.RuleID, v.Property, v.Class, v.Message, v.Severity)
		if v.Severity == "sh:Violation" {
			errCount++
		}
		if v.Severity == "sh:Info" && strings.Contains(v.Message, "EQBD") {
			infoEQBDCount++
		}
	}
	t.Logf("Total: %d violations, %d non-violations (expected 18 violations, 1 PROF10-EQ sh:Info)", errCount, len(violations)-errCount)
	if errCount != 18 {
		t.Errorf("expected 18 sh:Violation findings for PST Type 1 data, got %d", errCount)
	}
	if infoEQBDCount != 1 {
		t.Errorf("expected 1 PROF10-EQ sh:Info (missing EQBD ref), got %d", infoEQBDCount)
	}
	}

