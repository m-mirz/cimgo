package cimprofiles

import (
	"bytes"
	"cimgo/cimgostructs"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func DecodeTest(t *testing.T, filePattern string) {

	entries, err := filepath.Glob(filePattern)
	if err != nil {
		log.Fatal(err)
	}
	t.Log("Read files:", entries)

	for _, entry := range entries {

		b, err := os.ReadFile(entry)
		if err != nil {
			log.Fatal(err)
		}

		cimData, err := DecodeProfile(bytes.NewReader(b), nil)
		if err != nil {
			log.Fatal(err)
		}

		jsonOut, err := json.MarshalIndent(cimData.Elements, "", "  ")
		if err != nil {
			t.Fatalf("Failed to create a nicely formatted JSON: %v", err)
		}
		t.Log("CIM data:\n" + string(jsonOut))
	}
}

func TestDecode001(t *testing.T) {
	DecodeTest(t, "../testdata/test_001.xml")
}

func TestDecode002(t *testing.T) {
	// The file contains an elements with rdf:about attribute, which is a common way to identify elements in RDF/XML.
	// The test checks if the decoder can handle this format correctly and extract the relevant information from the rdf:about attribute.
	// It is not required that rdf:about is also used for the export since this would require information which profile is exported.
	// The test is more about the flexibility of the decoder to handle different formats of CIM data.
	DecodeTest(t, "../testdata/test_002.xml")
}

func TestDecode003(t *testing.T) {
	DecodeTest(t, "../testdata/test_003.xml")
}

func TestDecode004(t *testing.T) {
	DecodeTest(t, "../testdata/test_004.xml")
}

func TestDecode005(t *testing.T) {
	DecodeTest(t, "../testdata/test_005.xml")
}

func TestMergeData(t *testing.T) {
	t.Log("Start merge decoding test")

	entries, err := filepath.Glob("../testdata/test_009_*[^out.].xml")
	if err != nil {
		log.Fatal(err)
	}
	t.Log("Read files:", entries)

	mergedCIMData := cimgostructs.NewCIMElementList()

	for _, entry := range entries {

		b, err := os.ReadFile(entry)
		if err != nil {
			panic(err)
		}

		_, err = DecodeProfile(bytes.NewReader(b), mergedCIMData)
		if err != nil {
			panic(err)
		}

		jsonOut, err := json.MarshalIndent(mergedCIMData.Elements, "", "  ")
		if err != nil {
			t.Fatalf("Failed to create a nicely formatted JSON: %v", err)
		}
		t.Log("Decoded CIM data:\n" + string(jsonOut))
	}
}
