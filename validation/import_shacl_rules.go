package validation

import (
	"cimgo/rdf/shacl"
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"
)

func simplifyConstraints(constraints []ConstraintInfo) []ConstraintInfo {
	// Pre-scan
	hasAnyDatatype := false
	hasMinCount0 := false
	hasMinCount1 := false
	hasMaxCount1 := false
	for _, c := range constraints {
		switch c.Component {
		case "sh.DatatypeConstraintComponent":
			hasAnyDatatype = true
		case "sh.MinCountConstraintComponent":
			switch anyToFloat(c.Payload["MinCount"]) {
			case 0:
				hasMinCount0 = true
			case 1:
				hasMinCount1 = true
			}
		case "sh.MaxCountConstraintComponent":
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
		if hasAnyDatatype && c.Component == "sh.NodeKindConstraintComponent" {
			if nk, ok := c.Payload["NodeKind"].(string); ok && nk == "sh.Literal" {
				continue
			}
		}
		// Rule 2: drop NodeKind=BlankNodeOrIRI (the SHACL non-literal default), and
		// drop NodeKind=IRI unconditionally — every IRI-typed property in the CIM
		// schema is generated as a Go reference field (*struct{ MRID string }), so
		// the type system already enforces the IRI shape and the runtime check
		// can never fail on well-formed Go data.
		if c.Component == "sh.NodeKindConstraintComponent" {
			if nk, ok := c.Payload["NodeKind"].(string); ok {
				nk = strings.TrimPrefix(nk, "sh.")
				if nk == "BlankNodeOrIRI" || nk == "IRI" {
					continue
				}
			}
		}
		// Rule 3: minCount=0 is vacuously true — drop it
		if c.Component == "sh.MinCountConstraintComponent" && anyToFloat(c.Payload["MinCount"]) == 0 {
			continue
		}
		// Drop sh:datatype for xsd types that map to native Go scalars: the
		// generated struct field is already typed (int, float64, bool, string),
		// so the XML decoder rejects malformed values before validation runs.
		// Non-native types (dateTime, gMonthDay, anyURI map to Go string and
		// have no format enforcement) are left in for future format checking.
		if c.Component == "sh.DatatypeConstraintComponent" {
			if dt, ok := c.Payload["Datatype"].(string); ok {
				switch strings.TrimPrefix(dt, "xsd.") {
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
		if c.Component == "sh.InConstraintComponent" {
			if vals, ok := c.Payload["Values"].([]any); ok && len(vals) == 1 {
				result = append(result, ConstraintInfo{
					Path:      c.Path,
					Severity:  c.Severity,
					Message:   c.Message,
					Component: "sh.HasValueConstraintComponent",
					Payload:   map[string]any{"Value": vals[0]},
				})
				continue
			}
		}
		// Rule 4: minCount=0 + maxCount=1 → Optional (maxCount=1 entry is replaced;
		// the minCount=0 entry is already dropped by Rule 3 above)
		if mergeOptional && c.Component == "sh.MaxCountConstraintComponent" && anyToFloat(c.Payload["MaxCount"]) == 1 {
			if !optionalAdded {
				result = append(result, ConstraintInfo{
					Path:      c.Path,
					Severity:  c.Severity,
					Message:   c.Message,
					Component: "sh.OptionalConstraintComponent",
				})
				optionalAdded = true
			}
			continue
		}
		// Existing rule: minCount=1 + maxCount=1 → Required
		if mergeRequired {
			if c.Component == "sh.MinCountConstraintComponent" && anyToFloat(c.Payload["MinCount"]) == 1 {
				if !requiredAdded {
					result = append(result, ConstraintInfo{
						Path:      c.Path,
						Severity:  c.Severity,
						Message:   c.Message,
						Component: "sh.RequiredConstraintComponent",
					})
					requiredAdded = true
				}
				continue
			}
			if c.Component == "sh.MaxCountConstraintComponent" && anyToFloat(c.Payload["MaxCount"]) == 1 {
				continue
			}
		}
		result = append(result, c)
	}
	for i := range result {
		if result[i].Severity == "sh.Violation" {
			result[i].Severity = ""
		}
		for k, v := range result[i].Payload {
			result[i].Payload[k] = stripSHPrefix(v)
		}
	}
	return result
}

// stripSHPrefix recursively walks a payload value and removes the leading "sh."
// prefix from string values (e.g. "sh.IRI" → "IRI"). Severity fields on any
// nested ConstraintInfo are also normalized to drop the default "sh.Violation".
func stripSHPrefix(v any) any {
	switch val := v.(type) {
	case string:
		return strings.TrimPrefix(val, "sh.")
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
			if val[i].Severity == "sh.Violation" {
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
		Classes:  make([]ClassInfo, 0, len(fr.Classes)),
	}
	for _, cls := range fr.Classes {
		sCls := ClassInfo{
			Name:        cls.Name,
			Constraints: simplifyConstraints(cls.Constraints),
			Attributes:  make([]AttributeInfo, 0, len(cls.Attributes)),
		}
		for _, attr := range cls.Attributes {
			sCls.Attributes = append(sCls.Attributes, AttributeInfo{
				Name:        attr.Name,
				Description: attr.Description,
				Constraints: simplifyConstraints(attr.Constraints),
			})
		}
		simplified.Classes = append(simplified.Classes, sCls)
	}
	return simplified
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
		Classes:  []ClassInfo{},
	}

	allWrapped := make(map[string]*ShapeWrapper)
	for k := range shapes {
		allWrapped[k] = WrapShape(shapes[k])
	}

	isNested := IdentifyNestedShapes(shapes)

	classMap := make(map[string]*ClassInfo)

	for k := range shapes {
		if isNested[k] {
			continue
		}

		sw := allWrapped[k]

		// Determine target classes/nodes.
		// For class targets (explicit and implicit) follow rdfs:subClassOf to
		// the base CIM class so that profile-specific subclasses like
		// prof10:FullModel-EQ resolve to their canonical class (mdc:FullModel).
		var explicitTargets []string
		var implicitTargets []string
		for _, t := range sw.Targets {
			if t.Kind == shacl.TargetNode {
				explicitTargets = append(explicitTargets, SimplifyTerm(t.Value))
			} else if t.Kind == shacl.TargetClass {
				resolved := resolveSubClassOf(g, t.Value)
				explicitTargets = append(explicitTargets, SimplifyTerm(resolved))
			} else if t.Kind == shacl.TargetImplicitClass {
				resolved := resolveSubClassOf(g, t.Value)
				implicitTargets = append(implicitTargets, SimplifyTerm(resolved))
			}
		}

		var targets []string
		if len(explicitTargets) > 0 {
			targets = explicitTargets
		} else {
			targets = implicitTargets
		}

		for _, className := range targets {
			if strings.HasPrefix(className, "_:") {
				continue
			}
			if info, ok := classMap[className]; ok {
				mergeIntoClassInfo(info, sw, allWrapped)
			} else {
				classInfo := ConvertToClassInfo(sw, allWrapped, className)
				classMap[className] = &classInfo
			}
		}
	}

	// Sort class names for deterministic output
	var classNames []string
	for name := range classMap {
		classNames = append(classNames, name)
	}
	sort.Strings(classNames)

	for _, name := range classNames {
		results.Classes = append(results.Classes, *classMap[name])
	}

	return results, nil
}

func mergeIntoClassInfo(info *ClassInfo, sw *ShapeWrapper, allWrapped map[string]*ShapeWrapper) {
	// Add class-level constraints
	visited := map[string]bool{sw.ID.String(): true}
	info.Constraints = append(info.Constraints, ExtractConstraints(sw, allWrapped, visited)...)

	// Add attributes
	for _, pw := range sw.Properties {
		attrInfo := ConvertToAttributeInfo(pw, allWrapped)
		if len(attrInfo.Constraints) > 0 {
			// Check if attribute already exists to merge constraints
			found := false
			for i := range info.Attributes {
				if info.Attributes[i].Name == attrInfo.Name {
					info.Attributes[i].Constraints = append(info.Attributes[i].Constraints, attrInfo.Constraints...)
					found = true
					break
				}
			}
			if !found {
				info.Attributes = append(info.Attributes, attrInfo)
			}
		}
	}
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

func ConvertToClassInfo(sw *ShapeWrapper, allWrapped map[string]*ShapeWrapper, name string) ClassInfo {
	visited := map[string]bool{sw.ID.String(): true}
	classInfo := ClassInfo{
		Name:        name,
		Constraints: ExtractConstraints(sw, allWrapped, visited),
		Attributes:  []AttributeInfo{},
	}

	for _, pw := range sw.Properties {
		attrInfo := ConvertToAttributeInfo(pw, allWrapped)
		if len(attrInfo.Constraints) > 0 {
			classInfo.Attributes = append(classInfo.Attributes, attrInfo)
		}
	}

	return classInfo
}

func ConvertToAttributeInfo(pw *ShapeWrapper, allWrapped map[string]*ShapeWrapper) AttributeInfo {
	name := FormatPathString(pw.Path)

	var descriptions []string
	for _, d := range pw.Description {
		descriptions = append(descriptions, d.Value())
	}

	visited := map[string]bool{pw.ID.String(): true}
	attrInfo := AttributeInfo{
		Name:        name,
		Description: strings.Join(descriptions, "\n"),
		Constraints: ExtractConstraints(pw, allWrapped, visited),
	}

	return attrInfo
}

func ExtractConstraints(sw *ShapeWrapper, allWrapped map[string]*ShapeWrapper, visited map[string]bool) []ConstraintInfo {
	var constraints []ConstraintInfo

	defaultSeverity := SimplifyTerm(sw.Severity)
	if defaultSeverity == "" {
		defaultSeverity = "sh.Violation"
	}

	var messages []string
	for _, m := range sw.Messages {
		messages = append(messages, SimplifyTerm(m))
	}
	defaultMessage := strings.Join(messages, "; ")

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
			Path:      path,
			Severity:  severity,
			Message:   defaultMessage,
			Component: SimplifyIRI(cw.Type),
			Payload:   payload,
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
