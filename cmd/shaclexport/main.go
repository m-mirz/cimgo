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

// generateMarkdown creates a Markdown string for the given shapes, filtering constraints based on the provided filter function
func generateMarkdown(title string, wrapped map[string]*ShapeWrapper, filter func(ConstraintWrapper) bool) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", title))

	var keys []string
	for k, w := range wrapped {
		if hasContent(w, filter) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	if len(keys) == 0 {
		return ""
	}

	var shapes []*ShapeWrapper
	for _, k := range keys {
		shapes = append(shapes, wrapped[k])
	}

	renderShapes(&sb, shapes, 2, filter)

	return sb.String()
}

// hasContent checks if the shape or any of its nested properties contain constraints that match the filter
func hasContent(sw *ShapeWrapper, filter func(ConstraintWrapper) bool) bool {
	for _, c := range sw.Constraints {
		if filter(c) {
			return true
		}
	}
	for _, p := range sw.Properties {
		if hasContent(p, filter) {
			return true
		}
	}
	return false
}

// logicKey generates a unique key for a shape based on its constraints and properties, used for grouping similar shapes together
func logicKey(sw *ShapeWrapper, includeIdentity bool, filter func(ConstraintWrapper) bool) string {
	type logic struct {
		ID          string
		Path        string
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

	if includeIdentity {
		l.ID = simplifyTerm(sw.ID)
		if sw.Path != nil {
			l.Path = formatPath(sw.Path)
		}
		for _, d := range sw.Description {
			l.Description = append(l.Description, d.String())
		}
		sort.Strings(l.Description)
	}

	for _, m := range sw.Messages {
		l.Messages = append(l.Messages, m.String())
	}
	sort.Strings(l.Messages)
	for _, ip := range sw.IgnoredProperties {
		l.Ignored = append(l.Ignored, ip.Value())
	}
	sort.Strings(l.Ignored)

	for _, c := range sw.Constraints {
		if filter(c) {
			l.Constraints = append(l.Constraints, c)
		}
	}

	for _, p := range sw.Properties {
		if hasContent(p, filter) {
			l.Properties = append(l.Properties, logicKey(p, includeIdentity, filter))
		}
	}
	sort.Strings(l.Properties)

	data, _ := json.Marshal(l)
	return string(data)
}

func renderShapes(sb *strings.Builder, shapes []*ShapeWrapper, level int, filter func(ConstraintWrapper) bool) {
	if len(shapes) == 0 {
		return
	}

	groups := groupShapes(shapes, level == 2, filter)
	for _, key := range groups.keys {
		group := groups.m[key]
		first := group[0]

		renderShapeHeading(sb, group, level)
		renderShapeBasicInfo(sb, first)
		renderShapeTargets(sb, group)

		var sparqlQueries []sparqlInfo
		sparqlQueries = collectSPARQLValues(sb, first, sparqlQueries)

		filteredConstraints := filterConstraints(first, filter)
		if len(filteredConstraints) > 0 {
			sparqlQueries = renderConstraintsTable(sb, filteredConstraints, sparqlQueries)
		}

		renderSPARQLQueries(sb, sparqlQueries)
		renderNestedProperties(sb, first, level, filter)
	}
}

type shapeGroups struct {
	m    map[string][]*ShapeWrapper
	keys []string
}

func groupShapes(shapes []*ShapeWrapper, includeIdentity bool, filter func(ConstraintWrapper) bool) shapeGroups {
	groups := make(map[string][]*ShapeWrapper)
	var keys []string
	for _, s := range shapes {
		key := logicKey(s, includeIdentity, filter)
		if _, ok := groups[key]; !ok {
			keys = append(keys, key)
		}
		groups[key] = append(groups[key], s)
	}
	sort.Strings(keys)
	return shapeGroups{m: groups, keys: keys}
}

func renderShapeHeading(sb *strings.Builder, group []*ShapeWrapper, level int) {
	var titles []string
	seenTitle := make(map[string]bool)
	for _, s := range group {
		title := simplifyTerm(s.ID)
		if s.IsProperty && s.Path != nil {
			title = fmt.Sprintf("`%s`", formatPath(s.Path))
		}
		if !seenTitle[title] {
			titles = append(titles, title)
			seenTitle[title] = true
		}
	}
	sort.Strings(titles)

	heading := strings.Join(titles, ", ")
	if level > 2 && len(group) > 1 {
		if len(titles) <= 3 {
			heading = fmt.Sprintf("Property Group (%d): %s", len(group), heading)
		} else {
			heading = fmt.Sprintf("Property Group (%d)", len(group))
		}
	}
	sb.WriteString(fmt.Sprintf("%s %s\n\n", strings.Repeat("#", level), heading))

	if level > 2 && len(group) > 1 {
		sb.WriteString("**Properties:**\n")
		for _, t := range titles {
			sb.WriteString(fmt.Sprintf("- %s\n", t))
		}
		sb.WriteString("\n")
	}
}

func renderShapeBasicInfo(sb *strings.Builder, first *ShapeWrapper) {
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
}

func renderShapeTargets(sb *strings.Builder, group []*ShapeWrapper) {
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
}

type sparqlInfo struct {
	id    string
	query string
}

func collectSPARQLValues(sb *strings.Builder, sw *ShapeWrapper, queries []sparqlInfo) []sparqlInfo {
	if sw.Values == nil {
		return queries
	}
	query := sw.Values.Prefixes + sw.Values.Select
	if sw.Values.Expr != "" {
		query = sw.Values.Prefixes + "SELECT (" + sw.Values.Expr + " AS ?value) WHERE { $this ?p ?o }"
	}
	queries = append(queries, sparqlInfo{id: "Values", query: query})
	sb.WriteString("**SPARQL Values:** [See below](#sparql-values)\n\n")
	return queries
}

func filterConstraints(sw *ShapeWrapper, filter func(ConstraintWrapper) bool) []ConstraintWrapper {
	var filtered []ConstraintWrapper
	for _, c := range sw.Constraints {
		if filter(c) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func renderConstraintsTable(sb *strings.Builder, constraints []ConstraintWrapper, queries []sparqlInfo) []sparqlInfo {
	sb.WriteString("**Constraints:**\n\n")
	sb.WriteString("| Component | Details |\n")
	sb.WriteString("| --- | --- |\n")
	for i, c := range constraints {
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

		if sc, ok := displayData.(*shacl.SPARQLConstraint); ok {
			id := fmt.Sprintf("SPARQL-%d", i+1)
			queries = append(queries, sparqlInfo{id: id, query: sc.Prefixes + sc.Select})
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
				if k != "Prefixes" {
					details = append(details, fmt.Sprintf("%s: `%s` ", k, formatValue(v)))
				}
			}
		}
		sort.Strings(details)
		sb.WriteString(fmt.Sprintf("| %s | %s%s |\n", typeName, strings.Join(details, "<br>"), severityOverride))
	}
	sb.WriteString("\n")
	return queries
}

func renderSPARQLQueries(sb *strings.Builder, queries []sparqlInfo) {
	if len(queries) > 0 {
		sb.WriteString("#### SPARQL Queries\n\n")
		for _, sq := range queries {
			sb.WriteString(fmt.Sprintf("##### %s\n```sparql\n%s\n```\n\n", sq.id, sq.query))
		}
	}
}

func renderNestedProperties(sb *strings.Builder, sw *ShapeWrapper, level int, filter func(ConstraintWrapper) bool) {
	var filteredProperties []*ShapeWrapper
	for _, p := range sw.Properties {
		if hasContent(p, filter) {
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

func processFile(file string, doJSON, doMD bool, outputDir string, allSHACLTypes, allSPARQLTypes map[string]bool) (fileStats, error) {
	g, err := shacl.LoadTurtleFile(file)
	if err != nil {
		return fileStats{}, err
	}

	shapes := shacl.ParseShapes(g)
	wrapped := make(map[string]*ShapeWrapper)
	baseName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	stats := fileStats{
		name:         baseName,
		shaclCounts:  make(map[string]int),
		sparqlCounts: make(map[string]int),
	}

	isNested := identifyNestedShapes(shapes)

	for k, s := range shapes {
		w := wrapShape(s)
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

func updateStats(stats *fileStats, w *ShapeWrapper, allSHACLTypes, allSPARQLTypes map[string]bool) {
	for _, c := range w.Constraints {
		typeName := simplifyIRI(c.Type)
		if c.IsSPARQL() {
			stats.sparqlCounts[typeName]++
			allSPARQLTypes[typeName] = true
		} else {
			stats.shaclCounts[typeName]++
			allSHACLTypes[typeName] = true
		}
	}
}

func exportJSON(outputDir, baseName string, wrapped map[string]*ShapeWrapper) error {
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

func exportMD(stats *fileStats, outputDir, baseName string, wrapped map[string]*ShapeWrapper) {
	shaclMD := generateMarkdown(baseName, wrapped, func(cw ConstraintWrapper) bool { return cw.IsSHACL() })
	if shaclMD != "" {
		mdOutDir := filepath.Join(outputDir, "SHACL")
		os.MkdirAll(mdOutDir, 0755)
		mdFile := filepath.Join(mdOutDir, baseName+".md")
		stats.shaclPath = mdFile
		os.WriteFile(mdFile, []byte(shaclMD), 0644)
		fmt.Printf("Exported SHACL MD to %s\n", mdFile)
	}

	sparqlMD := generateMarkdown(baseName, wrapped, func(cw ConstraintWrapper) bool { return cw.IsSPARQL() })
	if sparqlMD != "" {
		mdOutDir := filepath.Join(outputDir, "SPARQL")
		os.MkdirAll(mdOutDir, 0755)
		mdFile := filepath.Join(mdOutDir, baseName+".md")
		stats.sparqlPath = mdFile
		os.WriteFile(mdFile, []byte(sparqlMD), 0644)
		fmt.Printf("Exported SPARQL MD to %s\n", mdFile)
	}
}

func writeOverview(outputDir string, allStats []fileStats, allSHACLTypes, allSPARQLTypes map[string]bool) {
	writeOverviewFile(filepath.Join(outputDir, "SHACL-Overview.md"), "SHACL", allStats, allSHACLTypes, true)
	writeOverviewFile(filepath.Join(outputDir, "SPARQL-Overview.md"), "SPARQL", allStats, allSPARQLTypes, false)
}

func writeOverviewFile(path, title string, allStats []fileStats, typesMap map[string]bool, isSHACL bool) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s Constraints Overview\n\n", title))

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

func renderOverviewRow(sb *strings.Builder, s fileStats, types []string, isSHACL bool, totals map[string]int) map[string]int {
	var filePath string
	var counts map[string]int
	if isSHACL {
		filePath, counts = s.shaclPath, s.shaclCounts
	} else {
		filePath, counts = s.sparqlPath, s.sparqlCounts
	}

	if len(counts) == 0 {
		return totals
	}

	fileName := s.name
	if filePath != "" {
		relPath := filepath.Base(filepath.Dir(filePath)) + "/" + filepath.Base(filePath)
		fileName = fmt.Sprintf("[%s](%s)", s.name, relPath)
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
