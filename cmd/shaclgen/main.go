// Command shaclgen generates Go validation functions directly from CGMES
// SHACL Turtle files. Each (class, attribute, constraint) tuple becomes one
// Check<...> function in the shaclgen package. Constraint shapes that can't
// be expressed with simple field access are skipped with a reason; there is
// no runtime SHACL evaluator behind shaclgen, so a skipped constraint is not
// validated at all. The skip-reason audit (see comment block below
// componentShort) tracks every unsupported shape against the live CGMES
// SHACL files so we can tell at a glance whether a skip is intentional
// (structurally satisfied by the Go type system) or genuinely unimplemented.
//
// Input is the SHACL TTL files matching `-shacl` (defaulting to
// shaclimport.DefaultSHACLPattern). Each file is parsed and simplified
// in-memory via shaclimport.ProcessFileToResults + SimplifyFileResults; no
// JSON intermediate is written or read.
//
// Output is written into a sibling package (default cimgo/shaclgen) so the
// generated code stays segregated from the hand-written validation package.
// The generated code uses shaclmodel.Violation for its return type. The
// generator depends only on shaclimport (parser) and cimstructs, so
// `go generate` can build it on a clean checkout even before any generated
// code exists in cimgo/shaclgen.
package main

import (
	"cimgo/shaclimport"
	"embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"text/template"
)

//go:embed all:templates
var templatesFS embed.FS

// fileCount is one row of the -rule-report "Per-File Rule Counts" table --
// the same (checks, skipped) totals cimoxide's --rule-report prints per
// file, in the same PERFILE\t<name>\t<checks>\t<skipped> format, so the two
// tools' outputs can be directly grepped/diffed against each other without
// an external script.
type fileCount struct {
	Name    string
	Checks  int
	Skipped int
}

func main() {
	shaclPattern := flag.String("shacl", shaclimport.DefaultSHACLPattern, "glob pattern for SHACL Turtle files")
	outDir := flag.String("out", "shaclgen", "output directory for generated Go files")
	pkg := flag.String("pkg", "shaclgen", "package name in generated files")
	skipReport := flag.Bool("skip-report", false, "instead of writing files, print every skip reason to stderr")
	ruleReport := flag.Bool("rule-report", false, "print hand-written + generated rule counts for README.md")
	validationDir := flag.String("validation-dir", "validation", "directory of hand-written SPARQL validation Go source, used by -rule-report")
	flag.Parse()

	matches, err := filepath.Glob(*shaclPattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "shacl glob %q: %v\n", *shaclPattern, err)
		os.Exit(1)
	}
	if len(matches) == 0 {
		fmt.Fprintf(os.Stderr, "shacl pattern %q matched no files\n", *shaclPattern)
		os.Exit(1)
	}
	sort.Strings(matches)

	tmpl, err := loadTemplate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "template: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}

	// Clean up existing generated files to ensure stale profiles are removed
	existing, _ := filepath.Glob(filepath.Join(*outDir, "generated_*.go"))
	for _, f := range existing {
		os.Remove(f)
	}

	var orchestrators []string
	totalChecks, totalSkipped, totalFiles := 0, 0, 0
	globalCounts := map[string]int{}
	groupChecks := map[string]int{}
	groupSkipped := map[string]int{}
	var perFile []fileCount
	for _, src := range matches {
		fr, drops, err := loadFromTTL(src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", src, err)
			os.Exit(1)
		}

		spec, skipReasons := buildFileSpec(*pkg, fr)
		skipReasons = append(dropsToSkipEntries(drops), skipReasons...)

		fileCheckCount := uniqueCheckPatterns(spec.Checks)
		totalChecks += fileCheckCount
		totalSkipped += len(skipReasons)

		group := ttlGroupLabel(filepath.Base(src))
		groupChecks[group] += fileCheckCount
		groupSkipped[group] += len(skipReasons)
		perFile = append(perFile, fileCount{Name: spec.FileName, Checks: fileCheckCount, Skipped: len(skipReasons)})

		if *skipReport {
			fmt.Fprintf(os.Stderr, "PERFILE\t%s\t%d\t%d\n", spec.FileName, fileCheckCount, len(skipReasons))
			for _, r := range skipReasons {
				fmt.Fprintf(os.Stderr, "%s\t%s\n", spec.FileName, r)
			}
			printFileSummary(os.Stderr, spec.FileName, fileCheckCount, skipReasons)
		}
		accumulateCounts(globalCounts, skipReasons)
		fmt.Printf("Generated %s (%d checks, %d skipped)\n", spec.FileName, fileCheckCount, len(skipReasons))

		if len(spec.Checks) == 0 {
			continue
		}

		err = writeGeneratedFile(tmpl, spec, *outDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", src, err)
			os.Exit(1)
		}

		orchestrators = append(orchestrators, spec.OrchestratorName)
		totalFiles++
	}

	if err := writeIndex(*outDir, *pkg, orchestrators); err != nil {
		fmt.Fprintf(os.Stderr, "index: %v\n", err)
		os.Exit(1)
	}
	if !*ruleReport {
		// Under -rule-report, this exact total (and more, broken down by profile) is
		// already in the "Generated SHACL Rules by Profile" table below.
		fmt.Printf("Total: %d files, %d checks, %d skipped\n", totalFiles, totalChecks, totalSkipped)
	}
	if *skipReport {
		printGlobalSummary(os.Stderr, globalCounts)
	}

	if *ruleReport {
		fmt.Fprintln(os.Stderr, "\n########## README rule-count report ##########")
		printGlobalSummary(os.Stderr, globalCounts)

		fmt.Fprintln(os.Stderr, "\n=== Generated SHACL Rules by Profile ===")
		fmt.Fprintf(os.Stderr, "  %-32s %10s %9s %8s\n", "Profile Group", "Generated", "Skipped", "Total")
		genTotal, skipTotal := 0, 0
		for _, g := range ttlGroupLabelOrder {
			checks, skipped := groupChecks[g], groupSkipped[g]
			genTotal += checks
			skipTotal += skipped
			fmt.Fprintf(os.Stderr, "  %-32s %10d %9d %8d\n", g, checks, skipped, checks+skipped)
		}
		fmt.Fprintln(os.Stderr, "  -----")
		fmt.Fprintf(os.Stderr, "  %-32s %10d %9d %8d\n", "Total", genTotal, skipTotal, genTotal+skipTotal)

		// Per-file breakdown, for diffing directly against cimoxide's --rule-report
		// output (same PERFILE\t<name>\t<checks>\t<skipped>\t<total> line format on
		// both sides): `grep PERFILE cimgo.log | sort > a; grep PERFILE cimoxide.log
		// | sort > b; diff a b` finds every field-level difference, or to compare
		// just the per-file Total (the meaningful cross-tool check -- Generated vs
		// Skipped legitimately differs by codegen capability even when Total
		// agrees): `awk -F'\t' '{print $2, $5}' a | diff - <(awk -F'\t' '{print $2,
		// $5}' b)`. No external script needed either way.
		fmt.Fprintln(os.Stderr, "\n=== Per-File Rule Counts (grep PERFILE to diff against cimoxide) ===")
		sort.Slice(perFile, func(i, j int) bool { return perFile[i].Name < perFile[j].Name })
		for _, f := range perFile {
			fmt.Fprintf(os.Stderr, "PERFILE\t%s\t%d\t%d\t%d\n", f.Name, f.Checks, f.Skipped, f.Checks+f.Skipped)
		}

		groups, err := sparqlReport(*validationDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "rule-report: %v\n", err)
			os.Exit(1)
		}
		ttl, err := ttlSparqlNames(*shaclPattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "rule-report: %v\n", err)
			os.Exit(1)
		}
		rows := combineCoverage(groups, ttl)

		fmt.Fprintf(os.Stderr, "\n=== SPARQL Check Coverage (%s vs %s) ===\n", *validationDir, *shaclPattern)
		fmt.Fprintf(os.Stderr, "  %-32s %12s %10s %9s\n", "Profile Group", "Implemented", "TTL Total", "Coverage")
		totalImpl, totalTTL := 0, 0
		for _, r := range rows {
			if r.TTLTotal < 0 {
				fmt.Fprintf(os.Stderr, "  %-32s %12d %10s %9s\n", r.Label, r.Implemented, "n/a", "n/a")
				continue
			}
			totalImpl += r.Implemented
			totalTTL += r.TTLTotal
			coverage := 100 * float64(r.Implemented) / float64(r.TTLTotal)
			fmt.Fprintf(os.Stderr, "  %-32s %12d %10d %8.1f%%\n", r.Label, r.Implemented, r.TTLTotal, coverage)
		}
		fmt.Fprintln(os.Stderr, "  -----")
		if totalTTL > 0 {
			fmt.Fprintf(os.Stderr, "  %-32s %12d %10d %8.1f%%\n", "Total", totalImpl, totalTTL, 100*float64(totalImpl)/float64(totalTTL))
		}

		for _, r := range rows {
			if len(r.Missing) == 0 {
				continue
			}
			fmt.Fprintf(os.Stderr, "\n  Not yet implemented in %s:\n", r.Label)
			for _, m := range r.Missing {
				fmt.Fprintf(os.Stderr, "    %s\n", m)
			}
		}
	}
}

// loadFromTTL parses one SHACL Turtle file and runs the simplify pipeline,
// keeping the result in memory.
func loadFromTTL(file string) (*shaclimport.FileResults, []shaclimport.SimplifiedDrop, error) {
	fr, err := shaclimport.ProcessFileToResults(file)
	if err != nil {
		return nil, nil, err
	}
	simplified, drops := shaclimport.SimplifyFileResultsWithDrops(fr)
	return simplified, drops, nil
}

// writeGeneratedFile writes the spec to a generated_*.go file.
func writeGeneratedFile(tmpl *template.Template, spec fileSpec, outDir string) error {
	stem := profileStem(spec.FileName)
	outPath := filepath.Join(outDir, "generated_"+stem+".go")
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	defer f.Close()
	if err := tmpl.Execute(f, spec); err != nil {
		return fmt.Errorf("execute: %w", err)
	}
	return nil
}

// writeIndex emits generated_index.go, which exposes a single
// ValidateAllGeneratedProfiles function chaining every per-file orchestrator.
// This is the top-level entry point; per-profile wiring into existing
// Validate*Profile functions in sparql_rules.go remains a separate decision.
func writeIndex(outDir, pkg string, orchestrators []string) error {
	sort.Strings(orchestrators)
	outPath := filepath.Join(outDir, "generated_index.go")
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "// Code generated by cmd/shaclgen. DO NOT EDIT.\n\n")
	fmt.Fprintf(f, "package %s\n\n", pkg)
	fmt.Fprintf(f, "import (\n\t\"cimgo/cimstructs\"\n\t\"cimgo/shaclmodel\"\n)\n\n")
	fmt.Fprintf(f, "// ValidateAllGeneratedProfiles runs every generated SHACL profile orchestrator.\n")
	fmt.Fprintf(f, "func ValidateAllGeneratedProfiles(dataset *cimstructs.CIMDataset) []shaclmodel.Violation {\n")
	if len(orchestrators) == 0 {
		fmt.Fprintf(f, "\treturn nil\n}\n")
		return nil
	}
	fmt.Fprintf(f, "\tvar violations []shaclmodel.Violation\n")
	for _, o := range orchestrators {
		fmt.Fprintf(f, "\tviolations = append(violations, %s(dataset)...)\n", o)
	}
	fmt.Fprintf(f, "\treturn violations\n}\n")
	return nil
}

func loadTemplate() (*template.Template, error) {
	data, err := templatesFS.ReadFile("templates/validation_file.tmpl")
	if err != nil {
		return nil, err
	}
	return template.New("validation_file").Parse(string(data))
}
