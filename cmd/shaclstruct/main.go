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

	"cimgo/validation"
)

const (
	SHACL_SCHEMA = "application-profiles-library/CGMES/CurrentRelease/SHACL/TTL/*.ttl"
)

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

	results := validation.FileResults{
		FileName: baseName,
		Classes:  []validation.ClassInfo{},
	}

	allWrapped := make(map[string]*validation.ShapeWrapper)
	for k := range shapes {
		allWrapped[k] = validation.WrapShape(shapes[k])
	}

	isNested := validation.IdentifyNestedShapes(shapes)

	classMap := make(map[string]*validation.ClassInfo)

	for k := range shapes {
		if isNested[k] {
			continue
		}

		sw := allWrapped[k]

		// Determine target classes/nodes
		var explicitTargets []string
		var implicitTargets []string
		for _, t := range sw.Targets {
			termVal := validation.SimplifyTerm(t.Value)
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
				classInfo := validation.ConvertToClassInfo(sw, allWrapped, className)
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

func mergeIntoClassInfo(info *validation.ClassInfo, sw *validation.ShapeWrapper, allWrapped map[string]*validation.ShapeWrapper) {
	// Add class-level constraints
	info.Constraints = append(info.Constraints, validation.ExtractConstraints(sw, allWrapped, make(map[string]bool))...)

	// Add attributes
	for _, pw := range sw.Properties {
		attrInfo := validation.ConvertToAttributeInfo(pw, allWrapped)
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
