package validation

import (
	"cimgo/cimgostructs"
	"strings"
)

const (
	profBase = "http://iec.ch/TC57/ns/CIM/"
	profEQ   = profBase + "CoreEquipment-EU/3.0"
	profEQBD = profBase + "EquipmentBoundary-EU/3.0"
	profDY   = profBase + "Dynamics-EU/1.0"
	profDL   = profBase + "DiagramLayout-EU/3.0"
	profSC   = profBase + "ShortCircuit-EU/3.0"
	profOP   = profBase + "Operation-EU/3.0"
	profGL   = profBase + "GeographicalLocation-EU/3.0"
	profSV   = profBase + "StateVariables-EU/3.0"
	profTP   = profBase + "Topology-EU/3.0"
	profSSH  = profBase + "SteadyStateHypothesis-EU/3.0"
)

func profileURI(m *cimgostructs.Model) string {
	for _, p := range m.Profile {
		if p = strings.TrimSpace(p); p != "" {
			return p
		}
	}
	return ""
}

// dependentOnProfiles returns the resolved profile URI for each DependentOn entry.
// Entries whose referenced model is not in the dataset are represented as "external".
func dependentOnProfiles(m *cimgostructs.Model, dataset *cimgostructs.CIMElementList) []string {
	if len(m.DependentOn) == 0 {
		return nil
	}
	out := make([]string, 0, len(m.DependentOn))
	for _, ref := range m.DependentOn {
		mrid := strings.TrimPrefix(ref.MRID, "#")
		if fm, ok := dataset.FullModels[mrid]; ok {
			out = append(out, profileURI(&fm.Model))
		} else if dm, ok := dataset.DifferenceModels[mrid]; ok {
			out = append(out, profileURI(&dm.Model))
		} else {
			out = append(out, "external")
		}
	}
	return out
}

// hasValue reports whether any entry in deps equals target.
func hasValue(deps []string, target string) bool {
	for _, d := range deps {
		if d == target {
			return true
		}
	}
	return false
}

// allInSet reports whether every resolved (non-external) entry in deps is in allowed.
func allInSet(deps []string, allowed map[string]bool) bool {
	for _, d := range deps {
		if d != "external" && !allowed[d] {
			return false
		}
	}
	return true
}

// datasetHasProfile reports whether the dataset contains at least one loaded model with the given profile URI.
func datasetHasProfile(dataset *cimgostructs.CIMElementList, prof string) bool {
	for _, fm := range dataset.FullModels {
		if profileURI(&fm.Model) == prof {
			return true
		}
	}
	for _, dm := range dataset.DifferenceModels {
		if profileURI(&dm.Model) == prof {
			return true
		}
	}
	return false
}

func prof10violation(id, msg, severity string) Violation {
	return Violation{
		ObjectID:    id,
		Class:       "FullModel",
		Property:    "Model.DependentOn",
		Message:     msg,
		Severity:    severity,
		RuleID:      "prof10:PROF10",
		Name:        "PROF10",
		Description: "CGMES instance file (distribution) dependency shall be declared by md:Model.DependentOn in the header according to Figure 1 and the associated rules.",
	}
}

// ValidateProf10HeaderRules checks PROF10 file-header dependency constraints
// from 61970-600-1_Prof10-Header-AP-Con-Complex-SHACL.ttl.
func ValidateProf10HeaderRules(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	for id, fm := range dataset.FullModels {
		violations = append(violations, checkProf10Model(id, &fm.Model, dataset)...)
	}
	return violations
}

func checkProf10Model(id string, m *cimgostructs.Model, dataset *cimgostructs.CIMElementList) []Violation {
	switch profileURI(m) {
	case profEQ:
		return checkProf10EQ(id, m, dataset)
	case profDY:
		return checkProf10DY(id, m, dataset)
	case profDL:
		return checkProf10DL(id, m, dataset)
	case profSC:
		return checkProf10SC(id, m, dataset)
	case profOP:
		return checkProf10OP(id, m, dataset)
	case profGL:
		return checkProf10GL(id, m, dataset)
	case profSV:
		return checkProf10SV(id, m, dataset)
	case profTP:
		return checkProf10TP(id, m, dataset)
	case profSSH:
		return checkProf10SSH(id, m, dataset)
	}
	return nil
}

const msgEQ = "The EQ does not have reference to EQBD. The file header dependencies cardinalities and types for EQ profile are not according to PROF10."

// checkProf10EQ: sh:hasValue EQBD, sh:Info, no sh:minCount.
// Passes if any dep is EQBD, or if any dep is external (can't verify).
func checkProf10EQ(id string, m *cimgostructs.Model, dataset *cimgostructs.CIMElementList) []Violation {
	deps := dependentOnProfiles(m, dataset)
	if hasValue(deps, profEQBD) || hasValue(deps, "external") {
		return nil
	}
	return []Violation{prof10violation(id, msgEQ, "sh:Info")}
}

const msgDY = "The file header dependencies cardinalities and types for DY profile are not according to PROF10."

// checkProf10DY: sh:minCount 1, sh:hasValue EQ, sh:Violation.
func checkProf10DY(id string, m *cimgostructs.Model, dataset *cimgostructs.CIMElementList) []Violation {
	deps := dependentOnProfiles(m, dataset)
	if len(deps) == 0 {
		return []Violation{prof10violation(id, msgDY, "sh:Violation")}
	}
	if hasValue(deps, profEQ) {
		return nil
	}
	if hasValue(deps, "external") && !datasetHasProfile(dataset, profEQ) {
		return nil
	}
	return []Violation{prof10violation(id, msgDY, "sh:Violation")}
}

const msgDL = "The file header dependencies cardinalities and types for DL profile are not according to PROF10."

// checkProf10DL: sh:in {DY,TP,EQ,SC,OP}, no sh:minCount, sh:Violation.
// External deps pass vacuously; each resolved dep must be in the allowed set.
func checkProf10DL(id string, m *cimgostructs.Model, dataset *cimgostructs.CIMElementList) []Violation {
	deps := dependentOnProfiles(m, dataset)
	allowed := map[string]bool{profDY: true, profTP: true, profEQ: true, profSC: true, profOP: true}
	if !allInSet(deps, allowed) {
		return []Violation{prof10violation(id, msgDL, "sh:Violation")}
	}
	return nil
}

const msgSC = "The file header dependencies cardinalities and types for SC profile are not according to PROF10."

// checkProf10SC: sh:minCount 1, sh:in {EQ,EQBD,OP}, sh:Violation.
func checkProf10SC(id string, m *cimgostructs.Model, dataset *cimgostructs.CIMElementList) []Violation {
	deps := dependentOnProfiles(m, dataset)
	if len(deps) == 0 {
		return []Violation{prof10violation(id, msgSC, "sh:Violation")}
	}
	allowed := map[string]bool{profEQ: true, profEQBD: true, profOP: true}
	if !allInSet(deps, allowed) {
		return []Violation{prof10violation(id, msgSC, "sh:Violation")}
	}
	return nil
}

const msgOP = "The file header dependencies cardinalities and types for OP profile are not according to PROF10."

// checkProf10OP: sh:minCount 1, sh:in {EQ,EQBD,SC}, sh:Violation.
func checkProf10OP(id string, m *cimgostructs.Model, dataset *cimgostructs.CIMElementList) []Violation {
	deps := dependentOnProfiles(m, dataset)
	if len(deps) == 0 {
		return []Violation{prof10violation(id, msgOP, "sh:Violation")}
	}
	allowed := map[string]bool{profEQ: true, profEQBD: true, profSC: true}
	if !allInSet(deps, allowed) {
		return []Violation{prof10violation(id, msgOP, "sh:Violation")}
	}
	return nil
}

const msgGL = "The file header dependencies cardinalities and types for GL profile are not according to PROF10."

// checkProf10GL: sh:in {EQBD,EQ,SC,OP}, no sh:minCount, sh:Violation.
// External deps pass vacuously; each resolved dep must be in the allowed set.
func checkProf10GL(id string, m *cimgostructs.Model, dataset *cimgostructs.CIMElementList) []Violation {
	deps := dependentOnProfiles(m, dataset)
	allowed := map[string]bool{profEQBD: true, profEQ: true, profSC: true, profOP: true}
	if !allInSet(deps, allowed) {
		return []Violation{prof10violation(id, msgGL, "sh:Violation")}
	}
	return nil
}

const msgSV = "The file header dependencies cardinalities and types for SV profile are not according to PROF10."

// checkProf10SV: sh:minCount 1, sh:hasValue TP, sh:Violation.
func checkProf10SV(id string, m *cimgostructs.Model, dataset *cimgostructs.CIMElementList) []Violation {
	deps := dependentOnProfiles(m, dataset)
	if len(deps) == 0 {
		return []Violation{prof10violation(id, msgSV, "sh:Violation")}
	}
	if hasValue(deps, profTP) {
		return nil
	}
	if hasValue(deps, "external") && !datasetHasProfile(dataset, profTP) {
		return nil
	}
	return []Violation{prof10violation(id, msgSV, "sh:Violation")}
}

const msgTP = "The file header dependencies cardinalities and types for TP profile are not according to PROF10."

// checkProf10TP: sh:minCount 1, sh:hasValue SSH, sh:Violation.
func checkProf10TP(id string, m *cimgostructs.Model, dataset *cimgostructs.CIMElementList) []Violation {
	deps := dependentOnProfiles(m, dataset)
	if len(deps) == 0 {
		return []Violation{prof10violation(id, msgTP, "sh:Violation")}
	}
	if hasValue(deps, profSSH) {
		return nil
	}
	if hasValue(deps, "external") && !datasetHasProfile(dataset, profSSH) {
		return nil
	}
	return []Violation{prof10violation(id, msgTP, "sh:Violation")}
}

const msgSSH = "The file header dependencies cardinalities and types for SSH profile are not according to PROF10."

// checkProf10SSH: sh:minCount 1, sh:hasValue EQ, sh:Violation.
func checkProf10SSH(id string, m *cimgostructs.Model, dataset *cimgostructs.CIMElementList) []Violation {
	deps := dependentOnProfiles(m, dataset)
	if len(deps) == 0 {
		return []Violation{prof10violation(id, msgSSH, "sh:Violation")}
	}
	if hasValue(deps, profEQ) {
		return nil
	}
	if hasValue(deps, "external") && !datasetHasProfile(dataset, profEQ) {
		return nil
	}
	return []Violation{prof10violation(id, msgSSH, "sh:Violation")}
}
