package main

import (
	"cimgo/rdf/shacl"
	"fmt"
	"regexp"
	"strings"
)

type fileStats struct {
	name         string
	shaclPath    string
	sparqlPath   string
	shaclCounts  map[string]int
	sparqlCounts map[string]int
}

type ConstraintWrapper struct {
	Type string
	Data shacl.Constraint
}

func (cw ConstraintWrapper) IsSPARQL() bool {
	_, ok := cw.Data.(*shacl.SPARQLConstraint)
	return ok
}

func (cw ConstraintWrapper) IsSHACL() bool {
	return !cw.IsSPARQL()
}

type ShapeWrapper struct {
	*shacl.Shape
	Constraints []ConstraintWrapper
	Properties  []*ShapeWrapper
}

func wrapShape(s *shacl.Shape) *ShapeWrapper {
	if s == nil {
		return nil
	}
	sw := &ShapeWrapper{Shape: s}
	for _, c := range s.Constraints {
		sw.Constraints = append(sw.Constraints, ConstraintWrapper{Type: c.ComponentIRI(), Data: c})
	}
	for _, ps := range s.Properties {
		sw.Properties = append(sw.Properties, wrapShape(ps))
	}
	return sw
}

var prefixes = map[string]string{
	"http://www.w3.org/1999/02/22-rdf-syntax-ns#": "rdf",
	"http://www.w3.org/2000/01/rdf-schema#":       "rdfs",
	"http://www.w3.org/2001/XMLSchema#":           "xsd",
	"http://www.w3.org/ns/shacl#":                 "sh",
	"http://www.w3.org/2002/07/owl#":              "owl",
	"http://iec.ch/TC57/CIM100#":                  "cim",
}

func simplifyIRI(iri string) string {
	for ns, pref := range prefixes {
		if strings.HasPrefix(iri, ns) {
			return pref + ":" + strings.TrimPrefix(iri, ns)
		}
	}
	if strings.HasPrefix(iri, "http") {
		return "<" + iri + ">"
	}
	return iri
}

func simplifyTerm(t shacl.Term) string {
	if t.IsIRI() {
		return simplifyIRI(t.Value())
	}
	return t.String()
}

func formatValue(v any) string {
	switch val := v.(type) {
	case map[string]any:
		if kind, ok := val["kind"].(string); ok {
			vIRI, _ := val["value"].(string)
			if kind == "IRI" {
				return simplifyIRI(vIRI)
			}
			return vIRI
		}
	case []any:
		var items []string
		for _, item := range val {
			items = append(items, formatValue(item))
		}
		return "[" + strings.Join(items, ", ") + "]"
	}
	return fmt.Sprint(v)
}

func formatPath(p *shacl.PropertyPath) string {
	if p == nil {
		return ""
	}
	switch p.Kind {
	case shacl.PathPredicate:
		return simplifyTerm(p.Pred)
	case shacl.PathInverse:
		return "^" + formatPath(p.Sub)
	case shacl.PathSequence:
		var parts []string
		for _, e := range p.Elements {
			parts = append(parts, formatPath(e))
		}
		return strings.Join(parts, " / ")
	case shacl.PathAlternative:
		var parts []string
		for _, e := range p.Elements {
			parts = append(parts, formatPath(e))
		}
		return "(" + strings.Join(parts, " | ") + ")"
	case shacl.PathZeroOrMore:
		return formatPath(p.Sub) + "*"
	case shacl.PathOneOrMore:
		return formatPath(p.Sub) + "+"
	case shacl.PathZeroOrOne:
		return formatPath(p.Sub) + "?"
	}
	return "unknown"
}

func CleanSparqlKeepNewlines(query string) string {
	// 1. Remove single-line comments (#...)
	// We use the (?m) flag for multi-line mode so ^ and $ match line boundaries
	reComments := regexp.MustCompile(`(?m)#.*$`)
	query = reComments.ReplaceAllString(query, "")

	// 2. Replace multiple horizontal spaces/tabs with a single space
	// \t = tab, \f = form feed, \r = carriage return (optional to keep)
	reHorizontalSpace := regexp.MustCompile(`[\t ]+`)
	query = reHorizontalSpace.ReplaceAllString(query, " ")

	// 3. Remove leading/trailing spaces on each individual line
	var lines []string
	for _, line := range strings.Split(query, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" { // This also removes empty lines created by deleted comments
			lines = append(lines, trimmed)
		}
	}

	return strings.Join(lines, "\n")
}

func MinifySparql(query string) string {
	// 1. Remove single-line comments (#...)
	// This matches a # that isn't inside a URI or string
	reComments := regexp.MustCompile(`(?m)^[ \t]*#.*$|#.*$`)
	query = reComments.ReplaceAllString(query, "")

	// 2. Replace all whitespace (tabs, newlines, multiple spaces) with a single space
	reWhitespace := regexp.MustCompile(`\s+`)
	query = reWhitespace.ReplaceAllString(query, " ")

	return strings.TrimSpace(query)
}
