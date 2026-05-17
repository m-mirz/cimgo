package main

import (
	"cimgo/cimstructs"
	"reflect"
	"sort"
	"strings"
)

// concreteSubclassesEmbedding returns the sorted list of concrete cimstructs
// class names (those present in StructMap) whose Go type embeds — directly or
// transitively — an anonymous struct named `abstract`. This is how the
// generator handles SHACL inverse paths that target abstract base classes
// like ExcitationSystemDynamics: those classes have no instances of their
// own, so the generated check has to dispatch over every concrete subclass.
func concreteSubclassesEmbedding(abstract string) []string {
	var matches []string
	for name, factory := range cimstructs.StructMap {
		t := reflect.TypeOf(factory()).Elem()
		if typeEmbedsName(t, abstract) {
			matches = append(matches, name)
		}
	}
	sort.Strings(matches)
	return matches
}

// classNameFromPayload normalises a SHACL `sh:class` payload (an IRI in
// either `<cim:X>` or `cim:X` form) to a bare Go struct name. Returns false
// if the payload isn't a string in the cim namespace.
func classNameFromPayload(payload any) (string, bool) {
	s, ok := payload.(string)
	if !ok {
		return "", false
	}
	s = strings.TrimPrefix(strings.TrimSuffix(s, ">"), "<")
	return stripCIMPrefix(s)
}

// isClassOrAncestor reports whether the class named `class` satisfies a SHACL
// `sh:class asserted` requirement structurally — i.e. either class == asserted
// or class embeds asserted (Go's encoding of subclass-of, transitively). Used
// to short-circuit inverse-path Class constraints when every concrete target
// type is already a subtype of the asserted class.
func isClassOrAncestor(class, asserted string) bool {
	if class == asserted {
		return true
	}
	factory, ok := cimstructs.StructMap[class]
	if !ok {
		return false
	}
	return typeEmbedsName(reflect.TypeOf(factory()).Elem(), asserted)
}

func typeEmbedsName(t reflect.Type, name string) bool {
	if t.Kind() != reflect.Struct {
		return false
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.Anonymous {
			continue
		}
		if f.Type.Name() == name {
			return true
		}
		if f.Type.Kind() == reflect.Struct && typeEmbedsName(f.Type, name) {
			return true
		}
	}
	return false
}

func zeroLiteralFor(k reflect.Kind) string {
	switch k {
	case reflect.String:
		return `""`
	case reflect.Bool:
		return "false"
	default:
		return "0"
	}
}

// xmlTagExistsOnAnyStruct returns true if any concrete struct in StructMap has
// a field whose XML tag local name matches tag. Used by pairCompare to
// distinguish genuine TTL typos (tag missing from the schema entirely) from
// cross-class comparisons (tag exists on a sibling type, making the
// constraint vacuously satisfied for the current target).
func xmlTagExistsOnAnyStruct(tag string) bool {
	for _, factory := range cimstructs.StructMap {
		t := reflect.TypeOf(factory()).Elem()
		if _, ok := findFieldByXMLTag(t, tag); ok {
			return true
		}
	}
	return false
}

func findFieldByXMLTag(t reflect.Type, tag string) (reflect.StructField, bool) {
	if t.Kind() != reflect.Struct {
		return reflect.StructField{}, false
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if got := xmlLocal(f.Tag.Get("xml")); got != "" && got == tag {
			return f, true
		}
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			if sub, ok := findFieldByXMLTag(f.Type, tag); ok {
				return sub, true
			}
		}
	}
	return reflect.StructField{}, false
}
