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
		match:   func(e skipEntry) bool { return contains(e.Reason, "Optional") && contains(e.Reason, "structurally satisfied") },
	},
	// Skipped — ordered to match README table rows
	{
		Label:   "`sh:required` on `float` fields",
		Section: "skipped",
		match:   func(e skipEntry) bool { return contains(e.Reason, "float Required is unreliable") },
	},
	{
		Label:   "`sh:maxCount 1` on scalar fields",
		Section: "skipped",
		match:   func(e skipEntry) bool { return contains(e.Reason, "MaxCount=1 on scalar field is structurally satisfied") },
	},
	{
		Label:   "`sh:required` on `bool` fields",
		Section: "skipped",
		match:   func(e skipEntry) bool { return contains(e.Reason, "bool Required is structurally satisfied") },
	},
	{
		Label:   "`sh:maxCount 1` on pointer fields",
		Section: "skipped",
		match:   func(e skipEntry) bool { return contains(e.Reason, "MaxCount=1 on pointer field is structurally satisfied") },
	},
	{
		Label:   "`sh:maxCount 1` on multi-hop paths",
		Section: "skipped",
		match:   func(e skipEntry) bool { return contains(e.Reason, "multi-segment MaxCount=1 is structurally satisfied") },
	},
	{
		Label:   "`sh:nodeKind` on `rdf:type` paths",
		Section: "skipped",
		match:   func(e skipEntry) bool { return contains(e.Reason, "NodeKind on path ending in rdf:type") },
	},
	{
		Label:   "Inverse `sh:class`",
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
		Label:   "SPARQL (not in README)",
		Section: "other",
		match:   func(e skipEntry) bool { return e.Component == "sh:SPARQLConstraintComponent" },
	},
}

var skipCategoryOther = skipCategory{Label: "Other unsupported", Section: "other"}

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
		{"Other (not in README)", "other"},
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
