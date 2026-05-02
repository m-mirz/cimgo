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
	outputDir := flag.String("out", "shacljson/struct", "output directory for generated files")
	simplifiedDir := flag.String("out-simplified", "shacljson/struct-simplified", "output directory for simplified generated files")
	writeStruct := flag.Bool("struct", false, "write non-simplified struct output")
	flag.Parse()

	shaclFiles, err := filepath.Glob(*shaclPattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error globbing files: %v\n", err)
		return
	}

	dirs := []string{*simplifiedDir}
	if *writeStruct {
		dirs = append(dirs, *outputDir)
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
			return
		}
	}

	for _, file := range shaclFiles {
		if err := processFile(file, *outputDir, *simplifiedDir, *writeStruct); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", file, err)
		}
	}
}

func processFile(file string, outputDir string, simplifiedDir string, writeStruct bool) error {
	results, err := validation.ProcessFileToResults(file)
	if err != nil {
		return err
	}
	if len(results.Classes) == 0 {
		return nil
	}

	if writeStruct {
		if err := writeJSON(results, filepath.Join(outputDir, results.FileName+".json")); err != nil {
			return err
		}
		fmt.Printf("Exported struct to %s\n", filepath.Join(outputDir, results.FileName+".json"))
	}

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
