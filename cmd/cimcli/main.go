package main

import (
	"bytes"
	"cimgo/cgmesxml"
	"cimgo/cimconv"
	"cimgo/cimstructs"
	"cimgo/validation"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"google.golang.org/protobuf/proto"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Fprintf(os.Stderr, "Usage: cimcli <command> [options]\n\nCommands:\n  validate  Validate CGMES XML files against SHACL and SPARQL rules\n  convert   Convert CGMES XML files to JSON, proto, or back to XML\n  import    Parse CGMES XML files and print element counts\n")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "validate":
		runValidate(os.Args[2:])
	case "convert":
		runConvert(os.Args[2:])
	case "import":
		runImport(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Usage: cimcli <command> [options]\n\nCommands:\n  validate  Validate CGMES XML files against SHACL and SPARQL rules\n  convert   Convert CGMES XML files to JSON, proto, or back to XML\n  import    Parse CGMES XML files and print element counts\n\nunknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func runValidate(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)

	var cfg validation.Config
	var profStr, silenceStr string
	var jsonOutput bool

	fs.StringVar(&profStr, "profile", "", "Comma-separated list of profiles to check (EQ, SSH, TP, DY, SC, SV, DL, GL, OP, EQBD). Default: all.")
	fs.StringVar(&silenceStr, "silence", "", "Comma-separated list of rule IDs to silence.")
	fs.BoolVar(&cfg.Solved, "solved", false, "Enable SolvedMAS checks.")
	fs.BoolVar(&cfg.NotSolved, "notsolved", true, "Enable NotSolvedMAS checks.")
	fs.BoolVar(&cfg.Common, "common", true, "Enable Common/AllProfiles rules.")
	fs.BoolVar(&cfg.Quality, "quality", false, "Enable CIMdesk-style modeling quality checks.")
	fs.BoolVar(&jsonOutput, "json", false, "Output results in JSON format.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: cimcli validate [options] <xml-file1> [<xml-file2> ...]\n\nOptions:\n")
		fs.PrintDefaults()
	}

	fs.Parse(args)

	explicitFlags := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) { explicitFlags[f.Name] = true })

	cfg.SilencedRules = []string{
		"dl:DiagramObject.IdentifiedObject-valueType",
		"sv:SvStatus.ConductingEquipment-valueType",
	}
	if silenceStr != "" {
		cfg.SilencedRules = append(cfg.SilencedRules, strings.Split(silenceStr, ",")...)
	}

	files := fs.Args()
	if len(files) == 0 {
		fs.Usage()
		os.Exit(1)
	}

	readers := make([]io.Reader, len(files))
	for i, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", file, err)
			os.Exit(1)
		}
		readers[i] = bytes.NewReader(b)
	}
	isolatedDatasets, err := cgmesxml.DecodeProfilesSeparate(readers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding files: %v\n", err)
		os.Exit(1)
	}

	dataset := cimstructs.NewCIMDataset()
	profileDatasets := make(map[string]*cimstructs.CIMDataset)
	eqbdBVIDs := make(map[string]struct{})
	for i, isolated := range isolatedDatasets {
		dc := validation.DetectConfig(isolated)
		if len(dc.Profiles) == 1 {
			name := dc.Profiles[0]
			profileDatasets[name] = isolated
			if name == "EQBD" {
				for id := range isolated.BaseVoltages {
					eqbdBVIDs[id] = struct{}{}
				}
			}
		}
		if err := cgmesxml.MergeInto(dataset, isolated); err != nil {
			fmt.Fprintf(os.Stderr, "Error merging %s: %v\n", files[i], err)
			os.Exit(1)
		}
	}
	cfg.EQBDBaseVoltageIDs = eqbdBVIDs
	cfg.PerProfileDatasets = profileDatasets

	detected := validation.DetectConfig(dataset)
	if profStr != "" {
		cfg.Profiles = strings.Split(strings.ToUpper(profStr), ",")
	} else {
		cfg.Profiles = detected.Profiles
	}
	if !explicitFlags["solved"] && !explicitFlags["notsolved"] {
		cfg.Solved = detected.Solved
		cfg.NotSolved = detected.NotSolved
	}

	if !jsonOutput {
		fmt.Printf("Loaded %d elements from %d files\n", len(dataset.ByID), len(files))
		fmt.Println("Running validation...")
	}

	violations := validation.RunValidation(dataset, cfg)

	if jsonOutput {
		data, _ := json.MarshalIndent(violations, "", "  ")
		fmt.Println(string(data))
	} else {
		if len(violations) == 0 {
			fmt.Println("No violations found.")
		} else {
			fmt.Printf("Found %d violations:\n\n", len(violations))
			sort.Slice(violations, func(i, j int) bool {
				if violations[i].ObjectID != violations[j].ObjectID {
					return violations[i].ObjectID < violations[j].ObjectID
				}
				return violations[i].RuleID < violations[j].RuleID
			})

			for _, v := range violations {
				fmt.Printf("[%s] Node: %s | Rule: %s\n", v.Severity, v.ObjectID, v.RuleID)
				if v.Name != "" {
					fmt.Printf("    Name:     %s\n", v.Name)
				}
				fmt.Printf("    Message:  %s\n", v.Message)
				if v.Property != "" {
					fmt.Printf("    Property: %s\n", v.Property)
				}
				if v.Description != "" {
					fmt.Printf("    Info:     %s\n", v.Description)
				}
				fmt.Println()
			}
		}
	}

	hasViolations := false
	for _, v := range violations {
		if v.Severity == "sh:Violation" {
			hasViolations = true
			break
		}
	}

	if hasViolations {
		os.Exit(1)
	}
}

func runImport(args []string) {
	fs := flag.NewFlagSet("import", flag.ExitOnError)
	var jsonOutput bool
	fs.BoolVar(&jsonOutput, "json", false, "Output results in JSON format.")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: cimcli import [options] <xml-file1> [<xml-file2> ...]\n\nOptions:\n")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	files := fs.Args()
	if len(files) == 0 {
		fs.Usage()
		os.Exit(1)
	}

	readers := make([]io.Reader, len(files))
	for i, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", file, err)
			os.Exit(1)
		}
		readers[i] = bytes.NewReader(b)
	}
	isolatedDatasets, err := cgmesxml.DecodeProfilesSeparate(readers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding files: %v\n", err)
		os.Exit(1)
	}

	dataset := cimstructs.NewCIMDataset()
	perFile := make([]map[string]interface{}, 0, len(files))
	for i, isolated := range isolatedDatasets {
		perFile = append(perFile, map[string]interface{}{
			"file":  filepath.Base(files[i]),
			"count": len(isolated.ByID),
		})
		if err := cgmesxml.MergeInto(dataset, isolated); err != nil {
			fmt.Fprintf(os.Stderr, "Error merging %s: %v\n", files[i], err)
			os.Exit(1)
		}
	}

	typeCounts := make(map[string]int)
	for _, elem := range dataset.ByID {
		typeCounts[reflect.TypeOf(elem).Elem().Name()]++
	}

	if jsonOutput {
		result := map[string]interface{}{
			"total":      len(dataset.ByID),
			"files":      perFile,
			"type_counts": typeCounts,
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Total elements: %d (from %d file(s))\n\n", len(dataset.ByID), len(files))
		for _, f := range perFile {
			fmt.Printf("  %s: %d elements\n", f["file"], f["count"])
		}
		if len(typeCounts) > 0 {
			fmt.Println("\nBy type:")
			types := make([]string, 0, len(typeCounts))
			for t := range typeCounts {
				types = append(types, t)
			}
			sort.Strings(types)
			for _, t := range types {
				fmt.Printf("  %-45s %d\n", t, typeCounts[t])
			}
		}
	}
}

func runConvert(args []string) {
	fs := flag.NewFlagSet("convert", flag.ExitOnError)
	var toFmt, outPath, profileStr string
	fs.StringVar(&toFmt, "to", "json", "output format: json, proto, xml")
	fs.StringVar(&outPath, "out", "", "output file for json/proto (default: stdout / output.pb) or directory for xml (default: .)")
	fs.StringVar(&profileStr, "profile", "", "comma-separated profile codes for --to xml (default: all)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: cimcli convert [options] <file1> [<file2> ...]\n\nInput: one or more CGMES XML files, or a single .json file produced by --to json.\nOutput:\n  --to json   merged dataset as JSON to stdout, or --out file\n  --to proto  binary Protobuf written to --out file (default: output.pb)\n  --to xml    CGMES XML profiles written to --out directory (default: .)\n\nOptions:\n")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	files := fs.Args()
	if len(files) == 0 {
		fs.Usage()
		os.Exit(1)
	}

	fromJSON := strings.HasSuffix(strings.ToLower(files[0]), ".json")

	var dataset *cimstructs.CIMDataset

	if fromJSON {
		if len(files) != 1 {
			fmt.Fprintf(os.Stderr, "JSON input: provide exactly one .json file\n")
			os.Exit(1)
		}
		b, err := os.ReadFile(files[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", files[0], err)
			os.Exit(1)
		}
		dataset, err = unmarshalWithType(b)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing JSON %s: %v\n", files[0], err)
			os.Exit(1)
		}
	} else {
		dataset = cimstructs.NewCIMDataset()
		for _, file := range files {
			b, err := os.ReadFile(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", file, err)
				os.Exit(1)
			}
			isolated, err := cgmesxml.DecodeProfile(bytes.NewReader(b), nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error decoding %s: %v\n", file, err)
				os.Exit(1)
			}
			if err := cgmesxml.MergeInto(dataset, isolated); err != nil {
				fmt.Fprintf(os.Stderr, "Error merging %s: %v\n", file, err)
				os.Exit(1)
			}
		}
	}

	switch toFmt {
	case "json":
		data, err := marshalWithType(dataset)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshalling JSON: %v\n", err)
			os.Exit(1)
		}
		if outPath != "" {
			if err := os.WriteFile(outPath, data, 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", outPath, err)
				os.Exit(1)
			}
		} else {
			fmt.Println(string(data))
		}

	case "proto":
		protoList, err := cimconv.ToProto(dataset)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error converting to proto: %v\n", err)
			os.Exit(1)
		}
		data, err := proto.Marshal(protoList)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshalling proto: %v\n", err)
			os.Exit(1)
		}
		dest := outPath
		if dest == "" {
			dest = "output.pb"
		}
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", dest, err)
			os.Exit(1)
		}

	case "xml":
		outDir := outPath
		if outDir == "" {
			outDir = "."
		}
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output directory %s: %v\n", outDir, err)
			os.Exit(1)
		}
		codes := resolveProfileCodes(profileStr)
		for _, code := range codes {
			filePath := filepath.Join(outDir, code+".xml")
			f, err := os.Create(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating %s: %v\n", filePath, err)
				os.Exit(1)
			}
			encErr := cgmesxml.EncodeForProfile(f, dataset, code)
			f.Close()
			if encErr != nil {
				fmt.Fprintf(os.Stderr, "Error encoding %s: %v\n", code, encErr)
				os.Exit(1)
			}
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown --to format: %s (want json, proto, xml)\n", toFmt)
		os.Exit(1)
	}
}

var knownProfileCodes = []string{"EQ", "SSH", "TP", "SV", "DL", "DY", "GL", "OP", "SC", "EQBD", "FH"}

func resolveProfileCodes(profileStr string) []string {
	if profileStr != "" {
		return strings.Split(strings.ToUpper(profileStr), ",")
	}
	return knownProfileCodes
}

func marshalWithType(dataset *cimstructs.CIMDataset) ([]byte, error) {
	out := make(map[string]map[string]interface{}, len(dataset.ByID))
	for id, elem := range dataset.ByID {
		typeName := reflect.TypeOf(elem).Elem().Name()
		b, err := json.Marshal(elem)
		if err != nil {
			return nil, err
		}
		var m map[string]interface{}
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		m["_type"] = typeName
		out[id] = m
	}
	return json.MarshalIndent(out, "", "  ")
}

func unmarshalWithType(data []byte) (*cimstructs.CIMDataset, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	dataset := cimstructs.NewCIMDataset()
	for _, elemRaw := range raw {
		var typeHolder struct {
			Type string `json:"_type"`
		}
		if err := json.Unmarshal(elemRaw, &typeHolder); err != nil {
			return nil, err
		}
		factory, ok := cimstructs.StructMap[typeHolder.Type]
		if !ok {
			continue
		}
		instance := factory()
		if err := json.Unmarshal(elemRaw, instance); err != nil {
			return nil, fmt.Errorf("%s: %w", typeHolder.Type, err)
		}
		if err := dataset.AddElement(instance); err != nil {
			return nil, err
		}
	}
	return dataset, nil
}
