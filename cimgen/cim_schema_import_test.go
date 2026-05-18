package cimgen

import (
	"crypto/sha256"
	"fmt"
	"os"
	"testing"
)

func TestSchemaImport(t *testing.T) {
	t.Log("Start CIM schema import test")

	output := "cim_schema_import_test.json"
	t.Log("Write imported schema to file:", output)
	f, err := os.Create(output)
	if err != nil {
		t.Error("Cannot create output file:", err)
	}
	defer f.Close()

	cimSpec := NewCIMSpecification()
	if err := cimSpec.ImportCIMSchemaFiles(CGMES3_SCHEMA); err != nil {
		t.Fatalf("ImportCIMSchemaFiles failed: %v", err)
	}

	if err := cimSpec.printSpecification(f); err != nil {
		t.Fatalf("printSpecification failed: %v", err)
	}

	// Compute hash of the output file for verification
	f.Sync()
	data, err := os.ReadFile(output)
	if err != nil {
		t.Error("Cannot read output file for hashing:", err)
	}
	hash := sha256.Sum256(data)
	t.Logf("SHA256 hash of output file: %x", hash)

	// Test output file against expected hash
	expectedHash := "4e1820630d12ba82772827b59857c1a90beff3455b2db105904d161346a76f94"
	if fmt.Sprintf("%x", hash) != expectedHash {
		t.Error("decoder tests failed, output file hash does not match expected hash")
	}
}
