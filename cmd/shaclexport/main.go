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
	sb.WriteString(fmt.Sprintf("# %s\n\n", title))

	var keys []string
	for k := range wrapped {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var shapes []*ShapeWrapper
	for _, k := range keys {
		shapes = append(shapes, wrapped[k])
	}

	renderShapes(&sb, shapes, 2)

	return sb.String()
}

func logicKey(sw *ShapeWrapper) string {
	type logic struct {
		Severity    string
		Messages    []string
		Description []string
		Constraints []ConstraintWrapper
		Closed      bool
		ClosedBy    bool
		Ignored     []string
		Properties  []string // recursion
	}

	l := logic{
		Severity: sw.Severity.Value(),
		Closed:   sw.Closed,
		ClosedBy: sw.ClosedByTypes,
	}
	for _, m := range sw.Messages {
		l.Messages = append(l.Messages, m.String())
	}
	sort.Strings(l.Messages)
	for _, d := range sw.Description {
		l.Description = append(l.Description, d.String())
	}
	sort.Strings(l.Description)
	for _, ip := range sw.IgnoredProperties {
		l.Ignored = append(l.Ignored, ip.Value())
	}
	sort.Strings(l.Ignored)
	l.Constraints = sw.Constraints // ConstraintWrapper already marshals to stable JSON
	for _, p := range sw.Properties {
		l.Properties = append(l.Properties, logicKey(p))
	}
	sort.Strings(l.Properties)

	data, _ := json.Marshal(l)
	return string(data)
}

func renderShapes(sb *strings.Builder, shapes []*ShapeWrapper, level int) {
	if len(shapes) == 0 {
		return
	}

	// Group shapes by logicKey
	groups := make(map[string][]*ShapeWrapper)
	var keys []string
	for _, s := range shapes {
		key := logicKey(s)
		if _, ok := groups[key]; !ok {
			keys = append(keys, key)
		}
		groups[key] = append(groups[key], s)
	}

	for _, key := range keys {
		group := groups[key]
		first := group[0]

		var titles []string
		for _, s := range group {
			title := simplifyTerm(s.ID)
			if s.IsProperty && s.Path != nil {
				title = fmt.Sprintf("`%s`", formatPath(s.Path))
			}
			titles = append(titles, title)
		}
		sort.Strings(titles)

		heading := ""
		if first.IsProperty && first.Path != nil {
			if len(group) > 1 {
				heading = fmt.Sprintf("Properties (%d)", len(group))
			} else {
				heading = "Property"
			}
		} else {
			if len(group) > 1 {
				heading = fmt.Sprintf("Shapes (%d)", len(group))
			} else {
				heading = "Shape"
			}
		}

		sb.WriteString(fmt.Sprintf("%s %s\n\n", strings.Repeat("#", level), heading))

		for _, t := range titles {
			sb.WriteString(fmt.Sprintf("- %s\n", t))
		}
		sb.WriteString("\n")

		if len(first.Description) > 0 {
			for _, d := range first.Description {
				sb.WriteString(fmt.Sprintf("%s\n\n", d.Value()))
			}
		}

		if first.Severity.Value() != "" && first.Severity.Value() != "http://www.w3.org/ns/shacl#Violation" {
			sb.WriteString(fmt.Sprintf("**Severity:** %s\n\n", simplifyTerm(first.Severity)))
		}

		if len(first.Messages) > 0 {
			sb.WriteString("**Messages:**\n")
			for _, m := range first.Messages {
				sb.WriteString(fmt.Sprintf("- %s\n", simplifyTerm(m)))
			}
			sb.WriteString("\n")
		}

		// Targets are usually on NodeShapes (not property shapes)
		// If we grouped NodeShapes, we should show all their targets.
		var allTargets []string
		seenTargets := make(map[string]bool)
		for _, s := range group {
			if !s.IsProperty {
				for _, t := range s.Targets {
					ts := fmt.Sprintf("- %s: %s", t.Kind.String(), simplifyTerm(t.Value))
					if !seenTargets[ts] {
						allTargets = append(allTargets, ts)
						seenTargets[ts] = true
					}
				}
			}
		}
		if len(allTargets) > 0 {
			sb.WriteString("**Targets:**\n")
			for _, t := range allTargets {
				sb.WriteString(t + "\n")
			}
			sb.WriteString("\n")
		}

		type sparqlInfo struct {
			id    string
			query string
		}
		var sparqlQueries []sparqlInfo

		if first.Values != nil {
			query := first.Values.Prefixes + first.Values.Select
			if first.Values.Expr != "" {
				query = first.Values.Prefixes + "SELECT (" + first.Values.Expr + " AS ?value) WHERE { $this ?p ?o }"
			}
			sparqlQueries = append(sparqlQueries, sparqlInfo{id: "Values", query: query})
			sb.WriteString("**SPARQL Values:** [See below](#sparql-values)\n\n")
		}

		if len(first.Constraints) > 0 {
			sb.WriteString("**Constraints:**\n\n")
			sb.WriteString("| Component | Details |\n")
			sb.WriteString("| --- | --- |\n")
			for i, c := range first.Constraints {
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

		if len(first.Properties) > 0 {
			sb.WriteString("**Nested Properties:**\n\n")
			renderShapes(sb, first.Properties, level+1)
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

type fileStats struct {
	name   string
	mdPath string
	counts map[string]int
}

func main() {
	flagJSON := flag.Bool("json", false, "Generate JSON output")
	flagMD := flag.Bool("md", false, "Generate Markdown output")
	shaclPattern := flag.String("shacl", SHACL_SCHEMA, "glob pattern for shacl files")
	outputDir := flag.String("out", "pages/docs", "output directory for generated files")
	flag.Parse()

	doJSON, doMD := *flagJSON, *flagMD
	if !doJSON && !doMD {
		doMD = true
	}

	shaclFiles, err := filepath.Glob(*shaclPattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error globbing files: %v\n", err)
		return
	}

	var allStats []fileStats
	allConstraintTypes := make(map[string]bool)

	for _, file := range shaclFiles {
		stats, err := processFile(file, doJSON, doMD, *outputDir, allConstraintTypes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", file, err)
			continue
		}
		allStats = append(allStats, stats)
	}

	if doMD && len(allStats) > 0 {
		writeOverview(*outputDir, allStats, allConstraintTypes)
	}
}

func processFile(file string, doJSON, doMD bool, outputDir string, allConstraintTypes map[string]bool) (fileStats, error) {
	g, err := shacl.LoadTurtleFile(file)
	if err != nil {
		return fileStats{}, err
	}

	shapes := shacl.ParseShapes(g)
	wrapped := make(map[string]*ShapeWrapper)
	baseName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	stats := fileStats{name: baseName, counts: make(map[string]int)}

	isNested := make(map[string]bool)
	for _, s := range shapes {
		for _, ps := range s.Properties {
			isNested[ps.ID.String()] = true
		}
	}

	for k, s := range shapes {
		w := wrapShape(s)
		if !isNested[k] {
			wrapped[k] = w
		}
		for _, c := range w.Constraints {
			typeName := simplifyIRI(c.Type)
			stats.counts[typeName]++
			allConstraintTypes[typeName] = true
		}
	}

	if doJSON {
		jsonOutDir := filepath.Join(outputDir, "json")
		if err := os.MkdirAll(jsonOutDir, 0755); err != nil {
			return stats, err
		}
		data, _ := json.MarshalIndent(wrapped, "", "  ")
		jsonFile := filepath.Join(jsonOutDir, baseName+".json")
		if err := os.WriteFile(jsonFile, data, 0644); err != nil {
			return stats, err
		}
		fmt.Printf("Exported JSON to %s\n", jsonFile)
	}

	if doMD {
		mdOutDir := filepath.Join(outputDir, "SHACL")
		if err := os.MkdirAll(mdOutDir, 0755); err != nil {
			return stats, err
		}
		mdFile := filepath.Join(mdOutDir, baseName+".md")
		stats.mdPath = mdFile
		mdData := generateMarkdown(baseName, wrapped)
		if err := os.WriteFile(mdFile, []byte(mdData), 0644); err != nil {
			return stats, err
		}
		fmt.Printf("Exported Markdown to %s\n", mdFile)
	}

	return stats, nil
}

func writeOverview(outputDir string, allStats []fileStats, allConstraintTypes map[string]bool) {
	var sb strings.Builder
	sb.WriteString("# SHACL Constraints Overview\n\n")

	var types []string
	for t := range allConstraintTypes {
		types = append(types, t)
	}
	sort.Strings(types)

	sb.WriteString("| File | " + strings.Join(types, " | ") + " |\n")
	sb.WriteString("| --- | " + strings.Repeat(" --- |", len(types)) + "\n")

	totals := make(map[string]int)
	for _, s := range allStats {
		fileName := s.name
		if s.mdPath != "" {
			fileName = fmt.Sprintf("[%s](SHACL/%s.md)", s.name, s.name)
		}
		row := "| " + fileName
		for _, t := range types {
			count := s.counts[t]
			totals[t] += count
			if count == 0 {
				row += " | -"
			} else {
				row += fmt.Sprintf(" | %d", count)
			}
		}
		row += " |\n"
		sb.WriteString(row)
	}

	row := "| **Total**"
	for _, t := range types {
		row += fmt.Sprintf(" | **%d**", totals[t])
	}
	row += " |\n"
	sb.WriteString(row)

	shaclOverviewFile := filepath.Join(outputDir, "SHACL-Overview.md")
	if err := os.WriteFile(shaclOverviewFile, []byte(sb.String()), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing overview Markdown: %v\n", err)
	} else {
		fmt.Println("Exported Overview to SHACL-Overview.md")
	}
}

