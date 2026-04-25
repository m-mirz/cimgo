package encoding

import (
	"bytes"
	"cimgo/encoding/cimgostructs"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Models for the struct JSON documentation
type ConstraintInfo struct {
	Path      string         `json:"path"`
	Severity  string         `json:"severity"`
	Message   string         `json:"message"`
	Component string         `json:"component"`
	Payload   map[string]any `json:"payload"`
	Details   string         `json:"details"`
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

func loadAllRules(t *testing.T, structDir string) map[string]ClassInfo {
	rules := make(map[string]ClassInfo)
	files, err := filepath.Glob(filepath.Join(structDir, "*.json"))
	if err != nil {
		t.Fatal(err)
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

func getCIMTypeName(obj interface{}) string {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return "cim." + t.Name()
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
		fieldName := attrRule.Name
		if lastDot := strings.LastIndex(fieldName, "."); lastDot != -1 {
			fieldName = fieldName[lastDot+1:]
		}
		fieldName = strings.ToUpper(fieldName[:1]) + fieldName[1:]

		field := val.FieldByName(fieldName)
		if !field.IsValid() {
			continue
		}

		for _, constraint := range attrRule.Constraints {
			v := checkConstraint(field, constraint, dataset)
			if v != "" {
				violations = append(violations, fmt.Sprintf("[%s] %s: %s", cimType, attrRule.Name, v))
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
			min := int(reflect.ValueOf(minVal).Float()) // JSON numbers are float64
			var count int
			if c.Path != "" && strings.HasPrefix(c.Path, "^") {
				count = countInverseRelations(field, c.Path, dataset)
			} else {
				count = getCount(field)
			}
			if count < min {
				return fmt.Sprintf("MinCount violation: got %d, want %d", count, min)
			}
		}

	case "sh.MaxCountConstraintComponent":
		if maxVal, ok := c.Payload["MaxCount"]; ok {
			max := int(reflect.ValueOf(maxVal).Float())
			var count int
			if c.Path != "" && strings.HasPrefix(c.Path, "^") {
				count = countInverseRelations(field, c.Path, dataset)
			} else {
				count = getCount(field)
			}
			if count > max {
				return fmt.Sprintf("MaxCount violation: got %d, want %d", count, max)
			}
		}

	case "sh.MinInclusiveConstraintComponent":
		if minVal, ok := c.Payload["Value"]; ok {
			min := reflect.ValueOf(minVal).Float()
			val, ok := getFloat(field)
			if ok && val < min {
				return fmt.Sprintf("MinInclusive violation: got %v, want %v", val, min)
			}
		}
	}
	return ""
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
			refId := f.Elem().FieldByName("Id")
			if refId.IsValid() && refId.String() == id {
				count++
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
		
		// Map map[string]interface{} to ConstraintInfo
		ci := ConstraintInfo{
			Component: ciMap["component"].(string),
			Path:      ciMap["path"].(string),
			Payload:   ciMap["payload"].(map[string]interface{}),
			Details:   ciMap["details"].(string),
		}

		if checkConstraint(field, ci, dataset) != "" {
			return false
		}
	}
	return true
}

func TestValidateCIMData(t *testing.T) {
	rules := loadAllRules(t, "../pages/docs/struct")
	if len(rules) == 0 {
		t.Skip("No rules found in ../pages/docs/struct")
	}

	dataFiles := []string{
		"../testdata/test_001.xml",
		"../testdata/test_009_EQ.xml",
	}

	mergedCIMData := cimgostructs.NewCIMElementList()
	for _, file := range dataFiles {
		b, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		DecodeProfile(bytes.NewReader(b), mergedCIMData)
	}

	var allViolations []string
	for id, obj := range mergedCIMData.Elements {
		violations := validateObject(t, obj, rules, mergedCIMData)
		for _, v := range violations {
			allViolations = append(allViolations, fmt.Sprintf("Object %s: %s", id, v))
		}
	}

	if len(allViolations) > 0 {
		t.Errorf("Found %d validation violations:", len(allViolations))
		for _, v := range allViolations {
			t.Log(v)
		}
	} else {
		t.Log("No validation violations found.")
	}
}
