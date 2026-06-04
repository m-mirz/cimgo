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

func main() {
	shaclPattern := flag.String("shacl", shaclimport.DefaultSHACLPattern, "glob pattern for SHACL Turtle files")
	outDir := flag.String("out", "shaclgen", "output directory for generated Go files")
	pkg := flag.String("pkg", "shaclgen", "package name in generated files")
	skipReport := flag.Bool("skip-report", false, "instead of writing files, print every skip reason to stderr")
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
	for _, src := range matches {
		fr, err := loadFromTTL(src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", src, err)
			os.Exit(1)
		}

		spec, skipReasons := buildFileSpec(*pkg, fr)

		if len(spec.Checks) == 0 {
			totalSkipped += len(skipReasons)
			if *skipReport {
				for _, r := range skipReasons {
					fmt.Fprintf(os.Stderr, "%s\t%s\n", spec.FileName, r)
				}
				printFileSummary(os.Stderr, spec.FileName, 0, skipReasons)
				accumulateCounts(globalCounts, skipReasons)
			}
			continue
		}

		err = writeGeneratedFile(tmpl, spec, *outDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", src, err)
			os.Exit(1)
		}

		orchestrators = append(orchestrators, spec.OrchestratorName)
		totalChecks += len(spec.Checks)
		totalSkipped += len(skipReasons)
		totalFiles++

		if *skipReport {
			for _, r := range skipReasons {
				fmt.Fprintf(os.Stderr, "%s\t%s\n", spec.FileName, r)
			}
			printFileSummary(os.Stderr, spec.FileName, len(spec.Checks), skipReasons)
			accumulateCounts(globalCounts, skipReasons)
		}
		fmt.Printf("Generated %s (%d checks, %d skipped)\n", spec.FileName, len(spec.Checks), len(skipReasons))
	}

	if err := writeIndex(*outDir, *pkg, orchestrators); err != nil {
		fmt.Fprintf(os.Stderr, "index: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Total: %d files, %d checks, %d skipped\n", totalFiles, totalChecks, totalSkipped)
	if *skipReport {
		printGlobalSummary(os.Stderr, globalCounts)
	}
}

// loadFromTTL parses one SHACL Turtle file and runs the simplify pipeline,
// keeping the result in memory.
func loadFromTTL(file string) (*shaclimport.FileResults, error) {
	fr, err := shaclimport.ProcessFileToResults(file)
	if err != nil {
		return nil, err
	}
	return shaclimport.SimplifyFileResults(fr), nil
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
