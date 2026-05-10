package validation

import (
	"cimgo/cimgostructs"
	"math"
	"reflect"
	"regexp"
	"strings"
)

// ValidateCommonRules runs hand-written checks for common rules (all600, io).
func ValidateCommonRules(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	// Profile: 61970-600-2_IdentifiedObjectCommon_AP-Con-Complex
	violations = append(violations, CheckIdentifiedObjectStringLengths(dataset)...)
	// Profile: 61970-600-1_AllProfiles-AP-Con-Complex
	violations = append(violations, CheckFloatSpecialValues(dataset)...)
	violations = append(violations, CheckModelDateTimeUTC(dataset)...)
	violations = append(violations, CheckMRIDUniqueness(dataset)...)
	violations = append(violations, CheckIDUUID(dataset)...)
	violations = append(violations, CheckIDDeprecated(dataset)...)
	violations = append(violations, CheckModelingAuthoritySetNotEmpty(dataset)...)
	// Complex SHACL rules that don't fit into a single attribute constraint
	violations = append(violations, CheckFileHeaderExists(dataset)...)
	return violations
}

// CheckMRIDUniqueness implements all600:All-GENC1
// Profile: 61970-600-1_AllProfiles-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: All IdentifiedObject-s shall have a persistent and globally unique identifier (Master Resource Identifier - mRID).
func CheckMRIDUniqueness(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	seenMRIDs := make(map[string]string) // mRID -> first object ID found with it

	for id, obj := range dataset.Elements {
		io, ok := getIdentifiedObject(obj)
		if !ok || io.MRID == "" {
			continue
		}
		if firstID, seen := seenMRIDs[io.MRID]; seen && firstID != id {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    goTypeName(obj),
				Property: "IdentifiedObject.mRID",
				Message:  "Not a unique identifier.",
				Severity: "sh.Violation",
			})
		}
		seenMRIDs[io.MRID] = id
	}
	return violations
}

// getIdentifiedObject extracts the IdentifiedObject from a CIM struct.
func getIdentifiedObject(obj interface{}) (*cimgostructs.IdentifiedObject, bool) {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, false
	}

	// Try to find IdentifiedObject field by name
	field := val.FieldByName("IdentifiedObject")
	if field.IsValid() {
		if io, ok := field.Interface().(cimgostructs.IdentifiedObject); ok {
			return &io, true
		}
	}

	// Fallback for deep embedding: iterate over fields
	for i := 0; i < val.NumField(); i++ {
		f := val.Field(i)
		if f.Kind() == reflect.Struct {
			if io, ok := f.Interface().(cimgostructs.IdentifiedObject); ok {
				return &io, true
			}
			// One level deeper for classes that inherit from subclasses of IO
			if f.Type().Kind() == reflect.Struct {
				for j := 0; j < f.NumField(); j++ {
					sf := f.Field(j)
					if sf.Kind() == reflect.Struct {
						if io, ok := sf.Interface().(cimgostructs.IdentifiedObject); ok {
							return &io, true
						}
					}
				}
			}
		}
	}

	return nil, false
}

var uuidRegex = regexp.MustCompile("(?i)^([0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{12})$")
var urnUuidRegex = regexp.MustCompile("(?i)^urn:uuid:[0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{12}$")

// CheckIDUUID implements all600:All-GENC4
// Profile: 61970-600-1_AllProfiles-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: IEC 61970-301 strongly recommends to use UUID, as specified in RFC 4122, for the .mRID. CGMES requires the usage of UUID.
func CheckIDUUID(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	for id, obj := range dataset.Elements {
		rawID := id
		cleanID := rawID
		if strings.Contains(rawID, "#_") {
			cleanID = strings.Split(rawID, "#_")[1]
		} else if strings.HasPrefix(rawID, "urn:uuid:") {
			// already urn:uuid format
		} else if strings.Contains(rawID, "#") {
			cleanID = strings.Split(rawID, "#")[1]
			if strings.HasPrefix(cleanID, "_") {
				cleanID = cleanID[1:]
			}
		} else if strings.HasPrefix(rawID, "_") {
			cleanID = rawID[1:]
		}

		if !uuidRegex.MatchString(cleanID) && !urnUuidRegex.MatchString(rawID) {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    goTypeName(obj),
				Property: "rdf:ID",
				Message:  "Invalid syntax of ID (rdf:ID or rdf:about). UUID expected.",
				Severity: "sh.Info",
			})
		}
	}
	return violations
}

// CheckIDDeprecated implements all600:All-GENC5
// Profile: 61970-600-1_AllProfiles-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: (deprecated) Transition rule for ID length and underscore prefix.
func CheckIDDeprecated(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	for id, obj := range dataset.Elements {
		if strings.HasPrefix(id, "urn:uuid:") {
			continue
		}

		var secondPart string
		if strings.Contains(id, "#_") {
			secondPart = strings.Split(id, "#_")[1]
		} else if strings.HasPrefix(id, "_") {
			secondPart = id[1:]
		}

		if len(secondPart) > 59 || secondPart == "" {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    goTypeName(obj),
				Property: "rdf:ID",
				Message:  "The ID string is more than 60 characters or the string does not begin with underscore.",
				Severity: "sh.Violation",
			})
		}
	}
	return violations
}

// CheckModelDateTimeUTC implements all600:Model.created-HGEN4 and Model.scenarioTime-HGEN4
// Profile: 61970-600-1_AllProfiles-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: European exchanges shall refer to UTC (marked with Z suffix).
func CheckModelDateTimeUTC(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	for id, obj := range dataset.Elements {
		val := reflect.ValueOf(obj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		if val.Kind() != reflect.Struct {
			continue
		}

		createdField := val.FieldByName("Created")
		if createdField.IsValid() && createdField.Kind() == reflect.String {
			v := createdField.String()
			if v != "" && !strings.HasSuffix(v, "Z") {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    goTypeName(obj),
					Property: "Model.created",
					Message:  "File header Model.created is not a valid UTC date time (missing 'Z').",
					Severity: "sh.Violation",
				})
			}
		}

		scenarioField := val.FieldByName("ScenarioTime")
		if scenarioField.IsValid() && scenarioField.Kind() == reflect.String {
			v := scenarioField.String()
			if v != "" && !strings.HasSuffix(v, "Z") {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    goTypeName(obj),
					Property: "Model.scenarioTime",
					Message:  "File header Model.scenarioTime is not a valid UTC date time (missing 'Z').",
					Severity: "sh.Violation",
				})
			}
		}
	}
	return violations
}

// CheckFloatSpecialValues implements all600:Float-specialValues
// Profile: 61970-600-1_AllProfiles-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Float attributes are restricted not to use INF and NaN values.
func CheckFloatSpecialValues(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	for id, obj := range dataset.Elements {
		val := reflect.ValueOf(obj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		if val.Kind() != reflect.Struct {
			continue
		}

		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			if field.Kind() == reflect.Struct {
				// Check embedded struct fields (one level for Conductor in ACLineSegment)
				for j := 0; j < field.NumField(); j++ {
					subField := field.Field(j)
					if subField.Kind() == reflect.Float64 || subField.Kind() == reflect.Float32 {
						f := subField.Float()
						if math.IsNaN(f) || math.IsInf(f, 0) {
							violations = append(violations, Violation{
								ObjectID: id,
								Class:    goTypeName(obj),
								Property: field.Type().Field(j).Name,
								Message:  "INF or NaN used in an attribute defined as float.",
								Severity: "sh.Violation",
							})
						}
					}
				}
				continue
			}
			if field.Kind() == reflect.Float64 || field.Kind() == reflect.Float32 {
				f := field.Float()
				if math.IsNaN(f) || math.IsInf(f, 0) {
					violations = append(violations, Violation{
						ObjectID: id,
						Class:    goTypeName(obj),
						Property: val.Type().Field(i).Name,
						Message:  "INF or NaN used in an attribute defined as float.",
						Severity: "sh.Violation",
					})
				}
			}
		}
	}
	return violations
}

// CheckModelingAuthoritySetNotEmpty implements all600:Model.modelingAuthoritySet-marp10-12
// Profile: 61970-600-1_AllProfiles-AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: The modelingAuthoritySet property in the header must not be empty.
func CheckModelingAuthoritySetNotEmpty(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	for id, obj := range dataset.Elements {
		val := reflect.ValueOf(obj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		if val.Kind() != reflect.Struct {
			continue
		}

		maField := val.FieldByName("ModelingAuthoritySet")
		if maField.IsValid() && !maField.IsNil() {
			mridField := maField.Elem().FieldByName("MRID")
			if mridField.IsValid() && mridField.String() == "" {
				violations = append(violations, Violation{
					ObjectID: id,
					Class:    goTypeName(obj),
					Property: "Model.modelingAuthoritySet",
					Message:  "The modelingAuthoritySet property is defined as empty.",
					Severity: "sh.Violation",
				})
			}
		}
	}
	return violations
}

// CheckIdentifiedObjectStringLengths implements iosl.IdentifiedObject.shortName-stringLength, iosl.IdentifiedObject.energyIdentCodeEic-stringLength, iosl.IdentifiedObject.name-stringLength and iosl.IdentifiedObject.description-stringLength
// Profile: 61970-600-2_IdentifiedObjectCommon_AP-Con-Complex
// Origin: Derived from a SPARQL constraint.
// Description: Validates maximum string lengths for various IdentifiedObject attributes.
func CheckIdentifiedObjectStringLengths(dataset *cimgostructs.CIMElementList) []Violation {
	var violations []Violation
	for id, obj := range dataset.Elements {
		io, ok := getIdentifiedObject(obj)
		if !ok {
			continue
		}

		if len(io.ShortName) > 12 {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    goTypeName(obj),
				Property: "IdentifiedObject.shortName",
				Message:  "String length is greater than 12 characters.",
				Severity: "sh.Violation",
			})
		}
		if io.EnergyIdentCodeEic != "" && len(io.EnergyIdentCodeEic) != 16 {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    goTypeName(obj),
				Property: "IdentifiedObject.energyIdentCodeEic",
				Message:  "String length is not 16 characters.",
				Severity: "sh.Violation",
			})
		}
		if len(io.Name) > 128 {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    goTypeName(obj),
				Property: "IdentifiedObject.name",
				Message:  "String length is greater than 128 characters.",
				Severity: "sh.Violation",
			})
		}
		if len(io.Description) > 256 {
			violations = append(violations, Violation{
				ObjectID: id,
				Class:    goTypeName(obj),
				Property: "IdentifiedObject.description",
				Message:  "String length is greater than 256 characters.",
				Severity: "sh.Violation",
			})
		}
	}
	return violations
}

// CheckFileHeaderExists implements all600:All-HGEN2
// Profile: 61970-600-1_AllProfiles-AP-Con-Complex
// Origin: Derived from a complex SHACL constraint.
// Description: Each type of instance file (full or difference) shall have a file header.
func CheckFileHeaderExists(dataset *cimgostructs.CIMElementList) []Violation {
	if len(dataset.FullModels) == 0 && len(dataset.DifferenceModels) == 0 {
		return []Violation{{
			ObjectID: "global",
			Class:    "FullModel",
			Property: "rdf:type",
			Message:  "File header is missing.",
			Severity: "sh.Violation",
		}}
	}
	return nil
}
