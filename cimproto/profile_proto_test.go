package cimproto

import (
	"bytes"
	"cimgo/cimgostructs"
	"cimgo/cimprofiles"
	apiv1 "cimgo/proto/api/v1"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"google.golang.org/protobuf/proto"
)

func TestProfileToProto(t *testing.T) {
	t.Log("Start CIM-Data to Protobuf test")

	entry := "../testdata/test_001.xml"
	b, err := os.ReadFile(entry)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	cimData, err := cimprofiles.DecodeProfile(bytes.NewReader(b), nil)
	if err != nil {
		t.Fatalf("Failed to decode CIM profile: %v", err)
	}

	// 1. Convert cimgostructs.CIMElementList to apiv1.CIMElementList using exported function
	protoList, err := ToProto(cimData)
	if err != nil {
		t.Fatalf("Failed to convert to proto: %v", err)
	}

	// 2. Serialize to Protobuf
	protoData, err := proto.Marshal(protoList)
	if err != nil {
		t.Fatalf("Failed to marshal to protobuf: %v", err)
	}
	t.Logf("Serialized to protobuf: %d bytes", len(protoData))

	// 3. Deserialize from Protobuf
	decodedProtoList := &apiv1.CIMElementList{}
	err = proto.Unmarshal(protoData, decodedProtoList)
	if err != nil {
		t.Fatalf("Failed to unmarshal from protobuf: %v", err)
	}

	// 4. Output as JSON
	jsonOut, err := json.MarshalIndent(decodedProtoList, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal decoded proto to JSON: %v", err)
	}
	t.Log("Decoded CIM data from Protobuf (JSON):\n" + string(jsonOut))
}

func TestMergedProfileToProto(t *testing.T) {
	t.Log("Start Merged CIM-Data to Protobuf test")

	// Same as TestMergeData but with Proto conversion
	entries, err := filepath.Glob("../testdata/test_009_*[^out.].xml")
	if err != nil {
		t.Fatalf("Failed to glob test files: %v", err)
	}

	mergedCIMData := cimgostructs.NewCIMElementList()
	for _, entry := range entries {
		b, err := os.ReadFile(entry)
		if err != nil {
			t.Fatalf("Failed to read test file %s: %v", entry, err)
		}

		_, err = cimprofiles.DecodeProfile(bytes.NewReader(b), mergedCIMData)
		if err != nil {
			t.Fatalf("Failed to decode CIM profile %s: %v", entry, err)
		}
	}

	// 1. Convert cimgostructs.CIMElementList to apiv1.CIMElementList
	protoList, err := ToProto(mergedCIMData)
	if err != nil {
		t.Fatalf("Failed to convert merged data to proto: %v", err)
	}

	// 2. Serialize to Protobuf
	protoData, err := proto.Marshal(protoList)
	if err != nil {
		t.Fatalf("Failed to marshal to protobuf: %v", err)
	}
	t.Logf("Serialized merged data to protobuf: %d bytes", len(protoData))

	// 3. Deserialize from Protobuf
	decodedProtoList := &apiv1.CIMElementList{}
	err = proto.Unmarshal(protoData, decodedProtoList)
	if err != nil {
		t.Fatalf("Failed to unmarshal from protobuf: %v", err)
	}

	// 4. Output as JSON (just a summary for large data)
	t.Logf("Successfully decoded %d types from merged protobuf", reflect.ValueOf(decodedProtoList).Elem().NumField())
}
