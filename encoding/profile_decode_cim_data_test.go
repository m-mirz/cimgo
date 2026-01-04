package encoding

import (
	"bytes"
	"cimgo/encoding/cimgostructs"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeCIMData(t *testing.T) {
	t.Log("Start CIM-Data decoding test")

	entries, err := filepath.Glob("../testdata/test_001.xml")
	if err != nil {
		log.Fatal(err)
	}
	t.Log("Read files:", entries)

	for _, entry := range entries {

		b, err := os.ReadFile(entry)
		if err != nil {
			panic(err)
		}

		cimData, err := DecodeProfile(bytes.NewReader(b), nil)
		if err != nil {
			panic(err)
		}

		jsonOut, err := json.MarshalIndent(cimData.Elements, "", "  ")
		if err != nil {
			t.Fatalf("Failed to create a nicely formatted JSON: %v", err)
		}
		t.Log("Decoded CIM data:\n" + string(jsonOut))
	}
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
