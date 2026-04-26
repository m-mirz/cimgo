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
	t.Run("ACDCTerminalSequenceNumbering", func(t *testing.T) {
		violations := CheckACDCTerminalSequenceNumbering(dataset)
		if len(violations) > 0 {
			t.Logf("Found %d violations", len(violations))
			for _, v := range violations {
				t.Logf("Violation: %+v", v)
			}
		} else {
			t.Log("No violations found")
		}
	})

	t.Run("TerminalPhasesConsistency", func(t *testing.T) {
		violations := CheckTerminalPhasesConsistencyEquipment(dataset)
		if len(violations) > 0 {
			t.Logf("Found %d violations", len(violations))
			for _, v := range violations {
				t.Logf("Violation: %+v", v)
			}
		} else {
			t.Log("No violations found")
		}
	})

	t.Run("ConductingEquipmentBaseVoltageUsage", func(t *testing.T) {
		violations := CheckConductingEquipmentBaseVoltageUsage(dataset)
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
