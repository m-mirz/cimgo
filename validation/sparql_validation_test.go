package validation

import (
	"bytes"
	"cimgo/cimgostructs"
	"cimgo/cimprofiles"
	"os"
	"testing"
)

func TestSPARQLRules(t *testing.T) {
	// 1. Load data
	dataFiles := []string{
		"../testdata/test_001.xml",
		"../testdata/test_009_EQ.xml",
	}

	dataset := cimgostructs.NewCIMElementList()
	for _, file := range dataFiles {
		b, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", file, err)
		}
		_, err = cimprofiles.DecodeProfile(bytes.NewReader(b), dataset)
		if err != nil {
			t.Fatalf("Failed to decode %s: %v", file, err)
		}
	}

	t.Logf("Loaded %d elements", len(dataset.Elements))

	// 2. Run validations
	t.Run("EquipmentProfile", func(t *testing.T) {
		violations := ValidateEquipmentProfile(dataset)
		if len(violations) > 0 {
			t.Logf("Found %d violations", len(violations))
			for _, v := range violations {
				t.Logf("Violation: %+v", v)
			}
		} else {
			t.Log("No violations found")
		}
	})

	t.Run("SSHProfile", func(t *testing.T) {
		violations := ValidateSSHProfile(dataset)
		if len(violations) > 0 {
			t.Logf("Found %d violations", len(violations))
			for _, v := range violations {
				t.Logf("Violation: %+v", v)
			}
		} else {
			t.Log("No violations found")
		}
	})

	t.Run("DynamicsProfile", func(t *testing.T) {
		violations := ValidateDynamicsProfile(dataset)
		if len(violations) > 0 {
			t.Logf("Found %d violations", len(violations))
			for _, v := range violations {
				t.Logf("Violation: %+v", v)
			}
		} else {
			t.Log("No violations found")
		}
	})

	t.Run("ShortCircuitProfile", func(t *testing.T) {
		violations := ValidateShortCircuitProfile(dataset)
		if len(violations) > 0 {
			t.Logf("Found %d violations", len(violations))
			for _, v := range violations {
				t.Logf("Violation: %+v", v)
			}
		} else {
			t.Log("No violations found")
		}
	})

	t.Run("StateVariablesProfile", func(t *testing.T) {
		violations := ValidateStateVariablesProfile(dataset)
		if len(violations) > 0 {
			t.Logf("Found %d violations", len(violations))
			for _, v := range violations {
				t.Logf("Violation: %+v", v)
			}
		} else {
			t.Log("No violations found")
		}
	})
}
