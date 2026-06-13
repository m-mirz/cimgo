package main

import (
	"cimgo/cimstructs"
	"cimgo/shaclimport"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

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
		if _, ok := cimstructs.StructMap[structName]; ok {
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

func buildFileSpec(pkg string, fr *shaclimport.FileResults) (fileSpec, []skipEntry) {
	stem := profileStem(fr.FileName)
	stemCamel := camelCaseFromStem(stem)
	spec := fileSpec{
		FileName:         fr.FileName,
		Pkg:              pkg,
		OrchestratorName: "ValidateGenerated" + stemCamel + "Profile",
	}
	var skipEntries []skipEntry
	skipIndex := map[string]int{}
	used := map[string]int{}
	importSet := map[string]struct{}{
		"cimgo/cimstructs": {},
		"cimgo/shaclmodel": {},
	}

	var processShape func(shape shaclimport.ShapeInfo, currentClasses []string)
	processShape = func(shape shaclimport.ShapeInfo, currentClasses []string) {
		concreteNames := resolveConcreteClasses(shape.Targets)
		if len(concreteNames) > 0 {
			currentClasses = concreteNames
		}

		if len(currentClasses) > 0 {
			for _, concrete := range currentClasses {
				factory := cimstructs.StructMap[concrete]
				structType := reflect.TypeOf(factory()).Elem()

				constraints := append([]shaclimport.ConstraintInfo(nil), shape.Constraints...)
				sortByNameAndSig(constraints, func(c shaclimport.ConstraintInfo) string { return c.Component })

				for _, c := range constraints {
					cs, imports, err := buildCheckSpec(stemCamel, concrete, shape.ID, structType, c, used)
					if err != nil {
						prop := ""
						if len(c.Path) > 0 {
							prop = "." + strings.Join(c.Path, "/")
						}
						key := prop + "\x00" + c.Component + "\x00" + c.Name
						if i, ok := skipIndex[key]; ok {
							skipEntries[i].Classes = append(skipEntries[i].Classes, concrete)
						} else {
							skipIndex[key] = len(skipEntries)
							skipEntries = append(skipEntries, skipEntry{
								Classes:   []string{concrete},
								Prop:      prop,
								Component: c.Component,
								Name:      c.Name,
								Reason:    err.Error(),
							})
						}
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
	return spec, skipEntries
}

func buildCheckSpec(stemCamel, structName, shapeID string, structType reflect.Type, c shaclimport.ConstraintInfo, used map[string]int) (checkSpec, []string, error) {
	// Detect inverse and multi-segment paths up front. Inverse paths
	// (`^cim:X.Y`) flip the constraint sense from "look at this object's
	// field" to "scan the dataset for objects pointing at this one". Some
	// multi-segment shapes ([ref, rdf:type]) are recognised as a class-of-
	// referenced-object check; everything else is currently skipped.
	if len(c.Path) == 0 {
		// Compound constraints (sh:Or, sh:And, sh:Xone) carry no top-level path;
		// their sub-shapes each have their own paths.
		switch c.Component {
		case "sh:OrConstraintComponent", "sh:AndConstraintComponent", "sh:XoneConstraintComponent":
			result, err := buildCompoundCheck(c, structType)
			if err != nil {
				return checkSpec{}, nil, err
			}
			cs := checkSpec{
				Name:        "Check" + stemCamel + structName + camelize(strings.TrimPrefix(strings.TrimSuffix(c.Component, "ConstraintComponent"), "sh:")),
				ShapeID:     shapeID,
				RuleID:      shapeID,
				RuleName:    c.Name,
				Description: c.Description,
				Class:       structName,
				Tag:         c.Component,
				Component:   c.Component,
				Property:    c.Component,
				PathKey:     "",
				Message:     strings.Trim(c.Message, "\""),
				Severity:    c.Severity,
				Prelude:     result.Prelude,
				Guard:       result.Guard,
				Condition:   result.Condition,
			}
			if cs.Severity == "" {
				cs.Severity = "sh:Violation"
			}
			// Only bind v when the generated code actually uses it.
			if !strings.Contains(cs.Guard, "v.") && !strings.Contains(cs.Condition, "v.") {
				cs.NoV = true
			}
			used[cs.Name]++
			if used[cs.Name] > 1 {
				cs.Name = fmt.Sprintf("%s_%d", cs.Name, used[cs.Name])
			}
			imports := result.Imports
			return cs, imports, nil
		}
		return checkSpec{}, nil, fmt.Errorf("empty path")
	}

	rawPath := c.Path[0]
	isInverse := strings.HasPrefix(rawPath, "^")
	if isInverse {
		rawPath = rawPath[1:]
	}

	// `^rdf:type` is a dataset-level cardinality check: "count of instances
	// whose rdf:type is the focus class". MinCount=N → at least N instances
	// must exist; MaxCount=N → at most N. Handled here before generic
	// inverse-path machinery, which would try to parse "rdf:type" as a
	// class.field and fail with "no Go struct rdf".
	if isInverse && len(c.Path) == 1 && rawPath == "rdf:type" {
		return buildDatasetCardinalityCheck(stemCamel, structName, shapeID, c, used)
	}

	// Classify multi-segment paths. The dominant shape is a forward chain
	// ending in `rdf:type` (657 of 669 multi-segment HasValue/In); we also
	// accept a forward chain *without* the trailing rdf:type for Or, where
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
			if c.Path[len(c.Path)-1] == "rdf:type" {
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
	// targetClasses is the list of concrete cimstructs class names to
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
		if factory, ok := cimstructs.StructMap[targetClass]; ok {
			lookupType = reflect.TypeOf(factory()).Elem()
			targetClasses = []string{targetClass}
		} else {
			subclasses := concreteSubclassesEmbedding(targetClass)
			if len(subclasses) == 0 {
				return checkSpec{}, nil, fmt.Errorf("inverse target class %q has no Go struct", targetClass)
			}
			// Use the first subclass's reflect.Type for field lookup —
			// the abstract field is reachable via embedded promotion.
			firstFactory := cimstructs.StructMap[subclasses[0]]
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

	// sh:NodeKind on a path ending in rdf:type is structurally satisfied:
	// Go's static type system already enforces literal/IRI/blank-node
	// distinctions at compile time. A pointer-to-struct field with no MRID
	// is a blank node by construction, an MRID-bearing reference is an IRI,
	// and a primitive field is a literal — none of which can be violated
	// at runtime.
	if c.Component == "sh:NodeKindConstraintComponent" && len(c.Path) >= 1 && c.Path[len(c.Path)-1] == "rdf:type" {
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
		severity = "sh:Violation"
	}

	ruleID := shapeID

	cs := checkSpec{
		Name:        name,
		ShapeID:     shapeID,
		RuleID:      ruleID,
		RuleName:    c.Name,
		Description: c.Description,
		Class:       structName,
		Tag:         tag,
		Component:   c.Component,
		Property:    tag,
		PathKey:     strings.Join(c.Path, "/"),
		Message:     strings.Trim(c.Message, "\""),
		Severity:    severity,
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
			if c.Component != "sh:HasValueConstraintComponent" {
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
		case "sh:RequiredConstraintComponent":
			prelude, cond := inverseCountCheck(targetClasses, field, "==", "0")
			cs.Prelude, cs.Condition = prelude, cond
			cs.NoV = true
			imports = append(imports, "strings")
		case "sh:MinCountConstraintComponent":
			min := int(shaclimport.AnyToFloat(c.Payload["MinCount"]))
			prelude, cond := inverseCountCheck(targetClasses, field, "<", fmt.Sprintf("%d", min))
			cs.Prelude, cs.Condition = prelude, cond
			cs.NoV = true
			imports = append(imports, "strings")
		case "sh:MaxCountConstraintComponent":
			max := int(shaclimport.AnyToFloat(c.Payload["MaxCount"]))
			prelude, cond := inverseCountCheck(targetClasses, field, ">", fmt.Sprintf("%d", max))
			cs.Prelude, cs.Condition = prelude, cond
			cs.NoV = true
			imports = append(imports, "strings")
		case "sh:ClassConstraintComponent":
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
				return checkSpec{}, nil, fmt.Errorf("inverse Class payload not a cim:* string")
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
	// "forwardChainOnly only handles Or" / "ends-rdf:type only handles
	// HasValue/In/Required" rejections and get reported as plain
	// "not supported" — which understates how many are actually OK.
	if len(c.Path) > 1 && !isInverse && c.Component == "sh:MaxCountConstraintComponent" &&
		(forwardChainEndsRdfType || forwardChainOnly) {
		if int(shaclimport.AnyToFloat(c.Payload["MaxCount"])) == 1 {
			return checkSpec{}, nil, fmt.Errorf("multi-segment MaxCount=1 is structurally satisfied (refs are 0..1)")
		}
	}

	// Forward-chain-ending-rdf:type branch. Covers any number of forward
	// reference hops followed by a final rdf:type segment. The chain
	// resolves cleanly to a class-of-referenced-object check; the trailing
	// rdf:type is satisfied by the Go type system once the chain lands.
	// HasValue → exact match; In → allow-set; Required → reduce to "first
	// ref must be present" (the chain's existence is the constraint).
	if forwardChainEndsRdfType {
		refSegs := c.Path[:len(c.Path)-1]
		switch c.Component {
		case "sh:HasValueConstraintComponent":
			guard, cond, err := refClassEqualCondition(refSegs, structType, c.Payload["Value"])
			if err != nil {
				// Fallback: first hop may be a slice-of-MRID-struct.
				if len(refSegs) == 1 {
					seg, _ := stripCIMPrefix(refSegs[0])
					if sliceField, ok2 := findFieldByXMLTag(structType, seg); ok2 && sliceField.Type.Kind() == reflect.Slice {
						want, _ := c.Payload["Value"].(string)
						want = strings.TrimPrefix(strings.TrimSuffix(want, ">"), "<")
						if wantClass, ok3 := stripCIMPrefix(want); ok3 {
							var classes []string
							if _, ok4 := cimstructs.StructMap[wantClass]; ok4 {
								classes = []string{wantClass}
							} else {
								classes = concreteSubclassesEmbedding(wantClass)
							}
							if len(classes) > 0 {
								if sliceGuard, sliceErr := sliceRefClassInCondition(sliceField, classes, cs); sliceErr == nil {
									cs.Guard = sliceGuard
									cs.SelfContained = true
									imports = append(imports, "strings")
									return cs, imports, nil
								}
							}
						}
					}
				}
				return checkSpec{}, nil, err
			}
			cs.Guard, cs.Condition = guard, cond
			imports = append(imports, "strings")
		case "sh:InConstraintComponent":
			guard, cond, err := refClassInCondition(refSegs, structType, c.Payload["Values"])
			if err != nil {
				// Fallback: first hop may be a slice-of-MRID-struct (e.g. DependentOn, Supersedes).
				if len(refSegs) == 1 {
					seg, _ := stripCIMPrefix(refSegs[0])
					if sliceField, ok2 := findFieldByXMLTag(structType, seg); ok2 && sliceField.Type.Kind() == reflect.Slice {
						rawVals, _ := c.Payload["Values"].([]any)
						if classes, classErr := classListFromValues(rawVals); classErr == nil {
							if sliceGuard, sliceErr := sliceRefClassInCondition(sliceField, classes, cs); sliceErr == nil {
								cs.Guard = sliceGuard
								cs.SelfContained = true
								imports = append(imports, "strings")
								return cs, imports, nil
							}
						}
					}
				}
				return checkSpec{}, nil, err
			}
			cs.Guard, cs.Condition = guard, cond
			imports = append(imports, "strings")
		case "sh:RequiredConstraintComponent":
			// Required at end-rdf:type degenerates to "the chain's first
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
			return checkSpec{}, nil, fmt.Errorf("multi-segment %v ending in rdf:type not supported for %s", c.Path, c.Component)
		}
		return cs, imports, nil
	}

	// Forward chain WITHOUT trailing rdf:type. Two shapes map cleanly:
	//
	//   - sh:Or with all-Class shapes: the disjunction is the type test on
	//     the final referenced object.
	//   - sh:Datatype on a chain whose last segment is a literal field on
	//     a known compound class (e.g. Location.mainAddress / StreetAddress
	//     .status / Status.dateTime): walk the N-1 ref hops, type-assert
	//     to the literal's parent class, then apply the same datatype check
	//     used for single-segment string fields.
	if forwardChainOnly {
		if c.Component == "sh:DatatypeConstraintComponent" {
			cs2, imports2, err := multiSegDatatypeCheck(c, structType)
			if err == nil {
				cs.Guard = cs2.Guard
				cs.Condition = cs2.Condition
				imports = append(imports, imports2...)
				return cs, imports, nil
			}
			return checkSpec{}, nil, err
		}
		// A required-field check whose terminal path segment is in the rdf:
		// namespace (e.g. rdf:Statements.subject/predicate/object) is
		// structurally satisfied: the RDF spec mandates that every rdf:Statement
		// resource has subject, predicate, and object predicates.
		if c.Component == "sh:RequiredConstraintComponent" {
			lastSeg := c.Path[len(c.Path)-1]
			if strings.HasPrefix(lastSeg, "rdf:") {
				return checkSpec{}, nil, fmt.Errorf("multi-segment Required path ending in rdf: segment %q is structurally satisfied: RDF spec mandates rdf:Statement has subject/predicate/object", lastSeg)
			}
		}
		if c.Component != "sh:OrConstraintComponent" {
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
			fmt.Fprintf(&b, "\n\t\tcase *cimstructs.%s:\n\t\t\tisAllowedClass = true", cls)
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
	case "sh:MinExclusiveConstraintComponent":
		guard, cond, err := numericCompare(field, "<=", c.Payload["Value"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
	case "sh:MaxExclusiveConstraintComponent":
		guard, cond, err := numericCompare(field, ">=", c.Payload["Value"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
	case "sh:MinInclusiveConstraintComponent":
		guard, cond, err := numericCompare(field, "<", c.Payload["Value"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
	case "sh:MaxInclusiveConstraintComponent":
		guard, cond, err := numericCompare(field, ">", c.Payload["Value"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
	case "sh:RequiredConstraintComponent":
		cond, err := requiredCondition(field)
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Condition = cond
	case "sh:MinCountConstraintComponent":
		guard, cond, err := minCountCondition(field, c.Payload["MinCount"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
	case "sh:MaxCountConstraintComponent":
		guard, cond, err := maxCountCondition(field, c.Payload["MaxCount"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
	case "sh:HasValueConstraintComponent":
		guard, cond, err := hasValueCondition(field, c.Payload["Value"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
	case "sh:InConstraintComponent":
		guard, cond, err := inCondition(field, c.Payload["Values"])
		if err != nil {
			// Fallback: []string slice field (e.g. Model.Profile).
			if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.String {
				rawVals, _ := c.Payload["Values"].([]any)
				if sliceGuard, sliceErr := sliceStringInCondition(field, rawVals, cs); sliceErr == nil {
					cs.Guard = sliceGuard
					cs.SelfContained = true
					return cs, imports, nil
				}
			}
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
	case "sh:MinLengthConstraintComponent":
		guard, cond, err := minLengthCondition(field, c.Payload["MinLength"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
		imports = append(imports, "unicode/utf8")
	case "sh:MaxLengthConstraintComponent":
		guard, cond, err := maxLengthCondition(field, c.Payload["MaxLength"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
		imports = append(imports, "unicode/utf8")
	case "sh:PatternConstraintComponent":
		decl, guard, cond, err := patternCondition(field, c.Payload, name+"Regex")
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Decl, cs.Guard, cs.Condition = decl, guard, cond
		imports = append(imports, "regexp")
	case "sh:LessThanConstraintComponent":
		guard, cond, err := pairCompare(structType, field, c.Payload["Path"], ">=")
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
	case "sh:LessThanOrEqualsConstraintComponent":
		guard, cond, err := pairCompare(structType, field, c.Payload["Path"], ">")
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
	case "sh:ClassConstraintComponent":
		guard, cond, err := classCondition(field, c.Payload["Class"])
		if err != nil {
			// Fallback: slice-of-MRID-struct field.
			if field.Type.Kind() == reflect.Slice {
				want, _ := c.Payload["Class"].(string)
				want = strings.TrimPrefix(strings.TrimSuffix(want, ">"), "<")
				if wantClass, ok2 := stripCIMPrefix(want); ok2 {
					var classes []string
					if _, ok3 := cimstructs.StructMap[wantClass]; ok3 {
						classes = []string{wantClass}
					} else {
						classes = concreteSubclassesEmbedding(wantClass)
					}
					if len(classes) > 0 {
						if sliceGuard, sliceErr := sliceRefClassInCondition(field, classes, cs); sliceErr == nil {
							cs.Guard = sliceGuard
							cs.SelfContained = true
							imports = append(imports, "strings")
							return cs, imports, nil
						}
					}
				}
			}
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
		imports = append(imports, "strings")
	case "sh:DatatypeConstraintComponent":
		guard, cond, dtImports, err := datatypeCondition(field, c.Payload["Datatype"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		cs.Guard, cs.Condition = guard, cond
		imports = append(imports, dtImports...)
	case "sh:NotConstraintComponent":
		// Single-segment Not is a forward-Class assertion with the
		// condition inverted. The wrapped shape must reduce to a single
		// sh:Class — anything richer (Not of a length range, of an Or,
		// etc.) is too uncommon here to be worth recursive emission.
		notClass, err := notClassFromShapeRef(c.Payload["ShapeRef"])
		if err != nil {
			return checkSpec{}, nil, err
		}
		guard, _, err := classCondition(field, "cim:"+notClass)
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

// componentShort returns the camel-case suffix used in generated function
// names for each supported SHACL constraint component. Returning ok=false
// signals the caller to skip the constraint with a "not supported" reason.
func componentShort(component string) (string, bool) {
	switch component {
	case "sh:MinExclusiveConstraintComponent":
		return "MinExclusive", true
	case "sh:MaxExclusiveConstraintComponent":
		return "MaxExclusive", true
	case "sh:MinInclusiveConstraintComponent":
		return "MinInclusive", true
	case "sh:MaxInclusiveConstraintComponent":
		return "MaxInclusive", true
	case "sh:RequiredConstraintComponent":
		return "Required", true
	case "sh:MinCountConstraintComponent":
		return "MinCount", true
	case "sh:MaxCountConstraintComponent":
		return "MaxCount", true
	case "sh:HasValueConstraintComponent":
		return "HasValue", true
	case "sh:InConstraintComponent":
		return "In", true
	case "sh:MinLengthConstraintComponent":
		return "MinLength", true
	case "sh:MaxLengthConstraintComponent":
		return "MaxLength", true
	case "sh:PatternConstraintComponent":
		return "Pattern", true
	case "sh:LessThanConstraintComponent":
		return "LessThan", true
	case "sh:LessThanOrEqualsConstraintComponent":
		return "LessThanOrEquals", true
	case "sh:ClassConstraintComponent":
		return "Class", true
	case "sh:DatatypeConstraintComponent":
		return "Datatype", true
	case "sh:OrConstraintComponent":
		return "Or", true
	case "sh:NotConstraintComponent":
		return "Not", true
	}
	return "", false
}
