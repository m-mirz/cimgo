package main

import (
	"cimgo/cimstructs"
	"cimgo/shaclimport"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// buildDatasetCardinalityCheck handles `^rdf:type` cardinality constraints
// (MinCount/MaxCount on the count of focus-class instances in the dataset).
// Emitted as a DatasetCheck — the per-element loop is skipped entirely and
// a single violation is appended (or not) based on the global count.
func buildDatasetCardinalityCheck(stemCamel, structName, shapeID string, c shaclimport.ConstraintInfo, used map[string]int) (checkSpec, []string, error) {
	compShort, ok := componentShort(c.Component)
	if !ok {
		return checkSpec{}, nil, fmt.Errorf("component %s not supported on ^rdf:type", c.Component)
	}
	var op, threshold string
	switch c.Component {
	case "sh:MinCountConstraintComponent":
		min := int(shaclimport.AnyToFloat(c.Payload["MinCount"]))
		op, threshold = "<", fmt.Sprintf("%d", min)
	case "sh:MaxCountConstraintComponent":
		max := int(shaclimport.AnyToFloat(c.Payload["MaxCount"]))
		op, threshold = ">", fmt.Sprintf("%d", max)
	default:
		return checkSpec{}, nil, fmt.Errorf("only Min/MaxCount supported on ^rdf:type, got %s", c.Component)
	}
	// Resolve focus class to either a single concrete or a list of concrete
	// subclasses (when the focus is abstract). Counting is a union over all
	// matching concrete types.
	var classes []string
	if _, ok := cimstructs.StructMap[structName]; ok {
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
		severity = "sh:Violation"
	}

	ruleID := shapeID

	var b strings.Builder
	b.WriteString("\tdatasetCount := 0\n")
	b.WriteString("\tfor _, ref := range dataset.ByID {\n")
	b.WriteString("\t\tswitch ref.(type) {")
	for _, cls := range classes {
		fmt.Fprintf(&b, "\n\t\tcase *cimstructs.%s:\n\t\t\tdatasetCount++", cls)
	}
	b.WriteString("\n\t\t}\n")
	b.WriteString("\t}")
	cs := checkSpec{
		Name:         name,
		ShapeID:      shapeID,
		RuleID:       ruleID,
		RuleName:     c.Name,
		Description:  c.Description,
		Class:        structName,
		Tag:          "^rdf:type",
		Component:    c.Component,
		Property:     "^rdf:type",
		Message:      strings.Trim(c.Message, "\""),
		Severity:     severity,
		Prelude:      b.String(),
		Condition:    fmt.Sprintf("datasetCount %s %s", op, threshold),
		DatasetCheck: true,
	}
	return cs, nil, nil
}

// multiSegDatatypeCheck handles sh:Datatype on a forward chain where the
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
	factory, ok := cimstructs.StructMap[parentClass]
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
	dt, ok := c.Payload["Datatype"].(string)
	if !ok {
		return checkSpec{}, nil, fmt.Errorf("Datatype payload missing or not a string")
	}
	if field.Type.Kind() != reflect.String {
		return checkSpec{}, nil, fmt.Errorf("Datatype %q on %s field is structurally satisfied", dt, field.Type.Kind())
	}

	var b strings.Builder
	b.WriteString(chainGuard)
	fmt.Fprintf(&b, "\n\t\tparent, parentOk := %s.(*cimstructs.%s)\n\t\tif !parentOk {\n\t\t\tcontinue\n\t\t}\n", targetVar, parentClass)
	fmt.Fprintf(&b, "\t\tif parent.%s == \"\" {\n\t\t\tcontinue\n\t\t}\n", field.Name)
	var cond string
	var extraImports []string
	switch dt {
	case "xsd:dateTime":
		fmt.Fprintf(&b, "\t\t_, parseErr := time.Parse(time.RFC3339, parent.%s)", field.Name)
		cond = "parseErr != nil"
		extraImports = []string{"time", "strings"}
	case "xsd:date":
		fmt.Fprintf(&b, "\t\t_, parseErr := time.Parse(\"2006-01-02\", parent.%s)", field.Name)
		cond = "parseErr != nil"
		extraImports = []string{"time", "strings"}
	case "xsd:gMonthDay":
		fmt.Fprintf(&b, "\t\t_, parseErr1 := time.Parse(\"--01-02\", parent.%s)\n", field.Name)
		fmt.Fprintf(&b, "\t\t_, parseErr2 := time.Parse(\"01-02\", parent.%s)", field.Name)
		cond = "parseErr1 != nil && parseErr2 != nil"
		extraImports = []string{"time", "strings"}
	case "xsd:anyURI":
		fmt.Fprintf(&b, "\t\t_, parseErr := url.ParseRequestURI(parent.%s)", field.Name)
		cond = "parseErr != nil"
		extraImports = []string{"net/url", "strings"}
	default:
		return checkSpec{}, nil, fmt.Errorf("Datatype %q not supported on multi-segment chain", dt)
	}
	return checkSpec{Guard: b.String(), Condition: cond}, extraImports, nil
}

// walkForwardRefChain emits Guard text that chases a sequence of forward
// reference hops. Each segment must be `cim:<Class>.<field>` (already with
// cim: prefix). The function dereferences each reference, looks the target
// up in the dataset, and (for all but the final hop) type-asserts the
// target to the class implied by the *next* segment's class portion. The
// final hop returns the raw Element value as `targetVar`; the caller plugs
// in the final test (e.g. type-switch for In, single type-assert for
// HasValue).
//
// At every step, if a reference is missing or the dataset lookup fails or a
// type assertion fails, we `continue`. The chain not resolving means there
// is no "value" for the property path at this focus node — so by SHACL
// semantics no value-shape constraint can fail; sh:Required is the
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
		if field.Type.Kind() != reflect.Pointer {
			return "", "", fmt.Errorf("chain step %d: field %q is %s, expected pointer", i, seg, field.Type.Kind())
		}
		fmt.Fprintf(&b, "\t\tif %s.%s == nil {\n\t\t\tcontinue\n\t\t}\n", currentVar, field.Name)
		fmt.Fprintf(&b, "\t\trefID%d := strings.TrimPrefix(%s.%s.MRID, \"#\")\n", i, currentVar, field.Name)
		targetVar := fmt.Sprintf("target%d", i)
		fmt.Fprintf(&b, "\t\t%s, found%d := dataset.ByID[refID%d]\n", targetVar, i, i)
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
		factory, ok := cimstructs.StructMap[nextClass]
		if !ok {
			return "", "", fmt.Errorf("chain step %d: target class %q has no Go struct", i, nextClass)
		}
		tVar := fmt.Sprintf("t%d", i)
		fmt.Fprintf(&b, "\t\t%s, ok%d := %s.(*cimstructs.%s)\n", tVar, i, targetVar, nextClass)
		fmt.Fprintf(&b, "\t\tif !ok%d {\n\t\t\tcontinue\n\t\t}\n", i)
		currentVar = tVar
		currentType = reflect.TypeOf(factory()).Elem()
	}
	return strings.TrimRight(b.String(), "\n"), currentVar, nil
}

// refClassEqualCondition implements sh:HasValue along a forward chain ending
// in rdf:type: chase the chain, then require the final target's Go type to
// match the named class exactly.
func refClassEqualCondition(refSegs []string, startType reflect.Type, payload any) (string, string, error) {
	want, ok := payload.(string)
	if !ok {
		return "", "", fmt.Errorf("HasValue payload not a string")
	}
	want = strings.TrimPrefix(strings.TrimSuffix(want, ">"), "<")
	// rdf:-namespace types (e.g. rdf:Statements) cannot be resolved to a Go
	// struct, but CGMES format guarantees these collections only ever contain
	// elements of the named RDF type, so the constraint is structurally satisfied.
	if strings.HasPrefix(want, "rdf:") {
		return "", "", fmt.Errorf("HasValue rdf:type %q is structurally satisfied: CGMES format guarantees rdf: type", want)
	}
	wantClass, ok := stripCIMPrefix(want)
	if !ok {
		return "", "", fmt.Errorf("HasValue rdf:type %q not in cim namespace", want)
	}
	if _, ok := cimstructs.StructMap[wantClass]; !ok {
		return "", "", fmt.Errorf("HasValue rdf:type %q has no Go struct", wantClass)
	}
	guard, targetVar, err := walkForwardRefChain(refSegs, startType, "v")
	if err != nil {
		return "", "", err
	}
	guard += fmt.Sprintf("\n\t\t_, isWantedClass := %s.(*cimstructs.%s)", targetVar, wantClass)
	return guard, "!isWantedClass", nil
}

// refClassInCondition implements sh:In along a forward chain ending in
// rdf:type: chase the chain, then require the final target's Go type to be
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
		fmt.Fprintf(&b, "\n\t\tcase *cimstructs.%s:\n\t\t\tisAllowedClass = true", cls)
	}
	b.WriteString("\n\t\t}")
	return b.String(), "!isAllowedClass", nil
}

// sliceRefClassInCondition handles sh:In along a path whose sole reference hop
// is a slice-of-anonymous-MRID-struct field (e.g. Model.DependentOn,
// Model.Supersedes). walkForwardRefChain cannot handle slice fields, so this
// function emits a complete self-contained inner for-loop that looks up each
// MRID in dataset.ByID and appends violations directly. The caller must set
// SelfContained on the checkSpec so the template skips the outer
// `if Condition { violations = append }` block.
func sliceRefClassInCondition(field reflect.StructField, classes []string, cs checkSpec) (string, error) {
	if field.Type.Kind() != reflect.Slice {
		return "", fmt.Errorf("expected slice field, got %s", field.Type.Kind())
	}
	elemType := field.Type.Elem()
	if elemType.Kind() != reflect.Struct {
		return "", fmt.Errorf("slice element is not a struct")
	}
	var mridFieldName string
	for i := 0; i < elemType.NumField(); i++ {
		if f := elemType.Field(i); f.Type.Kind() == reflect.String {
			mridFieldName = f.Name
			break
		}
	}
	if mridFieldName == "" {
		return "", fmt.Errorf("slice element struct has no string (MRID) field")
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\t\tfor _, ref := range v.%s {\n", field.Name)
	fmt.Fprintf(&b, "\t\t\tmrid := strings.TrimPrefix(ref.%s, \"#\")\n", mridFieldName)
	b.WriteString("\t\t\ttarget, found := dataset.ByID[mrid]\n")
	b.WriteString("\t\t\tif !found {\n\t\t\t\tcontinue\n\t\t\t}\n")
	b.WriteString("\t\t\tisAllowedClass := false\n")
	b.WriteString("\t\t\tswitch target.(type) {")
	for _, cls := range classes {
		fmt.Fprintf(&b, "\n\t\t\tcase *cimstructs.%s:\n\t\t\t\tisAllowedClass = true", cls)
	}
	b.WriteString("\n\t\t\t}\n")
	b.WriteString("\t\t\tif !isAllowedClass {\n")
	b.WriteString("\t\t\t\tviolations = append(violations, shaclmodel.Violation{\n")
	fmt.Fprintf(&b, "\t\t\t\t\tObjectID:    id,\n")
	fmt.Fprintf(&b, "\t\t\t\t\tRuleID:      %q,\n", cs.RuleID)
	fmt.Fprintf(&b, "\t\t\t\t\tClass:       %q,\n", cs.Class)
	fmt.Fprintf(&b, "\t\t\t\t\tProperty:    %q,\n", cs.Property)
	fmt.Fprintf(&b, "\t\t\t\t\tMessage:     %q,\n", cs.Message)
	fmt.Fprintf(&b, "\t\t\t\t\tSeverity:    %q,\n", cs.Severity)
	fmt.Fprintf(&b, "\t\t\t\t\tName:        %q,\n", cs.RuleName)
	fmt.Fprintf(&b, "\t\t\t\t\tDescription: %q,\n", cs.Description)
	b.WriteString("\t\t\t\t})\n")
	b.WriteString("\t\t\t}\n")
	b.WriteString("\t\t}")
	return b.String(), nil
}

// classListFromValues parses a SHACL `Values` payload into a list of Go
// struct names, normalising the `<cim:Foo>`/`cim100.Foo` IRI forms and
// verifying each name resolves to a generated struct. Abstract base classes
// (not in StructMap directly) are expanded to their concrete subclasses so
// the resulting allow-set is exhaustive at the Go-type level. Shared by
// sh:In on rdf:type and sh:Or with all-Class shapes.
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
			return nil, fmt.Errorf("rdf:type list contains non-string %v", v)
		}
		s = strings.TrimPrefix(strings.TrimSuffix(s, ">"), "<")
		cls, ok := stripCIMPrefix(s)
		if !ok {
			return nil, fmt.Errorf("rdf:type %q not in cim namespace", s)
		}
		if _, ok := cimstructs.StructMap[cls]; ok {
			add(cls)
			continue
		}
		subs := concreteSubclassesEmbedding(cls)
		if len(subs) == 0 {
			return nil, fmt.Errorf("rdf:type %q has no Go struct", cls)
		}
		for _, s := range subs {
			add(s)
		}
	}
	sort.Strings(classes)
	return classes, nil
}

// orClassListFromShapes accepts the Or payload's `Shapes` list (a list of
// shape lists, each containing a single sh:ClassConstraintComponent) and
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
		if c.Component != "sh:ClassConstraintComponent" {
			return nil, fmt.Errorf("Or shape %d component %q is not Class", i, c.Component)
		}
		want, _ := c.Payload["Class"].(string)
		want = strings.TrimPrefix(strings.TrimSuffix(want, ">"), "<")
		cls, ok := stripCIMPrefix(want)
		if !ok {
			return nil, fmt.Errorf("Or shape %d Class %q not in cim namespace", i, want)
		}
		if _, ok := cimstructs.StructMap[cls]; !ok {
			return nil, fmt.Errorf("Or shape %d Class %q has no Go struct", i, cls)
		}
		classes = append(classes, cls)
	}
	return classes, nil
}

// notClassFromShapeRef accepts the Not payload's `ShapeRef` list (expected
// to contain a single sh:ClassConstraintComponent) and returns the negated
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
	if c.Component != "sh:ClassConstraintComponent" {
		return "", fmt.Errorf("Not component %q is not Class", c.Component)
	}
	want, _ := c.Payload["Class"].(string)
	want = strings.TrimPrefix(strings.TrimSuffix(want, ">"), "<")
	cls, ok := stripCIMPrefix(want)
	if !ok {
		return "", fmt.Errorf("Not Class %q not in cim namespace", want)
	}
	if _, ok := cimstructs.StructMap[cls]; !ok {
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
