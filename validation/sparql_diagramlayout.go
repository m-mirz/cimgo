package validation

import (
	"cimgo/cimstructs"
	"strings"
)

// ValidateDLProfileSPARQL runs hand-written checks for 61970-301_DiagramLayout-AP-Con-Complex-SHACL.
func ValidateDLProfileSPARQL(dataset *cimstructs.CIMElementList) []Violation {
	return CheckDiagramObjectIdentifiedObjectType(dataset)
}

// CheckDiagramObjectIdentifiedObjectType implements dlc.DiagramObject.IdentifiedObject-DLvalueType
// Profile: 61970-301_DiagramLayout-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: DiagramObject.IdentifiedObject must be an IRI and must NOT point to one of:
// Diagram, DiagramObject, VisibilityLayer, DiagramStyle, DiagramObjectStyle, TextDiagramObject.
func CheckDiagramObjectIdentifiedObjectType(dataset *cimstructs.CIMElementList) []Violation {
	var violations []Violation

	disallowed := func(o interface{}) bool {
		switch o.(type) {
		case *cimstructs.Diagram,
			*cimstructs.DiagramObject,
			*cimstructs.VisibilityLayer,
			*cimstructs.DiagramStyle,
			*cimstructs.DiagramObjectStyle,
			*cimstructs.TextDiagramObject:
			return true
		}
		return false
	}

	for id, obj := range dataset.Elements {
		var identifiedObject *struct {
			MRID string `xml:"resource,attr"`
		}

		switch v := obj.(type) {
		case *cimstructs.DiagramObject:
			identifiedObject = v.IdentifiedObject_
		case *cimstructs.TextDiagramObject:
			identifiedObject = v.IdentifiedObject_
		default:
			continue
		}

		if identifiedObject == nil {
			continue
		}
		targetID := strings.TrimPrefix(identifiedObject.MRID, "#")
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
				Severity: "sh:Violation",
			})
		}
	}

	return violations
}
