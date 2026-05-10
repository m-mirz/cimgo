package shaclimport

import (
	"cimgo/rdf/shacl"
	"fmt"
	"regexp"
	"strings"
)

const DefaultSHACLPattern = "application-profiles-library/CGMES/CurrentRelease/SHACL/TTL/*.ttl"

var prefixes = map[string]string{
	"http://www.w3.org/1999/02/22-rdf-syntax-ns#":                                                "rdf",
	"http://www.w3.org/2000/01/rdf-schema#":                                                      "rdfs",
	"http://www.w3.org/2001/XMLSchema#":                                                          "xsd",
	"http://www.w3.org/ns/shacl#":                                                                "sh",
	"http://www.w3.org/2002/07/owl#":                                                             "owl",
	"http://iec.ch/TC57/CIM100#":                                                                 "cim",
	"http://iec.ch/TC57/ns/CIM/GeographicalLocation-EU/constraints/IEC61968-13/3.0#":             "gl13c",
	"http://iec.ch/TC57/ns/CIM/GeographicalLocation-EU/constraints/IEC61968-13/notSolved/3.0#":   "gl13n",
	"http://iec.ch/TC57/ns/CIM/DiagramLayout-EU/constraints/IEC61970-301/notSolved/3.0#":         "dl301n",
	"http://iec.ch/TC57/ns/CIM/DiagramLayout-EU/constraints/IEC61970-301/3.0#":                   "dl301c",
	"http://iec.ch/TC57/ns/CIM/Equipment-EU/constraints/IEC61970-301/notSolved/3.0#":             "eq301n",
	"http://iec.ch/TC57/ns/CIM/Equipment-EU/constraints/IEC61970-301/3.0#":                       "eq301c",
	"http://iec.ch/TC57/ns/CIM/EquipmentBoundary-EU/constraints/IEC61970-301/notSolved/3.0#":     "eqbd301n",
	"http://iec.ch/TC57/ns/CIM/EquipmentBoundary-EU/constraints/IEC61970-301/3.0#":               "eqbd301c",
	"http://iec.ch/TC57/ns/CIM/ShortCircuit-EU/constraints/IEC61970-301/notSolved/3.0#":          "sc301n",
	"http://iec.ch/TC57/ns/CIM/ShortCircuit-EU/constraints/IEC61970-301/3.0#":                    "sc301c",
	"http://iec.ch/TC57/ns/CIM/Operation-EU/constraints/IEC61970-301/3.0#":                       "op301c",
	"http://iec.ch/TC57/ns/CIM/Operation-EU/constraints/IEC61970-301/notSolved/3.0#":             "op301n",
	"http://iec.ch/TC57/ns/CIM/StateVariables-EU/constraints/IEC61970-301/3.0#":                  "sv301c",
	"http://iec.ch/TC57/ns/CIM/StateVariables-EU/constraints/IEC61970-301/notSolved/3.0#":        "sv301n",
	"http://iec.ch/TC57/ns/CIM/SteadyStateHypothesis-EU/constraints/IEC61970-301/notSolved/3.0#": "ssh301n",
	"http://iec.ch/TC57/ns/CIM/SteadyStateHypothesis-EU/constraints/IEC61970-301/3.0#":           "ssh301c",
	"http://iec.ch/TC57/ns/CIM/Topology-EU/constraints/IEC61970-301/notSolved/3.0#":              "tp301n",
	"http://iec.ch/TC57/ns/CIM/Topology-EU/constraints/IEC61970-301/3.0#":                        "tp301c",
	"http://iec.ch/TC57/ns/CIM/Dynamics/constraints/IEC61970-302/1.0#":                           "dy302c",
	"http://iec.ch/TC57/ns/CIM/Equipment-EU/constraints/IEC61970-452/notSolved/3.0#":             "eq452n",
	"http://iec.ch/TC57/ns/CIM/Equipment-EU/constraints/IEC61970-452/3.0#":                       "eq452c",
	"http://iec.ch/TC57/ns/CIM/Operation-EU/constraints/IEC61970-452/notSolved/3.0#":             "op452n",
	"http://iec.ch/TC57/ns/CIM/Operation-EU/constraints/IEC61970-452/3.0#":                       "op452c",
	"http://iec.ch/TC57/ns/CIM/SC-CrossProfile/constraints/IEC61970-452/3.0#":                    "sc452cp",
	"http://iec.ch/TC57/ns/CIM/SC-CrossProfile/constraints/IEC61970-452/notSolved/3.0#":          "sc452cpn",
	"http://iec.ch/TC57/ns/CIM/DL-CrossProfileExplicit/constraints/IEC61970-453/3.0#":            "dl453cpe",
	"http://iec.ch/TC57/ns/CIM/DL-CrossProfileExplicit/constraints/IEC61970-453/notSolved/3.0#":  "dl453cpen",
	"http://iec.ch/TC57/ns/CIM/DL-CrossProfileImplicit/constraints/IEC61970-453/3.0#":            "dl453cpi",
	"http://iec.ch/TC57/ns/CIM/DiagramLayout-EU/constraints/IEC61970-453/3.0#":                   "dl453c",
	"http://iec.ch/TC57/ns/CIM/DiagramLayout-EU/constraints/IEC61970-453/notSolved/3.0#":         "dl453n",
	"http://iec.ch/TC57/ns/CIM/SolvedMAS/constraints/IEC61970-456/3.0#":                          "mas456sol",
	"http://iec.ch/TC57/ns/CIM/SV-CrossProfileExplicit/constraints/IEC61970-456/3.0#":            "sv456cpe",
	"http://iec.ch/TC57/ns/CIM/SV-CrossProfileImplicit/constraints/IEC61970-456/3.0#":            "sv456cpi",
	"http://iec.ch/TC57/ns/CIM/StateVariable-EU/constraints/IEC61970-456/3.0#":                   "sv456c",
	"http://iec.ch/TC57/ns/CIM/StateVariable-EU/constraints/IEC61970-456/solved/3.0#":            "sv456sol",
	"http://iec.ch/TC57/ns/CIM/SteadyStateHypothesis-EU/constraints/IEC61970-456/notSolved/3.0#": "ssh456n",
	"http://iec.ch/TC57/ns/CIM/SteadyStateHypothesis-EU/constraints/IEC61970-456/3.0#":           "ssh456c",
	"http://iec.ch/TC57/ns/CIM/TP-CrossProfileExplicit/constraints/IEC61970-456/3.0#":            "tp456cpe",
	"http://iec.ch/TC57/ns/CIM/TP-CrossProfileImplicit/constraints/IEC61970-456/3.0#":            "tp456cpi",
	"http://iec.ch/TC57/ns/CIM/Topology-EU/constraints/IEC61970-456/notSolved/3.0#":              "tp456n",
	"http://iec.ch/TC57/ns/CIM/Topology-EU/constraints/IEC61970-456/3.0#":                        "tp456c",
	"http://iec.ch/TC57/ns/CIM/DY-CrossProfileExplicit/constraints/IEC61970-457/3.0#":            "dy457cpe",
	"http://iec.ch/TC57/ns/CIM/DY-CrossProfileImplicit/constraints/IEC61970-457/3.0#":            "dy457cpi",
	"http://iec.ch/TC57/ns/CIM/Dynamics/constraints/IEC61970-457/notSolved/1.0#":                 "dy457n",
	"http://iec.ch/TC57/ns/CIM/Dynamics/constraints/IEC61970-457/1.0#":                           "dy457c",
	"http://iec.ch/TC57/61970-552/ModelDescription/Constraints#":                                 "mdc",
	"http://iec.ch/TC57/ns/CIM/Equipment-EU/constraints/IEC61970-600/notSolved/3.0#":             "eq600n",
	"http://iec.ch/TC57/ns/CIM/Topology-EU/constraints/IEC61970-600/notSolved/3.0#":              "tp600n",
	"http://iec.ch/TC57/ns/CIM/All-EU/constraints/IEC61970-600-1/3.0#":                           "all600",
	"http://iec.ch/TC57/ns/CIM/SolvedMAS/constraints/IEC61970-600/3.0#":                          "mas600",
	"http://iec.ch/TC57/ns/CIM/Equipment-EU/constraints/IEC61970-600-1/3.0#":                     "eq600",
	"http://iec.ch/TC57/ns/CIM/prof10/constraints/IEC61970-600-1/3.0#":                           "prof10",
	"http://iec.ch/TC57/ns/CIM/SolvedMAS/constraints/IEC61970-600-2/3.0#":                        "mas600-2",
	"http://iec.ch/TC57/ns/CIM/DiagramLayout-EU/constraints/inverseAssociations/3.0#":            "dl301ia",
	"http://iec.ch/TC57/ns/CIM/DiagramLayout-EU/Constraints#":                                    "dl301c",
	"http://iec.ch/TC57/ns/CIM/Dynamics/constraints/inverseAssociations/1.0#":                    "dyia",
	"http://iec.ch/TC57/ns/CIM/Dynamics-EU/Constraints#":                                         "dy302c",
	"http://iec.ch/TC57/ns/CIM/Equipment-EU/constraints/inverseAssociations/3.0#":                "eq301ia",
	"http://iec.ch/TC57/ns/CIM/Equipment-EU/constraints/IEC61970-600-2/3.0#":                     "eq600-2",
	"http://iec.ch/TC57/ns/CIM/CoreEquipment-EU/Constraints#":                                    "coreeqc",
	"http://iec.ch/TC57/ns/CIM/EquipmentBoundary-EU/constraints/inverseAssociations/3.0#":        "eqbd301ia",
	"http://iec.ch/TC57/ns/CIM/EquipmentBoundary-EU/Constraints#":                                "eqbd301c",
	"http://iec.ch/TC57/ns/CIM/GL-CrossProfileExplicit/constraints/IEC61968-13/3.0#":             "gl13cpe",
	"http://iec.ch/TC57/ns/CIM/GL-CrossProfileImplicit/constraints/IEC61968-13/3.0#":             "gl13cpi",
	"http://iec.ch/TC57/ns/CIM/GeographicalLocation-EU/constraints/inverseAssociations/3.0#":     "gl13ia",
	"http://iec.ch/TC57/ns/CIM/GeographicalLocation-EU/constraints/IEC61970-600-2/3.0#":          "gl600-2",
	"http://iec.ch/TC57/ns/CIM/GeographicalLocation-EU/Constraints#":                             "gl13c",
	"http://iec.ch/TC57/ns/CIM/IdentifiedObjectStringLength/constraints/3.0#":                    "iosl",
	"http://iec.ch/TC57/ns/CIM/OP-CrossProfileExplicit/constraints/IEC61970-452/3.0#":            "op452cpe",
	"http://iec.ch/TC57/ns/CIM/OP-CrossProfileImplicit/constraints/IEC61970-452/3.0#":            "op452cpi",
	"http://iec.ch/TC57/ns/CIM/Operation-EU/constraints/inverseAssociations/3.0#":                "op301ia",
	"http://iec.ch/TC57/ns/CIM/Operation-EU/Constraints#":                                        "op301c",
	"http://iec.ch/TC57/ns/CIM/ShortCircuit-EU/constraints/IEC61970-600-2/3.0#":                  "sc600-2",
	"http://iec.ch/TC57/ns/CIM/ShortCircuit-EU/Constraints#":                                     "sc301c",
	"http://iec.ch/TC57/ns/CIM/StateVariable-EU/constraints/inverseAssociations/3.0#":            "sv301ia",
	"http://iec.ch/TC57/ns/CIM/StateVariables-EU/Constraints#":                                   "sv301c",
	"http://iec.ch/TC57/ns/CIM/SteadyStateHypothesis-EU/Constraints#":                            "ssh301c",
	"http://iec.ch/TC57/ns/CIM/Topology-EU/constraints/inverseAssociations/3.0#":                 "tp301ia",
	"http://iec.ch/TC57/ns/CIM/Topology-EU/Constraints#":                                         "tp301c",
	"http://iec.ch/TC57/ns/CIM/ShortCircuit-EU/constraints/IEC61970-452/3.0#":                    "sc452c",
	"http://iec.ch/TC57/CIM100-European#":                                                        "cim100",
	"http://iec.ch/TC57/61970-552/DifferenceModel/1#":                                            "diff",
	"http://iec.ch/TC57/61970-552/ModelDescription/1#":                                           "mdc",
	"http://iec.ch/TC57/ns/CIM/IdentifiedObject/constraints/3.0#":                                "io",
}

type FileStats struct {
	Name         string
	ShaclPath    string
	SparqlPath   string
	ShaclCounts  map[string]int
	SparqlCounts map[string]int
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

func WrapShape(s *shacl.Shape) *ShapeWrapper {
	if s == nil {
		return nil
	}
	sw := &ShapeWrapper{Shape: s}
	for _, c := range s.Constraints {
		sw.Constraints = append(sw.Constraints, ConstraintWrapper{Type: c.ComponentIRI(), Data: c})
	}
	for _, ps := range s.Properties {
		sw.Properties = append(sw.Properties, WrapShape(ps))
	}
	return sw
}

func SimplifyIRI(iri string) string {
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

func SimplifyTerm(t shacl.Term) string {
	if t.IsIRI() {
		return SimplifyIRI(t.Value())
	}
	return t.String()
}

func FormatValue(v any) string {
	switch val := v.(type) {
	case map[string]any:
		if kind, ok := val["kind"].(string); ok {
			vIRI, _ := val["value"].(string)
			if kind == "IRI" {
				return SimplifyIRI(vIRI)
			}
			return vIRI
		}
	case []any:
		var items []string
		for _, item := range val {
			items = append(items, FormatValue(item))
		}
		return "[" + strings.Join(items, ", ") + "]"
	}
	return fmt.Sprint(v)
}

// FormatPath renders a SHACL property path as a list of sequence steps.
// A simple predicate or inverse is a single-element list; a sh:sequencePath
// becomes one element per step. Alternatives and modifiers (*, +, ?) are
// collapsed back into a single element so the list shape mirrors the
// sequence structure only.
func FormatPath(p *shacl.PropertyPath) []string {
	if p == nil {
		return nil
	}
	switch p.Kind {
	case shacl.PathPredicate:
		return []string{SimplifyTerm(p.Pred)}
	case shacl.PathInverse:
		return []string{"^" + FormatPathString(p.Sub)}
	case shacl.PathSequence:
		var parts []string
		for _, e := range p.Elements {
			parts = append(parts, FormatPath(e)...)
		}
		return parts
	case shacl.PathAlternative:
		var parts []string
		for _, e := range p.Elements {
			parts = append(parts, FormatPathString(e))
		}
		return []string{"(" + strings.Join(parts, " | ") + ")"}
	case shacl.PathZeroOrMore:
		return []string{FormatPathString(p.Sub) + "*"}
	case shacl.PathOneOrMore:
		return []string{FormatPathString(p.Sub) + "+"}
	case shacl.PathZeroOrOne:
		return []string{FormatPathString(p.Sub) + "?"}
	}
	return []string{"unknown"}
}

// FormatPathString joins the FormatPath segments with " / " for callers that
// need a single-string rendering (e.g. attribute names, markdown output).
func FormatPathString(p *shacl.PropertyPath) string {
	return strings.Join(FormatPath(p), " / ")
}

// hasContent checks if the shape or any of its nested properties contain constraints that match the filter
func HasContent(sw *ShapeWrapper, filter func(ConstraintWrapper) bool) bool {
	for _, c := range sw.Constraints {
		if filter(c) {
			return true
		}
	}
	for _, p := range sw.Properties {
		if HasContent(p, filter) {
			return true
		}
	}
	return false
}

type SparqlInfo struct {
	Id    string
	Query string
}

func CollectSPARQLValues(sb *strings.Builder, sw *ShapeWrapper, queries []SparqlInfo) []SparqlInfo {
	if sw.Values == nil {
		return queries
	}
	query := sw.Values.Prefixes + sw.Values.Select
	if sw.Values.Expr != "" {
		query = sw.Values.Prefixes + "SELECT (" + sw.Values.Expr + " AS ?value) WHERE { $this ?p ?o }"
	}
	queries = append(queries, SparqlInfo{Id: "Values", Query: query})
	sb.WriteString("**SPARQL Values:** [See below](#sparql-values)\n\n")
	return queries
}

func FilterConstraints(sw *ShapeWrapper, filter func(ConstraintWrapper) bool) []ConstraintWrapper {
	var filtered []ConstraintWrapper
	for _, c := range sw.Constraints {
		if filter(c) {
			filtered = append(filtered, c)
		}
	}
	return filtered
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
