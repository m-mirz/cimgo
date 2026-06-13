package main

import (
	"cimgo/shaclimport"
	"fmt"
	"strings"
)

// skipEntry groups one or more focus classes that share the same skip reason.
type skipEntry struct {
	Classes   []string
	Prop      string
	Component string
	Name      string
	Reason    string
}

// uniqueCheckPatterns counts unique (PathKey, Component, RuleName) triples —
// the same deduplication key used for skip entries — so that the check count
// is comparable to the skip count regardless of how many classes share a pattern.
// PathKey is the full constraint path joined with "/" (matching cimoxide's dedup).
func uniqueCheckPatterns(checks []checkSpec) int {
	seen := map[string]bool{}
	for _, c := range checks {
		seen[c.PathKey+"\x00"+c.Component+"\x00"+c.RuleName] = true
	}
	return len(seen)
}

func dropsToSkipEntries(drops []shaclimport.SimplifiedDrop) []skipEntry {
	// Deduplicate by (prop, component, name): same as buildFileSpec's skipIndex.
	index := map[string]int{}
	var entries []skipEntry
	for _, d := range drops {
		key := d.Prop + "\x00" + d.Component + "\x00" + d.Name
		if i, ok := index[key]; ok {
			for _, c := range d.Classes {
				found := false
				for _, existing := range entries[i].Classes {
					if existing == c {
						found = true
						break
					}
				}
				if !found {
					entries[i].Classes = append(entries[i].Classes, c)
				}
			}
		} else {
			index[key] = len(entries)
			entries = append(entries, skipEntry{
				Classes:   d.Classes,
				Prop:      d.Prop,
				Component: d.Component,
				Name:      d.Name,
				Reason:    d.Reason,
			})
		}
	}
	return entries
}

func (s skipEntry) String() string {
	if len(s.Classes) == 1 {
		return fmt.Sprintf("%s%s [%s] %q: %s", s.Classes[0], s.Prop, s.Component, s.Name, s.Reason)
	}
	return fmt.Sprintf("%s [%s] %q (%s): %s",
		s.Prop, s.Component, s.Name, strings.Join(s.Classes, ", "), s.Reason)
}

// fileSpec is the data passed to validation_file.tmpl.
type fileSpec struct {
	FileName         string
	Pkg              string
	OrchestratorName string
	Imports          []string // sorted, deduplicated; always includes cimgo/cimstructs
	Checks           []checkSpec
}

// checkSpec describes a single Check<...> function. Guard and Condition carry
// the only varying pieces of the function body; everything else is fixed
// scaffolding emitted by the template. Decl is an optional package-level
// declaration emitted directly above the function (used by Pattern checks to
// hoist the compiled regexp out of the per-call hot path). Prelude is an
// optional block emitted before the main per-element loop — used by inverse
// path checks to build an O(N) cross-reference index once per Check, instead
// of paying O(N²) by scanning inside the loop. NoV switches the type assertion
// from `v, ok := ...` to `_, ok := ...` for checks that don't need the
// instance value (typical for inverse-path checks that only consume `id`).
type checkSpec struct {
	Name         string
	ShapeID      string // Original SHACL Shape ID (e.g. eqc:ACLineSegment.length-length)
	RuleID       string // Extracted Rule ID (e.g. eq600:ACLineSegment.length-length)
	RuleName     string
	Description  string
	Class        string
	Tag          string
	Component    string
	Property     string
	PathKey      string // full path joined with "/" for dedup only; not emitted into generated code
	Message      string
	Severity     string
	Decl         string // optional package-level declaration emitted before the function
	Prelude      string // optional block emitted before the main loop (e.g. inverse-ref index)
	NoV          bool   // suppress the v binding when the loop body doesn't use it
	Guard         string // tab-indented, may span multiple lines; empty if none
	Condition     string // single expression; opens the violation block as `if <Condition> {`
	DatasetCheck  bool   // emit a single dataset-level check (no per-element loop) when true
	SelfContained bool   // Guard appends violations directly; skip the outer `if Condition` block
}
