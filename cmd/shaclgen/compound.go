package main

import (
	"cimgo/cimstructs"
	"cimgo/shaclimport"
	"fmt"
	"reflect"
	"strings"
)

// compoundResult holds the generated code for a compound constraint (sh:or,
// sh:and, sh:xone with no top-level sh:path). Prelude holds inverse-count map
// builds executed before the per-element loop; Condition is the violation
// expression inside the loop. Imports lists any extra Go imports needed.
type compoundResult struct {
	Prelude   string // code before the per-element loop (e.g. inverse count maps)
	Guard     string // code inside the per-element loop, before the condition
	Condition string
	Imports   []string
}

// buildCompoundCheck dispatches sh:Or/And/Xone node constraints (c.Path == nil)
// to specific handlers. Each handler reduces the compound to a Prelude + Condition
// that fit the existing template architecture.
func buildCompoundCheck(c shaclimport.ConstraintInfo, structType reflect.Type) (compoundResult, error) {
	shapes, ok := c.Payload["Shapes"].([]any)
	if !ok {
		return compoundResult{}, fmt.Errorf("%s: Shapes payload not a list", c.Component)
	}
	subShapes := make([][]shaclimport.ConstraintInfo, 0, len(shapes))
	for i, s := range shapes {
		cs, ok := asConstraintList(s)
		if !ok {
			return compoundResult{}, fmt.Errorf("%s: sub-shape %d is not a constraint list", c.Component, i)
		}
		subShapes = append(subShapes, cs)
	}

	switch c.Component {
	case "sh:XoneConstraintComponent":
		return xoneCheck(subShapes, structType)
	case "sh:OrConstraintComponent":
		return orCheck(subShapes, structType)
	case "sh:AndConstraintComponent":
		return andCheck(subShapes, structType)
	}
	return compoundResult{}, fmt.Errorf("unsupported compound component: %s", c.Component)
}

// xoneCheck handles sh:Xone where each sub-shape is a single sh:MinCountConstraintComponent
// on a forward pointer path (minCount=1 → field != nil). The violation fires
// when not exactly one sub-shape passes.
func xoneCheck(subShapes [][]shaclimport.ConstraintInfo, structType reflect.Type) (compoundResult, error) {
	type branch struct {
		field reflect.StructField
	}
	var branches []branch
	for i, cs := range subShapes {
		if len(cs) != 1 || cs[0].Component != "sh:MinCountConstraintComponent" || len(cs[0].Path) != 1 {
			return compoundResult{}, fmt.Errorf("sh:Xone sub-shape %d: expected single MinCount on a forward path", i)
		}
		if int(shaclimport.AnyToFloat(cs[0].Payload["MinCount"])) != 1 {
			return compoundResult{}, fmt.Errorf("sh:Xone sub-shape %d: MinCount is not 1", i)
		}
		seg, ok := stripCIMPrefix(cs[0].Path[0])
		if !ok {
			return compoundResult{}, fmt.Errorf("sh:Xone sub-shape %d: path %q not in cim namespace", i, cs[0].Path[0])
		}
		f, ok := findFieldByXMLTag(structType, seg)
		if !ok {
			return compoundResult{}, fmt.Errorf("sh:Xone sub-shape %d: no field with xml tag %q", i, seg)
		}
		if f.Type.Kind() != reflect.Pointer {
			return compoundResult{}, fmt.Errorf("sh:Xone sub-shape %d: field %q is %s, expected pointer", i, f.Name, f.Type.Kind())
		}
		branches = append(branches, branch{f})
	}
	if len(branches) < 2 {
		return compoundResult{}, fmt.Errorf("sh:Xone: need at least 2 branches, got %d", len(branches))
	}

	// Violation: not exactly one branch has a non-nil field.
	// Emit into Guard (inside per-element loop, accesses v).
	var b strings.Builder
	b.WriteString("\t\tpassCount := 0\n")
	for _, br := range branches {
		fmt.Fprintf(&b, "\t\tif v.%s != nil {\n\t\t\tpassCount++\n\t\t}\n", br.field.Name)
	}
	return compoundResult{Guard: b.String(), Condition: "passCount != 1"}, nil
}

// orCheck handles sh:Or where each sub-shape is a set of MinCount/MaxCount
// constraints on the same inverse path. Violation fires when no branch satisfies
// its count range (i.e., all branches fail).
func orCheck(subShapes [][]shaclimport.ConstraintInfo, structType reflect.Type) (compoundResult, error) {
	type branchCounts struct {
		mapVar   string
		prelude  string
		failCond string
	}
	var branches []branchCounts
	var allImports []string
	var preludeB strings.Builder

	for i, cs := range subShapes {
		min, max, path, err := extractInverseCountConstraints(cs, i)
		if err != nil {
			return compoundResult{}, fmt.Errorf("sh:Or sub-shape %d: %w", i, err)
		}
		mapVar := fmt.Sprintf("counts%d", i)
		prelude, err := buildInverseCountPrelude(mapVar, path, structType)
		if err != nil {
			return compoundResult{}, fmt.Errorf("sh:Or sub-shape %d: %w", i, err)
		}
		preludeB.WriteString(prelude)
		failCond := inverseCountFailCond(mapVar, min, max)
		branches = append(branches, branchCounts{mapVar, prelude, failCond})
		allImports = append(allImports, "strings")
	}

	// Violation: ALL branches fail (OR → pass if at least one passes).
	var parts []string
	for _, br := range branches {
		parts = append(parts, br.failCond)
	}
	return compoundResult{
		Prelude:   preludeB.String(),
		Condition: strings.Join(parts, " && "),
		Imports:   dedupeStrings(allImports),
	}, nil
}

// andCheck handles sh:And where each sub-shape is either:
//   - a single MaxCount=0 on a forward field (field must be absent), or
//   - a MinCount constraint on an inverse path (must have N referrers), or
//   - a set of MinCount+MaxCount on an inverse path.
//
// Violation fires when ANY branch fails (AND → pass if all pass).
func andCheck(subShapes [][]shaclimport.ConstraintInfo, structType reflect.Type) (compoundResult, error) {
	var preludeB strings.Builder
	var failConds []string
	var allImports []string

	for i, cs := range subShapes {
		// Try forward MaxCount=0 first (field must be absent).
		if len(cs) == 1 && cs[0].Component == "sh:MaxCountConstraintComponent" && len(cs[0].Path) == 1 {
			if !strings.HasPrefix(cs[0].Path[0], "^") {
				max := int(shaclimport.AnyToFloat(cs[0].Payload["MaxCount"]))
				if max != 0 {
					return compoundResult{}, fmt.Errorf("sh:And sub-shape %d: forward MaxCount=%d (only 0 supported in and)", i, max)
				}
				seg, ok := stripCIMPrefix(cs[0].Path[0])
				if !ok {
					return compoundResult{}, fmt.Errorf("sh:And sub-shape %d: path not in cim namespace", i)
				}
				f, ok := findFieldByXMLTag(structType, seg)
				if !ok {
					return compoundResult{}, fmt.Errorf("sh:And sub-shape %d: no field with xml tag %q", i, seg)
				}
				_, cond, err := maxCountCondition(f, cs[0].Payload["MaxCount"])
				if err != nil {
					return compoundResult{}, fmt.Errorf("sh:And sub-shape %d: %w", i, err)
				}
				failConds = append(failConds, cond)
				continue
			}
		}

		// Try forward MinCount=0+MaxCount=0 collapsed to required-absent (both must be 0).
		if len(cs) == 2 {
			allForwardMaxZero := true
			var subConds []string
			for _, c := range cs {
				if len(c.Path) != 1 || strings.HasPrefix(c.Path[0], "^") {
					allForwardMaxZero = false
					break
				}
				if c.Component == "sh:MaxCountConstraintComponent" && int(shaclimport.AnyToFloat(c.Payload["MaxCount"])) == 0 {
					seg, ok := stripCIMPrefix(c.Path[0])
					if !ok {
						allForwardMaxZero = false
						break
					}
					f, ok := findFieldByXMLTag(structType, seg)
					if !ok {
						allForwardMaxZero = false
						break
					}
					_, cond, err := maxCountCondition(f, c.Payload["MaxCount"])
					if err != nil {
						allForwardMaxZero = false
						break
					}
					subConds = append(subConds, cond)
				} else if c.Component == "sh:MinCountConstraintComponent" && int(shaclimport.AnyToFloat(c.Payload["MinCount"])) == 0 {
					// minCount=0 is vacuously true — ignore
				} else {
					allForwardMaxZero = false
					break
				}
			}
			if allForwardMaxZero && len(subConds) > 0 {
				failConds = append(failConds, strings.Join(subConds, " || "))
				continue
			}
		}

		// Try forward MinCount=1+MaxCount=1 (Required) — check if field is float/bool.
		if len(cs) == 2 {
			allRequired := true
			for _, c := range cs {
				isMin1 := c.Component == "sh:MinCountConstraintComponent" && int(shaclimport.AnyToFloat(c.Payload["MinCount"])) == 1
				isMax1 := c.Component == "sh:MaxCountConstraintComponent" && int(shaclimport.AnyToFloat(c.Payload["MaxCount"])) == 1
				if !(isMin1 || isMax1) || len(c.Path) != 1 || strings.HasPrefix(c.Path[0], "^") {
					allRequired = false
					break
				}
			}
			if allRequired {
				seg, ok := stripCIMPrefix(cs[0].Path[0])
				if !ok {
					return compoundResult{}, fmt.Errorf("sh:And sub-shape %d: path not in cim namespace", i)
				}
				f, ok := findFieldByXMLTag(structType, seg)
				if !ok {
					return compoundResult{}, fmt.Errorf("sh:And sub-shape %d: no field %q", i, seg)
				}
				_, err := requiredCondition(f)
				if err != nil {
					return compoundResult{}, fmt.Errorf("sh:And sub-shape %d: %w", i, err)
				}
				cond := fmt.Sprintf("v.%s == %s", f.Name, zeroLiteralFor(f.Type.Kind()))
				failConds = append(failConds, cond)
				continue
			}
		}

		// Try inverse path with MinCount+MaxCount (single or multi-segment).
		// Multi-segment: first seg is inverse, rest are forward hops.
		min, max, path, err := extractInverseCountConstraints(cs, i)
		if err != nil {
			// Try single MinCount on inverse path.
			if len(cs) == 1 && cs[0].Component == "sh:MinCountConstraintComponent" && len(cs[0].Path) == 1 && strings.HasPrefix(cs[0].Path[0], "^") {
				min = int(shaclimport.AnyToFloat(cs[0].Payload["MinCount"]))
				max = 1<<31 - 1 // no upper bound
				path = cs[0].Path[0]
				err = nil
			}
			// Try multi-segment where all constraints share the same multi-seg path.
			if err != nil && len(cs) >= 1 && len(cs[0].Path) > 1 && strings.HasPrefix(cs[0].Path[0], "^") {
				allSamePath := true
				for _, c := range cs {
					if len(c.Path) != len(cs[0].Path) {
						allSamePath = false
						break
					}
					for k, seg := range c.Path {
						if seg != cs[0].Path[k] {
							allSamePath = false
							break
						}
					}
				}
				if allSamePath {
					min, max = 0, 1<<31-1
					for _, c := range cs {
						switch c.Component {
						case "sh:MinCountConstraintComponent":
							min = int(shaclimport.AnyToFloat(c.Payload["MinCount"]))
						case "sh:MaxCountConstraintComponent":
							max = int(shaclimport.AnyToFloat(c.Payload["MaxCount"]))
						}
					}
					path = cs[0].Path[0]
					err = nil
				}
			}
			if err != nil {
				return compoundResult{}, fmt.Errorf("sh:And sub-shape %d: %w", i, err)
			}
		}

		// Extra forward hops after the first inverse segment.
		var forwardSegs []string
		if len(cs[0].Path) > 1 {
			forwardSegs = cs[0].Path[1:]
		}

		mapVar := fmt.Sprintf("andCounts%d", i)
		prelude, buildErr := buildInverseCountPrelude(mapVar, path, structType, forwardSegs...)
		if buildErr != nil {
			return compoundResult{}, fmt.Errorf("sh:And sub-shape %d: %w", i, buildErr)
		}
		preludeB.WriteString(prelude)
		allImports = append(allImports, "strings")

		var failCond string
		if max == 1<<31-1 {
			failCond = fmt.Sprintf("%s[id] < %d", mapVar, min)
		} else {
			failCond = inverseCountFailCond(mapVar, min, max)
		}
		failConds = append(failConds, failCond)
	}

	if len(failConds) == 0 {
		return compoundResult{}, fmt.Errorf("sh:And: no conditions generated")
	}
	return compoundResult{
		Prelude:   preludeB.String(),
		Condition: strings.Join(failConds, " || "),
		Imports:   dedupeStrings(allImports),
	}, nil
}

// extractInverseCountConstraints extracts (min, max, inversePath) from a sub-shape
// that consists of a MinCount and a MaxCount on the same inverse path.
func extractInverseCountConstraints(cs []shaclimport.ConstraintInfo, idx int) (min, max int, path string, err error) {
	if len(cs) < 1 || len(cs) > 2 {
		return 0, 0, "", fmt.Errorf("sub-shape %d: expected 1-2 constraints, got %d", idx, len(cs))
	}
	min, max = 0, 1<<31-1
	for _, c := range cs {
		if len(c.Path) != 1 || !strings.HasPrefix(c.Path[0], "^") {
			return 0, 0, "", fmt.Errorf("sub-shape %d: expected single inverse path", idx)
		}
		if path == "" {
			path = c.Path[0]
		} else if path != c.Path[0] {
			return 0, 0, "", fmt.Errorf("sub-shape %d: mixed paths", idx)
		}
		switch c.Component {
		case "sh:MinCountConstraintComponent":
			min = int(shaclimport.AnyToFloat(c.Payload["MinCount"]))
		case "sh:MaxCountConstraintComponent":
			max = int(shaclimport.AnyToFloat(c.Payload["MaxCount"]))
		default:
			return 0, 0, "", fmt.Errorf("sub-shape %d: unsupported component %s", idx, c.Component)
		}
	}
	return min, max, path, nil
}

// buildInverseCountPrelude generates a Prelude that builds a map[string]int
// counting references along a path. path is the first (inverse) segment like
// "^cim:Terminal.ConductingEquipment"; forwardSegs are optional additional
// forward hops like ["cim:Terminal.ConnectivityNode"] that must all be non-nil
// for the count to increment.
func buildInverseCountPrelude(mapVar, path string, structType reflect.Type, forwardSegs ...string) (string, error) {
	rawForward := strings.TrimPrefix(path, "^")
	seg, ok := stripCIMPrefix(rawForward)
	if !ok {
		seg = rawForward
	}
	// The forward field lives on the REFERRER class (derived from the path segment).
	parts := strings.SplitN(seg, ".", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("inverse path %q has no class.field shape", seg)
	}
	referrerClass := parts[0]
	factory, ok := cimstructs.StructMap[referrerClass]
	if !ok {
		return "", fmt.Errorf("referrer class %q has no Go struct", referrerClass)
	}
	referrerType := reflect.TypeOf(factory()).Elem()
	field, ok := findFieldByXMLTag(referrerType, seg)
	if !ok {
		return "", fmt.Errorf("no field with xml tag %q on %s", seg, referrerClass)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "\t%s := map[string]int{}\n", mapVar)
	b.WriteString("\tfor _, ref := range dataset.ByID {\n")
	fmt.Fprintf(&b, "\t\tr, ok := ref.(*cimstructs.%s)\n", referrerClass)
	b.WriteString("\t\tif !ok {\n\t\t\tcontinue\n\t\t}\n")

	switch field.Type.Kind() {
	case reflect.Pointer:
		fmt.Fprintf(&b, "\t\tif r.%s == nil {\n\t\t\tcontinue\n\t\t}\n", field.Name)
		// Resolve any additional forward hops on the referrer object.
		currentVar := "r"
		currentType := referrerType
		for j, fwdRaw := range forwardSegs {
			fwdSeg, _ := stripCIMPrefix(fwdRaw)
			fwdField, ok := findFieldByXMLTag(currentType, fwdSeg)
			if !ok {
				return "", fmt.Errorf("forward hop %d: no field with xml tag %q on %s", j, fwdSeg, currentType.Name())
			}
			if fwdField.Type.Kind() != reflect.Pointer {
				return "", fmt.Errorf("forward hop %d: field %q is %s, expected pointer", j, fwdField.Name, fwdField.Type.Kind())
			}
			fmt.Fprintf(&b, "\t\tif %s.%s == nil {\n\t\t\tcontinue\n\t\t}\n", currentVar, fwdField.Name)
			nextVar := fmt.Sprintf("fwd%d", j)
			// We only need to nil-check; we don't need to dereference for counting.
			// Just use the current variable check — no need to type-assert since
			// we only care whether the pointer is set.
			_ = nextVar
			currentVar = currentVar + "." + fwdField.Name
			factory2, ok2 := cimstructs.StructMap[fwdField.Type.Elem().Name()]
			if ok2 {
				currentType = reflect.TypeOf(factory2()).Elem()
			}
		}
		fmt.Fprintf(&b, "\t\t%s[strings.TrimPrefix(r.%s.MRID, \"#\")]++\n", mapVar, field.Name)
	case reflect.Slice:
		fmt.Fprintf(&b, "\t\tfor _, entry := range r.%s {\n", field.Name)
		fmt.Fprintf(&b, "\t\t\t%s[strings.TrimPrefix(entry.MRID, \"#\")]++\n", mapVar)
		b.WriteString("\t\t}\n")
	default:
		return "", fmt.Errorf("field %q on %s is %s, expected pointer or slice", field.Name, referrerClass, field.Type.Kind())
	}

	b.WriteString("\t}\n")
	return b.String(), nil
}

// inverseCountFailCond returns the condition expression for "inverse count does NOT
// satisfy [min, max]".
func inverseCountFailCond(mapVar string, min, max int) string {
	if min == max {
		return fmt.Sprintf("%s[id] != %d", mapVar, min)
	}
	if min == 0 {
		return fmt.Sprintf("%s[id] > %d", mapVar, max)
	}
	if max == 1<<31-1 {
		return fmt.Sprintf("%s[id] < %d", mapVar, min)
	}
	return fmt.Sprintf("%s[id] < %d || %s[id] > %d", mapVar, min, mapVar, max)
}

func dedupeStrings(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
