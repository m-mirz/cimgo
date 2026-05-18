package main

import (
	"fmt"
	"strings"
	"unicode"
)

func anyToFloatLiteral(v any) string {
	switch x := v.(type) {
	case float64:
		return fmt.Sprintf("%v", x)
	case string:
		return strings.Trim(x, "\"")
	default:
		return fmt.Sprintf("%v", x)
	}
}

func simpleClassName(s string) (string, bool) {
	name, ok := stripCIMPrefix(s)
	if !ok {
		return "", false
	}
	if strings.ContainsAny(name, "./ ") {
		return "", false
	}
	return name, true
}

// stripCIMPrefix removes a leading namespace prefix from a simplified SHACL
// identifier and returns the local name. Handles:
//   - cim: / cim<version>. (e.g. cim100.) — the main CIM namespace variants
//   - mdc: — ModelDescription (md:) classes such as FullModel
//   - diff: — DifferenceModel (dm:) classes such as DifferenceModel
func stripCIMPrefix(s string) (string, bool) {
	if strings.HasPrefix(s, "cim") {
		rest := s[len("cim"):]
		for len(rest) > 0 && rest[0] >= '0' && rest[0] <= '9' {
			rest = rest[1:]
		}
		if len(rest) == 0 || (rest[0] != '.' && rest[0] != ':') {
			return "", false
		}
		return rest[1:], true
	}
	for _, prefix := range []string{"mdc:", "diff:"} {
		if strings.HasPrefix(s, prefix) {
			return s[len(prefix):], true
		}
	}
	return "", false
}

// cimNamespaceFromPrefix maps a simplified IRI prefix to the full namespace URI.
func cimNamespaceFromPrefix(simplified string) string {
	switch {
	case strings.HasPrefix(simplified, "cim100.") || strings.HasPrefix(simplified, "cim100:"):
		return "http://iec.ch/TC57/CIM100-European#"
	case strings.HasPrefix(simplified, "mdc:"):
		return "http://iec.ch/TC57/61970-552/ModelDescription/1#"
	case strings.HasPrefix(simplified, "diff:"):
		return "http://iec.ch/TC57/61970-552/DifferenceModel/1#"
	default:
		return "http://iec.ch/TC57/CIM100#"
	}
}

func xmlLocal(tag string) string {
	if tag == "" {
		return ""
	}
	if i := strings.IndexByte(tag, ','); i >= 0 {
		tag = tag[:i]
	}
	return tag
}

func camelize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func camelCaseFromStem(stem string) string {
	parts := strings.Split(stem, "_")
	out := strings.Builder{}
	for _, p := range parts {
		if p == "" {
			continue
		}
		out.WriteString(camelize(p))
	}
	return out.String()
}

// profileStem reduces a SHACL file name to a stable lowercase identifier
// suitable for both filenames and Go identifiers. Format:
//
//	<profile>_<version>[_simple|_complex][_<variant>]
//
// Examples:
//
//	61970-301_DiagramLayout-AP-Con-Complex-SHACL                       -> diagramlayout_61970_301_complex
//	61970-301_DiagramLayout-AP-Con-Complex-NotSolvedMAS-SHACL          -> diagramlayout_61970_301_complex_notsolvedmas
//	61970-552-Header-AP-Con-Simple-SHACL                               -> header_61970_552_simple
//	61970-600-2_IdentifiedObjectCommon_AP-Con-Complex-SHACL            -> identifiedobjectcommon_61970_600_2_complex
//	61970-456_StateVariables-AP-Con-Complex-Explicit-CrossProfile-SHACL -> statevariables_61970_456_complex_explicit_crossprofile
//
// All four parts are needed for uniqueness — same profile name appears across
// CIM revisions (61970-301 vs 61970-600-2 etc.) and across Simple/Complex
// shape vocabularies.
func profileStem(fileName string) string {
	name := strings.TrimSuffix(fileName, "-SHACL")

	variants := []struct{ key, suffix string }{
		{"-NotSolvedMAS", "_notsolvedmas"},
		{"-SolvedMAS", "_solvedmas"},
		{"-Explicit-CrossProfile", "_explicit_crossprofile"},
		{"-Implicit-CrossProfile", "_implicit_crossprofile"},
		{"-CrossProfile", "_crossprofile"},
		{"-InverseAssociation", "_inverseassociation"},
	}
	variantSuffix := ""
	for _, v := range variants {
		if strings.Contains(name, v.key) {
			name = strings.ReplaceAll(name, v.key, "")
			variantSuffix = v.suffix
			break
		}
	}

	mode := ""
	for _, m := range []struct{ key, sfx string }{
		{"-AP-Con-Simple", "_simple"},
		{"-AP-Con-Complex", "_complex"},
		{"_AP-Con-Simple", "_simple"},
		{"_AP-Con-Complex", "_complex"},
	} {
		if i := strings.Index(name, m.key); i >= 0 {
			mode = m.sfx
			name = name[:i]
			break
		}
	}
	for _, sep := range []string{"-AP-", "_AP-"} {
		if i := strings.Index(name, sep); i >= 0 {
			name = name[:i]
			break
		}
	}

	version := ""
	profile := name
	if i := strings.LastIndexAny(name, "_-"); i >= 0 {
		version = name[:i]
		profile = name[i+1:]
	}
	versionPart := ""
	if version != "" {
		versionPart = "_" + sanitizeIdent(version)
	}
	return strings.ToLower(profile) + versionPart + mode + variantSuffix
}

// sanitizeIdent lowercases and replaces non-identifier separators with "_" so
// the result is safe to embed in a Go identifier.
func sanitizeIdent(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}
