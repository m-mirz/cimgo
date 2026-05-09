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
// generator depends only on shaclimport (parser) and cimgostructs, so
// `go generate` can build it on a clean checkout even before any generated
// code exists in cimgo/shaclgen.
package main

import (
	"cimgo/cimgostructs"
	"cimgo/shaclimport"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"unicode"
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
		}
		fmt.Printf("Generated %s (%d checks, %d skipped)\n", spec.FileName, len(spec.Checks), len(skipReasons))
	}

	if err := writeIndex(*outDir, *pkg, orchestrators); err != nil {
		fmt.Fprintf(os.Stderr, "index: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Total: %d files, %d checks, %d skipped\n", totalFiles, totalChecks, totalSkipped)
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
	fmt.Fprintf(f, "import (\n\t\"cimgo/cimgostructs\"\n\t\"cimgo/shaclmodel\"\n)\n\n")
	fmt.Fprintf(f, "// ValidateAllGeneratedProfiles runs every generated SHACL profile orchestrator.\n")
	fmt.Fprintf(f, "func ValidateAllGeneratedProfiles(dataset *cimgostructs.CIMElementList) []shaclmodel.Violation {\n")
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

// fileSpec is the data passed to validation_file.tmpl.
type fileSpec struct {
	FileName         string
	Pkg              string
	OrchestratorName string
	Imports          []string // sorted, deduplicated; always includes cimgo/cimgostructs
	Checks           []checkSpec
}

// checkSpec describes a single Check<...> function. Guard and Condition carry
// the only varying pieces of the function body; everything else is fixed
// scaffolding emitted by the template. Decl is an optional package-level
// declaration emitted directly above the function (used by Pattern checks to
// hoist the compiled regexp out of the per-call hot path). Prelude is an
// optional block emitted before the main per-element loop — used by inverse
// path checks to build an O(N) cross-reference index once per Check, instead
// of paying O(N²) by scanning inside the loop. NoV switches the type assertion
// from `v, ok := ...` to `_, ok := ...` for checks that don't need the
// instance value (typical for inverse-path checks that only consume `id`).
type checkSpec struct {
	Name         string
	ShapeID      string // Original SHACL Shape ID (e.g. eqc:ACLineSegment.length-length)
	Class        string
	Tag          string
	Component    string
	Property     string
	Message      string
	Severity     string
	Decl         string // optional package-level declaration emitted before the function
	Prelude      string // optional block emitted before the main loop (e.g. inverse-ref index)
	NoV          bool   // suppress the v binding when the loop body doesn't use it
	Guard        string // tab-indented, may span multiple lines; empty if none
	Condition    string // single expression; opens the violation block as `if <Condition> {`
	DatasetCheck bool   // emit a single dataset-level check (no per-element loop) when true
}

func resolveConcreteClasses(targets []shaclimport.TargetInfo) []string {
	var result []string
	for _, t := range targets {
		if t.Kind != "targetClass" && t.Kind != "targetImplicitClass" && t.Kind != "targetNode" {
			continue
		}
		structName, ok := simpleClassName(t.Value)
		if !ok {
			continue
		}
		if _, ok := cimgostructs.StructMap[structName]; ok {
			result = append(result, structName)
		} else {
			result = append(result, concreteSubclassesEmbedding(structName)...)
		}
	}
	sort.Strings(result)
	// deduplicate
	if len(result) < 2 {
		return result
	}
	j := 0
	for i := 1; i < len(result); i++ {
		if result[i] != result[j] {
			j++
			result[j] = result[i]
		}
	}
	return result[:j+1]
}

func buildFileSpec(pkg string, fr *shaclimport.FileResults) (fileSpec, []string) {
	stem := profileStem(fr.FileName)
	stemCamel := camelCaseFromStem(stem)
	spec := fileSpec{
		FileName:         fr.FileName,
		Pkg:              pkg,
		OrchestratorName: "ValidateGenerated" + stemCamel + "Profile",
	}
	var skipReasons []string
	used := map[string]int{}
	importSet := map[string]struct{}{
		"cimgo/cimgostructs": {},
		"cimgo/shaclmodel":   {},
	}

	var processShape func(shape shaclimport.ShapeInfo, currentClasses []string)
	processShape = func(shape shaclimport.ShapeInfo, currentClasses []string) {
		concreteNames := resolveConcreteClasses(shape.Targets)
		if len(concreteNames) > 0 {
			currentClasses = concreteNames
		}

		if len(currentClasses) > 0 {
			for _, concrete := range currentClasses {
				factory := cimgostructs.StructMap[concrete]
				structType := reflect.TypeOf(factory()).Elem()

				constraints := append([]shaclimport.ConstraintInfo(nil), shape.Constraints...)
				sortByNameAndSig(constraints, func(c shaclimport.ConstraintInfo) string { return c.Component })

				for _, c := range constraints {
					cs, imports, err := buildCheckSpec(stemCamel, concrete, shape.ID, structType, c, used)
					if err != nil {
						prop := ""
						if len(c.Path) > 0 {
							prop = "." + c.Path[0]
						}
						skipReasons = append(skipReasons, fmt.Sprintf("%s%s [%s]: %v", concrete, prop, c.Component, err))
						continue
					}
					for _, imp := range imports {
						importSet[imp] = struct{}{}
					}
					spec.Checks = append(spec.Checks, cs)
				}
			}
		}

		for _, prop := range shape.Properties {
			processShape(prop, currentClasses)
		}
	}

	for _, shape := range fr.Shapes {
		processShape(shape, nil)
	}

	spec.Imports = make([]string, 0, len(importSet))
	for imp := range importSet {
		spec.Imports = append(spec.Imports, imp)
	}
	sort.Strings(spec.Imports)
	return spec, skipReasons
}

func buildCheckSpec(stemCamel, structName, shapeID string, structType reflect.Type, c shaclimport.ConstraintInfo, used map[string]int) (checkSpec, []string, error) {
	// Detect inverse and multi-segment paths up front. Inverse paths
	// (`^cim.X.Y`) flip the constraint sense from "look at this object's
	// field" to "scan the dataset for objects pointing at this one". Some
	// multi-segment shapes ([ref, rdf.type]) are recognised as a class-of-
	// referenced-object check; everything else is currently skipped.
	if len(c.Path) == 0 {
		return checkSpec{}, nil, fmt.Errorf("empty path")
	}

	rawPath := c.Path[0]
	isInverse := strings.HasPrefix(rawPath, "^")
	if isInverse {
		rawPath = rawPath[1:]
	}

	// `^rdf.type` is a dataset-level cardinality check: "count of instances
	// whose rdf:type is the focus class". MinCount=N → at least N instances
	// must exist; MaxCount=N → at most N. Handled here before generic
	// inverse-path machinery, which would try to parse "rdf.type" as a
	// class.field and fail with "no Go struct rdf".
	if isInverse && len(c.Path) == 1 && rawPath == "rdf.type" {
		return buildDatasetCardinalityCheck(stemCamel, structName, shapeID, c, used)
	}

	// Classify multi-segment paths. The dominant shape is a forward chain
	// ending in `rdf.type` (657 of 669 multi-segment HasValue/In); we also
	// accept a forward chain *without* the trailing rdf.type for Or, where
	// the disjunction-of-Class shapes already encode the type assertion.
	forwardChainEndsRdfType := false
	forwardChainOnly := false
	if len(c.Path) > 1 && !isInverse {
		allForward := true
		for _, seg := range c.Path[1:] {
			if strings.HasPrefix(seg, "^") {
				allForward = false
				break
			}
		}
		if allForward {
			if c.Path[len(c.Path)-1] == "rdf.type" {
				forwardChainEndsRdfType = true
			} else {
				forwardChainOnly = true
			}
		}
	}

	tag, ok := stripCIMPrefix(rawPath)
	if !ok {
		tag = rawPath
	}

	// For inverse paths, the field lives on the *target* class encoded in
	// the path (e.g. `Terminal.ConductingEquipment` lives on Terminal),
	// not on `structName` (the class whose constraints we're processing).
	// Otherwise the field lives on structName.
	//
	// targetClasses is the list of concrete cimgostructs class names to
	// dispatch over. For a class that's directly in StructMap it's
	// {targetClass}; for an abstract base class (e.g. ExcitationSystemDynamics)
	// we discover its concrete subclasses by walking StructMap and emit a
	// switch over each.
	lookupType := structType
	targetClass := ""
	var targetClasses []string
	if isInverse {
		// Tag shape after stripCIMPrefix: "Terminal.ConductingEquipment".
		// First dot-separated segment is the target class name.
		parts := strings.SplitN(tag, ".", 2)
		if len(parts) != 2 {
			return checkSpec{}, nil, fmt.Errorf("inverse path %q has no class.field shape", tag)
		}
		targetClass = parts[0]
		if factory, ok := cimgostructs.StructMap[targetClass]; ok {
			lookupType = reflect.TypeOf(factory()).Elem()
			targetClasses = []string{targetClass}
		} else {
			subclasses := concreteSubclassesEmbedding(targetClass)
			if len(subclasses) == 0 {
				return checkSpec{}, nil, fmt.Errorf("inverse target class %q has no Go struct", targetClass)
			}
			// Use the first subclass's reflect.Type for field lookup —
			// the abstract field is reachable via embedded promotion.
			firstFactory := cimgostructs.StructMap[subclasses[0]]
			lookupType = reflect.TypeOf(firstFactory()).Elem()
			targetClasses = subclasses
		}
	}

	field, ok := findFieldByXMLTag(lookupType, tag)
	if !ok {
		owner := structName
		if isInverse {
			owner = targetClass
		}
		return checkSpec{}, nil, fmt.Errorf("no field with xml tag %q on %s", tag, owner)
	}

	// sh.NodeKind on a path ending in rdf:type is structurally satisfied:
	// Go's static type system already enforces literal/IRI/blank-node
	// distinctions at compile time. A pointer-to-struct field with no MRID
	// is a blank node by construction, an MRID-bearing reference is an IRI,
	// and a primitive field is a literal — none of which can be violated
	// at runtime.
	if c.Component == "sh.NodeKindConstraintComponent" && len(c.Path) >= 1 && c.Path[len(c.Path)-1] == "rdf.type" {
		return checkSpec{}, nil, fmt.Errorf("NodeKind on path ending in rdf:type is structurally satisfied")
	}

	compShort, ok := componentShort(c.Component)
	if !ok {
		return checkSpec{}, nil, fmt.Errorf("component %s not supported", c.Component)
	}
	if isInverse {
		// Inverse-path checks share function-name space with forward-path
		// checks of the same component on the same class+field, but
		// produce different code; differentiate via a suffix.
		compShort = compShort + "Inverse"
	}

	base := "Check" + stemCamel + structName + camelize(field.Name) + compShort
	used[base]++
	name := base
	if used[base] > 1 {
		name = fmt.Sprintf("%s_%d", base, used[base])
	}

	severity := c.Severity
	if severity == "" {
		severity = "sh.Violation"
	}

	cs := checkSpec{
		Name:      name,
		ShapeID:   shapeID,
		Class:     structName,
		Tag:       tag,
		Component: c.Component,
		Property:  tag,
		Message:   strings.Trim(c.Message, "\""),
		Severity:  severity,
	}

	var imports []string

	// Inverse-path branch. Required/MinCount/MaxCount/Class are the only
	// components that make sense via inverse traversal — they all reduce to
	// "count or classify objects whose forward reference points back here".
	if isInverse {
		// 2-segment inverse-then-forward path: walk the inverse hop to
		// the target class, then read a forward field on it. The only
		// component currently supported is HasValue against an enum-as-
		// IRI field (live pattern: ^Terminal.ConductingEquipment /
		// Terminal.phases hasValue PhaseCode.N).
		if len(c.Path) >= 2 {
			if len(c.Path) > 2 {
				return checkSpec{}, nil, fmt.Errorf("inverse path with %d segments not supported", len(c.Path))
			}
			if strings.HasPrefix(c.Path[1], "^") {
				return checkSpec{}, nil, fmt.Errorf("inverse path with second-hop inverse not supported")
			}
			forwardTag, ok := stripCIMPrefix(c.Path[1])
			if !ok {
				return checkSpec{}, nil, fmt.Errorf("forward segment %q not in cim namespace", c.Path[1])
			}
			forwardField, ok := findFieldByXMLTag(lookupType, forwardTag)
			if !ok {
				return checkSpec{}, nil, fmt.Errorf("no field with xml tag %q on %s", forwardTag, targetClass)
			}
			if c.Component != "sh.HasValueConstraintComponent" {
				return checkSpec{}, nil, fmt.Errorf("multi-segment inverse %s not supported", c.Component)
			}
			want, ok := c.Payload["Value"].(string)
			if !ok {
				return checkSpec{}, nil, fmt.Errorf("HasValue payload is not a string")
			}
			want = strings.TrimPrefix(strings.TrimSuffix(want, ">"), "<")
			constIdent, isEnum, err := enumURIFieldConst(forwardField, want)
			if !isEnum {
				return checkSpec{}, nil, fmt.Errorf("inverse HasValue forward field %q is not enum-URI typed", forwardField.Name)
			}
			if err != nil {
				return checkSpec{}, nil, err
			}
			prelude, cond := inverseHasEnumValueCheck(targetClasses, field, forwardField, constIdent)
			cs.Prelude, cs.Condition = prelude, cond
			cs.NoV = true
			imports = append(imports, "strings")
			return cs, imports, nil
		}
		switch c.Component {
		case "sh.RequiredConstraintComponent":
			prelude, cond := inverseCountCheck(targetClasses, field, "==", "0")
			cs.Prelude, cs.Condition = prelude, cond
			cs.NoV = true
			imports = append(imports, "strings")
		case "sh.MinCountConstraintComponent":
			min := int(anyToFloat(c.Payload["MinCount"]))
			prelude, cond := inverseCountCheck(targetClasses, field, "<", fmt.Sprintf("%d", min))
			cs.Prelude, cs.Condition = prelude, cond
			cs.NoV = true
			imports = append(imports, "strings")
		case "sh.MaxCountConstraintComponent":
			max := int(anyToFloat(c.Payload["MaxCount"]))
			prelude, cond := inverseCountCheck(targetClasses, field, ">", fmt.Sprintf("%d", max))
			cs.Prelude, cs.Condition = prelude, cond
			cs.NoV = true
			imports = append(imports, "strings")
		case "sh.ClassConstraintComponent":
			// Class on an inverse path asserts the *referrers* are of a
			// given class. Our inverse-index loop already filters to
			// *targetClasses, so the constraint is structurally satisfied
			// iff every concrete target class is the asserted class or
			// embeds it (Go's representation of subclass-of). Verify that
			// programmatically rather than empirically: if every target
			// satisfies, skip with a structural reason; otherwise refuse
			// rather than silently dropping the constraint, since there's
			// no runtime evaluator to fall back on.
			assertedClass, ok := classNameFromPayload(c.Payload["Class"])
			if !ok {
				return checkSpec{}, nil, fmt.Errorf("inverse Class payload not a cim.* string")
			}
			for _, tc := range targetClasses {
				if !isClassOrAncestor(tc, assertedClass) {
					return checkSpec{}, nil, fmt.Errorf("inverse Class %q not satisfied by target %q (would need a referrer-type check, not yet implemented)", assertedClass, tc)
				}
			}
			return checkSpec{}, nil, fmt.Errorf("inverse Class %q is a parent of every target subclass — structurally satisfied", assertedClass)
		default:
			return checkSpec{}, nil, fmt.Errorf("inverse %s not supported", c.Component)
		}
		return cs, imports, nil
	}

	// Multi-segment MaxCount=1 along any forward chain is structurally
	// satisfied: every reference hop in our Go data model is 0..1, so the
	// path-end value-count cannot exceed 1. Flag these explicitly before
	// the per-shape branches below, otherwise they fall through into the
	// "forwardChainOnly only handles Or" / "ends-rdf.type only handles
	// HasValue/In/Required" rejections and get reported as plain
	// "not supported" — which understates how many are actually OK.
	if len(c.Path) > 1 && !isInverse && c.Component == "sh.MaxCountConstraintComponent" &&
		(forwardChainEndsRdfType || forwardChainOnly) {
		if int(anyToFloat(c.Payload["MaxCount"])) == 1 {
			return checkSpec{}, nil, fmt.Errorf("multi-segment MaxCount=1 is structurally satisfied (refs are 0..1)")
		}
	}

	// Forward-chain-ending-rdf.type branch. Covers any number of forward
	// reference hops followed by a final rdf.type segment. The chain
	// resolves cleanly to a class-of-referenced-object check; the trailing
	// rdf.type is satisfied by the Go type system once the chain lands.
	// HasValue → exact match; In → allow-set; Required → reduce to "first
	// ref must be present" (the chain's existence is the constraint).
	if forwardChainEndsRdfType {
		refSegs := c.Path[:len(c.Path)-1]
		switch c.Component {
		case "sh.HasValueConstraintComponent":
			guard, cond, err := refClassEqualCondition(refSegs, structType, c.Payload["Value"])
			if err != nil {
				return checkSpec{}, nil, err
			}
			cs.Guard, cs.Condition = guard, cond
			imports = append(imports, "strings")
		case "sh.InConstraintComponent":
			guard, cond, err := refClassInCondition(refSegs, structType, c.Payload["Values"])
			if err != nil {
				return checkSpec{}, nil, err
			}
			cs.Guard, cs.Condition = guard, cond
			imports = append(imports, "strings")
		case "sh.RequiredConstraintComponent":
			// Required at end-rdf.type degenerates to "the chain's first
			// reference must be present". The chain implicitly requires
			// each subsequent hop to resolve too, but for the dominant
			// CIM modelling intent ("this must be set") checking the
			// first hop matches what the SHACL author meant. Half-set
			// multi-hop chains are rare in the live TTL and would need a
			// per-hop presence walk to detect; we don't emit that today.
			cond, err := requiredCondition(field)
			if err != nil {
				return checkSpec{}, nil, err
			}
			cs.Condition = cond
		default:
			return checkSpec{}, nil, fmt.Errorf("multi-segment %v ending in rdf.type not supported for %s", c.Path, c.Component)
		}
		return cs, imports, nil
	}

	// Forward chain WITHOUT trailing rdf.type. Two shapes map cleanly:
	//
	//   - sh.Or with all-Class shapes: the disjunction is the type test on
	//     the final referenced object.
	//   - sh.Datatype on a chain whose last segment is a literal field on
	//     a known compound class (e.g. Location.mainAddress / StreetAddress
	//     .status / Status.dateTime): walk the N-1 ref hops, type-assert
	//     to the literal's parent class, then apply the same datatype check
	//     used for single-segment string fields.
	if forwardChainOnly {
		if c.Component == "sh.DatatypeConstraintComponent" {
			cs2, imports2, err := multiSegDatatypeCheck(c, structType)
			if err == nil {
				cs.Guard = cs2.Guard
				cs.Condition = cs2.Condition
				imports = append(imports, imports2...)
				return cs, imports, nil
			}
			return checkSpec{}, nil, err
		}
		if c.Component != "sh.OrConstraintComponent" {
			return checkSpec{}, nil, fmt.Errorf("multi-segment path %v not supported for %s", c.Path, c.Component)
		}
		classes, err := orClassListFromShapes(c.Payload["Shapes"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		guard, targetVar, err := walkForwardRefChain(c.Path, structType, "v")
		if err != nil {
			return checkSpec{}, nil, err
		}
		var b strings.Builder
		b.WriteString(guard)
		fmt.Fprintf(&b, "\n\t\tisAllowedClass := false\n\t\tswitch %s.(type) {", targetVar)
		for _, cls := range classes {
			fmt.Fprintf(&b, "\n\t\tcase *cimgostructs.%s:\n\t\t\tisAllowedClass = true", cls)
		}
		b.WriteString("\n\t\t}")
		cs.Guard, cs.Condition = b.String(), "!isAllowedClass"
		imports = append(imports, "strings")
		return cs, imports, nil
	}

	// Any remaining multi-segment forward-or-mixed path that we couldn't
	// classify above is currently out of scope.
	if len(c.Path) > 1 {
		return checkSpec{}, nil, fmt.Errorf("multi-segment path %v not yet supported", c.Path)
	}

	switch c.Component {
	case "sh.MinExclusiveConstraintComponent":
		guard, cond, err := numericCompare(field, "<=", c.Payload["Value"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
	case "sh.MaxExclusiveConstraintComponent":
		guard, cond, err := numericCompare(field, ">=", c.Payload["Value"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
	case "sh.MinInclusiveConstraintComponent":
		guard, cond, err := numericCompare(field, "<", c.Payload["Value"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
	case "sh.MaxInclusiveConstraintComponent":
		guard, cond, err := numericCompare(field, ">", c.Payload["Value"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
	case "sh.RequiredConstraintComponent":
		cond, err := requiredCondition(field)
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Condition = cond
	case "sh.MinCountConstraintComponent":
		guard, cond, err := minCountCondition(field, c.Payload["MinCount"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
	case "sh.MaxCountConstraintComponent":
		guard, cond, err := maxCountCondition(field, c.Payload["MaxCount"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
	case "sh.HasValueConstraintComponent":
		guard, cond, err := hasValueCondition(field, c.Payload["Value"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
	case "sh.InConstraintComponent":
		guard, cond, err := inCondition(field, c.Payload["Values"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
	case "sh.MinLengthConstraintComponent":
		guard, cond, err := minLengthCondition(field, c.Payload["MinLength"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
		imports = append(imports, "unicode/utf8")
	case "sh.MaxLengthConstraintComponent":
		guard, cond, err := maxLengthCondition(field, c.Payload["MaxLength"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
		imports = append(imports, "unicode/utf8")
	case "sh.PatternConstraintComponent":
		decl, guard, cond, err := patternCondition(field, c.Payload, name+"Regex")
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Decl, cs.Guard, cs.Condition = decl, guard, cond
		imports = append(imports, "regexp")
	case "sh.LessThanConstraintComponent":
		guard, cond, err := pairCompare(structType, field, c.Payload["Path"], ">=")
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
	case "sh.LessThanOrEqualsConstraintComponent":
		guard, cond, err := pairCompare(structType, field, c.Payload["Path"], ">")
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
	case "sh.ClassConstraintComponent":
		guard, cond, err := classCondition(field, c.Payload["Class"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
		imports = append(imports, "strings")
	case "sh.DatatypeConstraintComponent":
		guard, cond, dtImports, err := datatypeCondition(field, c.Payload["Datatype"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
		imports = append(imports, dtImports...)
	case "sh.NotConstraintComponent":
		// Single-segment Not is a forward-Class assertion with the
		// condition inverted. The wrapped shape must reduce to a single
		// sh.Class — anything richer (Not of a length range, of an Or,
		// etc.) is too uncommon here to be worth recursive emission.
		notClass, err := notClassFromShapeRef(c.Payload["ShapeRef"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		guard, _, err := classCondition(field, "cim."+notClass)
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, "isWantedClass"
		imports = append(imports, "strings")
	default:
		return checkSpec{}, nil, fmt.Errorf("component %s not supported", c.Component)
	}

	return cs, imports, nil
}

// Status of remaining SHACL constraint shapes, audited against the live
// CGMES TTL files. The 2182 total skips break down roughly as:
//
//   Structurally satisfied — skipped intentionally; the Go data model
//   already encodes the constraint, no code worth emitting:
//
//     - MaxCount=1 on scalar field [1302]: omitempty ints/strings/floats
//       trivially have value count ≤ 1.
//     - MaxCount=1 on pointer field [334]: pointer slots are 0..1 by shape.
//     - Multi-segment MaxCount=1 along forward chains [56]: every ref hop
//       in the Go model is 0..1, so the path-end value-count cannot exceed 1.
//     - Inverse Class [15]: verified programmatically (isClassOrAncestor) —
//       every concrete target class is the asserted class or embeds it,
//       which combined with the inverse-index *X filter satisfies the
//       constraint by construction. If a future TTL ever breaks this
//       invariant the generator will refuse with a "not yet implemented"
//       error rather than silently dropping the constraint.
//     - sh.NodeKindConstraintComponent on a path ending in rdf:type [8]:
//       once the chain resolves to a typed *cim.X, NodeKind=Literal/IRI/
//       BlankNode is automatically satisfied by Go's static type system;
//       any emitted check would be a no-op.
//
//     Total structurally satisfied: ~1715.
//
//   Won't fix — see README.md "Known limitations":
//
//     - Required on bool fields [413]: would need *bool in cimgostructs;
//       wide refactor for callers, no detectable wins.
//     - LessThan paired field xml tag not found [22]: 4 are upstream TTL
//       typos, 18 are cross-class semantics that require shared-MRID
//       merging across concrete types — not representable in the current
//       single-type-per-MRID Go model.
//
//   Long-tail (won't fix without upstream changes):
//
//     - SPARQL-driven targetNode shapes [10]: "Class X has no Go struct"
//       cases like cim:IDchecks, cim:FloatSpecialValues, cim:DanglingReferences,
//       cim:AllGeneratingUnit, cim:TextDiagramObjectDiagramObject,
//       cim:SubstationCount, cim:IdentifiedObjectStringLength,
//       cim:AngleReference. These use sh:targetNode <sentinel> + sh:sparql
//       to define their target via a SPARQL query rather than a class —
//       out of shaclgen's static-field-check scope.
//     - sh.In payload empty [4]: TTL bug — empty `sh:in ()` lists in
//       Operation profile (would imply "no value acceptable"). Skip.
//     - non-CIM target classes [4]: diff.DifferenceModel and mdc.FullModel
//       — model-metadata classes, out of scope.
//     - Upstream TTL typos [≈10]: CSConverter (vs CsConverter), GovHydroIEEE1
//       (no such class), CrossCompoundTurbineGovernorDyanmics (Dynamics
//       misspelled), missing fields like CrossCompoundTurbineGovernorDynamics
//       .SynchronousMachineDynamics, WindTurbineType3or4IEC.WindContQIEC.

// componentShort returns the camel-case suffix used in generated function
// names for each supported SHACL constraint component. Returning ok=false
// signals the caller to skip the constraint with a "not supported" reason.
func componentShort(component string) (string, bool) {
	switch component {
	case "sh.MinExclusiveConstraintComponent":
		return "MinExclusive", true
	case "sh.MaxExclusiveConstraintComponent":
		return "MaxExclusive", true
	case "sh.MinInclusiveConstraintComponent":
		return "MinInclusive", true
	case "sh.MaxInclusiveConstraintComponent":
		return "MaxInclusive", true
	case "sh.RequiredConstraintComponent":
		return "Required", true
	case "sh.MinCountConstraintComponent":
		return "MinCount", true
	case "sh.MaxCountConstraintComponent":
		return "MaxCount", true
	case "sh.HasValueConstraintComponent":
		return "HasValue", true
	case "sh.InConstraintComponent":
		return "In", true
	case "sh.MinLengthConstraintComponent":
		return "MinLength", true
	case "sh.MaxLengthConstraintComponent":
		return "MaxLength", true
	case "sh.PatternConstraintComponent":
		return "Pattern", true
	case "sh.LessThanConstraintComponent":
		return "LessThan", true
	case "sh.LessThanOrEqualsConstraintComponent":
		return "LessThanOrEquals", true
	case "sh.ClassConstraintComponent":
		return "Class", true
	case "sh.DatatypeConstraintComponent":
		return "Datatype", true
	case "sh.OrConstraintComponent":
		return "Or", true
	case "sh.NotConstraintComponent":
		return "Not", true
	}
	return "", false
}

// numericCompare returns the guard block and condition expression for an
// ordering constraint. `op` is the comparison that means "violates" — e.g.
// "<=" for MinExclusive: a value at or below the threshold violates the rule.
func numericCompare(field reflect.StructField, op string, payload any) (string, string, error) {
	threshold := anyToFloatLiteral(payload)
	cast := ""
	switch field.Type.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		cast = "float64"
	case reflect.Float32, reflect.Float64:
	default:
		return "", "", fmt.Errorf("unsupported numeric kind %s", field.Type.Kind())
	}
	guard := fmt.Sprintf("\t\t// omitempty zero ≡ absent — skip per existing getCount semantics\n\t\tif v.%s == 0 {\n\t\t\tcontinue\n\t\t}", field.Name)
	var cond string
	if cast != "" {
		cond = fmt.Sprintf("%s(v.%s) %s %s", cast, field.Name, op, threshold)
	} else {
		cond = fmt.Sprintf("v.%s %s %s", field.Name, op, threshold)
	}
	return guard, cond, nil
}

func requiredCondition(field reflect.StructField) (string, error) {
	switch field.Type.Kind() {
	case reflect.Ptr:
		return fmt.Sprintf("v.%s == nil", field.Name), nil
	case reflect.Slice:
		return fmt.Sprintf("len(v.%s) == 0", field.Name), nil
	case reflect.String:
		return fmt.Sprintf("v.%s == \"\"", field.Name), nil
	case reflect.Int, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64:
		// omitempty caveat: zero is indistinguishable from absent for scalars.
		return fmt.Sprintf("v.%s == 0", field.Name), nil
	case reflect.Bool:
		// Bool fields are xml:",omitempty" — false is indistinguishable from
		// absent after decode. Required is structurally satisfied: the field
		// always has a value (true or false) in the Go struct.
		return "", fmt.Errorf("bool Required is structurally satisfied: false is indistinguishable from absent")
	default:
		return "", fmt.Errorf("unsupported required kind %s", field.Type.Kind())
	}
}

// minCountCondition handles sh.MinCount. For pointer fields, MinCount=1 is
// equivalent to Required. For slices we compare against the literal threshold.
// Other field kinds are skipped — a scalar's count is always 0 or 1, and a
// MinCount > 1 against that would be unsatisfiable, suggesting bad input.
func minCountCondition(field reflect.StructField, payload any) (string, string, error) {
	min := int(anyToFloat(payload))
	switch field.Type.Kind() {
	case reflect.Slice:
		return "", fmt.Sprintf("len(v.%s) < %d", field.Name, min), nil
	case reflect.Ptr:
		if min <= 1 {
			return "", fmt.Sprintf("v.%s == nil", field.Name), nil
		}
		return "", "", fmt.Errorf("MinCount=%d on pointer field is unsatisfiable", min)
	default:
		return "", "", fmt.Errorf("MinCount on %s field not supported", field.Type.Kind())
	}
}

// maxCountCondition handles sh.MaxCount. Pointers and scalars are structurally
// bounded at 1, so MaxCount >= 1 is vacuous and we skip it. MaxCount=0 (a
// "must not be set" rule) is rare but real; emit it as a presence check.
func maxCountCondition(field reflect.StructField, payload any) (string, string, error) {
	max := int(anyToFloat(payload))
	switch field.Type.Kind() {
	case reflect.Slice:
		return "", fmt.Sprintf("len(v.%s) > %d", field.Name, max), nil
	case reflect.Ptr:
		if max == 0 {
			return "", fmt.Sprintf("v.%s != nil", field.Name), nil
		}
		return "", "", fmt.Errorf("MaxCount=%d on pointer field is structurally satisfied", max)
	case reflect.String, reflect.Int, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64, reflect.Bool:
		if max == 0 {
			// omitempty: zero ≡ absent.
			return "", fmt.Sprintf("v.%s != %s", field.Name, zeroLiteralFor(field.Type.Kind())), nil
		}
		return "", "", fmt.Errorf("MaxCount=%d on scalar field is structurally satisfied", max)
	default:
		return "", "", fmt.Errorf("MaxCount on %s field not supported", field.Type.Kind())
	}
}

// hasValueCondition handles sh.HasValue for string-typed fields and for
// enum-as-IRI reference fields (pointer to struct{URI string}). Enum-URI
// fields need a special path because the value lives one pointer hop deeper
// than a plain string — comparing v.Field.URI against the matching
// cimgostructs constant catches violations a generic string compare on
// v.Field would miss.
func hasValueCondition(field reflect.StructField, payload any) (string, string, error) {
	want, ok := payload.(string)
	if !ok {
		return "", "", fmt.Errorf("HasValue payload is not a string")
	}
	want = strings.TrimPrefix(strings.TrimSuffix(want, ">"), "<")
	if constIdent, isEnum, err := enumURIFieldConst(field, want); isEnum {
		if err != nil {
			return "", "", err
		}
		guard := fmt.Sprintf("\t\tif v.%s == nil {\n\t\t\tcontinue\n\t\t}", field.Name)
		cond := fmt.Sprintf("v.%s.URI != cimgostructs.%s", field.Name, constIdent)
		return guard, cond, nil
	}
	switch field.Type.Kind() {
	case reflect.String:
		guard := fmt.Sprintf("\t\tif v.%s == \"\" {\n\t\t\tcontinue\n\t\t}", field.Name)
		cond := fmt.Sprintf("v.%s != %q", field.Name, want)
		return guard, cond, nil
	case reflect.Bool:
		bval := want == "true"
		guard := "" // false is a legitimate value; no skip-zero guard
		cond := fmt.Sprintf("v.%s != %v", field.Name, bval)
		return guard, cond, nil
	case reflect.Int, reflect.Int32, reflect.Int64:
		n := int64(anyToFloat(want))
		guard := fmt.Sprintf("\t\tif v.%s == 0 {\n\t\t\tcontinue\n\t\t}", field.Name)
		cond := fmt.Sprintf("int64(v.%s) != %d", field.Name, n)
		return guard, cond, nil
	case reflect.Float32, reflect.Float64:
		f := anyToFloat(want)
		guard := fmt.Sprintf("\t\tif v.%s == 0 {\n\t\t\tcontinue\n\t\t}", field.Name)
		cond := fmt.Sprintf("float64(v.%s) != %v", field.Name, f)
		return guard, cond, nil
	default:
		return "", "", fmt.Errorf("HasValue on %s field not supported", field.Type.Kind())
	}
}

// inCondition handles sh.In for string-typed fields and for enum-as-IRI
// reference fields. Same enum-URI rationale as hasValueCondition.
func inCondition(field reflect.StructField, payload any) (string, string, error) {
	rawValues, ok := payload.([]any)
	if !ok {
		return "", "", fmt.Errorf("In payload is not a list")
	}
	values := make([]string, 0, len(rawValues))
	for _, v := range rawValues {
		s, ok := v.(string)
		if !ok {
			return "", "", fmt.Errorf("In list contains non-string %v", v)
		}
		values = append(values, strings.TrimPrefix(strings.TrimSuffix(s, ">"), "<"))
	}
	if len(values) > 0 {
		if _, isEnum, _ := enumURIFieldConst(field, values[0]); isEnum {
			consts := make([]string, 0, len(values))
			for _, want := range values {
				constIdent, _, err := enumURIFieldConst(field, want)
				if err != nil {
					return "", "", err
				}
				consts = append(consts, constIdent)
			}
			var b strings.Builder
			fmt.Fprintf(&b, "\t\tif v.%s == nil {\n\t\t\tcontinue\n\t\t}\n", field.Name)
			b.WriteString("\t\tallowed := map[string]bool{")
			for i, c := range consts {
				if i > 0 {
					b.WriteString(", ")
				}
				fmt.Fprintf(&b, "cimgostructs.%s: true", c)
			}
			b.WriteString("}")
			cond := fmt.Sprintf("!allowed[v.%s.URI]", field.Name)
			return b.String(), cond, nil
		}
	}
	switch field.Type.Kind() {
	case reflect.String:
	// handled below
	case reflect.Bool:
		// With only two possible values, build an explicit comparison.
		// An empty allow-list means nothing is allowed — always a violation.
		allowTrue := false
		allowFalse := false
		for _, val := range values {
			if val == "true" {
				allowTrue = true
			} else {
				allowFalse = true
			}
		}
		if allowTrue && allowFalse {
			return "", "", fmt.Errorf("In on bool field allows both true and false: structurally satisfied")
		}
		want := allowTrue
		cond := fmt.Sprintf("v.%s != %v", field.Name, want)
		return "", cond, nil
	case reflect.Int, reflect.Int32, reflect.Int64:
		var b strings.Builder
		fmt.Fprintf(&b, "\t\tif v.%s == 0 {\n\t\t\tcontinue\n\t\t}\n", field.Name)
		b.WriteString("\t\tallowed := map[int64]bool{")
		for i, val := range values {
			if i > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(&b, "%d: true", int64(anyToFloat(val)))
		}
		b.WriteString("}")
		cond := fmt.Sprintf("!allowed[int64(v.%s)]", field.Name)
		return b.String(), cond, nil
	case reflect.Float32, reflect.Float64:
		var b strings.Builder
		fmt.Fprintf(&b, "\t\tif v.%s == 0 {\n\t\t\tcontinue\n\t\t}\n", field.Name)
		b.WriteString("\t\tallowed := map[float64]bool{")
		for i, val := range values {
			if i > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(&b, "%v: true", anyToFloat(val))
		}
		b.WriteString("}")
		cond := fmt.Sprintf("!allowed[float64(v.%s)]", field.Name)
		return b.String(), cond, nil
	default:
		return "", "", fmt.Errorf("In on %s field not supported", field.Type.Kind())
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\t\tif v.%s == \"\" {\n\t\t\tcontinue\n\t\t}\n", field.Name)
	b.WriteString("\t\tallowed := map[string]bool{")
	for i, val := range values {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%q: true", val)
	}
	b.WriteString("}")
	cond := fmt.Sprintf("!allowed[v.%s]", field.Name)
	return b.String(), cond, nil
}

// enumURIFieldConst inspects field for the enum-as-IRI shape (pointer to a
// struct with a single string field named URI). When the shape matches, the
// returned constIdent is the cimgostructs constant identifier corresponding
// to payload — derived by stripping the cim. prefix and dropping the dot
// between enum name and member, since cimgostructs names enum constants
// `<EnumName><Member>` (e.g. cim.InputSignalKind.generatorElectricalPower →
// InputSignalKindgeneratorElectricalPower). A missing constant turns into a
// build error in the generated code, not a silent miscompare here.
func enumURIFieldConst(field reflect.StructField, payload string) (string, bool, error) {
	if field.Type.Kind() != reflect.Ptr {
		return "", false, nil
	}
	elem := field.Type.Elem()
	if elem.Kind() != reflect.Struct || elem.NumField() != 1 {
		return "", false, nil
	}
	uriField := elem.Field(0)
	if uriField.Name != "URI" || uriField.Type.Kind() != reflect.String {
		return "", false, nil
	}
	rest, ok := stripCIMPrefix(payload)
	if !ok {
		return "", true, fmt.Errorf("enum payload %q not in cim namespace", payload)
	}
	dot := strings.IndexByte(rest, '.')
	if dot < 0 {
		return "", true, fmt.Errorf("enum payload %q missing '.member' segment", payload)
	}
	return rest[:dot] + rest[dot+1:], true, nil
}

// minLengthCondition handles sh.MinLength on string fields. Skip absent
// values (empty string ≡ omitempty absent) so an unrelated MinLength rule
// doesn't fire on every dataset element that simply doesn't set the field —
// Required handles "must be present" separately.
func minLengthCondition(field reflect.StructField, payload any) (string, string, error) {
	if field.Type.Kind() != reflect.String {
		return "", "", fmt.Errorf("MinLength on non-string field (%s) not supported", field.Type.Kind())
	}
	min := int(anyToFloat(payload))
	guard := fmt.Sprintf("\t\tif v.%s == \"\" {\n\t\t\tcontinue\n\t\t}", field.Name)
	cond := fmt.Sprintf("utf8.RuneCountInString(v.%s) < %d", field.Name, min)
	return guard, cond, nil
}

// maxLengthCondition handles sh.MaxLength on string fields. Empty string
// trivially passes (0 ≤ max), so the skip-empty guard is redundant but kept
// for parity with MinLength/HasValue/In so the loop body shape stays uniform.
func maxLengthCondition(field reflect.StructField, payload any) (string, string, error) {
	if field.Type.Kind() != reflect.String {
		return "", "", fmt.Errorf("MaxLength on non-string field (%s) not supported", field.Type.Kind())
	}
	max := int(anyToFloat(payload))
	guard := fmt.Sprintf("\t\tif v.%s == \"\" {\n\t\t\tcontinue\n\t\t}", field.Name)
	cond := fmt.Sprintf("utf8.RuneCountInString(v.%s) > %d", field.Name, max)
	return guard, cond, nil
}

// patternCondition handles sh.Pattern on string fields. The regex is hoisted
// to a package-level var via regexp.MustCompile so each Check call reuses one
// compiled pattern. We test-compile at generator time: if the SHACL pattern
// uses XSD/XPath-only syntax that Go's RE2 engine rejects (e.g. lookahead,
// character-class subtraction), we return an error and the caller skips the
// constraint rather than emitting code that would refuse to compile.
func patternCondition(field reflect.StructField, payload map[string]any, varName string) (string, string, string, error) {
	if field.Type.Kind() != reflect.String {
		return "", "", "", fmt.Errorf("Pattern on non-string field (%s) not supported", field.Type.Kind())
	}
	pat, ok := payload["Pattern"].(string)
	if !ok {
		return "", "", "", fmt.Errorf("Pattern payload missing or not a string")
	}
	full := pat
	if flags, ok := payload["Flags"].(string); ok && flags != "" {
		full = "(?" + flags + ")" + pat
	}
	if _, err := regexp.Compile(full); err != nil {
		return "", "", "", fmt.Errorf("Pattern regex %q: %w", full, err)
	}
	decl := fmt.Sprintf("var %s = regexp.MustCompile(%q)", varName, full)
	guard := fmt.Sprintf("\t\tif v.%s == \"\" {\n\t\t\tcontinue\n\t\t}", field.Name)
	cond := fmt.Sprintf("!%s.MatchString(v.%s)", varName, field.Name)
	return decl, guard, cond, nil
}

// pairCompare emits a cross-field comparison on the same struct, used for
// sh.LessThan and sh.LessThanOrEquals. `op` is the comparison that *violates*
// the rule: ">=" for LessThan (A must be < B; violation when A >= B), ">" for
// LessThanOrEquals. Skip-empty matches numericCompare's omitempty semantics:
// if either field is zero we treat it as absent and bow out of the check.
func pairCompare(structType reflect.Type, fieldA reflect.StructField, payloadPath any, op string) (string, string, error) {
	pathB, ok := payloadPath.(string)
	if !ok {
		return "", "", fmt.Errorf("LessThan/LessThanOrEquals payload Path missing or not a string")
	}
	tagB, ok := stripCIMPrefix(pathB)
	if !ok {
		tagB = pathB
	}
	fieldB, ok := findFieldByXMLTag(structType, tagB)
	if !ok {
		if xmlTagExistsOnAnyStruct(tagB) {
			// tagB is a real field on a sibling class, not on the current
			// target. In RDF the two are mutually exclusive subtypes of the
			// same parent (e.g. SynchronousMachineTimeConstantReactance vs.
			// SynchronousMachineEquivalentCircuit). In valid CGMES data the
			// comparison therefore cannot have both operands visible at once,
			// so the constraint is vacuously satisfied for this target.
			return "", "", fmt.Errorf("paired field %q is on a sibling class — constraint is vacuously satisfied for this target", tagB)
		}
		return "", "", fmt.Errorf("paired field xml tag %q not found", tagB)
	}
	castA, okA := numericCastFor(fieldA.Type.Kind())
	castB, okB := numericCastFor(fieldB.Type.Kind())
	if !okA {
		return "", "", fmt.Errorf("LessThan A field kind %s not numeric", fieldA.Type.Kind())
	}
	if !okB {
		return "", "", fmt.Errorf("LessThan B field kind %s not numeric", fieldB.Type.Kind())
	}
	guard := fmt.Sprintf("\t\tif v.%s == 0 || v.%s == 0 {\n\t\t\tcontinue\n\t\t}", fieldA.Name, fieldB.Name)
	cond := fmt.Sprintf("%s %s %s", castExpr(castA, "v."+fieldA.Name), op, castExpr(castB, "v."+fieldB.Name))
	return guard, cond, nil
}

// numericCastFor returns the wrapper cast (or "" if none needed) for a
// numeric kind so cross-kind comparisons can be normalised to float64.
// Returns ok=false for non-numeric kinds.
func numericCastFor(k reflect.Kind) (string, bool) {
	switch k {
	case reflect.Int, reflect.Int32, reflect.Int64:
		return "float64", true
	case reflect.Float32, reflect.Float64:
		return "", true
	}
	return "", false
}

func castExpr(cast, expr string) string {
	if cast == "" {
		return expr
	}
	return cast + "(" + expr + ")"
}

// classCondition handles sh.ClassConstraintComponent for forward reference
// fields. It dereferences the MRID, looks up the target object in the
// dataset, and verifies its Go type matches the expected class. Missing
// references and dangling MRIDs are treated as "skip" (no violation): a
// missing-reference complaint is the job of sh.Required, not sh.Class, so
// flagging here would double-report. When the named class is abstract (not
// in StructMap) we accept any concrete subclass — same dispatch as the
// inverse-path branch.
func classCondition(field reflect.StructField, payload any) (string, string, error) {
	want, ok := payload.(string)
	if !ok {
		return "", "", fmt.Errorf("Class payload missing or not a string")
	}
	wantClass, ok := stripCIMPrefix(want)
	if !ok {
		return "", "", fmt.Errorf("Class %q not in cim namespace", want)
	}
	if field.Type.Kind() != reflect.Ptr {
		return "", "", fmt.Errorf("Class on non-pointer field (%s) not supported", field.Type.Kind())
	}
	var classes []string
	if _, ok := cimgostructs.StructMap[wantClass]; ok {
		classes = []string{wantClass}
	} else {
		classes = concreteSubclassesEmbedding(wantClass)
		if len(classes) == 0 {
			return "", "", fmt.Errorf("Class %q has no Go struct", wantClass)
		}
	}
	if len(classes) == 1 {
		guard := fmt.Sprintf(`		if v.%s == nil {
			continue
		}
		refID := strings.TrimPrefix(v.%s.MRID, "#")
		target, found := dataset.Elements[refID]
		if !found {
			continue
		}
		_, isWantedClass := target.(*cimgostructs.%s)`, field.Name, field.Name, classes[0])
		return guard, "!isWantedClass", nil
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\t\tif v.%s == nil {\n\t\t\tcontinue\n\t\t}\n", field.Name)
	fmt.Fprintf(&b, "\t\trefID := strings.TrimPrefix(v.%s.MRID, \"#\")\n", field.Name)
	b.WriteString("\t\ttarget, found := dataset.Elements[refID]\n")
	b.WriteString("\t\tif !found {\n\t\t\tcontinue\n\t\t}\n")
	b.WriteString("\t\tisWantedClass := false\n")
	b.WriteString("\t\tswitch target.(type) {")
	for _, cls := range classes {
		fmt.Fprintf(&b, "\n\t\tcase *cimgostructs.%s:\n\t\t\tisWantedClass = true", cls)
	}
	b.WriteString("\n\t\t}")
	return b.String(), "!isWantedClass", nil
}

// datatypeCondition handles sh.DatatypeConstraintComponent. For the dominant
// xsd.dateTime case (and xsd.date) we emit a parse attempt against the Go
// time package; other XSD datatypes are mostly redundant with Go's static
// types and skipped with a structural-satisfaction reason.
func datatypeCondition(field reflect.StructField, payload any) (string, string, []string, error) {
	dt, ok := payload.(string)
	if !ok {
		return "", "", nil, fmt.Errorf("Datatype payload missing or not a string")
	}
	if field.Type.Kind() != reflect.String {
		// Strongly-typed Go fields satisfy the datatype constraint by
		// construction (an int field cannot hold a non-integer value).
		return "", "", nil, fmt.Errorf("Datatype %q on %s field is structurally satisfied", dt, field.Type.Kind())
	}
	switch dt {
	case "xsd.dateTime":
		guard := fmt.Sprintf("\t\tif v.%s == \"\" {\n\t\t\tcontinue\n\t\t}\n\t\t_, parseErr := time.Parse(time.RFC3339, v.%s)", field.Name, field.Name)
		return guard, "parseErr != nil", []string{"time"}, nil
	case "xsd.date":
		guard := fmt.Sprintf("\t\tif v.%s == \"\" {\n\t\t\tcontinue\n\t\t}\n\t\t_, parseErr := time.Parse(\"2006-01-02\", v.%s)", field.Name, field.Name)
		return guard, "parseErr != nil", []string{"time"}, nil
	case "xsd.gMonthDay":
		// xsd:gMonthDay canonical form is "--MM-DD" but CGMES instances
		// commonly write "MM-DD" without the leading dashes. Try both
		// before flagging a violation.
		guard := fmt.Sprintf(`		if v.%s == "" {
			continue
		}
		_, parseErr1 := time.Parse("--01-02", v.%s)
		_, parseErr2 := time.Parse("01-02", v.%s)`, field.Name, field.Name, field.Name)
		return guard, "parseErr1 != nil && parseErr2 != nil", []string{"time"}, nil
	}
	return "", "", nil, fmt.Errorf("Datatype %q not supported on string field", dt)
}

// buildDatasetCardinalityCheck handles `^rdf.type` cardinality constraints
// (MinCount/MaxCount on the count of focus-class instances in the dataset).
// Emitted as a DatasetCheck — the per-element loop is skipped entirely and
// a single violation is appended (or not) based on the global count.
func buildDatasetCardinalityCheck(stemCamel, structName, shapeID string, c shaclimport.ConstraintInfo, used map[string]int) (checkSpec, []string, error) {
	compShort, ok := componentShort(c.Component)
	if !ok {
		return checkSpec{}, nil, fmt.Errorf("component %s not supported on ^rdf.type", c.Component)
	}
	var op, threshold string
	switch c.Component {
	case "sh.MinCountConstraintComponent":
		min := int(anyToFloat(c.Payload["MinCount"]))
		op, threshold = "<", fmt.Sprintf("%d", min)
	case "sh.MaxCountConstraintComponent":
		max := int(anyToFloat(c.Payload["MaxCount"]))
		op, threshold = ">", fmt.Sprintf("%d", max)
	default:
		return checkSpec{}, nil, fmt.Errorf("only Min/MaxCount supported on ^rdf.type, got %s", c.Component)
	}
	// Resolve focus class to either a single concrete or a list of concrete
	// subclasses (when the focus is abstract). Counting is a union over all
	// matching concrete types.
	var classes []string
	if _, ok := cimgostructs.StructMap[structName]; ok {
		classes = []string{structName}
	} else {
		classes = concreteSubclassesEmbedding(structName)
		if len(classes) == 0 {
			return checkSpec{}, nil, fmt.Errorf("focus class %q has no Go struct", structName)
		}
	}
	base := "Check" + stemCamel + structName + "Type" + compShort + "Inverse"
	used[base]++
	name := base
	if used[base] > 1 {
		name = fmt.Sprintf("%s_%d", base, used[base])
	}
	severity := c.Severity
	if severity == "" {
		severity = "sh.Violation"
	}
	var b strings.Builder
	b.WriteString("\tdatasetCount := 0\n")
	b.WriteString("\tfor _, ref := range dataset.Elements {\n")
	b.WriteString("\t\tswitch ref.(type) {")
	for _, cls := range classes {
		fmt.Fprintf(&b, "\n\t\tcase *cimgostructs.%s:\n\t\t\tdatasetCount++", cls)
	}
	b.WriteString("\n\t\t}\n")
	b.WriteString("\t}")
	cs := checkSpec{
		Name:         name,
		ShapeID:      shapeID,
		Class:        structName,
		Tag:          "^rdf.type",
		Component:    c.Component,
		Property:     "^rdf.type",
		Message:      strings.Trim(c.Message, "\""),
		Severity:     severity,
		Prelude:      b.String(),
		Condition:    fmt.Sprintf("datasetCount %s %s", op, threshold),
		DatasetCheck: true,
	}
	return cs, nil, nil
}

// multiSegDatatypeCheck handles sh.Datatype on a forward chain where the
// last segment is a literal field on a known compound class. The strategy:
// walk the N-1 reference hops, type-assert the resulting Element to the
// parent class extracted from the last segment, then apply the existing
// single-field datatype guard against `parent.<LiteralField>`.
//
// Returns a partially-filled checkSpec (Guard + Condition) plus any
// extra imports the datatype check needs (e.g. "time").
func multiSegDatatypeCheck(c shaclimport.ConstraintInfo, structType reflect.Type) (checkSpec, []string, error) {
	if len(c.Path) < 2 {
		return checkSpec{}, nil, fmt.Errorf("Datatype on multi-segment path needs ≥ 2 segments")
	}
	refSegs := c.Path[:len(c.Path)-1]
	lastRaw := c.Path[len(c.Path)-1]
	lastSeg, ok := stripCIMPrefix(lastRaw)
	if !ok {
		return checkSpec{}, nil, fmt.Errorf("last segment %q not in cim namespace", lastRaw)
	}
	parts := strings.SplitN(lastSeg, ".", 2)
	if len(parts) != 2 {
		return checkSpec{}, nil, fmt.Errorf("last segment %q has no class.field shape", lastSeg)
	}
	parentClass := parts[0]
	factory, ok := cimgostructs.StructMap[parentClass]
	if !ok {
		return checkSpec{}, nil, fmt.Errorf("Datatype chain ends on %q which has no Go struct", parentClass)
	}
	parentType := reflect.TypeOf(factory()).Elem()
	field, ok := findFieldByXMLTag(parentType, lastSeg)
	if !ok {
		return checkSpec{}, nil, fmt.Errorf("no field with xml tag %q on %s", lastSeg, parentClass)
	}
	chainGuard, targetVar, err := walkForwardRefChain(refSegs, structType, "v")
	if err != nil {
		return checkSpec{}, nil, err
	}
	// datatypeCondition formats access as `v.<Field>`. We need access via
	// `parent.<Field>`. Build a lightweight adapter struct so we can reuse
	// it: the trick is to apply the same logic but with a different prefix.
	//
	// Cleanest path: rewrite the relevant guard/condition fragments here,
	// since datatypeCondition's per-case bodies are short.
	dt, ok := c.Payload["Datatype"].(string)
	if !ok {
		return checkSpec{}, nil, fmt.Errorf("Datatype payload missing or not a string")
	}
	if field.Type.Kind() != reflect.String {
		return checkSpec{}, nil, fmt.Errorf("Datatype %q on %s field is structurally satisfied", dt, field.Type.Kind())
	}

	var b strings.Builder
	b.WriteString(chainGuard)
	fmt.Fprintf(&b, "\n\t\tparent, parentOk := %s.(*cimgostructs.%s)\n\t\tif !parentOk {\n\t\t\tcontinue\n\t\t}\n", targetVar, parentClass)
	fmt.Fprintf(&b, "\t\tif parent.%s == \"\" {\n\t\t\tcontinue\n\t\t}\n", field.Name)
	var cond string
	var extraImports []string
	switch dt {
	case "xsd.dateTime":
		fmt.Fprintf(&b, "\t\t_, parseErr := time.Parse(time.RFC3339, parent.%s)", field.Name)
		cond = "parseErr != nil"
		extraImports = []string{"time", "strings"}
	case "xsd.date":
		fmt.Fprintf(&b, "\t\t_, parseErr := time.Parse(\"2006-01-02\", parent.%s)", field.Name)
		cond = "parseErr != nil"
		extraImports = []string{"time", "strings"}
	case "xsd.gMonthDay":
		fmt.Fprintf(&b, "\t\t_, parseErr1 := time.Parse(\"--01-02\", parent.%s)\n", field.Name)
		fmt.Fprintf(&b, "\t\t_, parseErr2 := time.Parse(\"01-02\", parent.%s)", field.Name)
		cond = "parseErr1 != nil && parseErr2 != nil"
		extraImports = []string{"time", "strings"}
	default:
		return checkSpec{}, nil, fmt.Errorf("Datatype %q not supported on multi-segment chain", dt)
	}
	return checkSpec{Guard: b.String(), Condition: cond}, extraImports, nil
}

// walkForwardRefChain emits Guard text that chases a sequence of forward
// reference hops. Each segment must be `cim.<Class>.<field>` (already with
// cim. prefix). The function dereferences each reference, looks the target
// up in the dataset, and (for all but the final hop) type-asserts the
// target to the class implied by the *next* segment's class portion. The
// final hop returns the raw Element value as `targetVar`; the caller plugs
// in the final test (e.g. type-switch for In, single type-assert for
// HasValue).
//
// At every step, if a reference is missing or the dataset lookup fails or a
// type assertion fails, we `continue`. The chain not resolving means there
// is no "value" for the property path at this focus node — so by SHACL
// semantics no value-shape constraint can fail; sh.Required is the
// constraint that signals "the chain must resolve", and lives separately.
func walkForwardRefChain(pathSegs []string, startType reflect.Type, startVar string) (string, string, error) {
	if len(pathSegs) == 0 {
		return "", "", fmt.Errorf("empty chain")
	}
	var b strings.Builder
	currentVar := startVar
	currentType := startType
	for i, raw := range pathSegs {
		seg, ok := stripCIMPrefix(raw)
		if !ok {
			seg = raw
		}
		field, ok := findFieldByXMLTag(currentType, seg)
		if !ok {
			return "", "", fmt.Errorf("chain step %d: no field %q on %s", i, seg, currentType.Name())
		}
		if field.Type.Kind() != reflect.Ptr {
			return "", "", fmt.Errorf("chain step %d: field %q is %s, expected pointer", i, seg, field.Type.Kind())
		}
		fmt.Fprintf(&b, "\t\tif %s.%s == nil {\n\t\t\tcontinue\n\t\t}\n", currentVar, field.Name)
		fmt.Fprintf(&b, "\t\trefID%d := strings.TrimPrefix(%s.%s.MRID, \"#\")\n", i, currentVar, field.Name)
		targetVar := fmt.Sprintf("target%d", i)
		fmt.Fprintf(&b, "\t\t%s, found%d := dataset.Elements[refID%d]\n", targetVar, i, i)
		fmt.Fprintf(&b, "\t\tif !found%d {\n\t\t\tcontinue\n\t\t}\n", i)

		isLast := i == len(pathSegs)-1
		if isLast {
			currentVar = targetVar
			currentType = nil
			continue
		}
		// Intermediate hop: derive the next class from the next segment's
		// class portion (the dot-separated prefix of its xml-tag form).
		nextRaw := pathSegs[i+1]
		nextSeg, ok := stripCIMPrefix(nextRaw)
		if !ok {
			nextSeg = nextRaw
		}
		parts := strings.SplitN(nextSeg, ".", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("chain step %d: next seg %q has no class.field", i, nextSeg)
		}
		nextClass := parts[0]
		factory, ok := cimgostructs.StructMap[nextClass]
		if !ok {
			return "", "", fmt.Errorf("chain step %d: target class %q has no Go struct", i, nextClass)
		}
		tVar := fmt.Sprintf("t%d", i)
		fmt.Fprintf(&b, "\t\t%s, ok%d := %s.(*cimgostructs.%s)\n", tVar, i, targetVar, nextClass)
		fmt.Fprintf(&b, "\t\tif !ok%d {\n\t\t\tcontinue\n\t\t}\n", i)
		currentVar = tVar
		currentType = reflect.TypeOf(factory()).Elem()
	}
	return strings.TrimRight(b.String(), "\n"), currentVar, nil
}

// refClassEqualCondition implements sh.HasValue along a forward chain ending
// in rdf.type: chase the chain, then require the final target's Go type to
// match the named class exactly.
func refClassEqualCondition(refSegs []string, startType reflect.Type, payload any) (string, string, error) {
	want, ok := payload.(string)
	if !ok {
		return "", "", fmt.Errorf("HasValue payload not a string")
	}
	want = strings.TrimPrefix(strings.TrimSuffix(want, ">"), "<")
	wantClass, ok := stripCIMPrefix(want)
	if !ok {
		return "", "", fmt.Errorf("HasValue rdf.type %q not in cim namespace", want)
	}
	if _, ok := cimgostructs.StructMap[wantClass]; !ok {
		return "", "", fmt.Errorf("HasValue rdf.type %q has no Go struct", wantClass)
	}
	guard, targetVar, err := walkForwardRefChain(refSegs, startType, "v")
	if err != nil {
		return "", "", err
	}
	guard += fmt.Sprintf("\n\t\t_, isWantedClass := %s.(*cimgostructs.%s)", targetVar, wantClass)
	return guard, "!isWantedClass", nil
}

// refClassInCondition implements sh.In along a forward chain ending in
// rdf.type: chase the chain, then require the final target's Go type to be
// one of the listed classes. Emitted as a `type switch` because Go interface
// values can't be keyed by their concrete type at the language level.
func refClassInCondition(refSegs []string, startType reflect.Type, payload any) (string, string, error) {
	if payload == nil {
		// Source TTL writes `sh:in ()` (empty list) — the simplifier
		// passes nil through. "No value is acceptable" is almost
		// certainly a TTL authoring error; skip the check rather than
		// emit a guaranteed-violating one.
		return "", "", fmt.Errorf("In payload is empty (likely TTL bug: empty `sh:in ()`)")
	}
	rawValues, ok := payload.([]any)
	if !ok {
		return "", "", fmt.Errorf("In payload not a list (got %T)", payload)
	}
	if len(rawValues) == 0 {
		return "", "", fmt.Errorf("In payload is an empty list (likely TTL bug)")
	}
	classes, err := classListFromValues(rawValues)
	if err != nil {
		return "", "", err
	}
	guard, targetVar, err := walkForwardRefChain(refSegs, startType, "v")
	if err != nil {
		return "", "", err
	}
	var b strings.Builder
	b.WriteString(guard)
	fmt.Fprintf(&b, "\n\t\tisAllowedClass := false\n\t\tswitch %s.(type) {", targetVar)
	for _, cls := range classes {
		fmt.Fprintf(&b, "\n\t\tcase *cimgostructs.%s:\n\t\t\tisAllowedClass = true", cls)
	}
	b.WriteString("\n\t\t}")
	return b.String(), "!isAllowedClass", nil
}

// classListFromValues parses a SHACL `Values` payload into a list of Go
// struct names, normalising the `<cim.Foo>`/`cim100.Foo` IRI forms and
// verifying each name resolves to a generated struct. Abstract base classes
// (not in StructMap directly) are expanded to their concrete subclasses so
// the resulting allow-set is exhaustive at the Go-type level. Shared by
// sh.In on rdf.type and sh.Or with all-Class shapes.
func classListFromValues(rawValues []any) ([]string, error) {
	seen := map[string]bool{}
	var classes []string
	add := func(cls string) {
		if !seen[cls] {
			seen[cls] = true
			classes = append(classes, cls)
		}
	}
	for _, v := range rawValues {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("rdf.type list contains non-string %v", v)
		}
		s = strings.TrimPrefix(strings.TrimSuffix(s, ">"), "<")
		cls, ok := stripCIMPrefix(s)
		if !ok {
			return nil, fmt.Errorf("rdf.type %q not in cim namespace", s)
		}
		if _, ok := cimgostructs.StructMap[cls]; ok {
			add(cls)
			continue
		}
		subs := concreteSubclassesEmbedding(cls)
		if len(subs) == 0 {
			return nil, fmt.Errorf("rdf.type %q has no Go struct", cls)
		}
		for _, s := range subs {
			add(s)
		}
	}
	sort.Strings(classes)
	return classes, nil
}

// orClassListFromShapes accepts the Or payload's `Shapes` list (a list of
// shape lists, each containing a single sh.ClassConstraintComponent) and
// returns the equivalent allowed-class list, or an error when the shape
// can't be reduced to a flat class disjunction. This is the only Or shape
// the generator handles — it captures the dominant "this reference must be
// one of these CIM classes" pattern.
func orClassListFromShapes(payload any) ([]string, error) {
	shapes, ok := payload.([]any)
	if !ok {
		return nil, fmt.Errorf("Or payload Shapes is not a list")
	}
	classes := make([]string, 0, len(shapes))
	for i, sh := range shapes {
		inner, ok := asConstraintList(sh)
		if !ok {
			return nil, fmt.Errorf("Or shape %d is not a constraint list", i)
		}
		if len(inner) != 1 {
			return nil, fmt.Errorf("Or shape %d has %d constraints (only single-Class shapes supported)", i, len(inner))
		}
		c := inner[0]
		if c.Component != "sh.ClassConstraintComponent" {
			return nil, fmt.Errorf("Or shape %d component %q is not Class", i, c.Component)
		}
		want, _ := c.Payload["Class"].(string)
		want = strings.TrimPrefix(strings.TrimSuffix(want, ">"), "<")
		cls, ok := stripCIMPrefix(want)
		if !ok {
			return nil, fmt.Errorf("Or shape %d Class %q not in cim namespace", i, want)
		}
		if _, ok := cimgostructs.StructMap[cls]; !ok {
			return nil, fmt.Errorf("Or shape %d Class %q has no Go struct", i, cls)
		}
		classes = append(classes, cls)
	}
	return classes, nil
}

// notClassFromShapeRef accepts the Not payload's `ShapeRef` list (expected
// to contain a single sh.ClassConstraintComponent) and returns the negated
// class name. Like Or, this only handles the simple Class-only shape.
func notClassFromShapeRef(payload any) (string, error) {
	shapes, ok := asConstraintList(payload)
	if !ok {
		return "", fmt.Errorf("Not payload ShapeRef is not a constraint list")
	}
	if len(shapes) != 1 {
		return "", fmt.Errorf("Not has %d constraints (only single-Class shape supported)", len(shapes))
	}
	c := shapes[0]
	if c.Component != "sh.ClassConstraintComponent" {
		return "", fmt.Errorf("Not component %q is not Class", c.Component)
	}
	want, _ := c.Payload["Class"].(string)
	want = strings.TrimPrefix(strings.TrimSuffix(want, ">"), "<")
	cls, ok := stripCIMPrefix(want)
	if !ok {
		return "", fmt.Errorf("Not Class %q not in cim namespace", want)
	}
	if _, ok := cimgostructs.StructMap[cls]; !ok {
		return "", fmt.Errorf("Not Class %q has no Go struct", cls)
	}
	return cls, nil
}

// sortByNameAndSig stable-sorts `slice` by `key(item)` first, with the JSON
// encoding of the whole item as a tie-breaker. SHACL inputs frequently carry
// multiple entries with the same primary key (same attribute name, same
// constraint component) that differ only in payload, so the JSON tail makes
// the order reproducible across runs without needing every caller to spell
// out a per-type secondary key. The slice is mutated in place.
func sortByNameAndSig[T any](slice []T, key func(T) string) {
	type entry struct {
		item T
		sig  string
	}
	entries := make([]entry, len(slice))
	for i, item := range slice {
		b, _ := json.Marshal(item)
		entries[i] = entry{item: item, sig: string(b)}
	}
	sort.SliceStable(entries, func(i, j int) bool {
		ki, kj := key(entries[i].item), key(entries[j].item)
		if ki != kj {
			return ki < kj
		}
		return entries[i].sig < entries[j].sig
	})
	for i, e := range entries {
		slice[i] = e.item
	}
}

// asConstraintList normalises a nested-shape payload value into a typed
// constraint slice. The TTL-direct loader leaves nested shapes as
// []shaclimport.ConstraintInfo; the JSON loader flattens them to []any of
// map[string]any with lowercase keys (per the json tags on ConstraintInfo).
// Both shapes are accepted so Or/Not work identically across loaders.
func asConstraintList(v any) ([]shaclimport.ConstraintInfo, bool) {
	if cs, ok := v.([]shaclimport.ConstraintInfo); ok {
		return cs, true
	}
	arr, ok := v.([]any)
	if !ok {
		return nil, false
	}
	out := make([]shaclimport.ConstraintInfo, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			return nil, false
		}
		comp, _ := m["component"].(string)
		pl, _ := m["payload"].(map[string]any)
		out = append(out, shaclimport.ConstraintInfo{Component: comp, Payload: pl})
	}
	return out, true
}

// inverseCountCheck emits the Prelude that builds an O(N) inverse-reference
// count map plus the loop-body condition that compares the count against
// `threshold` using `op` (the operator that *violates* the constraint —
// "==" for "must have ≥1", "<" for MinCount, ">" for MaxCount). When
// targetClasses has more than one entry the prelude dispatches over each
// concrete subclass — this is how abstract base classes (e.g.
// ExcitationSystemDynamics) are handled, since they're not in StructMap
// and have no instances of their own.
func inverseCountCheck(targetClasses []string, field reflect.StructField, op, threshold string) (string, string) {
	cond := fmt.Sprintf("inverseCounts[id] %s %s", op, threshold)
	if len(targetClasses) == 1 {
		prelude := fmt.Sprintf(`	inverseCounts := map[string]int{}
	for _, ref := range dataset.Elements {
		r, ok := ref.(*cimgostructs.%s)
		if !ok {
			continue
		}
		if r.%s == nil {
			continue
		}
		inverseCounts[strings.TrimPrefix(r.%s.MRID, "#")]++
	}`, targetClasses[0], field.Name, field.Name)
		return prelude, cond
	}
	var b strings.Builder
	b.WriteString("\tinverseCounts := map[string]int{}\n")
	b.WriteString("\tfor _, ref := range dataset.Elements {\n")
	b.WriteString("\t\tswitch r := ref.(type) {\n")
	for _, cls := range targetClasses {
		fmt.Fprintf(&b, "\t\tcase *cimgostructs.%s:\n", cls)
		fmt.Fprintf(&b, "\t\t\tif r.%s != nil {\n", field.Name)
		fmt.Fprintf(&b, "\t\t\t\tinverseCounts[strings.TrimPrefix(r.%s.MRID, \"#\")]++\n", field.Name)
		b.WriteString("\t\t\t}\n")
	}
	b.WriteString("\t\t}\n")
	b.WriteString("\t}")
	return b.String(), cond
}

// inverseHasEnumValueCheck emits the Prelude for a 2-segment inverse-then-
// forward HasValue check: scan the dataset once, flag every focus-node id
// that has at least one referrer (a *cimgostructs.<targetClasses[i]>) whose
// `refField` points back AND whose `valueField` (a *struct{URI string} enum
// field) carries the named enum constant. Violation is "no such referrer
// found". Multi-class targets dispatch over each concrete subclass.
func inverseHasEnumValueCheck(targetClasses []string, refField, valueField reflect.StructField, constIdent string) (string, string) {
	cond := "!hasEnumValue[id]"
	if len(targetClasses) == 1 {
		prelude := fmt.Sprintf(`	hasEnumValue := map[string]bool{}
	for _, ref := range dataset.Elements {
		r, ok := ref.(*cimgostructs.%s)
		if !ok {
			continue
		}
		if r.%s == nil || r.%s == nil {
			continue
		}
		if r.%s.URI != cimgostructs.%s {
			continue
		}
		hasEnumValue[strings.TrimPrefix(r.%s.MRID, "#")] = true
	}`, targetClasses[0], refField.Name, valueField.Name, valueField.Name, constIdent, refField.Name)
		return prelude, cond
	}
	var b strings.Builder
	b.WriteString("\thasEnumValue := map[string]bool{}\n")
	b.WriteString("\tfor _, ref := range dataset.Elements {\n")
	b.WriteString("\t\tswitch r := ref.(type) {\n")
	for _, cls := range targetClasses {
		fmt.Fprintf(&b, "\t\tcase *cimgostructs.%s:\n", cls)
		fmt.Fprintf(&b, "\t\t\tif r.%s == nil || r.%s == nil {\n", refField.Name, valueField.Name)
		b.WriteString("\t\t\t\tcontinue\n")
		b.WriteString("\t\t\t}\n")
		fmt.Fprintf(&b, "\t\t\tif r.%s.URI != cimgostructs.%s {\n", valueField.Name, constIdent)
		b.WriteString("\t\t\t\tcontinue\n")
		b.WriteString("\t\t\t}\n")
		fmt.Fprintf(&b, "\t\t\thasEnumValue[strings.TrimPrefix(r.%s.MRID, \"#\")] = true\n", refField.Name)
	}
	b.WriteString("\t\t}\n")
	b.WriteString("\t}")
	return b.String(), cond
}

// concreteSubclassesEmbedding returns the sorted list of concrete cimgostructs
// class names (those present in StructMap) whose Go type embeds — directly or
// transitively — an anonymous struct named `abstract`. This is how the
// generator handles SHACL inverse paths that target abstract base classes
// like ExcitationSystemDynamics: those classes have no instances of their
// own, so the generated check has to dispatch over every concrete subclass.
func concreteSubclassesEmbedding(abstract string) []string {
	var matches []string
	for name, factory := range cimgostructs.StructMap {
		t := reflect.TypeOf(factory()).Elem()
		if typeEmbedsName(t, abstract) {
			matches = append(matches, name)
		}
	}
	sort.Strings(matches)
	return matches
}

// classNameFromPayload normalises a SHACL `sh:class` payload (an IRI in
// either `<cim.X>` or `cim.X` form) to a bare Go struct name. Returns false
// if the payload isn't a string in the cim namespace.
func classNameFromPayload(payload any) (string, bool) {
	s, ok := payload.(string)
	if !ok {
		return "", false
	}
	s = strings.TrimPrefix(strings.TrimSuffix(s, ">"), "<")
	return stripCIMPrefix(s)
}

// isClassOrAncestor reports whether the class named `class` satisfies a SHACL
// `sh:class asserted` requirement structurally — i.e. either class == asserted
// or class embeds asserted (Go's encoding of subclass-of, transitively). Used
// to short-circuit inverse-path Class constraints when every concrete target
// type is already a subtype of the asserted class.
func isClassOrAncestor(class, asserted string) bool {
	if class == asserted {
		return true
	}
	factory, ok := cimgostructs.StructMap[class]
	if !ok {
		return false
	}
	return typeEmbedsName(reflect.TypeOf(factory()).Elem(), asserted)
}

func typeEmbedsName(t reflect.Type, name string) bool {
	if t.Kind() != reflect.Struct {
		return false
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.Anonymous {
			continue
		}
		if f.Type.Name() == name {
			return true
		}
		if f.Type.Kind() == reflect.Struct && typeEmbedsName(f.Type, name) {
			return true
		}
	}
	return false
}

func zeroLiteralFor(k reflect.Kind) string {
	switch k {
	case reflect.String:
		return `""`
	case reflect.Bool:
		return "false"
	default:
		return "0"
	}
}

func anyToFloat(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case string:
		var f float64
		fmt.Sscanf(x, "%f", &f)
		return f
	}
	return 0
}

func anyToFloatLiteral(v any) string {
	switch x := v.(type) {
	case float64:
		return fmt.Sprintf("%v", x)
	case string:
		return strings.Trim(x, "\"")
	default:
		return fmt.Sprintf("%v", x)
	}
}

func simpleClassName(s string) (string, bool) {
	name, ok := stripCIMPrefix(s)
	if !ok {
		return "", false
	}
	if strings.ContainsAny(name, "./ ") {
		return "", false
	}
	return name, true
}

// stripCIMPrefix removes a leading "cim." or "cim<version>." (e.g. "cim100.")
// segment from a SHACL identifier. Returns the remainder and true on match.
// CGMES profiles intermix CIM 16/17 (no version suffix or "cim16") and CIM 100
// (CGMES 3.0) names, but cimgostructs has a single Go struct per class
// regardless of CIM version, so the prefix is irrelevant for code generation.
func stripCIMPrefix(s string) (string, bool) {
	if !strings.HasPrefix(s, "cim") {
		return "", false
	}
	rest := s[len("cim"):]
	for len(rest) > 0 && rest[0] >= '0' && rest[0] <= '9' {
		rest = rest[1:]
	}
	if len(rest) == 0 || rest[0] != '.' {
		return "", false
	}
	return rest[1:], true
}

// xmlTagExistsOnAnyStruct returns true if any concrete struct in StructMap has
// a field whose XML tag local name matches tag. Used by pairCompare to
// distinguish genuine TTL typos (tag missing from the schema entirely) from
// cross-class comparisons (tag exists on a sibling type, making the
// constraint vacuously satisfied for the current target).
func xmlTagExistsOnAnyStruct(tag string) bool {
	for _, factory := range cimgostructs.StructMap {
		t := reflect.TypeOf(factory()).Elem()
		if _, ok := findFieldByXMLTag(t, tag); ok {
			return true
		}
	}
	return false
}

func findFieldByXMLTag(t reflect.Type, tag string) (reflect.StructField, bool) {
	if t.Kind() != reflect.Struct {
		return reflect.StructField{}, false
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if got := xmlLocal(f.Tag.Get("xml")); got != "" && got == tag {
			return f, true
		}
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			if sub, ok := findFieldByXMLTag(f.Type, tag); ok {
				return sub, true
			}
		}
	}
	return reflect.StructField{}, false
}

func xmlLocal(tag string) string {
	if tag == "" {
		return ""
	}
	if i := strings.IndexByte(tag, ','); i >= 0 {
		tag = tag[:i]
	}
	return tag
}

func camelize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func camelCaseFromStem(stem string) string {
	parts := strings.Split(stem, "_")
	out := strings.Builder{}
	for _, p := range parts {
		if p == "" {
			continue
		}
		out.WriteString(camelize(p))
	}
	return out.String()
}

// profileStem reduces a SHACL file name to a stable lowercase identifier
// suitable for both filenames and Go identifiers. Format:
//
//	<profile>_<version>[_simple|_complex][_<variant>]
//
// Examples:
//
//	61970-301_DiagramLayout-AP-Con-Complex-SHACL                       -> diagramlayout_61970_301_complex
//	61970-301_DiagramLayout-AP-Con-Complex-NotSolvedMAS-SHACL          -> diagramlayout_61970_301_complex_notsolvedmas
//	61970-552-Header-AP-Con-Simple-SHACL                               -> header_61970_552_simple
//	61970-600-2_IdentifiedObjectCommon_AP-Con-Complex-SHACL            -> identifiedobjectcommon_61970_600_2_complex
//	61970-456_StateVariables-AP-Con-Complex-Explicit-CrossProfile-SHACL -> statevariables_61970_456_complex_explicit_crossprofile
//
// All four parts are needed for uniqueness — same profile name appears across
// CIM revisions (61970-301 vs 61970-600-2 etc.) and across Simple/Complex
// shape vocabularies, and the JSON files distinguish them.
func profileStem(fileName string) string {
	name := strings.TrimSuffix(fileName, "-SHACL")

	// Variants hide between the profile name and the "-SHACL" tail. Detect and
	// strip them first so the "-AP-" cut below sees a clean tail.
	// Order matters: longer/more specific keys must come before shorter ones
	// that would otherwise match a prefix (e.g. -Explicit-CrossProfile must be
	// tried before the bare -CrossProfile).
	variants := []struct{ key, suffix string }{
		{"-NotSolvedMAS", "_notsolvedmas"},
		{"-SolvedMAS", "_solvedmas"},
		{"-Explicit-CrossProfile", "_explicit_crossprofile"},
		{"-Implicit-CrossProfile", "_implicit_crossprofile"},
		{"-CrossProfile", "_crossprofile"},
		{"-InverseAssociation", "_inverseassociation"},
	}
	variantSuffix := ""
	for _, v := range variants {
		if strings.Contains(name, v.key) {
			name = strings.ReplaceAll(name, v.key, "")
			variantSuffix = v.suffix
			break
		}
	}

	// Pull out the AP-Con-Simple/Complex marker and remove the metadata tail.
	mode := ""
	for _, m := range []struct{ key, sfx string }{
		{"-AP-Con-Simple", "_simple"},
		{"-AP-Con-Complex", "_complex"},
		{"_AP-Con-Simple", "_simple"},
		{"_AP-Con-Complex", "_complex"},
	} {
		if i := strings.Index(name, m.key); i >= 0 {
			mode = m.sfx
			name = name[:i]
			break
		}
	}
	// Anything left over (rare bare "_AP-" / "-AP-") gets cut blind.
	for _, sep := range []string{"-AP-", "_AP-"} {
		if i := strings.Index(name, sep); i >= 0 {
			name = name[:i]
			break
		}
	}

	// Split off the profile name (last "_"- or "-"-separated segment) from the
	// version prefix. For "61970-301_DiagramLayout" the split is at "_" giving
	// version "61970-301" + profile "DiagramLayout"; for "61970-552-Header"
	// the last "-" gives version "61970-552" + profile "Header".
	version := ""
	profile := name
	if i := strings.LastIndexAny(name, "_-"); i >= 0 {
		version = name[:i]
		profile = name[i+1:]
	}
	versionPart := ""
	if version != "" {
		versionPart = "_" + sanitizeIdent(version)
	}
	return strings.ToLower(profile) + versionPart + mode + variantSuffix
}

// sanitizeIdent lowercases and replaces non-identifier separators with "_" so
// the result is safe to embed in a Go identifier.
func sanitizeIdent(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}
