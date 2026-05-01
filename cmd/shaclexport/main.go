package main

import (
	"cimgo/rdf/shacl"
	"cimgo/validation"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// generateMarkdown creates a Markdown string for the given shapes, filtering constraints based on the provided filter function
func generateMarkdown(title string, wrapped map[string]*validation.ShapeWrapper, filter func(validation.ConstraintWrapper) bool) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", title))

	var keys []string
	for k, w := range wrapped {
		if validation.HasContent(w, filter) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	if len(keys) == 0 {
		return ""
	}

	var shapes []*validation.ShapeWrapper
	for _, k := range keys {
		shapes = append(shapes, wrapped[k])
	}

	renderShapes(&sb, shapes, 2, filter)

	return sb.String()
}

func renderShapes(sb *strings.Builder, shapes []*validation.ShapeWrapper, level int, filter func(validation.ConstraintWrapper) bool) {
	if len(shapes) == 0 {
		return
	}

	// Sort shapes: properties first, then by path, then by ID
	sort.Slice(shapes, func(i, j int) bool {
		si, sj := shapes[i], shapes[j]
		if si.IsProperty && sj.IsProperty && si.Path != nil && sj.Path != nil {
			return validation.FormatPathString(si.Path) < validation.FormatPathString(sj.Path)
		}
		return si.ID.String() < sj.ID.String()
	})

	for _, s := range shapes {
		renderShapeHeading(sb, s, level)
		renderShapeBasicInfo(sb, s)
		renderShapeTargets(sb, s)

		var sparqlQueries []validation.SparqlInfo
		sparqlQueries = validation.CollectSPARQLValues(sb, s, sparqlQueries)

		filteredConstraints := validation.FilterConstraints(s, filter)
		if len(filteredConstraints) > 0 {
			sparqlQueries = renderConstraintsTable(sb, filteredConstraints, sparqlQueries)
		}

		renderSPARQLQueries(sb, sparqlQueries)
		renderNestedProperties(sb, s, level, filter)
	}
}

func renderShapeHeading(sb *strings.Builder, s *validation.ShapeWrapper, level int) {
	title := validation.SimplifyTerm(s.ID)
	sb.WriteString(fmt.Sprintf("%s %s\n\n", strings.Repeat("#", level), title))
}

func renderShapeBasicInfo(sb *strings.Builder, first *validation.ShapeWrapper) {
	if first.IsProperty && first.Path != nil {
		sb.WriteString(fmt.Sprintf("**Path:** `%s`  \n", validation.FormatPathString(first.Path)))
	}

	if len(first.Description) > 0 {
		for _, d := range first.Description {
			sb.WriteString(fmt.Sprintf("%s\n\n", d.Value()))
		}
	}

	if first.Severity.Value() != "" && first.Severity.Value() != "http://www.w3.org/ns/shacl#Violation" {
		sb.WriteString(fmt.Sprintf("**Severity:** %s\n\n", validation.SimplifyTerm(first.Severity)))
	}

	if len(first.Messages) > 0 {
		sb.WriteString("**Messages:**\n")
		for _, m := range first.Messages {
			sb.WriteString(fmt.Sprintf("- %s\n", validation.SimplifyTerm(m)))
		}
		sb.WriteString("\n")
	}
}

func renderShapeTargets(sb *strings.Builder, s *validation.ShapeWrapper) {
	if s.IsProperty {
		return
	}
	if len(s.Targets) > 0 {
		sb.WriteString("**Targets:**\n")
		for _, t := range s.Targets {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", t.Kind.String(), validation.SimplifyTerm(t.Value)))
		}
		sb.WriteString("\n")
	}
}

func renderConstraintsTable(sb *strings.Builder, constraints []validation.ConstraintWrapper, queries []validation.SparqlInfo) []validation.SparqlInfo {
	sb.WriteString("**Constraints:**\n\n")
	sb.WriteString("| Component | Details |\n")
	sb.WriteString("| --- | --- |\n")
	for i, c := range constraints {
		typeName := validation.SimplifyIRI(c.Type)
		displayData := c.Data
		var severityOverride string
		if soc, ok := c.Data.(*shacl.SeverityOverrideConstraint); ok {
			displayData = soc.Inner()
			severityOverride = fmt.Sprintf("<br>**Severity:** %s", validation.SimplifyTerm(soc.Severity))
		}

		data, _ := json.Marshal(displayData)
		var m map[string]any
		json.Unmarshal(data, &m)
		var details []string

		if sc, ok := displayData.(*shacl.SPARQLConstraint); ok {
			id := fmt.Sprintf("SPARQL-%d", i+1)
			queries = append(queries, validation.SparqlInfo{Id: id, Query: sc.Prefixes + sc.Select})
			details = append(details, fmt.Sprintf("Query: [See %s below](#%s) ", id, strings.ToLower(id)))
			if len(sc.Messages) > 0 {
				var msgs []string
				for _, msg := range sc.Messages {
					msgs = append(msgs, validation.SimplifyTerm(msg))
				}
				details = append(details, fmt.Sprintf("Messages: `[%s]` ", strings.Join(msgs, ", ")))
			}
		} else {
			for k, v := range m {
				if k != "Prefixes" {
					details = append(details, fmt.Sprintf("%s: `%s` ", k, validation.FormatValue(v)))
				}
			}
		}
		sort.Strings(details)
		sb.WriteString(fmt.Sprintf("| %s | %s%s |\n", typeName, strings.Join(details, "<br>"), severityOverride))
	}
	sb.WriteString("\n")
	return queries
}

func renderSPARQLQueries(sb *strings.Builder, queries []validation.SparqlInfo) {
	if len(queries) > 0 {
		sb.WriteString("#### SPARQL Queries\n\n")
		for _, sq := range queries {
			sb.WriteString(fmt.Sprintf("##### %s\n```sparql\n%s\n```\n\n", sq.Id, sq.Query))
		}
	}
}

func renderNestedProperties(sb *strings.Builder, sw *validation.ShapeWrapper, level int, filter func(validation.ConstraintWrapper) bool) {
	var filteredProperties []*validation.ShapeWrapper
	for _, p := range sw.Properties {
		if validation.HasContent(p, filter) {
			filteredProperties = append(filteredProperties, p)
		}
	}

	if len(filteredProperties) > 0 {
		sb.WriteString("**Nested Properties:**\n\n")
		renderShapes(sb, filteredProperties, level+1, filter)
	}
}

func main() {
	flagJSON := flag.Bool("json", false, "Generate JSON output")
	flagMD := flag.Bool("md", false, "Generate Markdown output")
	shaclPattern := flag.String("shacl", validation.DefaultSHACLPattern, "glob pattern for shacl files")
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

	var allStats []validation.FileStats
	allSHACLTypes := make(map[string]bool)
	allSPARQLTypes := make(map[string]bool)

	for _, file := range shaclFiles {
		stats, err := processFile(file, doJSON, doMD, *outputDir, allSHACLTypes, allSPARQLTypes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", file, err)
			continue
		}
		allStats = append(allStats, stats)
	}

	if doMD && len(allStats) > 0 {
		writeOverview(*outputDir, allStats, allSHACLTypes, allSPARQLTypes)
	}
}

func processFile(file string, doJSON, doMD bool, outputDir string, allSHACLTypes, allSPARQLTypes map[string]bool) (validation.FileStats, error) {
	g, err := shacl.LoadTurtleFile(file)
	if err != nil {
		return validation.FileStats{}, err
	}

	shapes := shacl.ParseShapes(g)
	wrapped := make(map[string]*validation.ShapeWrapper)
	baseName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	stats := validation.FileStats{
		Name:         baseName,
		ShaclCounts:  make(map[string]int),
		SparqlCounts: make(map[string]int),
	}

	isNested := identifyNestedShapes(shapes)

	for k, s := range shapes {
		w := validation.WrapShape(s)
		if !isNested[k] {
			wrapped[k] = w
		}
		updateStats(&stats, w, allSHACLTypes, allSPARQLTypes)
	}

	if doJSON {
		if err := exportJSON(outputDir, baseName, wrapped); err != nil {
			return stats, err
		}
	}

	if doMD {
		exportMD(&stats, outputDir, baseName, wrapped)
	}

	return stats, nil
}

func identifyNestedShapes(shapes map[string]*shacl.Shape) map[string]bool {
	isNested := make(map[string]bool)
	for _, s := range shapes {
		for _, ps := range s.Properties {
			isNested[ps.ID.String()] = true
		}
	}
	return isNested
}

func updateStats(stats *validation.FileStats, w *validation.ShapeWrapper, allSHACLTypes, allSPARQLTypes map[string]bool) {
	for _, c := range w.Constraints {
		typeName := validation.SimplifyIRI(c.Type)
		if c.IsSPARQL() {
			stats.SparqlCounts[typeName]++
			allSPARQLTypes[typeName] = true
		} else {
			stats.ShaclCounts[typeName]++
			allSHACLTypes[typeName] = true
		}
	}
}

func exportJSON(outputDir, baseName string, wrapped map[string]*validation.ShapeWrapper) error {
	jsonOutDir := filepath.Join(outputDir, "json")
	if err := os.MkdirAll(jsonOutDir, 0755); err != nil {
		return err
	}
	data, _ := json.MarshalIndent(wrapped, "", "  ")
	jsonFile := filepath.Join(jsonOutDir, baseName+".json")
	if err := os.WriteFile(jsonFile, data, 0644); err != nil {
		return err
	}
	fmt.Printf("Exported JSON to %s\n", jsonFile)
	return nil
}

func exportMD(stats *validation.FileStats, outputDir, baseName string, wrapped map[string]*validation.ShapeWrapper) {
	shaclMD := generateMarkdown(baseName, wrapped, func(cw validation.ConstraintWrapper) bool { return cw.IsSHACL() })
	if shaclMD != "" {
		mdOutDir := filepath.Join(outputDir, "SHACL")
		os.MkdirAll(mdOutDir, 0755)
		mdFile := filepath.Join(mdOutDir, baseName+".md")
		stats.ShaclPath = mdFile
		os.WriteFile(mdFile, []byte(shaclMD), 0644)
		fmt.Printf("Exported SHACL MD to %s\n", mdFile)
	}

	sparqlMD := generateMarkdown(baseName, wrapped, func(cw validation.ConstraintWrapper) bool { return cw.IsSPARQL() })
	if sparqlMD != "" {
		mdOutDir := filepath.Join(outputDir, "SPARQL")
		os.MkdirAll(mdOutDir, 0755)
		mdFile := filepath.Join(mdOutDir, baseName+".md")
		stats.SparqlPath = mdFile
		os.WriteFile(mdFile, []byte(sparqlMD), 0644)
		fmt.Printf("Exported SPARQL MD to %s\n", mdFile)
	}
}

func writeOverview(outputDir string, allStats []validation.FileStats, allSHACLTypes, allSPARQLTypes map[string]bool) {
	writeOverviewFile(filepath.Join(outputDir, "SHACL-Overview.md"), "SHACL", allStats, allSHACLTypes, true)
	writeOverviewFile(filepath.Join(outputDir, "SPARQL-Overview.md"), "SPARQL", allStats, allSPARQLTypes, false)
}

func writeOverviewFile(path, title string, allStats []validation.FileStats, typesMap map[string]bool, isSHACL bool) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s Overview\n\n", title))

	types := sortedKeys(typesMap)
	if len(types) == 0 {
		return
	}

	renderOverviewTableHeader(&sb, types)

	totals := make(map[string]int)
	for _, s := range allStats {
		totals = renderOverviewRow(&sb, s, types, isSHACL, totals)
	}

	renderOverviewTableFooter(&sb, types, totals)

	if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", path, err)
	} else {
		fmt.Printf("Exported Overview to %s\n", path)
	}
}

func sortedKeys(m map[string]bool) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func renderOverviewTableHeader(sb *strings.Builder, types []string) {
	sb.WriteString("| File | " + strings.Join(types, " | ") + " |\n")
	sb.WriteString("| --- | " + strings.Repeat(" --- |", len(types)) + "\n")
}

func renderOverviewRow(sb *strings.Builder, s validation.FileStats, types []string, isSHACL bool, totals map[string]int) map[string]int {
	var filePath string
	var counts map[string]int
	if isSHACL {
		filePath, counts = s.ShaclPath, s.ShaclCounts
	} else {
		filePath, counts = s.SparqlPath, s.SparqlCounts
	}

	if len(counts) == 0 {
		return totals
	}

	fileName := s.Name
	if filePath != "" {
		relPath := filepath.Base(filepath.Dir(filePath)) + "/" + filepath.Base(filePath)
		fileName = fmt.Sprintf("[%s](%s)", s.Name, relPath)
	}

	sb.WriteString("| " + fileName)
	for _, t := range types {
		count := counts[t]
		totals[t] += count
		if count == 0 {
			sb.WriteString(" | -")
		} else {
			sb.WriteString(fmt.Sprintf(" | %d", count))
		}
	}
	sb.WriteString(" |\n")
	return totals
}

func renderOverviewTableFooter(sb *strings.Builder, types []string, totals map[string]int) {
	sb.WriteString("| **Total**")
	for _, t := range types {
		sb.WriteString(fmt.Sprintf(" | **%d**", totals[t]))
	}
	sb.WriteString(" |\n")
}
