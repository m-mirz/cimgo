package validation

import (
	"bytes"
	"cimgo/cimstructs"
	"cimgo/cgmesxml"
	"cimgo/shaclgen"
	"os"
	"testing"
)

// TestOccurrenceFalsePositiveRegression is the whole point of the
// FieldOccurrences fix: a float/bool field present exactly once with a
// legitimately-zero value must not be mistaken for "absent".
func TestOccurrenceFalsePositiveRegression(t *testing.T) {
	dataset := loadDataset(t, "../testdata/test_occurrence_regression.xml")

	if got := len(shaclgen.CheckShortcircuit619706002SimpleACLineSegmentB0chRequired(dataset)); got != 1 {
		t.Fatalf("expected exactly 1 b0ch-Required violation (only ACL.B0CH.MISSING), got %d", got)
	}
	byID := indexByID(shaclgen.CheckShortcircuit619706002SimpleACLineSegmentB0chRequired(dataset))
	if got := len(byID["ACL.B0CH.ZERO.OK"]); got != 0 {
		t.Errorf("ACL.B0CH.ZERO.OK (b0ch=0, present once): expected 0 violations, got %d: %v", got, byID["ACL.B0CH.ZERO.OK"])
	}

	byID = indexByID(shaclgen.CheckShortcircuit619706002SimplePowerTransformerEndGroundedRequired(dataset))
	if got := len(byID["PTE.GROUNDED.FALSE.OK"]); got != 0 {
		t.Errorf("PTE.GROUNDED.FALSE.OK (grounded=false, present once): expected 0 violations, got %d: %v", got, byID["PTE.GROUNDED.FALSE.OK"])
	}
}

// TestOccurrenceTruePositiveMissingField confirms a genuinely absent
// float/bool tag still fires sh:required.
func TestOccurrenceTruePositiveMissingField(t *testing.T) {
	dataset := loadDataset(t, "../testdata/test_occurrence_regression.xml")

	byID := indexByID(shaclgen.CheckShortcircuit619706002SimpleACLineSegmentB0chRequired(dataset))
	if got := len(byID["ACL.B0CH.MISSING"]); got != 1 {
		t.Errorf("ACL.B0CH.MISSING: expected 1 violation, got %d: %v", got, byID["ACL.B0CH.MISSING"])
	}

	byID = indexByID(shaclgen.CheckShortcircuit619706002SimplePowerTransformerEndGroundedRequired(dataset))
	if got := len(byID["PTE.GROUNDED.MISSING"]); got != 1 {
		t.Errorf("PTE.GROUNDED.MISSING: expected 1 violation, got %d: %v", got, byID["PTE.GROUNDED.MISSING"])
	}
}

// TestOccurrenceTruePositiveDuplicateScalar confirms a repeated scalar
// (string) XML tag fires sh:maxCount 1.
func TestOccurrenceTruePositiveDuplicateScalar(t *testing.T) {
	dataset := loadDataset(t, "../testdata/test_occurrence_regression.xml")

	byID := indexByID(shaclgen.CheckDiagramlayout619706002SimpleDiagramDescriptionMaxCount(dataset))
	if got := len(byID["DIAGRAM.DESC.DUP"]); got != 1 {
		t.Errorf("DIAGRAM.DESC.DUP (description repeated): expected 1 violation, got %d: %v", got, byID["DIAGRAM.DESC.DUP"])
	}
	if got := len(byID["DIAGRAM.DESC.OK"]); got != 0 {
		t.Errorf("DIAGRAM.DESC.OK (description present once): expected 0 violations, got %d: %v", got, byID["DIAGRAM.DESC.OK"])
	}
}

// TestOccurrenceTruePositiveDuplicatePointer confirms a repeated
// pointer/reference XML tag fires sh:maxCount 1.
func TestOccurrenceTruePositiveDuplicatePointer(t *testing.T) {
	dataset := loadDataset(t, "../testdata/test_occurrence_regression.xml")

	byID := indexByID(shaclgen.CheckDiagramlayout619706002SimpleDiagramDiagramStyleMaxCount(dataset))
	if got := len(byID["DIAGRAM.STYLE.DUP"]); got != 1 {
		t.Errorf("DIAGRAM.STYLE.DUP (DiagramStyle repeated): expected 1 violation, got %d: %v", got, byID["DIAGRAM.STYLE.DUP"])
	}
	if got := len(byID["DIAGRAM.STYLE.OK"]); got != 0 {
		t.Errorf("DIAGRAM.STYLE.OK (DiagramStyle present once): expected 0 violations, got %d: %v", got, byID["DIAGRAM.STYLE.OK"])
	}
}

// TestOccurrenceCrossFileMergeOverwrite pins down the resolved cross-file
// merge semantics: the same mRID + tag appearing once in each of two profile
// files must NOT sum to 2 — mergeOccurrences overwrites, matching
// cimbase.DeepMerge's own "later file wins" policy for the scalar value
// itself (see plan doc's "Resolved" section).
func TestOccurrenceCrossFileMergeOverwrite(t *testing.T) {
	dataset := cimstructs.NewCIMDataset()

	b1, err := os.ReadFile("../testdata/test_occurrence_merge_file1.xml")
	if err != nil {
		t.Fatalf("read file1: %v", err)
	}
	if _, err := cgmesxml.DecodeProfile(bytes.NewReader(b1), dataset); err != nil {
		t.Fatalf("decode file1: %v", err)
	}

	b2, err := os.ReadFile("../testdata/test_occurrence_merge_file2.xml")
	if err != nil {
		t.Fatalf("read file2: %v", err)
	}
	if _, err := cgmesxml.DecodeProfile(bytes.NewReader(b2), dataset); err != nil {
		t.Fatalf("decode file2: %v", err)
	}

	got := dataset.FieldOccurrences["ACL.CROSSFILE"]["ACLineSegment.b0ch"]
	if got != 1 {
		t.Errorf("FieldOccurrences[ACL.CROSSFILE][ACLineSegment.b0ch] = %d, want 1 (overwrite, not sum across files)", got)
	}
}

// TestOccurrenceEmbeddedFieldRegression confirms occurrence tracking works
// for a field declared on an embedded/superclass struct
// (RatioTapChanger.ControlEnabled, tagged "TapChanger.controlEnabled" on the
// embedded TapChanger), not just fields declared directly on the concrete
// class.
func TestOccurrenceEmbeddedFieldRegression(t *testing.T) {
	dataset := loadDataset(t, "../testdata/test_occurrence_regression.xml")

	if got := dataset.FieldOccurrences["RTC.CONTROLENABLED.DUP"]["TapChanger.controlEnabled"]; got != 2 {
		t.Errorf("FieldOccurrences[RTC.CONTROLENABLED.DUP][TapChanger.controlEnabled] = %d, want 2", got)
	}
	if got := dataset.FieldOccurrences["RTC.CONTROLENABLED.OK"]["TapChanger.controlEnabled"]; got != 1 {
		t.Errorf("FieldOccurrences[RTC.CONTROLENABLED.OK][TapChanger.controlEnabled] = %d, want 1", got)
	}
}
