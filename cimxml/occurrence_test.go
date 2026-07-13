package cimxml

import (
	"strings"
	"testing"
)

type occTestElem struct {
	Bch float64 `xml:"Class.bch,omitempty"`
}

// TestRecordOccurrence is the Phase 1 smoke test: a duplicated scalar XML tag
// must be counted twice in FieldOccurrences, isolating decoder-level
// correctness before cgmesxml wiring is added on top.
func TestRecordOccurrence(t *testing.T) {
	xmlDoc := `<Root><Class.bch>1</Class.bch><Class.bch>2</Class.bch></Root>`
	dec := NewDecoder(strings.NewReader(xmlDoc))
	dec.CurrentMRID = "mrid1"

	var v occTestElem
	if err := dec.Decode(&v); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if got := dec.FieldOccurrences["mrid1"]["Class.bch"]; got != 2 {
		t.Fatalf("FieldOccurrences[mrid1][Class.bch] = %d, want 2", got)
	}
	if v.Bch != 2 {
		t.Fatalf("Bch = %v, want 2 (last-write-wins, unchanged decode behavior)", v.Bch)
	}
}

// TestRecordOccurrenceNoOpWithoutMRID confirms recordOccurrence is a no-op
// when CurrentMRID is unset, so direct/low-level Decoder use (not going
// through cgmesxml) is unaffected.
func TestRecordOccurrenceNoOpWithoutMRID(t *testing.T) {
	xmlDoc := `<Root><Class.bch>1</Class.bch></Root>`
	dec := NewDecoder(strings.NewReader(xmlDoc))

	var v occTestElem
	if err := dec.Decode(&v); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(dec.FieldOccurrences) != 0 {
		t.Fatalf("FieldOccurrences = %v, want empty (no CurrentMRID set)", dec.FieldOccurrences)
	}
}
