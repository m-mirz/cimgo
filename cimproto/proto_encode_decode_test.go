package cimproto

import (
	apiv1 "cimgo/proto/api/v1"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"
)

func TestProtoEncodeDecode(t *testing.T) {
	line := &apiv1.ACLineSegment{
		R:   0.045,
		X:   0.12,
		Bch: 0.000005,
		Super: &apiv1.Conductor{
			Length: 50.5,
			Super: &apiv1.ConductingEquipment{
				BaseVoltage: "400kV",
				Super: &apiv1.Equipment{
					InService:         true,
					NormallyInService: true,
					Super: &apiv1.PowerSystemResource{
						Super: &apiv1.IdentifiedObject{
							MRID: "123-abc-456",
							Name: "North-South-Line-01",
						},
					},
				},
			},
		},
	}

	data, err := proto.Marshal(line)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	filePath := filepath.Join(t.TempDir(), "line_segment.bin")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	readData, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	decoded := &apiv1.ACLineSegment{}
	if err := proto.Unmarshal(readData, decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !proto.Equal(line, decoded) {
		t.Fatalf("round-trip mismatch:\noriginal: %v\ndecoded:  %v", line, decoded)
	}
}
