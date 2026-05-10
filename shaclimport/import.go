package shaclimport

import (
	"cimgo/rdf/shacl"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// ConstraintInfo, TargetInfo, ShapeInfo, FileResults are the simplified
// SHACL representation produced by ProcessFileToResults + SimplifyFileResults
// and consumed by cmd/shaclgen.
type ConstraintInfo struct {
	Path        []string       `json:"path"`
	Severity    string         `json:"severity,omitempty"`
	Message     string         `json:"message,omitempty"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Component   string         `json:"component"`
	Payload     map[string]any `json:"payload"`
}

func (c ConstraintInfo) IsSPARQL() bool {
	return c.Component == "sh:SPARQLConstraintComponent"
}

func (c ConstraintInfo) IsSHACL() bool {
	return !c.IsSPARQL()
}

type TargetInfo struct {
	Kind  string `json:"kind"`  // e.g. "targetClass", "targetNode"
	Value string `json:"value"` // Simplified IRI
}

type SparqlValuesInfo struct {
	Select   string `json:"select"`
	Prefixes string `json:"prefixes"`
	Expr     string `json:"expr,omitempty"`
}

type ShapeInfo struct {
	ID          string            `json:"id"`
	Targets     []TargetInfo      `json:"targets,omitempty"`
	Path        []string          `json:"path,omitempty"`
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Constraints []ConstraintInfo  `json:"constraints,omitempty"`
	Properties  []ShapeInfo       `json:"properties,omitempty"`
	Values      *SparqlValuesInfo `json:"values,omitempty"`
	Severity    string            `json:"severity,omitempty"`
	Messages    []string          `json:"messages,omitempty"`
}

type FileResults struct {
	FileName string      `json:"file_name"`
	Shapes   []ShapeInfo `json:"shapes"`
}

// anyToFloat coerces a SHACL payload value into a float64. SHACL counts and
// thresholds arrive as float64 (after JSON round-trip), int (in-memory ints
// from the simplifier), or string (literal values from the RDF parser).
func anyToFloat(v any) float64 {
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

func simplifyConstraints(constraints []ConstraintInfo) []ConstraintInfo {
	// Pre-scan
	hasAnyDatatype := false
	hasMinCount0 := false
	hasMinCount1 := false
	hasMaxCount1 := false
	for _, c := range constraints {
		switch c.Component {
		case "sh:DatatypeConstraintComponent":
			hasAnyDatatype = true
		case "sh:MinCountConstraintComponent":
			switch anyToFloat(c.Payload["MinCount"]) {
			case 0:
				hasMinCount0 = true
			case 1:
				hasMinCount1 = true
			}
		case "sh:MaxCountConstraintComponent":
			if anyToFloat(c.Payload["MaxCount"]) == 1 {
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
		// Rule 2: drop NodeKind=BlankNodeOrIRI (the SHACL non-literal default), and
		// drop NodeKind=IRI unconditionally — every IRI-typed property in the CIM
		// schema is generated as a Go reference field (*struct{ MRID string }), so
		// the type system already enforces the IRI shape and the runtime check
		// can never fail on well-formed Go data.
		if c.Component == "sh:NodeKindConstraintComponent" {
			if nk, ok := c.Payload["NodeKind"].(string); ok {
				nk = strings.TrimPrefix(nk, "sh:")
				if nk == "BlankNodeOrIRI" || nk == "IRI" {
					continue
				}
			}
		}
		// Rule 3: minCount=0 is vacuously true — drop it
		if c.Component == "sh:MinCountConstraintComponent" && anyToFloat(c.Payload["MinCount"]) == 0 {
			continue
		}
		// Drop sh:datatype for xsd types that map to native Go scalars: the
		// generated struct field is already typed (int, float64, bool, string),
		// so the XML decoder rejects malformed values before validation runs.
		// Non-native types (dateTime, gMonthDay, anyURI map to Go string and
		// have no format enforcement) are left in for future format checking.
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
		// Rule 4: minCount=0 + maxCount=1 → Optional (maxCount=1 entry is replaced;
		// the minCount=0 entry is already dropped by Rule 3 above)
		if mergeOptional && c.Component == "sh:MaxCountConstraintComponent" && anyToFloat(c.Payload["MaxCount"]) == 1 {
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
		// Existing rule: minCount=1 + maxCount=1 → Required
		if mergeRequired {
			if c.Component == "sh:MinCountConstraintComponent" && anyToFloat(c.Payload["MinCount"]) == 1 {
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
			if c.Component == "sh:MaxCountConstraintComponent" && anyToFloat(c.Payload["MaxCount"]) == 1 {
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

// stripSHPrefix recursively walks a payload value and removes the leading "sh:"
// prefix from string values (e.g. "sh:IRI" → "IRI"). Severity fields on any
// nested ConstraintInfo are also normalized to drop the default "sh:Violation".
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

func simplifyShape(s ShapeInfo) ShapeInfo {
	s.Constraints = simplifyConstraints(s.Constraints)
	for i := range s.Properties {
		s.Properties[i] = simplifyShape(s.Properties[i])
	}
	return s
}

// resolveSubClassOf follows rdfs:subClassOf in the graph until no further
// superclass is found, returning the most-general class available in the graph.
// This maps profile-specific subclasses (e.g. prof10:FullModel-EQ) to their
// canonical CIM base class (e.g. mdc:FullModel).
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

	allWrapped := make(map[string]*ShapeWrapper)
	for k := range shapes {
		allWrapped[k] = WrapShape(shapes[k])
	}

	isNested := IdentifyNestedShapes(shapes)

	// Sort shape IDs for deterministic output
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
		results.Shapes = append(results.Shapes, ConvertToShapeInfo(sw, allWrapped, g))
	}

	return results, nil
}

func IdentifyNestedShapes(shapes map[string]*shacl.Shape) map[string]bool {
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

func ConvertToShapeInfo(sw *ShapeWrapper, allWrapped map[string]*ShapeWrapper, g *shacl.Graph) ShapeInfo {
	var targets []TargetInfo
	for _, t := range sw.Targets {
		val := SimplifyTerm(t.Value)
		if t.Kind == shacl.TargetClass || t.Kind == shacl.TargetImplicitClass {
			resolved := resolveSubClassOf(g, t.Value)
			val = SimplifyTerm(resolved)
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
		messages = append(messages, SimplifyTerm(m))
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
		ID:          SimplifyTerm(sw.ID),
		Targets:     targets,
		Path:        FormatPath(sw.Path),
		Name:        strings.Join(names, "; "),
		Description: strings.Join(descriptions, "\n"),
		Constraints: ExtractConstraints(sw, allWrapped, visited),
		Properties:  []ShapeInfo{},
		Values:      values,
		Severity:    SimplifyTerm(sw.Severity),
		Messages:    messages,
	}

	for _, pw := range sw.Properties {
		info.Properties = append(info.Properties, ConvertToShapeInfo(pw, allWrapped, g))
	}

	return info
}

func ExtractConstraints(sw *ShapeWrapper, allWrapped map[string]*ShapeWrapper, visited map[string]bool) []ConstraintInfo {
	var constraints []ConstraintInfo

	defaultSeverity := SimplifyTerm(sw.Severity)
	if defaultSeverity == "" {
		defaultSeverity = "sh:Violation"
	}

	var messages []string
	for _, m := range sw.Messages {
		messages = append(messages, SimplifyTerm(m))
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

	path := FormatPath(sw.Path)

	for _, cw := range sw.Constraints {
		if !cw.IsSHACL() {
			continue
		}

		severity := defaultSeverity
		displayData := cw.Data

		if soc, ok := cw.Data.(*shacl.SeverityOverrideConstraint); ok {
			displayData = soc.Inner()
			severity = SimplifyTerm(soc.Severity)
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
			Component:   SimplifyIRI(cw.Type),
			Payload:     payload,
		})
	}

	return constraints
}

func formatValueWithResolution(v any, allWrapped map[string]*ShapeWrapper, visited map[string]bool) any {
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
					return SimplifyIRI(vVal)
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

func resolveShapeConstraints(sw *ShapeWrapper, allWrapped map[string]*ShapeWrapper, visited map[string]bool) any {
	id := sw.ID.String()
	if visited[id] {
		return "ref:" + SimplifyTerm(sw.ID)
	}
	visited[id] = true
	defer delete(visited, id)

	constraints := ExtractConstraints(sw, allWrapped, visited)
	if len(constraints) == 0 {
		return SimplifyTerm(sw.ID)
	}
	return constraints
}
