package main

import (
	apiv1 "cimgo/proto/api/v1"
	"fmt"
	"log"
	"os"

	"google.golang.org/protobuf/proto"
)

func main() {
	const filePath = "line_segment.bin"

	// Construct the nested hierarchy
	// ACLineSegment -> Conductor -> ConductingEquipment -> Equipment -> PSR -> IdentifiedObject
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

	// Serialize to Binary and write to file
	data, err := proto.Marshal(line)
	if err != nil {
		log.Fatalf("Marshaling error: %v", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		log.Fatalf("File write error: %v", err)
	}
	fmt.Printf("Serialized ACLineSegment to %s (%d bytes)\n", filePath, len(data))

	// Read from File and deserialize
	readData, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("File read error: %v", err)
	}

	decodedLine := &apiv1.ACLineSegment{}
	err = proto.Unmarshal(readData, decodedLine)
	if err != nil {
		log.Fatalf("Unmarshaling error: %v", err)
	}

	// Accessing nested data safely using Getters
	fmt.Println("--- Decoded Data ---")
	fmt.Printf("Name:    %s\n", decodedLine.GetSuper().GetSuper().GetSuper().GetSuper().GetSuper().GetName())
	fmt.Printf("Voltage: %s\n", decodedLine.GetSuper().GetSuper().GetBaseVoltage())
	fmt.Printf("R:       %f\n", decodedLine.GetR())
	fmt.Printf("Length:  %f\n", decodedLine.GetSuper().GetLength())
}
