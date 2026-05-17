package cgmesxml

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeEncodePSTPhaseTapChangerLinearType1(t *testing.T) {
	t.Log("Start PST PhaseTapChangerLinear Type1 decode-encode test")

	entries, err := filepath.Glob("../CGMES-Test-Configurations/v3.0/PST/PST_PhaseTapChangerLinear_Type1/PST_Type1_*.xml")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Read files:", entries)

	readers := make([]io.Reader, len(entries))
	for i, entry := range entries {
		b, err := os.ReadFile(entry)
		if err != nil {
			t.Fatal(err)
		}
		readers[i] = bytes.NewReader(b)
	}

	cimData, err := DecodeProfiles(readers, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Decoded %d CIM elements", len(cimData.Elements))

	const testDir = "../CGMES-Test-Configurations/v3.0/PST/PST_PhaseTapChangerLinear_Type1/"

	for _, profileCode := range []string{"EQ", "SSH", "SV", "TP", "DL"} {
		outPath := testDir + "PST_PhaseTapChangerLinear_Type1_" + profileCode + ".out.xml"
		f, err := os.Create(outPath)
		if err != nil {
			t.Fatal(err)
		}

		if err := EncodeForProfile(f, cimData, profileCode); err != nil {
			f.Close()
			t.Fatal(err)
		}
		f.Close()

		t.Logf("Profile %s: written to %s", profileCode, outPath)
	}
}
