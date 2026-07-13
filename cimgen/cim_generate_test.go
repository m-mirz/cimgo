package cimgen

import (
	"log"
	"testing"

	"golang.org/x/mod/sumdb/dirhash"
)

func TestGenerate(t *testing.T) {
	t.Log("Start CIM code generation test")

	cimSpec := NewCIMSpecification()
	err := cimSpec.ImportCIMSchemaFiles(CGMES3_SCHEMA)
	if err != nil {
		t.Fatalf("ImportCIMSchemaFiles failed: %v", err)
	}

	outputDir := "../cimstructs"
	err = cimSpec.GenerateGo(outputDir)
	if err != nil {
		t.Fatalf("GenerateGo failed: %v", err)
	}

	hash, err := dirhash.HashDir(outputDir, "", dirhash.Hash1)
	if err != nil {
		log.Fatal(err)
	}
	t.Logf("Directory Hash: %s\n", hash)

	// Test directory hash
	expectedHash := "h1:c2+iYun3XqSrzRAhNHgaqAgqgTKZ6d0uCe1VL/pcOYo="
	if hash != expectedHash {
		t.Error("decoder tests failed, output file hash does not match expected hash")
	}
}
