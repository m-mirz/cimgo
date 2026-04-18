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

type ConstraintWrapper struct {
	Type string
	Data shacl.Constraint
}

func (cw ConstraintWrapper) MarshalJSON() ([]byte, error) {
	// We want to combine the Type field and the data from the constraint
	data, err := json.Marshal(cw.Data)
	if err != nil {
		return nil, err
	}

	var m map[string]any
	err = json.Unmarshal(data, &m)
	if err != nil {
		// If it's not a map (e.g. basic type), just use a simple wrapper
		return json.Marshal(map[string]any{
			"type": cw.Type,
			"data": cw.Data,
		})
	}

	m["_type"] = cw.Type
	return json.Marshal(m)
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
	sw := &ShapeWrapper{
		Shape: s,
	}
	for _, c := range s.Constraints {
		sw.Constraints = append(sw.Constraints, ConstraintWrapper{
			Type: c.ComponentIRI(),
			Data: c,
		})
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

func generateMarkdown(title string, wrapped map[string]*ShapeWrapper) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# SHACL Shapes Export: %s\n\n", title))

	var keys []string
	for k := range wrapped {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		sw := wrapped[k]
		if sw.IsProperty && sw.Path != nil {
			// Skip standalone property shapes if they are already nested,
			// but wait, how do we know? Actually, we'll just list all shapes.
		}
		renderShape(&sb, sw, 2)
	}

	return sb.String()
}

func renderShape(sb *strings.Builder, sw *ShapeWrapper, level int) {
	title := simplifyTerm(sw.ID)
	if sw.IsProperty && sw.Path != nil {
		title = fmt.Sprintf("Property: `%s`", formatPath(sw.Path))
	}

	sb.WriteString(fmt.Sprintf("%s %s\n\n", strings.Repeat("#", level), title))

	if sw.Severity.Value() != "" && sw.Severity.Value() != "http://www.w3.org/ns/shacl#Violation" {
		sb.WriteString(fmt.Sprintf("**Severity:** %s\n\n", simplifyTerm(sw.Severity)))
	}

	if len(sw.Messages) > 0 {
		sb.WriteString("**Messages:**\n")
		for _, m := range sw.Messages {
			sb.WriteString(fmt.Sprintf("- %s\n", simplifyTerm(m)))
		}
		sb.WriteString("\n")
	}

	if !sw.IsProperty {
		if len(sw.Targets) > 0 {
			sb.WriteString("**Targets:**\n")
			for _, t := range sw.Targets {
				sb.WriteString(fmt.Sprintf("- %s: %s\n", t.Kind.String(), simplifyTerm(t.Value)))
			}
			sb.WriteString("\n")
		}
	}

	type sparqlInfo struct {
		id    string
		query string
	}
	var sparqlQueries []sparqlInfo

	if sw.Values != nil {
		query := sw.Values.Prefixes + sw.Values.Select
		if sw.Values.Expr != "" {
			query = sw.Values.Prefixes + "SELECT (" + sw.Values.Expr + " AS ?value) WHERE { $this ?p ?o }"
		}
		sparqlQueries = append(sparqlQueries, sparqlInfo{id: "Values", query: query})
		sb.WriteString("**SPARQL Values:** [See below](#sparql-values)\n\n")
	}

	if len(sw.Constraints) > 0 {
		sb.WriteString("**Constraints:**\n\n")
		sb.WriteString("| Component | Details |\n")
		sb.WriteString("| --- | --- |\n")
		for i, c := range sw.Constraints {
			typeName := simplifyIRI(c.Type)

			displayData := c.Data
			var severityOverride string
			if soc, ok := c.Data.(*shacl.SeverityOverrideConstraint); ok {
				displayData = soc.Inner()
				severityOverride = fmt.Sprintf("<br>**Severity:** %s", simplifyTerm(soc.Severity))
			}

			data, _ := json.Marshal(displayData)
			var m map[string]any
			json.Unmarshal(data, &m)
			var details []string

			// Special handling for SPARQLConstraint
			if sc, ok := displayData.(*shacl.SPARQLConstraint); ok {
				id := fmt.Sprintf("SPARQL-%d", i+1)
				sparqlQueries = append(sparqlQueries, sparqlInfo{id: id, query: sc.Prefixes + sc.Select})
				details = append(details, fmt.Sprintf("Query: [See %s below](#%s) ", id, strings.ToLower(id)))
				if len(sc.Messages) > 0 {
					var msgs []string
					for _, msg := range sc.Messages {
						msgs = append(msgs, simplifyTerm(msg))
					}
					details = append(details, fmt.Sprintf("Messages: `[%s]` ", strings.Join(msgs, ", ")))
				}
			} else {
				for k, v := range m {
					if k == "Prefixes" {
						continue
					}
					details = append(details, fmt.Sprintf("%s: `%s` ", k, formatValue(v)))
				}
			}
			sort.Strings(details)

			sb.WriteString(fmt.Sprintf("| %s | %s%s |\n", typeName, strings.Join(details, "<br>"), severityOverride))
		}
		sb.WriteString("\n")
	}

	if len(sparqlQueries) > 0 {
		sb.WriteString("#### SPARQL Queries\n\n")
		for _, sq := range sparqlQueries {
			sb.WriteString(fmt.Sprintf("##### %s\n```sparql\n%s\n```\n\n", sq.id, sq.query))
		}
	}

	if len(sw.Properties) > 0 {
		sb.WriteString("**Nested Properties:**\n\n")
		for _, ps := range sw.Properties {
			renderShape(sb, ps, level+1)
		}
	}
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

func main() {
	flagJSON := flag.Bool("json", false, "Generate JSON output")
	flagMD := flag.Bool("md", false, "Generate Markdown output")
	flag.Parse()

	doJSON := *flagJSON
	doMD := *flagMD
	if !doJSON && !doMD {
		doJSON = true
		doMD = true
	}

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: shacl_json_export [-json] [-md] <file1.ttl> [file2.ttl ...]")
		os.Exit(1)
	}

	for _, file := range args {
		g, err := shacl.LoadTurtleFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading %s: %v\n", file, err)
			continue
		}

		shapes := shacl.ParseShapes(g)
		wrapped := make(map[string]*ShapeWrapper)
		for k, s := range shapes {
			wrapped[k] = wrapShape(s)
		}

		dir := filepath.Dir(file)
		baseName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))

		// 1. JSON Export
		if doJSON {
			data, err := json.MarshalIndent(wrapped, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling shapes from %s: %v\n", file, err)
			} else {
				jsonDir := filepath.Join(dir, "json")
				if err := os.MkdirAll(jsonDir, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "Error creating JSON directory %s: %v\n", jsonDir, err)
				} else {
					jsonFile := filepath.Join(jsonDir, baseName+".json")
					err = os.WriteFile(jsonFile, data, 0644)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error writing JSON to %s: %v\n", jsonFile, err)
					} else {
						fmt.Printf("Exported JSON to %s\n", jsonFile)
					}
				}
			}
		}

		// 2. Markdown Export
		if doMD {
			mdDir := filepath.Join(dir, "markdown")
			if err := os.MkdirAll(mdDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating Markdown directory %s: %v\n", mdDir, err)
			} else {
				mdFile := filepath.Join(mdDir, baseName+".md")
				mdData := generateMarkdown(filepath.Base(file), wrapped)
				err = os.WriteFile(mdFile, []byte(mdData), 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error writing Markdown to %s: %v\n", mdFile, err)
				} else {
					fmt.Printf("Exported Markdown to %s\n", mdFile)
				}
			}
		}
	}
}
