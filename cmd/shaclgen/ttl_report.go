// TTL-side half of the SPARQL Check Coverage report: classifies each CGMES
// SHACL TTL file into the same profile groups as rule_report.go's
// sparqlGroups, and counts the distinct sh:sparql constraint shapes each
// group actually defines -- the denominator for "how many of the SPARQL
// constraints in the standard are implemented", as opposed to how many Go
// check functions exist. Combined with sparqlGroupReport's Names (matched
// against sh:name) in combineCoverage to produce a genuine
// Implemented/Total/Coverage figure per group.
package main

import (
	"cimgo/shaclimport"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// ttlGroupLabel classifies a SHACL TTL file's basename into one of
// sparqlGroups' labels, by simple substring match on the CGMES profile name
// embedded in the filename (e.g. "61970-452_Equipment-AP-Con-Complex-SHACL.ttl").
// Order matters: check the more specific profile names first. "Equipment"
// also matches "EquipmentBoundary" files, mirroring how
// sparql_equipmentboundary_notsolvedmas.go is grouped under Equipment (EQ) on
// the Go side. Used by both the "Generated SHACL Rules by Profile" table (via
// the caller in main.go) and the SPARQL Check Coverage table (via
// ttlSparqlNames below), so their rows line up: "Common / AllProfiles"
// absorbs "C:600 conformance" (both are cross-cutting, not tied to one
// profile) as well as AllProfiles/IdentifiedObjectCommon/GeographicalLocation/
// the plain Header file (none of which has its own profile group on the
// hand-written side either), and Topology/DiagramLayout/Operation each get
// their own row instead of being bundled into a generic "Others".
func ttlGroupLabel(filename string) string {
	switch {
	case strings.Contains(filename, "Equipment"):
		return "Equipment (EQ)"
	case strings.Contains(filename, "SteadyStateHypothesis"):
		return "Steady State Hypothesis (SSH)"
	case strings.Contains(filename, "Dynamics"):
		return "Dynamics (DY)"
	case strings.Contains(filename, "StateVariables"):
		return "State Variables (SV)"
	case strings.Contains(filename, "ShortCircuit"):
		return "Short Circuit (SC)"
	case strings.Contains(filename, "Topology"):
		return "Topology (TP)"
	case strings.Contains(filename, "DiagramLayout"):
		return "DiagramLayout (DL)"
	case strings.Contains(filename, "Operation"):
		return "Operation (OP)"
	default:
		// Prof10, AllProfiles, IdentifiedObjectCommon, GeographicalLocation, the
		// plain Header file, ... -- everything cross-cutting or without its own
		// hand-written profile group folds into the catch-all "Common /
		// AllProfiles" bucket.
		return "Common / AllProfiles"
	}
}

// ttlGroupLabelOrder lists every label ttlGroupLabel can return, in the fixed
// display order used for per-profile report tables (map iteration order in Go
// is randomized, so callers printing one row per group need this).
var ttlGroupLabelOrder = []string{
	"Equipment (EQ)",
	"Steady State Hypothesis (SSH)",
	"Dynamics (DY)",
	"State Variables (SV)",
	"Short Circuit (SC)",
	"Common / AllProfiles",
	"Topology (TP)",
	"DiagramLayout (DL)",
	"Operation (OP)",
}

// ttlSparqlNames globs shaclGlob and, for every SHACL TTL file, collects the
// sh:name of every shape carrying at least one sh:sparql constraint, grouped
// by ttlGroupLabel. A shape can appear more than once in the parsed tree
// (once per resolved concrete target class); the returned sets dedupe by
// name. Matching on sh:name (a plain string) rather than the shape's IRI
// sidesteps namespace-prefix normalization entirely -- the hand-written
// Go RuleID's prefix can legitimately differ from shaclimport's canonical
// one (see shaclimport/pipeline.go's `prefixes` map), but sh:name is copied
// verbatim into Violation.Name on both sides.
//
// A single shape's sh:name can itself be a "|"-joined compound of several
// rule names when one SPARQL constraint enforces multiple named conformance
// rules at once (e.g. 61970-600-2_IdentifiedObjectCommon_AP-Con-Complex-SHACL.ttl's
// "C:301:EQ:IdentifiedObject.shortName:stringLength|C:301:EQBD:...|...") --
// this is how the standard itself expresses it, not a formatting quirk, and
// the hand-written implementation may legitimately cover only one of the
// joined names in a single Violation. Each "|"-separated part is recorded as
// its own candidate name; empty parts (an upstream authoring gap in at least
// one shape) are skipped.
func ttlSparqlNames(shaclGlob string) (map[string]map[string]bool, error) {
	matches, err := filepath.Glob(shaclGlob)
	if err != nil {
		return nil, fmt.Errorf("invalid shacl glob %q: %w", shaclGlob, err)
	}
	sort.Strings(matches)

	groups := map[string]map[string]bool{}
	var collect func(group string, s shaclimport.ShapeInfo)
	collect = func(group string, s shaclimport.ShapeInfo) {
		for _, c := range s.Constraints {
			if c.IsSPARQL() {
				for part := range strings.SplitSeq(s.Name, "|") {
					if part != "" {
						groups[group][part] = true
					}
				}
				break
			}
		}
		for _, p := range s.Properties {
			collect(group, p)
		}
	}

	for _, path := range matches {
		if filepath.Ext(path) != ".ttl" {
			continue
		}
		fr, err := shaclimport.ProcessFileToResults(path)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		group := ttlGroupLabel(filepath.Base(path))
		if groups[group] == nil {
			groups[group] = map[string]bool{}
		}
		for _, s := range fr.Shapes {
			collect(group, s)
		}
	}
	return groups, nil
}

// coverageRow is one printed row of the combined TTL-vs-hand-written report.
type coverageRow struct {
	Label       string
	Implemented int      // distinct sh:names implemented that are also defined in the TTL
	TTLTotal    int      // distinct SPARQL constraint shapes defined in the TTL for this group; -1 if the group has no TTL backing (e.g. CIMdesk quality)
	Missing     []string // TTL sh:names with no matching implemented name, sorted
}

// combineCoverage matches each Go profile group's implemented sh:names
// against the TTL-derived constraint sets, producing the
// Implemented/Total/Coverage figures that replace the old hand-maintained
// "SPARQL Constraints | Implemented | Coverage" table.
func combineCoverage(groups []sparqlGroupReport, ttl map[string]map[string]bool) []coverageRow {
	rows := make([]coverageRow, 0, len(groups))
	for _, g := range groups {
		row := coverageRow{Label: g.Label, TTLTotal: -1}
		implemented := map[string]bool{}
		for _, n := range g.Names {
			implemented[n] = true
		}

		set, ok := ttl[g.Label]
		if !ok {
			row.Implemented = len(implemented)
			rows = append(rows, row)
			continue
		}
		row.TTLTotal = len(set)
		for name := range set {
			if implemented[name] {
				row.Implemented++
			} else {
				row.Missing = append(row.Missing, name)
			}
		}
		sort.Strings(row.Missing)
		rows = append(rows, row)
	}
	return rows
}
