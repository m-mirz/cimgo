package validation

import (
	"cimgo/cimgostructs"
	"cimgo/shaclmodel"
	"slices"
	"sync"
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
	violations = append(violations, ValidateSSHProfileSPARQL(dataset)...)
	violations = append(violations, ValidateGeneratedSteadystatehypothesisProfileSHACL(dataset)...)
	return violations
}

func ValidateSSHNotSolvedMASProfile(dataset *cimgostructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, ValidateSSHNotSolvedMASProfileSPARQL(dataset)...)
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
	violations = append(violations, ValidateEQBDProfileSPARQL(dataset)...)
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
	Profiles           []string
	Solved             bool
	NotSolved          bool
	Common             bool
	Quality            bool              // enables CIMdesk-style modeling quality checks
	SilencedRules      []string
	EQBDBaseVoltageIDs map[string]struct{} // enables EQBD2 check when non-nil
}

func RunValidation(dataset *cimgostructs.CIMElementList, cfg Config) []shaclmodel.Violation {
	profileSelected := func(p string) bool {
		return len(cfg.Profiles) == 0 || slices.Contains(cfg.Profiles, p)
	}

	type fn func() []shaclmodel.Violation
	var validators []fn

	if cfg.Common {
		validators = append(validators, func() []shaclmodel.Violation { return ValidateCommonRulesSPARQL(dataset) })
	}
	if profileSelected("EQ") {
		validators = append(validators, func() []shaclmodel.Violation { return ValidateEQProfile(dataset) })
		if cfg.NotSolved {
			validators = append(validators, func() []shaclmodel.Violation { return ValidateEQNotSolvedMASProfile(dataset) })
		}
	}
	if profileSelected("SSH") {
		validators = append(validators, func() []shaclmodel.Violation { return ValidateSSHProfile(dataset) })
		if cfg.NotSolved {
			validators = append(validators, func() []shaclmodel.Violation { return ValidateSSHNotSolvedMASProfile(dataset) })
		}
	}
	if profileSelected("TP") {
		validators = append(validators, func() []shaclmodel.Violation { return ValidateTPProfile(dataset) })
		if cfg.NotSolved {
			validators = append(validators, func() []shaclmodel.Violation { return ValidateTPNotSolvedMASProfile(dataset) })
		}
	}
	if profileSelected("DY") {
		validators = append(validators, func() []shaclmodel.Violation { return ValidateDYProfile(dataset) })
	}
	if profileSelected("SC") {
		validators = append(validators, func() []shaclmodel.Violation { return ValidateSCProfile(dataset) })
		if cfg.NotSolved {
			validators = append(validators, func() []shaclmodel.Violation { return ValidateSCNotSolvedMASProfile(dataset) })
		}
	}
	if profileSelected("SV") {
		validators = append(validators, func() []shaclmodel.Violation { return ValidateSVProfile(dataset) })
		if cfg.Solved {
			validators = append(validators, func() []shaclmodel.Violation { return ValidateSVSolvedMASProfile(dataset) })
		}
	}
	if profileSelected("DL") {
		validators = append(validators, func() []shaclmodel.Violation { return ValidateDLProfile(dataset) })
		if cfg.NotSolved {
			validators = append(validators, func() []shaclmodel.Violation { return ValidateDLNotSolvedMASProfile(dataset) })
		}
	}
	if profileSelected("EQBD") {
		validators = append(validators, func() []shaclmodel.Violation { return ValidateEQBDProfile(dataset) })
		if cfg.EQBDBaseVoltageIDs != nil {
			validators = append(validators, func() []shaclmodel.Violation {
				return CheckBaseVoltageInEQBD(dataset, cfg.EQBDBaseVoltageIDs)
			})
		}
	}
	if profileSelected("OP") {
		validators = append(validators, func() []shaclmodel.Violation { return ValidateOPProfile(dataset) })
	}
	if cfg.Quality {
		validators = append(validators, func() []shaclmodel.Violation { return ValidateCIMdeskQualityChecks(dataset) })
	}

	results := make([][]shaclmodel.Violation, len(validators))
	var wg sync.WaitGroup
	wg.Add(len(validators))
	for i, v := range validators {
		go func(i int, v fn) {
			defer wg.Done()
			results[i] = v()
		}(i, v)
	}
	wg.Wait()

	var violations []shaclmodel.Violation
	for _, r := range results {
		violations = append(violations, r...)
	}

	if len(cfg.SilencedRules) > 0 {
		filtered := make([]shaclmodel.Violation, 0, len(violations))
		silenced := make(map[string]bool)
		for _, r := range cfg.SilencedRules {
			silenced[r] = true
		}
		for _, v := range violations {
			if !silenced[v.RuleID] {
				filtered = append(filtered, v)
			}
		}
		violations = filtered
	}

	return violations
}
