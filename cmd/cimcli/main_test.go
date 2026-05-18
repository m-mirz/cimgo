package main

import (
	"bytes"
	"cimgo/cgmesxml"
	"cimgo/cimconv"
	"cimgo/cimstructs"
	apiv1 "cimgo/proto/api/v1"
	"encoding/json"
	"os"
	"testing"

	"google.golang.org/protobuf/proto"
)

// loadXML decodes one or more CGMES XML files and merges them into one dataset.
func loadXML(t *testing.T, paths ...string) *cimstructs.CIMElementList {
	t.Helper()
	dataset := cimstructs.NewCIMElementList()
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		isolated, err := cgmesxml.DecodeProfile(bytes.NewReader(b), nil)
		if err != nil {
			t.Fatalf("decode %s: %v", p, err)
		}
		if err := cgmesxml.MergeInto(dataset, isolated); err != nil {
			t.Fatalf("merge %s: %v", p, err)
		}
	}
	return dataset
}

// --- XML → JSON round-trip ---

func TestJSONRoundTrip_ElementCount(t *testing.T) {
	original := loadXML(t, "../../testdata/test_001.xml")

	jsonBytes, err := marshalWithType(original)
	if err != nil {
		t.Fatalf("marshalWithType: %v", err)
	}
	recovered, err := unmarshalWithType(jsonBytes)
	if err != nil {
		t.Fatalf("unmarshalWithType: %v", err)
	}

	if len(recovered.Elements) != len(original.Elements) {
		t.Errorf("element count: got %d, want %d", len(recovered.Elements), len(original.Elements))
	}
}

func TestJSONRoundTrip_TypeField(t *testing.T) {
	original := loadXML(t, "../../testdata/test_001.xml")

	jsonBytes, err := marshalWithType(original)
	if err != nil {
		t.Fatalf("marshalWithType: %v", err)
	}

	// Every entry in the flat map must carry a non-empty "_type".
	var raw map[string]map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &raw); err != nil {
		t.Fatalf("unmarshal raw JSON: %v", err)
	}
	for id, m := range raw {
		typ, ok := m["_type"]
		if !ok {
			t.Errorf("element %s missing _type", id)
		} else if typ == "" {
			t.Errorf("element %s has empty _type", id)
		}
	}
}

func TestJSONRoundTrip_FieldValues(t *testing.T) {
	// test_001.xml: 1 BaseVoltage (nominalVoltage=20) + 1 VoltageLevel (name="98")
	original := loadXML(t, "../../testdata/test_001.xml")

	jsonBytes, err := marshalWithType(original)
	if err != nil {
		t.Fatalf("marshalWithType: %v", err)
	}
	recovered, err := unmarshalWithType(jsonBytes)
	if err != nil {
		t.Fatalf("unmarshalWithType: %v", err)
	}

	// BaseVoltage: numeric field preserved
	if len(recovered.BaseVoltages) != 1 {
		t.Fatalf("BaseVoltage count: got %d, want 1", len(recovered.BaseVoltages))
	}
	for _, bv := range recovered.BaseVoltages {
		if bv.NominalVoltage != 20.0 {
			t.Errorf("NominalVoltage: got %v, want 20.0", bv.NominalVoltage)
		}
	}

	// VoltageLevel: string field and reference field preserved
	if len(recovered.VoltageLevels) != 1 {
		t.Fatalf("VoltageLevel count: got %d, want 1", len(recovered.VoltageLevels))
	}
	for _, vl := range recovered.VoltageLevels {
		if vl.Name != "98" {
			t.Errorf("VoltageLevel.Name: got %q, want %q", vl.Name, "98")
		}
		if vl.BaseVoltage == nil || vl.BaseVoltage.MRID == "" {
			t.Error("VoltageLevel.BaseVoltage reference not preserved")
		}
	}
}

func TestJSONRoundTrip_EnumField(t *testing.T) {
	// test_002.xml: Analog with UnitMultiplier and UnitSymbol URI fields
	original := loadXML(t, "../../testdata/test_002.xml")

	jsonBytes, err := marshalWithType(original)
	if err != nil {
		t.Fatalf("marshalWithType: %v", err)
	}
	recovered, err := unmarshalWithType(jsonBytes)
	if err != nil {
		t.Fatalf("unmarshalWithType: %v", err)
	}

	if len(recovered.Analogs) != 1 {
		t.Fatalf("Analog count: got %d, want 1", len(recovered.Analogs))
	}
	for _, a := range recovered.Analogs {
		if a.UnitMultiplier == nil || a.UnitMultiplier.URI == "" {
			t.Error("Analog.UnitMultiplier URI not preserved")
		}
		if a.UnitSymbol == nil || a.UnitSymbol.URI == "" {
			t.Error("Analog.UnitSymbol URI not preserved")
		}
	}
}

func TestJSONRoundTrip_MultiFile(t *testing.T) {
	// test_009: EQ + TP profiles merged
	original := loadXML(t,
		"../../testdata/test_009_EQ.xml",
		"../../testdata/test_009_TP.xml",
	)

	jsonBytes, err := marshalWithType(original)
	if err != nil {
		t.Fatalf("marshalWithType: %v", err)
	}
	recovered, err := unmarshalWithType(jsonBytes)
	if err != nil {
		t.Fatalf("unmarshalWithType: %v", err)
	}

	if len(recovered.Elements) != len(original.Elements) {
		t.Errorf("element count: got %d, want %d", len(recovered.Elements), len(original.Elements))
	}
}

// --- XML → proto round-trip ---

func TestProtoRoundTrip_MarshalUnmarshal(t *testing.T) {
	dataset := loadXML(t, "../../testdata/test_001.xml")

	protoList, err := cimconv.ToProto(dataset)
	if err != nil {
		t.Fatalf("ToProto: %v", err)
	}

	data, err := proto.Marshal(protoList)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("marshalled proto is empty")
	}

	decoded := &apiv1.CIMElementList{}
	if err := proto.Unmarshal(data, decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !proto.Equal(protoList, decoded) {
		t.Error("proto round-trip: messages differ after marshal/unmarshal")
	}
}

func TestProtoRoundTrip_FieldValues(t *testing.T) {
	// Verify that field values survive the proto round-trip.
	dataset := loadXML(t, "../../testdata/test_001.xml")

	protoList, err := cimconv.ToProto(dataset)
	if err != nil {
		t.Fatalf("ToProto: %v", err)
	}

	data, err := proto.Marshal(protoList)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	decoded := &apiv1.CIMElementList{}
	if err := proto.Unmarshal(data, decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if len(decoded.BaseVoltage) != 1 {
		t.Fatalf("BaseVoltage count: got %d, want 1", len(decoded.BaseVoltage))
	}
	if decoded.BaseVoltage[0].NominalVoltage != 20.0 {
		t.Errorf("NominalVoltage: got %v, want 20.0", decoded.BaseVoltage[0].NominalVoltage)
	}

	if len(decoded.VoltageLevel) != 1 {
		t.Fatalf("VoltageLevel count: got %d, want 1", len(decoded.VoltageLevel))
	}
	// VoltageLevel → EquipmentContainer → ConnectivityNodeContainer → PowerSystemResource → IdentifiedObject
	io := decoded.VoltageLevel[0].Super.Super.Super.Super
	if io == nil {
		t.Fatal("VoltageLevel IdentifiedObject chain should not be nil")
	}
	if io.Name != "98" {
		t.Errorf("VoltageLevel.Name: got %q, want %q", io.Name, "98")
	}
}

func TestProtoRoundTrip_MultiFile(t *testing.T) {
	dataset := loadXML(t,
		"../../testdata/test_009_EQ.xml",
		"../../testdata/test_009_TP.xml",
	)

	protoList, err := cimconv.ToProto(dataset)
	if err != nil {
		t.Fatalf("ToProto: %v", err)
	}

	data, err := proto.Marshal(protoList)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	decoded := &apiv1.CIMElementList{}
	if err := proto.Unmarshal(data, decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !proto.Equal(protoList, decoded) {
		t.Error("proto round-trip (multi-file): messages differ after marshal/unmarshal")
	}
}
