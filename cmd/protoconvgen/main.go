// Command protoconvgen generates typed conversion functions from cimstructs to
// proto/api/v1 types. Each CIM class — concrete and abstract alike — gets one
// FooToProto(*cimstructs.Foo) *apiv1.ProtoFoo function. A top-level ToProto
// function dispatches over all concrete element types in a CIMDataset.
//
// Output is written to cimconv/generated_conv.go (override with -out).
// Run via go generate (see gen.go) or directly: go run ./cmd/protoconvgen/
package main

import (
	"bytes"
	"cimgo/cimstructs"
	apiv1 "cimgo/proto/api/v1"
	"flag"
	"fmt"
	"go/format"
	"os"
	"reflect"
	"sort"
	"strings"
)

func main() {
	outFile := flag.String("out", "cimconv/generated_conv.go", "output file")
	flag.Parse()

	// Build cimMap first; proto lookups use it for case-insensitive name matching.
	cimMap := collectAllCIMTypes()
	protoMap := buildProtoTypeMap(cimMap)
	expandProtoTypeMap(protoMap, cimMap)

	names := sortedKeys(cimMap)

	var buf bytes.Buffer
	emitHeader(&buf)
	emitHelpers(&buf)
	emitToProto(&buf, protoMap)

	skipped := 0
	for _, name := range names {
		protoType, ok := protoMap[name]
		if !ok {
			skipped++
			continue
		}
		emitConvFunc(&buf, name, cimMap[name], protoType, protoMap)
	}

	src, err := format.Source(buf.Bytes())
	if err != nil {
		_ = os.WriteFile(*outFile, buf.Bytes(), 0o644)
		fmt.Fprintf(os.Stderr, "format error (raw output written for debugging): %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*outFile, src, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", *outFile, err)
		os.Exit(1)
	}
	fmt.Printf("Generated %s (%d converters, %d without proto type skipped)\n",
		*outFile, len(names)-skipped, skipped)
}

// collectAllCIMTypes walks cimstructs.StructMap and the anonymous embeds of
// every concrete type, returning a map of all CIM struct names to reflect.Type.
// Abstract base classes discovered via embedding are included.
func collectAllCIMTypes() map[string]reflect.Type {
	result := map[string]reflect.Type{}
	var walk func(t reflect.Type)
	walk = func(t reflect.Type) {
		name := t.Name()
		if _, seen := result[name]; seen {
			return
		}
		result[name] = t
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Anonymous && f.Type.Kind() == reflect.Struct &&
				strings.HasSuffix(f.Type.PkgPath(), "/cimstructs") {
				walk(f.Type)
			}
		}
	}
	for _, factory := range cimstructs.StructMap {
		walk(reflect.TypeOf(factory()).Elem())
	}
	return result
}

// buildProtoTypeMap builds a cimstructs-name → proto Go reflect.Type map for
// all concrete CIM classes by enumerating the []*Foo fields of CIMDataset.
// Names are matched case-insensitively against cimMap to handle the few spots
// where the proto generator capitalises differently (e.g. "1or2" → "1Or2").
func buildProtoTypeMap(cimMap map[string]reflect.Type) map[string]reflect.Type {
	cimLower := lowerIndex(cimMap)
	result := map[string]reflect.Type{}
	lt := reflect.TypeOf(apiv1.CIMDataset{})
	for i := 0; i < lt.NumField(); i++ {
		f := lt.Field(i)
		if f.Type.Kind() != reflect.Slice {
			continue
		}
		elem := f.Type.Elem()
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		if elem.Kind() != reflect.Struct {
			continue
		}
		key := cimKeyFor(f.Name, cimLower)
		result[key] = elem
	}
	return result
}

// expandProtoTypeMap discovers abstract base class types not in CIMDataset
// by following Super pointer fields transitively from known proto types.
func expandProtoTypeMap(m map[string]reflect.Type, cimMap map[string]reflect.Type) {
	cimLower := lowerIndex(cimMap)
	for {
		added := 0
		for _, pt := range m {
			sf, ok := pt.FieldByName("Super")
			if !ok {
				continue
			}
			parent := sf.Type
			if parent.Kind() == reflect.Ptr {
				parent = parent.Elem()
			}
			if parent.Kind() != reflect.Struct {
				continue
			}
			key := cimKeyFor(parent.Name(), cimLower)
			if _, exists := m[key]; !exists {
				m[key] = parent
				added++
			}
		}
		if added == 0 {
			break
		}
	}
}

// lowerIndex builds a lowercase → original-case map for fast case-insensitive lookup.
func lowerIndex(m map[string]reflect.Type) map[string]string {
	idx := make(map[string]string, len(m))
	for k := range m {
		idx[strings.ToLower(k)] = k
	}
	return idx
}

// cimKeyFor returns the cimstructs canonical name for protoGoName. If a
// case-insensitive match exists in cimLower it returns the cimstructs casing;
// otherwise it returns protoGoName unchanged.
func cimKeyFor(protoGoName string, cimLower map[string]string) string {
	if cimName, ok := cimLower[strings.ToLower(protoGoName)]; ok {
		return cimName
	}
	return protoGoName
}

func sortedKeys(m map[string]reflect.Type) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// xmlLocalName extracts the unqualified field name from an xml struct tag.
// "ACLineSegment.b0ch,omitempty" → "b0ch"
func xmlLocalName(tag string) string {
	if i := strings.IndexByte(tag, ','); i >= 0 {
		tag = tag[:i]
	}
	if i := strings.IndexByte(tag, '.'); i >= 0 {
		return tag[i+1:]
	}
	return tag
}

// findProtoField returns the Go struct field in protoType whose protobuf tag
// carries name=<localName> (case-insensitive).
func findProtoField(protoType reflect.Type, localName string) (reflect.StructField, bool) {
	lower := strings.ToLower(localName)
	for i := 0; i < protoType.NumField(); i++ {
		f := protoType.Field(i)
		tag := f.Tag.Get("protobuf")
		if tag == "" {
			continue
		}
		for _, part := range strings.Split(tag, ",") {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 && kv[0] == "name" && strings.ToLower(kv[1]) == lower {
				return f, true
			}
		}
	}
	return reflect.StructField{}, false
}

func emitHeader(buf *bytes.Buffer) {
	buf.WriteString("// Code generated by cmd/protoconvgen. DO NOT EDIT.\n\n")
	buf.WriteString("package cimconv\n\n")
	buf.WriteString("import (\n")
	buf.WriteString("\t\"cimgo/cimstructs\"\n")
	buf.WriteString("\tapiv1 \"cimgo/proto/api/v1\"\n")
	buf.WriteString("\t\"strings\"\n")
	buf.WriteString(")\n\n")
}

func emitHelpers(buf *bytes.Buffer) {
	buf.WriteString(`// uriToEnumKey converts a CIM URI like "http://iec.ch/TC57/CIM100#PhaseCode.ABC"
// to the proto enum map key "PhaseCode_ABC".
func uriToEnumKey(uri string) string {
	if i := strings.LastIndexByte(uri, '#'); i >= 0 {
		return strings.ReplaceAll(uri[i+1:], ".", "_")
	}
	return ""
}

`)
}

// emitToProto emits the top-level ToProto function that type-switches over all
// concrete CIM classes and dispatches to their individual converters.
func emitToProto(buf *bytes.Buffer, protoMap map[string]reflect.Type) {
	var concreteNames []string
	for name := range cimstructs.StructMap {
		if _, inProto := protoMap[name]; inProto {
			concreteNames = append(concreteNames, name)
		}
	}
	sort.Strings(concreteNames)

	buf.WriteString("// ToProto converts a CIMDataset to its Protobuf equivalent.\n")
	buf.WriteString("func ToProto(src *cimstructs.CIMDataset) (*apiv1.CIMDataset, error) {\n")
	buf.WriteString("\tdst := &apiv1.CIMDataset{}\n")
	buf.WriteString("\tfor _, elem := range src.ByID {\n")
	buf.WriteString("\t\tswitch v := elem.(type) {\n")
	for _, name := range concreteNames {
		pt := protoMap[name]
		fmt.Fprintf(buf, "\t\tcase *cimstructs.%s:\n", name)
		fmt.Fprintf(buf, "\t\t\tdst.%s = append(dst.%s, %sToProto(v))\n",
			pt.Name(), pt.Name(), name)
	}
	buf.WriteString("\t\t}\n\t}\n\treturn dst, nil\n}\n\n")
}

// emitConvFunc emits one FooToProto conversion function. The function name and
// parameter type use the cimstructs name; the return type uses the proto Go name
// (they differ when proto capitalises differently, e.g. "1or2" vs "1Or2").
func emitConvFunc(buf *bytes.Buffer, name string, cimType, protoType reflect.Type, protoMap map[string]reflect.Type) {
	protoName := protoType.Name()
	fmt.Fprintf(buf, "func %sToProto(src *cimstructs.%s) *apiv1.%s {\n", name, name, protoName)
	buf.WriteString("\tif src == nil {\n\t\treturn nil\n\t}\n")
	fmt.Fprintf(buf, "\tdst := &apiv1.%s{}\n", protoName)

	// CIM parent class: first anonymous embed whose package is cimstructs.
	// Emit dst.Super only when the proto type has a Super field and the parent
	// has its own converter (i.e. appears in protoMap).
	for i := 0; i < cimType.NumField(); i++ {
		f := cimType.Field(i)
		if !f.Anonymous || f.Type.Kind() != reflect.Struct {
			continue
		}
		if !strings.HasSuffix(f.Type.PkgPath(), "/cimstructs") {
			continue
		}
		parentCIMName := f.Type.Name()
		if _, hasProto := protoMap[parentCIMName]; !hasProto {
			break
		}
		if sf, ok := protoType.FieldByName("Super"); ok && sf.Type.Kind() == reflect.Ptr {
			fmt.Fprintf(buf, "\tdst.Super = %sToProto(&src.%s)\n", parentCIMName, parentCIMName)
		}
		break
	}

	// Regular fields: match by xml tag local name against proto field name.
	for i := 0; i < cimType.NumField(); i++ {
		f := cimType.Field(i)
		if f.Anonymous {
			continue
		}
		local := xmlLocalName(f.Tag.Get("xml"))
		if local == "" {
			continue
		}
		pf, ok := findProtoField(protoType, local)
		if !ok {
			continue
		}
		emitFieldConv(buf, f, pf)
	}

	buf.WriteString("\treturn dst\n}\n\n")
}

func emitFieldConv(buf *bytes.Buffer, src, dst reflect.StructField) {
	ft := src.Type

	// Pointer to single-field anonymous struct: reference or enum.
	if ft.Kind() == reflect.Ptr && ft.Elem().Kind() == reflect.Struct && ft.Elem().NumField() == 1 {
		switch ft.Elem().Field(0).Name {
		case "MRID":
			fmt.Fprintf(buf, "\tif src.%s != nil {\n", src.Name)
			fmt.Fprintf(buf, "\t\tdst.%s = strings.TrimPrefix(src.%s.MRID, \"#\")\n", dst.Name, src.Name)
			fmt.Fprintf(buf, "\t}\n")
		case "URI":
			if dst.Type.Kind() != reflect.Int32 {
				return
			}
			enum := dst.Type.Name()
			fmt.Fprintf(buf, "\tif src.%s != nil {\n", src.Name)
			fmt.Fprintf(buf, "\t\tdst.%s = apiv1.%s(apiv1.%s_value[uriToEnumKey(src.%s.URI)])\n",
				dst.Name, enum, enum, src.Name)
			fmt.Fprintf(buf, "\t}\n")
		}
		return
	}

	// Slice fields.
	if ft.Kind() == reflect.Slice {
		elem := ft.Elem()
		// []*struct{MRID string} — pointer-element reference slice
		if elem.Kind() == reflect.Ptr && elem.Elem().Kind() == reflect.Struct &&
			elem.Elem().NumField() == 1 && elem.Elem().Field(0).Name == "MRID" {
			fmt.Fprintf(buf, "\tfor _, ref := range src.%s {\n", src.Name)
			fmt.Fprintf(buf, "\t\tif ref != nil {\n")
			fmt.Fprintf(buf, "\t\t\tdst.%s = append(dst.%s, strings.TrimPrefix(ref.MRID, \"#\"))\n", dst.Name, dst.Name)
			fmt.Fprintf(buf, "\t\t}\n")
			fmt.Fprintf(buf, "\t}\n")
			return
		}
		// []struct{MRID string} — value-element reference slice
		if elem.Kind() == reflect.Struct && elem.NumField() == 1 && elem.Field(0).Name == "MRID" {
			fmt.Fprintf(buf, "\tfor _, ref := range src.%s {\n", src.Name)
			fmt.Fprintf(buf, "\t\tdst.%s = append(dst.%s, strings.TrimPrefix(ref.MRID, \"#\"))\n", dst.Name, dst.Name)
			fmt.Fprintf(buf, "\t}\n")
			return
		}
		// []string or any other identical slice — direct copy
		if ft == dst.Type {
			fmt.Fprintf(buf, "\tdst.%s = src.%s\n", dst.Name, src.Name)
		}
		return
	}

	// Identical types: direct assignment.
	if ft == dst.Type {
		fmt.Fprintf(buf, "\tdst.%s = src.%s\n", dst.Name, src.Name)
		return
	}

	// Numeric type mismatch (e.g. int → int64, float32 → float64).
	if isNumericKind(ft.Kind()) && isNumericKind(dst.Type.Kind()) {
		fmt.Fprintf(buf, "\tdst.%s = %s(src.%s)\n", dst.Name, dst.Type.String(), src.Name)
	}
}

func isNumericKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}
