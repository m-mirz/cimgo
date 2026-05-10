package validation

import (
	"cimgo/cimgostructs"
	"cimgo/shaclgen"
	"cimgo/shaclmodel"
	"reflect"
)

// Violation is re-exported from shaclmodel so existing callers of
// validation.Violation keep compiling. The actual type lives in shaclmodel
// (a leaf package) so the generated shaclgen package can return it without
// pulling in validation, which would be a cycle.
type Violation = shaclmodel.Violation

func goTypeName(obj interface{}) string {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

// Hand-written profile checks. Each Validate<Profile>Profile bundles only the
// SPARQL-style cross-cutting rules that don't reduce to a single attribute
// constraint; per-attribute SHACL checks come from the generated profile
// orchestrators (see generated_index.go) and are run separately by
// ValidateAllProfiles.

// ValidateAllProfiles runs every generated SHACL profile orchestrator plus
// every hand-written profile-level check. The two are independent: the
// generated set comes from shaclgen.ValidateAllGeneratedProfiles; the
// hand-written ones are the Validate<Profile>Profile functions above.
func ValidateAllProfiles(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	violations = append(violations, shaclgen.ValidateAllGeneratedProfiles(dataset)...)
	violations = append(violations, ValidateCommonRulesSolvedMASProfile(dataset)...)
	violations = append(violations, ValidateCommonRules(dataset)...)
	violations = append(violations, ValidateEquipmentProfile(dataset)...)
	violations = append(violations, ValidateEquipmentNotSolvedMASProfile(dataset)...)
	violations = append(violations, ValidateSSHProfile(dataset)...)
	violations = append(violations, ValidateSSHNotSolvedMASProfile(dataset)...)
	violations = append(violations, ValidateDynamicsProfile(dataset)...)
	violations = append(violations, ValidateShortCircuitProfile(dataset)...)
	violations = append(violations, ValidateShortCircuitNotSolvedMASProfile(dataset)...)
	violations = append(violations, ValidateStateVariablesProfile(dataset)...)
	violations = append(violations, ValidateStateVariablesSolvedMASProfile(dataset)...)
	violations = append(violations, ValidateDiagramLayoutProfile(dataset)...)
	violations = append(violations, ValidateTopologyNotSolvedMASProfile(dataset)...)
	violations = append(violations, ValidateEquipmentBoundaryProfile(dataset)...)
	violations = append(violations, ValidateOperationProfile(dataset)...)
	return violations
}
