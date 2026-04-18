package main

import (
	"cimgo/rdf/shacl"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ConstraintWrapper struct {
	Type string
	Data shacl.Constraint
}

func (cw ConstraintWrapper) MarshalJSON() ([]byte, error) {
	// We want to combine the Type field and the data from the constraint
	data, err := json.Marshal(cw.Data)
	if err != nil {
		return nil, err
	}

	var m map[string]any
	err = json.Unmarshal(data, &m)
	if err != nil {
		// If it's not a map (e.g. basic type), just use a simple wrapper
		return json.Marshal(map[string]any{
			"type": cw.Type,
			"data": cw.Data,
		})
	}

	m["_type"] = cw.Type
	return json.Marshal(m)
}

type ShapeWrapper struct {
	*shacl.Shape
	Constraints []ConstraintWrapper
	Properties  []*ShapeWrapper
}

func wrapShape(s *shacl.Shape) *ShapeWrapper {
	if s == nil {
		return nil
	}
	sw := &ShapeWrapper{
		Shape: s,
	}
	for _, c := range s.Constraints {
		sw.Constraints = append(sw.Constraints, ConstraintWrapper{
			Type: c.ComponentIRI(),
			Data: c,
		})
	}
	for _, ps := range s.Properties {
		sw.Properties = append(sw.Properties, wrapShape(ps))
	}
	return sw
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: shacl_json_export <file1.ttl> [file2.ttl ...]")
		os.Exit(1)
	}

	for _, file := range os.Args[1:] {
		g, err := shacl.LoadTurtleFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading %s: %v\n", file, err)
			continue
		}

		shapes := shacl.ParseShapes(g)
		wrapped := make(map[string]*ShapeWrapper)
		for k, s := range shapes {
			wrapped[k] = wrapShape(s)
		}
		
		data, err := json.MarshalIndent(wrapped, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling shapes from %s: %v\n", file, err)
			continue
		}
		
		fmt.Printf("--- %s ---\n", filepath.Base(file))
		fmt.Println(string(data))
	}
}
