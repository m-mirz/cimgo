package main

import (
	"cimgo/cimstructs"
	"cimgo/shaclimport"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

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
	case reflect.Pointer:
		return fmt.Sprintf("v.%s == nil", field.Name), nil
	case reflect.Slice:
		return fmt.Sprintf("len(v.%s) == 0", field.Name), nil
	case reflect.String:
		return fmt.Sprintf("v.%s == \"\"", field.Name), nil
	case reflect.Int, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("v.%s == 0", field.Name), nil
	case reflect.Float32, reflect.Float64:
		return "", fmt.Errorf("float Required is unreliable: zero is indistinguishable from absent")
	case reflect.Bool:
		return "", fmt.Errorf("bool Required is structurally satisfied: false is indistinguishable from absent")
	default:
		return "", fmt.Errorf("unsupported required kind %s", field.Type.Kind())
	}
}

// minCountCondition handles sh:MinCount. For pointer fields, MinCount=1 is
// equivalent to Required. For slices we compare against the literal threshold.
func minCountCondition(field reflect.StructField, payload any) (string, string, error) {
	min := int(shaclimport.AnyToFloat(payload))
	switch field.Type.Kind() {
	case reflect.Slice:
		return "", fmt.Sprintf("len(v.%s) < %d", field.Name, min), nil
	case reflect.Pointer:
		if min <= 1 {
			return "", fmt.Sprintf("v.%s == nil", field.Name), nil
		}
		return "", "", fmt.Errorf("MinCount=%d on pointer field is unsatisfiable", min)
	default:
		return "", "", fmt.Errorf("MinCount on %s field not supported", field.Type.Kind())
	}
}

// maxCountCondition handles sh:MaxCount. Pointers and scalars are structurally
// bounded at 1, so MaxCount >= 1 is vacuous. MaxCount=0 is a "must not be set" rule.
func maxCountCondition(field reflect.StructField, payload any) (string, string, error) {
	max := int(shaclimport.AnyToFloat(payload))
	switch field.Type.Kind() {
	case reflect.Slice:
		return "", fmt.Sprintf("len(v.%s) > %d", field.Name, max), nil
	case reflect.Pointer:
		if max == 0 {
			return "", fmt.Sprintf("v.%s != nil", field.Name), nil
		}
		return "", "", fmt.Errorf("MaxCount=%d on pointer field is structurally satisfied", max)
	case reflect.String, reflect.Int, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64, reflect.Bool:
		if max == 0 {
			return "", fmt.Sprintf("v.%s != %s", field.Name, zeroLiteralFor(field.Type.Kind())), nil
		}
		return "", "", fmt.Errorf("MaxCount=%d on scalar field is structurally satisfied", max)
	default:
		return "", "", fmt.Errorf("MaxCount on %s field not supported", field.Type.Kind())
	}
}

// hasValueCondition handles sh:HasValue for string-typed fields and for
// enum-as-IRI reference fields.
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
		var cond string
		rest, _ := stripCIMPrefix(want)
		fullURI := cimNamespaceFromPrefix(want) + rest
		if strings.HasPrefix(fullURI, "http://iec.ch/TC57/CIM100-European#") {
			cond = fmt.Sprintf("v.%s.URI != %q", field.Name, fullURI)
		} else {
			cond = fmt.Sprintf("v.%s.URI != cimstructs.%s", field.Name, constIdent)
		}
		return guard, cond, nil
	}
	switch field.Type.Kind() {
	case reflect.String:
		guard := fmt.Sprintf("\t\tif v.%s == \"\" {\n\t\t\tcontinue\n\t\t}", field.Name)
		cond := fmt.Sprintf("v.%s != %q", field.Name, want)
		return guard, cond, nil
	case reflect.Bool:
		bval := want == "true"
		cond := fmt.Sprintf("v.%s != %v", field.Name, bval)
		return "", cond, nil
	case reflect.Int, reflect.Int32, reflect.Int64:
		n := int64(shaclimport.AnyToFloat(want))
		guard := fmt.Sprintf("\t\tif v.%s == 0 {\n\t\t\tcontinue\n\t\t}", field.Name)
		cond := fmt.Sprintf("int64(v.%s) != %d", field.Name, n)
		return guard, cond, nil
	case reflect.Float32, reflect.Float64:
		f := shaclimport.AnyToFloat(want)
		guard := fmt.Sprintf("\t\tif v.%s == 0 {\n\t\t\tcontinue\n\t\t}", field.Name)
		cond := fmt.Sprintf("float64(v.%s) != %v", field.Name, f)
		return guard, cond, nil
	default:
		return "", "", fmt.Errorf("HasValue on %s field not supported", field.Type.Kind())
	}
}

// inCondition handles sh:In for string-typed fields and for enum-as-IRI fields.
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
			type enumEntry struct{ constIdent, fullURI string }
			entries := make([]enumEntry, 0, len(values))
			for _, want := range values {
				constIdent, _, err := enumURIFieldConst(field, want)
				if err != nil {
					return "", "", err
				}
				rest, _ := stripCIMPrefix(want)
				fullURI := cimNamespaceFromPrefix(want) + rest
				entries = append(entries, enumEntry{constIdent, fullURI})
			}
			var b strings.Builder
			fmt.Fprintf(&b, "\t\tif v.%s == nil {\n\t\t\tcontinue\n\t\t}\n", field.Name)
			b.WriteString("\t\tallowed := map[string]bool{")
			for i, e := range entries {
				if i > 0 {
					b.WriteString(", ")
				}
				if strings.HasPrefix(e.fullURI, "http://iec.ch/TC57/CIM100-European#") {
					fmt.Fprintf(&b, "%q: true", e.fullURI)
				} else {
					fmt.Fprintf(&b, "cimstructs.%s: true", e.constIdent)
				}
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
			fmt.Fprintf(&b, "%d: true", int64(shaclimport.AnyToFloat(val)))
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
			fmt.Fprintf(&b, "%v: true", shaclimport.AnyToFloat(val))
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

// sliceStringInCondition handles sh:In for []string slice fields. It emits a
// self-contained inner for-loop that checks each element against the allowed
// set and appends violations directly. The caller must set SelfContained on
// the checkSpec.
func sliceStringInCondition(field reflect.StructField, rawValues []any, cs checkSpec) (string, error) {
	if field.Type.Kind() != reflect.Slice || field.Type.Elem().Kind() != reflect.String {
		return "", fmt.Errorf("expected []string field, got %s", field.Type)
	}
	values := make([]string, 0, len(rawValues))
	for _, v := range rawValues {
		s, ok := v.(string)
		if !ok {
			return "", fmt.Errorf("In list contains non-string %v", v)
		}
		values = append(values, strings.TrimPrefix(strings.TrimSuffix(s, ">"), "<"))
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\t\tfor _, val := range v.%s {\n", field.Name)
	b.WriteString("\t\t\tif val == \"\" {\n\t\t\t\tcontinue\n\t\t\t}\n")
	b.WriteString("\t\t\tallowed := map[string]bool{")
	for i, val := range values {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%q: true", val)
	}
	b.WriteString("}\n")
	b.WriteString("\t\t\tif !allowed[val] {\n")
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

// enumURIFieldConst inspects field for the enum-as-IRI shape (pointer to a
// struct with a single string field named URI). When the shape matches, the
// returned constIdent is the cimstructs constant identifier corresponding
// to payload.
func enumURIFieldConst(field reflect.StructField, payload string) (string, bool, error) {
	if field.Type.Kind() != reflect.Pointer {
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
	class, member, ok2 := strings.Cut(rest, ".")
	if !ok2 {
		return "", true, fmt.Errorf("enum payload %q missing '.member' segment", payload)
	}
	return class + member, true, nil
}

// minLengthCondition handles sh:MinLength on string fields.
func minLengthCondition(field reflect.StructField, payload any) (string, string, error) {
	if field.Type.Kind() != reflect.String {
		return "", "", fmt.Errorf("MinLength on non-string field (%s) not supported", field.Type.Kind())
	}
	min := int(shaclimport.AnyToFloat(payload))
	guard := fmt.Sprintf("\t\tif v.%s == \"\" {\n\t\t\tcontinue\n\t\t}", field.Name)
	cond := fmt.Sprintf("utf8.RuneCountInString(v.%s) < %d", field.Name, min)
	return guard, cond, nil
}

// maxLengthCondition handles sh:MaxLength on string fields.
func maxLengthCondition(field reflect.StructField, payload any) (string, string, error) {
	if field.Type.Kind() != reflect.String {
		return "", "", fmt.Errorf("MaxLength on non-string field (%s) not supported", field.Type.Kind())
	}
	max := int(shaclimport.AnyToFloat(payload))
	guard := fmt.Sprintf("\t\tif v.%s == \"\" {\n\t\t\tcontinue\n\t\t}", field.Name)
	cond := fmt.Sprintf("utf8.RuneCountInString(v.%s) > %d", field.Name, max)
	return guard, cond, nil
}

// patternCondition handles sh:Pattern on string fields. The regex is hoisted
// to a package-level var so each Check call reuses one compiled pattern.
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
// sh:LessThan and sh:LessThanOrEquals.
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

// numericCastFor returns the wrapper cast for a numeric kind to normalise
// cross-kind comparisons to float64. Returns ok=false for non-numeric kinds.
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

// classCondition handles sh:ClassConstraintComponent for forward reference fields.
func classCondition(field reflect.StructField, payload any) (string, string, error) {
	want, ok := payload.(string)
	if !ok {
		return "", "", fmt.Errorf("Class payload missing or not a string")
	}
	wantClass, ok := stripCIMPrefix(want)
	if !ok {
		return "", "", fmt.Errorf("Class %q not in cim namespace", want)
	}
	if field.Type.Kind() != reflect.Pointer {
		return "", "", fmt.Errorf("Class on non-pointer field (%s) not supported", field.Type.Kind())
	}
	var classes []string
	if _, ok := cimstructs.StructMap[wantClass]; ok {
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
		target, found := dataset.ByID[refID]
		if !found {
			continue
		}
		_, isWantedClass := target.(*cimstructs.%s)`, field.Name, field.Name, classes[0])
		return guard, "!isWantedClass", nil
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\t\tif v.%s == nil {\n\t\t\tcontinue\n\t\t}\n", field.Name)
	fmt.Fprintf(&b, "\t\trefID := strings.TrimPrefix(v.%s.MRID, \"#\")\n", field.Name)
	b.WriteString("\t\ttarget, found := dataset.ByID[refID]\n")
	b.WriteString("\t\tif !found {\n\t\t\tcontinue\n\t\t}\n")
	b.WriteString("\t\tisWantedClass := false\n")
	b.WriteString("\t\tswitch target.(type) {")
	for _, cls := range classes {
		fmt.Fprintf(&b, "\n\t\tcase *cimstructs.%s:\n\t\t\tisWantedClass = true", cls)
	}
	b.WriteString("\n\t\t}")
	return b.String(), "!isWantedClass", nil
}

// datatypeCondition handles sh:DatatypeConstraintComponent on string fields.
func datatypeCondition(field reflect.StructField, payload any) (string, string, []string, error) {
	dt, ok := payload.(string)
	if !ok {
		return "", "", nil, fmt.Errorf("Datatype payload missing or not a string")
	}
	if field.Type.Kind() != reflect.String {
		return "", "", nil, fmt.Errorf("Datatype %q on %s field is structurally satisfied", dt, field.Type.Kind())
	}
	switch dt {
	case "xsd:dateTime":
		guard := fmt.Sprintf("\t\tif v.%s == \"\" {\n\t\t\tcontinue\n\t\t}\n\t\t_, parseErr := time.Parse(time.RFC3339, v.%s)", field.Name, field.Name)
		return guard, "parseErr != nil", []string{"time"}, nil
	case "xsd:date":
		guard := fmt.Sprintf("\t\tif v.%s == \"\" {\n\t\t\tcontinue\n\t\t}\n\t\t_, parseErr := time.Parse(\"2006-01-02\", v.%s)", field.Name, field.Name)
		return guard, "parseErr != nil", []string{"time"}, nil
	case "xsd:gMonthDay":
		guard := fmt.Sprintf(`		if v.%s == "" {
			continue
		}
		_, parseErr1 := time.Parse("--01-02", v.%s)
		_, parseErr2 := time.Parse("01-02", v.%s)`, field.Name, field.Name, field.Name)
		return guard, "parseErr1 != nil && parseErr2 != nil", []string{"time"}, nil
	case "xsd:anyURI":
		guard := fmt.Sprintf("\t\tif v.%s == \"\" {\n\t\t\tcontinue\n\t\t}\n\t\t_, parseErr := url.ParseRequestURI(v.%s)", field.Name, field.Name)
		return guard, "parseErr != nil", []string{"net/url"}, nil
	}
	return "", "", nil, fmt.Errorf("Datatype %q not supported on string field", dt)
}
