package validation

import (
	"cimgo/cimstructs"
	"cimgo/shaclgen"
	"cimgo/shaclmodel"
)

func ValidateGeneratedDiagramlayoutProfileSHACL(dataset *cimstructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, shaclgen.ValidateGeneratedDiagramlayout61970301ComplexProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedDiagramlayout61970453ComplexProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedDiagramlayout619706002SimpleProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedDiagramlayout619706002ComplexInverseassociationProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedDiagramlayout61970453ComplexExplicitCrossprofileProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedDiagramlayout61970453ComplexImplicitCrossprofileProfile(dataset)...)
	return violations
}

func ValidateGeneratedDiagramlayoutNotsolvedmasProfileSHACL(dataset *cimstructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, shaclgen.ValidateGeneratedDiagramlayout61970301ComplexNotsolvedmasProfile(dataset)...)
	return violations
}

func ValidateGeneratedDynamicsProfileSHACL(dataset *cimstructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, shaclgen.ValidateGeneratedDynamics61970302ComplexProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedDynamics61970457ComplexExplicitCrossprofileProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedDynamics61970457ComplexImplicitCrossprofileProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedDynamics619706002ComplexInverseassociationProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedDynamics619706002SimpleProfile(dataset)...)
	return violations
}

func ValidateGeneratedEquipmentProfileSHACL(dataset *cimstructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, shaclgen.ValidateGeneratedEquipment61970301ComplexProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedEquipment61970452ComplexProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedEquipment619706001ComplexProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedEquipment619706002ComplexInverseassociationProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedEquipment619706002ComplexProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedEquipment619706002SimpleProfile(dataset)...)
	return violations
}

func ValidateGeneratedEquipmentboundaryProfileSHACL(dataset *cimstructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, shaclgen.ValidateGeneratedEquipmentboundary61970301ComplexProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedEquipmentboundary619706002ComplexInverseassociationProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedEquipmentboundary619706002SimpleProfile(dataset)...)
	return violations
}

func ValidateGeneratedGeographicallocationProfileSHACL(dataset *cimstructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, shaclgen.ValidateGeneratedGeographicallocation6196813ComplexProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedGeographicallocation619706002ComplexExplicitCrossprofileProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedGeographicallocation619706002ComplexImplicitCrossprofileProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedGeographicallocation619706002ComplexInverseassociationProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedGeographicallocation619706002SimpleProfile(dataset)...)
	return violations
}

func ValidateGeneratedOperationProfileSHACL(dataset *cimstructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, shaclgen.ValidateGeneratedOperation61970301ComplexProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedOperation61970452ComplexProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedOperation619706002ComplexExplicitCrossprofileProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedOperation619706002ComplexImplicitCrossprofileProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedOperation619706002ComplexInverseassociationProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedOperation619706002SimpleProfile(dataset)...)
	return violations
}

func ValidateGeneratedShortcircuitProfileSHACL(dataset *cimstructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, shaclgen.ValidateGeneratedShortcircuit61970301ComplexProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedShortcircuit61970452ComplexCrossprofileProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedShortcircuit619706002SimpleProfile(dataset)...)
	return violations
}

func ValidateGeneratedStatevariablesProfileSHACL(dataset *cimstructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, shaclgen.ValidateGeneratedStatevariables61970301ComplexProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedStatevariables61970456ComplexExplicitCrossprofileProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedStatevariables61970456ComplexImplicitCrossprofileProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedStatevariables61970456ComplexProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedStatevariables619706002ComplexInverseassociationProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedStatevariables619706002SimpleProfile(dataset)...)
	return violations
}

func ValidateGeneratedSteadystatehypothesisProfileSHACL(dataset *cimstructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, shaclgen.ValidateGeneratedSteadystatehypothesis61970301ComplexProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedSteadystatehypothesis61970456ComplexProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedSteadystatehypothesis619706002SimpleProfile(dataset)...)
	return violations
}

func ValidateGeneratedSteadystatehypothesisNotsolvedmasProfileSHACL(dataset *cimstructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, shaclgen.ValidateGeneratedSteadystatehypothesis61970301ComplexNotsolvedmasProfile(dataset)...)
	return violations
}

func ValidateGeneratedTopologyProfileSHACL(dataset *cimstructs.CIMElementList) []shaclmodel.Violation {
	var violations []shaclmodel.Violation
	violations = append(violations, shaclgen.ValidateGeneratedTopology61970456ComplexExplicitCrossprofileProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedTopology61970456ComplexImplicitCrossprofileProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedTopology61970456ComplexProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedTopology619706002ComplexInverseassociationProfile(dataset)...)
	violations = append(violations, shaclgen.ValidateGeneratedTopology619706002SimpleProfile(dataset)...)
	return violations
}
