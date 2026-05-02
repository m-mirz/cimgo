package validation

import (
	"cimgo/cimgostructs"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

// Models for the struct JSON documentation
type ConstraintInfo struct {
	Path      []string       `json:"path"`
	Severity  string         `json:"severity,omitempty"`
	Message   string         `json:"message,omitempty"`
	Component string         `json:"component"`
	Payload   map[string]any `json:"payload"`
}

// isInversePath reports whether p is a single-step inverse path like ^cim.X.Y.
// countInverseRelations only handles single-step inverses, so a multi-step
// sequence beginning with an inverse is not treated as one.
func isInversePath(p []string) bool {
	return len(p) == 1 && strings.HasPrefix(p[0], "^")
}

type AttributeInfo struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Constraints []ConstraintInfo `json:"constraints"`
}

type ClassInfo struct {
	Name        string           `json:"name"`
	Constraints []ConstraintInfo `json:"constraints"`
	Attributes  []AttributeInfo  `json:"attributes"`
}

type FileResults struct {
	FileName string      `json:"file_name"`
	Classes  []ClassInfo `json:"classes"`
}

func loadAllRules(t *testing.T, structPaths ...string) map[string]ClassInfo {
	rules := make(map[string]ClassInfo)

	var files []string
	for _, structPath := range structPaths {
		info, err := os.Stat(structPath)
		if err != nil {
			t.Fatal(err)
		}
		if info.IsDir() {
			glob, err := filepath.Glob(filepath.Join(structPath, "*.json"))
			if err != nil {
				t.Fatal(err)
			}
			files = append(files, glob...)
		} else {
			files = append(files, structPath)
		}
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var res FileResults
		if err := json.Unmarshal(data, &res); err != nil {
			continue
		}

		for _, cls := range res.Classes {
			if existing, ok := rules[cls.Name]; ok {
				existing.Constraints = append(existing.Constraints, cls.Constraints...)
				existing.Attributes = append(existing.Attributes, cls.Attributes...)
				rules[cls.Name] = existing
			} else {
				rules[cls.Name] = cls
			}
		}
	}
	return rules
}

// goTypeToCIMPrefix maps Go struct type names to their CIM namespace prefix
// for types that are not in the default "cim" namespace (http://iec.ch/TC57/CIM100#).
var goTypeToCIMPrefix = map[string]string{
	"FullModel":       "mdc",
	"DifferenceModel": "diff",
	"Statements":      "rdf",
}

func getCIMTypeName(obj interface{}) string {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	name := t.Name()
	if prefix, ok := goTypeToCIMPrefix[name]; ok {
		return prefix + "." + name
	}
	return "cim." + name
}

func validateObject(t *testing.T, obj interface{}, rules map[string]ClassInfo, dataset *cimgostructs.CIMElementList) []string {
	var violations []string
	cimType := getCIMTypeName(obj)
	classRule, ok := rules[cimType]
	if !ok {
		return nil
	}

	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 1. Validate Class-level Constraints
	for _, constraint := range classRule.Constraints {
		v := checkConstraint(val, constraint, dataset)
		if v != "" {
			violations = append(violations, fmt.Sprintf("[%s] Class error: %s", cimType, v))
		}
	}

	// 2. Validate Attributes
	for _, attrRule := range classRule.Attributes {
		path := attrRule.Name

		// Inverse paths (^...) are resolved by checkConstraint via
		// countInverseRelations, which needs the focus node, not a field.
		if strings.HasPrefix(strings.TrimSpace(path), "^") {
			for _, constraint := range attrRule.Constraints {
				if v := checkConstraint(val, constraint, dataset); v != "" {
					violations = append(violations, fmt.Sprintf("[%s] %s: %s", cimType, path, v))
				}
			}
			continue
		}

		segments := strings.Split(path, " / ")
		fields := resolvePathSegments(val, segments)

		for _, constraint := range attrRule.Constraints {
			if constraint.Component == "sh.HasValueConstraintComponent" {
				// HasValue requires at least one terminal value to match.
				if v := checkHasValue(fields, constraint); v != "" {
					violations = append(violations, fmt.Sprintf("[%s] %s: %s", cimType, path, v))
				}
			} else {
				for _, field := range fields {
					if v := checkConstraint(field, constraint, dataset); v != "" {
						violations = append(violations, fmt.Sprintf("[%s] %s: %s", cimType, path, v))
					}
				}
			}
		}
	}

	return violations
}

func checkConstraint(field reflect.Value, c ConstraintInfo, dataset *cimgostructs.CIMElementList) string {
	switch c.Component {
	case "sh.OrConstraintComponent":
		// Payload has "Shapes" which is []any (nested shapes)
		if shapes, ok := c.Payload["Shapes"].([]interface{}); ok {
			for _, shape := range shapes {
				if checkOrOption(field, shape, dataset) {
					return ""
				}
			}
			return "OrConstraint violation: none of the required shapes matched"
		}

	case "sh.MinCountConstraintComponent":
		if minVal, ok := c.Payload["MinCount"]; ok {
			min := int(anyToFloat(minVal))
			var count int
			if isInversePath(c.Path) {
				count = countInverseRelations(field, c.Path[0], dataset)
			} else {
				count = getCount(field)
			}
			if count < min {
				return fmt.Sprintf("MinCount violation: got %d, want %d", count, min)
			}
		}

	case "sh.MaxCountConstraintComponent":
		if maxVal, ok := c.Payload["MaxCount"]; ok {
			max := int(anyToFloat(maxVal))
			var count int
			if isInversePath(c.Path) {
				count = countInverseRelations(field, c.Path[0], dataset)
			} else {
				count = getCount(field)
			}
			if count > max {
				return fmt.Sprintf("MaxCount violation: got %d, want %d", count, max)
			}
		}

	case "sh.MinInclusiveConstraintComponent":
		if minVal, ok := c.Payload["Value"]; ok {
			min := anyToFloat(minVal)
			val, ok := getFloat(field)
			if ok && getCount(field) > 0 && val < min {
				return fmt.Sprintf("MinInclusive violation: got %v, want >= %v", val, min)
			}
		}

	case "sh.MaxInclusiveConstraintComponent":
		if maxVal, ok := c.Payload["Value"]; ok {
			max := anyToFloat(maxVal)
			val, ok := getFloat(field)
			if ok && getCount(field) > 0 && val > max {
				return fmt.Sprintf("MaxInclusive violation: got %v, want <= %v", val, max)
			}
		}

	case "sh.MinExclusiveConstraintComponent":
		if minVal, ok := c.Payload["Value"]; ok {
			min := anyToFloat(minVal)
			val, ok := getFloat(field)
			if ok && getCount(field) > 0 && val <= min {
				return fmt.Sprintf("MinExclusive violation: got %v, want > %v", val, min)
			}
		}

	case "sh.MaxExclusiveConstraintComponent":
		if maxVal, ok := c.Payload["Value"]; ok {
			max := anyToFloat(maxVal)
			val, ok := getFloat(field)
			if ok && getCount(field) > 0 && val >= max {
				return fmt.Sprintf("MaxExclusive violation: got %v, want < %v", val, max)
			}
		}

	case "sh.RequiredConstraintComponent":
		var count int
		if isInversePath(c.Path) {
			count = countInverseRelations(field, c.Path[0], dataset)
		} else {
			count = getCount(field)
		}
		if count != 1 {
			return fmt.Sprintf("Required violation: got %d values, want exactly 1", count)
		}

	case "sh.OptionalConstraintComponent":
		var count int
		if isInversePath(c.Path) {
			count = countInverseRelations(field, c.Path[0], dataset)
		} else {
			count = getCount(field)
		}
		if count > 1 {
			return fmt.Sprintf("Optional violation: got %d values, want at most 1", count)
		}

	case "sh.MinLengthConstraintComponent":
		if minVal, ok := c.Payload["MinLength"]; ok {
			min := int(anyToFloat(minVal))
			if s, ok := getString(field); ok && utf8.RuneCountInString(s) < min {
				return fmt.Sprintf("MinLength violation: got %d chars, want at least %d", utf8.RuneCountInString(s), min)
			}
		}

	case "sh.MaxLengthConstraintComponent":
		if maxVal, ok := c.Payload["MaxLength"]; ok {
			max := int(anyToFloat(maxVal))
			if s, ok := getString(field); ok && utf8.RuneCountInString(s) > max {
				return fmt.Sprintf("MaxLength violation: got %d chars, want at most %d", utf8.RuneCountInString(s), max)
			}
		}

	case "sh.PatternConstraintComponent":
		if patVal, ok := c.Payload["Pattern"].(string); ok {
			pat := patVal
			if flagsVal, ok := c.Payload["Flags"].(string); ok && flagsVal != "" {
				pat = "(?" + flagsVal + ")" + pat
			}
			if re, err := regexp.Compile(pat); err == nil {
				if s, ok := getString(field); ok && !re.MatchString(s) {
					return fmt.Sprintf("Pattern violation: %q does not match %q", s, patVal)
				}
			}
		}

	case "sh.NodeKindConstraintComponent":
		if getCount(field) == 0 {
			return ""
		}
		nk, ok := c.Payload["NodeKind"].(string)
		if !ok {
			return ""
		}
		// NodeKind=IRI and NodeKind=BlankNodeOrIRI are dropped by the simplifier
		// (see import_shacl_rules.go) because all IRI-typed CIM properties are
		// already typed as Go reference fields. Only Literal can fail in practice.
		if strings.TrimPrefix(nk, "sh.") == "Literal" && !isLiteral(field) {
			return "NodeKind violation: value is not a Literal"
		}

	case "sh.ClassConstraintComponent":
		if getCount(field) == 0 {
			return ""
		}
		want, ok := c.Payload["Class"].(string)
		if !ok {
			return ""
		}
		got := getReferencedClass(field, dataset)
		if got == "" {
			return ""
		}
		if got != want {
			return fmt.Sprintf("Class violation: target is %s, want %s", got, want)
		}

	case "sh.NotConstraintComponent":
		shapes, ok := c.Payload["ShapeRef"].([]interface{})
		if !ok || len(shapes) == 0 {
			return ""
		}
		// The wrapped shape conforms iff ALL its constraints conform. The Not
		// fires (= violation) precisely when the wrapped shape conforms.
		for _, ciRaw := range shapes {
			ciMap, ok := ciRaw.(map[string]interface{})
			if !ok {
				return ""
			}
			component, _ := ciMap["component"].(string)
			payload, _ := ciMap["payload"].(map[string]interface{})
			if checkConstraint(field, ConstraintInfo{Component: component, Payload: payload}, dataset) != "" {
				return ""
			}
		}
		return "Not violation: value conforms to forbidden shape"
	}
	return ""
}

func isIRI(v reflect.Value) bool {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	if mrid := v.FieldByName("MRID"); mrid.IsValid() && mrid.Kind() == reflect.String && mrid.String() != "" {
		return true
	}
	if id := v.FieldByName("Id"); id.IsValid() && id.Kind() == reflect.String && id.String() != "" {
		return true
	}
	return false
}

func isLiteral(v reflect.Value) bool {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// getReferencedClass returns the cim.* type name of the dataset element that
// the field's MRID/Id points to. Returns "" if the field is not a reference,
// has no dataset entry, or no dataset was supplied.
func getReferencedClass(field reflect.Value, dataset *cimgostructs.CIMElementList) string {
	if dataset == nil {
		return ""
	}
	for field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return ""
		}
		field = field.Elem()
	}
	if field.Kind() != reflect.Struct {
		return ""
	}
	idField := field.FieldByName("MRID")
	if !idField.IsValid() {
		idField = field.FieldByName("Id")
	}
	if !idField.IsValid() || idField.Kind() != reflect.String {
		return ""
	}
	id := strings.TrimPrefix(idField.String(), "#")
	if obj, ok := dataset.Elements[id]; ok {
		return getCIMTypeName(obj)
	}
	return ""
}

func anyToFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	}
	return 0
}

func getCount(v reflect.Value) int {
	if !v.IsValid() {
		return 0
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return v.Len()
	case reflect.Ptr:
		if v.IsNil() {
			return 0
		}
		return 1
	default:
		if v.IsZero() {
			return 0
		}
		return 1
	}
}

func getFloat(v reflect.Value) (float64, bool) {
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return v.Float(), true
	case reflect.Int, reflect.Int64:
		return float64(v.Int()), true
	}
	return 0, false
}

func getString(v reflect.Value) (string, bool) {
	switch v.Kind() {
	case reflect.String:
		return v.String(), true
	case reflect.Ptr:
		if !v.IsNil() {
			return getString(v.Elem())
		}
	}
	return "", false
}

// resolvePathSegments walks a struct following a sequence of property path
// segments (split from " / ") and returns all terminal reflect.Values.
// Each segment's Go field name is the last dot-separated token, capitalised.
// Pointers are dereferenced and slices are fanned out at every step.
func resolvePathSegments(val reflect.Value, segments []string) []reflect.Value {
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		var results []reflect.Value
		for i := 0; i < val.Len(); i++ {
			results = append(results, resolvePathSegments(val.Index(i), segments)...)
		}
		return results
	}
	if len(segments) == 0 {
		return []reflect.Value{val}
	}
	if val.Kind() != reflect.Struct {
		return nil
	}
	seg := strings.TrimSpace(segments[0])
	fieldName := seg
	if dot := strings.LastIndex(seg, "."); dot != -1 {
		fieldName = seg[dot+1:]
	}
	if fieldName == "" {
		return nil
	}
	fieldName = strings.ToUpper(fieldName[:1]) + fieldName[1:]
	// Prefer the renamed pointer field (e.g. "IdentifiedObject_") when it
	// exists alongside an embedded struct of the same base name. The generator
	// suffixes the relation field with "_" to disambiguate from the embedded
	// supertype, and that renamed field is the one carrying the rdf:resource
	// MRID we want to traverse.
	field := val.FieldByName(fieldName + "_")
	if !field.IsValid() {
		field = val.FieldByName(fieldName)
	}
	if !field.IsValid() {
		return nil
	}
	return resolvePathSegments(field, segments[1:])
}

// checkHasValue reports a violation if none of the terminal values matches the
// expected value from the sh:hasValue payload. The expected value may be an IRI
// stored as "<http://...>" or a plain string literal.
func checkHasValue(fields []reflect.Value, c ConstraintInfo) string {
	want, ok := c.Payload["Value"].(string)
	if !ok {
		return ""
	}
	want = strings.TrimPrefix(strings.TrimSuffix(want, ">"), "<")
	for _, f := range fields {
		if hasValue(f, want) {
			return ""
		}
	}
	return fmt.Sprintf("HasValue violation: want %q", want)
}

func hasValue(v reflect.Value, want string) bool {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}
	if v.Kind() == reflect.String {
		return v.String() == want
	}
	if v.Kind() == reflect.Struct {
		if id := v.FieldByName("Id"); id.IsValid() && id.Kind() == reflect.String {
			return id.String() == want
		}
	}
	return false
}

func countInverseRelations(field reflect.Value, path string, dataset *cimgostructs.CIMElementList) int {
	if dataset == nil {
		return 0
	}
	idField := field.FieldByName("Id")
	if !idField.IsValid() {
		return 0
	}
	id := idField.String()

	parts := strings.Split(strings.TrimPrefix(path, "^"), ".")
	if len(parts) < 3 {
		return 0
	}
	targetClass := parts[1]
	targetField := parts[2]

	count := 0
	for _, other := range dataset.Elements {
		if getCIMTypeName(other) != "cim."+targetClass {
			continue
		}
		otherVal := reflect.ValueOf(other)
		if otherVal.Kind() == reflect.Ptr {
			otherVal = otherVal.Elem()
		}
		f := otherVal.FieldByName(targetField)
		if !f.IsValid() {
			continue
		}
		if f.Kind() == reflect.Ptr && !f.IsNil() {
			elem := f.Elem()
			refId := elem.FieldByName("Id")
			if !refId.IsValid() {
				// CIM reference stubs use MRID instead of Id
				refId = elem.FieldByName("MRID")
			}
			if refId.IsValid() {
				// rdf:resource="#_uuid" → strip leading "#" before comparing to rdf:ID="_uuid"
				refVal := strings.TrimPrefix(refId.String(), "#")
				if refVal == id {
					count++
				}
			}
		}
	}
	return count
}

func checkOrOption(field reflect.Value, shape any, dataset *cimgostructs.CIMElementList) bool {
	// A shape in an OR constraint is a list of constraints (ANDed)
	// After JSON unmarshaling, it's a []interface{} where each item is a map matching ConstraintInfo
	constraints, ok := shape.([]interface{})
	if !ok {
		return false
	}

	for _, ciRaw := range constraints {
		ciMap, ok := ciRaw.(map[string]interface{})
		if !ok {
			return false
		}

		component, _ := ciMap["component"].(string)
		var path []string
		if rawPath, ok := ciMap["path"].([]interface{}); ok {
			for _, p := range rawPath {
				if s, ok := p.(string); ok {
					path = append(path, s)
				}
			}
		}
		payload, _ := ciMap["payload"].(map[string]interface{})
		ci := ConstraintInfo{
			Component: component,
			Path:      path,
			Payload:   payload,
		}

		if checkConstraint(field, ci, dataset) != "" {
			return false
		}
	}
	return true
}
