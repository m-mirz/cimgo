package main

import (
	"bytes"
	"cimgo/cimgostructs"
	"cimgo/cimprofiles"
	"cimgo/validation"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

func main() {
	var cfg validation.Config
	var profStr, silenceStr string
	var jsonOutput bool

	flag.StringVar(&profStr, "profile", "", "Comma-separated list of profiles to check (EQ, SSH, TP, DY, SC, SV, DL, GL, OP, EQBD). Default: all.")
	flag.StringVar(&silenceStr, "silence", "", "Comma-separated list of rule IDs to silence.")
	flag.BoolVar(&cfg.Solved, "solved", false, "Enable SolvedMAS checks.")
	flag.BoolVar(&cfg.NotSolved, "notsolved", true, "Enable NotSolvedMAS checks.")
	flag.BoolVar(&cfg.Common, "common", true, "Enable Common/AllProfiles rules.")
	flag.BoolVar(&cfg.Quality, "quality", false, "Enable CIMdesk-style modeling quality checks.")
	flag.BoolVar(&jsonOutput, "json", false, "Output results in JSON format.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: cimval [options] <xml-file1> [<xml-file2> ...]\n\nOptions:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	explicitFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { explicitFlags[f.Name] = true })

	if silenceStr != "" {
		cfg.SilencedRules = strings.Split(silenceStr, ",")
	}

	files := flag.Args()
	if len(files) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	dataset := cimgostructs.NewCIMElementList()
	profileDatasets := make(map[string]*cimgostructs.CIMElementList)
	eqbdBVIDs := make(map[string]struct{})
	for _, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", file, err)
			os.Exit(1)
		}
		isolated, err := cimprofiles.DecodeProfile(bytes.NewReader(b), nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error decoding %s: %v\n", file, err)
			os.Exit(1)
		}
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
		if err := cimprofiles.MergeInto(dataset, isolated); err != nil {
			fmt.Fprintf(os.Stderr, "Error merging %s: %v\n", file, err)
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
		fmt.Printf("Loaded %d elements from %d files\n", len(dataset.Elements), len(files))
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
			// Sort violations for stable output
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
