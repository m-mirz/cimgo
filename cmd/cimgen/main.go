package main

import (
	"cimgo/cimgen"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	var schemaPattern string
	var language string
	var cgmesVersion string
	var verbose bool

	flag.StringVar(&schemaPattern, "schema", cimgen.DefaultRDFSPattern, "glob pattern for CIM schema files")
	flag.StringVar(&language, "lang", "go", "output language (go, proto)")
	flag.StringVar(&cgmesVersion, "version", cimgen.CGMESVersion_3_0_0, "CGMES version")
	flag.BoolVar(&verbose, "v", false, "verbose logging")
	flag.Parse()

	logger := log.New(os.Stderr, "", 0)
	if verbose {
		logger.Printf("schema: %s", schemaPattern)
		logger.Printf("language: %s", language)
		logger.Printf("version: %s", cgmesVersion)
	}

	if err := run(logger, schemaPattern, language, cgmesVersion); err != nil {
		logger.Fatalf("Error: %v", err)
	}
}

func run(logger *log.Logger, schemaPattern, language, cgmesVersion string) error {
	logger.Println("Generate code for", language, "from schema files matching", schemaPattern)

	// create and populate specification
	cimSpec := cimgen.NewCIMSpecification()
	cimSpec.CGMESVersion = cgmesVersion
	if err := cimSpec.ImportCIMSchemaFiles(schemaPattern); err != nil {
		return fmt.Errorf("failed to import CIM schema files: %w", err)
	}

	outputDir := "cimgostructs"
	if language == "proto" {
		outputDir = "proto/definitions"
	}

	type generatorFunc func(spec *cimgen.CIMSpecification, outputDir string) error

	generators := map[string]generatorFunc{
		"go":    (*cimgen.CIMSpecification).GenerateGo,
		"proto": (*cimgen.CIMSpecification).GenerateProto,
	}

	generator, ok := generators[language]
	if !ok {
		return fmt.Errorf("unsupported language: %s", language)
	}

	if err := generator(cimSpec, outputDir); err != nil {
		return fmt.Errorf("failed to generate %s code: %w", language, err)
	}

	logger.Println("Generated source files in:", outputDir)
	return nil
}
