package main

import (
	"cimgo/rdf/shacl"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	SHACL_SCHEMA = "application-profiles-library/CGMES/CurrentRelease/SHACL/TTL/*.ttl"
)

type ConstraintInfo struct {
	Path      string `json:"path,omitempty"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	Component string `json:"component"`
	Details   string `json:"details"`
}

type AttributeInfo struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Constraints []ConstraintInfo `json:"constraints"`
}

type ClassInfo struct {
	Name        string           `json:"name"`
	Constraints []ConstraintInfo `json:"constraints"`
	Attributes  []AttributeInfo  `json:"attributes"`
}

type FileResults struct {
	FileName string      `json:"file_name"`
	Classes  []ClassInfo `json:"classes"`
}

func main() {
	shaclPattern := flag.String("shacl", SHACL_SCHEMA, "glob pattern for shacl files")
	outputDir := flag.String("out", "pages/docs/struct", "output directory for generated files")
	flag.Parse()

	shaclFiles, err := filepath.Glob(*shaclPattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error globbing files: %v\n", err)
		return
	}

	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		return
	}

	for _, file := range shaclFiles {
		if err := processFile(file, *outputDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", file, err)
		}
	}
}

func processFile(file string, outputDir string) error {
	g, err := shacl.LoadTurtleFile(file)
	if err != nil {
		return err
	}

	shapes := shacl.ParseShapes(g)
	baseName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))

	results := FileResults{
		FileName: baseName,
		Classes:  []ClassInfo{},
	}

	allWrapped := make(map[string]*ShapeWrapper)
	for k := range shapes {
		allWrapped[k] = wrapShape(shapes[k])
	}

	isNested := identifyNestedShapes(shapes)

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
			termVal := simplifyTerm(t.Value)
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
				classInfo := convertToClassInfo(sw, allWrapped, className)
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

	if len(results.Classes) == 0 {
		return nil
	}

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}

	outFile := filepath.Join(outputDir, baseName+".json")
	if err := os.WriteFile(outFile, data, 0644); err != nil {
		return err
	}
	fmt.Printf("Exported struct to %s\n", outFile)
	return nil
}

func mergeIntoClassInfo(info *ClassInfo, sw *ShapeWrapper, allWrapped map[string]*ShapeWrapper) {
	// Add class-level constraints
	info.Constraints = append(info.Constraints, extractConstraints(sw, allWrapped, make(map[string]bool))...)

	// Add attributes
	for _, pw := range sw.Properties {
		attrInfo := convertToAttributeInfo(pw, allWrapped)
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

func convertToClassInfo(sw *ShapeWrapper, allWrapped map[string]*ShapeWrapper, name string) ClassInfo {
	classInfo := ClassInfo{
		Name:        name,
		Constraints: extractConstraints(sw, allWrapped, make(map[string]bool)),
		Attributes:  []AttributeInfo{},
	}

	for _, pw := range sw.Properties {
		attrInfo := convertToAttributeInfo(pw, allWrapped)
		if len(attrInfo.Constraints) > 0 {
			classInfo.Attributes = append(classInfo.Attributes, attrInfo)
		}
	}

	return classInfo
}

func convertToAttributeInfo(pw *ShapeWrapper, allWrapped map[string]*ShapeWrapper) AttributeInfo {
	name := formatPath(pw.Path)

	var descriptions []string
	for _, d := range pw.Description {
		descriptions = append(descriptions, d.Value())
	}

	attrInfo := AttributeInfo{
		Name:        name,
		Description: strings.Join(descriptions, "\n"),
		Constraints: extractConstraints(pw, allWrapped, make(map[string]bool)),
	}

	return attrInfo
}

func extractConstraints(sw *ShapeWrapper, allWrapped map[string]*ShapeWrapper, visited map[string]bool) []ConstraintInfo {
	var constraints []ConstraintInfo

	defaultSeverity := simplifyTerm(sw.Severity)
	if defaultSeverity == "" {
		defaultSeverity = "sh.Violation"
	}

	var messages []string
	for _, m := range sw.Messages {
		messages = append(messages, simplifyTerm(m))
	}
	defaultMessage := strings.Join(messages, "; ")

	path := formatPath(sw.Path)

	for _, cw := range sw.Constraints {
		if !cw.IsSHACL() {
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

		var details []string
		for k, v := range m {
			details = append(details, fmt.Sprintf("%s: %s", k, formatValueWithResolution(v, allWrapped, visited)))
		}

		constraints = append(constraints, ConstraintInfo{
			Path:      path,
			Severity:  severity,
			Message:   defaultMessage,
			Component: simplifyIRI(cw.Type),
			Details:   strings.Join(details, ", "),
		})
	}

	return constraints
}

func formatValueWithResolution(v any, allWrapped map[string]*ShapeWrapper, visited map[string]bool) string {
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
		var items []string
		for _, item := range val {
			items = append(items, formatValueWithResolution(item, allWrapped, visited))
		}
		return "[" + strings.Join(items, ", ") + "]"
	}
	return fmt.Sprint(v)
}

func resolveShapeConstraints(sw *ShapeWrapper, allWrapped map[string]*ShapeWrapper, visited map[string]bool) string {
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
	return "{" + strings.Join(parts, ", ") + "}"
}
