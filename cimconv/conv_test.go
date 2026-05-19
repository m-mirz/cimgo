package cimconv

import (
	"bytes"
	"cimgo/cgmesxml"
	"cimgo/cimbase"
	"cimgo/cimstructs"
	apiv1 "cimgo/proto/api/v1"
	"os"
	"testing"

	"google.golang.org/protobuf/proto"
)

// --- uriToEnumKey unit tests ---

func TestUriToEnumKey(t *testing.T) {
	cases := []struct {
		uri  string
		want string
	}{
		{"http://iec.ch/TC57/CIM100#PhaseCode.ABC", "PhaseCode_ABC"},
		{"http://iec.ch/TC57/2013/CIM-schema-cim16#UnitMultiplier.k", "UnitMultiplier_k"},
		{"http://iec.ch/TC57/2013/CIM-schema-cim16#UnitSymbol.V", "UnitSymbol_V"},
		{"http://example.org/ns#SomeEnum.Value1", "SomeEnum_Value1"},
		// no '#' → empty string
		{"not-a-uri", ""},
	}
	for _, c := range cases {
		got := uriToEnumKey(c.uri)
		if got != c.want {
			t.Errorf("uriToEnumKey(%q) = %q, want %q", c.uri, got, c.want)
		}
	}
}

// --- nil safety ---

func TestNilSafeConverters(t *testing.T) {
	if IdentifiedObjectToProto(nil) != nil {
		t.Error("IdentifiedObjectToProto(nil) should be nil")
	}
	if TerminalToProto(nil) != nil {
		t.Error("TerminalToProto(nil) should be nil")
	}
	if ACLineSegmentToProto(nil) != nil {
		t.Error("ACLineSegmentToProto(nil) should be nil")
	}
	if BaseVoltageToProto(nil) != nil {
		t.Error("BaseVoltageToProto(nil) should be nil")
	}
	if AnalogToProto(nil) != nil {
		t.Error("AnalogToProto(nil) should be nil")
	}
}

// --- IdentifiedObject converter ---

func TestIdentifiedObjectConverter(t *testing.T) {
	src := &cimstructs.IdentifiedObject{
		Base:        cimbase.Base{Id: "should-not-appear"},
		MRID:        "abc-123",
		Name:        "Test Object",
		Description: "A test",
		ShortName:   "TO",
	}
	dst := IdentifiedObjectToProto(src)
	if dst == nil {
		t.Fatal("expected non-nil result")
	}
	if dst.MRID != "abc-123" {
		t.Errorf("MRID: got %q, want %q", dst.MRID, "abc-123")
	}
	if dst.Name != "Test Object" {
		t.Errorf("Name: got %q, want %q", dst.Name, "Test Object")
	}
	if dst.Description != "A test" {
		t.Errorf("Description: got %q, want %q", dst.Description, "A test")
	}
	if dst.ShortName != "TO" {
		t.Errorf("ShortName: got %q, want %q", dst.ShortName, "TO")
	}
}

// --- BaseVoltage: Super chain + primitive field ---

func TestBaseVoltageConverter(t *testing.T) {
	src := &cimstructs.BaseVoltage{
		IdentifiedObject: cimstructs.IdentifiedObject{
			MRID: "bv-001",
			Name: "20kV",
		},
		NominalVoltage: 20.0,
	}
	dst := BaseVoltageToProto(src)
	if dst == nil {
		t.Fatal("expected non-nil result")
	}
	if dst.NominalVoltage != 20.0 {
		t.Errorf("NominalVoltage: got %v, want %v", dst.NominalVoltage, 20.0)
	}
	if dst.Super == nil {
		t.Fatal("Super should not be nil")
	}
	if dst.Super.MRID != "bv-001" {
		t.Errorf("Super.MRID: got %q, want %q", dst.Super.MRID, "bv-001")
	}
	if dst.Super.Name != "20kV" {
		t.Errorf("Super.Name: got %q, want %q", dst.Super.Name, "20kV")
	}
}

// --- ACLineSegment: full Super chain + numeric fields ---

func TestACLineSegmentConverter(t *testing.T) {
	src := &cimstructs.ACLineSegment{
		Conductor: cimstructs.Conductor{
			ConductingEquipment: cimstructs.ConductingEquipment{
				Equipment: cimstructs.Equipment{
					PowerSystemResource: cimstructs.PowerSystemResource{
						IdentifiedObject: cimstructs.IdentifiedObject{
							MRID: "line-001",
							Name: "North-South-Line",
						},
					},
				},
			},
			Length: 50.0,
		},
		R:   0.045,
		X:   0.12,
		Bch: 0.000005,
	}
	dst := ACLineSegmentToProto(src)
	if dst == nil {
		t.Fatal("expected non-nil result")
	}
	if dst.R != 0.045 {
		t.Errorf("R: got %v, want %v", dst.R, 0.045)
	}
	if dst.X != 0.12 {
		t.Errorf("X: got %v, want %v", dst.X, 0.12)
	}
	if dst.Bch != 0.000005 {
		t.Errorf("Bch: got %v, want %v", dst.Bch, 0.000005)
	}
	// traverse Super chain: ACLineSegment → Conductor → ConductingEquipment → Equipment → PSR → IdentifiedObject
	io := dst.Super.Super.Super.Super.Super
	if io == nil {
		t.Fatal("IdentifiedObject Super should not be nil")
	}
	if io.MRID != "line-001" {
		t.Errorf("Super chain MRID: got %q, want %q", io.MRID, "line-001")
	}
	if io.Name != "North-South-Line" {
		t.Errorf("Super chain Name: got %q, want %q", io.Name, "North-South-Line")
	}
	if dst.Super.Length != 50.0 {
		t.Errorf("Conductor.Length: got %v, want %v", dst.Super.Length, 50.0)
	}
}

// --- Terminal: '#' stripping on reference fields ---

func TestTerminalConverter_Reference(t *testing.T) {
	src := &cimstructs.Terminal{
		TopologicalNode: &struct {
			MRID string `xml:"resource,attr"`
		}{MRID: "#N0"},
		ConnectivityNode: &struct {
			MRID string `xml:"resource,attr"`
		}{MRID: "#CN1"},
	}
	dst := TerminalToProto(src)
	if dst == nil {
		t.Fatal("expected non-nil result")
	}
	if dst.TopologicalNode != "N0" {
		t.Errorf("TopologicalNode: got %q, want %q", dst.TopologicalNode, "N0")
	}
	if dst.ConnectivityNode != "CN1" {
		t.Errorf("ConnectivityNode: got %q, want %q", dst.ConnectivityNode, "CN1")
	}
	// reference without '#' prefix should pass through unchanged
	src.TopologicalNode.MRID = "plain-id"
	dst2 := TerminalToProto(src)
	if dst2.TopologicalNode != "plain-id" {
		t.Errorf("TopologicalNode (no #): got %q, want %q", dst2.TopologicalNode, "plain-id")
	}
}

// --- Terminal: URI → proto enum ---

func TestTerminalConverter_Enum(t *testing.T) {
	src := &cimstructs.Terminal{
		Phases: &struct {
			URI string `xml:"resource,attr"`
		}{URI: "http://iec.ch/TC57/CIM100#PhaseCode.ABC"},
	}
	dst := TerminalToProto(src)
	if dst == nil {
		t.Fatal("expected non-nil result")
	}
	want := apiv1.PhaseCode(apiv1.PhaseCode_value["PhaseCode_ABC"])
	if dst.Phases != want {
		t.Errorf("Phases: got %v, want %v", dst.Phases, want)
	}
	if dst.Phases == 0 {
		t.Error("Phases should be non-zero for a known enum value")
	}
}

// --- AccumulatorLimitSet: slice of references ---

func TestAccumulatorLimitSetConverter_Slices(t *testing.T) {
	src := &cimstructs.AccumulatorLimitSet{
		Measurements: []struct {
			MRID string `xml:"resource,attr"`
		}{
			{MRID: "#Meas1"},
			{MRID: "#Meas2"},
			{MRID: "Meas3"}, // no '#'
		},
	}
	dst := AccumulatorLimitSetToProto(src)
	if dst == nil {
		t.Fatal("expected non-nil result")
	}
	if len(dst.Measurements) != 3 {
		t.Fatalf("Measurements len: got %d, want %d", len(dst.Measurements), 3)
	}
	if dst.Measurements[0] != "Meas1" {
		t.Errorf("[0]: got %q, want %q", dst.Measurements[0], "Meas1")
	}
	if dst.Measurements[1] != "Meas2" {
		t.Errorf("[1]: got %q, want %q", dst.Measurements[1], "Meas2")
	}
	if dst.Measurements[2] != "Meas3" {
		t.Errorf("[2]: got %q, want %q", dst.Measurements[2], "Meas3")
	}
}

// --- Integration: test_001.xml (VoltageLevel + BaseVoltage) ---

func TestToProto_VoltageLevel(t *testing.T) {
	b, err := os.ReadFile("../testdata/test_001.xml")
	if err != nil {
		t.Fatalf("read test file: %v", err)
	}
	cimData, err := cgmesxml.DecodeProfile(bytes.NewReader(b), nil)
	if err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	dst, err := ToProto(cimData)
	if err != nil {
		t.Fatalf("ToProto: %v", err)
	}

	if len(dst.VoltageLevel) != 1 {
		t.Fatalf("VoltageLevel count: got %d, want 1", len(dst.VoltageLevel))
	}
	vl := dst.VoltageLevel[0]
	if vl.BaseVoltage != "BaseVoltage.20" {
		t.Errorf("VoltageLevel.BaseVoltage: got %q, want %q", vl.BaseVoltage, "BaseVoltage.20")
	}
	// VoltageLevel → EquipmentContainer → ConnectivityNodeContainer → PowerSystemResource → IdentifiedObject
	io := vl.Super.Super.Super.Super
	if io == nil {
		t.Fatal("VoltageLevel IdentifiedObject Super should not be nil")
	}
	if io.Name != "98" {
		t.Errorf("VoltageLevel Name: got %q, want %q", io.Name, "98")
	}

	if len(dst.BaseVoltage) != 1 {
		t.Fatalf("BaseVoltage count: got %d, want 1", len(dst.BaseVoltage))
	}
	bv := dst.BaseVoltage[0]
	if bv.NominalVoltage != 20.0 {
		t.Errorf("BaseVoltage.NominalVoltage: got %v, want %v", bv.NominalVoltage, 20.0)
	}
}

// --- Integration: test_003.xml (Terminal '#' reference stripping) ---

func TestToProto_TerminalReference(t *testing.T) {
	b, err := os.ReadFile("../testdata/test_003.xml")
	if err != nil {
		t.Fatalf("read test file: %v", err)
	}
	cimData, err := cgmesxml.DecodeProfile(bytes.NewReader(b), nil)
	if err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	dst, err := ToProto(cimData)
	if err != nil {
		t.Fatalf("ToProto: %v", err)
	}

	if len(dst.Terminal) != 1 {
		t.Fatalf("Terminal count: got %d, want 1", len(dst.Terminal))
	}
	if dst.Terminal[0].TopologicalNode != "N0" {
		t.Errorf("Terminal.TopologicalNode: got %q, want %q", dst.Terminal[0].TopologicalNode, "N0")
	}
}

// --- Integration: test_002.xml (Analog enum fields) ---

func TestToProto_EnumFields(t *testing.T) {
	b, err := os.ReadFile("../testdata/test_002.xml")
	if err != nil {
		t.Fatalf("read test file: %v", err)
	}
	cimData, err := cgmesxml.DecodeProfile(bytes.NewReader(b), nil)
	if err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	dst, err := ToProto(cimData)
	if err != nil {
		t.Fatalf("ToProto: %v", err)
	}

	if len(dst.Analog) != 1 {
		t.Fatalf("Analog count: got %d, want 1", len(dst.Analog))
	}
	a := dst.Analog[0]
	// Super chain leads to Measurement which holds the enum fields
	m := a.Super
	if m == nil {
		t.Fatal("Analog.Super (Measurement) should not be nil")
	}
	wantMult := apiv1.UnitMultiplier(apiv1.UnitMultiplier_value["UnitMultiplier_k"])
	if m.UnitMultiplier != wantMult {
		t.Errorf("UnitMultiplier: got %v, want %v", m.UnitMultiplier, wantMult)
	}
	wantSym := apiv1.UnitSymbol(apiv1.UnitSymbol_value["UnitSymbol_V"])
	if m.UnitSymbol != wantSym {
		t.Errorf("UnitSymbol: got %v, want %v", m.UnitSymbol, wantSym)
	}
	if m.UnitMultiplier == 0 {
		t.Error("UnitMultiplier should be non-zero for a known enum value")
	}
	if m.UnitSymbol == 0 {
		t.Error("UnitSymbol should be non-zero for a known enum value")
	}
}

// --- Integration: proto round-trip (test_001.xml) ---

func TestToProto_RoundTrip(t *testing.T) {
	b, err := os.ReadFile("../testdata/test_001.xml")
	if err != nil {
		t.Fatalf("read test file: %v", err)
	}
	cimData, err := cgmesxml.DecodeProfile(bytes.NewReader(b), nil)
	if err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	protoList, err := ToProto(cimData)
	if err != nil {
		t.Fatalf("ToProto: %v", err)
	}

	data, err := proto.Marshal(protoList)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	decoded := &apiv1.CIMDataset{}
	if err := proto.Unmarshal(data, decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !proto.Equal(protoList, decoded) {
		t.Error("proto round-trip: marshaled and unmarshaled messages differ")
	}
}
