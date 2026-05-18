package shaclimport

import (
	"cimgo/rdf/shacl"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// AnyToFloat coerces a SHACL payload value into a float64. SHACL counts and
// thresholds arrive as float64 (after JSON round-trip), int (in-memory ints
// from the simplifier), or string (literal values from the RDF parser).
func AnyToFloat(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case string:
		var f float64
		fmt.Sscanf(x, "%f", &f)
		return f
	}
	return 0
}

// ProcessFileToResults loads a SHACL Turtle file and converts it to a
// FileResults value. Call SimplifyFileResults to normalise the constraints
// before passing to a code generator or renderer.
func ProcessFileToResults(file string) (*FileResults, error) {
	g, err := shacl.LoadTurtleFile(file)
	if err != nil {
		return nil, err
	}

	shapes := shacl.ParseShapes(g)
	baseName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))

	results := &FileResults{
		FileName: baseName,
		Shapes:   []ShapeInfo{},
	}

	allWrapped := make(map[string]*shapeWrapper)
	for k := range shapes {
		allWrapped[k] = wrapShape(shapes[k])
	}

	isNested := identifyNestedShapes(shapes)

	shapeKeys := make([]string, 0, len(shapes))
	for k := range shapes {
		shapeKeys = append(shapeKeys, k)
	}
	sort.Strings(shapeKeys)

	for _, k := range shapeKeys {
		if isNested[k] {
			continue
		}
		sw := allWrapped[k]
		results.Shapes = append(results.Shapes, convertToShapeInfo(sw, allWrapped, g))
	}

	return results, nil
}

func SimplifyFileResults(fr *FileResults) *FileResults {
	simplified := &FileResults{
		FileName: fr.FileName,
		Shapes:   make([]ShapeInfo, 0, len(fr.Shapes)),
	}
	for _, s := range fr.Shapes {
		simplified.Shapes = append(simplified.Shapes, simplifyShape(s))
	}
	return simplified
}

// ---- internal types ----

type constraintWrapper struct {
	Type string
	Data shacl.Constraint
}

func (cw constraintWrapper) isSPARQL() bool {
	_, ok := cw.Data.(*shacl.SPARQLConstraint)
	return ok
}

func (cw constraintWrapper) isSHACL() bool {
	return !cw.isSPARQL()
}

type shapeWrapper struct {
	*shacl.Shape
	Constraints []constraintWrapper
	Properties  []*shapeWrapper
}

type sparqlInfo struct {
	Id    string
	Query string
}

// ---- pipeline helpers ----

func wrapShape(s *shacl.Shape) *shapeWrapper {
	if s == nil {
		return nil
	}
	sw := &shapeWrapper{Shape: s}
	for _, c := range s.Constraints {
		sw.Constraints = append(sw.Constraints, constraintWrapper{Type: c.ComponentIRI(), Data: c})
	}
	for _, ps := range s.Properties {
		sw.Properties = append(sw.Properties, wrapShape(ps))
	}
	return sw
}

func identifyNestedShapes(shapes map[string]*shacl.Shape) map[string]bool {
	isNested := make(map[string]bool)
	for _, s := range shapes {
		for _, ps := range s.Properties {
			isNested[ps.ID.String()] = true
		}
		for _, c := range s.Constraints {
			markNested(isNested, c)
		}
	}
	return isNested
}

func markNested(isNested map[string]bool, c shacl.Constraint) {
	switch con := c.(type) {
	case *shacl.AndConstraint:
		for _, sRef := range con.Shapes {
			isNested[sRef.String()] = true
		}
	case *shacl.OrConstraint:
		for _, sRef := range con.Shapes {
			isNested[sRef.String()] = true
		}
	case *shacl.XoneConstraint:
		for _, sRef := range con.Shapes {
			isNested[sRef.String()] = true
		}
	case *shacl.NotConstraint:
		isNested[con.ShapeRef.String()] = true
	case *shacl.NodeConstraint:
		isNested[con.ShapeRef.String()] = true
	case *shacl.QualifiedValueShapeConstraint:
		isNested[con.ShapeRef.String()] = true
	case *shacl.SeverityOverrideConstraint:
		markNested(isNested, con.Inner())
	}
}

// resolveSubClassOf follows rdfs:subClassOf in the graph until no further
// superclass is found, returning the most-general class available in the graph.
func resolveSubClassOf(g *shacl.Graph, class shacl.Term) shacl.Term {
	visited := map[string]bool{class.Value(): true}
	current := class
	for {
		supers := g.Objects(current, shacl.IRI(shacl.RDFSSubClassOf))
		if len(supers) == 0 {
			return current
		}
		next := supers[0]
		if visited[next.Value()] {
			return current
		}
		visited[next.Value()] = true
		current = next
	}
}

func convertToShapeInfo(sw *shapeWrapper, allWrapped map[string]*shapeWrapper, g *shacl.Graph) ShapeInfo {
	var targets []TargetInfo
	for _, t := range sw.Targets {
		val := simplifyTerm(t.Value)
		if t.Kind == shacl.TargetClass || t.Kind == shacl.TargetImplicitClass {
			resolved := resolveSubClassOf(g, t.Value)
			val = simplifyTerm(resolved)
		}
		targets = append(targets, TargetInfo{
			Kind:  t.Kind.String(),
			Value: val,
		})
	}

	var descriptions []string
	for _, d := range sw.Description {
		descriptions = append(descriptions, d.Value())
	}

	var names []string
	for _, n := range sw.Name {
		names = append(names, n.Value())
	}

	var messages []string
	for _, m := range sw.Messages {
		messages = append(messages, simplifyTerm(m))
	}

	var values *SparqlValuesInfo
	if sw.Values != nil {
		values = &SparqlValuesInfo{
			Select:   sw.Values.Select,
			Prefixes: sw.Values.Prefixes,
			Expr:     sw.Values.Expr,
		}
	}

	visited := map[string]bool{sw.ID.String(): true}
	info := ShapeInfo{
		ID:          simplifyTerm(sw.ID),
		Targets:     targets,
		Path:        formatPath(sw.Path),
		Name:        strings.Join(names, "; "),
		Description: strings.Join(descriptions, "\n"),
		Constraints: extractConstraints(sw, allWrapped, visited),
		Properties:  []ShapeInfo{},
		Values:      values,
		Severity:    simplifyTerm(sw.Severity),
		Messages:    messages,
	}

	for _, pw := range sw.Properties {
		info.Properties = append(info.Properties, convertToShapeInfo(pw, allWrapped, g))
	}

	return info
}

func extractConstraints(sw *shapeWrapper, allWrapped map[string]*shapeWrapper, visited map[string]bool) []ConstraintInfo {
	var constraints []ConstraintInfo

	defaultSeverity := simplifyTerm(sw.Severity)
	if defaultSeverity == "" {
		defaultSeverity = "sh:Violation"
	}

	var messages []string
	for _, m := range sw.Messages {
		messages = append(messages, simplifyTerm(m))
	}
	defaultMessage := strings.Join(messages, "; ")

	var descriptions []string
	for _, d := range sw.Description {
		descriptions = append(descriptions, d.Value())
	}
	defaultDescription := strings.Join(descriptions, "\n")

	var names []string
	for _, n := range sw.Name {
		names = append(names, n.Value())
	}
	defaultName := strings.Join(names, "; ")

	path := formatPath(sw.Path)

	for _, cw := range sw.Constraints {
		if !cw.isSHACL() {
			if sc, ok := cw.Data.(*shacl.SPARQLConstraint); ok {
				var msgs []string
				for _, m := range sc.Messages {
					msgs = append(msgs, simplifyTerm(m))
				}
				msg := strings.Join(msgs, "; ")
				if msg == "" {
					msg = defaultMessage
				}
				constraints = append(constraints, ConstraintInfo{
					Path:        path,
					Severity:    defaultSeverity,
					Message:     msg,
					Name:        defaultName,
					Description: defaultDescription,
					Component:   simplifyIRI(cw.Type),
					Payload: map[string]any{
						"Prefixes": sc.Prefixes,
						"Select":   sc.Select,
					},
				})
			}
			continue
		}

		severity := defaultSeverity
		displayData := cw.Data

		if soc, ok := cw.Data.(*shacl.SeverityOverrideConstraint); ok {
			displayData = soc.Inner()
			severity = simplifyTerm(soc.Severity)
		}

		data, _ := json.Marshal(displayData)
		var m map[string]any
		json.Unmarshal(data, &m)

		payload := make(map[string]any)
		for k, v := range m {
			payload[k] = formatValueWithResolution(v, allWrapped, visited)
		}

		constraints = append(constraints, ConstraintInfo{
			Path:        path,
			Severity:    severity,
			Message:     defaultMessage,
			Name:        defaultName,
			Description: defaultDescription,
			Component:   simplifyIRI(cw.Type),
			Payload:     payload,
		})
	}

	return constraints
}

func formatValueWithResolution(v any, allWrapped map[string]*shapeWrapper, visited map[string]bool) any {
	switch val := v.(type) {
	case map[string]any:
		if kind, ok := val["kind"].(string); ok {
			vVal, _ := val["value"].(string)
			var key string
			if kind == "IRI" {
				key = "<" + vVal + ">"
			} else if kind == "BlankNode" {
				key = "_:" + vVal
			}

			if key != "" {
				if resolved, ok := allWrapped[key]; ok {
					return resolveShapeConstraints(resolved, allWrapped, visited)
				}
				if kind == "IRI" {
					return simplifyIRI(vVal)
				}
			}
			return vVal
		}
	case []any:
		items := make([]any, 0, len(val))
		for _, item := range val {
			items = append(items, formatValueWithResolution(item, allWrapped, visited))
		}
		return items
	}
	return v
}

func resolveShapeConstraints(sw *shapeWrapper, allWrapped map[string]*shapeWrapper, visited map[string]bool) any {
	id := sw.ID.String()
	if visited[id] {
		return "ref:" + simplifyTerm(sw.ID)
	}
	visited[id] = true
	defer delete(visited, id)

	constraints := extractConstraints(sw, allWrapped, visited)
	if len(constraints) == 0 {
		return simplifyTerm(sw.ID)
	}
	return constraints
}

func simplifyShape(s ShapeInfo) ShapeInfo {
	s.Constraints = simplifyConstraints(s.Constraints)
	for i := range s.Properties {
		s.Properties[i] = simplifyShape(s.Properties[i])
	}
	return s
}

func simplifyConstraints(constraints []ConstraintInfo) []ConstraintInfo {
	hasAnyDatatype := false
	hasMinCount0 := false
	hasMinCount1 := false
	hasMaxCount1 := false
	for _, c := range constraints {
		switch c.Component {
		case "sh:DatatypeConstraintComponent":
			hasAnyDatatype = true
		case "sh:MinCountConstraintComponent":
			switch AnyToFloat(c.Payload["MinCount"]) {
			case 0:
				hasMinCount0 = true
			case 1:
				hasMinCount1 = true
			}
		case "sh:MaxCountConstraintComponent":
			if AnyToFloat(c.Payload["MaxCount"]) == 1 {
				hasMaxCount1 = true
			}
		}
	}
	mergeRequired := hasMinCount1 && hasMaxCount1
	mergeOptional := hasMinCount0 && hasMaxCount1

	var result []ConstraintInfo
	requiredAdded := false
	optionalAdded := false
	for _, c := range constraints {
		// Rule 1: any sh:datatype implies sh:Literal — drop redundant NodeKind=Literal
		if hasAnyDatatype && c.Component == "sh:NodeKindConstraintComponent" {
			if nk, ok := c.Payload["NodeKind"].(string); ok && nk == "sh:Literal" {
				continue
			}
		}
		// Rule 2: drop NodeKind=BlankNodeOrIRI and NodeKind=IRI — the Go type system
		// already enforces the IRI shape; the runtime check can never fail on
		// well-formed Go data.
		if c.Component == "sh:NodeKindConstraintComponent" {
			if nk, ok := c.Payload["NodeKind"].(string); ok {
				nk = strings.TrimPrefix(nk, "sh:")
				if nk == "BlankNodeOrIRI" || nk == "IRI" {
					continue
				}
			}
		}
		// Rule 3: minCount=0 is vacuously true — drop it
		if c.Component == "sh:MinCountConstraintComponent" && AnyToFloat(c.Payload["MinCount"]) == 0 {
			continue
		}
		// Drop sh:datatype for xsd types that map to native Go scalars.
		if c.Component == "sh:DatatypeConstraintComponent" {
			if dt, ok := c.Payload["Datatype"].(string); ok {
				switch strings.TrimPrefix(dt, "xsd:") {
				case "integer", "int", "long", "short", "byte",
					"nonNegativeInteger", "positiveInteger",
					"nonPositiveInteger", "negativeInteger",
					"unsignedInt", "unsignedLong", "unsignedShort", "unsignedByte",
					"float", "double", "decimal",
					"boolean",
					"string", "normalizedString", "token":
					continue
				}
			}
		}
		// sh:in with a single value is equivalent to sh:hasValue.
		if c.Component == "sh:InConstraintComponent" {
			if vals, ok := c.Payload["Values"].([]any); ok && len(vals) == 1 {
				result = append(result, ConstraintInfo{
					Path:      c.Path,
					Severity:  c.Severity,
					Message:   c.Message,
					Component: "sh:HasValueConstraintComponent",
					Payload:   map[string]any{"Value": vals[0]},
				})
				continue
			}
		}
		// Rule 4: minCount=0 + maxCount=1 → Optional
		if mergeOptional && c.Component == "sh:MaxCountConstraintComponent" && AnyToFloat(c.Payload["MaxCount"]) == 1 {
			if !optionalAdded {
				result = append(result, ConstraintInfo{
					Path:      c.Path,
					Severity:  c.Severity,
					Message:   c.Message,
					Component: "sh:OptionalConstraintComponent",
				})
				optionalAdded = true
			}
			continue
		}
		// Rule 5: minCount=1 + maxCount=1 → Required
		if mergeRequired {
			if c.Component == "sh:MinCountConstraintComponent" && AnyToFloat(c.Payload["MinCount"]) == 1 {
				if !requiredAdded {
					result = append(result, ConstraintInfo{
						Path:      c.Path,
						Severity:  c.Severity,
						Message:   c.Message,
						Component: "sh:RequiredConstraintComponent",
					})
					requiredAdded = true
				}
				continue
			}
			if c.Component == "sh:MaxCountConstraintComponent" && AnyToFloat(c.Payload["MaxCount"]) == 1 {
				continue
			}
		}
		result = append(result, c)
	}
	for i := range result {
		if result[i].Severity == "sh:Violation" {
			result[i].Severity = ""
		}
		for k, v := range result[i].Payload {
			result[i].Payload[k] = stripSHPrefix(v)
		}
	}
	return result
}

func stripSHPrefix(v any) any {
	switch val := v.(type) {
	case string:
		return strings.TrimPrefix(val, "sh:")
	case []any:
		for i := range val {
			val[i] = stripSHPrefix(val[i])
		}
		return val
	case map[string]any:
		for k, vv := range val {
			val[k] = stripSHPrefix(vv)
		}
		return val
	case []ConstraintInfo:
		for i := range val {
			if val[i].Severity == "sh:Violation" {
				val[i].Severity = ""
			}
			for k, vv := range val[i].Payload {
				val[i].Payload[k] = stripSHPrefix(vv)
			}
		}
		return val
	}
	return v
}

// ---- IRI / term formatting ----

var prefixes = map[string]string{
	"http://www.w3.org/1999/02/22-rdf-syntax-ns#":                                                "rdf",
	"http://www.w3.org/2000/01/rdf-schema#":                                                      "rdfs",
	"http://www.w3.org/2001/XMLSchema#":                                                          "xsd",
	"http://www.w3.org/ns/shacl#":                                                                "sh",
	"http://www.w3.org/2002/07/owl#":                                                             "owl",
	"http://iec.ch/TC57/CIM100#":                                                                 "cim",
	"http://iec.ch/TC57/ns/CIM/GeographicalLocation-EU/constraints/IEC61968-13/3.0#":             "gl",
	"http://iec.ch/TC57/ns/CIM/GeographicalLocation-EU/constraints/IEC61968-13/notSolved/3.0#":   "gl13n",
	"http://iec.ch/TC57/ns/CIM/DiagramLayout-EU/constraints/IEC61970-301/notSolved/3.0#":         "dl301n",
	"http://iec.ch/TC57/ns/CIM/DiagramLayout-EU/constraints/IEC61970-301/3.0#":                   "dl",
	"http://iec.ch/TC57/ns/CIM/Equipment-EU/constraints/IEC61970-301/notSolved/3.0#":             "eq301n",
	"http://iec.ch/TC57/ns/CIM/Equipment-EU/constraints/IEC61970-301/3.0#":                       "eq",
	"http://iec.ch/TC57/ns/CIM/EquipmentBoundary-EU/constraints/IEC61970-301/notSolved/3.0#":     "eqbd301n",
	"http://iec.ch/TC57/ns/CIM/EquipmentBoundary-EU/constraints/IEC61970-301/3.0#":               "eqbd",
	"http://iec.ch/TC57/ns/CIM/ShortCircuit-EU/constraints/IEC61970-301/notSolved/3.0#":          "sc301n",
	"http://iec.ch/TC57/ns/CIM/ShortCircuit-EU/constraints/IEC61970-301/3.0#":                    "sc",
	"http://iec.ch/TC57/ns/CIM/Operation-EU/constraints/IEC61970-301/3.0#":                       "op",
	"http://iec.ch/TC57/ns/CIM/Operation-EU/constraints/IEC61970-301/notSolved/3.0#":             "op301n",
	"http://iec.ch/TC57/ns/CIM/StateVariables-EU/constraints/IEC61970-301/3.0#":                  "sv",
	"http://iec.ch/TC57/ns/CIM/StateVariables-EU/constraints/IEC61970-301/notSolved/3.0#":        "sv301n",
	"http://iec.ch/TC57/ns/CIM/SteadyStateHypothesis-EU/constraints/IEC61970-301/notSolved/3.0#": "ssh301n",
	"http://iec.ch/TC57/ns/CIM/SteadyStateHypothesis-EU/constraints/IEC61970-301/3.0#":           "ssh",
	"http://iec.ch/TC57/ns/CIM/Topology-EU/constraints/IEC61970-301/notSolved/3.0#":              "tp301n",
	"http://iec.ch/TC57/ns/CIM/Topology-EU/constraints/IEC61970-301/3.0#":                        "tp",
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
	"http://iec.ch/TC57/ns/CIM/DiagramLayout-EU/Constraints#":                                    "dl",
	"http://iec.ch/TC57/ns/CIM/Dynamics/constraints/inverseAssociations/1.0#":                    "dyia",
	"http://iec.ch/TC57/ns/CIM/Dynamics-EU/Constraints#":                                         "dy302c",
	"http://iec.ch/TC57/ns/CIM/Equipment-EU/constraints/inverseAssociations/3.0#":                "eq301ia",
	"http://iec.ch/TC57/ns/CIM/Equipment-EU/constraints/IEC61970-600-2/3.0#":                     "eq600-2",
	"http://iec.ch/TC57/ns/CIM/CoreEquipment-EU/Constraints#":                                    "coreeqc",
	"http://iec.ch/TC57/ns/CIM/EquipmentBoundary-EU/constraints/inverseAssociations/3.0#":        "eqbd301ia",
	"http://iec.ch/TC57/ns/CIM/EquipmentBoundary-EU/Constraints#":                                "eqbd",
	"http://iec.ch/TC57/ns/CIM/GL-CrossProfileExplicit/constraints/IEC61968-13/3.0#":             "gl13cpe",
	"http://iec.ch/TC57/ns/CIM/GL-CrossProfileImplicit/constraints/IEC61968-13/3.0#":             "gl13cpi",
	"http://iec.ch/TC57/ns/CIM/GeographicalLocation-EU/constraints/inverseAssociations/3.0#":     "gl13ia",
	"http://iec.ch/TC57/ns/CIM/GeographicalLocation-EU/constraints/IEC61970-600-2/3.0#":          "gl600-2",
	"http://iec.ch/TC57/ns/CIM/GeographicalLocation-EU/Constraints#":                             "gl",
	"http://iec.ch/TC57/ns/CIM/IdentifiedObjectStringLength/constraints/3.0#":                    "iosl",
	"http://iec.ch/TC57/ns/CIM/OP-CrossProfileExplicit/constraints/IEC61970-452/3.0#":            "op452cpe",
	"http://iec.ch/TC57/ns/CIM/OP-CrossProfileImplicit/constraints/IEC61970-452/3.0#":            "op452cpi",
	"http://iec.ch/TC57/ns/CIM/Operation-EU/constraints/inverseAssociations/3.0#":                "op301ia",
	"http://iec.ch/TC57/ns/CIM/Operation-EU/Constraints#":                                        "op",
	"http://iec.ch/TC57/ns/CIM/ShortCircuit-EU/constraints/IEC61970-600-2/3.0#":                  "sc600-2",
	"http://iec.ch/TC57/ns/CIM/ShortCircuit-EU/Constraints#":                                     "sc",
	"http://iec.ch/TC57/ns/CIM/StateVariable-EU/constraints/inverseAssociations/3.0#":            "sv301ia",
	"http://iec.ch/TC57/ns/CIM/StateVariables-EU/Constraints#":                                   "sv",
	"http://iec.ch/TC57/ns/CIM/SteadyStateHypothesis-EU/Constraints#":                            "ssh",
	"http://iec.ch/TC57/ns/CIM/Topology-EU/constraints/inverseAssociations/3.0#":                 "tp301ia",
	"http://iec.ch/TC57/ns/CIM/Topology-EU/Constraints#":                                         "tp",
	"http://iec.ch/TC57/ns/CIM/ShortCircuit-EU/constraints/IEC61970-452/3.0#":                    "sc452c",
	"http://iec.ch/TC57/CIM100-European#":                                                        "cim100",
	"http://iec.ch/TC57/61970-552/DifferenceModel/1#":                                            "diff",
	"http://iec.ch/TC57/61970-552/ModelDescription/1#":                                           "mdc",
	"http://iec.ch/TC57/ns/CIM/IdentifiedObject/constraints/3.0#":                                "io",
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

// formatPath renders a SHACL property path as a list of sequence steps.
func formatPath(p *shacl.PropertyPath) []string {
	if p == nil {
		return nil
	}
	switch p.Kind {
	case shacl.PathPredicate:
		return []string{simplifyTerm(p.Pred)}
	case shacl.PathInverse:
		return []string{"^" + formatPathString(p.Sub)}
	case shacl.PathSequence:
		var parts []string
		for _, e := range p.Elements {
			parts = append(parts, formatPath(e)...)
		}
		return parts
	case shacl.PathAlternative:
		var parts []string
		for _, e := range p.Elements {
			parts = append(parts, formatPathString(e))
		}
		return []string{"(" + strings.Join(parts, " | ") + ")"}
	case shacl.PathZeroOrMore:
		return []string{formatPathString(p.Sub) + "*"}
	case shacl.PathOneOrMore:
		return []string{formatPathString(p.Sub) + "+"}
	case shacl.PathZeroOrOne:
		return []string{formatPathString(p.Sub) + "?"}
	}
	return []string{"unknown"}
}

func formatPathString(p *shacl.PropertyPath) string {
	return strings.Join(formatPath(p), " / ")
}

// ---- shape inspection ----

func hasContent(sw *shapeWrapper, filter func(constraintWrapper) bool) bool {
	for _, c := range sw.Constraints {
		if filter(c) {
			return true
		}
	}
	for _, p := range sw.Properties {
		if hasContent(p, filter) {
			return true
		}
	}
	return false
}

func filterConstraints(sw *shapeWrapper, filter func(constraintWrapper) bool) []constraintWrapper {
	var filtered []constraintWrapper
	for _, c := range sw.Constraints {
		if filter(c) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func collectSPARQLValues(sb *strings.Builder, sw *shapeWrapper, queries []sparqlInfo) []sparqlInfo {
	if sw.Values == nil {
		return queries
	}
	query := sw.Values.Prefixes + sw.Values.Select
	if sw.Values.Expr != "" {
		query = sw.Values.Prefixes + "SELECT (" + sw.Values.Expr + " AS ?value) WHERE { $this ?p ?o }"
	}
	queries = append(queries, sparqlInfo{Id: "Values", Query: query})
	sb.WriteString("**SPARQL Values:** [See below](#sparql-values)\n\n")
	return queries
}

// ---- SPARQL text utilities ----

func cleanSparqlKeepNewlines(query string) string {
	reComments := regexp.MustCompile(`(?m)#.*$`)
	query = reComments.ReplaceAllString(query, "")
	reHorizontalSpace := regexp.MustCompile(`[\t ]+`)
	query = reHorizontalSpace.ReplaceAllString(query, " ")
	var lines []string
	for _, line := range strings.Split(query, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return strings.Join(lines, "\n")
}

func minifySparql(query string) string {
	reComments := regexp.MustCompile(`(?m)^[ \t]*#.*$|#.*$`)
	query = reComments.ReplaceAllString(query, "")
	reWhitespace := regexp.MustCompile(`\s+`)
	query = reWhitespace.ReplaceAllString(query, " ")
	return strings.TrimSpace(query)
}
