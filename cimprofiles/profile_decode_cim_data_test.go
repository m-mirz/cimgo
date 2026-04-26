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
