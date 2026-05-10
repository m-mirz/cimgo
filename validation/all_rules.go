package validation

import (
	"cimgo/cimgostructs"
	"cimgo/shaclmodel"
)

func ValidateCommonRules(dataset *cimgostructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, ValidateCommonRulesSPARQL(dataset)...)
	return violations
}

func ValidateEQProfile(dataset *cimgostructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, ValidateEQProfileSPARQL(dataset)...)
	violations = append(violations, ValidateGeneratedEquipmentProfileSHACL(dataset)...)
	return violations
}

func ValidateEQNotSolvedMASProfile(dataset *cimgostructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, ValidateEQNotSolvedMASProfileSPARQL(dataset)...)
	return violations
}

func ValidateSSHProfile(dataset *cimgostructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, ValidateSSHProfile(dataset)...)
	violations = append(violations, ValidateGeneratedSteadystatehypothesisProfileSHACL(dataset)...)
	return violations
}

func ValidateSSHNotSolvedMASProfile(dataset *cimgostructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, ValidateSSHNotSolvedMASProfile(dataset)...)
	violations = append(violations, ValidateGeneratedSteadystatehypothesisNotsolvedmasProfileSHACL(dataset)...)
	return violations
}

func ValidateDYProfile(dataset *cimgostructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, ValidateDYProfileSPARQL(dataset)...)
	violations = append(violations, ValidateGeneratedDynamicsProfileSHACL(dataset)...)
	return violations
}

func ValidateSCProfile(dataset *cimgostructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, ValidateSCProfileSPARQL(dataset)...)
	violations = append(violations, ValidateGeneratedShortcircuitProfileSHACL(dataset)...)
	return violations
}

func ValidateSCNotSolvedMASProfile(dataset *cimgostructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, ValidateSCNotSolvedMASProfileSPARQL(dataset)...)
	return violations
}

func ValidateSVProfile(dataset *cimgostructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, ValidateSVProfileSPARQL(dataset)...)
	violations = append(violations, ValidateGeneratedStatevariablesProfileSHACL(dataset)...)
	return violations
}

func ValidateSVSolvedMASProfile(dataset *cimgostructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, ValidateSVSolvedMASProfileSPARQL(dataset)...)
	return violations
}

func ValidateDLProfile(dataset *cimgostructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, ValidateDLProfileSPARQL(dataset)...)
	violations = append(violations, ValidateGeneratedDiagramlayoutProfileSHACL(dataset)...)
	return violations
}

func ValidateDLNotSolvedMASProfile(dataset *cimgostructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, ValidateGeneratedDiagramlayoutNotsolvedmasProfileSHACL(dataset)...)
	return violations
}

func ValidateTPProfile(dataset *cimgostructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, ValidateGeneratedTopologyProfileSHACL(dataset)...)
	return violations
}

func ValidateTPNotSolvedMASProfile(dataset *cimgostructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, ValidateTPNotSolvedMASProfileSPARQL(dataset)...)
	return violations
}

func ValidateEQBDProfile(dataset *cimgostructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, ValidateEQBDProfile(dataset)...)
	violations = append(violations, ValidateGeneratedEquipmentboundaryProfileSHACL(dataset)...)
	return violations
}

func ValidateOPProfile(dataset *cimgostructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, ValidateOPProfileSPARQL(dataset)...)
	violations = append(violations, ValidateGeneratedOperationProfileSHACL(dataset)...)
	return violations
}

type Config struct {
	Profiles  []string
	Solved    bool
	NotSolved bool
	Common    bool
}

func RunValidation(dataset *cimgostructs.CIMElementList, cfg Config) []shaclmodel.Violation {
	var violations []shaclmodel.Violation

	profileSelected := func(p string) bool {
		if len(cfg.Profiles) == 0 {
			return true
		}
		for _, sp := range cfg.Profiles {
			if sp == p {
				return true
			}
		}
		return false
	}

	if cfg.Common {
		violations = append(violations, ValidateCommonRulesSPARQL(dataset)...)
	}

	if profileSelected("EQ") {
		violations = append(violations, ValidateEQProfile(dataset)...)
		if cfg.NotSolved {
			violations = append(violations, ValidateEQNotSolvedMASProfile(dataset)...)
		}
	}
	if profileSelected("SSH") {
		violations = append(violations, ValidateSSHProfile(dataset)...)
		if cfg.NotSolved {
			violations = append(violations, ValidateSSHNotSolvedMASProfile(dataset)...)
		}
	}
	if profileSelected("TP") {
		violations = append(violations, ValidateTPProfile(dataset)...)
		if cfg.NotSolved {
			violations = append(violations, ValidateTPNotSolvedMASProfile(dataset)...)
		}
	}
	if profileSelected("DY") {
		violations = append(violations, ValidateDYProfile(dataset)...)
	}
	if profileSelected("SC") {
		violations = append(violations, ValidateSCProfile(dataset)...)
		if cfg.NotSolved {
			violations = append(violations, ValidateSCNotSolvedMASProfile(dataset)...)
		}
	}
	if profileSelected("SV") {
		violations = append(violations, ValidateSVProfile(dataset)...)
		if cfg.Solved {
			violations = append(violations, ValidateSVSolvedMASProfile(dataset)...)
		}
	}
	if profileSelected("DL") {
		violations = append(violations, ValidateDLProfile(dataset)...)
		if cfg.NotSolved {
			violations = append(violations, ValidateDLNotSolvedMASProfile(dataset)...)
		}
	}
	if profileSelected("EQBD") {
		violations = append(violations, ValidateEQBDProfile(dataset)...)
	}
	if profileSelected("OP") {
		violations = append(violations, ValidateOPProfile(dataset)...)
	}

	return violations
}
