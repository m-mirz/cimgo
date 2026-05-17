package cgmesxml

import (
	"cimgo/cimbase"
	"cimgo/cimstructs"
	"cimgo/cimxml"
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// reverseNamespaces maps namespace URI → xmlns prefix (e.g. "cim", "eu", "md").
var reverseNamespaces = func() map[string]string {
	m := make(map[string]string, len(cimstructs.CIMNamespaces))
	for prefix, uri := range cimstructs.CIMNamespaces {
		m[uri] = prefix
	}
	return m
}()

// profileURLKeywords maps profile code → substring found in the Model.profile URL.
var profileURLKeywords = map[string]string{
	"EQ":   "CoreEquipment",
	"EQBD": "EquipmentBoundary",
	"SSH":  "SteadyStateHypothesis",
	"SV":   "StateVariables",
	"TP":   "Topology",
	"DL":   "DiagramLayout",
	"DY":   "Dynamics",
	"GL":   "GeographicalLocation",
	"OP":   "Operation",
	"SC":   "ShortCircuit",
	"FH":   "FileHeader",
}

func nsPrefix(namespace string) string {
	if p, ok := reverseNamespaces[namespace]; ok {
		return p
	}
	return "cim"
}

func containsOrigin(origins []string, code string) bool {
	for _, o := range origins {
		if o == code {
			return true
		}
	}
	return false
}

// rdfAttr builds either rdf:ID or rdf:about depending on whether this profile
// is the primary owner of the element.
func rdfAttr(id, profileCode string, typeInfo cimbase.CIMTypeInfo) xml.Attr {
	if typeInfo.Origin == profileCode {
		return xml.Attr{Name: xml.Name{Local: "rdf:ID"}, Value: id}
	}
	value := id
	if strings.HasPrefix(id, "_") {
		value = "#" + id
	}
	return xml.Attr{Name: xml.Name{Local: "rdf:about"}, Value: value}
}

// profileAttrBelongsToType reports whether attrInfo's attribute should be emitted
// for profileCode given the concrete element's type origins.
// An attribute A belongs to profile P for type T iff P ∈ A.Origins AND P ∈ typeOrigins.
func profileAttrBelongsToType(attrInfo cimbase.CIMAttributeInfo, profileCode string, typeOrigins []string) bool {
	if !containsOrigin(attrInfo.Origins, profileCode) {
		return false
	}
	return containsOrigin(typeOrigins, profileCode)
}

// hasStrictProfileField reports whether any non-zero field in v has an attribute
// that belongs exclusively to profileCode (len(Origins)==1 && Origins[0]==profileCode).
// Used for secondary-profile elements: prevents false positives from universal
// attrs like IdentifiedObject.name that appear in many profiles.
func hasStrictProfileField(v reflect.Value, profileCode string) bool {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fv := v.Field(i)
		if field.Anonymous {
			if field.Type == reflect.TypeOf(cimbase.Base{}) {
				continue
			}
			if hasStrictProfileField(fv, profileCode) {
				return true
			}
			continue
		}
		tag := field.Tag.Get("xml")
		if tag == "" || tag == "-" {
			continue
		}
		tagName := strings.Split(tag, ",")[0]
		if tagName == "ID" {
			continue
		}
		attrInfo, ok := cimstructs.AttributeInfoMap[tagName]
		if !ok || len(attrInfo.Origins) != 1 || attrInfo.Origins[0] != profileCode {
			continue
		}
		if !fv.IsZero() {
			return true
		}
	}
	return false
}

// hasProfileFields reports whether any non-zero field in v (walking embedded
// structs) would be emitted for profileCode given the concrete element's typeOrigins.
func hasProfileFields(v reflect.Value, profileCode string, typeOrigins []string) bool {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fv := v.Field(i)
		if field.Anonymous {
			if field.Type == reflect.TypeOf(cimbase.Base{}) {
				continue
			}
			if hasProfileFields(fv, profileCode, typeOrigins) {
				return true
			}
			continue
		}
		tag := field.Tag.Get("xml")
		if tag == "" || tag == "-" {
			continue
		}
		tagName := strings.Split(tag, ",")[0]
		if tagName == "ID" {
			continue
		}
		attrInfo, ok := cimstructs.AttributeInfoMap[tagName]
		if !ok || !profileAttrBelongsToType(attrInfo, profileCode, typeOrigins) {
			continue
		}
		if !fv.IsZero() {
			return true
		}
	}
	return false
}

// encodeFields walks v (a struct) recursively, emitting XML elements for every
// field whose AttributeInfoMap entry belongs to profileCode for typeOrigins.
func encodeFields(enc *cimxml.Encoder, v reflect.Value, profileCode string, typeOrigins []string) error {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fv := v.Field(i)

		// Recurse into anonymous (embedded) structs, but skip cimbase.Base
		// which only holds the rdf:ID attribute (already emitted).
		if field.Anonymous {
			if field.Type == reflect.TypeOf(cimbase.Base{}) {
				continue
			}
			if err := encodeFields(enc, fv, profileCode, typeOrigins); err != nil {
				return err
			}
			continue
		}

		tag := field.Tag.Get("xml")
		if tag == "" || tag == "-" {
			continue
		}
		tagName := strings.Split(tag, ",")[0]
		if tagName == "ID" { // rdf:ID attr — handled at element level
			continue
		}

		attrInfo, ok := cimstructs.AttributeInfoMap[tagName]
		if !ok || !profileAttrBelongsToType(attrInfo, profileCode, typeOrigins) {
			continue
		}

		if err := encodeField(enc, fv, tagName, attrInfo, profileCode); err != nil {
			return err
		}
	}
	return nil
}

func encodeField(enc *cimxml.Encoder, fv reflect.Value, tagName string, attrInfo cimbase.CIMAttributeInfo, profileCode string) error {
	prefix := nsPrefix(attrInfo.Namespace)
	localName := prefix + ":" + tagName

	// Dereference pointer
	if fv.Kind() == reflect.Ptr {
		if fv.IsNil() {
			return nil
		}
		fv = fv.Elem()
	}

	// Zero value — omit
	if fv.IsZero() {
		return nil
	}

	switch fv.Kind() {
	case reflect.Struct:
		// *struct{ MRID string } or *struct{ URI string } → self-closing with rdf:resource
		mridField := fv.FieldByName("MRID")
		uriField := fv.FieldByName("URI")
		var resourceVal string
		if mridField.IsValid() && mridField.String() != "" {
			mrid := mridField.String()
			if strings.HasPrefix(mrid, "_") {
				resourceVal = "#" + mrid
			} else {
				resourceVal = mrid
			}
		} else if uriField.IsValid() && uriField.String() != "" {
			resourceVal = uriField.String()
		}
		if resourceVal != "" {
			return emitResource(enc, localName, resourceVal)
		}

	case reflect.Slice:
		if fv.Type().Elem().Kind() == reflect.Struct {
			// []struct{ MRID string } or []struct{ URI string }
			for j := 0; j < fv.Len(); j++ {
				entry := fv.Index(j)
				mridField := entry.FieldByName("MRID")
				uriField := entry.FieldByName("URI")
				var resourceVal string
				if mridField.IsValid() && mridField.String() != "" {
					mrid := mridField.String()
					if strings.HasPrefix(mrid, "_") {
						resourceVal = "#" + mrid
					} else {
						resourceVal = mrid
					}
				} else if uriField.IsValid() && uriField.String() != "" {
					resourceVal = uriField.String()
				}
				if resourceVal != "" {
					if err := emitResource(enc, localName, resourceVal); err != nil {
						return err
					}
				}
			}
			return nil
		}
		// []string — one element per entry (e.g. Model.profile)
		for j := 0; j < fv.Len(); j++ {
			if err := emitCharData(enc, localName, fv.Index(j).String()); err != nil {
				return err
			}
		}
		return nil

	default:
		return emitCharData(enc, localName, fmt.Sprintf("%v", fv.Interface()))
	}
	return nil
}

func emitResource(enc *cimxml.Encoder, localName, resourceVal string) error {
	return enc.EncodeToken(cimxml.StartElementSelfClosing{
		Name: xml.Name{Local: localName},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "rdf:resource"}, Value: resourceVal},
		},
	})
}

func emitCharData(enc *cimxml.Encoder, localName, value string) error {
	start := xml.StartElement{Name: xml.Name{Local: localName}}
	if err := enc.EncodeToken(start); err != nil {
		return err
	}
	if err := enc.EncodeToken(xml.CharData(value)); err != nil {
		return err
	}
	return enc.EncodeToken(start.End())
}

// EncodeForProfile encodes only the elements and attributes that belong to
// profileCode into w, using rdf:ID for primary-profile elements and rdf:about
// for secondary-profile references, with correct namespace prefixes.
func EncodeForProfile(w io.Writer, cimData *cimstructs.CIMElementList, profileCode string) error {
	if _, err := w.Write([]byte("<?xml version=\"1.0\" encoding=\"utf-8\" ?>\n")); err != nil {
		return err
	}

	enc := cimxml.NewEncoder(w)
	enc.Indent("", "  ")

	root := xml.StartElement{
		Name: xml.Name{Local: "rdf:RDF"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "xmlns:rdf"}, Value: cimstructs.CIMNamespaces["rdf"]},
			{Name: xml.Name{Local: "xmlns:cim"}, Value: cimstructs.CIMNamespaces["cim"]},
			{Name: xml.Name{Local: "xmlns:eu"}, Value: cimstructs.CIMNamespaces["eu"]},
			{Name: xml.Name{Local: "xmlns:md"}, Value: cimstructs.CIMNamespaces["md"]},
		},
	}
	if err := enc.EncodeToken(root); err != nil {
		return err
	}

	urlKeyword := profileURLKeywords[profileCode]

	for _, element := range cimData.Elements {
		rv := reflect.TypeOf(element)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		typeName := rv.Name()

		typeInfo, ok := cimstructs.TypeInfoMap[typeName]
		if !ok {
			continue
		}

		id := element.(cimbase.CIMElement).GetId()
		elemPrefix := nsPrefix(typeInfo.Namespace)
		startLocal := elemPrefix + ":" + typeName

		// FullModel: include only the one matching this profile, always rdf:about.
		if typeName == "FullModel" {
			fm, ok := element.(interface{ GetProfiles() []string })
			_ = fm
			_ = ok
			// Access Profile field via reflection.
			ev := reflect.ValueOf(element).Elem()
			profileField := ev.FieldByName("Profile")
			if !profileField.IsValid() {
				continue
			}
			matched := false
			for i := 0; i < profileField.Len(); i++ {
				if strings.Contains(profileField.Index(i).String(), urlKeyword) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
			value := id
			if strings.HasPrefix(id, "_") {
				value = "#" + id
			}
			start := xml.StartElement{
				Name: xml.Name{Local: startLocal},
				Attr: []xml.Attr{{Name: xml.Name{Local: "rdf:about"}, Value: value}},
			}
			if err := enc.EncodeToken(start); err != nil {
				return err
			}
			if err := encodeFields(enc, reflect.ValueOf(element), "FH", typeInfo.Origins); err != nil {
				return err
			}
			if err := enc.EncodeToken(start.End()); err != nil {
				return err
			}
			continue
		}

		if typeInfo.Origin == profileCode {
			// Primary: use intersection check (attr in profile AND type in profile).
			if !hasProfileFields(reflect.ValueOf(element), profileCode, typeInfo.Origins) {
				continue
			}
		} else {
			// Secondary (rdf:about): require at least one attribute whose primary origin
			// IS profileCode — prevents spurious inclusion via universal attrs like mRID.
			if !hasStrictProfileField(reflect.ValueOf(element), profileCode) {
				continue
			}
		}

		start := xml.StartElement{
			Name: xml.Name{Local: startLocal},
			Attr: []xml.Attr{rdfAttr(id, profileCode, typeInfo)},
		}
		if err := enc.EncodeToken(start); err != nil {
			return err
		}
		if err := encodeFields(enc, reflect.ValueOf(element), profileCode, typeInfo.Origins); err != nil {
			return err
		}
		if err := enc.EncodeToken(start.End()); err != nil {
			return err
		}
	}

	if err := enc.EncodeToken(root.End()); err != nil {
		return err
	}
	if err := enc.Flush(); err != nil {
		return err
	}
	if _, err := w.Write([]byte("\n")); err != nil {
		return err
	}
	return nil
}
