package main

import (
	"cimgo/rdf/shacl"
	"cimgo/shaclimport"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// generateMarkdown creates a Markdown string for the given shapes, filtering constraints based on the provided filter function
func generateMarkdown(title string, topLevel []*shaclimport.ShapeWrapper, allWrapped map[string]*shaclimport.ShapeWrapper, filter func(shaclimport.ConstraintWrapper) bool) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", title))

	if len(topLevel) == 0 {
		return ""
	}

	renderShapes(&sb, topLevel, 2, filter, allWrapped, make(map[string]bool))

	return sb.String()
}

func renderShapes(sb *strings.Builder, shapes []*shaclimport.ShapeWrapper, level int, filter func(shaclimport.ConstraintWrapper) bool, allWrapped map[string]*shaclimport.ShapeWrapper, visited map[string]bool) {
	if len(shapes) == 0 {
		return
	}

	// Sort shapes: properties first, then by path, then by ID
	sort.Slice(shapes, func(i, j int) bool {
		si, sj := shapes[i], shapes[j]
		if si.IsProperty && sj.IsProperty && si.Path != nil && sj.Path != nil {
			return shaclimport.FormatPathString(si.Path) < shaclimport.FormatPathString(sj.Path)
		}
		return si.ID.String() < sj.ID.String()
	})

	for i, s := range shapes {
		id := s.ID.String()
		if visited[id] {
			renderShapeHeading(sb, s, level)
			sb.WriteString(fmt.Sprintf("*(Recursive reference to %s)*\n\n", shaclimport.SimplifyTerm(s.ID)))
			continue
		}
		visited[id] = true

		if s.ID.IsBlank() && len(shapes) > 1 {
			sb.WriteString(fmt.Sprintf("**Item %d:**\n\n", i+1))
		} else {
			renderShapeHeading(sb, s, level)
		}

		renderShapeContent(sb, s, level, filter, allWrapped, visited)

		delete(visited, id)
	}
}

func renderShapeContent(sb *strings.Builder, s *shaclimport.ShapeWrapper, level int, filter func(shaclimport.ConstraintWrapper) bool, allWrapped map[string]*shaclimport.ShapeWrapper, visited map[string]bool) {
	renderShapeBasicInfo(sb, s)
	renderShapeTargets(sb, s)

	if s.Values != nil {
		query := s.Values.Prefixes + s.Values.Select
		if s.Values.Expr != "" {
			query = s.Values.Prefixes + "SELECT (" + s.Values.Expr + " AS ?value) WHERE { $this ?p ?o }"
		}
		sb.WriteString("**SPARQL Values:**\n\n```sparql\n")
		sb.WriteString(query)
		sb.WriteString("\n```\n\n")
	}

	filteredConstraints := shaclimport.FilterConstraints(s, filter)
	if len(filteredConstraints) > 0 {
		renderConstraintsList(sb, filteredConstraints, level, filter, allWrapped, visited)
	}

	renderNestedProperties(sb, s, level, filter, allWrapped, visited)
}

func renderShapeHeading(sb *strings.Builder, s *shaclimport.ShapeWrapper, level int) {
	if s.ID.IsBlank() {
		return
	}
	title := shaclimport.SimplifyTerm(s.ID)
	sb.WriteString(fmt.Sprintf("%s %s\n\n", strings.Repeat("#", level), title))
}

func renderShapeBasicInfo(sb *strings.Builder, first *shaclimport.ShapeWrapper) {
	if first.IsProperty && first.Path != nil {
		sb.WriteString(fmt.Sprintf("**Path:** `%s`  \n", shaclimport.FormatPathString(first.Path)))
	}

	if len(first.Description) > 0 {
		for _, d := range first.Description {
			sb.WriteString(fmt.Sprintf("%s\n\n", d.Value()))
		}
	}

	if first.Severity.Value() != "" && first.Severity.Value() != "http://www.w3.org/ns/shacl#Violation" {
		sb.WriteString(fmt.Sprintf("**Severity:** %s\n\n", shaclimport.SimplifyTerm(first.Severity)))
	}

	if len(first.Messages) > 0 {
		sb.WriteString("**Messages:**\n")
		for _, m := range first.Messages {
			sb.WriteString(fmt.Sprintf("- %s\n", shaclimport.SimplifyTerm(m)))
		}
		sb.WriteString("\n")
	}
}

func renderShapeTargets(sb *strings.Builder, s *shaclimport.ShapeWrapper) {
	if s.IsProperty {
		return
	}
	if len(s.Targets) > 0 {
		sb.WriteString("**Targets:**\n")
		for _, t := range s.Targets {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", t.Kind.String(), shaclimport.SimplifyTerm(t.Value)))
		}
		sb.WriteString("\n")
	}
}

func renderConstraintsList(sb *strings.Builder, constraints []shaclimport.ConstraintWrapper, level int, filter func(shaclimport.ConstraintWrapper) bool, allWrapped map[string]*shaclimport.ShapeWrapper, visited map[string]bool) {
	sb.WriteString("**Constraints:**\n\n")
	for _, c := range constraints {
		typeName := shaclimport.SimplifyIRI(c.Type)
		sb.WriteString(fmt.Sprintf("- **%s**", typeName))

		displayData := c.Data
		var severityOverride string
		if soc, ok := c.Data.(*shacl.SeverityOverrideConstraint); ok {
			displayData = soc.Inner()
			severityOverride = fmt.Sprintf(" (Severity: %s)", shaclimport.SimplifyTerm(soc.Severity))
		}
		sb.WriteString(severityOverride + "\n")

		if sc, ok := displayData.(*shacl.SPARQLConstraint); ok {
			query := sc.Prefixes + sc.Select
			sb.WriteString("\n```sparql\n")
			sb.WriteString(query)
			sb.WriteString("\n```\n")
			if len(sc.Messages) > 0 {
				var msgs []string
				for _, msg := range sc.Messages {
					msgs = append(msgs, shaclimport.SimplifyTerm(msg))
				}
				sb.WriteString(fmt.Sprintf("  - Messages: `[%s]`\n", strings.Join(msgs, ", ")))
			}
		} else {
			renderConstraintDetails(sb, displayData, level+1, filter, allWrapped, visited)
		}
	}
	sb.WriteString("\n")
}

func renderConstraintDetails(sb *strings.Builder, c shacl.Constraint, level int, filter func(shaclimport.ConstraintWrapper) bool, allWrapped map[string]*shaclimport.ShapeWrapper, visited map[string]bool) {
	var nestedShapes []*shaclimport.ShapeWrapper
	switch con := c.(type) {
	case *shacl.AndConstraint:
		for _, sRef := range con.Shapes {
			if sw, ok := allWrapped[sRef.String()]; ok {
				nestedShapes = append(nestedShapes, sw)
			}
		}
	case *shacl.OrConstraint:
		for _, sRef := range con.Shapes {
			if sw, ok := allWrapped[sRef.String()]; ok {
				nestedShapes = append(nestedShapes, sw)
			}
		}
	case *shacl.XoneConstraint:
		for _, sRef := range con.Shapes {
			if sw, ok := allWrapped[sRef.String()]; ok {
				nestedShapes = append(nestedShapes, sw)
			}
		}
	case *shacl.NotConstraint:
		if sw, ok := allWrapped[con.ShapeRef.String()]; ok {
			nestedShapes = append(nestedShapes, sw)
		}
	case *shacl.NodeConstraint:
		if sw, ok := allWrapped[con.ShapeRef.String()]; ok {
			nestedShapes = append(nestedShapes, sw)
		}
	case *shacl.QualifiedValueShapeConstraint:
		if sw, ok := allWrapped[con.ShapeRef.String()]; ok {
			nestedShapes = append(nestedShapes, sw)
		}
	}

	if len(nestedShapes) > 0 {
		sb.WriteString("\n")
		var sub strings.Builder
		renderShapes(&sub, nestedShapes, level, filter, allWrapped, visited)

		lines := strings.Split(strings.TrimSpace(sub.String()), "\n")
		inCodeBlock := false
		for _, line := range lines {
			if strings.HasPrefix(line, "```") {
				inCodeBlock = !inCodeBlock
			}
			if line != "" {
				if inCodeBlock || strings.HasPrefix(line, "```") {
					sb.WriteString(line + "\n")
				} else {
					sb.WriteString("  " + line + "\n")
				}
			} else {
				sb.WriteString("\n")
			}
		}
		return
	}

	data, _ := json.Marshal(c)
	var m map[string]any
	json.Unmarshal(data, &m)
	var details []string
	for k, v := range m {
		if k != "Prefixes" {
			details = append(details, fmt.Sprintf("- %s: `%s` ", k, shaclimport.FormatValue(v)))
		}
	}
	sort.Strings(details)
	for _, d := range details {
		sb.WriteString("  " + d + "\n")
	}
}

func renderNestedProperties(sb *strings.Builder, sw *shaclimport.ShapeWrapper, level int, filter func(shaclimport.ConstraintWrapper) bool, allWrapped map[string]*shaclimport.ShapeWrapper, visited map[string]bool) {
	var filteredProperties []*shaclimport.ShapeWrapper
	for _, p := range sw.Properties {
		if shaclimport.HasContent(p, filter) {
			filteredProperties = append(filteredProperties, p)
		}
	}

	if len(filteredProperties) > 0 {
		sb.WriteString("**Nested Properties:**\n\n")
		renderShapes(sb, filteredProperties, level+1, filter, allWrapped, visited)
	}
}

func main() {
	flagJSON := flag.Bool("json", false, "Generate JSON output")
	flagMD := flag.Bool("md", false, "Generate Markdown output")
	shaclPattern := flag.String("shacl", shaclimport.DefaultSHACLPattern, "glob pattern for shacl files")
	outputDir := flag.String("out", "docs", "output directory for generated files")
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

	var allStats []shaclimport.FileStats
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

func processFile(file string, doJSON, doMD bool, outputDir string, allSHACLTypes, allSPARQLTypes map[string]bool) (shaclimport.FileStats, error) {
	g, err := shacl.LoadTurtleFile(file)
	if err != nil {
		return shaclimport.FileStats{}, err
	}

	shapes := shacl.ParseShapes(g)
	baseName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	stats := shaclimport.FileStats{
		Name:         baseName,
		ShaclCounts:  make(map[string]int),
		SparqlCounts: make(map[string]int),
	}

	isNestedMD := shaclimport.IdentifyNestedShapes(shapes)
	isNestedJSON := identifyPropertyNestedShapes(shapes)

	allWrapped := make(map[string]*shaclimport.ShapeWrapper)
	for k, s := range shapes {
		allWrapped[k] = shaclimport.WrapShape(s)
	}

	topLevelJSON := make(map[string]*shaclimport.ShapeWrapper)
	var topLevelMD []*shaclimport.ShapeWrapper
	for k, w := range allWrapped {
		if !isNestedMD[k] {
			topLevelMD = append(topLevelMD, w)
		}
		if !isNestedJSON[k] {
			topLevelJSON[k] = w
		}
		updateStats(&stats, w, allSHACLTypes, allSPARQLTypes)
	}

	if doJSON {
		if err := exportJSON(outputDir, baseName, topLevelJSON); err != nil {
			return stats, err
		}
	}

	if doMD {
		exportMD(&stats, outputDir, baseName, topLevelMD, allWrapped)
	}

	return stats, nil
}

func identifyPropertyNestedShapes(shapes map[string]*shacl.Shape) map[string]bool {
	isNested := make(map[string]bool)
	for _, s := range shapes {
		for _, ps := range s.Properties {
			isNested[ps.ID.String()] = true
		}
	}
	return isNested
}

func updateStats(stats *shaclimport.FileStats, w *shaclimport.ShapeWrapper, allSHACLTypes, allSPARQLTypes map[string]bool) {
	for _, c := range w.Constraints {
		typeName := shaclimport.SimplifyIRI(c.Type)
		if c.IsSPARQL() {
			stats.SparqlCounts[typeName]++
			allSPARQLTypes[typeName] = true
		} else {
			stats.ShaclCounts[typeName]++
			allSHACLTypes[typeName] = true
		}
	}
}

func exportJSON(outputDir, baseName string, wrapped map[string]*shaclimport.ShapeWrapper) error {
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

func exportMD(stats *shaclimport.FileStats, outputDir, baseName string, topLevel []*shaclimport.ShapeWrapper, allWrapped map[string]*shaclimport.ShapeWrapper) {
	shaclMD := generateMarkdown(baseName, topLevel, allWrapped, func(cw shaclimport.ConstraintWrapper) bool { return cw.IsSHACL() })
	if shaclMD != "" {
		mdOutDir := filepath.Join(outputDir, "SHACL")
		os.MkdirAll(mdOutDir, 0755)
		mdFile := filepath.Join(mdOutDir, baseName+".md")
		stats.ShaclPath = mdFile
		os.WriteFile(mdFile, []byte(shaclMD), 0644)
		fmt.Printf("Exported SHACL MD to %s\n", mdFile)
	}

	sparqlMD := generateMarkdown(baseName, topLevel, allWrapped, func(cw shaclimport.ConstraintWrapper) bool { return cw.IsSPARQL() })
	if sparqlMD != "" {
		mdOutDir := filepath.Join(outputDir, "SPARQL")
		os.MkdirAll(mdOutDir, 0755)
		mdFile := filepath.Join(mdOutDir, baseName+".md")
		stats.SparqlPath = mdFile
		os.WriteFile(mdFile, []byte(sparqlMD), 0644)
		fmt.Printf("Exported SPARQL MD to %s\n", mdFile)
	}
}

func writeOverview(outputDir string, allStats []shaclimport.FileStats, allSHACLTypes, allSPARQLTypes map[string]bool) {
	writeOverviewFile(filepath.Join(outputDir, "SHACL-Overview.md"), "SHACL", allStats, allSHACLTypes, true)
	writeOverviewFile(filepath.Join(outputDir, "SPARQL-Overview.md"), "SPARQL", allStats, allSPARQLTypes, false)
}

func writeOverviewFile(path, title string, allStats []shaclimport.FileStats, typesMap map[string]bool, isSHACL bool) {
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

func renderOverviewRow(sb *strings.Builder, s shaclimport.FileStats, types []string, isSHACL bool, totals map[string]int) map[string]int {
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
