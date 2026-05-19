package main

import (
	"fmt"
	"reflect"
	"strings"
)

// inverseCountCheck emits the Prelude that builds an O(N) inverse-reference
// count map plus the loop-body condition that compares the count against
// `threshold` using `op` (the operator that *violates* the constraint —
// "==" for "must have ≥1", "<" for MinCount, ">" for MaxCount). When
// targetClasses has more than one entry the prelude dispatches over each
// concrete subclass — this is how abstract base classes (e.g.
// ExcitationSystemDynamics) are handled, since they're not in StructMap
// and have no instances of their own.
func inverseCountCheck(targetClasses []string, field reflect.StructField, op, threshold string) (string, string) {
	cond := fmt.Sprintf("inverseCounts[id] %s %s", op, threshold)
	isList := field.Type.Kind() == reflect.Slice
	if len(targetClasses) == 1 {
		var prelude string
		if isList {
			prelude = fmt.Sprintf(`	inverseCounts := map[string]int{}
	for _, ref := range dataset.ByID {
		r, ok := ref.(*cimstructs.%s)
		if !ok {
			continue
		}
		for _, entry := range r.%s {
			inverseCounts[strings.TrimPrefix(entry.MRID, "#")]++
		}
	}`, targetClasses[0], field.Name)
		} else {
			prelude = fmt.Sprintf(`	inverseCounts := map[string]int{}
	for _, ref := range dataset.ByID {
		r, ok := ref.(*cimstructs.%s)
		if !ok {
			continue
		}
		if r.%s == nil {
			continue
		}
		inverseCounts[strings.TrimPrefix(r.%s.MRID, "#")]++
	}`, targetClasses[0], field.Name, field.Name)
		}
		return prelude, cond
	}
	var b strings.Builder
	b.WriteString("\tinverseCounts := map[string]int{}\n")
	b.WriteString("\tfor _, ref := range dataset.ByID {\n")
	b.WriteString("\t\tswitch r := ref.(type) {\n")
	for _, cls := range targetClasses {
		fmt.Fprintf(&b, "\t\tcase *cimstructs.%s:\n", cls)
		if isList {
			fmt.Fprintf(&b, "\t\t\tfor _, entry := range r.%s {\n", field.Name)
			b.WriteString("\t\t\t\tinverseCounts[strings.TrimPrefix(entry.MRID, \"#\")]++\n")
			b.WriteString("\t\t\t}\n")
		} else {
			fmt.Fprintf(&b, "\t\t\tif r.%s != nil {\n", field.Name)
			fmt.Fprintf(&b, "\t\t\t\tinverseCounts[strings.TrimPrefix(r.%s.MRID, \"#\")]++\n", field.Name)
			b.WriteString("\t\t\t}\n")
		}
	}
	b.WriteString("\t\t}\n")
	b.WriteString("\t}")
	return b.String(), cond
}

// inverseHasEnumValueCheck emits the Prelude for a 2-segment inverse-then-
// forward HasValue check: scan the dataset once, flag every focus-node id
// that has at least one referrer (a *cimstructs.<targetClasses[i]>) whose
// `refField` points back AND whose `valueField` (a *struct{URI string} enum
// field) carries the named enum constant. Violation is "no such referrer
// found". Multi-class targets dispatch over each concrete subclass.
func inverseHasEnumValueCheck(targetClasses []string, refField, valueField reflect.StructField, constIdent string) (string, string) {
	cond := "!hasEnumValue[id]"
	if len(targetClasses) == 1 {
		prelude := fmt.Sprintf(`	hasEnumValue := map[string]bool{}
	for _, ref := range dataset.ByID {
		r, ok := ref.(*cimstructs.%s)
		if !ok {
			continue
		}
		if r.%s == nil || r.%s == nil {
			continue
		}
		if r.%s.URI != cimstructs.%s {
			continue
		}
		hasEnumValue[strings.TrimPrefix(r.%s.MRID, "#")] = true
	}`, targetClasses[0], refField.Name, valueField.Name, valueField.Name, constIdent, refField.Name)
		return prelude, cond
	}
	var b strings.Builder
	b.WriteString("\thasEnumValue := map[string]bool{}\n")
	b.WriteString("\tfor _, ref := range dataset.ByID {\n")
	b.WriteString("\t\tswitch r := ref.(type) {\n")
	for _, cls := range targetClasses {
		fmt.Fprintf(&b, "\t\tcase *cimstructs.%s:\n", cls)
		fmt.Fprintf(&b, "\t\t\tif r.%s == nil || r.%s == nil {\n", refField.Name, valueField.Name)
		b.WriteString("\t\t\t\tcontinue\n")
		b.WriteString("\t\t\t}\n")
		fmt.Fprintf(&b, "\t\t\tif r.%s.URI != cimstructs.%s {\n", valueField.Name, constIdent)
		b.WriteString("\t\t\t\tcontinue\n")
		b.WriteString("\t\t\t}\n")
		fmt.Fprintf(&b, "\t\t\thasEnumValue[strings.TrimPrefix(r.%s.MRID, \"#\")] = true\n", refField.Name)
	}
	b.WriteString("\t\t}\n")
	b.WriteString("\t}")
	return b.String(), cond
}
