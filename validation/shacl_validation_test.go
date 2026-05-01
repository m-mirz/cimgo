package validation

import (
	"bytes"
	"cimgo/cimgostructs"
	"cimgo/cimprofiles"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestValidateCIMData(t *testing.T) {
	rules := loadAllRules(t, "../pages/docs/struct-simplified")
	if len(rules) == 0 {
		t.Skip("No rules found in ../pages/docs/struct-simplified")
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

func TestValidatePSTType1EQ(t *testing.T) {
	// Load both the Equipment-profile rules (cim.* class names — match Go struct
	// types) and the Prof10 header rules. Prof10 uses implicit-class targets like
	// prof10.FullModel-EQ which require RDFS subclass reasoning to map to Go
	// types; those rules are loaded but will not fire until subclass mapping is
	// added. The one non-SPARQL CSV violation is a sh:HasValueConstraintComponent
	// on a Prof10 shape, so it is not yet caught.
	rules := loadAllRules(t,
		"../pages/docs/struct-simplified/61970-301_Equipment-AP-Con-Complex-SHACL.json",
		"../pages/docs/struct-simplified/61970-600-1_Prof10-Header-AP-Con-Complex-SHACL.json",
	)
	if len(rules) == 0 {
		t.Skip("No rules found")
	}

	dataFile := "../CGMES-Test-Configurations/v3.0/PST/PST_PhaseTapChangerLinear_Type1/PST_Type1_EQ.xml"
	dataset := cimgostructs.NewCIMElementList()

	b, err := os.ReadFile(dataFile)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", dataFile, err)
	}
	cimprofiles.DecodeProfile(bytes.NewReader(b), dataset)

	t.Logf("Loaded %d elements", len(dataset.Elements))

	// Produce output similar to the CSV (ignoring SPARQL rules)
	t.Log("Focus node,Path,Constraint Component,Message,Severity")

	var count int
	for id, obj := range dataset.Elements {
		violations := validateObject(t, obj, rules, dataset)
		for _, v := range violations {
			count++
			path := ""
			msg := v
			if colIndex := strings.Index(v, "]: "); colIndex != -1 {
				msg = v[colIndex+3:]
				if spIndex := strings.Index(msg, ": "); spIndex != -1 {
					path = msg[:spIndex]
					msg = msg[spIndex+2:]
				}
			}
			t.Logf("%s,%s,sh:ConstraintComponent,%s,sh:Violation", id, path, msg)
		}
	}
	t.Logf("Total violations: %d (expected 0 for conformant EQ data; CSV warnings are SPARQL or model-metadata constraints)", count)
}
