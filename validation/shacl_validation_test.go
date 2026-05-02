package validation

import (
	"bytes"
	"cimgo/cimgostructs"
	"cimgo/cimprofiles"
	"os"
	"strings"
	"testing"
)

func TestValidateCoordinateSystemCrsUrn(t *testing.T) {
	rules := loadAllRules(t,
		"../shacljson/struct-simplified/61968-13_GeographicalLocation-AP-Con-Complex-SHACL.json",
	)
	if len(rules) == 0 {
		t.Skip("No rules found")
	}

	dataFile := "../testdata/test_shacl_001_GL.xml"
	dataset := cimgostructs.NewCIMElementList()

	b, err := os.ReadFile(dataFile)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", dataFile, err)
	}
	cimprofiles.DecodeProfile(bytes.NewReader(b), dataset)

	t.Logf("Loaded %d elements", len(dataset.Elements))

	var violationsByID = map[string][]string{}
	for id, obj := range dataset.Elements {
		for _, v := range validateObject(t, obj, rules, dataset) {
			violationsByID[id] = append(violationsByID[id], v)
		}
	}

	if got := len(violationsByID["CoordinateSystem.WGS84"]); got != 0 {
		t.Errorf("CoordinateSystem.WGS84 (default crsUrn): expected 0 violations, got %d: %v",
			got, violationsByID["CoordinateSystem.WGS84"])
	}
	if got := len(violationsByID["CoordinateSystem.ETRS89"]); got != 1 {
		t.Errorf("CoordinateSystem.ETRS89 (non-default crsUrn): expected 1 violation, got %d: %v",
			got, violationsByID["CoordinateSystem.ETRS89"])
	}

	for id, vs := range violationsByID {
		for _, v := range vs {
			t.Logf("Object %s: %s", id, v)
		}
	}
}

func TestValidateDiagramObjectIdentifiedObject(t *testing.T) {
	// The rule says DiagramObject.IdentifiedObject must be an IRI and must NOT
	// point to a cim.GeneratingUnit (it should reference SynchronousMachine).
	rules := loadAllRules(t,
		"../shacljson/struct-simplified/61970-301_DiagramLayout-AP-Con-Complex-NotSolvedMAS-SHACL.json",
	)
	if len(rules) == 0 {
		t.Skip("No rules found")
	}

	dataFile := "../testdata/test_shacl_002_DL.xml"
	dataset := cimgostructs.NewCIMElementList()

	b, err := os.ReadFile(dataFile)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", dataFile, err)
	}
	cimprofiles.DecodeProfile(bytes.NewReader(b), dataset)

	t.Logf("Loaded %d elements", len(dataset.Elements))

	var violationsByID = map[string][]string{}
	for id, obj := range dataset.Elements {
		for _, v := range validateObject(t, obj, rules, dataset) {
			violationsByID[id] = append(violationsByID[id], v)
		}
	}

	if got := len(violationsByID["DiagramObject.OK"]); got != 0 {
		t.Errorf("DiagramObject.OK (points to SynchronousMachine): expected 0 violations, got %d: %v",
			got, violationsByID["DiagramObject.OK"])
	}
	for _, badID := range []string{"DiagramObject.BAD", "TextDiagramObject.BAD"} {
		if got := len(violationsByID[badID]); got != 1 {
			t.Errorf("%s (points to GeneratingUnit): expected 1 violation, got %d: %v",
				badID, got, violationsByID[badID])
		}
	}

	for id, vs := range violationsByID {
		for _, v := range vs {
			t.Logf("Object %s: %s", id, v)
		}
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
		"../shacljson/struct-simplified/61970-301_Equipment-AP-Con-Complex-SHACL.json",
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
