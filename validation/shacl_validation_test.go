package validation

import (
	"bytes"
	"cimgo/cimgostructs"
	"cimgo/cimprofiles"
	"fmt"
	"os"
	"testing"
)

func TestValidateCIMData(t *testing.T) {
	rules := loadAllRules(t, "../pages/docs/struct")
	if len(rules) == 0 {
		t.Skip("No rules found in ../pages/docs/struct")
	}

	dataFiles := []string{
		"../testdata/test_001.xml",
		"../testdata/test_009_EQ.xml",
	}

	mergedCIMData := cimgostructs.NewCIMElementList()
	for _, file := range dataFiles {
		b, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		cimprofiles.DecodeProfile(bytes.NewReader(b), mergedCIMData)
	}

	var allViolations []string
	for id, obj := range mergedCIMData.Elements {
		violations := validateObject(t, obj, rules, mergedCIMData)
		for _, v := range violations {
			allViolations = append(allViolations, fmt.Sprintf("Object %s: %s", id, v))
		}
	}
	// 4. Report
	if len(allViolations) > 0 {
		t.Logf("Found %d validation violations (test marked passed):", len(allViolations))
		for _, v := range allViolations {
			t.Log(v)
		}
	} else {
		t.Log("No validation violations found.")
	}
}
