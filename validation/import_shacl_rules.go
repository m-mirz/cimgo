package validation

import (
	"cimgo/rdf/shacl"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	SHACL_SCHEMA = "application-profiles-library/CGMES/CurrentRelease/SHACL/TTL/*.ttl"
)

func importSHACLRules() {
	outputDir := "pages/docs/struct"

	shaclFiles, err := filepath.Glob(SHACL_SCHEMA)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error globbing files: %v\n", err)
		return
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		return
	}

	for _, file := range shaclFiles {
		results, err := processFileToResults(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", file, err)
			continue
		}

		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling %s: %v\n", file, err)
			continue
		}

		baseName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
		outFile := filepath.Join(outputDir, baseName+".json")
		if err := os.WriteFile(outFile, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", outFile, err)
			continue
		}
		fmt.Printf("Exported struct to %s\n", outFile)
	}
}

func GetSHACLRules(shaclPattern string) (map[string]ClassInfo, error) {
	rules := make(map[string]ClassInfo)
	shaclFiles, err := filepath.Glob(shaclPattern)
	if err != nil {
		return nil, err
	}

	for _, file := range shaclFiles {
		results, err := processFileToResults(file)
		if err != nil {
			return nil, err
		}

		for _, cls := range results.Classes {
			if existing, ok := rules[cls.Name]; ok {
				existing.Constraints = append(existing.Constraints, cls.Constraints...)
				existing.Attributes = append(existing.Attributes, cls.Attributes...)
				rules[cls.Name] = existing
			} else {
				rules[cls.Name] = cls
			}
		}
	}
	return rules, nil
}

func processFileToResults(file string) (*FileResults, error) {
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

		// Determine target classes/nodes
		var explicitTargets []string
		var implicitTargets []string
		for _, t := range sw.Targets {
			termVal := SimplifyTerm(t.Value)
			if t.Kind == shacl.TargetClass || t.Kind == shacl.TargetNode {
				explicitTargets = append(explicitTargets, termVal)
			} else if t.Kind == shacl.TargetImplicitClass {
				implicitTargets = append(implicitTargets, termVal)
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
	name := FormatPath(pw.Path)

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
		var details []string
		for k, v := range m {
			resolvedVal, strVal := formatValueWithResolution(v, allWrapped, visited)
			payload[k] = resolvedVal
			details = append(details, fmt.Sprintf("%s: %s", k, strVal))
		}

		constraints = append(constraints, ConstraintInfo{
			Path:      path,
			Severity:  severity,
			Message:   defaultMessage,
			Component: SimplifyIRI(cw.Type),
			Payload:   payload,
			Details:   strings.Join(details, ", "),
		})
	}

	return constraints
}

func formatValueWithResolution(v any, allWrapped map[string]*ShapeWrapper, visited map[string]bool) (any, string) {
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
					return SimplifyIRI(vVal), SimplifyIRI(vVal)
				}
			}
			return vVal, vVal
		}
	case []any:
		var items []any
		var strs []string
		for _, item := range val {
			rv, sv := formatValueWithResolution(item, allWrapped, visited)
			items = append(items, rv)
			strs = append(strs, sv)
		}
		return items, "[" + strings.Join(strs, ", ") + "]"
	}
	s := fmt.Sprint(v)
	return v, s
}

func resolveShapeConstraints(sw *ShapeWrapper, allWrapped map[string]*ShapeWrapper, visited map[string]bool) (any, string) {
	id := sw.ID.String()
	if visited[id] {
		ref := "ref:" + SimplifyTerm(sw.ID)
		return ref, ref
	}
	visited[id] = true
	defer delete(visited, id)

	constraints := ExtractConstraints(sw, allWrapped, visited)
	if len(constraints) == 0 {
		s := SimplifyTerm(sw.ID)
		return s, s
	}

	var parts []string
	for _, c := range constraints {
		details := c.Details
		if c.Path != "" {
			if details != "" {
				details = "path: " + c.Path + ", " + details
			} else {
				details = "path: " + c.Path
			}
		}
		parts = append(parts, fmt.Sprintf("%s(%s)", c.Component, details))
	}
	return constraints, "{" + strings.Join(parts, ", ") + "}"
}
