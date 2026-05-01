package main

import (
	"cimgo/validation"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	shaclPattern := flag.String("shacl", validation.DefaultSHACLPattern, "glob pattern for shacl files")
	outputDir := flag.String("out", "pages/docs/struct", "output directory for generated files")
	simplifiedDir := flag.String("out-simplified", "pages/docs/struct-simplified", "output directory for simplified generated files")
	flag.Parse()

	shaclFiles, err := filepath.Glob(*shaclPattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error globbing files: %v\n", err)
		return
	}

	for _, dir := range []string{*outputDir, *simplifiedDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
			return
		}
	}

	for _, file := range shaclFiles {
		if err := processFile(file, *outputDir, *simplifiedDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", file, err)
		}
	}
}

func processFile(file string, outputDir string, simplifiedDir string) error {
	results, err := validation.ProcessFileToResults(file)
	if err != nil {
		return err
	}
	if len(results.Classes) == 0 {
		return nil
	}

	if err := writeJSON(results, filepath.Join(outputDir, results.FileName+".json")); err != nil {
		return err
	}
	fmt.Printf("Exported struct to %s\n", filepath.Join(outputDir, results.FileName+".json"))

	simplified := validation.SimplifyFileResults(results)
	if err := writeJSON(simplified, filepath.Join(simplifiedDir, results.FileName+".json")); err != nil {
		return err
	}
	fmt.Printf("Exported simplified struct to %s\n", filepath.Join(simplifiedDir, results.FileName+".json"))

	return nil
}

func writeJSON(v any, path string) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
