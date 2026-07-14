package main

import (
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
func generateMarkdown(title string, shapes []shaclimport.ShapeInfo, filter func(shaclimport.ConstraintInfo) bool) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", title))

	if len(shapes) == 0 {
		return ""
	}

	renderShapes(&sb, shapes, 2, filter, make(map[string]bool))

	return sb.String()
}

func renderShapes(sb *strings.Builder, shapes []shaclimport.ShapeInfo, level int, filter func(shaclimport.ConstraintInfo) bool, visited map[string]bool) {
	if len(shapes) == 0 {
		return
	}

	// Sort shapes: properties first, then by path, then by ID
	sort.Slice(shapes, func(i, j int) bool {
		si, sj := shapes[i], shapes[j]
		if len(si.Path) > 0 && len(sj.Path) > 0 {
			return strings.Join(si.Path, " / ") < strings.Join(sj.Path, " / ")
		}
		return si.ID < sj.ID
	})

	for i, s := range shapes {
		if s.ID != "" && visited[s.ID] {
			renderShapeHeading(sb, s, level)
			sb.WriteString(fmt.Sprintf("*(Recursive reference to %s)*\n\n", s.ID))
			continue
		}
		if s.ID != "" {
			visited[s.ID] = true
		}

		if (s.ID == "" || strings.HasPrefix(s.ID, "_:")) && len(shapes) > 1 {
			sb.WriteString(fmt.Sprintf("**Item %d:**\n\n", i+1))
		} else {
			renderShapeHeading(sb, s, level)
		}

		renderShapeContent(sb, s, level, filter, visited)

		if s.ID != "" {
			delete(visited, s.ID)
		}
	}
}

func renderShapeContent(sb *strings.Builder, s shaclimport.ShapeInfo, level int, filter func(shaclimport.ConstraintInfo) bool, visited map[string]bool) {
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

	var filteredConstraints []shaclimport.ConstraintInfo
	for _, c := range s.Constraints {
		if filter(c) {
			filteredConstraints = append(filteredConstraints, c)
		}
	}

	if len(filteredConstraints) > 0 {
		renderConstraintsList(sb, filteredConstraints, level, visited)
	}

	if len(s.Properties) > 0 {
		var filteredProperties []shaclimport.ShapeInfo
		for _, p := range s.Properties {
			if hasContent(p, filter) {
				filteredProperties = append(filteredProperties, p)
			}
		}

		if len(filteredProperties) > 0 {
			sb.WriteString("**Nested Properties:**\n\n")
			renderShapes(sb, filteredProperties, level+1, filter, visited)
		}
	}
}

func hasContent(s shaclimport.ShapeInfo, filter func(shaclimport.ConstraintInfo) bool) bool {
	for _, c := range s.Constraints {
		if filter(c) {
			return true
		}
	}
	for _, p := range s.Properties {
		if hasContent(p, filter) {
			return true
		}
	}
	return false
}

func renderShapeHeading(sb *strings.Builder, s shaclimport.ShapeInfo, level int) {
	if s.ID == "" || strings.HasPrefix(s.ID, "_:") {
		return
	}
	sb.WriteString(fmt.Sprintf("%s %s\n\n", strings.Repeat("#", level), s.ID))
}

func renderShapeBasicInfo(sb *strings.Builder, s shaclimport.ShapeInfo) {
	if len(s.Path) > 0 {
		sb.WriteString(fmt.Sprintf("**Path:** `%s`  \n", strings.Join(s.Path, " / ")))
	}

	if s.Name != "" {
		sb.WriteString(fmt.Sprintf("**Name:** %s  \n", s.Name))
	}

	if s.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", s.Description))
	}

	if s.Severity != "" && s.Severity != "Violation" {
		sb.WriteString(fmt.Sprintf("**Severity:** %s\n\n", s.Severity))
	}

	if len(s.Messages) > 0 {
		sb.WriteString("**Messages:**\n")
		for _, m := range s.Messages {
			sb.WriteString(fmt.Sprintf("- %s\n", m))
		}
		sb.WriteString("\n")
	}
}

func renderShapeTargets(sb *strings.Builder, s shaclimport.ShapeInfo) {
	if len(s.Targets) > 0 {
		sb.WriteString("**Targets:**\n")
		for _, t := range s.Targets {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", t.Kind, t.Value))
		}
		sb.WriteString("\n")
	}
}

func renderConstraintsList(sb *strings.Builder, constraints []shaclimport.ConstraintInfo, level int, visited map[string]bool) {
	sb.WriteString("**Constraints:**\n\n")
	for _, c := range constraints {
		sb.WriteString(fmt.Sprintf("- **%s**", c.Component))

		if c.Severity != "" {
			sb.WriteString(fmt.Sprintf(" (Severity: %s)", c.Severity))
		}
		sb.WriteString("\n")

		if c.IsSPARQL() {
			query, _ := c.Payload["Select"].(string)
			prefixes, _ := c.Payload["Prefixes"].(string)
			sb.WriteString("\n```sparql\n")
			sb.WriteString(prefixes + query)
			sb.WriteString("\n```\n")
			if c.Message != "" {
				sb.WriteString(fmt.Sprintf("  - Messages: `[%s]`\n", c.Message))
			}
		} else {
			renderConstraintDetails(sb, c, level+1, visited)
		}
	}
	sb.WriteString("\n")
}

func renderConstraintDetails(sb *strings.Builder, c shaclimport.ConstraintInfo, level int, visited map[string]bool) {
	var nestedLists [][]shaclimport.ConstraintInfo
	var details []string

	keys := make([]string, 0, len(c.Payload))
	for k := range c.Payload {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := c.Payload[k]
		if list, ok := v.([]shaclimport.ConstraintInfo); ok {
			nestedLists = append(nestedLists, list)
		} else if list, ok := v.([]any); ok {
			if len(list) > 0 {
				if _, isSubList := list[0].([]shaclimport.ConstraintInfo); isSubList {
					// List-of-alternatives (e.g. sh:or Shapes): compact if all are single sh:class
					if classes := classNamesFromAlternatives(list); classes != nil {
						details = append(details, fmt.Sprintf("- %s: %s ", k, strings.Join(classes, ", ")))
					} else {
						for _, item := range list {
							if subList, ok := item.([]shaclimport.ConstraintInfo); ok {
								nestedLists = append(nestedLists, subList)
							}
						}
					}
				} else {
					// Try to convert []any to flat []ConstraintInfo
					var ciList []shaclimport.ConstraintInfo
					allCI := true
					for _, item := range list {
						if ci, ok := item.(shaclimport.ConstraintInfo); ok {
							ciList = append(ciList, ci)
						} else if m, ok := item.(map[string]any); ok {
							data, _ := json.Marshal(m)
							var ci shaclimport.ConstraintInfo
							if err := json.Unmarshal(data, &ci); err == nil && ci.Component != "" {
								ciList = append(ciList, ci)
							} else {
								allCI = false; break
							}
						} else {
							allCI = false; break
						}
					}
					if allCI && len(ciList) > 0 {
						nestedLists = append(nestedLists, ciList)
					} else if !allCI && k != "Prefixes" && k != "Select" {
						details = append(details, fmt.Sprintf("- %s: `%v` ", k, v))
					}
				}
			}
		} else if ci, ok := v.(shaclimport.ConstraintInfo); ok {
			nestedLists = append(nestedLists, []shaclimport.ConstraintInfo{ci})
		} else if k != "Prefixes" && k != "Select" {
			details = append(details, fmt.Sprintf("- %s: `%v` ", k, v))
		}
	}

	for _, d := range details {
		sb.WriteString("  " + d + "\n")
	}

	if len(nestedLists) > 0 {
		sb.WriteString("\n")
		var sub strings.Builder
		for i, list := range nestedLists {
			if len(nestedLists) > 1 {
				sub.WriteString(fmt.Sprintf("**Item %d:**\n\n", i+1))
			}
			sub.WriteString("**Constraints:**\n\n")
			for _, ci := range list {
				sub.WriteString(fmt.Sprintf("- **%s**\n", ci.Component))
				renderConstraintDetails(&sub, ci, level+1, visited)
			}
		}

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
	}
}

// classNamesFromAlternatives returns a compact list of backtick-quoted class names if every
// alternative in a list-of-alternatives is a single sh:ClassConstraintComponent, or nil otherwise.
func classNamesFromAlternatives(list []any) []string {
	classes := make([]string, 0, len(list))
	for _, item := range list {
		subList, ok := item.([]shaclimport.ConstraintInfo)
		if !ok || len(subList) != 1 || subList[0].Component != "sh:ClassConstraintComponent" {
			return nil
		}
		cls, ok := subList[0].Payload["Class"].(string)
		if !ok || cls == "" {
			return nil
		}
		classes = append(classes, "`"+cls+"`")
	}
	return classes
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
	fr, err := shaclimport.ProcessFileToResults(file)
	if err != nil {
		return shaclimport.FileStats{}, err
	}

	baseName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	stats := shaclimport.FileStats{
		Name:         baseName,
		ShaclCounts:  make(map[string]int),
		SparqlCounts: make(map[string]int),
	}

	updateStats(&stats, fr.Shapes, allSHACLTypes, allSPARQLTypes)

	if doJSON {
		if err := exportJSON(outputDir, baseName, fr.Shapes); err != nil {
			return stats, err
		}
	}

	if doMD {
		exportMD(&stats, outputDir, baseName, fr.Shapes)
	}

	return stats, nil
}

func updateStats(stats *shaclimport.FileStats, shapes []shaclimport.ShapeInfo, allSHACLTypes, allSPARQLTypes map[string]bool) {
	updateStatsDedup(stats, shapes, allSHACLTypes, allSPARQLTypes, make(map[string]bool))
}

func updateStatsDedup(stats *shaclimport.FileStats, shapes []shaclimport.ShapeInfo, allSHACLTypes, allSPARQLTypes, seen map[string]bool) {
	for _, s := range shapes {
		for _, c := range s.Constraints {
			if c.IsSPARQL() {
				sel, _ := c.Payload["Select"].(string)
				key := s.ID + "::" + sel
				if seen[key] {
					continue
				}
				seen[key] = true
				stats.SparqlCounts[c.Component]++
				allSPARQLTypes[c.Component] = true
			} else {
				key := s.ID + "::" + c.Component
				if seen[key] {
					continue
				}
				seen[key] = true
				stats.ShaclCounts[c.Component]++
				allSHACLTypes[c.Component] = true
			}
		}
		updateStatsDedup(stats, s.Properties, allSHACLTypes, allSPARQLTypes, seen)
	}
}

func exportJSON(outputDir, baseName string, shapes []shaclimport.ShapeInfo) error {
	jsonOutDir := filepath.Join(outputDir, "json")
	if err := os.MkdirAll(jsonOutDir, 0755); err != nil {
		return err
	}
	data, _ := json.MarshalIndent(shapes, "", "  ")
	jsonFile := filepath.Join(jsonOutDir, baseName+".json")
	if err := os.WriteFile(jsonFile, data, 0644); err != nil {
		return err
	}
	fmt.Printf("Exported JSON to %s\n", jsonFile)
	return nil
}

func exportMD(stats *shaclimport.FileStats, outputDir, baseName string, shapes []shaclimport.ShapeInfo) {
	shaclMD := generateMarkdown(baseName, shapes, func(c shaclimport.ConstraintInfo) bool { return c.IsSHACL() })
	if shaclMD != "" {
		mdOutDir := filepath.Join(outputDir, "SHACL")
		os.MkdirAll(mdOutDir, 0755)
		mdFile := filepath.Join(mdOutDir, baseName+".md")
		stats.ShaclPath = mdFile
		os.WriteFile(mdFile, []byte(shaclMD), 0644)
		fmt.Printf("Exported SHACL MD to %s\n", mdFile)
	}

	sparqlMD := generateMarkdown(baseName, shapes, func(c shaclimport.ConstraintInfo) bool { return c.IsSPARQL() })
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

	var types []string
	for k := range typesMap {
		types = append(types, k)
	}
	sort.Strings(types)

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
