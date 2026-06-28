package cgmesxml

import (
	"bytes"
	"cimgo/cimstructs"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDecodeVoltageLevelAndBaseVoltage(t *testing.T) {
	b, err := os.ReadFile("../testdata/test_001_EQ.xml")
	if err != nil {
		t.Fatal(err)
	}
	cimData, err := DecodeProfile(bytes.NewReader(b), nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(cimData.ByID) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(cimData.ByID))
	}

	vl, ok := cimData.ByID["VoltageLevel.98"].(*cimstructs.VoltageLevel)
	if !ok {
		t.Fatal("VoltageLevel.98 not found or wrong type")
	}
	if vl.Name != "98" {
		t.Errorf("VoltageLevel.Name: got %q, want %q", vl.Name, "98")
	}
	if vl.BaseVoltage == nil || vl.BaseVoltage.MRID != "#BaseVoltage.20" {
		t.Errorf("VoltageLevel.BaseVoltage: got %v, want MRID=#BaseVoltage.20", vl.BaseVoltage)
	}

	bv, ok := cimData.ByID["BaseVoltage.20"].(*cimstructs.BaseVoltage)
	if !ok {
		t.Fatal("BaseVoltage.20 not found or wrong type")
	}
	if bv.NominalVoltage != 20.0 {
		t.Errorf("BaseVoltage.NominalVoltage: got %v, want 20.0", bv.NominalVoltage)
	}
}

// TestDecodeRDFAbout verifies that rdf:about IDs (with leading "#") are
// decoded to the same element map key as rdf:ID, so callers don't need to
// distinguish the two RDF identification forms.
func TestDecodeRDFAbout(t *testing.T) {
	b, err := os.ReadFile("../testdata/test_002_OP.xml")
	if err != nil {
		t.Fatal(err)
	}
	cimData, err := DecodeProfile(bytes.NewReader(b), nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(cimData.ByID) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(cimData.ByID))
	}

	a, ok := cimData.ByID["Analog.N0.Voltage"].(*cimstructs.Analog)
	if !ok {
		t.Fatal("Analog.N0.Voltage not found or wrong type")
	}
	if a.Name != "Voltage Magnitude Measurement at N0" {
		t.Errorf("Analog.Name: got %q, want %q", a.Name, "Voltage Magnitude Measurement at N0")
	}
	if a.MeasurementType != "Voltage" {
		t.Errorf("Analog.MeasurementType: got %q, want %q", a.MeasurementType, "Voltage")
	}

	av, ok := cimData.ByID["AnalogValue.N0.Voltage"].(*cimstructs.AnalogValue)
	if !ok {
		t.Fatal("AnalogValue.N0.Voltage not found or wrong type")
	}
	if av.Analog == nil || av.Analog.MRID != "#Analog.N0.Voltage" {
		t.Errorf("AnalogValue.Analog: got %v, want MRID=#Analog.N0.Voltage", av.Analog)
	}
}

func TestDecodeTerminalTopologicalNodeReference(t *testing.T) {
	b, err := os.ReadFile("../testdata/test_003_TP.xml")
	if err != nil {
		t.Fatal(err)
	}
	cimData, err := DecodeProfile(bytes.NewReader(b), nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(cimData.ByID) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(cimData.ByID))
	}

	tn, ok := cimData.ByID["N0"].(*cimstructs.TopologicalNode)
	if !ok {
		t.Fatal("TopologicalNode N0 not found or wrong type")
	}
	if tn.Name != "N0" {
		t.Errorf("TopologicalNode.Name: got %q, want %q", tn.Name, "N0")
	}

	term, ok := cimData.ByID["Terminal.N0"].(*cimstructs.Terminal)
	if !ok {
		t.Fatal("Terminal.N0 not found or wrong type")
	}
	if term.TopologicalNode == nil || term.TopologicalNode.MRID != "#N0" {
		t.Errorf("Terminal.TopologicalNode: got %v, want MRID=#N0", term.TopologicalNode)
	}
}

// TestDecodeEuropeanNamespaceExtension verifies that attributes in the eu:
// namespace (e.g. eu:IdentifiedObject.shortName) are decoded alongside
// standard cim: attributes on the same element.
func TestDecodeEuropeanNamespaceExtension(t *testing.T) {
	b, err := os.ReadFile("../testdata/test_004_EQ.xml")
	if err != nil {
		t.Fatal(err)
	}
	cimData, err := DecodeProfile(bytes.NewReader(b), nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(cimData.ByID) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(cimData.ByID))
	}

	fm, ok := cimData.ByID["test004"].(*cimstructs.FullModel)
	if !ok {
		t.Fatal("FullModel test004 not found or wrong type")
	}
	if len(fm.Profile) != 1 || fm.Profile[0] != "http://iec.ch/TC57/ns/CIM/CoreEquipment-EU/3.0" {
		t.Errorf("FullModel.Profile: got %v, want [http://iec.ch/TC57/ns/CIM/CoreEquipment-EU/3.0]", fm.Profile)
	}

	bv, ok := cimData.ByID["BaseVoltage.20"].(*cimstructs.BaseVoltage)
	if !ok {
		t.Fatal("BaseVoltage.20 not found or wrong type")
	}
	if bv.NominalVoltage != 20.0 {
		t.Errorf("BaseVoltage.NominalVoltage: got %v, want 20.0", bv.NominalVoltage)
	}
	if bv.ShortName != "20" {
		t.Errorf("BaseVoltage.ShortName: got %q, want %q", bv.ShortName, "20")
	}
}

// TestDecodeBoundaryPoint verifies that elements declared in the eu: namespace
// (eu:BoundaryPoint) are decoded as concrete Go types when loaded from separate
// EQ and EQBD profile files.
func TestDecodeBoundaryPoint(t *testing.T) {
	var readers []io.Reader
	for _, path := range []string{"../testdata/test_005_EQ.xml", "../testdata/test_005_EQBD.xml"} {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		readers = append(readers, bytes.NewReader(b))
	}
	cimData, err := DecodeProfiles(readers, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(cimData.ByID) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(cimData.ByID))
	}

	cn, ok := cimData.ByID["N0"].(*cimstructs.ConnectivityNode)
	if !ok {
		t.Fatal("ConnectivityNode N0 not found or wrong type")
	}
	if cn.Name != "N0" {
		t.Errorf("ConnectivityNode.Name: got %q, want %q", cn.Name, "N0")
	}

	bp, ok := cimData.ByID["N0_BP"].(*cimstructs.BoundaryPoint)
	if !ok {
		t.Fatal("BoundaryPoint N0_BP not found or wrong type")
	}
	if bp.ConnectivityNode == nil || bp.ConnectivityNode.MRID != "#N0" {
		t.Errorf("BoundaryPoint.ConnectivityNode: got %v, want MRID=#N0", bp.ConnectivityNode)
	}
}

// TestMergeProfiles verifies that decoding multiple profile files into a shared
// CIMDataset merges elements with the same ID rather than duplicating them.
// The EQ file declares Terminal.N0 (with name); the TP file re-declares it
// with a TopologicalNode reference. After merging both files the terminal must
// carry both the name and the reference.
func TestMergeProfiles(t *testing.T) {
	merged := cimstructs.NewCIMDataset()

	for _, file := range []string{
		"../testdata/test_009_EQ.xml",
		"../testdata/test_009_TP.xml",
	} {
		b, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := DecodeProfile(bytes.NewReader(b), merged); err != nil {
			t.Fatalf("%s: %v", file, err)
		}
	}

	if len(merged.ByID) != 2 {
		t.Fatalf("expected 2 elements after merge, got %d", len(merged.ByID))
	}

	tn, ok := merged.ByID["N0"].(*cimstructs.TopologicalNode)
	if !ok {
		t.Fatal("TopologicalNode N0 not found or wrong type after merge")
	}
	if tn.Name != "N0" {
		t.Errorf("TopologicalNode.Name: got %q, want %q", tn.Name, "N0")
	}

	term, ok := merged.ByID["Terminal.N0"].(*cimstructs.Terminal)
	if !ok {
		t.Fatal("Terminal.N0 not found or wrong type after merge")
	}
	if term.Name != "Terminal.N0" {
		t.Errorf("Terminal.Name after merge: got %q, want %q", term.Name, "Terminal.N0")
	}
	if term.TopologicalNode == nil || term.TopologicalNode.MRID != "#N0" {
		t.Errorf("Terminal.TopologicalNode after merge: got %v, want MRID=#N0", term.TopologicalNode)
	}
}

var benchDataset *cimstructs.CIMDataset

func BenchmarkImportRealGrid(b *testing.B) {
	const dir = "../CGMES-Test-Configurations/v3.0/RealGrid/RealGrid-Merged"
	entries, err := os.ReadDir(dir)
	if err != nil {
		b.Skipf("test data not available: %v", err)
	}
	var xmlFiles []string
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".xml" {
			xmlFiles = append(xmlFiles, filepath.Join(dir, e.Name()))
		}
	}

	var total time.Duration
	var iters int
	for b.Loop() {
		start := time.Now()
		dataset := cimstructs.NewCIMDataset()
		for _, file := range xmlFiles {
			raw, err := os.ReadFile(file)
			if err != nil {
				b.Fatal(err)
			}
			isolated, err := DecodeProfile(bytes.NewReader(raw), nil)
			if err != nil {
				b.Fatal(err)
			}
			if err := MergeInto(dataset, isolated); err != nil {
				b.Fatal(err)
			}
		}
		total += time.Since(start)
		iters++
		benchDataset = dataset
	}
	b.ReportMetric(float64(total.Milliseconds())/float64(iters), "ms/op")
}
