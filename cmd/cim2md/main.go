package main

import (
	"cimgo/cimgen"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

func main() {
	schemaPattern := flag.String("schema", cimgen.DefaultRDFSPattern, "glob pattern for CIM schema files")
	outputDir := flag.String("out", "docs", "output directory for markdown files")
	profileAttributes := flag.Bool("profile-attributes", false, "include attribute lists in profile overview mermaid diagrams")
	flag.Parse()

	classesDir := filepath.Join(*outputDir, "Classes")
	profilesDir := filepath.Join(*outputDir, "Profiles")

	if err := os.MkdirAll(classesDir, 0755); err != nil {
		fmt.Printf("Error creating classes directory: %v\n", err)
		return
	}
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		fmt.Printf("Error creating profiles directory: %v\n", err)
		return
	}

	cimSpec := cimgen.NewCIMSpecification()
	fmt.Printf("Importing schema files: %s\n", *schemaPattern)
	if err := cimSpec.ImportCIMSchemaFiles(*schemaPattern); err != nil {
		fmt.Printf("Error importing schema files: %v\n", err)
		return
	}

	profiles := make(map[string][]string)
	for name, entity := range cimSpec.Types {
		for _, cat := range entity.CIMCategories {
			profiles[cat] = append(profiles[cat], name)
		}
	}

	// Build subtypes map for inheritance diagrams
	subtypes := make(map[string][]string)
	for name, entity := range cimSpec.Types {
		if entity.SuperType != "" {
			subtypes[entity.SuperType] = append(subtypes[entity.SuperType], name)
		}
	}

	// Generate Class Pages
	for name, entity := range cimSpec.Types {
		generateClassPage(name, entity, classesDir, subtypes, cimSpec.Enums, cimSpec.Types)
	}

	// Generate Enum Pages
	for name, entity := range cimSpec.Enums {
		generateEnumPage(name, entity, classesDir)
	}

	// Generate Profile Pages
	for name, clsList := range profiles {
		generateProfilePage(name, clsList, profilesDir, cimSpec.Types, cimSpec.Enums, *profileAttributes)
	}

	// Generate Indexes
	generateProfilesIndex(*outputDir, profiles)
	generateClassesIndex(*outputDir, cimSpec.Types, cimSpec.Enums)
	fmt.Printf("Documentation generated in %s\n", *outputDir)
}

func sanitizeFilename(name string) string {
	var sb strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '.' || r == '_' {
			sb.WriteRune(r)
		}
	}
	return strings.ReplaceAll(strings.TrimSpace(sb.String()), " ", "_")
}

func stringifyMultiplicity(attribute *cimgen.CIMAttribute) string {
	multi := ""
	if strings.Contains(attribute.CIMMultiplicity, "#M:") {
		parts := strings.Split(attribute.CIMMultiplicity, "#M:")
		multi = parts[len(parts)-1]
	}
	return multi
}

func stringifyCIMDataType(attribute *cimgen.CIMAttribute) string {
	attrType := attribute.DataType
	if attrType == "" {
		attrType = attribute.RDFRange
	}
	if attrType == "" {
		attrType = "N/A"
	}
	return attrType
}

func generateAttributesForMermaid(f *os.File, className string, attributes []*cimgen.CIMAttribute) {
	for _, attr := range attributes {
		fmt.Fprintf(f, "    %s : +%s %s[%s]\n", className, stringifyCIMDataType(attr), attr.Label, stringifyMultiplicity(attr))
	}
}

func generateClassPage(name string, data *cimgen.CIMType, outDir string, subtypes map[string][]string, enums map[string]*cimgen.CIMEnum, allClasses map[string]*cimgen.CIMType) {
	filename := filepath.Join(outDir, sanitizeFilename(name)+".md")
	label := data.Label
	if label == "" {
		label = name
	}
	comment := data.Comment
	if comment == "" {
		comment = "No description available."
	}

	f, _ := os.Create(filename)
	defer f.Close()

	fmt.Fprintf(f, "# %s\n\n", label)
	fmt.Fprintf(f, "%s\n\n", comment)

	if data.SuperType != "" || len(subtypes[name]) > 0 {
		fmt.Fprintf(f, "## Inheritance\n\n")
		fmt.Fprintf(f, "```mermaid\n---\n  config:\n    class:\n      hideEmptyMembersBox: true\n---\nclassDiagram\n")

		if data.SuperType != "" {
			fmt.Fprintf(f, "    %s <|-- %s\n", data.SuperType, name)
			superType, superTypeExists := allClasses[data.SuperType]
			if superTypeExists && len(superType.Attributes) != 0 {
				generateAttributesForMermaid(f, data.SuperType, superType.Attributes)
			}
		}
		for _, sub := range subtypes[name] {
			fmt.Fprintf(f, "    %s <|-- %s\n", name, sub)
			subType, subTypeExists := allClasses[sub]
			if subTypeExists && len(subType.Attributes) != 0 {
				generateAttributesForMermaid(f, sub, subType.Attributes)
			}
		}
		generateAttributesForMermaid(f, name, data.Attributes)

		fmt.Fprintf(f, "```\n")
		fmt.Fprintf(f, "<button class=\"mermaid-enlarge-button\">Enlarge Diagram</button>\n\n")
	}

	fmt.Fprintf(f, "## Attributes\n\n")
	fmt.Fprintf(f, "| Label | Type | Multiplicity | Comment |\n")
	fmt.Fprintf(f, "|-------|------|--------------|---------|\n")

	if len(data.Attributes) == 0 {
		fmt.Fprintf(f, "| No attributes | | | |\n")
	} else {
		for _, attr := range data.Attributes {
			attrType := stringifyCIMDataType(attr)

			typeLink := attrType
			_, isClass := allClasses[attrType]
			_, isEnum := enums[attrType]
			if isClass || isEnum {
				typeLink = fmt.Sprintf("[%s](%s.md)", attrType, sanitizeFilename(attrType))
			}

			attrComment := strings.ReplaceAll(attr.Comment, "\n", " ")
			fmt.Fprintf(f, "| %s | %s | %s | %s |\n", attr.Label, typeLink, stringifyMultiplicity(attr), attrComment)
		}
	}
	fmt.Fprint(f, "\n")
}

func generateEnumPage(name string, data *cimgen.CIMEnum, outDir string) {
	filename := filepath.Join(outDir, sanitizeFilename(name)+".md")
	label := data.Label
	if label == "" {
		label = name
	}
	comment := data.Comment
	if comment == "" {
		comment = "No description available."
	}

	f, _ := os.Create(filename)
	defer f.Close()

	fmt.Fprintf(f, "# %s (Enumeration)\n\n", label)
	fmt.Fprintf(f, "%s\n\n", comment)

	fmt.Fprintf(f, "## Values\n\n")
	fmt.Fprintf(f, "| Label | Comment |\n")
	fmt.Fprintf(f, "|-------|---------|\n")

	for _, val := range data.Values {
		valComment := strings.ReplaceAll(val.Comment, "\n", " ")
		fmt.Fprintf(f, "| %s | %s |\n", val.Label, valComment)
	}
	fmt.Fprint(f, "\n")
}

func generateProfilePage(name string, clsNames []string, outDir string, allClasses map[string]*cimgen.CIMType, allEnums map[string]*cimgen.CIMEnum, showAttributes bool) {
	filename := filepath.Join(outDir, sanitizeFilename(name)+".md")
	f, _ := os.Create(filename)
	defer f.Close()

	fmt.Fprintf(f, "# %s\n\n", name)

	if len(clsNames) > 1 {
		fmt.Fprintf(f, "## Overview Diagram\n\n")
		fmt.Fprintf(f, "```mermaid\nclassDiagram\n")

		relevant := make(map[string]bool)
		for _, n := range clsNames {
			relevant[n] = true
		}

		for _, n := range clsNames {
			clsData, ok := allClasses[n]
			if !ok {
				continue
			}

			// rendering type inheritance inside package
			super := clsData.SuperType
			if relevant[super] {
				fmt.Fprintf(f, "    %s <|-- %s\n", super, n)
				superType, superTypeExists := allClasses[super]
				if showAttributes && superTypeExists && len(superType.Attributes) != 0 {
					generateAttributesForMermaid(f, super, superType.Attributes)
				}
			}

			// rendering package internal references
			for _, attr := range clsData.Attributes {
				attrType := attr.DataType
				if attrType == "" {
					attrType = attr.RDFRange
				}
				if relevant[attrType] {
					fmt.Fprintf(f, "    %s --> %s : %s\n", n, attrType, attr.Label)
				}
			}

			// rendering attributes of package objects
			if showAttributes {
				generateAttributesForMermaid(f, n, clsData.Attributes)
			}
		}
		fmt.Fprintf(f, "```\n")
		fmt.Fprintf(f, "<button class=\"mermaid-enlarge-button\">Enlarge Diagram</button>\n\n")
	}

	fmt.Fprintf(f, "## Classes\n\n")
	sort.Strings(clsNames)
	for _, n := range clsNames {
		label := n
		comment := ""
		if cls, ok := allClasses[n]; ok {
			if cls.Label != "" {
				label = cls.Label
			}
			comment = cls.Comment
		} else if en, ok := allEnums[n]; ok {
			if en.Label != "" {
				label = en.Label
			}
			comment = en.Comment
		}

		firstDot := strings.Index(comment, ".")
		if firstDot != -1 {
			comment = comment[:firstDot+1]
		}

		fmt.Fprintf(f, "- [%s](../Classes/%s): %s\n", label, sanitizeFilename(n), comment)
	}
}

func generateProfilesIndex(outDir string, profiles map[string][]string) {
	filename := filepath.Join(outDir, "Profiles.md")
	f, _ := os.Create(filename)
	defer f.Close()

	fmt.Fprintf(f, "# CIM Profiles\n\n")

	var profileNames []string
	for p := range profiles {
		profileNames = append(profileNames, p)
	}
	sort.Strings(profileNames)
	for _, p := range profileNames {
		fmt.Fprintf(f, "- [%s](Profiles/%s)\n", p, sanitizeFilename(p))
	}
}

func generateClassesIndex(outDir string, classes map[string]*cimgen.CIMType, enums map[string]*cimgen.CIMEnum) {
	filename := filepath.Join(outDir, "Classes.md")
	f, _ := os.Create(filename)
	defer f.Close()

	fmt.Fprintf(f, "# CIM Classes and Enums\n\n")

	var allNames []string
	for n := range classes {
		allNames = append(allNames, n)
	}
	for n := range enums {
		allNames = append(allNames, n)
	}
	sort.Strings(allNames)

	for _, n := range allNames {
		label := n
		suffix := ""
		if cls, ok := classes[n]; ok && cls.Label != "" {
			label = cls.Label
		} else if en, ok := enums[n]; ok {
			if en.Label != "" {
				label = en.Label
			}
			suffix = " (Enum)"
		}
		fmt.Fprintf(f, "- [%s](Classes/%s)%s\n", label, sanitizeFilename(n), suffix)
	}
}
