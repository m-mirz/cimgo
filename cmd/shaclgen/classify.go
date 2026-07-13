package main

import (
	"fmt"
	"io"
	"strings"
)

type skipCategory struct {
	Label   string
	Section string // "simplified" | "skipped" | "cannot_be_conducted" | "other"
	match   func(e skipEntry) bool
}

func contains(s, sub string) bool { return strings.Contains(s, sub) }

// classesContain reports whether any label in classes contains sub. Used to
// inspect the "(kind1,kind2,...)" target-kind label pushUnsupportedTargetSkips
// stores in skipEntry.Classes for shapes whose target isn't a concrete class.
func classesContain(classes []string, sub string) bool {
	for _, c := range classes {
		if contains(c, sub) {
			return true
		}
	}
	return false
}

var skipCategories = []skipCategory{
	// Simplified — dropped in SimplifyFileResults before code generation
	{
		Label:   "`sh:nodeKind` simplified (type-system guarantee)",
		Section: "simplified",
		match: func(e skipEntry) bool {
			return contains(e.Reason, "NodeKind=Literal") || contains(e.Reason, "NodeKind structurally satisfied by")
		},
	},
	{
		Label:   "`sh:datatype` simplified (native Go type)",
		Section: "simplified",
		match:   func(e skipEntry) bool { return contains(e.Reason, "Datatype structurally satisfied") },
	},
	{
		Label:   "`sh:minCount=0` vacuously true",
		Section: "simplified",
		match:   func(e skipEntry) bool { return contains(e.Reason, "MinCount=0 vacuously true") },
	},
	{
		Label:   "`sh:optional` (minCount=0 + maxCount=1) structurally satisfied",
		Section: "simplified",
		match: func(e skipEntry) bool {
			return contains(e.Reason, "Optional") && contains(e.Reason, "structurally satisfied")
		},
	},
	// Skipped — ordered to match README table rows
	{
		Label:   "`sh:maxCount 1` on multi-hop paths",
		Section: "skipped",
		match: func(e skipEntry) bool {
			return contains(e.Reason, "multi-segment MaxCount=1 is structurally satisfied")
		},
	},
	{
		Label:   "`sh:nodeKind` on `rdf:type` paths",
		Section: "skipped",
		match:   func(e skipEntry) bool { return contains(e.Reason, "NodeKind on path ending in rdf:type") },
	},
	{
		Label:   "`sh:class` vacuously true (inverse-index already type-asserts)",
		Section: "skipped",
		match: func(e skipEntry) bool {
			return contains(e.Reason, "inverse Class") && contains(e.Reason, "structurally satisfied")
		},
	},
	{
		Label:   "Cross-class `sh:lessThan` on sibling subtypes",
		Section: "skipped",
		match:   func(e skipEntry) bool { return contains(e.Reason, "sibling class") },
	},
	{
		Label:   "`sh:hasValue rdf:type rdf:Statements`",
		Section: "skipped",
		match: func(e skipEntry) bool {
			return contains(e.Reason, "HasValue rdf:type") && contains(e.Reason, "rdf:Statements")
		},
	},
	{
		Label:   "`sh:datatype xsd:anyURI` on slice fields",
		Section: "skipped",
		match: func(e skipEntry) bool {
			return contains(e.Reason, "anyURI") && contains(e.Reason, "slice")
		},
	},
	{
		Label:   "Multi-segment `sh:required` on `rdf:Statements`",
		Section: "skipped",
		match:   func(e skipEntry) bool { return contains(e.Reason, "multi-segment Required path ending in rdf:") },
	},

	// Cannot be conducted — ordered to match README table rows
	{
		Label:   "Field name typos (`sh:lessThan`)",
		Section: "cannot_be_conducted",
		match: func(e skipEntry) bool {
			return contains(e.Reason, "paired field xml tag") && contains(e.Reason, "not found")
		},
	},
	{
		Label:   "Class name typo (Dyanmics)",
		Section: "cannot_be_conducted",
		match:   func(e skipEntry) bool { return contains(e.Reason, "Dyanmics") },
	},
	{
		Label:   "Class name capitalisation mismatch (CSConverter)",
		Section: "cannot_be_conducted",
		match: func(e skipEntry) bool {
			return contains(e.Reason, "CSConverter") && contains(e.Reason, "no Go struct")
		},
	},
	{
		Label:   "Wrong field names in inverse paths",
		Section: "cannot_be_conducted",
		match: func(e skipEntry) bool {
			return contains(e.Prop, "^") && contains(e.Reason, "no field with xml tag")
		},
	},
	{
		Label:   "Stale field reference",
		Section: "cannot_be_conducted",
		match:   func(e skipEntry) bool { return contains(e.Reason, "AccumulatorValue.value") },
	},
	{
		Label:   "Empty `sh:in` list",
		Section: "cannot_be_conducted",
		match:   func(e skipEntry) bool { return contains(e.Reason, "In payload is empty") },
	},

	// Other
	{
		// Checked ahead of the SPARQL-derived rule below so that target-kind takes
		// precedence over the underlying constraint's own component: a handful of
		// these (e.g. C:600:ALL:NA:FBOD4, the IdentifiedObject.* stringLength rules)
		// happen to have Component == sh:SPARQLConstraintComponent too, since the
		// CGMES rule they express is itself SPARQL-derived -- but the reason they
		// can't be code-generated here is the unresolvable targetSubjectsOf/
		// targetObjectsOf target, not the component type, so they're classified by
		// target kind first. This matches cimoxide's precedence (codegen.rs's
		// push_unsupported_target_skips matches on target kind before anything else):
		// see the identically-named entries in both tools' `--skip-report` output for
		// the same 5 sh:names.
		Label:   "Unsupported `sh:target` kind (`targetSubjectsOf`/`targetObjectsOf`)",
		Section: "other",
		match: func(e skipEntry) bool {
			return contains(e.Reason, "unsupported SHACL target kind") &&
				(classesContain(e.Classes, "targetSubjectsOf") || classesContain(e.Classes, "targetObjectsOf"))
		},
	},
	{
		// See SPARQL Check Coverage in the README: this total isn't directly
		// comparable to that table's TTL Total, even though both are "how much
		// SPARQL is there" counts -- this one is every distinct (property,
		// component, sh:name) skip entry, deduped per TTL file (a fresh
		// skipIndex per buildFileSpec call) and *not* split on "|" for compound
		// sh:name values, whereas ttl_report.go's sh:name-based count is deduped
		// per profile group across all its files and does split on "|".
		//
		// Also catches sh:sparql-based sh:target shapes (unsupportedTargetKinds'
		// "sparqlTarget" label) -- those need a hand-written implementation for the
		// exact same reason plain sh:sparql constraints do (no SPARQL evaluator),
		// so they belong in this bucket rather than the target-kind rule above.
		// This mirrors cimoxide's grouping (cimgen/src/shacl/skip.rs), which folds
		// sh:target SPARQLTarget into its "SPARQL-derived constraints" row.
		Label:   "SPARQL-derived constraints (see SPARQL Check Coverage below)",
		Section: "other",
		match: func(e skipEntry) bool {
			if e.Component == "sh:SPARQLConstraintComponent" {
				return true
			}
			return contains(e.Reason, "unsupported SHACL target kind") && classesContain(e.Classes, "sparqlTarget")
		},
	},
}

// skipCategoryOther is a safety-net fallback; every skip entry is expected to
// match one of the categories above (see -skip-report's global total check).
var skipCategoryOther = skipCategory{Label: "Unclassified", Section: "other"}

func classify(e skipEntry) *skipCategory {
	for i := range skipCategories {
		if skipCategories[i].match(e) {
			return &skipCategories[i]
		}
	}
	return &skipCategoryOther
}

func accumulateCounts(counts map[string]int, entries []skipEntry) {
	for _, e := range entries {
		counts[classify(e).Label]++
	}
}

func printFileSummary(w io.Writer, fileName string, checks int, entries []skipEntry) {
	fmt.Fprintf(w, "-- %s (%d checks, %d skipped) --\n", fileName, checks, len(entries))
	if len(entries) == 0 {
		return
	}
	counts := map[string]int{}
	accumulateCounts(counts, entries)
	for _, cat := range append(skipCategories, skipCategoryOther) {
		n := counts[cat.Label]
		if n > 0 {
			fmt.Fprintf(w, "  %5d  %s\n", n, cat.Label)
		}
	}
}

func printGlobalSummary(w io.Writer, counts map[string]int) {
	sections := []struct {
		title string
		key   string
	}{
		{"Simplified (type-system guarantees)", "simplified"},
		{"Skipped", "skipped"},
		{"Cannot be conducted", "cannot_be_conducted"},
		{"Other", "other"},
	}
	for _, sec := range sections {
		total := 0
		var lines []string
		for _, cat := range append(skipCategories, skipCategoryOther) {
			if cat.Section != sec.key {
				continue
			}
			n := counts[cat.Label]
			if n > 0 {
				lines = append(lines, fmt.Sprintf("  %5d  %s", n, cat.Label))
				total += n
			}
		}
		if total == 0 {
			continue
		}
		fmt.Fprintf(w, "\n=== %s ===\n", sec.title)
		for _, l := range lines {
			fmt.Fprintln(w, l)
		}
		fmt.Fprintf(w, "  -----\n  %5d  total\n", total)
	}
}
