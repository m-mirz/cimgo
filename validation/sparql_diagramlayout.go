package validation

import (
	"cimgo/cimgostructs"
	"strings"
)

// CheckDiagramObjectIdentifiedObjectType implements dlc.DiagramObject.IdentifiedObject-DLvalueType
// Description: DiagramObject.IdentifiedObject must be an IRI and must NOT point to one of:
// Diagram, DiagramObject, VisibilityLayer, DiagramStyle, DiagramObjectStyle, TextDiagramObject.
func CheckDiagramObjectIdentifiedObjectType(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation

	disallowed := func(o interface{}) bool {
		switch o.(type) {
		case *cimgostructs.Diagram,
			*cimgostructs.DiagramObject,
			*cimgostructs.VisibilityLayer,
			*cimgostructs.DiagramStyle,
			*cimgostructs.DiagramObjectStyle,
			*cimgostructs.TextDiagramObject:
			return true
		}
		return false
	}

	for id, obj := range dataset.Elements {
		do, ok := obj.(*cimgostructs.DiagramObject)
		if !ok || do.IdentifiedObject_ == nil {
			continue
		}
		targetID := strings.TrimPrefix(do.IdentifiedObject_.MRID, "#")
		targetObj, ok := dataset.Elements[targetID]
		if !ok {
			continue
		}
		if disallowed(targetObj) {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    goTypeName(obj),
				Property: "DiagramObject.IdentifiedObject",
				Message:  "The value type shall not be an instance of cim:Diagram, cim:DiagramObject, cim:VisibilityLayer, cim:DiagramStyle, cim:DiagramObjectStyle or cim:TextDiagramObject.",
				Severity: "sh.Violation",
			})
		}
	}

	return violations
}
